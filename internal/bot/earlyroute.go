package bot

import (
	"encoding/json"
	"fmt"
	"sync"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/task"
)

// The early route guard (user request 2026-11-25): a quest walk through the Underground Passage to the Dark Wood killed a clvl 4
// Paladin on the Questy bed (no potions, Vile Hunters at monster level 6 to 7). The legs and their floors are data
// (internal/plandata/earlyroutes.json): below the level of its class family or short of the kit a hero does not enter them, and the
// level plan farms elsewhere meanwhile. The owner's try anyway overrides, like the other route guards.

type earlyRoute struct {
	Name     string         `json:"name"`
	Areas    []int          `json:"areas"`
	MinLevel map[string]int `json:"minLevel"`
	MinLife  map[string]int `json:"minLife"` // max life by class family; 0 = no rule
	Kit      struct {
		Healing int `json:"healing"`
		Portal  int `json:"portal"`
	} `json:"kit"`
}

var (
	earlyOnce   sync.Once
	earlyRoutes []earlyRoute
	earlyErr    error
)

func loadEarlyRoutes() ([]earlyRoute, error) {
	earlyOnce.Do(func() {
		b, err := plandata.FS.ReadFile("earlyroutes.json")
		if err != nil {
			earlyErr = err
			return
		}
		var doc struct {
			Routes []earlyRoute `json:"routes"`
		}
		if earlyErr = json.Unmarshal(b, &doc); earlyErr == nil {
			earlyRoutes = doc.Routes
		}
	})
	return earlyRoutes, earlyErr
}

// earlyRouteAt is the route that holds area a, if any.
func earlyRouteAt(a int) (earlyRoute, bool) {
	rs, _ := loadEarlyRoutes()
	for _, r := range rs {
		for _, x := range r.Areas {
			if x == a {
				return r, true
			}
		}
	}
	return earlyRoute{}, false
}

// earlyRouteWhy says why area a is closed to the hero now ("" = open). Normal difficulty only.
func (r *Runner) earlyRouteWhy(s *d2.Snapshot, a int) string {
	why, _ := r.earlyRouteCheck(s, a)
	return why
}

// earlyRouteCheck is earlyRouteWhy that also says whether only the kit is missing (a chores trip fixes it).
func (r *Runner) earlyRouteCheck(s *d2.Snapshot, a int) (why string, kitOnly bool) {
	if r.corpseRun.Load() {
		return "", false // a corpse run is exempt: the corpse is worth more than the kit rule (Questy 2026-11-26: skipped for 1 potion, then lost)
	}
	rt, ok := earlyRouteAt(a)
	if !ok || s == nil || s.Player == nil || r.e.Tune.Route.EarlyRoute == 0 || s.Player.Difficulty != d2.Normal || task.TryActive() {
		return "", false
	}
	family := "melee"
	if s.Player.Class.Caster() {
		family = "caster"
	}
	if min := rt.MinLevel[family]; min > 0 && s.Player.Level < min {
		return fmt.Sprintf("level %d needed for the %s road (clvl %d)", min, rt.Name, s.Player.Level), false
	}
	if min := rt.MinLife[family]; min > 0 && s.Player.MaxLife > 0 && s.Player.MaxLife < min {
		return fmt.Sprintf("life %d needed for the %s road (max life %d)", min, rt.Name, s.Player.MaxLife), false
	}
	in := r.resInput(s)
	if n := healingOf(in); n < rt.Kit.Healing {
		return fmt.Sprintf("%d healing potions < %d for the %s road", n, rt.Kit.Healing, rt.Name), true
	}
	if rt.Kit.Portal > 0 && !in.HasTP {
		return "no town portal for the " + rt.Name + " road", true
	}
	return "", false
}
