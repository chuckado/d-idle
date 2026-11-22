package main

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"d2lod-bot/internal/senses"
	"d2lod-bot/internal/watch"
)

// The Map-tab check (card mapphone, user bug 2026-11-22: on a phone the Area map was empty and its controls floated over the
// tab row). /__mapphone?u=<page>&w=&h=&fx=<variant>&click=<sel|sel> loads the page in an iframe of exactly w by h with the
// senses state of fx (the cookie mapfx), presses the selectors, measures the map (mapphone.js rules) and prints one line
// "MAPPHONE PASS|FAIL ..." into <pre id="mp-result">. tools/uishots.sh --map dumps that DOM in headless Chromium.

// mapFx is the senses state a variant serves: legacy (the frozen fixture, no fog fields), fresh (a new game or restart: the
// hero stands somewhere, nothing walked yet beyond the first reveal), walked (a trail of walked ground and an objective) and
// waiting (the bot has not observed the hero yet).
var mapFx = map[string]*senses.Publisher{}

func initMapFx() error {
	for name, stage := range map[string]string{"fresh": "start", "walked": "walked"} {
		p, err := fogPublisher(stage)
		if err != nil {
			return err
		}
		mapFx[name] = p
	}
	mapFx["waiting"] = &senses.Publisher{}
	return nil
}

// sensesFor answers /api/senses for the variant in the request's cookie; no cookie (or legacy) is the frozen fixture.
func sensesFor(r *http.Request, def http.Handler) http.Handler {
	c, err := r.Cookie("mapfx")
	if err != nil {
		return def
	}
	p := mapFx[c.Value]
	if p == nil {
		return def
	}
	return (&watch.Senses{Src: p, Lock: watch.FixedTier(1)}).Handler()
}

func mapPhoneHarness(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	u, ww, hh, fx := q.Get("u"), atoi(q.Get("w")), atoi(q.Get("h")), q.Get("fx")
	if !strings.HasPrefix(u, "/watch") || strings.HasPrefix(u, "//") || ww < 200 || hh < 200 || ww > 4000 || hh > 4000 {
		http.Error(w, "u must be a /watch path on this server; w and h 200 to 4000", http.StatusBadRequest)
		return
	}
	if fx == "" {
		fx = "legacy"
	}
	http.SetCookie(w, &http.Cookie{Name: "mapfx", Value: fx, Path: "/"})
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><title>mapphone</title></head><body style="margin:0;background:#888">
<iframe id="f" src="%s" width="%d" height="%d" style="border:0;display:block"></iframe><pre id="mp-result"></pre>
<script type="module">
import { measureMap } from '/__fit/mapphone.js';
const f = document.getElementById('f'), out = document.getElementById('mp-result');
const wait = (ms) => new Promise((r) => setTimeout(r, ms));
f.addEventListener('load', async () => {
  await wait(4000);
  try {
    for (const sel of %q.split('|').filter(Boolean)) { const n = f.contentDocument.querySelector(sel); if (n) n.click(); await wait(1500); }
    await wait(1500);
    out.textContent = measureMap(f.contentWindow, %q);
  } catch (e) { out.textContent = 'MAPPHONE FAIL harness: ' + e; }
});
</script></body></html>`, html.EscapeString(u+sep+"fitframe=1"), ww, hh, q.Get("click"), fx)
}
