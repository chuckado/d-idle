package merc

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

// two builds: the revive bar does not depend on the class, but the request path must behave the same for both
func reviveInputs(class string, level, gold int) Inputs {
	c := ctx()
	c.PlayerClass, c.Level, c.Gold = class, level, gold
	return Inputs{State: MercState{Known: true, Owned: true, Level: 10}, Ctx: c} // dead, price 750
}

func newPR() (*PlayerRevive, *time.Time) {
	t := time.Unix(1_800_000_000, 0)
	return NewPlayerRevive(func() time.Time { return t }), &t
}

func code(err error) string {
	var r *Refusal
	if errors.As(err, &r) {
		return r.Code
	}
	return ""
}

func TestPlayerReviveDeadWithGoldRequestsTheTrip(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		p, _ := newPR()
		in := reviveInputs(class, 10, 750)
		v := p.View(in)
		if v.State != "dead" || v.Price != 750 || !v.Button.Enabled || v.Button.Action != ActionRequest || v.Schema != ViewSchema {
			t.Fatalf("%s: %+v", class, v)
		}
		if p.Ready(in) {
			t.Fatalf("%s: ready without a request", class)
		}
		if _, err := p.Request(in); err != nil {
			t.Fatal(err)
		}
		if !p.Ready(in) || !p.Queued() {
			t.Fatalf("%s: not ready after the request", class)
		}
		if v := p.View(in); v.Button.Action != ActionCancel || v.Request == nil || v.Request.State != ReqQueued {
			t.Errorf("%s: %+v", class, v)
		}
		if _, err := p.Request(in); code(err) != "queued" {
			t.Errorf("%s: second request %v", class, err)
		}
		if !p.Start() {
			t.Fatal("start")
		}
		p.Finish(true, "", false)
		if v := p.View(in); v.Request.State != ReqDone || p.Queued() {
			t.Errorf("%s: %+v", class, v)
		}
	}
}

func TestPlayerReviveGoldShortSaysHowMuch(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		p, _ := newPR()
		// clvl 14: the revive reserve applies, so the bar is price + reserve
		in := reviveInputs(class, 14, 700)
		need := ReviveNeed(in.Ctx, 750)
		v := p.View(in)
		want := "need " + strconv.Itoa(need-700) + " more gold"
		if v.Button.Enabled || v.Button.Reason != want || v.Need != need {
			t.Fatalf("%s: %+v want %q", class, v, want)
		}
		if _, err := p.Request(in); code(err) != "state" || err.Error() != want {
			t.Errorf("%s: %v", class, err)
		}
		if p.Ready(in) {
			t.Error("ready without gold")
		}
	}
}

func TestPlayerReviveAliveNoneAndUnknownAreOff(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		p, _ := newPR()
		for name, c := range map[string]struct {
			st    MercState
			state string
			why   string
		}{
			"alive":   {MercState{Known: true, Owned: true, Alive: true, Level: 10}, "alive", "alive"},
			"none":    {MercState{Known: true}, "none", "no mercenary is hired"},
			"unknown": {MercState{}, "none", "not been read"},
		} {
			in := reviveInputs(class, 10, 9999)
			in.State = c.st
			v := p.View(in)
			if v.State != c.state || v.Button.Enabled || !strings.Contains(v.Button.Reason, c.why) {
				t.Errorf("%s %s: %+v", class, name, v)
			}
			if _, err := p.Request(in); err == nil {
				t.Errorf("%s %s: request accepted", class, name)
			}
		}
	}
}

func TestPlayerReviveCapKeepsTheButtonOff(t *testing.T) {
	p, _ := newPR()
	in := reviveInputs("Paladin", 10, 5000)
	in.Ctx.RevivesLastHour = in.Ctx.Knobs.ReviveMaxPerHour
	if v := p.View(in); v.Button.Enabled || !strings.Contains(v.Button.Reason, "limit") {
		t.Errorf("%+v", v)
	}
}

func TestPlayerReviveFailureKeepsTheReasonAndACooldown(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		p, now := newPR()
		in := reviveInputs(class, 10, 750)
		_, _ = p.Request(in)
		p.Start()
		p.Finish(false, "the hireling could not be reached", true)
		v := p.View(in)
		if v.Request.State != ReqFailed || v.Request.Reason != "the hireling could not be reached" || v.Button.Enabled || !strings.Contains(v.Button.Reason, "try again in 5 min") {
			t.Fatalf("%s: %+v", class, v)
		}
		if _, err := p.Request(in); err == nil {
			t.Errorf("%s: request inside the cooldown", class)
		}
		*now = now.Add(PlayerReviveCooldown + time.Second)
		if v := p.View(in); !v.Button.Enabled {
			t.Errorf("%s: still off after the cooldown: %+v", class, v)
		}
		if _, err := p.Request(in); err != nil {
			t.Errorf("%s: %v", class, err)
		}
	}
}

func TestPlayerReviveCancelPostponeAndNoCooldownForGold(t *testing.T) {
	p, _ := newPR()
	in := reviveInputs("Sorceress", 10, 750)
	if err := p.Cancel(); code(err) != "none" {
		t.Errorf("cancel with nothing: %v", err)
	}
	_, _ = p.Request(in)
	if err := p.Cancel(); err != nil {
		t.Fatal(err)
	}
	if p.Queued() || p.Ready(in) {
		t.Error("cancelled request still waits")
	}
	_, _ = p.Request(in)
	p.Start()
	if err := p.Cancel(); err == nil {
		t.Error("a running request was cancelled")
	}
	p.Postpone()
	if !p.Queued() {
		t.Error("postpone did not requeue")
	}
	p.Start()
	p.Finish(false, "gold", false) // gold is not a failure of the trip: no cooldown
	if _, err := p.Request(in); err != nil {
		t.Errorf("cooldown after a gold refusal: %v", err)
	}
}
