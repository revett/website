// Package site renders the markdown content in content/ through the
// templates in web/templates/ into a static site in dist/.
package site

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	baseURL  = "https://revcd.com"
	distDir  = "dist"
	siteName = "Charlie Revett (@revcd)"
)

// Build renders the full site into dist/.
func Build() error {
	return build(false)
}

func build(dev bool) error {
	s, err := loadSite()
	if err != nil {
		return fmt.Errorf("loading content: %w", err)
	}

	t, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("parsing templates: %w", err)
	}

	if err := os.RemoveAll(distDir); err != nil {
		return fmt.Errorf("cleaning %s: %w", distDir, err)
	}

	if err := renderPage(t, s, s.Home, "home", dev); err != nil {
		return fmt.Errorf("rendering home: %w", err)
	}

	if err := renderPage(t, s, s.CV, "page", dev); err != nil {
		return fmt.Errorf("rendering cv: %w", err)
	}

	if err := renderPage(t, s, nowPage(s), "now", dev); err != nil {
		return fmt.Errorf("rendering now: %w", err)
	}

	for _, p := range s.Posts {
		if err := renderPage(t, s, p, "post", dev); err != nil {
			return fmt.Errorf("rendering %s: %w", p.Slug, err)
		}
	}

	if err := renderNotFound(t, dev); err != nil {
		return fmt.Errorf("rendering 404: %w", err)
	}

	for _, e := range s.Now {
		if err := writeRedirect(t, e); err != nil {
			return fmt.Errorf("writing redirect for %s: %w", e.Anchor, err)
		}
	}

	if err := writeSitemap(s); err != nil {
		return fmt.Errorf("writing sitemap: %w", err)
	}

	if err := writeLLMs(s); err != nil {
		return fmt.Errorf("writing llms.txt: %w", err)
	}

	if err := writeMirrors(s); err != nil {
		return fmt.Errorf("writing markdown mirrors: %w", err)
	}

	for src, dst := range map[string]string{
		"web/static":     distDir,
		"content/images": filepath.Join(distDir, "images"),
		"content/files":  filepath.Join(distDir, "files"),
	} {
		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("copying %s: %w", src, err)
		}
	}

	return nil
}
