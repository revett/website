package site

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

type frontmatter struct {
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
	frontmatter
	Slug   string
	Anchor string
	Body   template.HTML
	Source string
}

type project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
	Image       string `yaml:"image"`
}

type site struct {
	Home  *page
	CV    *page
	Posts []*page
	Now   []*page
}

var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->\n?`)

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

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
	if err := yaml.Unmarshal(front, &p.frontmatter); err != nil {
		return nil, fmt.Errorf("parsing %s frontmatter: %w", path, err)
	}

	p.Source = strings.TrimSpace(htmlCommentRe.ReplaceAllString(string(body), ""))

	var buf bytes.Buffer
	if err := markdown.Convert([]byte(p.Source), &buf); err != nil {
		return nil, fmt.Errorf("rendering %s markdown: %w", path, err)
	}
	p.Body = template.HTML(buf.String())

	return p, nil
}

func loadSite() (*site, error) {
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
		return nil, fmt.Errorf("globbing posts: %w", err)
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
		return nil, fmt.Errorf("globbing now editions: %w", err)
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

func nowPage(s *site) *page {
	p := &page{Slug: "/now/"}
	p.Title = "Now"
	p.Description = "What Charlie Revett is working on, up to, and thinking about right now; " +
		"updated a couple of times a year."

	if len(s.Now) > 0 {
		current := s.Now[0]

		p.Date = current.Date
		p.Updated = current.Updated

		p.OGImage = current.OGImage
		if p.OGImage == "" {
			p.OGImage = current.Cover
		}
	}

	return p
}
