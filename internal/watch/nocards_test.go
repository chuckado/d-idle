package watch

import (
	"strings"
	"testing"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/unlock"
)

// The identify-reveal card feature is gone (user report 2026-11-19: no "unidentified item" card anywhere). 50
// unidentified drops, on two classes, must produce no card route, no card toast or event line and no card payload; the
// page and the presets carry no Cards tab; unlocking never names cards.
func TestNoUnidentifiedCardFromDrops(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		t.Run(class, func(t *testing.T) {
			l, err := unlock.Shipped()
			if err != nil {
				t.Fatal(err)
			}
			eng, err := unlock.Open(t.TempDir(), strings.ToLower(class), l, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := newFunRigWith(t, class, 1, func(_ *funRig, fc *FunConfig) { fc.Unlock = eng })
			for i := 1; i <= 50; i++ {
				q := []string{"magic", "rare", "unique", "set"}[i%4]
				r.feed(goals.Event{Type: goals.EvDrop, Seq: uint64(i), Item: &goals.Item{Name: "Short Sword", Quality: q}})
			}
			if eng.Allowed("idreveal") || eng.Allowed("idreveal_later") {
				t.Error("the ladder still has a card row")
			}
			for _, route := range []string{"/api/reveal", "/api/reveal/flip"} {
				for _, base := range []string{r.view, r.answer} {
					if resp, _ := r.do("GET", base, route, funToken, ""); resp.StatusCode == 200 {
						t.Errorf("%s answers %d, want no card route", route, resp.StatusCode)
					}
				}
			}
			for _, route := range []string{"/api/away", "/api/chronicle", "/api/needs"} {
				_, body := r.do("GET", r.view, route, funToken, "")
				for _, bad := range []string{"Unidentified item", "face-down", "Flip the card", "idreveal"} {
					if strings.Contains(string(body), bad) {
						t.Errorf("%s mentions %q", route, bad)
					}
				}
			}
		})
	}
}

func TestShellHasNoCardsTab(t *testing.T) {
	ents, err := webFS.ReadDir("web/layouts")
	if err != nil || len(ents) == 0 {
		t.Fatal(err)
	}
	for _, e := range ents {
		if e.Name() == "classic.json" || e.Name() == "mockup.json" || e.Name() == "live.json" || e.Name() == "command.json" {
			continue // frozen (hashed) files keep the id; shell.json reserves it, so it mounts nothing
		}
		b, _ := webFS.ReadFile("web/layouts/" + e.Name())
		if strings.Contains(string(b), `"id": "cards"`) {
			t.Errorf("preset %s has a cards panel", e.Name())
		}
	}
	h, _ := webFS.ReadFile("web/watch.html")
	for _, bad := range []string{`id="tab-cards"`, "revealsheet", "cardspane", "Unidentified item"} {
		if strings.Contains(string(h), bad) {
			t.Errorf("watch.html still has %q", bad)
		}
	}
	for _, f := range []string{"web/inventory-actions.js"} {
		b, _ := webFS.ReadFile(f)
		if strings.Contains(string(b), "/api/reveal/flip") || strings.Contains(string(b), "'Unidentified item'") {
			t.Errorf("%s still builds the card", f)
		}
	}
}
