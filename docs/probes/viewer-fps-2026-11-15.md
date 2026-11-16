# Viewer game picture, fps report (2026-11-15)

Report: "very bad fps in the game view of the friends app". Friend's machine (user, 2026-11-15): Arch Linux, Radeon RX 7900 XTX,
fast CPU, browser Waterfox (Firefox based). So raw CPU is not the cause. Tags: V verified (how), S read in code, U unverified.

## Path (S, read in code)
d2d frame -> capture (`play` default: ffmpeg x11grab of the private display, 8 fps, `-q:v 5`; with the tap: the d2dview
`/stream.mjpg` relay) -> `watch.Video` (one capture, one pump, JPEG shared by all viewers, no per-viewer encode) -> `GET /stream`
multipart MJPEG -> page: `fetch` reader, `partParser` splits parts, each part becomes a Blob object URL set as `<img>.src`
(`web/mjpeg.js`, `web/panels/video/panel.js`). No canvas, no WebGL, no base64, no JSON per frame.

## Findings
1. Server side is sound (V by tests): the pump hands each viewer a 1-slot channel, a slow viewer only misses frames
   (`TestVideoSlowViewerMissesFramesAndNothingQueues`), a viewer that stops reading is dropped after 10 s, one encode serves all viewers.
2. `play` default is the x11grab source at 8 fps by design (`DefaultVideoFPS`), so a friend without the d2d frame feed sees at most 8 fps
   at JPEG q5 whatever the machine: this is the most likely "bad fps" on a fast machine (S). The d2d feed/tap path (20 fps) is the
   lab sandbox's. Raising it needs the friends build to ship the tap; not done here.
3. Page side (S): the old `partParser` re-copied the whole pending buffer on every chunk and again per part, and each frame costs a
   Blob + object URL + image decode on the main thread, which Firefox does synchronously with the rest of the cockpit's work.
   No `backdrop-filter`, CSS filter or per-frame overlay repaint exists in the watch CSS (V: grep over `internal/watch/web`), so the
   cockpit overlay is not a cause; the card head only changes when its text changes.
4. Nothing in the page told the friend or us the rate (no stat outside `?debug`).

## Changes (all additive)
- Server: `/stream?fps=N` per-viewer cap; identical frames skipped (at most 1 s); counters; `/api/glance` field `stream`.
- Page: parser without per-chunk copies; "Low rate" button (8 fps, remembered per browser); a hint after 6 s under 15 fps; a
  `stream: N fps, K KB/s` line in the card head while slow or with `?debug`; `window.d2video.stats()`.
- Tests: `go test ./internal/watch -run 'TestVideo'`, `node internal/watch/testdata/video_panel.js internal/watch/web/panels/video/panel.js`
  (via `go test ./internal/watch -run Layout`). Knobs: docs/knobs.md section 9.

## Measurements
Server side, local loopback, fake 90 KB frames every 40 ms (V, ad-hoc test, 5 s each; not committed):

| viewers | `/stream` | parts per viewer / 5 s | fps |
|---|---|---|---|
| 1 | all frames | 125 | 25 |
| 3 | all frames | 125 each | 25 |
| 1 | `?fps=8` | 43 | 8.6 |
| 3 | `?fps=8` | 43 each | 8.6 |

NOT measured (U): browser fps in Chromium or Firefox, frame bytes at 1280x720, 1914x930, 412x780, CPU throttling, cockpit vs classic.
The Tower builder was saturated (load 15.8 on 16 cores, 3 GB free) and did not run even a `which` probe within 100 s, and the brief
bans Orion and herdr for this. Checklist once a builder is free: start `d2bot --watch --watch-video` against the fake capture, load
`/watch` in headless Chromium (CDP `Emulation.setCPUThrottlingRate 4`) and Firefox (Playwright), read `window.d2video.stats()` for 30 s.

## Ask the friend
OS and Waterfox version, `about:support` Graphics section (compositing, WebRender, GPU), window size, the page address with `?debug`
(or the `stream:` line in the video card head and `/api/glance` `stream`) as a screenshot, and whether `Low rate` helps.
