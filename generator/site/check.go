package site

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var linkRe = regexp.MustCompile(`(?:href|src)="([^"]+)"`)

// Check validates that every internal link and asset reference in dist/ resolves to a generated
// file.
func Check() error {
	var broken []string

	err := filepath.WalkDir(distDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		for _, m := range linkRe.FindAllStringSubmatch(string(data), -1) {
			link := strings.TrimPrefix(m[1], baseURL)
			if !strings.HasPrefix(link, "/") || strings.HasPrefix(link, "//") {
				continue
			}

			if !resolves(link) {
				broken = append(broken, fmt.Sprintf("%s -> %s", path, m[1]))
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking %s: %w", distDir, err)
	}

	if len(broken) > 0 {
		return fmt.Errorf("%d broken internal links:\n  %s", len(broken), strings.Join(broken, "\n  "))
	}

	fmt.Println("links ok")

	return nil
}

func resolves(link string) bool {
	link, _, _ = strings.Cut(link, "#")
	link, _, _ = strings.Cut(link, "?")

	if link == "" || link == "/" {
		link = "/index.html"
	}

	path := filepath.Join(distDir, filepath.FromSlash(strings.TrimPrefix(link, "/")))

	for _, candidate := range []string{path, filepath.Join(path, "index.html")} {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return true
		}
	}

	return false
}
