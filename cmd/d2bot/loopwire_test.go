package main

import (
	"testing"

	"d2lod-bot/internal/watch"
)

// The Loop is served by the real wiring, shut at the start, and a start is refused (409 locked) until its row opens.
func TestOrderLoopWired(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		t.Run(class, func(t *testing.T) {
			w := newWireRig(t, class, t.TempDir(), nil)
			if code := httpDo(t, "GET", w.view+"/api/orderloop", nil, nil); code != 401 {
				t.Errorf("loop without a token: %d", code)
			}
			var v watch.LoopView
			w.get(t, "/api/orderloop", &v)
			if v.Open || v.State != watch.LoopLocked || v.Teaser == "" {
				t.Fatalf("locked view: %+v", v)
			}
			if code := w.post(t, "/api/orderloop", map[string]any{"action": "start", "area": "Blood Moor", "minutes": 10}, nil); code != 409 {
				t.Fatalf("start while locked: %d", code)
			}
			w.rig.Fun.Unlocks.Seen(watch.LoopManualEvent)
			w.get(t, "/api/orderloop", &v)
			if !v.Open || v.State != watch.LoopIdle {
				t.Fatalf("open view: %+v", v)
			}
		})
	}
}
