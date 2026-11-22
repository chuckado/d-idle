package d2

import "fmt"

// Difficulty helpers for d2doctor and the logs (docs/difficulty.md P3 item 26; DIFFP2).

func (d Difficulty) String() string {
	switch d {
	case Normal:
		return "Normal"
	case Nightmare:
		return "Nightmare"
	case Hell:
		return "Hell"
	}
	return fmt.Sprintf("Difficulty(%d)", int(d))
}

// ResistPenalty is the player resist penalty of the difficulty: 0 / -40 / -100 (V: difficultylevels.tsv of the
// 1.14.3.71 install, see docs/difficulty.md section 1); the numbers live in ResistTable.
func (d Difficulty) ResistPenalty() int { return ResistTable.PenaltyFor(d) }

// DoctorLine is the one-line difficulty report for d2doctor: difficulty with the byte read, the resist penalty, the
// resists with the penalty applied (penalty first, then the cap: ResistRules) and, when areaMLvl is known (levels.tsv, V), the gap
// between the area monster level and the character. Whether the Resists read from the game already contain the
// penalty is U (check live with a Nightmare character).
func DoctorLine(p *Player, areaMLvl int) string {
	if p == nil {
		return "difficulty: no player"
	}
	pen := p.Difficulty.ResistPenalty()
	eff := func(v, bonus int) int { return ResistTable.Effective(v, bonus, p.Difficulty) }
	s := fmt.Sprintf("difficulty: %s (byte %d) penalty=%d level=%d res(fire/cold/ltng/pois)=%d/%d/%d/%d",
		p.Difficulty, int(p.Difficulty), pen, p.Level,
		eff(p.Resist.Fire, p.MaxResist.Fire), eff(p.Resist.Cold, p.MaxResist.Cold),
		eff(p.Resist.Lightning, p.MaxResist.Lightning), eff(p.Resist.Poison, p.MaxResist.Poison))
	if areaMLvl > 0 {
		s += fmt.Sprintf(" area=%d mlvl=%d gap=%d", p.Area, areaMLvl, areaMLvl-p.Level)
	}
	return s
}
