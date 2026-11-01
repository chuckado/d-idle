# d.idle branding

> status: DECIDED. The user chose the cursor concept on 2026-11-01 ("brand: cursor"). The final set is in `assets/branding/final/` (cards `brand`, `brandfinal`). Nothing is applied to any public surface yet: see the checklist. The ready-to-apply in-repo set is on branch `agent/brandapply` (card `brandapply`), not merged.

The name is always written with a lowercase d: **d.idle**, never "D.idle". The repo is `d2lod-bot`, the public repo `d-idle`.

## Decision

Cursor: a prompt chevron and a block cursor in a terminal ember palette. The bot sits at the prompt for days, waiting. The other two concepts (`rune`, `tick`) and the first-round cursor drafts stay in `assets/branding/concepts/` as archive; they are not used. Palette and contrast ratios: `assets/branding/final/` colours listed below and `assets/branding/concepts/cursor/palette.md`.

## Final asset set (`assets/branding/final/`)

SVG masters (all shapes, own pixel font, no installed fonts):

| File | Size | Use |
|---|---|---|
| `logo.svg` | 468x108 | chevron + cursor + wordmark, transparent; for dark and light surfaces (P `#e0630f`, Q `#a8700a`, both >= 3:1) |
| `logo-light.svg` | 468x108 | darker colours (`#b34700`, `#7d5200`, 4.9 and 6.1:1 on paper) for light surfaces and print |
| `mark.svg` | 132x84 | chevron + cursor only |
| `mark-mono.svg` | 132x84 | single colour (coal `#0e0c0a`) for stamps, favicons in one ink, embossing; recolour by editing the one fill |
| `wordmark.svg` | 108x28 | wordmark + block cursor, 4 px cells: tight spaces (badges, nav bars) |
| `icon.svg` | 512x512 | rounded tile (rx 104) for avatars and favicons; dark surfaces |
| `header.svg` | 1280x640 | GitHub social preview, README header |
| `header-discord.svg` | 960x540 | Discord banner |

PNG renders (all under 300 KB; the biggest is 6.3 KB): `icon-512/256/128/64/48/32/16.png` (hinted per size: whole-pixel cells, 1 px cells at 16 px), `favicon.ico` (16, 32, 48; PNG-compressed entries), `logo.png` and `logo-light.png` (936x216, 2x), `wordmark.png` (108x28, 1x, scale it by whole numbers with `image-rendering: pixelated`), `header-1280x640.png`, `header-discord-960x540.png`.

Text: `banner.txt` (half-block art, 52 columns, plus the tagline), `banner-ascii.txt` (plain ASCII `#` blocks, for terminals and logs without block characters), `banner-line.txt` (one-line variants: `> d.idle_` and `d.idle: unattended. offline. idle.`; both plain ASCII). `scenes.json` is the rect list every SVG and PNG is made from.

Colours: coal `#0e0c0a` (dark), paper `#f6f1e7` (light), ember `#ff8a1f` and amber `#ffc15a` (header and icon accents, dark only), ash `#ece2d4` (header wordmark, 15.2:1 on coal), muted `#b4a797` (tagline, 8.3:1), icon tile `#15100c` (ember 8.0:1, amber 11.7:1). The generator prints the full contrast table and fails below 3:1.

## Regeneration

```
python3 assets/branding/gen.py      # concepts + final/*.svg, scenes.json, banners (stdlib only)
python3 assets/branding/render.py   # final/*.png and favicon.ico, then decodes every PNG and checks pixel colours
python3 assets/branding/render.py --preview   # ASCII occupancy of the icon at 16, 32, 48, 64 px
```

`render.py` is a small stdlib rasterizer for the project's own subset (axis-aligned rects with opacity, scanlines, one rounded background; exact area coverage, 4x4 supersampling on the corners). It is not a general SVG renderer: new shapes go into `final_gen.py` as rects first.

## Usage rules

- Lowercase d, always: `d.idle`. The dot is part of the name and keeps its accent colour.
- Clear space: at least one wordmark pixel cell (the logo files include it) on every side; for the icon, nothing within the rounded corner.
- Minimum size: logo 120 px wide, wordmark 108 px, mark and icon 16 px (use the icon, not the header, below 64 px). At 16 and 32 px the chevron and the block stay separate shapes (checked by the ASCII preview). Do not scale pixel art to non-integer sizes; in browsers use `image-rendering: pixelated` for the PNGs.
- Do not: recolour outside the palette, stretch, add shadows or outlines, set the name in another font, put header or icon colours on a light background (use `logo-light`), or combine concepts.
- Everything is drawn as shapes by our own pixel font; no installed font, third-party image, Blizzard or Diablo II logo, font, art, sprite, icon or screenshot is used or may be mixed in.
- Licence: original work of the project; the licence follows the repository licence, the GNU General Public License v3 (see `LICENSE`).

## Application checklist (each needs the user's go)

Prepared in-repo on branch `agent/brandapply` (not merged; card `brandapply`): README header image, favicon links in the public web pages of this repo, the banner printed by `d2bot --version` and `-h`.

- [ ] Merge `brandfinal`, then `brandapply`: needs the user's go
- [ ] GitHub social preview (`header-1280x640.png`): needs the user's go
- [ ] GitHub avatar (`icon-512.png`): needs the user's go
- [ ] Discord server icon and banner (`icon-512.png`, `header-discord-960x540.png`): needs the user's go
- [ ] README header on main: needs the user's go
- [ ] Favicons on the status page and pages repo (private ops repo, not this one): needs the user's go
- [ ] The bot's startup banner on main: needs the user's go
