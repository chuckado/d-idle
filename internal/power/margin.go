package power

import (
	"fmt"
	"math"
	"slices"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/resource"
)

// Mon is one attacker of a reference pack in the numbers Margin uses.
type Mon struct {
	Life, Def, ToHit, Dmg float64
	Ranged                bool
	// Unblockable is the damage share (0 to 1) of this attacker that a shield cannot block: a missile with no
	// physical record (Fallen Shaman fire, Radament, vampires; danger.Monster.Unblockable). 0 for melee and arrows.
	Unblockable float64
	Hazards     []string
}

// Pack is a reference encounter: monsters killed in order, melee first (they are nearest), then the ranged ones.
type Pack struct {
	MLvl int
	Mons []Mon
}

// Model ties the data file to the danger generator's constants.
type Model struct {
	P  *Profiles
	DP danger.Params
}

// NewModel loads the shipped profiles.
func NewModel(dp danger.Params) (*Model, error) {
	p, err := Load()
	if err != nil {
		return nil, err
	}
	return &Model{P: p, DP: dp}, nil
}

// Terms are the numbers behind one margin, for the log and the API (design 8.1).
type Terms struct {
	Budget      float64 `json:"budget"`      // life pool + usable potions + sustain over the fight
	Taken       float64 `json:"taken"`       // life the encounter takes while the hero kills it
	KillSeconds float64 `json:"killSeconds"` // time to kill the whole pack
	PerKill     float64 `json:"perKill"`     // seconds per monster
	Potions     float64 `json:"potions"`     // healing potions the fight is expected to cost
}

// Result is the margin of one encounter: Best uses what is known, Low discounts every term that is not measured.
type Result struct {
	Best  float64  `json:"best"`
	Low   float64  `json:"low"`
	Terms Terms    `json:"terms"`
	Flags []string `json:"flags,omitempty"` // "<term>_<quality>" for each term below measured that this encounter depends on
}

// Margin is the pure function of design 8.2: life the hero can spend divided by the life the pack takes while the hero
// kills it. An unreadable max life is an error: there is no pool to divide.
func (m *Model) Margin(h Hero, arch *Archetype, pk Pack) (Result, error) {
	if h.MaxLife <= 0 {
		return Result{}, fmt.Errorf("power: hero %q has no max life (not read)", h.Class)
	}
	if arch == nil {
		return Result{}, fmt.Errorf("power: no archetype")
	}
	best := m.margin(h, arch, pk, false)
	low := m.margin(h, arch, pk, true)
	best.Low = low.Best
	best.Flags = m.flags(h, arch)
	return best, nil
}

func (m *Model) flags(h Hero, arch *Archetype) []string {
	var f []string
	add := func(term string) {
		if q := h.q(term); q < Measured {
			f = append(f, term+"_"+q.String())
		}
	}
	if arch.Exposure.DefenceCounts {
		add(TermDefence)
		add(TermBlock)
	}
	add(TermOffence)
	add(TermResist)
	add(TermPotions)
	if h.Merc.Hired {
		add(TermMerc)
	}
	return f
}

// discount is the multiplier of a term in the low estimate.
func discount(a *Archetype, q Quality) float64 {
	switch q {
	case Measured:
		return 1
	case Derived:
		return a.Uncertainty.Derived
	}
	return a.Uncertainty.Default
}

