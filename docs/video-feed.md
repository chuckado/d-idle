# Lossless delta frame feed for the watch video (card videofeed, 2026-11-15)

Tags: V = verified (read in source or run, how is said), S = secondary (read in docs), U = unverified.
Numbers: `docs/probes/video-feed-2026-11-15.md` (real d2d frames, section 5 below); this page is the design and the status.
> status: mounted behind the knob `watch.paletteFeed` (default OFF) by card videofeed2 (2026-11-16); not deployed. Sections
> 9 and 10 below are the current state; the bandwidth verdict is in docs/probes/video-feed-2026-11-15.md (walking-heavy
> recording).

User go 2026-11-15: design and prototype a better feed than ffmpeg x11grab / the container tap. User go 2026-11-15 "keep
working till everything is fully implemented": videofeed2 mounts it (section 9). Nothing is deployed.

## 1. The frame source: what d2d really holds (the premise needs a correction)

The idea was that d2d keeps one 800x600 8-bit palettised frame. It does not. V (read, fork branch `d2d.lab3-framefeed`,
`apps/d2d/main.cpp`, `apps/d2d/common.hpp`, `apps/d2d/world_view.hpp`, `apps/d2d/tile_blit.hpp`):

- The framebuffer is `std::vector<std::uint8_t> framebuffer`, RGBA8, `fb_width(framebuffer) * kScreenHeight * 4` bytes
  (`fb_width` = size / 4 / 600, so 800 by default and up to `kMaxViewWidth` 1400 in widescreen; the height is 600).
  It is uploaded as-is with `SDL_UpdateTexture(win.texture, ..., framebuffer.data(), fb_width * 4)`.
- Sprites and tiles are palette indexed ONLY at blit time: `blit_sprite(framebuffer, frame, pal, x, y)`,
  `blit_dcc_frame(..., pal, ..., trans)`, `blit_dt1_tile_lit(...)`. Every call converts index to RGBA through a palette
  chosen per call: the act palette, a per-light-level palette (`Lighting::palette(x, y)` = `(*pal)[level >> 3]`, one per
  subtile), tinted palettes (co-op cursor), and draw modes 3 and 4 (missiles, translucent) that blend with sums in RGB
  ("the same sums in RGB, not the tables' nearest palette colours", `world_view.hpp`). So one frame is drawn with many
  palettes and some RGB arithmetic: there is no single index buffer plus one palette to stream, and no 256-colour bound.
- The exports that exist (V, docs/viewer-fps.md section 2, fork commit `80cc7adc`): verb `screenshot <abs path>` (PNG,
  encoded on the game thread, about 57 ms) and verb `framefeed <abs path> <every> <lease ms>` (every drawn frame as
  RGBA8 into a two-slot shared file on tmpfs, one memcpy on the game thread; header and slots in `components/screenshot/
  framefeed.hpp`; `main.cpp` offers the frame right after the draw, skipped in lockstep `skip_draw`). The sandbox runs it
  (docs/viewer-fps.md section 6: 21.6 fps delivered, d2d CPU down). `cmd/d2dview` reads the ring (`readRing`, `ring.go`).

Decision: NO new d2d verb or patch. The ring already is the lossless source of truth (RGBA8, 60 Hz draw, sampled at the
feed's `every`). An indexed framebuffer would mean changing the renderer (a parallel index plane that cannot represent
lit, tinted and blended pixels), which is large, risky for fidelity and buys nothing the delta coder does not (see the
probe: real frames hold thousands of distinct colours, so a per-frame palette is not applicable). If the numbers ever
demand it, the wire format below has a reserved flag bit for an indexed payload; it is not implemented.

Where the encoder runs: in the process that already holds the ring reader and the viewer sockets, `cmd/d2dview` (the
tap), NOT in d2d: zero extra work on the game thread, and a lease-less feed costs nothing when nobody watches (the ring
lease stops after 3 s without a viewer, V by viewerfps design).

## 2. Wire format (version 1; code `internal/watch/palfeed/codec.go`, V by tests)

One message per frame, little endian, then a payload of raw DEFLATE (RFC 1951, Go `compress/flate`, level 1):

```
0  "PF"  2 u8 version=1  3 u8 flags (bit0 key frame; other bits reserved: bit1 indexed payload, bit2 palette change)
4  u16 width  6 u16 height  8 u8 tile edge (16)  9 u8 0  10 u32 seq  14 u32 base seq (0 for a key frame)
18 u32 game ms  22 u32 payload length  26 i8 dx  27 i8 dy (scroll found by the encoder)  28 payload
payload (inflated)
  key frame : every tile's RGB bytes, row-major tiles; each row left-neighbour filtered (byte minus the byte 3 to the left)
  delta     : 2 bits per tile (row-major, 4 tiles per byte): 0 same as the base, 1 base XOR data, 2 shifted base XOR data,
              3 exactly the shifted base; then the RGB bytes of every mode 1 and 2 tile. "Shifted base" = the base pixel at
              (x-dx, y-dy), black outside the frame.
```

