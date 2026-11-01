# d.idle branding

> status: first concept round (card `brand`), nothing applied anywhere yet. Assets: `assets/branding/concepts/`; contact sheet: `assets/branding/concepts/index.html` (open in a browser; dark/light toggle). Regenerate with `python3 assets/branding/gen.py` (stdlib only).

The name is always written with a lowercase d: **d.idle**, never "D.idle". The repo is `d2lod-bot`, the public repo `d-idle`.

## The three concepts

- **cursor** (recommended): the prompt chevron and a block cursor, terminal ember palette. The bot sits at the prompt for days, waiting. Pixel wordmark with a block cursor in the header.
- **rune**: an original ring sigil (stem, triangular bowl, diamond dot, twelve clock ticks), arcane violet and teal on night. Closest to the game's mood without borrowing anything from it.
- **tick**: a pixel mage with a glowing staff on an isometric tile and a segmented progress bar that never fills; cool blues, a nod to the lockstep tick.

Each concept folder holds `logo.svg` (horizontal lockup, transparent, readable on dark and light), `mark.svg`, `icon.svg` (512x512, rounded), `header.svg` (1280x640, GitHub social preview and README), `header-discord.svg` (960x540), `ascii.txt` (large and one-line variants), `palette.md` (colours and contrast ratios).

## Recommendation: cursor

It is the most distinctive at small sizes (a chevron and a block read at 32 px; the icon grid is whole 2 px cells at 32 px), it says "idle" and "terminal" without any game imagery, the pixel wordmark is cheap to reproduce in ASCII for the startup banner and README, and its dark ember look sits well with the stream page (`#0d1016`) and the status page dark theme. Rune is the fallback if a more illustrative icon is wanted.

## Usage rules

- Lowercase d, always: `d.idle`. The dot is part of the name and keeps its accent colour.
- Clear space: at least one wordmark pixel cell (the logo files already include it) on every side; for the icon, nothing within the rounded corner.
- Minimum size: logo 120 px wide, mark and icon 16 px (use the icon, not the header, below 64 px). Do not scale pixel art to non-integer sizes in print; in browsers use `image-rendering: pixelated` for PNG renders.
- Do not: recolour outside the palette, stretch, add shadows or outlines, set the name in another font, put header colours on a light background, or combine concepts.
- Everything is drawn as shapes by our own pixel font; no installed font, third-party image, Blizzard or Diablo II logo, font, art, sprite, icon or screenshot is used or may be mixed in.
- Licence: original work of the project; the licence follows the repository licence, which is the GNU General Public License v3 (see `LICENSE`).
- PNG renders: not produced in this run (the worker had no rasterizer: no rsvg-convert, inkscape, ImageMagick, cairosvg or PIL). The SVGs are the masters; render later where a rasterizer exists.

## Application checklist (nothing done; each needs the user's go)

- [ ] GitHub social preview (header 1280x640): needs the user's go
- [ ] GitHub avatar (icon): needs the user's go
- [ ] Discord server icon and banner (icon, header-discord 960x540): needs the user's go
- [ ] README header: needs the user's go
- [ ] Status page and stream page favicon (icon.svg): needs the user's go
- [ ] The bot's startup banner (ascii.txt): needs the user's go
