package watch

// The Points page (docs/points.md): the sheet of GET /api/hero/points with one + button per attribute, the banked skill
// points with a link to the skills page, the per-hero autopilot switch and the trail of what came through the API. web/points.js
// draws it; every text goes through textContent and nothing is inline (CSP style-src 'self').

const pointsPage = `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="dark light"><title>d.idle - points</title><link rel="icon" href="data:,">
<link rel="stylesheet" href="/watch/tokens.css"><link rel="stylesheet" href="/watch/reset.css"><link rel="stylesheet" href="/watch/shell.css">
<link rel="stylesheet" href="/watch/components.css"><link rel="stylesheet" href="/watch/utilities.css"><link rel="stylesheet" href="/watch/points.css">
</head><body>
<div class="d-shell">
<header class="top d-band-status" data-band="status">
 <div class="d-brand"><span class="d-brand-mark" aria-hidden="true"></span><h1>d.idle</h1></div>
 <div class="d-band-main"></div>
 <nav class="d-links" aria-label="Other pages">
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch?live=1">Live screen</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/skills">Skills</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/inventory">Inventory</a>
  <button class="d-btn" data-size="sm" data-variant="ghost" id="btn-theme" type="button" aria-label="Switch colour theme">Theme: dark</button>
 </nav>
</header>
<main class="pt-page" id="pt">
 <p class="d-note" data-sev="alarm" id="msg" role="status" hidden></p>
 <section class="d-card" aria-labelledby="t-sum">
  <div class="d-card-head"><h2 id="t-sum">Points</h2><span class="meta" id="sum"></span></div>
  <p class="d-hint" id="waiting"></p>
  <p class="d-hint" id="skillline"></p>
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
<script type="module" src="/watch/shell.js"></script>
<script src="/watch/points.js"></script></body></html>`
