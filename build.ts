// Builds the site. Every markdown file in content/ becomes a page:
// content/index.md is the home page, content/foo.md renders at /foo/.
// Each page gets a raw markdown twin (index.md) next to its index.html,
// and the site gets /llms.txt and /sitemap.xml. static/ is copied verbatim,
// and style.css is compiled by Tailwind into public/style.css.
//
//   node build.ts      build into public/
//   node build.ts dev  build, serve public/ on http://localhost:1234,
//                      and rebuild when source files change

import { execFileSync } from "node:child_process";
import fs from "node:fs";
import http from "node:http";
import path from "node:path";
import { marked } from "marked";

const SITE_TITLE = "Charlie Revett";
const BANNER_ALT = "Serene lake with a stilted house, green hills, and misty mountains";
const CV = `I am not actively looking to change role, but am always happy to talk to folks building interesting
products and incredible teams.

I'm interested in applied AI roles within small, high agency, high impact teams, with a genuine
product engineering culture, with founders that have done it before, and that are remote friendly,
requiring max 1/day per week in London (expensed).`;

const PORT = Number(process.env.PORT ?? 1234);
// In dev mode, every generated link (pages, llms.txt, sitemap) points at the
// server you're actually looking at, unless SITE_URL was set explicitly
// (e.g. to rehearse a deploy target).
if (process.argv[2] === "dev" && !process.env.SITE_URL) {
  process.env.SITE_URL = `http://localhost:${PORT}`;
}

// Full origin the site is served from, no trailing slash. Override at build
// time with SITE_URL, e.g. while living at a GitHub Pages project URL
// (https://revett.github.io/website) before a custom domain is wired up.
const BASE_URL = (process.env.SITE_URL ?? "https://revcd.com").replace(/\/$/, "");
// The path portion of BASE_URL ("" for a root domain, "/website" for a
// project page), used to prefix root-relative asset links.
const BASE_PATH = new URL(BASE_URL).pathname.replace(/\/$/, "");

type Page = {
  slug: string; // "" for the home page
  file: string; // path to the source content file
  title: string;
  description: string;
  markdown: string; // raw markdown body, frontmatter stripped
  body: string; // rendered markdown
};

function pageURL(page: Page): string {
  return page.slug === "" ? `${BASE_URL}/` : `${BASE_URL}/${page.slug}/`;
}

// The date content/foo.md was last committed, YYYY-MM-DD, or undefined if
// it isn't tracked yet (a new, uncommitted page).
function lastmod(file: string): string | undefined {
  try {
    const date = execFileSync(
      "git",
      ["log", "-1", "--format=%cd", "--date=short", "--", file],
      { encoding: "utf8" },
    ).trim();
    return date === "" ? undefined : date;
  } catch {
    return undefined;
  }
}

// A content file is a "---" delimited frontmatter block with title and
// description lines, then the markdown body.
function parse(file: string): Page {
  const raw = fs.readFileSync(file, "utf8");
  if (!raw.startsWith("---\n")) throw new Error(`${file}: missing frontmatter`);
  const end = raw.indexOf("\n---\n");
  if (end === -1) throw new Error(`${file}: unterminated frontmatter`);
  const front = raw.slice(4, end);
  const markdown = raw.slice(end + 5).replace(/^\n+/, "");

  let slug = path.basename(file, ".md");
  if (slug === "index") slug = "";
  let title = "";
  let description = "";
  for (const line of front.split("\n")) {
    const colon = line.indexOf(":");
    if (colon === -1) continue;
    const key = line.slice(0, colon).trim();
    const value = line.slice(colon + 1).trim();
    if (key === "title") title = value;
    if (key === "description") description = value;
  }
  if (!title || !description) {
    throw new Error(`${file}: frontmatter needs title and description`);
  }

  const body = marked.parse(markdown) as string;
  return { slug, file, title, description, markdown, body };
}

