package watch

import "testing"

// realPause answers like bot.PlayerPauseView: no "ok" field, state and frozen only (the shape the live bug showed).
type realPause struct {
	frozen bool
	refuse string
}

func (p *realPause) view() any {
	m := map[string]any{"schema": "watch.pause/1", "state": "running", "can_freeze": true}
	if p.frozen {
		m["state"], m["frozen"] = "paused", true
	}
	if p.refuse != "" {
		m["error"] = p.refuse
	}
	return m
}
func (p *realPause) View() any { return p.view() }
func (p *realPause) Set(on bool) any {
	if p.refuse == "" {
		p.frozen = on
	}
	return p.view()
}

// Live bug 2026-11-16: the game froze, yet the Orders Pause verb (POST /api/commands -> POST /api/pause) was answered without
// "ok", and the panel read that as a refusal ("could not do that"). Every accepted press carries ok:true, a double press and
// Resume are harmless, and a real refusal keeps its own reason with ok:false.
func TestCommandsPauseReplyCarriesOK(t *testing.T) {
	desk := &realPause{}
	r := newFunRigWith(t, "Sorceress", 5, func(r *funRig, fc *FunConfig) {
		r.pause = NewPauseAPI()
		r.pause.Bind(desk)
	})
	for i, c := range []struct {
		body   string
		ok     bool
		frozen bool
	}{
		{`{"intent":"pause"}`, true, true},
		{`{"intent":"pause"}`, true, true}, // a double press
		{`{"intent":"resume"}`, true, false},
		{`{"intent":"resume"}`, true, false},
	} {
		code, m := cmdPost(r, funToken, CommandsRoute, c.body)
		if code != 200 || m["ok"] != c.ok || desk.frozen != c.frozen {
			t.Fatalf("press %d %s: %d %v frozen=%v", i, c.body, code, m, desk.frozen)
		}
	}
	desk.refuse = "could not freeze the game: no lockstep verb"
	code, m := cmdPost(r, funToken, CommandsRoute, `{"intent":"pause"}`)
	if m["ok"] != false || m["error"] != desk.refuse || m["code"] != "pause_refused" {
		t.Fatalf("a refusal must carry its reason: %d %v", code, m)
	}
	code, m = cmdPost(r, funToken, PauseRoute, `{"on":true}`)
	if code != 200 || m["ok"] != false || m["error"] != desk.refuse {
		t.Fatalf("direct route refusal: %d %v", code, m)
	}
}
