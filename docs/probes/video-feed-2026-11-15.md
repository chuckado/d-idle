# Probe: lossless delta feed on real d2d frames (card videofeed, 2026-11-15)

Tags: V = measured here (how is said), S = from another doc, U = not measured. Design: `docs/video-feed.md`.

## Method (V)

- Source: a lab d2d (container-pin build `lab3-labverbs` f75816c6, `--lockstep --no-save --seed 7`, fresh level 1
  sorceress `Vfeed`, Rogue Encampment, throwaway, nothing written to any save, no cheat writes) on the worker host (builder)
  (a locked builder job, nothing on the user desktop, not the sandbox hero). `tools/videofeed/record.py` steps ONE 40 ms world tick with
  `step 1 render` and takes a `screenshot` PNG after each, so consecutive files are consecutive 25 Hz frames (the rate of
  the sandbox feed, docs/viewer-fps.md section 6). The pinned build has no `framefeed` verb, so the PNGs (800x600 RGBA)
  stand in for the ring's RGBA slots: the same pixels, `readRing` gives the same layout.
- 630 frames, six scenes: town idle (100), then a click that walks the hero east, north, west, south (100 frames each,
  the whole view scrolls while the hero walks), inventory panel open (40), character panel open (40), town idle (50).
- Replay: `internal/watch/palfeed/replay_test.go` (`TestReplay`) encodes each frame against the previous one (a 25 fps
  client) or against the frame 2 or 3 earlier (a client taking every 2nd or 3rd frame), decodes it, and FAILS unless the
  decoded RGB equals the recorded frame (V: lossless on all 630 frames, all configurations). Run:
  `labrun <builder> -- sh tools/videofeed/run.sh` (record + replay + node + headless Chromium; `VF_REUSE=1` keeps the frames).
  Frames stay on the builder in /tmp and are never committed.
- Timing caveat: the builder is a shared container (6 CPU cap, load 2 to 5 while this ran), Go 1.27 and node 20 on one
  thread each. Read ms as a loaded-host figure, not a sandbox one. The sandbox has ONE core for d2d, d2dview and d2bot.

## Findings

1. d2d's frame is not palettised (V, fork source, design doc section 1) and the pictures confirm it: unique RGB colours
   per frame mean 340, max 656, and ALL 630 frames exceed 256. A per-frame 8-bit palette would be lossy. Alpha is 255 in
   every pixel (V), so the wire drops it.
2. A plain frame-to-frame XOR is great while the view is still and useless while it scrolls (V): walking north every
   delta is 484 KB at p50, as big as the 480 KB key frame, 6.6 MB/s at 25 fps. Walk east/west are cheap at p50 only
   because the hero reaches the click target early in the 100 frames; their p95 is 500 KB.
3. One global integer scroll per frame fixes it (V). With the scroll search, walk north p50 drops from 484 KB to 38 KB,
   and p95 over ALL frames from 521 KB to 55 KB (25 fps). The remaining outliers are single frames: opening a panel
   (inventory, character) moves the camera by half a screen at once (375 and 737 KB, about one key frame each) and is not
   a scroll within the search range. They are rare events, not a rate.

## Bytes per frame at 25 fps, tile 16 (the chosen tile), scroll search ON vs OFF

| scene | n | B p50 | B p95 | B max | MB/s mean | OFF: B p50 | OFF: B p95 | OFF: MB/s |
|---|---|---|---|---|---|---|---|---|
| town idle | 99 | 5712 | 9157 | 22175 | 0.15 | same | same | 0.15 |
| walk east | 100 | 10369 | 56860 | 62615 | 0.60 | 10369 | 501820 | 4.50 |
| walk north | 100 | 37735 | 52875 | 56567 | 0.71 | 483758 | 524751 | 6.63 |
| walk west | 100 | 13549 | 63095 | 84767 | 0.75 | 13549 | 524746 | 5.50 |
| walk south | 100 | 9936 | 12697 | 25856 | 0.25 | same | same | 0.25 |
| inventory open | 40 | 5077 | 5910 | 374161 | 0.36 | same | same | 0.36 |
| character panel | 40 | 4369 | 7126 | 737295 | 0.58 | same | same | 0.58 |
| ALL | 629 | 9716 | 54976 | 737295 | 0.49 | 9716 | 520693 | 2.80 |

