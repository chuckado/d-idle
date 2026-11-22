package town

import (
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// Stage 3 of town.plan: the purse target of a visit is the carry line plus the gold the plan spends; stages 0 to 2 keep
// the carry line alone. The two knobs bankFirst and withdrawShop do not matter at stage 3. Runs without game tables.
func TestGoldKeepPurseTarget(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		for _, tc := range []struct {
			stage, spend, want int
		}{{0, 400, 300 + 50*10}, {2, 400, 300 + 50*10}, {3, 0, 300 + 50*10}, {3, 400, 300 + 50*10 + 400}} {
			tune := env.DefaultTuning()
			tune.Town.Plan = tc.stage
			tune.Town.BankFirst = 0
			tn := &Town{e: &env.Env{Tune: &tune}, purseSpend: tc.spend}
			if got := tn.goldKeep(10); got != tc.want {
				t.Errorf("%v stage %d spend %d: keep %d, want %d", class, tc.stage, tc.spend, got, tc.want)
			}
		}
	}
}
