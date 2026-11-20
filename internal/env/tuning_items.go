package env

// ItemsTuning holds the item-rule switches (ADR 0024). Owner: the town and loot planners.
type ItemsTuning struct {
	// LadderRunewords: 1 tells the planner the engine accepts the ladder-only runewords (the Runes.txt "server" rows:
	// Spirit, Insight, ...). Only true when the d2d fork runs with --ladder-runewords (d.idle mode); gates, soaks and
	// fidelity runs keep 0.
	LadderRunewords int
}

func DefaultItemsTuning() ItemsTuning { return ItemsTuning{} }

func itemsKnobs() []Knob {
	return []Knob{
		{Key: "items.ladderRunewords", Label: "Ladder-only runewords", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: the planner may socket the ladder-only runewords (Spirit, Insight and the rest of the Runes.txt server rows) and the Return screen shows 'ladder runewords: on'. Set it only together with the fork flag --ladder-runewords (ADR 0024, d.idle mode); read at start. 0 (default, vanilla): they stay unavailable, as in every gate and fidelity run."},
	}
}
