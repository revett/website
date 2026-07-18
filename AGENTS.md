# AGENTS.md

## Commands

- `npm run build` builds the site into `public/`
- `npm run dev` builds, serves on http://localhost:1234, and rebuilds on change
- `npm test` builds the way the deploy does, then asserts on generated files

## Code Style

It is critically important that you abide by all the rules set out in the `typescript-as-go` skill
when writing Typescript, no exceptions.

Line length is set to 100 characters for all project files.

## Footer

Ensure to update the date in the `<footer>` in `template.html` for me, as changes are made in the
project.

## Tests

Tests live in `build.test.ts`, table driven with `node:test` and `node:assert/strict`, no framework.
They build the site the way the deploy does and assert that some generated files render correctly.
`npm test` gates the deploy workflow, so run it before pushing. When you change what the build
emits, update the checks table to match.

## Links

Asset and page links must be prefixed with the site base path, `{{base}}` in `template.html` or
`BASE_PATH` in `build.ts`, never root relative (`/style.css`). The site lives at a GitHub Pages
project URL (`/website/`) for now, so a root relative link resolves against the wrong path and
silently breaks.

## Documentation

It is your responsibility to update documentation as changes are made in the project. This covers
the following files:

- `AGENTS.md`
- `README.md`

## Scope

The lightweight static site generator in `build.ts` is simple and functional by design. We like
simplicity. We are building a personal website, not a rocket ship, so keep that in mind. Always
remember that less is always more, simple is always better, boring is best, and to avoid the magic;
but we must still meet requirements.