Key frame 480 KB (deflate level 1 of the left-filtered RGB; the join cost, once per viewer and per 250 deltas).
References on the same frames (every 3rd frame): JPEG q72 (what `/stream.mjpg` sends today, lossy) p50 68.7 KB, p95 103 KB
= 1.7 to 2.6 MB/s at 25 fps; whole-frame XOR + deflate without tiles p50 17.2 KB, p95 522 KB.

So: still or slowly changing town, 5 to 10 KB per frame (8 to 14 times smaller than JPEG and exact); walking 10 to 63 KB
(1 to 7 times smaller than JPEG, exact); mean 0.49 MB/s (3.9 Mbit/s) over the whole recording against 1.7 MB/s for
MJPEG q72 (S: 1.9 MB/s measured in the sandbox, docs/viewer-fps.md section 6).

## Tile size and slow clients

| config (scroll search on) | ALL B p50 | ALL B p95 | ALL MB/s |
|---|---|---|---|
| tile 8, 25 fps | 9211 | 50641 | 0.45 |
| tile 16, 25 fps | 9716 | 54976 | 0.49 |
| tile 32, 25 fps | 10898 | 59876 | 0.53 |
| tile 16, 12.5 fps (every 2nd frame) | 12423 | 84767 | 0.37 |

Tile 8 is 6 percent smaller than 16 but has 4 times the tiles (bitmap and per-tile bookkeeping, not measured separately); 16 is the
default. A slower client (it skips frames) pays more per message but sends fewer: 0.37 MB/s at 12.5 fps.
`MaxShift` was 16 first: with it the 12.5 fps case had p95 509 KB (the scroll between two frames 2 apart is larger than
16 px); at 32 it is 85 KB, the price is a slower worst-case search (encode p95 5.7 to 11.6 ms; p50 stays 0.5 ms because the
previous shift is tried first).

## CPU (V on the loaded builder; Go, one thread)

| step | p50 | p95 |
|---|---|---|
| encode, tile 16, scroll search on | 0.51 ms (3.1 ms while walking steadily) | 11.3 ms (search when a walk starts or stops) |
| encode, scroll search off | 0.48 ms | 8.1 ms (the huge scrolled deltas deflating) |
| JPEG q72 encode, same frames (stdlib) | 8.1 ms | 8.8 ms |
| decode in Go (tests) | n/a | 1.8 ms |
| decode in node 20 (`inflateRawSync` + apply) | 0.64 ms | 3.3 ms |
| headless Chromium 156 (`DecompressionStream` + apply + `putImageData` of the dirty rect) | 1.2 ms | 4.4 ms |

d2d itself does nothing new: the ring is the existing `framefeed` (one memcpy). The encoder sits in d2dview where the JPEG
encode is today: median cost about 16 times lower than JPEG (0.5 against 8.1 ms), walking about 2.6 times lower, worst
single frames comparable (11 against 9 ms). On the sandbox's one core that moves d2dview from about 21 percent
(25 fps JPEG, docs/viewer-fps.md section 6) to a few percent, U until measured there.

## What this does not show (U)

- Only the Rogue Encampment, by day, one level-1 hero, no monsters, no spells, no dungeon darkness or light radius
  changes. Combat (missiles with additive blends, corpses, damage numbers, lighting that pulses) will cost more per
  frame; the first sandbox A/B should measure it with the real bot.
- Level changes, the death screen, and cutscene frames are key-frame sized by nature (a few per session).
- Real network and browser GPU path: headless Chromium here is software-rendered on a loaded host. Waterfox/Firefox is
  not installed on the builder, so Chromium only. The user's devices (dpr 1 to 3, phone) are not measured.
- The d2d widescreen framebuffer (up to 1400 px wide) is handled by the format (width in the header) but was not recorded.

## Walking-heavy recording and the honest bandwidth comparison (2026-11-16, user challenge: "walking is most of the game")

