package drill

import (
	"strings"
	"testing"
	"time"
)

func min(t time.Time, restarts int, xp, kills int64, area int) Minute {
	return Minute{Time: t, Area: area, Clvl: 3, XP: xp, Kills: kills, Restarts: restarts}
}

func facts() D10Facts {
	k := time.Date(2026, 10, 31, 14, 0, 0, 0, time.UTC)
	return D10Facts{KillAt: k, KilledPID: 42, SavedPre: 100, SavedEnd: 300, StagePre: "clean", StagePost: "clean", NewD2DAt: k.Add(20 * time.Second), NewD2DPID: 43,
		Minutes: []Minute{min(k.Add(-10*time.Second), 2, 500, 20, 2), min(k.Add(40*time.Second), 3, 150, 20, 2), min(k.Add(60*time.Second), 3, 400, 25, 2)}}
}

func TestEvalD10(t *testing.T) {
	if o := EvalD10(facts()); o.Verdict != Pass {
		t.Fatalf("good run: %s %s", o.Verdict, o.Reason)
	}
	cases := map[string]func(*D10Facts){
		"restarts counter": func(f *D10Facts) { f.Minutes[2].Restarts = 4 },
		"no restart":       func(f *D10Facts) { f.Minutes[1].Restarts, f.Minutes[2].Restarts = 2, 2 },
		"back in game after": func(f *D10Facts) {
			f.Minutes[1].Time = f.KillAt.Add(4 * time.Minute)
			f.Minutes[2].Time = f.KillAt.Add(5 * time.Minute)
		},
		"below the saved xp":         func(f *D10Facts) { f.Minutes[1].XP = 50 },
		"did not resume":             func(f *D10Facts) { f.Minutes[2].XP, f.Minutes[2].Kills = 150, 20 },
		"saved xp fell":              func(f *D10Facts) { f.SavedEnd = 99 },
		"P35":                        func(f *D10Facts) { f.StagePost = "open: inventory" },
		"still in the process table": func(f *D10Facts) { f.StillThere = true },
		"names the killed":           func(f *D10Facts) { f.Leftover = "orphan pid=42" },
		"no new d2d":                 func(f *D10Facts) { f.NewD2DAt = time.Time{} },
	}
	for want, mut := range cases {
		f := facts()
		mut(&f)
		if o := EvalD10(f); o.Verdict != Fail || !strings.Contains(o.Reason, want) {
			t.Errorf("%s: %s %s", want, o.Verdict, o.Reason)
		}
	}
	f := facts()
	f.Minutes = f.Minutes[1:]
	if o := EvalD10(f); o.Verdict != Blocked {
		t.Errorf("no pre-kill line: %s", o.Verdict)
	}
}

func TestParseMinute(t *testing.T) {
	m, ok := ParseMinute(`{"time":"2026-10-31T14:37:11.017-04:00","level":"INFO","msg":"minute","area":4,"clvl":9,"xp":55701,"kills":2110,"restarts":6,"xpRisk":5,"xpLost":0}`)
	if !ok || m.XP != 55701 || m.Restarts != 6 || m.Area != 4 || m.Time.IsZero() {
		t.Fatalf("%+v %v", m, ok)
	}
	if _, ok := ParseMinute(`{"msg":"pace"}`); ok {
		t.Error("pace is not a minute")
	}
}
