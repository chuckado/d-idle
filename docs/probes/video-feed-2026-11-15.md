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
