package main

import (
	"net/http"
	"strings"
)

// Defects of the overlay policy (card overlayfix, bug report 2026-11-15: a zone sheet covered the "..." menu and its text was cut):
// the real page with ONE defect put back. defectCSS entries named "ov-*" are the style ones; overlayJS is the script ones (a
// script from this viewer's own origin, so the page CSP allows it). The fit pages overlaydefect-NAME must FAIL with the verdict
// mustFail names (fit.go).
func init() {
	defectCSS["ov-clip"] = `[data-panel="zone-map"] .zm-l { white-space: nowrap !important; overflow: hidden !important; text-overflow: ellipsis !important; }`
	defectCSS["ov-header"] = `.d-tilesheet { inset-block-start: -90px !important; }`
	defectCSS["ov-two"] = `.d-tilesheet[hidden] { display: flex !important; }`
}

// overlayJS: a transparent full-window popover shown after the menu opens sits above it in the top layer, as a sheet that
// covered the menu would (a stacking-order bug cannot be put back by CSS once the menu is in the top layer).
var overlayJS = map[string]string{
	"ov-under": `document.addEventListener('toggle', function (e) {
  if (!e.target || e.target.id !== 'menu-more' || e.newState !== 'open') return;
  var c = document.createElement('div'); c.setAttribute('popover', 'manual');
  c.style.cssText = 'position:fixed;inset:0;margin:0;border:0;padding:0;background:transparent;inline-size:auto;block-size:auto';
  document.body.appendChild(c); c.showPopover();
}, true);`,
}

func serveDefectJS(w http.ResponseWriter, r *http.Request) bool {
	n, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/__fit/defect-"), ".js")
	if !ok || overlayJS[n] == "" {
		return false
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Write([]byte(overlayJS[n]))
	return true
}
