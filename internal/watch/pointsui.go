package watch

import (
	"io"
	"net/http"
)

// The Points page (docs/points.md): the sheet of GET /api/hero/points with one + button per attribute, the banked skill
// points with a link to the skills page, the per-hero autopilot switch and the trail of what came through the API. web/points.js
// draws it; every text goes through textContent and nothing is inline (CSP style-src 'self').

const pointsPage = `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="dark light"><title>d.idle - points</title><link rel="icon" href="data:,">
<link rel="stylesheet" href="/watch/tokens.css"><link rel="stylesheet" href="/watch/reset.css"><link rel="stylesheet" href="/watch/shell.css">
<link rel="stylesheet" href="/watch/components.css"><link rel="stylesheet" href="/watch/utilities.css"><link rel="stylesheet" href="/watch/points.css"><link rel="stylesheet" href="/watch/nav.css">
</head><body>
<div class="d-shell">
<header class="top d-band-status" data-band="status">
 <div class="d-brand"><span class="d-brand-mark" aria-hidden="true"></span><h1>d.idle</h1></div>
 <div class="d-band-main"></div>
 <div class="d-links" data-nav-host>
  <button class="d-btn" data-size="sm" data-variant="ghost" id="btn-theme" type="button" aria-label="Switch colour theme">Theme: dark</button>
 </div>
</header>
<main class="pt-page" id="pt">
 <p class="d-note" data-sev="bad" id="msg" role="status" hidden></p>
 <section class="d-card" aria-labelledby="t-sum">
  <div class="d-card-head"><h2 id="t-sum">Points</h2><span class="meta" id="sum"></span></div>
  <p class="d-hint" id="waiting"></p>
  <p class="d-hint" id="skillline"></p>
  <p class="d-hint" id="buildline"></p>
 </section>
 <section class="d-card" aria-labelledby="t-stats">
  <div class="d-card-head"><h2 id="t-stats">Attributes</h2><span class="meta" id="statmeta"></span></div>
  <ul class="pt-stats" id="stats"></ul>
  <p class="d-hint" id="note"></p>
 </section>
 <section class="d-card" aria-labelledby="t-gear">
  <div class="d-card-head"><h2 id="t-gear">What your gear asks for</h2></div>
  <ul class="pt-gear" id="gear"></ul>
 </section>
 <section class="d-card" data-part="build" aria-labelledby="t-build">
  <div class="d-card-head"><h2 id="t-build" tabindex="-1">Build</h2><span class="meta" id="bmeta"></span></div>
  <p class="d-hint" id="bstate"></p>
  <p class="d-note" data-sev="info" id="boffnote" hidden></p>
  <p class="d-hint" id="bheld" hidden></p>
  <p class="d-hint" id="boff" hidden></p>
  <p class="d-hint" id="brespecline"></p>
  <p class="d-note" data-sev="bad" id="bcerr" role="status" hidden></p>
  <ul class="bc-list" id="bcards"></ul>
 </section>
 <section class="d-card" aria-labelledby="t-auto">
  <div class="d-card-head"><h2 id="t-auto">Autopilot</h2></div>
  <p class="d-hint" id="autodesc"></p>
  <button class="d-btn" data-size="sm" id="auto" type="button"></button>
 </section>
 <section class="d-card" aria-labelledby="t-trail">
  <div class="d-card-head"><h2 id="t-trail">Recent actions</h2></div>
  <ul class="pt-trail" id="trail"></ul>
 </section>
</main>
</div>
<div id="toasts" class="d-toasts" role="status" aria-live="polite"></div>
<script src="/watch/nav.js"></script>
<script type="module" src="/watch/shell.js"></script>
<script src="/watch/sheet.js"></script>
<script src="/watch/buildcards.js"></script>
<script src="/watch/points.js"></script></body></html>`

// MountPointsPage serves the page and its two assets without a desk behind them, for the screenshot suite's fixture viewer.
func MountPointsPage(m *http.ServeMux) {
	m.HandleFunc("/watch/points", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, pointsPage)
	})
	for _, f := range []struct{ name, ct string }{{"points.js", "text/javascript; charset=utf-8"}, {"buildcards.js", "text/javascript; charset=utf-8"}, {"points.css", "text/css; charset=utf-8"}} {
		m.HandleFunc("/watch/"+f.name, func(w http.ResponseWriter, r *http.Request) {
			b, _ := webFS.ReadFile("web/" + f.name)
			w.Header().Set("Content-Type", f.ct)
			w.Write(b)
		})
	}
}
