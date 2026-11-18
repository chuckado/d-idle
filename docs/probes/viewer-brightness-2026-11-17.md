# Probe: why the watch viewer looks "a bit too bright" (card brightness, 2026-11-17)

Tags: V = measured or read here (how is said), S = from another doc, U = not measured. Offline only: no game, nothing on
the sandbox, no live instance. Frames are three recorded 800x600 town frames from the video-feed probe
(`docs/probes/video-feed-2026-11-15.md`: f0050 idle, f0300 and f0450 walking), never committed.

## Verdict

1. The viewer is NOT brighter than d2d's own picture (V, numbers below). Lossless palette feed: luma identical to the
   frame; JPEG q72 (d2dview): mean luma 0.3 levels LOWER (about 1.2 %), p50 and p95 within 1 level. No CSS filter,
   overlay or blend touches the canvas or the `<img>`; the canvas is a plain sRGB 2D context.
2. The brightness is made in d2d, before the frame leaves the game: d2d.cfg `gamma` defaults to 155 and `contrast` to 100,
   and at contrast 100 the ramp is a straight gain, level = min(255, i x 155 / 100) (V: fork `palette.hpp` `make_ramp`,
   asserted by the fork's `test_gamma`; `docs/research/re/video-drivers.md` in the fork, read from game.exe `FUN_006be190`).
   Everything on screen is 1.55 times the file colour; 100 would leave the palette alone.
3. That 1.55 is what game.exe applies under DirectDraw (its default fullscreen driver); under GDI, which is what `-w`
   (the windowed game the bot runs and every live check is compared with) uses, the slot is a bare `ret`: no gamma (S, same
   fork doc: "GDI does nothing", read from the code, not compared on screen). So d2d's default is faithful to a default
   fullscreen D2 and about 35 % darker-looking "flat" in the windowed game: a viewer who remembers the windowed picture
   sees the d2d picture as brighter.
4. The sandbox ships no `gamma` or `contrast` key (V: grep of the repo's deploy, packaging, tools and docs: none), so it
   runs the 155 default.

## Numbers (V, `go run ./tools/videofeed/lumastat FRAME.png...`)

BT.601 luma of the frame as each stage delivers it (0..255):

| frame | stage | mean | p50 | p95 |
|---|---|---|---|---|
| f0050 | reference (d2d) | 25.03 | 20 | 55 |
| f0050 | lossless palette feed | 25.03 | 20 | 55 |
| f0050 | JPEG q72 | 24.72 | 20 | 55 |
| f0300 | reference / palfeed / JPEG | 26.78 / 26.78 / 26.47 | 22 / 22 / 21 | 60 / 60 / 59 |
| f0450 | reference / palfeed / JPEG | 25.70 / 25.70 / 25.36 | 20 / 20 / 20 | 59 / 58 / 58 |

Estimated gamma 100 (U: each channel divided by 1.55; exact when no channel clipped, and 0.00 % of the channels in these
frames are 255, so the division is invertible up to rounding): mean 16.2 / 17.3 / 16.6, p50 13 / 14 / 13, p95 36 / 39 / 38.
So gamma 155 against 100 is a luma ratio of 1.55 (mean, p50 and p95 all about 1.55 times), far above the viewer's 0.99.
The frames are dark either way (mean 25 of 255): the "washed" impression is lift in the shadows, not clipped highlights.
U: that the pinned build that recorded these frames used 155 (the fork source says so; the build's cfg was not read). The
ready-to-run check is `tools/videofeed/gamma_frames.py` (below).

## The viewer pipeline (V, code read)

- Palette feed: lossless RGB tiles (`internal/watch/palfeed`), `putImageData` on `getContext('2d', {alpha:false})` (no
  `colorSpace` given, so sRGB; a wide-gamut screen colour-manages an sRGB canvas, it does not stretch it). No filter.
- MJPEG: Go `image/jpeg` writes a full-range JFIF (BT.601 YCbCr, no limited-range yuvj ambiguity), decoded by the
  browser as sRGB. The 0.3 level mean drop above is quantisation and 4:2:0, not a range error. (ffmpeg x11grab is not in
  the sandbox path; d2dview encodes from the frame ring.)
- CSS: `grep` of `web/` and `internal/watch/web/` for `filter:`, `mix-blend-mode`, `opacity` over the picture: only the
  quest caption and badge (small, outside the picture) and the primary-button hover; `image-rendering: pixelated` does
  not change values. `color-scheme` is page chrome only.

## What shipped (card brightness)

- A viewer-side Brightness button in the video card (`internal/watch/web/panels/video/brightness.js`, additive, one entry
  in the card's actions): Darkest 70 / Dark 85 / Default 100 / Bright 115 / Brightest 130 percent, per browser
  (`ctx.store`), default 100 and then no filter at all. It sets `--vid-bright` on the frame; `panel.css` applies
  `filter: brightness()` to the picture only when the level is not 100. It never touches the game's setting. Note 70 %
  approximates undoing the 155 gain on an unclipped frame (1 / 1.55 = 65 %); 100 -> 65 would be the windowed-game look.
- Regression: `TestViewerPathsKeepLuminance` (palfeed: exact lossless luma, JPEG q72 within 1.0 level) and the node panel
  test for the control (`testdata/video_panel.js`).

## Decisions for the user

- Sandbox look: ship `gamma = 100` in the sandbox/friends d2d.cfg (looks like the windowed game, about 35 % less
  light) or keep 155 (faithful to default fullscreen D2)? Not changed here. A friend can already tune it in the game's
  own Video Options if the sandbox exposes that menu; the viewer control above works either way.
- Fidelity: the ramp itself matches `FUN_006be190` (run through game.exe's own code in the fork). NOT shown: an
  on-screen differential of a full frame against game.exe, and double application on UI sprites that go through
  `Palette::operator[]` (U). No d2d deviation is claimed; no fork branch was made.

## Re-check (ready to run, builder)

`labrun <builder> -- sh -c 'python3 tools/videofeed/gamma_frames.py /tmp/brightness-fr "$D2D" /gamedata 155 100 120 && go run ./tools/videofeed/lumastat /tmp/brightness-fr/*.png'`
records the same seeded town frame at gamma 155, 100 and 120 (d2d.cfg `gamma` via `D2D_USER_DIR`) and prints luma stats;
expect mean(155) / mean(100) about 1.55 and the frames otherwise identical. Not run: the builder's main queue held six jobs.
