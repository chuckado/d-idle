package d2

// ResistRules is the player's resistance model in game.exe 1.14d order. Source: docs/research/decomp37-combat.md
// section 7 (B1) and model-ab-skresist.md Q10. The numbers are data: change the table, not the callers.
//
// Order (V read of FUN_0057be00, 0x57be3d..0x57bf02, V run model-ab-skresist Q10): the difficulty penalty is added to
// the raw sum FIRST, then the result is capped at BaseCap + the character's max-resist bonus (stat 0x26 for magic and
// the matching max stat of each element), then floored at Floor. A raw 90 in Nightmare is 50 (not 35).
type ResistRules struct {
	Penalty  [3]int // per Difficulty (Normal, Nightmare, Hell); V difficultylevels.tsv, expansion rows
	BaseCap  int    // 75 (V read)
	CapLimit int    // ceiling of the cap with max-resist bonuses (S: 95, the published limit; U, decomp37 did not trace it)
	Floor    int    // -100 (V, sk-resist-trace)
	// AbsorbCapPct is the cap on a percent absorb (V read FUN_0057bf10: 40).
	AbsorbCapPct int
	// FoldAbsorb makes the intake estimate apply the percent absorb. It only changes a number when the hero's absorb
	// stat is non-zero, so a hero without absorb gear is unaffected.
	FoldAbsorb bool
}

// ResistTable is the rule set the bot uses.
var ResistTable = ResistRules{
	Penalty:      [3]int{0, -40, -100},
	BaseCap:      75,
	CapLimit:     95,
	Floor:        -100,
	AbsorbCapPct: 40,
	FoldAbsorb:   true,
}

// PenaltyFor is the penalty of difficulty d; an unknown difficulty has none.
func (r ResistRules) PenaltyFor(d Difficulty) int {
	if d < 0 || int(d) >= len(r.Penalty) {
		return 0
	}
	return r.Penalty[d]
}

// Cap is the highest effective resist for a character with maxBonus points of max resist.
func (r ResistRules) Cap(maxBonus int) int { return min(r.BaseCap+max(maxBonus, 0), r.CapLimit) }

// Effective is the resist game.exe uses against a hit: penalty first, then the cap, then the floor.
// raw is the stat sum before the difficulty penalty (whether the stat read already holds it is U, docs/difficulty.md).
func (r ResistRules) Effective(raw, maxBonus int, d Difficulty) int {
	return max(min(raw+r.PenaltyFor(d), r.Cap(maxBonus)), r.Floor)
}

// Intake folds a percent absorb into an effective resist: the absorb is applied after the resistance (V run,
// decomp37 Q5), so the share of a hit taken is (100-res)/100 x (100-abs)/100, and the result is the resist that gives
// the same share. abs <= 0 or FoldAbsorb off returns res unchanged. The flat absorb (stats 143/145/147/149) needs
// the hit size and is not folded.
func (r ResistRules) Intake(res, absPct int) int {
	if !r.FoldAbsorb || absPct <= 0 {
		return res
	}
	absPct = min(absPct, r.AbsorbCapPct)
	return 100 - (100-res)*(100-absPct)/100
}
