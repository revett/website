# Website (revcd.com)

Personal website 👋

Markdown pages built into a static site by `build.ts`, a ~180 line generator
with no conventions: read it top to bottom and you know everything. Run
directly by Node (24+), no build step. Hosted on GitHub Pages; pushing to
`main` deploys.

## Layout

```
content/        Markdown pages (index.md is the home page)
template.html   the HTML shell every page is rendered into
style.css       Tailwind input, compiled to public/style.css
static/         favicon, robots.txt (copied verbatim)
build.ts        the generator
```

## Usage

```
npm install

npm run build   # build into public/
npm run dev     # build, serve on http://localhost:8080 (override with
                # PORT=n), and rebuild when source files change
```

## Content

New page: create `content/foo.md` with `title` and `description` frontmatter,
it renders at `/foo/`.

Each page also gets a raw markdown twin (`/foo/index.md`, linked via
`<link rel="alternate">`), and the build emits `/llms.txt` (an index of all
pages for machines) and `/sitemap.xml`. All generated, never edited.

The generator's scope is capped on purpose: markdown in, HTML + twins out,
copy static. No image pipeline, no shortcodes, no aliases. If it starts
needing features, use a real static site generator instead of growing this.
