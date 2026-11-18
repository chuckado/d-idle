package questboard

import (
	"strings"
	"testing"
	"time"
)

func (a *autoRig) deskAuto(paused *bool) *Desk {
	st, err := Open(a.b, a.dir, a.hero)
	if err != nil {
		a.t.Fatal(err)
	}
	st.now = a.c.Now
	return NewDesk(a.b, DeskConfig{
		Exec:   func() (*Exec, *Store) { return a.x, st },
		Paused: func() bool { return paused != nil && *paused },
		Class:  func() string { return a.f.Class },
		Level:  func() int { return a.f.Level },
		Now:    a.c.Now,
	})
}

func TestDeskSetAuto(t *testing.T) {
	for _, bd := range builds {
		t.Run(bd.class, func(t *testing.T) {
			a := newAutoRig(t, bd.class, 6, opts(20))
			d := a.deskAuto(nil)
			if v := d.View(); !v.AutoStart.Available || v.AutoStart.DelaySec != 20 {
				t.Fatalf("auto view %+v", v.AutoStart)
			}
			if r := d.SetAuto(den, true); !r.OK || r.Already || !strings.Contains(r.Reason, "on") {
				t.Fatalf("on %+v", r)
			}
			if r := d.SetAuto(den, true); !r.OK || !r.Already {
				t.Fatalf("again %+v", r)
			}
			v := d.viewFresh()
			var row QuestRow
			for _, q := range v.Quests {
				if q.ID == den {
					row = q
				}
			}
			if !row.Auto || !row.AutoOK {
				t.Fatalf("row %+v", row)
			}
			if r := d.SetAuto("search_for_cain", true); r.OK || r.Code != CodeAutoNever || r.Status != 409 {
				t.Fatalf("never %+v", r)
			}
			if r := d.SetAuto("nope", true); r.OK || r.Code != CodeUnknown {
				t.Fatalf("unknown %+v", r)
			}
			if r := d.SetAuto(den, false); !r.OK || r.Already {
				t.Fatalf("off %+v", r)
			}
			if a.plan().AutoStart(0, den) {
				t.Fatal("flag stayed")
			}
		})
	}
}

func TestDeskCountdownCancelAndNeeds(t *testing.T) {
	a := newAutoRig(t, "Paladin", 6, opts(20))
	d := a.deskAuto(nil)
	// an offer without the flag: a Needs row with the word as its chip
	v := d.viewFresh()
	var offer *NeedRow
	for i := range v.Needs {
		if v.Needs[i].Kind == "quest_offer" {
			offer = &v.Needs[i]
		}
	}
	if offer == nil || offer.Sub != "safe" || offer.Actions[0].Label != "Start" || offer.Actions[0].Post != "/api/quests/"+den+"/start" || !strings.Contains(offer.Text, "ready") {
		t.Fatalf("needs %+v", v.Needs)
	}
	d.SetAuto(den, true)
	a.tickAuto(calm)
	v = d.viewFresh()
	c := v.AutoStart.Countdown
	if c == nil || c.Quest != den || c.SecondsLeft != 20 {
		t.Fatalf("countdown %+v", c)
	}
	var cancel *NeedRow
	for i := range v.Needs {
		if v.Needs[i].Kind == "quest_offer" {
			t.Fatal("offer row while a countdown runs")
		}
		if v.Needs[i].Kind == "quest_auto" {
			cancel = &v.Needs[i]
		}
	}
	if cancel == nil || cancel.Actions[0].Label != "Cancel" || cancel.Actions[0].Post != "/api/quests/"+den+"/cancel" {
		t.Fatalf("needs %+v", v.Needs)
	}
	a.c.add(8 * time.Second)
	if left := d.viewFresh().AutoStart.Countdown.SecondsLeft; left != 12 {
		t.Fatalf("left %d", left)
	}
	if r := d.Cancel(den, ""); !r.OK || r.Already || !strings.Contains(r.Reason, "will not start by itself") {
		t.Fatalf("cancel %+v", r)
	}
	if v = d.viewFresh(); v.AutoStart.Countdown != nil {
		t.Fatal("countdown stayed")
	}
	a.c.add(time.Minute)
	if act := a.tickAuto(calm); act.Kind != AutoNone {
		t.Fatalf("started after cancel: %+v", act)
	}
	if !d.viewFresh().Quests[0].Auto {
		t.Log("flag per quest row")
	}
}