Tags: V = measured here. Setup: `tools/videofeed/run2.sh` on the builder (one locked job, 8 min; not the user desktop,
not the sandbox): `record2.py` steps a lockstep lab d2d (`lab3-labverbs` f75816c6, `--lab-verbs --no-save --seed 7`,
fresh throwaway level 1 Sorceress, 25 Hz, one PNG per 40 ms tick) through 2860 frames: town idle 100, a 20 s town walk
(20 clicks at random points of the encampment, 500 frames), `debug goto 2` (Blood Moor) + 60 idle, a 72 s field walk
(24 random clicks, 1800 frames), a fight (5 spawned Fallen, Fire Bolt with `debug cast 36 near` every 12 ticks, 300
frames), idle 100. The spawn, goto and skill points are lab SETUP writes on a throwaway hero: fine for bytes per frame,
never evidence about play. `TestCompare` (`internal/watch/palfeed/compare_test.go`) then encodes each frame every way
against the previous frame, nothing sampled. Bytes are per frame; MB/s = mean x 25. Methods: JPEG q72 (what `/stream.mjpg`
sends), q50, q35 (Go stdlib JPEG is already 4:2:0 chroma, so "2x subsampled chroma" is what these are); plain tile delta
(tile 16, no scroll search); tile delta + scroll search (the shipped codec); the same with colour truncated to 6 bits
(lossy); a hybrid. WebP is not stdlib: not measured (U).

| scene (frames) | JPEG q72 B / MB/s | JPEG q35 B / MB/s | delta plain B / MB/s | delta + scroll B mean / p50 / p95 / MB/s | 6-bit lossy delta+scroll MB/s |
|---|---|---|---|---|---|
| town idle (99) | 68 326 / 1.71 | 33 894 / 0.85 | 6 038 / 0.15 | 6 038 / 5 712 / 9 157 / 0.15 | 0.15 |
| town walk (500; scroll found in 472) | 69 974 / 1.75 | 35 027 / 0.88 | 489 234 / 12.23 | 49 579 / 51 372 / 62 310 / 1.24 | 1.22 |
| field walk (1800; scroll found in 644, the hero often stood between clicks) | 70 604 / 1.77 | 35 129 / 0.88 | 193 379 / 4.83 | 30 966 / 19 605 / 64 178 / 0.77 | 0.75 |
| fight (300) | 77 946 / 1.95 | 40 250 / 1.01 | 9 255 / 0.23 | 9 255 / 8 900 / 11 137 / 0.23 | 0.23 |
| field idle (160) | 64 048 to 79 230 / 1.6 to 2.0 | 30 883 to 41 085 / 0.77 to 1.03 | 9 184 to 20 021 / 0.23 to 0.50 | 9 184 to 20 021 / 0.23 to 0.50 | 0.23 to 0.48 |
| ALL (2859) | 71 349 / 1.78 | 35 725 / 0.89 | 209 232 / 5.23 | 30 088 / 18 948 / 63 552 / 0.75 | 0.73 |

Weighted mean MB/s for a share w of the time walking (town and field walk pooled by their frame counts), 10 percent fight
(none above 90 percent walking), the rest standing. The shares are not measured by this recording: S, soak analyses put
the walking share of KILL time at 68 to 77 percent (docs/d2d-soak-full-fix.md, docs/BACKLOG.md decomp walk entry) but
only 29 percent of a Blood Moor hour is walking once waiting for the next monster is counted (docs/roamspacing.md); the
table brackets both.

| method | w = 30% | w = 50% | w = 70% | w = 77% |
|---|---|---|---|---|
| JPEG q72 | 1.80 | 1.79 | 1.79 | 1.78 |
| JPEG q50 | 1.19 | 1.19 | 1.18 | 1.18 |
| JPEG q35 (lossy) | 0.90 | 0.90 | 0.89 | 0.89 |
| delta plain (no scroll) | 2.11 | 3.35 | 4.59 | 5.02 |
| delta + scroll (lossless) | 0.44 | 0.57 | 0.69 | 0.73 |
| delta + scroll, 6-bit colour (lossy) | 0.43 | 0.55 | 0.67 | 0.71 |
| hybrid, JPEG q72 when the delta is larger | 1.44 | 1.55 | 1.66 | 1.70 |

