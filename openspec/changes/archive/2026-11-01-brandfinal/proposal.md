## Why

The user decided on 2026-11-01 ("brand: cursor") that the cursor concept (prompt chevron + block cursor, terminal ember palette) is the d.idle brand. The concept round (card `brand`) left SVG drafts only; the final asset set, PNG renders and the banner are still missing.

## What Changes

- assets/branding/final/: the cursor set as SVG masters (logo, logo-light, mark, mark-mono, wordmark, icon, header, header-discord), PNG renders (icon 16..512, favicon.ico, logo, logo-light, wordmark, header 1280x640, header-discord 960x540), scenes.json (the rect lists both outputs come from), and the banner files (big, plain ASCII, one-line).
- assets/branding/final_gen.py (called by gen.py) and assets/branding/render.py (stdlib PNG/ICO writer and decoder, no system rasterizer).
- Geometry fixes found on the re-check: the discord header is laid out natively on whole pixels (the old 0.75 scale gave a 4.5 px tagline cell), the icon is hinted per size (whole-pixel cells from 16 to 512 px), the headers' lockup is vertically centred.
- docs/branding.md: the decision, asset list, regeneration, usage rules, licence, application checklist.

## Relation to card `brand`

`brand` (branch agent/brand, tip 4cbe0d8) produced the concepts and stays in-progress there. This card starts from that tip, so it overlaps its `owns:` on purpose: brandfinal continues the same files and must be merged after `brand` (or instead of it: brandfinal contains every brand commit). The coordinator archives both together. The other two concepts stay in assets/branding/concepts/ as archive.

## Not done here

Nothing is applied to GitHub, Discord, the README on main, the web pages or the bot: that is the separate branch `agent/brandapply` (card `brandapply`), awaiting the user's go.
