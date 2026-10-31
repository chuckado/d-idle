package town_test

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/gear/geartest"
)

// FIXPRICE B1, drill D3.r: a Short Staff at 1/20 durability cost 19 gold at Charsi (cost x 19/20 / 8); the old
// estimate (quarter of cost) said 39.
func TestFixPriceRepairEstimate(t *testing.T) {
	geartest.Need(t)
	for _, c := range []struct {
		name  string
		flags uint32
		stats d2.Stats
		want  int
	}{
		{"identified weapon pays cost/8 of the missing part", d2.FlagIdentified, d2.Stats{72: 1, 73: 20}, 19},
		{"unidentified cannot be repaired: floor of 1", 0, d2.Stats{72: 1, 73: 20}, 1},
		{"ethereal cannot be repaired: floor of 1", d2.FlagIdentified | d2.FlagEthereal, d2.Stats{72: 1, 73: 20}, 1},
		{"whole item needs nothing", d2.FlagIdentified, d2.Stats{72: 20, 73: 20}, 0},
	} {
		r := audit(t, 500, 5, true)
		u := r.W.Add("sst", d2.ContEquipment, 0, 0)
		u.Item.Body, u.Item.Flags = d2.BodyRHand, c.flags
		r.W.Stats[u.Addr] = c.stats
		for _, o := range r.W.Items { // the dagger of dressed() is not the item under test
			if o.Item.Where == d2.ContEquipment && o.Addr != u.Addr {
				r.W.Stats[o.Addr] = d2.Stats{72: 20, 73: 20}
			}
		}
		got, err := r.T.RepairEstimateForTest()
		if err != nil || got != c.want {
			t.Errorf("%s: got %d (%v) want %d", c.name, got, err, c.want)
		}
	}
}