Reading it (V for the numbers, the interpretation is mine):

1. YES, it is a bandwidth saving while walking, but a smaller one than the first probe suggested. The worst measured case is
   the town walk: 49.6 KB per frame against 70.0 KB for JPEG q72 (1.24 against 1.75 MB/s, 29 percent less; its p95 62 KB is
   below JPEG's p50). Walking a field is 2.3 times smaller (0.77 against 1.77 MB/s) and everything that stands (town, a fight,
   idle, panels) is 8 to 11 times smaller. Weighted: 0.57 against 1.79 MB/s at half the time walking (3.1 times less), 0.73
   against 1.78 at 77 percent walking (2.4 times less), and always lossless. The reason walking costs more than it did in the
   first probe is that a town walk scrolls on almost every frame (472 of 500), and a scrolled frame carries every tile that
   is not exactly a shifted copy (sprites, the lit ground), about 50 KB.
2. JPEG costs the same whether the picture moves or not (64 to 79 KB per frame), so its MB/s is flat across the table; the feed's
   MB/s rises with walking. In this data the feed is below JPEG q72 in every scene. Cheaper LOSSY JPEGs beat it only in
   a town walk (q50 1.15 and q35 0.88 against 1.24 MB/s); in a field walk, a fight and every still scene the lossless feed is
   smaller than even q35. Truncating colour to 6 bits saves 2 to 3 percent (d2d's frames carry few distinct colours per
   tile already) and is not worth a lossy mode.
3. Plain tile delta without the scroll search is NOT usable: 12.2 MB/s in a town walk, 7 times the JPEG. The scroll search is
   what makes the idea work (as the first probe said); `MaxShift` 32 stays.
4. The hybrid does not work as a per-frame choice. A JPEG frame leaves the client's lossless base behind, so the next delta
   is built from an older base and is larger than a JPEG almost every time: the simulation (the client keeps its last lossless
   frame, the delta is computed against it) stays on JPEG on every walking frame (its field-walk mean equals the JPEG's to the byte) and ends at 1.55 MB/s, 2.7 times WORSE than the pure
   feed at w = 50 percent. Nothing is gained by a JPEG fallback because the delta is not larger than the JPEG on any scene
   mean or p95 here; the exceptions are single frames (a panel opening is key-sized, 375 to 737 KB, see the earlier table),
   which the server paces instead (`Options.BigMsg`, `BigGap`, `MaxBytesPerSec`).
5. Where it wins: CPU on d2dview (encode 0.5 ms p50 and about 3 ms while walking against 8 ms for the JPEG, which the feed
   also skips entirely while only canvas viewers are connected), lossless pixels (no JPEG ringing around text), static
   scenes (5 to 10 KB), fps headroom (the codec is not the limit at 60 Hz; the bytes of a walking frame are). Where it does not
   win by much: a town walk (1.4 times smaller than JPEG q72, and a lower quality JPEG is smaller still).

Verdict: ship behind the knob (off by default), keep MJPEG as the fallback; do not build the hybrid. The bandwidth case is
real but modest while walking in town and strong everywhere else; the CPU and lossless cases do not depend on walking.
The A/B on the sandbox (docs/video-feed.md section 10) decides whether the default turns on.

## Canvas upload: putImageData against a WebGL texture (V, headless Chromium 156 on the builder, SwiftShader GL, 100 runs)

`tools/videofeed/bench2.sh`: `putImageData` of the full 800x600 frame p50 0.2 ms, p95 1.0 ms; WebGL `texSubImage2D` + draw +
`finish` p50 0.4 ms, p95 2.1 ms; a 200x150 or 16x16 dirty rectangle is below the timer's 0.1 ms for both. The 2D path is as
fast or faster in software GL, so the player keeps `putImageData` of the dirty rectangle (no WebGL context to lose, no
shader). A real GPU path and Waterfox/Firefox are NOT measured (U: Firefox is not installed on the builder; the player
uses only `DecompressionStream("deflate-raw")` (Firefox 113+, Waterfox follows), `ImageData`, `putImageData` and
`fetch` streaming, all present there; a browser without `DecompressionStream` stays on MJPEG, logged to the console).
