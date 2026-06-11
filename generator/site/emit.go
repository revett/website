package site

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func absURL(path string) string {
	if path == "" || strings.HasPrefix(path, "http") {
		return path
	}

	return baseURL + path
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	return os.CopyFS(dst, os.DirFS(src))
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func writeLLMs(s *site) error {
	funcs := template.FuncMap{
		"trimSlash": func(s string) string { return strings.TrimSuffix(s, "/") },
	}

	t, err := template.New("llms.txt.tmpl").Funcs(funcs).ParseFiles("web/templates/llms.txt.tmpl")
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = t.Execute(&out, struct {
		Base string
		Name string
		Site *site
	}{
		Base: baseURL,
		Name: siteName,
		Site: s,
	})
	if err != nil {
		return err
	}

	return writeFile(filepath.Join(distDir, "llms.txt"), out.Bytes())
}

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
		if err := write(strings.Trim(p.Slug, "/")+".md", p.Title, p.Source); err != nil {
			return err
		}
	}

	return nil
}

func writeSitemap(s *site) error {
	type entry struct {
		Loc     string
		Lastmod string
	}

	format := func(p *page) string {
		if p.Updated.IsZero() {
			return ""
		}

		return p.Updated.Format("2006-01-02")
	}

	entries := []entry{
		{Loc: baseURL + "/", Lastmod: format(s.Home)},
		{Loc: baseURL + "/cv/", Lastmod: format(s.CV)},
	}

	if len(s.Now) > 0 {
		entries = append(entries, entry{Loc: baseURL + "/now/", Lastmod: format(s.Now[0])})
	}

	for _, p := range s.Posts {
		entries = append(entries, entry{Loc: baseURL + p.Slug, Lastmod: format(p)})
	}

	t, err := template.ParseFiles("web/templates/sitemap.xml.tmpl")
	if err != nil {
		return err
	}

	var out bytes.Buffer
	if err := t.Execute(&out, entries); err != nil {
		return err
	}

	return writeFile(filepath.Join(distDir, "sitemap.xml"), out.Bytes())
}
