package town

import (
	"testing"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

func bulkSnap(area int, cain bool, packItems int) *d2.Snapshot {
	s := &d2.Snapshot{Player: &d2.Player{Area: area}, Inventory: &d2.Inventory{}}
	for i := 0; i < packItems; i++ {
		s.Inventory.Items = append(s.Inventory.Items, d2.Unit{ID: uint32(i + 1), Item: &d2.Item{Where: d2.ContInventory}})
	}
	if cain {
		s.Monsters = []d2.Unit{{Class: cainClasses()[0], Mode: d2.MonModeStand}}
	}
	return s
}

func cainClasses() []int { return cain }

func needTown(policy int, rescued bool, journal ...uint32) *Town {
	tu := env.DefaultIdentifyTuning()
	tu.Policy = policy
	t := &Town{}
	t.SetBulkID(&BulkID{Tuning: func() env.IdentifyTuning { return tu }, Journal: &bulkid.Journal{IDs: journal},
		Rescued: func(*d2.Snapshot) bool { return rescued }})
	return t
}

// Must-fail: no trip (no stash opened, nothing moved) unless the policy is on AND Cain is rescued AND he stands here.
func TestNeedBulkOnlyWithRescuedCainHere(t *testing.T) {
	town := bulkSnap(1, true, 3)
	for _, c := range []struct {
		name string
		t    *Town
		s    *d2.Snapshot
		want bool
	}{
		{"cain-free, rescued, here", needTown(env.IdentifyCainFree, true), town, true},
		{"cain-free, not rescued", needTown(env.IdentifyCainFree, false), town, false},
		{"cain-free, rescued, absent", needTown(env.IdentifyCainFree, true), bulkSnap(1, false, 3), false},
		{"cain-free, rescued, in the field", needTown(env.IdentifyCainFree, true), bulkSnap(2, true, 3), false},
		{"manual", needTown(env.IdentifyManual, true), town, false},
		{"never", needTown(env.IdentifyNever, true), town, false},
		{"scrolls need no Cain", needTown(env.IdentifyScrolls, false), bulkSnap(1, false, 3), true},
		{"journal pending beats manual", needTown(env.IdentifyManual, false, 7), bulkSnap(1, false, 3), true},
		{"journal pending beats never", needTown(env.IdentifyNever, false, 7), bulkSnap(1, false, 3), true},
	} {
		if got := c.t.needBulk(c.s); got != c.want {
			t.Errorf("%s: need %v, want %v", c.name, got, c.want)
		}
	}
}

func TestNeedBulkIdleLatch(t *testing.T) {
	tw := needTown(env.IdentifyCainFree, true)
	s := bulkSnap(1, true, 3)
	tw.bulkIdle = packCount(s) // the last trip planned nothing for this pack
	if tw.needBulk(s) {
		t.Fatal("same pack, nothing planned: no second walk to the chest")
	}
	if !tw.needBulk(bulkSnap(1, true, 4)) {
		t.Fatal("a changed pack must be looked at again")
	}
}

func TestMeasuredRunForcesManual(t *testing.T) {
	tw := needTown(env.IdentifyCainFree, true)
	tw.bulk.Measured = func() bool { return true }
	if tw.needBulk(bulkSnap(1, true, 3)) {
		t.Fatal("a measured run never identifies in bulk")
	}
}
