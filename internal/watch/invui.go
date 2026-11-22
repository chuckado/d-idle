package watch

// The inventory page (docs/ui-design-system.md components 13 and 14): the shared shell (tokens, the top band, theme)
// with a Diablo II paper doll, the backpack, belt and stash as cell grids, and the repair card. web/inventory.js draws
// /api/inventory and /api/repair; every text goes through textContent and nothing is inline (CSP style-src 'self': the
// old page's <style> block was refused, so it rendered unstyled). Writes go to the answer listener.

const invPage = `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="dark light"><title>d.idle - inventory</title><link rel="icon" href="data:,">
<link rel="stylesheet" href="/watch/tokens.css"><link rel="stylesheet" href="/watch/reset.css"><link rel="stylesheet" href="/watch/shell.css">
<link rel="stylesheet" href="/watch/components.css"><link rel="stylesheet" href="/watch/inventory.css"><link rel="stylesheet" href="/watch/utilities.css">
<link rel="stylesheet" href="/watch/inventory-tip.css"><link rel="stylesheet" href="/watch/inventory-actions.css"><link rel="stylesheet" href="/watch/inventory-merc.css"><link rel="stylesheet" href="/watch/inventory-sets.css">
</head><body>
<div class="d-shell">
<header class="top d-band-status" data-band="status">
 <div class="d-brand"><span class="d-brand-mark" aria-hidden="true"></span><h1>d.idle</h1></div>
 <div class="d-band-main"></div>
 <nav class="d-links" aria-label="Other pages">
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch?live=1">Live screen</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/skills">Skills</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/policy">Policy</a>
  <button class="d-btn" data-size="sm" data-variant="ghost" id="btn-theme" type="button" aria-label="Switch colour theme">Theme: dark</button>
 </nav>
</header>
<main class="inv-page" id="inv">
 <p class="d-note" data-sev="alarm" id="msg" role="status" hidden></p>
 <p class="d-note" data-sev="info" id="gearprompt" role="note" hidden></p>
 <section class="d-card inv-equip" aria-labelledby="t-equip">
  <div class="d-card-head"><h2 id="t-equip">Equipped</h2><span class="meta" id="equipmeta"></span></div>
  <div class="inv-doll-wrap"><div class="d-doll" id="equip" role="group" aria-label="Paper doll"></div><div class="inv-detail" id="detail"></div></div>
  <div class="d-legend" aria-label="Legend"><span><b class="lg-sell">S</b>sell at the next vendor</span><span><b class="lg-keep">K</b>keep, never sold</span><span>pips: durability, amber under 30 percent, red and BROKEN at 0</span></div>
  <div class="inv-merc" id="merc-doll" hidden></div>
  <div class="inv-sets" id="gearsets" role="group" aria-label="Gear sets" hidden></div>
 </section>
 <section class="d-card inv-pack" aria-labelledby="t-pack">
  <div class="d-card-head"><h2 id="t-pack">Backpack</h2><span class="meta" id="packmeta"></span></div>
  <div class="d-cells" id="pack"></div>
  <p class="d-hint" id="note"></p>
 </section>
 <section class="d-card inv-repair" aria-labelledby="t-repair">
  <div class="d-card-head"><h2 id="t-repair">Repair</h2></div>
  <div class="inv-repair-body" id="repair"></div>
 </section>
 <section class="d-card inv-belt" aria-labelledby="t-belt">
  <div class="d-card-head"><h2 id="t-belt">Belt</h2></div>
  <div class="d-cells" id="belt"></div>
 </section>
 <section class="d-card inv-stash" aria-labelledby="t-stash">
  <div class="d-card-head"><h2 id="t-stash">Stash</h2><span class="meta" id="stashmeta"></span></div>
  <div class="d-cells" id="stash"></div>
 </section>
 <section class="d-card inv-queue" aria-labelledby="t-queue">
  <div class="d-card-head"><h2 id="t-queue">Item actions</h2></div>
  <div class="inv-errands" id="errands" hidden></div>
  <div class="inv-errands inv-stashgold" id="stashgold" hidden></div>
  <div id="queue"></div>
 </section>
</main>
</div>
<div id="toasts" class="d-toasts" role="status" aria-live="polite"></div>
<script type="module" src="/watch/shell.js"></script>
<script src="/watch/inventory.js"></script>
<script src="/watch/inventory-tip.js"></script>
<script src="/watch/inventory-actions.js"></script>
<script src="/watch/inventory-merc.js"></script>
<script src="/watch/inventory-drag.js"></script>
<script src="/watch/inventory-sets.js"></script></body></html>`

