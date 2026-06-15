# Website (revcd.com)

Personal website 👋

A static site built with [Hugo](https://gohugo.io), styled with Tailwind
(standalone binary, no Node), content in Markdown. Hosted on GitHub Pages.

## Layout

```
content/    Markdown: home (_index.md), cv, now editions, posts
layouts/    Hugo templates
assets/     css/main.css (Tailwind entrypoint)
static/     fonts, favicon, robots.txt, images, files
hugo.toml   config
```

## Usage

```
make serve    # dev server on :1313 with live reload
make build    # compile CSS + render the site into public/
make css      # compile Tailwind only
make thumbs   # regenerate the committed post-cover thumbnails (macOS)
```

Requires `hugo` (extended) and `tailwindcss` on `PATH`:

```
brew install hugo tailwindcss
```

Hugo can't drive the standalone Tailwind binary (its CSS pipeline expects the
npm build), so Tailwind runs as its own step and Hugo serves the output at
`/style.css`. `make serve` runs both with live reload.

## Content

Pages are Markdown with YAML frontmatter. Posts live in `content/posts/` and
their filename is their URL slug, served at the site root. Now editions live in
`content/now/`; each is marked headless (`build.render: never`) so they all
render onto `/now` newest-first, and `content/now/_index.md` lists the old
edition URLs under `aliases` to emit redirect stubs.

Hugo emits `sitemap.xml`, `llms.txt` (a home output format), and a `404.html`
for free.
