package resource

import "fmt"

// CastGuard decides whether a town portal cast (scroll or tome) is a safe escape right now (card portalguard,
// refines survival reflex R6). V run, game.exe (docs/research/decomp28-trace.md Q5): a hit, block or knock-back does
// NOT interrupt a Town Portal cast, so the old reason "melee cuts the cast" is false. The real risk is the damage
// taken during the cast and the walk into the portal.
//
// Mode 0: the old count rule (CastAdjacentMax or more adjacent attackers block the cast).
// Mode 1: expected damage over the exposure against a fraction of current life. It falls back to the count rule
// whenever a term is missing (no damage terms for an adjacent monster, no exposure, no life), so it never allows a
// cast the old rule blocked unless every adjacent attacker has a damage term and the sum is shown safe.
type CastGuard struct {
	Mode        int     // bot.portalGuard
	ExposureSec float64 // cast time plus the walk into the portal (bot.portalCastMs + bot.portalWalkMs), U until drill R3
	LifeFrac    float64 // block when expected damage reaches this fraction of current life (bot.portalLifePct / 100)
	Incoming    float64 // summed damage per second of the adjacent attackers; only read when DmgKnown
	DmgKnown    bool    // every adjacent attacker had a damage term
}

// Blocked reports whether the cast must not be tried, with the reason for the log and the status sentence.
func (g CastGuard) Blocked(adjacent, life int) (bool, string) {
	old := adjacent >= CastAdjacentMax
	if g.Mode != 1 || !g.DmgKnown || g.ExposureSec <= 0 || g.LifeFrac <= 0 || life <= 0 {
		if old {
			return true, fmt.Sprintf("%d adjacent attackers (count rule, damage terms unknown)", adjacent)
		}
		return false, ""
	}
	exp := g.Incoming * g.ExposureSec
	lim := float64(life) * g.LifeFrac
	if exp >= lim {
		return true, fmt.Sprintf("expected damage %.0f over %.1f s reaches %.0f (%.0f%% of life %d), %d adjacent",
			exp, g.ExposureSec, lim, g.LifeFrac*100, life, adjacent)
	}
	return false, ""
}
