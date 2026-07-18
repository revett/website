// Builds the site the way the deploy does (SITE_URL is the GitHub Pages
// project URL) and asserts each generated file renders correctly. A failing
// check exits non-zero, which blocks the deploy workflow before it uploads.

import { execFileSync } from "node:child_process";
import fs from "node:fs";
import assert from "node:assert/strict";
import { before, test } from "node:test";

// The URL the deploy builds with; kept in sync with .github/workflows/deploy.yml.
const SITE_URL = "https://revett.github.io/website";

type Check = {
  name: string;
  file: string;
  want: string[];
};

const checks: Check[] = [
  {
    name: "/ renders with the right title, base-prefixed assets, and content",
    file: "public/index.html",
    want: [
      "<title>Charlie Revett</title>",
      'href="/website/style.css"',
      'href="/website/favicon.png"',
      "plain.com",
      "<footer>",
    ],
  },
  {
    name: "/index.md renders the raw markdown twin pointing back at llms.txt",
    file: "public/index.md",
    want: [
      "> See: https://revett.github.io/website/llms.txt",
      "# Charlie Revett",
      "plain.com",
    ],
  },
  {
    name: "/llms.txt renders with the CV, pages, and home links",
    file: "public/llms.txt",
    want: [
      "# Charlie Revett",
      "## CV",
      "## Pages",
      "- [Home](https://revett.github.io/website/index.md)",
    ],
  },
  {
    name: "404 renders styled, with base-prefixed assets and a home link",
    file: "public/404.html",
    want: [
      "<title>Not found · Charlie Revett</title>",
      'href="/website/style.css"',
      "That page doesn't exist",
      'href="/website/">head back home</a>',
    ],
  },
  {
    name: "/sitemap.xml renders with the home URL at top priority",
    file: "public/sitemap.xml",
    want: [
      '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
      "<loc>https://revett.github.io/website/</loc>",
      "<priority>1.0</priority>",
    ],
  },
  {
    name: "/robots.txt renders allowing all with a sitemap reference",
    file: "public/robots.txt",
    want: [
      "User-agent: *",
      "Allow: /",
      "Sitemap: https://revett.github.io/website/sitemap.xml",
    ],
  },
];

before(() => {
  execFileSync("node", ["build.ts"], { env: { ...process.env, SITE_URL } });
});

for (const check of checks) {
  test(check.name, () => {
    const content = fs.readFileSync(check.file, "utf8");
    for (const want of check.want) {
      assert.ok(content.includes(want), `${check.file}: missing ${JSON.stringify(want)}`);
    }
  });
}
