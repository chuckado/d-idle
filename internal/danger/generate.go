package danger

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"d2lod-bot/internal/data"
)

var elementNames = [6]string{"physical", "magic", "fire", "lightning", "cold", "poison"}

// Build derives the table from the game tables. Towns and areas without spawns are left out. A nil or incomplete db
// is an error; overrides (may be nil) are applied last and listed in Table.Overrides.
func Build(db *data.DB, p Params, ov *Overrides) (*Table, error) { return build(db, p, ov, bossesJSON) }

func build(db *data.DB, p Params, ov *Overrides, bossRaw []byte) (*Table, error) {
	if db == nil || len(db.Levels) == 0 || len(db.Monsters) == 0 {
		return nil, errors.New("danger: the generator needs the level and monster tables")
	}
	if p.FPS <= 0 || p.AttackFrames <= 0 {
		return nil, errors.New("danger: params need fps and attackFrames")
	}
	byKey := make(map[string]*data.Monster, len(db.Monsters))
	for _, m := range db.Monsters {
		byKey[strings.ToLower(m.Key)] = m
	}
	t := &Table{Schema: Schema, Params: p}
	for _, l := range db.Levels {
		for d := 0; d < 3; d++ {
			pool := l.Mon
			if d > 0 {
				pool = l.NMon
			}
			z, err := buildZone(db, byKey, l, pool, d, p)
			if err != nil {
				return nil, err
			}
			if z != nil {
				t.Zones = append(t.Zones, z)
			}
		}
	}
	if len(t.Zones) == 0 {
		return nil, errors.New("danger: no area has spawns in the level table")
	}
	bs, err := buildBosses(db, byKey, bossRaw)
	if err != nil {
		return nil, err
	}
	t.Bosses = bs
	for _, b := range bs {
		for _, z := range t.Zones {
			if z.Area == b.Area && z.Diff == b.Diff {
				z.Bosses = append(z.Bosses, b.Key)
			}
		}
	}
	if ov != nil {
		if err := ov.apply(t); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func scaleRow(db *data.DB, mlvl int) *data.MonLevel {
	if mlvl >= 0 && mlvl < len(db.MonLevels) {
		return db.MonLevels[mlvl]
	}
	return nil
}

func pct(v int, s *data.MonLevel, pick func(*data.MonLevel) [3]int, d int, noRatio bool) float64 {
	if s == nil || noRatio || pick(s)[d] == 0 {
		return float64(v)
	}
	return float64(v) * float64(pick(s)[d]) / 100
}

// stats returns life, defence, to-hit and mean attack damage of a monster at difficulty d on monster level mlvl.
func stats(db *data.DB, m *data.Monster, mlvl, d int) (life, def, th, dmg float64) {
	s := scaleRow(db, mlvl)
	hp := func(s *data.MonLevel) [3]int { return s.HP }
	life = pct((m.HPMin[d]+m.HPMax[d])/2, s, hp, d, m.NoRatio)
	def = pct(m.AC[d], s, func(s *data.MonLevel) [3]int { return s.AC }, d, m.NoRatio)
	var n, sum, thSum float64
	add := func(lo, hi, h int) {
		if hi <= 0 {
			return
		}
		n++
		sum += pct((lo+hi)/2, s, func(s *data.MonLevel) [3]int { return s.DM }, d, m.NoRatio)
		thSum += pct(h, s, func(s *data.MonLevel) [3]int { return s.TH }, d, m.NoRatio)
	}
	add(m.A1Min[d], m.A1Max[d], m.A1TH[d])
	add(m.A2Min[d], m.A2Max[d], m.A2TH[d])
	add(m.S1Min[d], m.S1Max[d], m.S1TH[d])
	if n == 0 {
		return life, def, 0, 0
	}
	dmg = sum / n
	for _, e := range m.Elem { // elemental on-hit damage rides on the attack at its percent chance
		if e.Pct > 0 && e.Max > 0 {
			dmg += float64(e.Pct) / 100 * float64(e.Min+e.Max) / 2
		}
	}
	return life, def, thSum / n, dmg
}

func hitChance(th, def float64, mlvl, clvl int) float64 {
	if th+def <= 0 {
		return 0.95
	}
	c := th / (th + def) * 2 * float64(mlvl) / float64(mlvl+clvl)
	return math.Min(0.95, math.Max(0.05, c))
}

func hazardsOf(m *data.Monster, d int, p Params) []string {
	var h []string
	if m.Ranged {
		h = append(h, "ranged")
	}
	if m.Run >= p.FastRun {
		h = append(h, "fast")
	}
	for _, e := range m.Elem {
		if e.Pct > 0 && e.Max > 0 && e.Type != "" {
			if !slices.Contains(h, strings.ToLower(e.Type)) {
				h = append(h, strings.ToLower(e.Type))
			}
		}
	}
	return h
}

func buildZone(db *data.DB, byKey map[string]*data.Monster, l *data.Level, pool []string, d int, p Params) (*Zone, error) {
	mlvl := l.MonLvl[d]
	var mons []*data.Monster
	for _, k := range pool {
		m := byKey[strings.ToLower(k)]
		if m == nil || !m.Enabled || m.NPC || !m.Killable || m.Boss {
			continue
		}
		mons = append(mons, m)
	}
	if len(mons) == 0 {
		return nil, nil
	}
	name := l.Name
	if name == "" {
		name = l.Key
	}
	z := &Zone{Area: l.ID, Name: name, Act: l.Act + 1, Diff: DiffNames[d], MLvl: mlvl,
		Source: map[string]string{"tables": TagV, "shares": TagI, "dps": TagI, "density": TagI, "encounters": TagI, "xpGap": TagU}}
	weights := make([]float64, len(mons))
	var wsum float64
	for i, m := range mons {
		weights[i] = math.Max(1, float64(m.MinGrp+m.MaxGrp)/2)
		wsum += weights[i]
	}
	period := p.AttackFrames / p.FPS
	var meleeShare, rangedShare, groupMean, xpMean float64
	for i, m := range mons {
		life, def, th, dmg := stats(db, m, mlvl, d)
		hc := hitChance(th, float64(p.RefDefense), mlvl, mlvl)
		mon := Monster{Key: m.Key, Share: weights[i] / wsum, Life: int(life), Def: int(def), ToHit: int(th), Dmg: round2(dmg),
			Dps: round2(dmg * hc / period), Reach: "melee", Speed: m.Run, Group: [2]int{m.MinGrp, m.MaxGrp}, Hazards: hazardsOf(m, d, p)}
		if m.Ranged {
			mon.Reach = "ranged"
			rangedShare += mon.Share
		} else {
			meleeShare += mon.Share
		}
		groupMean += mon.Share * math.Max(1, float64(m.MinGrp+m.MaxGrp)/2)
		xpMean += mon.Share * pct(m.Exp[d], scaleRow(db, mlvl), func(s *data.MonLevel) [3]int { return s.XP }, d, m.NoRatio)
		z.Monsters = append(z.Monsters, mon)
	}
	// density: units per clear from MonDen and the level's size, then the units on one screen.
	k, ok := p.UnitsPerDensityGroup[l.DrlgType]
	if !ok {
		return nil, fmt.Errorf("danger: area %d %q: no units factor for level type %d", l.ID, name, l.DrlgType)
	}
	tiles := float64(l.SizeX[d] * l.SizeY[d])
	z.Density.Units = round2(float64(l.MonDen[d]) * groupMean * k)
	if tiles > 0 {
		z.Density.PerScreen = round2(z.Density.Units / tiles * p.ScreenTiles)
	}
	typ := max(1, int(math.Round(groupMean)))
	bad := typ + int(math.Round(z.Density.PerScreen*p.BadScreens))
	z.Encounters.Typical = split(typ, rangedShare, meleeShare)
	z.Encounters.Bad = split(bad, rangedShare, meleeShare)
	z.Uniques = round2(float64(l.MonUMin[d]+l.MonUMax[d]) / 2)
	z.Encounters.Elite = Encounter{N: typ + 1, Melee: z.Encounters.Typical.Melee + 1, Ranged: z.Encounters.Typical.Ranged,
		Leader: "unique", Rate: round2(z.Uniques)}
	z.Pay.XPPerKill = round2(xpMean)
	z.Pay.XPByClvl = map[string]float64{}
	for _, c := range p.ClvlBands {
		z.Pay.XPByClvl[fmt.Sprintf("c%d", c)] = round2(xpMean * gapFactor(c-mlvl))
	}
	z.Immune, z.EliteImmune = immunities(mons, d)
	return z, nil
}

func split(n int, ranged, melee float64) Encounter {
	r := 0
	if ranged+melee > 0 {
		r = int(math.Round(float64(n) * ranged / (ranged + melee)))
	}
	return Encounter{N: n, Melee: n - r, Ranged: r}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// immunities lists elements some spawn is immune to, and the elements an elite enchant could make immune (fire, cold,
// lightning +75 or magic resistant +40; poison +40: docs/difficulty.md section 4, as readiness.BuildAreaTable).
func immunities(mons []*data.Monster, d int) (imm, elite []string) {
	im, el := map[int]bool{}, map[int]bool{}
	for _, m := range mons {
		for e := 0; e < len(elementNames); e++ {
			r := m.Res[d][e]
			switch {
			case r >= 100:
				im[e] = true
			case e >= int(data.Fire) && e <= int(data.Cold) && r+75 >= 100:
				el[e] = true
			case e == int(data.Poison) && r+40 >= 100:
				el[e] = true
			}
		}
	}
	for e := 0; e < len(elementNames); e++ {
		if im[e] {
			imm = append(imm, elementNames[e])
		}
		if el[e] {
			elite = append(elite, elementNames[e])
		}
	}
	return imm, elite
}
