package survival

import (
	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/power"
)

// Eco is what a zone pays per minute for the hero (design 5.1, 5.6): the supply model's kills per minute times the
// level-gap XP, gold per kill minus the potions a kill costs.
type Eco struct {
	KillsPerMin      float64
	XPPerMin         float64
	GoldPerMin       float64 // gold income, before potions
	PotionCostPerMin float64
	PotionsPerKill   float64
}

// Net is gold income minus the potions burned, per minute.
func (e Eco) Net() float64 { return e.GoldPerMin - e.PotionCostPerMin }

// Economy derives Eco from the zone and its typical-encounter margin. Kill time comes from the margin (so a hero with
// a better weapon farms faster), plus the walk between kills. Gold is the plain-kill table of internal/data (V for Blood
// Moor, S elsewhere).
func Economy(z *danger.Zone, zm power.ZoneMargin, clvl int, p *Params) Eco {
	secs := zm.Typical.Terms.PerKill + p.MoveSecondsPerKill
	var e Eco
	if secs <= 0 {
		return e
	}
	e.KillsPerMin = 60 / secs
	e.XPPerMin = z.Pay.XPPerKill * danger.GapFactor(clvl-z.MLvl) * e.KillsPerMin
	e.GoldPerMin = data.GoldPerKill(data.KillPlain, z.MLvl) * e.KillsPerMin
	if n := float64(z.Encounters.Typical.N); n > 0 {
		e.PotionsPerKill = zm.Typical.Terms.Potions / n
	}
	e.PotionCostPerMin = e.PotionsPerKill * e.KillsPerMin * p.PotionPrice
	return e
}
