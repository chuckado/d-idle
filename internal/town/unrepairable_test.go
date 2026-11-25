package town_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/town"
)

// V live 2026-11-25 08:17 to 09:14 EST: a worn Hand Axe read 0% and the smith's Repair All never raised it ("repair not
// confirmed (durability 0% unchanged)"); the trip trigger read the same 0% again after four minutes, ten trips in an hour. The
// other pieces were repaired each time. Now the repair counts any worn item that rose, and an item left as it was stops counting.
// Two classes.
func TestUnrepairableItemDoesNotBringRepairTripsForever(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r := audit(t, 800, 1, true)
		r.W.Player.Class = class
		r.W.RepairBtn = town.RepairAll
		var worn []d2.Unit
		for _, u := range r.W.Items {
			if u.Item.Where == d2.ContEquipment {
				worn = append(worn, u)
			}
		}
		if len(worn) < 2 {
			t.Skipf("%v: the rig wears %d items", class, len(worn))
		}
		r.W.Stats[worn[0].Addr] = d2.Stats{72: 0, 73: 100}  // the stuck one
		r.W.Stats[worn[1].Addr] = d2.Stats{72: 30, 73: 100} // an ordinary worn piece
		r.W.NoRepair = map[uint32]bool{worn[0].ID: true}
		r.W.Add("tsc", d2.ContInventory, 0, 0) // no other trip reason hides the durability one
		s0, _ := r.W.Snapshot()
		if _, why := r.T.NeedsTrip(s0); !strings.Contains(why, "durability") {
			t.Fatalf("%v: the rig does not raise the repair trip to begin with: %q", class, why)
		}
		plan := town.ChorePlan{Repair: true, Force: true}
		if _, err := r.T.Chores(context.Background(), plan); err != nil {
			t.Fatalf("%v: %v", class, err)
		}
		r.W.Player.Gold += 5000          // and the purse grew since, as it did between the live trips
		r.Env.Clock.Sleep(2 * time.Hour) // the futile-reason quiet time is over, as it was between the live trips
		s, _ := r.W.Snapshot()
		if _, why := r.T.NeedsTrip(s); strings.Contains(why, "durability") {
			t.Fatalf("%v: still a repair trip reason after the stuck item was set aside: %q", class, why)
		}
	}
}
