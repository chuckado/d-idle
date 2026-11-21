package env

// InputTuning holds the input-layer knobs (docs/design/fable-review-2026-11-19/07-input-layer.md; planner in
// internal/input). Owner: the inputlayer card.
type InputTuning struct {
	// Planner: 1 lets the input planner choose verb > key > verified click for the actions migrated to it (right and
	// left skill selection). 0 (default) keeps the legacy path byte for byte.
	Planner int
	// HotkeyMap: 1 writes the hotkey map generated from the planner's slots into the hero's hotkey table once a game
	// starts, then reads it back (internal/hotkeymap). 0 (default): the hotkeys stay as they are.
	HotkeyMap int
}

func DefaultInputTuning() InputTuning { return InputTuning{Planner: 0, HotkeyMap: 0} }

func inputKnobs() []Knob {
	return []Knob{
		{Key: "input.planner", Label: "Input planner", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: skill selection (right hand and aura, left hand) runs through the input planner: the fork's select verb when the connected d2d advertises input.select in caps, else the bound key, else the popup click, and every path is verified by reading the hand back; the path used is counted in the minute line as input=verb/key/click. 0 (default): the legacy popup path, unchanged. Off until the Proving Ground scenario and a live check pass on the devp14 image (docs/process/sandbox-devp14.md; question Q13 of the Fable review is open)."},
		{Key: "input.hotkeyMap", Label: "Generated hotkey map", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: when a game starts the bot derives the hero's skill hotkey map from the planner's slots (main attack, second element, movement, buffs in upkeep order, aura, summons, town portal), writes it into the game's own hotkey table through the skill popup (F1 to F8) and reads the table back from the game state, logging any slot that does not match (docs/design/fable-review-2026-11-19/07-input-layer.md 4.3, I7). Never writes a save file. 0 (default): the hotkeys are left alone. U: not verified live."},
	}
}
