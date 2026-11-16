# Lossless delta frame feed for the watch video (card videofeed, 2026-11-15)

Tags: V = verified (read in source or run, how is said), S = secondary (read in docs), U = unverified.
Numbers are in `docs/probes/video-feed-2026-11-15.md` (real d2d frames); this page is the design and the status.
User go 2026-11-15: design and prototype a better feed than ffmpeg x11grab / the container tap. NOT deployed, nothing wired
into a page or a server: `internal/watch/palfeed` (codec, hub, handlers) and `internal/watch/web/palfeed.js` (player) are
additive and unused until a host mounts them; knob name reserved: `watch.paletteFeed` (default off, not registered: the
knob registry is not this card's).

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
18 u32 game ms  22 u32 payload length  26 payload
payload (inflated): dirty-tile bitmap (row-major, bit 0 first), then per dirty tile its RGB bytes, row by row
  key frame : every tile; each row left-neighbour filtered (byte minus the byte 3 to the left, first pixel raw)
  delta     : only tiles that differ from the base frame; each byte is XOR with the base frame's byte
```

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

## 5. CPU, bytes, fps

See the probe for the measured table (real recorded d2d frames: bytes per frame p50/p95 per scene, encoder ms, decoder
ms in node and in headless Chromium, JPEG q72 and whole-frame XOR as references). Sandbox context (V, docs/viewer-fps.md
section 6): one CPU for the whole container, JPEG q72 of a full frame costs about 10 ms and 77 to 90 KB, 21.6 fps
delivered at 1.9 MB/s. The delta feed has no per-frame cost on d2d, and in d2dview the encoder replaces the JPEG encode
(the comparison is in the probe, measured on a loaded builder core, so read it as an upper bound).

## 6. Security

Same token as the existing streams; no new listener; the feed carries only what the picture already shows; the decoder
caps a message at 16 MiB and inflates at most `w*h*3 + w*h/8 + 1024` bytes (Go) / checks lengths before reading (JS)
so a hostile or broken stream cannot make the page allocate unboundedly. The ring file stays mode 0600 on tmpfs (unchanged).

## 7. Friends pack impact

None until wired. When wired: no new binary, no new dependency, no new port; `cmd/d2dview` and the watch viewer grow by
the package (about 400 lines of Go) and one 200-line JS file in the embedded web tree; the d2d image is unchanged (it
already carries `framefeed`, the pinned container ref predates it: the friends container needs the `d2d.lab3-framefeed`
build that the sandbox runs, or the feed falls back to MJPEG from screenshots, which cannot feed this codec at full rate).
The sandbox's `-feed` ring is the prerequisite; a friends release without it keeps MJPEG (the capability check says so).

## 8. Status and next steps

Prototyped (V, tests): codec both ways, hub, stream and capability handlers, auth, latest-frame-wins, JS decoder
byte-compatible with Go (shared fixtures), splitter, fit, detect, player draw pacing; replay and measurement tooling
(`tools/videofeed/run.sh`). Only designed: mounting in d2dview and the watch viewer, ring-client counting, per-write
deadline, the `watch.paletteFeed` knob and page switch, indexed payload.
Next: (1) wire behind the flag in d2dview + `watch.Video` relay (one card, `cmd/d2dview` and `internal/watch` owners);
(2) sandbox A/B against MJPEG with the coordinator's `ab.sh`: fps, d2dview CPU, bytes, bot pace; (3) a real-browser pass on
the user's devices (dpr 1 to 3).
