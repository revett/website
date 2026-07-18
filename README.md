# Website

Repo for my personal website, which really is just a bunch of Markdown pages that I hand write,
which are then built into a static site for GitHub pages to host by `build.ts`, a ~200 line
generator that Claude maintains.

```plaintext
content/        Markdown pages
template.html   HTML shell
style.css       Tailwind input → compiled to public/style.css
static/         favicon, banner etc; robots.txt is generated, not here
build.ts        Generator
```

## Inspiration

The following websites have inspired this project:

- https://deadsimplesites.com
- https://emilkowal.ski
- https://shud.in
- https://paco.me
- https://raphaelsalaja.com
- https://benji.org
- https://nat.org
- https://gregbrockman.com
- https://adityainduraj.com
- https://megans.website
- https://kyh.io
- https://shedsgns.me
- https://hamiltonulmer.com
- https://david.kjelkerud.com

## Development

```bash
npm i
npm run dev
```

## Acknowledgements

- Favicon is the `:mountain:` emoji from
  [Samsung One UI 8.5](https://emojipedia.org/samsung/one-ui-8.5) set