Why the scroll modes (V, measured on real frames, probe doc): while the hero walks the whole isometric view moves a few
pixels per frame, so a plain frame-to-frame XOR changes nearly every pixel (about 480 to 540 KB per delta, as large as a
key frame). The encoder looks for one global integer shift within +-32 px (`EstimateShift`: a sparse sample lattice, the
previous frame's shift first, then a coarse search refined on a finer lattice) and codes each tile against whichever of
"same place" or "shifted" leaves fewer differing pixels; the HUD, which does not scroll, stays on mode 0 or 1.

- "Palette change flag": not needed for RGB; the reserved bits describe the indexed variant (a palette of N RGB triples
  ahead of the bitmap, tile bytes become indices) for a future source that really has indices. Not implemented (U).
- Alpha is dropped on the wire (d2d draws opaque: every recorded frame had alpha 255, V in the probe).
- Delta base: the encoder is given the frame THE CLIENT HOLDS, so the server needs no shared reference chain. A client
  that is slow gets, at its next write, a delta from what it has to the newest frame (frames in between never exist for
  it): "latest frame wins" without a queue (`Hub.Stream`). A key frame goes first on every connection, and again every
  `KeyEvery` (250) deltas as resync insurance. TCP is reliable and ordered, so a delta never misses its base; a decoder
  that nevertheless sees the wrong base throws `needKey`/`ErrNeedKey` and the player reconnects (new key frame).
- Join and request: connect = key frame; a key on demand is a reconnect (no extra verb, nothing to authenticate twice).
- Why tiles and not row runs: unchanged tiles cost one bit and no deflate input; the dirty-rect also lets the client call
  `putImageData` on the changed bounding box only. Tile edge is in the header (8, 16 or 32 all decode); the probe picks 16.
- Why DEFLATE only: the Go stdlib has it, browsers decode it natively (`DecompressionStream("deflate-raw")`, no wasm,
  no library), node has `zlib.inflateRawSync` for the tests. zstd would need a module (not stdlib) and a wasm in the page.

## 3. Transport and auth

- `GET /stream.pfeed[?fps=N]` (`Hub.Stream`): chunked body of `u32 length` + message pairs, `Content-Type:
  application/x-palfeed`, `Cache-Control: no-store`. A plain HTTP stream read with `fetch()` and `ReadableStream`; no
  WebSocket needed (one direction; the page controls the rate by reconnecting with `fps`). A WebSocket stays optional for
  a client that wants to send keyframe requests without reconnecting (U, not built).
- `GET /api/palfeed` (`Hub.Capability`): `{"paletteFeed":true,"version":1,"width":W,"height":H}`, 503 before the first
  frame, 401 without auth. The page calls it once per load (`PalFeed.detect`).
- Auth: both handlers take the host's own check as `Options.Auth` (nil refuses everything): the tap's bearer token
  (`d2dview -token-file`, same as `/stream.mjpg`) and, on the watch viewer (8891), the watch session/token check that
  guards `/stream`. Nothing new to configure; no unauthenticated default (V by `TestStreamLatestFrameWinsAndAuth`,
  `TestCapability`). A per-write deadline as in `watch.Video` (`WriteTimeout`) must wrap the writer when mounted, so a
  viewer that stopped reading frees its slot (the prototype's `Stream` blocks in `Write` like any handler).
- Wiring (not done, about 25 lines each): `cmd/d2dview`: after `readRing` call `hub.Publish(Frame{...}, h.Stride)`
  while a feed client is connected (the Hub has no client count yet: add one before mounting, to stop the ring lease when
  idle); mount both routes behind the existing token middleware. `internal/watch`: `Video` today relays the tap's
  `/stream.mjpg` through `TapCapture`; add a second relay of `/stream.pfeed` (bytes copied, not decoded) and mount
  `/stream.pfeed` + `/api/palfeed` on the viewer behind the watch auth, with `SetFrameURL` unchanged for MJPEG.

## 4. Client player (`internal/watch/web/palfeed.js`; node tests `tools/videofeed/palfeed.test.js`, V 7/7)

- `PalFeed.detect(fetch, base, headers)`: asks `/api/palfeed`; anything but 200 + `paletteFeed:true` + `version:1`
  returns null and the page keeps its MJPEG `<img>` (fallback, unchanged). After 3 consecutive stream failures
  `Player` calls `onFallback(reason)` and the page switches back to the `<img>`.
- `Player`: `fetch` stream, `Splitter` (any chunking), messages applied strictly in order (async inflate via
  `DecompressionStream("deflate-raw")`), the decoder writes into one `Uint8ClampedArray` that backs an `ImageData`;
  the draw happens ONCE per `requestAnimationFrame` with the union of the dirty rects since the last draw
  (`putImageData` dirty rect): two messages in one display frame cost one draw (latest wins, tested).
- Canvas: intrinsic size = frame size (800x600), CSS `image-rendering: pixelated`; `PalFeed.fit(frameW, frameH, availW,
  availH, devicePixelRatio)` chooses an INTEGER number of device pixels per frame pixel (CSS size = frame * s / dpr) so
  pixels stay square and sharp at dpr 1, 1.5, 2, 3; below 1x it falls back to the exact fit.
- Low rate toggle: `Player.setLowRate(true)` reconnects with `?fps=10`; the server applies `min(fps, MaxFPS)` to that client only.
- GPU palette lookup is not needed: the stream is RGB, `putImageData` is the whole draw cost (probe numbers). A WebGL
  texture upload would only matter at dpr 3 full-screen scaling; U.

## 5. CPU, bytes, fps (V, 630 real d2d frames, docs/probes/video-feed-2026-11-15.md)

- Lossless on every frame and every configuration (the replay fails otherwise). Real frames hold 340 colours on average
  (max 656, all over 256): a per-frame 8-bit palette would be lossy, so the wire is RGB.
- Bytes per frame at 25 fps, tile 16: still town 5.7 KB p50; walking 10 to 38 KB p50, 53 to 63 KB p95; whole recording
  p50 9.7 KB, p95 55 KB, mean 0.49 MB/s (3.9 Mbit/s). JPEG q72 of the same frames: 69 KB p50, 103 KB p95, about 1.7 MB/s.
  Without the scroll modes walking costs 484 KB per frame (6.6 MB/s): they are what makes the idea work.
- Key frame 480 KB; panel open or a level change is one key-sized delta (375 to 737 KB), rare.
- CPU: encode 0.5 ms p50 (3 ms while walking, 11 ms worst) against 8.1 ms for the JPEG it would replace; decode 1.2 ms p50,
  4.4 ms p95 in headless Chromium, 0.2 ms to put the dirty rect. d2d: nothing new (the ring's memcpy).
- fps: the codec is not the limit at 25 fps (the budget per frame is 40 ms); 60 fps is possible on the encode side
  (the ring publishes at the draw rate) but the sandbox tap is configured for 25, and 60 would double the bytes.
  A slow client skips frames and costs fewer bytes per second (0.37 MB/s at 12.5 fps).
- Sandbox context (V, docs/viewer-fps.md section 6): one CPU for the whole container, JPEG q72 about 10 ms and 77 to 90 KB
  per frame, 21.6 fps delivered at 1.9 MB/s.

## 6. Security

Same token as the existing streams; no new listener; the feed carries only what the picture already shows; the decoder
caps a message at 16 MiB and inflates at most `w*h*3 + w*h/8 + 1024` bytes (Go) / checks lengths before reading (JS)
so a hostile or broken stream cannot make the page allocate unboundedly. The ring file stays mode 0600 on tmpfs (unchanged).

## 7. Friends pack impact

PLAINLY: the friends image (pin `lab3-labverbs` f75816c6, `deploy/d2d-container/context.sh`) does NOT have the `framefeed`
verb (V: the probe recorded with that exact build had to use `screenshot` PNGs), so with the current friends image the feed
cannot run: `D2_PALETTE_FEED=1` there logs "needs -feed" or finds no ring and the page keeps MJPEG. A friends release needs
a newer d2d pin that contains the lineage of fork commit 80cc7adc (framefeed + nocursor) AND the user's go to move the pin;
then `D2_VIEW_FLAGS="-fps 25 -feed /dev/shm/d2dview/frames.ring"` (+ `shm_size: 128m`) and `D2_PALETTE_FEED=1` in `.env` turn it on
(docs/friend-guide.md).

(Written before the mount) None until wired. When wired: no new binary, no new dependency, no new port; `cmd/d2dview` and the watch viewer grow by
the package (about 400 lines of Go) and one 200-line JS file in the embedded web tree; the d2d image is unchanged (it
already carries `framefeed`, the pinned container ref predates it: the friends container needs the `d2d.lab3-framefeed`
build that the sandbox runs, or the feed falls back to MJPEG from screenshots, which cannot feed this codec at full rate).
The sandbox's `-feed` ring is the prerequisite; a friends release without it keeps MJPEG (the capability check says so).

## 8. Status and next steps

Prototyped (V, tests): codec both ways, hub, stream and capability handlers, auth, latest-frame-wins, JS decoder
byte-compatible with Go (shared fixtures), splitter, fit, detect, player draw pacing; replay and measurement tooling
(`tools/videofeed/run.sh`). Only designed: mounting in d2dview and the watch viewer, ring-client counting, per-write
deadline, the `watch.paletteFeed` knob and page switch, indexed payload.
(Done by videofeed2: item (1) below and the client polish; see section 9. The A/B, item (2), is section 10.)
Next: (1) wire behind the flag in d2dview + `watch.Video` relay (one card, `cmd/d2dview` and `internal/watch` owners);
(2) sandbox A/B against MJPEG with the coordinator's `ab.sh`: fps, d2dview CPU, bytes, bot pace; (3) a real-browser pass on
the user's devices (dpr 1 to 3).

## 9. Mounted (card videofeed2, 2026-11-16; offline tests only, nothing deployed)

Turn it on (default OFF, nothing changes without it): the environment switch `D2_PALETTE_FEED=1` (friends pack: a line in
`.env`, passed through `tools/friendpack/compose.yml`) or the knob `watch.paletteFeed` = 1 (docs/knobs.md; `env.PaletteFeedOn`
is the one rule: `D2_PALETTE_FEED=0` forces off, `1` forces on, else the knob). d2dview also has `-palette-feed` (default from that
rule). It needs the framefeed ring (`-feed ...`, `-fps > 0`): without it d2dview logs one line and the feed stays off.

- d2dview: `Frames.Pal` receives every ring frame (`palfeed.Hub.Publish`); `/stream.pfeed` and `/api/palfeed` are mounted next
  to `/stream.mjpg` behind the same token middleware (`server.auth`); `Frames.TouchPal` (the hub's `Options.Touch`, called on
  connect, on a probe and about once a second) keeps the ring lease and the sampler alive, so the feed alone drives d2d's
  ring. The JPEG is encoded only while an MJPEG viewer touched the frames in the last 5 s (`jpegWanted`): a canvas-only
  viewer saves the 8 ms encode.
- Watch screen: `Hub.SetPalFeed` (internal/watch/palfeed.go) makes `/api/config` carry `palFeedUrl` and `/api/glance` list
  `stream.formats` = `["mjpeg","palfeed"]` (absent feed: no `formats`, the page stays MJPEG). In d2dview the routes are
  mounted by d2dview itself; for the d2bot watch viewer `watch.PalFeedRelay(tapURL, token)` is a reverse proxy to the tap
  (bytes copied, not decoded; the viewer's token and cookie are NOT forwarded, the tap gets its own bearer, only `fps` passes), wired in
  `cmd/d2bot/ladder.go` (`palFeed`) when `--watch-video-source tap` and the switch is on.
- Page (`web/panels/video/panel.js`): asks `/api/glance` `stream.formats`; with `palfeed` listed it loads `palfeed.js`, probes
  `/api/palfeed` (503 until the first frame, so up to 5 tries a second apart; the probe itself wakes the ring) and then
  uses a `<canvas>` player; the controller (backoff, 3 s stall watchdog, hidden-tab pause, Retry, the fps and KB/s stat,
  Low rate which adds `fps=8` to the feed's URL too) is the same one MJPEG uses. A fatal stream error (bad header,
  short payload), 3 failures in a row with no frame between, a browser without `DecompressionStream` or `ImageData`, or a
  glance without `palfeed` all leave or return to MJPEG, logged with `console.info` only (no toast). A need-key error (a
  delta without its base) is an ordinary failure: reconnect = key frame.
- Client polish: integer device-pixel scaling with `devicePixelRatio` (`PalFeed.fit`, applied by a `ResizeObserver`;
  without it the CSS `object-fit: contain` fits), `image-rendering: pixelated`, one draw per `requestAnimationFrame` with the
  union of the dirty rectangles (latest frame wins), one `ImageData` per decoder buffer so a delta stays a dirty rectangle,
  `desynchronized: true` context. Rejoin after a hiccup = reconnect = key frame. `putImageData` was kept over a WebGL texture
  upload (probe doc: 0.2 ms against 0.4 ms for a full frame in software GL; Firefox not measured, U).
- Server guard (`palfeed.Options`): per client `MaxBytesPerSec` 2 MiB/s (after a message of n bytes the next waits n/cap, so
  a heavy stretch lowers that client's rate and the next delta is built from the newest frame), key-sized messages
  (`BigMsg` 200 KB) are followed by at least `BigGap` 400 ms (a panel flickering open and shut cannot send 700 KB 25 times a
  second), a per-write deadline (`WriteTimeout` 10 s) drops a client that stopped reading, a heartbeat (an empty delta of a few
  dozen bytes) after 1 s without a new frame keeps the page's stall watchdog quiet in a paused game.
  `Hub.Stats()` has the counters (messages, bytes, big messages, guard pauses); not yet on `/api/glance` (U).
- Not built: a hybrid JPEG fallback per frame (it makes things worse, probe doc), the indexed payload, a WebSocket.

## 10. Live A/B checklist for the sandbox (NOT run; do not deploy from this card)

Prerequisites: a sandbox d2d with the `framefeed` verb (fork 80cc7adc, already there) and d2dview with `-fps 25 -feed
/dev/shm/d2dview/frames.ring` (docs/viewer-fps.md section 6). Friends images do not qualify (section 7).

1. Baseline (switch off): note `/api/glance` `stream` (fpsOut, kBps, viewers), d2dview CPU (`top` on the container) and the
   page's card head with `?debug` for 2 min while the hero walks and 2 min while it fights. Screenshot the card.
2. `D2_PALETTE_FEED=1` on d2dview (restart the tap only), reload the page: expect d2dview's log line
   `lossless picture feed on /stream.pfeed`, `GET /api/glance` `stream.formats` = `["mjpeg","palfeed"]`, the card
   showing a canvas (devtools: `canvas.shot`), `console.info` quiet. Check `kBps` in the card head against the baseline
   (expect 0.4 to 0.8 MB/s walking, 0.15 to 0.25 standing) and the fps (expect 25 or the ring's rate).
3. d2dview CPU while one canvas viewer only (expect lower than baseline: no JPEG), then with an MJPEG viewer added
   (`curl -N .../stream.mjpg?t=...`): both work, CPU about the baseline plus the feed's 0.5 ms a frame.
4. Bot pace: compare ticks per wall second and kills per hour of the same 10 min window on and off (expect no change; the
   encoder is in d2dview, d2d does one memcpy as before).
5. Fault checks: `curl -N /stream.pfeed?t=...` without a token gives 401; kill d2d (the page stalls, then reconnects with a
   key frame); stop reading (a paused tab): the server drops the client after the write deadline and the lease runs out.
6. Panel open and close ten times in a minute: the feed's bytes stay under about 2 MiB/s per client (the guard), no tear.
7. Devices: dpr 1 desktop, dpr 2 laptop, a phone: the canvas is sharp (no blur at 100 percent zoom), Low rate works on
   the canvas, the card falls back to MJPEG with a blocked `DecompressionStream` (devtools override) and stays up. Waterfox.
8. Verdict: keep the switch off (default), or ask the user to turn it on for the sandbox. Record the numbers in
   docs/probes/video-feed-<date>.md.

## 11. Rate, presentation clock and fit (card videofeed3, 2026-11-16, after the first live look)

User feedback on the sandbox (desktop Waterfox, ~1075 px window): "~14-15 fps", "sorta jittery", "does not fill the window"
(a centred 800x600 with black bars on all four sides).

1. Why 14-15 fps (V by test, cause read in `cmd/d2dview/frames.go`; the sandbox flags themselves were not looked at). d2dview asks
   d2d for every k-th drawn frame (`divisor`): the smallest whole k whose period is at least the `-fps` target AND the back-off
   budget of the JPEG encode (4 x cost). A whole k is kept so game time stays evenly spaced. That rule belongs to the JPEG
   stream; the lossless feed inherited it. With `-fps 15` (or a JPEG cost of 15 to 20 ms on the sandbox's single core) k is 4 =
   15 fps; `-fps 25` gives k = 3 = 20 fps because 25 fps (40 ms) rounds UP to every 3rd draw of a 60 Hz source. Nothing
   downstream capped it: the hub paces at 60 fps per client, the guard is a byte cap, and the d2bot tap relay copies bytes
   (`D2_VIDEO_FPS`/`D2BOT_WATCH_VIDEO_FPS` sets the ffmpeg or tap-capture rate of the MJPEG path only).
   Fix: while a lossless viewer is connected (`Frames.palActive`, kept alive by `TouchPal`) the target is 33 ms (`palPeriod`):
   every 2nd draw of a 60 Hz source = 30 fps, every draw of a 30 Hz one. The JPEG keeps its own cadence next to it
   (`jpegWanted`: at most one encode per `-fps` gap), so a mixed audience does not raise the JPEG cost. The hub's pace is now a
   schedule (previous slot plus one period) instead of "write time plus period": the loop's own work no longer slips a slot
   every thirty frames (test `TestDeliversEveryFrameOfA30HzSource`, `TestDivisorFollowsRenderRateForPalViewers`).
2. Jitter: presentation clock in `palfeed.js` (`Player`). Every message carries the d2d game time of its frame. The player
   keeps the smallest (arrival - game time) of the last 2 s, presents frame g at g + that + D, with D = one frame interval +
   1.5 x the measured jitter (clamped to 1 to 2 frames, at most 70 ms), decodes what is due in order and draws once per
   animation frame (latest wins); a message older than 80 ms at presentation is not waited for; heartbeats (same game
   time) are applied at once and ignored by the clock; a game clock that does not follow the wall clock (fast-forwarded
   lockstep) degrades to "present on the next animation frame". Allocation per frame: the splitter keeps one growing
   buffer (no per-chunk array), the decoder reuses two pixel buffers and one `ImageData` per buffer; the compressed message
   is handed to the inflater at once (a Blob copies it) so the splitter buffer can be reused. Stats (`Player.stats()`, shown in the
   card head while slow or with `?debug`): delivered, decoded, shown (draws), dropped (decoded and merged into a later draw),
   draw-interval standard deviation, latency, current delay, main-thread work per animation frame.
3. Fill: the canvas is sized by `PalFeed.fitFill` (contain, 4:3 kept, ANY fractional scale, rounded to whole device pixels,
   `devicePixelRatio` aware, re-fitted by a `ResizeObserver` on the frame and in full screen). The old integer scale left
   bars (1x in a 1000 px box). Scaling: nearest-neighbour only at a whole-number scale of 1x or more, smooth otherwise
   (a fractional scale with nearest-neighbour shimmers while the picture scrolls); the card button "Sharp pixels" (remembered per
   browser) forces nearest-neighbour. The MJPEG `<img>` already filled its 4:3 frame (`object-fit: contain`, panel.css); the
   uishots pictures (`pal=1` fixtures, `tools/uishots`) show both.


Measured (V, `tools/videofeed/run3.sh` on the builder, headless Chromium 156, animation frames every 16.67 ms; the real `Player`
against `stream3.js`, a paced feed of 376 recorded frames at 25 Hz stamps with 0 to 25 ms of network-like jitter and 3 percent
80 ms stalls with a burst after them, 15 s per phase). The 25 Hz source on a 60 Hz display has an unavoidable draw-interval
spread of 8.2 ms (intervals alternate 33 and 50 ms: phase C, no jitter); a 30 Hz source on 60 Hz has none.

| phase | delivered | drawn | merged (dropped) | draw fps | interval sd | delay | main thread per frame |
|---|---|---|---|---|---|---|---|
| A jitter, present on arrival (the old behaviour) | 376 | 357 | 18 | 23.6 | 17.3 ms | 9 ms | 2.65 ms |
| B jitter, presentation clock | 376 | 373 | 1 | 24.8 | 10.5 ms | 64 ms | 1.23 ms |
| C no jitter, presentation clock | 376 | 375 | 0 | 25.0 | 8.2 ms | 50 ms | 1.33 ms |

So the clock removes 17 of 18 merged frames and cuts the interval spread from 17.3 to 10.5 ms (the floor is 8.2) for a
presentation delay of one to two frames (64 ms here, bound 70 ms). Not measured: Firefox and Waterfox (not installed on the
builder; the player uses only `fetch` streams, `DecompressionStream("deflate-raw")`, `ImageData` and `requestAnimationFrame`, all in
Firefox 113+), a real GPU, the sandbox itself. The same node tests drive the clock on a fake rAF and fake time
(`tools/videofeed/palfeed.test.js`: jitter 3.0 ms against 14.3 ms for present-on-arrival at 30 fps, delay 65 ms, 150 of 150
frames drawn).

A/B additions to section 10: with `?debug` the card head must show ~30 fps (not 14 to 15) and `dropped` near 0 while walking;
`jitter` below 12 ms; the window must be filled (no black bars at 1075 px); the Sharp pixels button must appear only with the feed;
the d2dview `-fps` flag is no longer the cap while a canvas viewer is connected.

## 12. Hitching on the sandbox: measured cause, the monitor, the fixes (card videofeed4, 2026-11-16 to 17)

Symptom (user, sandbox with the feed on, steady walking, no combat): "smooth, then fps drops, then hitches", the low-fps hint
up for 2 to 3 s and gone again, "arbitrary". Read-only sampling inside the sandbox container (V; one `docker exec` per
sample, per-second `/proc/<pid>/stat` ticks of d2dview, d2d and d2bot plus the cgroup counters; nothing was changed there; a
temporary sampling script I copied into the container was removed again):

- The container has a ONE-core CPU quota (`cpu.max` 100000 100000, `docker inspect` NanoCpus 1e9), enforced in 100 ms periods.
  Over a 254 s window with the page open: d2d 30 percent of the core on average (11 to 68), d2dview 11 percent (0 to 46), d2bot 33
  percent (2 to 87). Over a longer history 9 percent of all periods were throttled.
- The bot is bursty: stretches of 10 to 47 s at 64 to 71 percent of the core (the logs show a repair trip, town walks and
  `nav.trace` lines every 100 ms at the time) against 15 percent in between. In those stretches the container is throttled in
  7 to 10 of every 10 periods; 61 of the 254 seconds had 8 or more of 10 periods throttled, and their bot CPU averaged 67 percent
  against 15 for the calm seconds. Burst starts: 02:01:56, 02:03:23, 02:04:30 (container clock), 67 to 87 s apart.
- While the quota is saturated the renderer itself is starved: d2d falls from 33 to 18 to 22 ticks per second, in runs of 2 to 7 s
  (starts 02:03:32, :38, :45, 04:00, :03, :39: 3 to 16 s apart, 9 s typical), and a throttled period freezes d2d, d2dview and the
  bot together for the rest of that 100 ms. The signal that LEADS each dip is the bot's CPU already at 65 percent or more and
  7 to 10 of 10 periods throttled in the 3 s before it. That is the smooth -> drop -> hitch cycle: the picture's source (d2d) and
  its encoder (d2dview) lose the CPU to the bot, the arrivals gap by 100 ms or more, and the old 66 ms buffer could not hide it.
  It is not content, not the guard (a 2 s bucket never reached 4 MiB/s), not a key frame schedule, and not network: the cause is
  in the container's CPU budget.

What changed:
1. Stream monitor (`web/streamstats.js`, the card's Stats button, `?pfdebug`): a 30 s record of both players on one canvas, drawn
   at 5 Hz: shown/delivered/dropped fps against the target line, kB/s bars with key-frame ticks and the host's throttle share,
   a frame-time strip (hitches over 50 ms in vermillion and taller), the presentation delay, a time axis (HH:MM:SS at both ends with
   the zone abbreviation, ticks every 5 s, a "now" mark), a header with the window, build id, player, canvas size and scale,
   a summary line (fps, kB/s, p95 gap, hitches in 30 s) and the last three events with times (hitch, resync, key, hold, slow-on/off,
   fallback, reconnecting). "Copy 30 s" puts the same record on the clipboard as JSON (ISO start and end, timezone offset, build, view,
   per-second series, every gap, events, and the host sample's clock so server and page series align). The host's numbers come
   from `/api/palfeed` (`stats`: hub counters plus `server`: container throttle share, d2dview CPU, ring fps, measured from cgroup
   `cpu.stat` by `cmd/d2dview/hoststats.go`), polled once a second only while the overlay is on. The MJPEG player feeds the same
   record (arrival and image-decode times), without host numbers.
2. The low-fps hint has hysteresis (`SlowGate`): on after 3 s below 12 fps, off after 3 s above 18; the fps is the monitor's mean
   of shown frames over 3 whole seconds. The rate line in the card head and the hint no longer flap; every flip is an event with
   its time and a `console.info` line.
3. The presentation clock is stable and deeper: base transit follows a lower value at once and creeps up 0.2 percent per message;
   the lateness estimate follows a later arrival at once and decays with a half-life of about 11 s; delay = two frame intervals +
   1.1 x lateness, 66 to 200 ms. A synthetic walking trace with a 95 ms freeze every 1.5 s shows no presented gap over 52 ms after
   the first stall (the old clock shows them; both are asserted, `tools/videofeed/palfeed.test.js`), and the delay only ever relaxes
   slowly after a stall.
4. The feed degrades gracefully: while more than 60 percent of the container's CPU periods are throttled the lossless feed asks d2d
   for every 4th draw (15 fps) instead of every 2nd, which halves d2dview's work when the container is starved (it returns to 30
   fps when throttling stops; `throttleHigh`).
5. d2dview feed CPU: published frames come from a pool with reference counts (no 1.9 MB allocation per frame), the guard's key-size
   hold is 150 ms (was 400), the byte cap 4 MiB/s.
Not fixed here (not this card's code): the bot's own CPU bursts. They are the root cause; the bot, d2d and d2dview share one
core. A quota of 2 cores, or lower bot CPU in town (`nav.trace` at 10 Hz), would remove the stalls at the source.

## 13. The presentation stall seen in the overlay, and the build id (card videofeed5, 2026-11-17)

A user's Stats overlay (palette player, Waterfox) showed frames ARRIVING at 30 fps while nothing was shown for about 12 s, "dropped"
spiking to 35-40 several times in that stretch, no delay line while it lasted and ~200 ms after: a client-side stall, not CPU.
Reading `Player._tick`: it compared each message's due time with the animation frame's own timestamp, while the due times are built
from `performance.now()`. A browser whose frame timestamps lag or jump against `performance.now()` (timer clamping, a throttled
or occluded window) holds every frame back until the two clocks meet and then releases the backlog in one draw (the "dropped"
spike); animation frames that are merely withheld do the same. Not confirmed on the user's machine (no data from it, U), but the three
fixes below remove every mechanism of that shape and the new tests fail on the old player:
1. `_tick` uses `this.now()` for every comparison (the animation frame's timestamp is only compared, to log `clock-skew`).
2. A fallback timer (100 ms) drives ticks while animation frames are withheld; the monitor logs `raf-stall`.
3. Watchdog: a head message still being inflated 500 ms after it arrived (a key frame stuck in the decompressor) reconnects for a
   fresh key frame (`resync inflate stall N ms`); `delay-reset` is logged when a game-time jump back resets the clock estimate.
Regression tests (`tools/videofeed/palfeed.test.js`, fake clock, fake animation frames, fake timers): timestamp skew of -4 s, 3 s
without animation frames, a stuck key frame inflate, a game-time jump +10 s / -400 ms / -5 s, a 300 ms arrival pause; each asserts
no gap between draws over 250 ms while frames arrive and shown >= 90 percent of delivered. The first, second and third fail on the
player of #857.
Build id: the page header said `a756ec2-dirty`, the commit of the checkout the dev update was started from on the worker host
(a copied tree has no `.git`, a stale mirror's `.git` names the wrong commit; the go tool embeds whichever it finds). Now
`tools/dev-update.sh` computes the id of the tree it ships (`<sha7>` of `--commit`, of the `--worktree`, or of this checkout; `-dirty` only
when TRACKED files differ) and passes it with `--stamp` to `tools/dev-update-host.sh`, which links it as
`-X d2lod-bot/internal/buildinfo.Version`; `deploy/d2d-container/context.sh` takes `BUILD_STAMP` the same way. Test:
`internal/buildinfo/stamp_test.go` (a temp git repo: clean, named commit, tracked change, untracked file).

## 14. Item labels toggle (task labels, 2026-11-17; offline tests only, nothing deployed)

How the labels work in d2d (V, read in the fork branch `didle/rebase-20261115`): the names over ground items are D2's "Show Items" (Alt held). `apps/d2d/ingame.cpp` draws them into the world framebuffer when `show_items` is true: each name over its item (`rules::show_items`, at most 0x20, nearest rooms), black at half in the item's colour, the one under the mouse solid blue with white text. `town/draw.cpp` passes `alt_held || pad_show_items || function_held(kShowItems)`: the real Alt key, a skill slot bound to Show Items, or `Town::alt_held`. The labels are IN the framebuffer, so the palette feed and the MJPEG capture both show them (U live: pictures not taken here). The only older switch was `debug alt on|off`, which needs `--lab-verbs` (cheat family) and gives no read-back.

Verb (fork branch `didle/labels-verb`, off `didle/rebase-20261115`, private remote, fork-only, never upstream without the user's go): `labels [on|off|toggle|status]`, refused with `err lab verbs off: labels needs d2d --item-verbs` unless the process runs with `--item-verbs` (or `--lab-verbs`, which implies it). Reply `labels=on|off` then `ok`. `caps` lists `ui.labels` in the same condition. It sets `Town::alt_held`, which stays until set off (a click or a focus change cannot release it, unlike a key press); no game start, Character Select switch or save resets it inside one d2d process, a new process starts off. V: fork headless test `tests/labels_d2d.py` (`ctest -R labels_d2d`; skips without the MPQs).

Go side: `d2dctl.Client.Labels(ctx, mode)` (`internal/d2dctl/labels.go`, cap name `d2dctl.CapLabels`); `GET /api/labels` (viewer) and `POST /api/labels {"on":bool}` (answer listener, the same token and origin gate as `/api/pause`), reply `{ok, on, available, code?, reason?}` where `on` is what the game read back; `available:false` / code `unsupported` for a d2d without the verb (the button stays hidden), `game` when the d2d is unreachable, `refused` otherwise. The wish is per hero in `<instance state dir>/labels.json` (default off); the desk applies it when the bot enters a game (the Prepare hook) and on every GET (idempotent: reads `status`, sends `on|off` only when the game differs). Code: `internal/watch/labelsapi.go`, `cmd/d2bot/labelswire.go`.

Watch UI: a `Labels` toggle (`aria-pressed`, `data-key="labels"`) on the video card next to Full screen, Low rate, Stats and Sharp pixels, hidden until GET says available, re-read every 20 s; an `Item labels` entry in the "..." menu clicks it. Size follows the sibling buttons (`data-size="sm"`, 38 px in the phone cockpit).

Does it break the bot (V by code reading)? The bot on d2d reads `ground`, `dump` and memory-like verbs, never the picture: the only `Screenshot` caller is `cmd/d2dview` (the feed). In the fork `alt_held` is read in exactly one place, the draw call; clicks are not label-aware (only the label under the mouse becomes the hovered unit for the name tooltip), so a click lands where it did. U live: a click on a label while it is shown, in game.exe it picks the item up, in d2d no change is expected.

No knob: the feature exists only when the d2d answers `caps` with `ui.labels`, and the default is off. Live checklist (NOT run): start a sandbox d2d with `--item-verbs` from the `didle/labels-verb` build, drop an item (`debug`), open the watch page, tap Labels: names appear in the lossless and the MJPEG picture, the button stays pressed after a reload; restart the d2d: the label is back on when the bot is in game; switch hero: off for a hero that never set it.
