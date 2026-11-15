package combat

import (
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/resource"
)

// eliteDmgMult scales the damage of a unique, champion or minion: extra strong and the other damage affixes are not
// read per monster, so the base row is tripled (U, deliberately high: it can only make the guard block more).
const eliteDmgMult = 3

// MonsterDps is the damage per second one monster deals when every swing lands (hit chance 1, no block, no
// resistance: an upper bound, U), from the monstats columns of the character's difficulty: the strongest of attack 1,
// attack 2, special 1 and the elemental slots, divided by the attack interval hitSec. ok is false when the row is
// unknown or carries no damage column, so the caller falls back to the count rule instead of reading zero as safe.
func MonsterDps(m *data.Monster, d d2.Difficulty, hitSec float64, elite bool) (dps float64, ok bool) {
	if m == nil || d < 0 || int(d) > 2 || hitSec <= 0 {
		return 0, false
	}
	hit := max(m.A1Max[d], m.A2Max[d], m.S1Max[d])
	for _, e := range m.Elem {
		hit = max(hit, e.Max)
	}
	if hit <= 0 {
		return 0, false
	}
	dps = float64(hit) / hitSec
	if elite {
		dps *= eliteDmgMult
	}
	return dps, true
}

// castGuard builds the portal exposure test of the current hostiles: the knobs and the summed damage of those in
// melee reach. A cast is a scroll or tome read: its time is the bot.portalCastMs knob (the Skills.txt frames of the
// scroll/tome row are not parsed by internal/data; R3 measures it), plus the walk into the portal.
func castGuard(bt env.BotTuning, ct env.CombatTuning, p *d2.Player, hs []Target) resource.CastGuard {
	g := resource.CastGuard{
		Mode:        bt.PortalGuard,
		ExposureSec: float64(bt.PortalCastMs+bt.PortalWalkMs) / 1000,
		LifeFrac:    float64(bt.PortalLifePct) / 100,
		DmgKnown:    true,
	}
	if g.Mode != 1 || p == nil {
		return g
	}
	for _, h := range hs {
		if h.Dist > float64(ct.CrowdAdjacent) {
			continue
		}
		dps, ok := MonsterDps(h.Mon, p.Difficulty, float64(bt.PortalHitMs)/1000, h.Elite || h.Boss)
		if !ok {
			g.DmgKnown = false
			return g
		}
		g.Incoming += dps
	}
	return g
}
