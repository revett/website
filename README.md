# Website (revcd.com)

Personal website 👋

A static site built with a tiny custom Go generator, Tailwind (standalone binary, no Node), and
Markdown content. Deployed to GitHub Pages via Actions.

## Layout

```
content/    Markdown pages, posts, now editions, images, files
web/        Templates, CSS, static assets (fonts, favicon, robots.txt)
generator/  The Go tool that renders content + web into dist/
```

## Usage

```
make serve   # local dev server on :8080 with live reload
make build   # render the full site into dist/
make check   # validate internal links in dist/
```

The only dependencies are Go (plus goldmark and yaml.v3) and the Tailwind standalone binary, which
`make` downloads into `bin/` on first run.

## Content

Pages are Markdown with YAML frontmatter. Posts live in `content/posts/` and their filename is their
URL slug. Now editions live in `content/now/`; they all render onto `/now`, newest first, and an
`alias` field emits a redirect stub at an edition's legacy URL.

The generator also emits `sitemap.xml`, `llms.txt`, a `404.html`, and a raw markdown mirror of every
page (e.g. `/cv.md`).

