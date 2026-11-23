package main

import (
	"net/http"

	"d2lod-bot/internal/watch"
)

// Defects of the shared navigation (card uinav2): the real page with ONE defect put back. The fit pages navdefect-NAME must FAIL
// with the verdict mustFail names (fit.go); a pass there means the check is blind to it.
func init() {
	defectCSS["nav-covers"] = `body { padding-block-end: 0 !important; }`                                                       // the bar over the end of the page
	defectCSS["nav-small"] = `.d-nav-i { min-block-size: 20px !important; block-size: 20px !important; }`                       // targets under 44 px
	defectCSS["nav-sheet"] = `.d-navsheet { inset-block-end: 0 !important; }`                                                   // the More sheet over the bar
	defectCSS["ts-vcut"] = `.d-btn { overflow: hidden !important; block-size: 16px !important; min-block-size: 0 !important; }` // controls with a fixed height that the larger text no longer fits (the ts200 pages must fail)
	defectCSS["nav-wide"] = `.d-nav { position: fixed !important; inset-block-end: 0 !important; inset-inline: 0 !important; }` // a bar on a wide window
}

// hallFixture serves the Hall of Fame page of the fixture viewer (the real renderer, two heroes), so the suite can render it
// next to the other pages of the shared frame.
func hallFixture(m *http.ServeMux) {
	m.HandleFunc(watch.HallPage, func(w http.ResponseWriter, r *http.Request) {
		v := &watch.HallView{Account: "2 heroes", Heroes: []watch.HallLine{
			{HallRecord: watch.HallRecord{Name: "Zapp", Class: "sorceress", Level: 31, Kills: 1204, GameSecs: 90000, WallSecs: 40000, BestItem: "Griffon's Eye"}, Status: "retired", Title: "Slayer of Andariel"},
			{HallRecord: watch.HallRecord{Name: "Holy", Class: "paladin", Level: 24, Deaths: 1, GameSecs: 30000, WallSecs: 12000, Cause: "killed by Mephisto"}, Status: "dead", Title: "Fallen Paladin"},
		}}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
		w.Write(watch.RenderHall(v))
	})
}
