package questboard

import (
	"sort"
	"strings"
	"testing"
)

func changedKeys(a Reply) string {
	var s []string
	for _, c := range a.Changed {
		s = append(s, c.Quest+"/"+c.Chunk)
	}
	sort.Strings(s)
	return strings.Join(s, ",")
}

func findRow(v View, quest, chunk string) (ChunkRow, bool) {
	for _, q := range v.Quests {
		for _, c := range q.Chunks {
			if q.ID == quest && c.ID == chunk {
				return c, true
			}
		}
	}
	return ChunkRow{}, false
}

// D2: the Stop reply lists the chunk and every run that waited for it; Start the chain it accepted; both with a reason.
func TestDeskChangedAndReasons(t *testing.T) {
	const cain = "search_for_cain"
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := newRigSM(t, class, 17, true)
			d := r.desk(nil)
			st := d.Start(cain, "stones") // the chain: scroll, read, stones
			if !st.OK || changedKeys(st) != cain+"/read,"+cain+"/scroll,"+cain+"/stones" {
				t.Fatalf("start changed %q: %+v", changedKeys(st), st)
			}
			v := d.viewFresh()
			for _, c := range []string{"read", "stones"} {
				row, ok := findRow(v, cain, c)
				if !ok || row.Reason == nil || row.Reason.Code != "waits_for" || strings.Contains(row.Reason.Text, "{") {
					t.Fatalf("%s: want waits_for with a sentence, got %+v", c, row)
				}
			}
			if row, _ := findRow(v, cain, "scroll"); row.Reason != nil {
				t.Fatalf("the head moves, it has no reason: %+v", row.Reason)
			}
			stop := d.Cancel(cain, "scroll")
			if !stop.OK || stop.Already || changedKeys(stop) != cain+"/read,"+cain+"/scroll,"+cain+"/stones" {
				t.Fatalf("stop changed %q: %+v", changedKeys(stop), stop)
			}
			if again := d.Cancel(cain, "scroll"); !again.OK || !again.Already || len(again.Changed) != 0 {
				t.Fatalf("second stop: %+v", again)
			}
		})
	}
}

// A chunk without the machine keeps `changed` too (the stop lists the dependants of the old store).
func TestDeskChangedKnobOff(t *testing.T) {
	skipSM(t, "the machine is on for the whole suite")
	r := newRigSM(t, "Paladin", 17, false)
	d := r.desk(nil)
	if a := d.Start("search_for_cain", "scroll"); !a.OK || changedKeys(a) != "search_for_cain/scroll" {
		t.Fatalf("%+v", a)
	}
}

// Appendix B.1: a parked run says its stored code.
func TestDeskParkedRowReason(t *testing.T) {
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := parkRig(t, class, true)
			d := r.desk(nil)
			v := d.viewFresh()
			var found bool
			for _, q := range v.Quests {
				for _, c := range q.Chunks {
					if c.State == Blocked {
						found = true
						if c.Reason == nil || c.Reason.Code == "" || c.Reason.Text == "" {
							t.Fatalf("blocked row without reason: %+v", c)
						}
					}
				}
			}
			if !found {
				t.Fatalf("parkRig left nothing blocked: %+v", v.Quests)
			}
			if v.Blocked == nil || v.Blocked.Reason == nil {
				t.Fatalf("the blocked run carries no reason: %+v", v.Blocked)
			}
		})
	}
}

// C5: the game counts the field work done and no run exists. The row says handin_ready; Start creates the run, which
// then waits for the tap (handin_wanted) instead of answering "already started".
func TestDeskHandInReady(t *testing.T) {
	const q, c = "sisters_burial_grounds", "blood_raven"
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := newRigSM(t, class, 17, true)
			r.f.PendingBit = true
			ch, _ := r.b.Chunk(q, c)
			r.f.Progress[ChunkKey(q, c)] = len(ch.Steps)
			d := r.desk(nil)
			row, _ := findRow(d.viewFresh(), q, c)
			if row.State != Awaiting || row.Reason == nil || row.Reason.Code != CodeHandInReady {
				t.Fatalf("want awaiting + handin_ready: %+v", row)
			}
			a := d.Start(q, c)
			if !a.OK || a.Already || changedKeys(a) != q+"/"+c {
				t.Fatalf("start: %+v", a)
			}
			row, _ = findRow(d.viewFresh(), q, c)
			if row.State != Awaiting || row.Reason == nil || row.Reason.Code != "handin_wanted" {
				t.Fatalf("after start: %+v", row)
			}
			if again := d.Start(q, c); !again.OK || !again.Already {
				t.Fatalf("second start: %+v", again)
			}
		})
	}
}