function escapeHTML(text: string): string {
  return text
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function render(template: string, page: Page): string {
  const pageTitle =
    page.slug === "" ? SITE_TITLE : `${page.title} · ${SITE_TITLE}`;
  return template
    .replaceAll("{{pageTitle}}", escapeHTML(pageTitle))
    .replaceAll("{{title}}", escapeHTML(page.title))
    .replaceAll("{{description}}", escapeHTML(page.description))
    .replaceAll("{{url}}", pageURL(page))
    .replaceAll("{{base}}", BASE_PATH)
    .replaceAll("{{body}}", page.body);
}

function llmsTxt(pages: Page[]): string {
  const home = pages.find((page) => page.slug === "");
  if (!home) throw new Error("content/index.md is required");
  const banner = `![${BANNER_ALT}](${BASE_URL}/banner.png)`;
  const links = [`- [Home](${pageURL(home)}index.md)`];
  for (const page of pages) {
    if (page.slug === "") continue;
    links.push(`- [${page.title}](${pageURL(page)}index.md): ${page.description}`);
  }
  return (
    [
      `# ${SITE_TITLE}`,
      home.markdown.trimEnd(),
      banner,
      `## CV\n\n${CV}`,
      `## Pages\n\n${links.join("\n")}`,
    ].join("\n\n") + "\n"
  );
}

function sitemapXML(pages: Page[]): string {
  const urls = pages
    .map((page) => {
      const modified = lastmod(page.file);
      const fields = [`    <loc>${pageURL(page)}</loc>`];
      if (modified) fields.push(`    <lastmod>${modified}</lastmod>`);
      fields.push("    <changefreq>monthly</changefreq>");
      fields.push(`    <priority>${page.slug === "" ? "1.0" : "0.5"}</priority>`);
      return `  <url>\n${fields.join("\n")}\n  </url>`;
    })
    .join("\n");
  return `<?xml version="1.0" encoding="utf-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls}
</urlset>
`;
}

function build(): void {
  const template = fs.readFileSync("template.html", "utf8");
  fs.rmSync("public", { recursive: true, force: true });
  fs.cpSync("static", "public", { recursive: true });

  const pages = fs
    .readdirSync("content")
    .filter((file) => file.endsWith(".md"))
    .sort()
    .map((file) => parse(path.join("content", file)));

  for (const page of pages) {
    const dir = page.slug === "" ? "public" : path.join("public", page.slug);
    fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(path.join(dir, "index.html"), render(template, page));
    fs.writeFileSync(
      path.join(dir, "index.md"),
      `# ${page.title}\n\n${page.markdown}`,
    );
  }

  fs.writeFileSync("public/llms.txt", llmsTxt(pages));
  fs.writeFileSync("public/sitemap.xml", sitemapXML(pages));

  execFileSync("node_modules/.bin/tailwindcss", [
    "--input",
    "style.css",
    "--output",
    "public/style.css",
  ]);
}

function watch(): void {
  let timer: NodeJS.Timeout | undefined;
  const rebuild = () => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      try {
        build();
        console.log("rebuilt");
      } catch (error) {
        console.error(error);
      }
    }, 100);
  };
  for (const target of ["content", "static", "template.html", "style.css"]) {
    fs.watch(target, { recursive: true }, rebuild);
  }
}

function serve(): void {
  const types: Record<string, string> = {
    ".html": "text/html",
    ".css": "text/css",
    ".md": "text/markdown",
    ".txt": "text/plain",
    ".xml": "application/xml",
    ".png": "image/png",
    ".jpg": "image/jpeg",
    ".svg": "image/svg+xml",
  };
  const root = path.resolve("public");
  http
    .createServer((req, res) => {
      const pathname = new URL(req.url ?? "/", BASE_URL).pathname;
      let file = path.join(root, decodeURIComponent(pathname));
      if (!file.startsWith(root)) {
        res.writeHead(400).end();
        return;
      }
      if (fs.existsSync(file) && fs.statSync(file).isDirectory()) {
        file = path.join(file, "index.html");
      }
      if (!fs.existsSync(file)) {
        res.writeHead(404, { "content-type": "text/plain" }).end("not found");
        return;
      }
      res.writeHead(200, {
        "content-type": types[path.extname(file)] ?? "application/octet-stream",
      });
      res.end(fs.readFileSync(file));
    })
    .listen(PORT, () => console.log(`serving on http://localhost:${PORT}`));
}

build();
if (process.argv[2] === "dev") {
  watch();
  serve();
}
