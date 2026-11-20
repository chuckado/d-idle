package plandata

import "fmt"

// ladderRunewords is the table of ladder-only runewords per class, used only while the knob items.ladderRunewords is on
// (ADR 0024, d.idle mode). Names only; the rune order, bases and mods stay in the game's Runes.txt (data.DB). Spirit:
// sword or shield; Insight: polearm or staff. U: which a class should prefer is a judgment (docs/sme/builds-gear.md).
var ladderRunewords = map[string][]string{
	"paladin":   {"Spirit", "Insight"},
	"sorceress": {"Spirit", "Insight"},
}

// LadderRunewords returns the ladder-only runewords a class's builds may socket when items.ladderRunewords is on.
// A class without a row is an error that names it: no fallback to another class.
func LadderRunewords(class string) ([]string, error) {
	words, ok := ladderRunewords[class]
	if !ok {
		return nil, fmt.Errorf("plandata: no ladder runewords row for class %q", class)
	}
	return append([]string(nil), words...), nil
}
