package combat

import "time"

const tickMs = 40 // one game tick (25 per second)

// Cast lengths in 40 ms ticks, derived from the game's own animation rules (docs/fixcastgap.md section 2). A plain
// SC cast lasts ceil(N / speed) - 1 ticks and a sequence ceil(N / speed), speed = floor(256 * rate / 100), rate =
// min(100 + 120 FCR / (120 + FCR), 175); N is 14 frames for the Sorceress SC (V game.exe, oracle table at FCR
// 0 / 9 / 20 / 37 / 63 / 105 / 200 = 13 / 12 / 11 / 10 / 9 / 8 / 7) and 19 steps for the Lightning sequence (V).
const (
	scFrames  = 14 * 256
	seqFrames = 19 * 256
)

// seqSkills are the SQ skills whose seqtrans is SC; every skill in scSkills is a plain SC cast (Skills.txt, D).
var seqSkills = map[string]bool{"Lightning": true, "Chain Lightning": true}

var scSkills = map[string]bool{
	"Fire Bolt": true, "Charged Bolt": true, "Ice Bolt": true, "Frozen Armor": true, "Static Field": true,
	"Telekinesis": true, "Frost Nova": true, "Ice Blast": true, "Blaze": true, "Fire Ball": true, "Nova": true,
	"Shiver Armor": true, "Chilling Armor": true, "Fire Wall": true, "Enchant": true, "Teleport": true,
	"Glacial Spike": true, "Meteor": true, "Thunder Storm": true, "Energy Shield": true, "Blizzard": true,
}

// castTicks is the length of one cast of the named skill at the given faster cast rate, 0 for a skill the table does
// not know. game.exe (d2d false): the model above. d2d: Lightning and Chain Lightning are plain SC casts there (d2d
// ignores seqtrans, sklight D3, S), so they follow the SC row. THE ONE PLACE to change when upstream's
// fix/lightning-sequence lands in the lab binary: drop the `d2d` test below.
func castTicks(name string, fcr int, d2d bool) int {
	if fcr < 0 {
		fcr = 0
	}
	rate := 100 + 120*fcr/(120+fcr)
	if rate > 175 {
		rate = 175
	}
	speed := 256 * rate / 100
	switch {
	case seqSkills[name] && !d2d:
		return (seqFrames + speed - 1) / speed
	case scSkills[name] || seqSkills[name]:
		return (scFrames+speed-1)/speed - 1
	}
	return 0
}

// castGapFor is the minimum time between a cast of the named skill and the next command. A known skill gets its
// cast length plus one dispatch tick (d2d took 14 ticks for a 13 tick SC cast, V-d2d docs/d2d-casts-real.md) plus
// the margin, but never less than CastGapMs; an unknown skill, or CastPerSkill 0, gets CastGapMs.
func castGapFor(name string, castGapMs, margin, fcr int, perSkill, d2d bool) time.Duration {
	g := castGapMs
	if perSkill {
		if tk := castTicks(name, fcr, d2d); tk > 0 {
			if p := (tk+1)*tickMs + margin; p > g {
				g = p
			}
		}
	}
	return ms(g)
}

func (f *Fighter) castGap() time.Duration {
	t := f.tune()
	return castGapFor(f.lastCastName, t.CastGapMs, t.CastMarginMs, t.CastFCR, t.CastPerSkill != 0, f.D2D)
}
