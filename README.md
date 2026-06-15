# Website (revcd.com)

Personal website 👋

A static site built with [Hugo](https://gohugo.io). Content in Markdown,
hand-written CSS, no JavaScript. Hosted on GitHub Pages.

## Layout

```
content/    Markdown: home (_index.md), cv, now, posts
layouts/    Hugo templates
static/     style.css, fonts, favicon, robots.txt, images, files
hugo.toml   config
```

## Usage

```
hugo server   # dev server on :1313 with live reload
hugo          # render the site into public/
```

The only dependency is Hugo:

```
brew install hugo
```

## Content

Pages are Markdown with YAML frontmatter. Posts live in `content/posts/` and
their filename is their URL slug, served at the site root (`/:contentbasename/`).
`content/now.md` is a single page with one section per edition, newest first;
its `aliases` emit redirect stubs for the old per-edition URLs.

Hugo emits `sitemap.xml`, `llms.txt` (a home output format), and a `404.html`
for free.