func TestDeskOfferNeedsRowFollowsTheOffer(t *testing.T) {
	a := newAutoRig(t, "Sorceress", 6, opts(0))
	d := a.deskAuto(nil)
	has := func(v View) bool {
		for _, n := range v.Needs {
			if n.Kind == "quest_offer" {
				return true
			}
		}
		return false
	}
	if !has(d.viewFresh()) {
		t.Fatal("no offer row")
	}
	if r := d.Start(den, ""); !r.OK {
		t.Fatalf("%+v", r)
	}
	if has(d.viewFresh()) {
		t.Fatal("offer row after Start")
	}
	d.Cancel(den, "")
	paused := true
	dp := a.deskAuto(&paused)
	if has(dp.viewFresh()) {
		t.Fatal("offer row while paused")
	}
}

func TestDeskStepListAndLevelsToGo(t *testing.T) {
	for _, bd := range builds {
		t.Run(bd.class, func(t *testing.T) {
			a := newAutoRig(t, bd.class, 5, opts(0))
			d := a.deskAuto(nil)
			var br *OfferView
			v := d.viewFresh()
			for i := range v.Offers {
				if v.Offers[i].Quest == brq {
					br = &v.Offers[i]
				}
			}
			// Paladin floor 7 at clvl 5: risky, two levels to go; Sorceress floor 8 is too hard (no offer)
			if bd.burialSeen {
				if br == nil || br.LevelsToGo != 2 {
					t.Fatalf("offer %+v", br)
				}
			} else if br != nil {
				t.Fatalf("offer %+v", br)
			}
			a.f.Level = 12
			if v = d.viewFresh(); func() bool {
				for _, o := range v.Offers {
					if o.Readiness == ReadySafe && o.LevelsToGo != 0 {
						return true
					}
				}
				return false
			}() {
				t.Fatalf("levels to go on a safe offer: %+v", v.Offers)
			}
			if r := d.Start(brq, ""); !r.OK {
				t.Fatalf("start %+v", r)
			}
			run := d.viewFresh().Running
			c, _ := a.b.Chunk(brq, brc)
			if run == nil || run.StepCount != len(c.Steps) || run.StepNo != 1 || len(run.Steps) != len(c.Steps) || run.Steps[0].State != "now" || run.Steps[len(c.Steps)-1].State != "todo" {
				t.Fatalf("running %+v", run)
			}
			a.tick(1)
			if run = d.viewFresh().Running; run == nil || run.StepNo < 2 || run.Steps[0].State != "done" {
				t.Fatalf("after one step %+v", run)
			}
			if s := d.viewFresh().Strip; s == nil || s.StepCount != len(c.Steps) || s.StepNo < 2 {
				t.Fatalf("strip %+v", s)
			}
		})
	}
}

func TestDeskAwayCardAndDismiss(t *testing.T) {
	a := newAutoRig(t, "Paladin", 6, opts(0))
	d := a.deskAuto(nil)
	d.SetAuto(den, true)
	a.c.add(time.Hour) // away
	a.tickAuto(calm)
	a.tick(8)
	a.c.add(time.Hour)
	a.tickAuto(calm)
	waitEvents(t, a, 2)
	v := d.viewFresh() // the player is back
	if v.WhileAway == nil || len(v.WhileAway.Events) != 2 || v.WhileAway.Events[0].Text != "Started Den of Evil by itself" || v.WhileAway.Events[1].Kind != "handedIn" {
		t.Fatalf("card %+v", v.WhileAway)
	}
	if r := d.DismissAway(); !r.OK {
		t.Fatalf("%+v", r)
	}
	if d.viewFresh().WhileAway != nil {
		t.Fatal("card stayed")
	}
}

func TestDeskViewTouchesPresence(t *testing.T) {
	a := newAutoRig(t, "Sorceress", 6, opts(0))
	d := a.deskAuto(nil)
	a.c.add(11 * time.Minute)
	if !a.x.Away() {
		t.Fatal("should be away")
	}
	d.View()
	if a.x.Away() {
		t.Fatal("a GET did not count as presence")
	}
}

func TestDeskAutoKnobOff(t *testing.T) {
	a := newAutoRig(t, "Sorceress", 6, opts(0))
	a.x.SetEnabled(false)
	d := a.deskAuto(nil)
	v := d.viewFresh()
	if v.Enabled || v.AutoStart.Available || v.WhileAway != nil || len(v.Needs) != 0 {
		t.Fatalf("%+v", v)
	}
	if r := d.SetAuto(den, true); r.OK || r.Code != CodeOff {
		t.Fatalf("%+v", r)
	}
	if r := d.DismissAway(); r.OK || r.Code != CodeOff {
		t.Fatalf("%+v", r)
	}
}
