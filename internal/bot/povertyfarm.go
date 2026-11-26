package bot

import (
	"encoding/json"
	"fmt"
	"sync"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/task"
)

// The poverty farm (user request 2026-11-26): a hero with no healing potion and not enough gold for the kit died six times on the
// Questy bed in fields above what it could survive without potions. The data (plandata/povertyfarm.json) lists the farm areas with
// their measured gold per minute; the hero farms the best open one whose monster level is at most clvl minus noPotionGap.

type povertyFarmData struct {
	NoPotionGap int `json:"noPotionGap"`
	FromLevel   int `json:"noPotionFromLevel"`
	Farm        []struct {
		Area       int     `json:"area"`
		Name       string  `json:"name"`
		GoldPerMin float64 `json:"goldPerMin"`
	} `json:"farm"`
}

var (
	povOnce sync.Once
	povData povertyFarmData
)

func loadPovertyFarm() povertyFarmData {
	povOnce.Do(func() {
		if b, err := plandata.FS.ReadFile("povertyfarm.json"); err == nil {
			_ = json.Unmarshal(b, &povData)
		}
	})
	return povData
}

// noPotions: the hero carries no healing potion at all (belt and pack, rejuvenation counted).
func (r *Runner) noPotions(s *d2.Snapshot) bool {
	return s != nil && s.Player != nil && healingOf(r.resInput(s)) == 0
}

// povertyAreaWhy says why a hero with no potions may not enter area a ("" = fine): its monsters are above clvl minus the gap.
func (r *Runner) povertyAreaWhy(s *d2.Snapshot, a int) string {
	d := loadPovertyFarm()
	if d.NoPotionGap <= 0 || r.e.Tune.Route.EarlyRoute == 0 || !r.noPotions(s) || s.Player.Level < d.FromLevel || s.Player.Difficulty != d2.Normal || task.TryActive() ||
		r.corpseRun.Load() || a == 0 || r.d.Plan.IsTown(a) {
		return ""
	}
	if ml := r.areaMLvl(a, s.Player.Difficulty); ml > s.Player.Level-d.NoPotionGap {
		return fmt.Sprintf("no healing potions: monster level %d is above clvl %d minus %d", ml, s.Player.Level, d.NoPotionGap)
	}
	return ""
}

// povertyFarmArea is the best open farm area for a hero with no potions and no gold for them: the highest measured gold per
// minute among the data's areas that its level allows and that are neither cleared, banned nor guarded.
func (r *Runner) povertyFarmArea(s *d2.Snapshot) (int, bool) {
	d := loadPovertyFarm()
	if d.NoPotionGap <= 0 || !r.noPotions(s) || s.Player.Difficulty != d2.Normal || s.Player.Level < d.FromLevel {
		return 0, false
	}
	ex := map[int]bool{}
	for a := range r.loop.cleared {
		ex[a] = true
	}
	for a := range r.avoidedNow() {
		ex[a] = true
	}
	r.guardedAreas(s, ex)
	best, bestG := 0, -1.0
	for _, f := range d.Farm {
		if ex[f.Area] || r.povertyAreaWhy(s, f.Area) != "" || f.GoldPerMin <= bestG {
			continue
		}
		best, bestG = f.Area, f.GoldPerMin
	}
	return best, best != 0
}
