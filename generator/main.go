// Command generator builds revcd.com. It renders the markdown files in
// content/ through the templates in web/templates/ into dist/, and also emits
// the sitemap, llms.txt, markdown mirrors, and redirect stubs.
//
// Usage:
//
//	go run ./generator build    render the site into dist/
//	go run ./generator serve    build, serve on :8080, and live reload
//	go run ./generator check    validate internal links in dist/
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

const (
	baseURL  = "https://revcd.com"
	siteName = "Charlie Revett (@revcd)"
	distDir  = "dist"
)

type project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
	Image       string `yaml:"image"`
}

type meta struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	Updated     time.Time `yaml:"updated"`
	Cover       string    `yaml:"cover"`
	OGImage     string    `yaml:"ogImage"`
	Caption     string    `yaml:"caption"`
	Period      string    `yaml:"period"`
	Alias       string    `yaml:"alias"`
	Portrait    string    `yaml:"portrait"`
	Projects    []project `yaml:"projects"`
}

type page struct {
	meta
	Slug   string // URL path with trailing slash, e.g. "/cv/"
	Anchor string // fragment on /now/ for now editions
	Body   template.HTML
	Source string // raw markdown body, used for the .md mirrors
}

type site struct {
	Home  *page
	CV    *page
	Posts []*page // newest first
	Now   []*page // newest first; first edition is current
}

