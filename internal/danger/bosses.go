package danger

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"d2lod-bot/internal/data"
)

//go:embed bosses.json
var bossesJSON []byte

type bossRow struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	Monster   string   `json:"monster"`
	SuperUniq string   `json:"superunique"`
	Area      int      `json:"area"`
	Mods      []string `json:"mods"`
	DmgPct    int      `json:"dmgPct"`
	Minions   int      `json:"minions"`
}

type bossFile struct {
	Schema   string    `json:"schema"`
	Evidence string    `json:"evidence"`
	Bosses   []bossRow `json:"bosses"`
}

// Superuniques are +3 monster levels for damage and to-hit, x4 life and x5 experience; act bosses are x1. V run:
// docs/research/boss1-trace.md Q1 and deviation D4 (life, defence and experience are read at the MonStats level).
const (
	superLevelBonus = 3
	superLifeMult   = 4
	superXPMult     = 5
)

func buildBosses(db *data.DB, byKey map[string]*data.Monster, raw []byte) ([]*Boss, error) {
	var f bossFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("danger: bosses.json: %w", err)
	}
	var out []*Boss
	for _, r := range f.Bosses {
		class, kind := r.Monster, "act"
		if r.SuperUniq != "" {
			kind = "superunique"
			su := findSuper(db, r.SuperUniq)
			if su == nil {
				return nil, fmt.Errorf("danger: boss %q: superunique %q is not in the tables", r.Key, r.SuperUniq)
			}
			class = su.Class
		}
		m := byKey[strings.ToLower(class)]
		if m == nil {
			return nil, fmt.Errorf("danger: boss %q: monster %q is not in the tables", r.Key, class)
		}
		for d := 0; d < 3; d++ {
			lvl := m.Level[d]
			life, def, _, _ := stats(db, m, lvl, d)
			bonus, lifeMult, xpMult := 0, 1, 1
			if kind == "superunique" {
				bonus, lifeMult, xpMult = superLevelBonus, superLifeMult, superXPMult
			}
			_, _, th, dmg := stats(db, m, lvl+bonus, d)
			if r.DmgPct > 0 {
				dmg = dmg * float64(r.DmgPct) / 100
			}
			xp := pct(m.Exp[d], scaleRow(db, lvl), func(s *data.MonLevel) [3]int { return s.XP }, d, m.NoRatio)
			b := &Boss{Key: r.Key, Name: r.Name, Class: m.Key, Area: r.Area, Diff: DiffNames[d], Kind: kind,
				Level: lvl + bonus, Life: int(life) * lifeMult, Def: int(def), Dmg: round2(dmg), ToHit: int(th),
				XP: int(xp) * xpMult, Mods: r.Mods, Minions: r.Minions, Evidence: f.Evidence, Source: TagV + " tables, " + TagI + " mods"}
			b.Immune, _ = immunities([]*data.Monster{m}, d)
			out = append(out, b)
		}
	}
	return out, nil
}

func findSuper(db *data.DB, name string) *data.SuperUnique {
	for _, s := range db.SuperUniques {
		if strings.EqualFold(s.Key, name) || strings.EqualFold(s.Name, name) {
			return s
		}
	}
	return nil
}
