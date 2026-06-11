package site

import (
	"bytes"
	"html/template"
	"path/filepath"
	"time"
)

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

type pageData struct {
	Site *site
	Page *page
	Dev  bool
}

func parseTemplates() (*template.Template, error) {
	funcs := template.FuncMap{
		"date":  func(t time.Time) string { return t.Format("January 2, 2006") },
		"iso":   func(t time.Time) string { return t.Format("2006-01-02") },
		"thumb": func(src string) string { return "/images/thumbs/" + filepath.Base(src) },
	}

	return template.New("").Funcs(funcs).ParseGlob("web/templates/*.html")
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

func renderPage(t *template.Template, s *site, p *page, name string, dev bool) error {
	var content bytes.Buffer
	if err := t.ExecuteTemplate(&content, name, pageData{Site: s, Page: p, Dev: dev}); err != nil {
		return err
	}

	title := p.Title + " · " + siteName
	ogType := "article"
	if p.Slug == "/" {
		title = siteName + " · Software Engineer"
		ogType = "website"
	}

	var out bytes.Buffer
	err := t.ExecuteTemplate(&out, "base", baseData{
		Title:       title,
		Description: p.Description,
		Canonical:   baseURL + p.Slug,
		OGImage:     absURL(p.OGImage),
		OGType:      ogType,
		Year:        time.Now().Year(),
		Dev:         dev,
		Content:     template.HTML(content.String()),
	})
	if err != nil {
		return err
	}

	return writeFile(filepath.Join(distDir, filepath.FromSlash(p.Slug), "index.html"), out.Bytes())
}

func writeRedirect(t *template.Template, e *page) error {
	if e.Alias == "" {
		return nil
	}

	var out bytes.Buffer
	err := t.ExecuteTemplate(&out, "redirect", struct {
		Canonical string
		Target    string
	}{
		Canonical: baseURL + "/now/",
		Target:    "/now/#" + e.Anchor,
	})
	if err != nil {
		return err
	}

	return writeFile(filepath.Join(distDir, filepath.FromSlash(e.Alias), "index.html"), out.Bytes())
}
