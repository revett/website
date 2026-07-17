# AGENTS.md

## What This Is

A hand-rolled static site generator (`build.ts`) with no framework and no conventions to memorise:
read it top to bottom and you know everything. Markdown in `content/` becomes HTML pages via
`template.html`, styled by Tailwind (`style.css`). No client side JavaScript.

## Development

- `npm run build` builds into `public/`
- `npm run dev` builds, serves on `:1234` (override with `PORT`), and rebuilds on file changes
- Line length is set to 100 characters for all project files

## Code Style

Write TypeScript as if it were Go: simple, explicit, boring. When unsure, ask what the dullest Go
programmer would do.

- Pure functions only; no classes
- Named exports only
- String literal unions over enums; `erasableSyntaxOnly` in tsconfig enforces strippable syntax
- `type` over `interface`: data shapes are structs, not contracts
- Errors are values where practical; `build.ts` throws on malformed content on purpose, so a bad
  page fails the build loudly instead of shipping broken
- Guard clauses and early returns; flat beats nested; no `else` after a `return`
- No ternary expressions; write the `if`
- Braces on every `if`, even a one line body
- No clever generics, no decorators, no magic
- A small, focused dependency beats hand rolling something fiddly to get right (`marked` for
  markdown, the Tailwind CLI for CSS); it loses to hand rolling the moment it drags in a framework
  or an SDK we don't need

## Scope

`build.ts` is capped on purpose: markdown in, HTML + markdown twins out, copy static files, compile
CSS. No image pipeline, no shortcodes, no aliases, no plugins. If a change needs more than that,
the answer is a real static site generator, not a bigger `build.ts`.

## Footer

Ensure to update the date in the `<footer>` in `template.html` for me.