// charPage is the character canvas (card uic2, docs/ui/redesign-2026-11-14/briefs/C2-character-canvas.md): the same data,
// scripts and element ids as invPage, laid out as ONE frame on one cell unit: the hero's and the mercenary's dolls, the set
// bar and the stat sheets (who wears), the pack over the stash with the belt, the selected item and the town errands beside
// it (storage), and a bar with search, filters, gold, repair and sell. /watch/inventory serves it; /watch/inventory?view=classic
// serves invPage (the cards page) until the canvas has been looked at live. The data-cockpit-slot and data-cockpit-panel
// attributes are what the fit check (tools/uishots.sh --fit) measures. Nothing is inline (CSP style-src 'self').
const charPage = `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="dark light"><title>d.idle - character</title><link rel="icon" href="data:,">
<link rel="stylesheet" href="/watch/tokens.css"><link rel="stylesheet" href="/watch/reset.css"><link rel="stylesheet" href="/watch/shell.css">
<link rel="stylesheet" href="/watch/components.css"><link rel="stylesheet" href="/watch/inventory.css"><link rel="stylesheet" href="/watch/utilities.css">
<link rel="stylesheet" href="/watch/inventory-tip.css"><link rel="stylesheet" href="/watch/inventory-actions.css"><link rel="stylesheet" href="/watch/inventory-merc.css"><link rel="stylesheet" href="/watch/inventory-sets.css"><link rel="stylesheet" href="/watch/inventory-canvas.css">
</head><body class="cv-root" data-canvas>
<div class="d-shell cv" id="cv" data-tab="hero">
<header class="top d-band-status" data-band="status" data-cockpit-slot="band" data-fit-deep>
 <div class="d-brand"><span class="d-brand-mark" aria-hidden="true"></span><h1>d.idle</h1></div>
 <div class="d-band-main"></div>
 <nav class="d-links" aria-label="Other pages">
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch?live=1">Live screen</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/skills">Skills</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/points">Points</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/vault">Vault</a>
  <a class="d-btn" data-size="sm" data-variant="ghost" href="/watch/policy">Policy</a>
  <button class="d-btn" data-size="sm" data-variant="ghost" id="btn-theme" type="button" aria-label="Switch colour theme">Theme: dark</button>
  <button class="d-btn cv-morebtn" data-size="sm" data-variant="ghost" id="cv-more" type="button" aria-label="More pages" aria-expanded="false" aria-controls="cv-morelist">&hellip;</button>
  <div class="cv-morelist" id="cv-morelist" role="group" aria-label="Pages" hidden><a href="/watch/inventory" aria-current="page">Inventory</a><a href="/watch/skills">Skills</a><a href="/watch/points">Points</a><a href="/watch/vault">Vault</a><a href="/watch/policy">Policy</a><a href="/watch/return">Return</a></div>
 </nav>
</header>
<div class="cv-bar" id="cv-bar" role="toolbar" aria-label="Search, filters, gold and errands" data-cockpit-slot="band" data-fit-deep>
 <input class="cv-q" id="cv-q" type="search" placeholder="Search pack and stash  /" aria-label="Search pack and stash" autocomplete="off">
 <span class="cv-chips" id="cv-chips" data-hscroll></span>
 <span class="cv-sp"></span>
 <span class="cv-kv">Gold <b id="cv-gold">-</b></span>
 <span class="cv-kv cv-opt">Stash <b id="cv-stashgold">-</b></span>
 <button class="cv-btn cv-repair" id="cv-repair" type="button" hidden></button>
 <button class="cv-btn cv-sell" id="cv-sell" type="button" aria-disabled="true">Sell marked (0)</button>
</div>
<div class="cv-tabbar" id="cv-tabbar" role="group" aria-label="Page part" data-cockpit-slot="band"></div>
<main class="cv-canvas" id="inv" data-vscroll>
 <section class="cv-slot" data-slot="who" data-cockpit-slot="who">
  <div class="cv-zone cv-who" data-cockpit-panel data-panel="char-who" aria-label="Hero and mercenary">
   <p class="d-note" data-sev="info" id="gearprompt" data-cv="prompt" role="note" hidden></p>
   <div class="cv-wear">
    <div class="cv-hero"><h2 id="t-equip">Hero <em id="equipmeta"></em></h2><div class="d-doll" id="equip" role="group" aria-label="Paper doll"></div></div>
    <div class="cv-mercbox"><h2>Mercenary <em id="merc-sub"></em></h2><div class="inv-merc" id="merc-doll" data-canvas hidden></div></div>
   </div>
   <div class="cv-setbar" id="cv-sets" role="group" aria-label="Gear sets"></div>
   <div class="cv-sheet" id="hsheet" data-cv="hsheet" aria-label="Hero values"></div>
   <div class="cv-log" id="cv-log"><h2 id="t-queue">Item actions</h2><div id="queue" data-cv="queue"></div></div>
   <div class="cv-rules" id="loot-rules"><h2>Loot rules</h2><p>Your rules and the starter preset show here once the loot rules tab is built</p></div>
  </div>
 </section>
 <section class="cv-slot" data-slot="store" data-cockpit-slot="store">
  <div class="cv-zone cv-storez" data-cockpit-panel data-panel="char-store" aria-label="Pack and stash">
   <h2 id="t-pack" class="cv-hpack">Pack <em id="packmeta"></em></h2>
   <div class="d-cells" id="pack" data-cv="pack"></div>
   <h2 id="t-stash" class="cv-hstash">Stash <em id="stashmeta"></em></h2>
   <div class="cv-store">
    <div class="d-cells" id="stash" data-cv="stash"></div>
    <div class="cv-side">
     <h2 class="cv-beltlabel" id="t-belt">Belt</h2>
     <div class="d-cells" id="belt" data-cv="belt" aria-label="Belt"></div>
     <div class="cv-sel" id="cv-sel" aria-live="polite" data-vscroll></div>
     <div class="cv-errwrap" id="cv-errwrap">
      <button class="cv-errtoggle" id="cv-errtoggle" type="button" aria-expanded="false" aria-controls="errands stashgold">Town errands</button>
      <div class="inv-errands" id="errands" hidden></div>
      <div class="inv-errands inv-stashgold" id="stashgold" hidden></div>
     </div>
     <p class="cv-lock" id="cv-lock">Stash from anywhere: locked (skips a game rule)</p>
     <p class="d-note" data-sev="alarm" id="msg" data-cv="msg" role="status" hidden></p>
    </div>
   </div>
  </div>
 </section>
</main>
<div class="cv-pop" id="cv-pop" hidden role="dialog" aria-label="Gear sets"><button class="cv-btn cv-popx" id="cv-popx" type="button">Close</button><div id="gearsets" hidden></div></div>
<div class="cv-hidden" hidden><div id="repair"></div><div id="detail"></div><span id="note"></span></div>
</div>
<div id="toasts" class="d-toasts" role="status" aria-live="polite"></div>
<script type="module" src="/watch/shell.js"></script>
<script src="/watch/inventory.js"></script>
<script src="/watch/inventory-tip.js"></script>
<script src="/watch/inventory-actions.js"></script>
<script src="/watch/inventory-merc.js"></script>
<script src="/watch/inventory-drag.js"></script>
<script src="/watch/inventory-sets.js"></script>
<script src="/watch/inventory-canvas.js"></script></body></html>`