func (m *Model) margin(h Hero, a *Archetype, pk Pack, low bool) Result {
	pr := m.P.Params
	d := func(term string) float64 {
		if !low {
			return 1
		}
		return discount(a, h.q(term))
	}
	var mel, ran []Mon
	for _, mo := range pk.Mons {
		if mo.Ranged {
			ran = append(ran, mo)
		} else {
			mel = append(mel, mo)
		}
	}
	order := append(slices.Clone(mel), ran...)
	period := m.DP.AttackFrames / m.DP.FPS

	// kill speed per monster, merc included
	qm := 0.0
	if h.Merc.Hired && h.Merc.Alive {
		lvl := h.Merc.Level
		if lvl == 0 {
			lvl = h.Level
		}
		qm = float64(lvl) * pr.MercDpsPerLevel * d(TermMerc)
	}
	heroDPS := func(mo Mon) float64 {
		switch {
		case h.MeasuredDPS > 0:
			return h.MeasuredDPS
		case h.DmgMax > 0:
			hit := float64(resource.HitChance(h.AR, 0, int(mo.Def), h.Level, pk.MLvl)) / 100
			mean := float64(h.DmgMin+h.DmgMax) / 2 * (1 + float64(h.DmgPct)/100)
			return hit * mean / pr.SwingSeconds * d(TermOffence)
		}
		return pr.DefaultDpsPerClvl * float64(h.Level) * d(TermOffence)
	}
	kill := make([]float64, len(order))
	var t0 float64
	for i, mo := range order {
		kill[i] = mo.Life / math.Max(heroDPS(mo)+qm, 1e-6)
		t0 += kill[i]
	}
	// a mana-gated attack slows down once the pool is empty (design 4: kill speed is the lesser of burst and mana income)
	if a.Offence.Mode == "mana_burst" && h.MaxMana > 0 && t0 > 0 {
		sec := float64(h.MaxMana) / pr.ManaPerSecond
		if t0 > sec {
			f := (sec + (t0-sec)*pr.ManaFloor) / t0
			for i := range kill {
				kill[i] /= f
			}
			t0 /= f
		}
	}

	// damage in
	holds := 0.0
	if h.Merc.Hired && h.Merc.Alive {
		holds = pr.MercHold * d(TermMerc)
	}
	holds = math.Min(0.9, holds+a.Exposure.ScreenShare)
	hitOn := func(mo Mon) float64 {
		def := float64(m.DP.RefDefense) // the reference hero of the generator when defence is not read
		if h.q(TermDefence) == Measured || h.Defence > 0 {
			def = float64(h.Defence)
		}
		if !a.Exposure.DefenceCounts {
			def = 0
		}
		c := 0.95
		if mo.ToHit+def > 0 {
			c = mo.ToHit / (mo.ToHit + def) * 2 * float64(pk.MLvl) / float64(pk.MLvl+h.Level)
		}
		c = math.Min(0.95, math.Max(0.05, c))
		if a.Exposure.DefenceCounts {
			c *= 1 - math.Min(75, float64(h.BlockPct))/100*(1-mo.Unblockable)
		}
		return c
	}
	dps := func(mo Mon) float64 {
		w := 1.0
		fr := 1.0
		for _, hz := range mo.Hazards {
			if v, ok := a.HazardWeights[hz]; ok {
				w = math.Max(w, v)
			}
			if i, ok := m.P.ElementIndex[hz]; ok && i < len(h.Resist) {
				res := math.Min(75, float64(h.Resist[i]))
				fr = math.Min(fr, 1-pr.ElemShare+pr.ElemShare*(1-res/100))
			}
		}
		return mo.Dmg * hitOn(mo) / period * w * fr
	}
	var taken, el float64
	for i := range order {
		var meleeIn, rangedIn float64
		nm := 0
		for _, o := range order[i:] {
			if o.Ranged {
				rangedIn += dps(o)
			} else {
				nm++
			}
		}
		if nm > 0 {
			reach := math.Min(float64(nm)*(1-holds), float64(a.Exposure.ContactCap))
			for _, o := range order[i:] {
				if !o.Ranged {
					meleeIn = dps(o) * reach
					break
				}
			}
		}
		t := kill[i]
		active := math.Max(0, el+t-math.Max(el, a.Exposure.CloseSeconds))
		taken += rangedIn*t + meleeIn*active
		el += t
	}
	if low {
		cut := d(TermResist)
		if a.Exposure.DefenceCounts {
			cut *= d(TermDefence)
		}
		taken /= cut
	}
	heal := pr.PotionHeal
	if h.PotionHeal > 0 {
		heal = h.PotionHeal
	}
	pool := float64(h.MaxLife) * (1 - pr.ChickenPct)
	potions := math.Min(float64(h.Potions)*heal*d(TermPotions), heal/pr.PotionEffectSeconds*el)
	budget := pool + potions + h.Sustain*el
	r := Result{Terms: Terms{Budget: budget, Taken: taken, KillSeconds: el}}
	if len(order) > 0 {
		r.Terms.PerKill = el / float64(len(order))
	}
	if over := taken - pool - h.Sustain*el; over > 0 {
		r.Terms.Potions = over / heal
	}
	r.Best = pr.MaxMargin
	if taken > 1e-9 {
		r.Best = math.Min(pr.MaxMargin, budget/taken)
	}
	return r
}