func main() {
	cmd := "build"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	switch cmd {
	case "build":
		err = build(false)
	case "serve":
		err = serve(":8080")
	case "check":
		err = check()
	default:
		err = fmt.Errorf("unknown command %q (want build, serve, or check)", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

var commentRe = regexp.MustCompile(`(?s)<!--.*?-->\n?`)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func build(dev bool) error {
	s, err := load()
	if err != nil {
		return err
	}
	t, err := parseTemplates()
	if err != nil {
		return err
	}

	if err := os.RemoveAll(distDir); err != nil {
		return err
	}

	pages := append([]*page{s.Home, s.CV, nowPage(s)}, s.Posts...)
	for _, p := range pages {
		if err := renderPage(t, s, p, dev); err != nil {
			return err
		}
	}
	if err := renderNotFound(t, dev); err != nil {
		return err
	}
	for _, e := range s.Now {
		if err := writeRedirect(e); err != nil {
			return err
		}
	}
	if err := writeSitemap(s); err != nil {
		return err
	}
	if err := writeLLMs(s); err != nil {
		return err
	}
	if err := writeMirrors(s); err != nil {
		return err
	}

	for src, dst := range map[string]string{
		"web/static":     distDir,
		"content/images": filepath.Join(distDir, "images"),
		"content/files":  filepath.Join(distDir, "files"),
	} {
		if err := copyDir(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func load() (*site, error) {
	s := &site{}

	var err error
	if s.Home, err = loadPage("content/home.md", "/"); err != nil {
		return nil, err
	}
	if s.CV, err = loadPage("content/cv.md", "/cv/"); err != nil {
		return nil, err
	}

	posts, err := filepath.Glob("content/posts/*.md")
	if err != nil {
		return nil, err
	}
	for _, f := range posts {
		slug := strings.TrimSuffix(filepath.Base(f), ".md")
		p, err := loadPage(f, "/"+slug+"/")
		if err != nil {
			return nil, err
		}
		s.Posts = append(s.Posts, p)
	}
	sort.Slice(s.Posts, func(i, j int) bool { return s.Posts[i].Date.After(s.Posts[j].Date) })

	editions, err := filepath.Glob("content/now/*.md")
	if err != nil {
		return nil, err
	}
	for _, f := range editions {
		p, err := loadPage(f, "/now/")
		if err != nil {
			return nil, err
		}
		p.Anchor = strings.TrimSuffix(filepath.Base(f), ".md")
		s.Now = append(s.Now, p)
	}
	sort.Slice(s.Now, func(i, j int) bool { return s.Now[i].Date.After(s.Now[j].Date) })

	return s, nil
}

func loadPage(path, slug string) (*page, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	front, body, ok := bytes.Cut(bytes.TrimPrefix(raw, []byte("---\n")), []byte("\n---\n"))
	if !ok || len(front) == len(raw) {
		return nil, fmt.Errorf("%s: missing frontmatter", path)
	}

	p := &page{Slug: slug}
	if err := yaml.Unmarshal(front, &p.meta); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	// HTML comments (e.g. TODO notes) stay in the source files; they never
	// ship in the rendered HTML or the markdown mirrors.
	p.Source = strings.TrimSpace(commentRe.ReplaceAllString(string(body), ""))

	var buf bytes.Buffer
	if err := md.Convert([]byte(p.Source), &buf); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	p.Body = template.HTML(buf.String())
	return p, nil
}

// nowPage wraps the now editions into a single renderable page at /now/.
func nowPage(s *site) *page {
	p := &page{Slug: "/now/"}
	p.Title = "Now"
	p.Description = "What Charlie Revett is working on, up to, and thinking about right now; updated a couple of times a year."
	if len(s.Now) > 0 {
		current := s.Now[0]
		p.Updated = current.Updated
		p.Date = current.Date
		p.OGImage = current.OGImage
		if p.OGImage == "" {
			p.OGImage = current.Cover
		}
	}
	return p
}

func parseTemplates() (*template.Template, error) {
	funcs := template.FuncMap{
		"date": func(t time.Time) string { return t.Format("January 2, 2006") },
		"iso":  func(t time.Time) string { return t.Format("2006-01-02") },
		// thumb maps an image to its small variant in /images/thumbs/,
		// regenerated via `make thumbs`. check catches missing ones.
		"thumb": func(src string) string {
			return "/images/thumbs/" + filepath.Base(src)
		},
	}
	return template.New("").Funcs(funcs).ParseGlob("web/templates/*.html")
}

type renderData struct {
	Site *site
	Page *page
	Dev  bool
}

type baseData struct {
	Title       string
	Description string
	Canonical   string
	OGImage     string
	OGType      string
	Year        int
	Dev         bool
	Content     template.HTML
}

func renderPage(t *template.Template, s *site, p *page, dev bool) error {
	name := map[string]string{"/": "home", "/cv/": "page", "/now/": "now"}[p.Slug]
	if name == "" {
		name = "post"
	}

	var content bytes.Buffer
	if err := t.ExecuteTemplate(&content, name, renderData{Site: s, Page: p, Dev: dev}); err != nil {
		return err
	}

	title := p.Title + " · " + siteName
	ogType := "article"
	if p.Slug == "/" {
		title = siteName + " · Software Engineer"
		ogType = "website"
	}
	b := baseData{
		Title:       title,
		Description: p.Description,
		Canonical:   baseURL + p.Slug,
		OGImage:     absURL(p.OGImage),
		OGType:      ogType,
		Year:        time.Now().Year(),
		Dev:         dev,
		Content:     template.HTML(content.String()),
	}

	var out bytes.Buffer
	if err := t.ExecuteTemplate(&out, "base", b); err != nil {
		return err
	}
	return writeFile(filepath.Join(distDir, filepath.FromSlash(p.Slug), "index.html"), out.Bytes())
}

func renderNotFound(t *template.Template, dev bool) error {
	var content bytes.Buffer
	if err := t.ExecuteTemplate(&content, "404", nil); err != nil {
		return err
	}
	var out bytes.Buffer
	err := t.ExecuteTemplate(&out, "base", baseData{
		Title:       "Page not found · " + siteName,
		Description: "Page not found.",
		Canonical:   baseURL + "/404.html",
		OGType:      "website",
		Year:        time.Now().Year(),
		Dev:         dev,
		Content:     template.HTML(content.String()),
	})
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(distDir, "404.html"), out.Bytes())
}

// writeRedirect emits a stub at a now edition's legacy URL (e.g.
// /now/summer-autumn-2024) pointing at its section on /now/.
func writeRedirect(e *page) error {
	if e.Alias == "" {
		return nil
	}
	target := "/now/#" + e.Anchor
	stub := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Redirecting…</title>
<link rel="canonical" href="%s/now/">
<meta http-equiv="refresh" content="0; url=%s">
<meta name="robots" content="noindex">
</head>
<body><p>This page has moved to <a href="%s">revcd.com/now</a>.</p></body>
</html>
`, baseURL, target, target)
	return writeFile(filepath.Join(distDir, filepath.FromSlash(e.Alias), "index.html"), []byte(stub))
}

func writeSitemap(s *site) error {
	type entry struct {
		loc     string
		lastmod time.Time
	}
	entries := []entry{
		{baseURL + "/", s.Home.Updated},
		{baseURL + "/cv/", s.CV.Updated},
	}
	if len(s.Now) > 0 {
		entries = append(entries, entry{baseURL + "/now/", s.Now[0].Updated})
	}
	for _, p := range s.Posts {
		entries = append(entries, entry{baseURL + p.Slug, p.Updated})
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, e := range entries {
		b.WriteString("  <url>\n    <loc>" + e.loc + "</loc>\n")
		if !e.lastmod.IsZero() {
			b.WriteString("    <lastmod>" + e.lastmod.Format("2006-01-02") + "</lastmod>\n")
		}
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>\n")
	return writeFile(filepath.Join(distDir, "sitemap.xml"), []byte(b.String()))
}

func writeLLMs(s *site) error {
	var b strings.Builder
	b.WriteString("# " + siteName + "\n\n")
	b.WriteString("> " + s.Home.Description + "\n\n")
	b.WriteString("Markdown versions of every page are linked below.\n\n")
	b.WriteString("## Pages\n\n")
	b.WriteString(fmt.Sprintf("- [About](%s/index.md): %s\n", baseURL, s.Home.Description))
	b.WriteString(fmt.Sprintf("- [CV](%s/cv.md): %s\n", baseURL, s.CV.Description))
	b.WriteString(fmt.Sprintf("- [Now](%s/now.md): What Charlie is working on and thinking about right now.\n", baseURL))
	b.WriteString("\n## Posts\n\n")
	for _, p := range s.Posts {
		b.WriteString(fmt.Sprintf("- [%s](%s%s.md): %s\n", p.Title, baseURL, strings.TrimSuffix(p.Slug, "/"), p.Description))
	}
	return writeFile(filepath.Join(distDir, "llms.txt"), []byte(b.String()))
}

// writeMirrors emits a raw markdown version of every page alongside its HTML,
// for agents and curl-minded humans.
func writeMirrors(s *site) error {
	write := func(path, title, body string) error {
		return writeFile(filepath.Join(distDir, path), []byte("# "+title+"\n\n"+body+"\n"))
	}
	if err := write("index.md", s.Home.Title, s.Home.Source); err != nil {
		return err
	}
	if err := write("cv.md", "CV", s.CV.Source); err != nil {
		return err
	}

	var now strings.Builder
	for i, e := range s.Now {
		if i > 0 {
			now.WriteString("\n\n")
		}
		fmt.Fprintf(&now, "## %s\n\nCovering %s.\n\n%s", e.Title, e.Period, e.Source)
	}
	if err := write("now.md", "Now", now.String()); err != nil {
		return err
	}

	for _, p := range s.Posts {
		if err := write(strings.TrimSuffix(strings.TrimPrefix(p.Slug, "/"), "/")+".md", p.Title, p.Source); err != nil {
			return err
		}
	}
	return nil
}

func absURL(path string) string {
	if path == "" || strings.HasPrefix(path, "http") {
		return path
	}
	return baseURL + path
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFile(target, data)
	})
}
