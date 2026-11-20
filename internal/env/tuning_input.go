package env

// InputTuning holds the input-layer knobs (docs/design/fable-review-2026-11-19/07-input-layer.md; planner in
// internal/input). Owner: the inputlayer card.
type InputTuning struct {
	// Planner: 1 lets the input planner choose verb > key > verified click for the actions migrated to it (right and
	// left skill selection). 0 (default) keeps the legacy path byte for byte.
	Planner int
}

func DefaultInputTuning() InputTuning { return InputTuning{Planner: 0} }

func inputKnobs() []Knob {
	return []Knob{
		{Key: "input.planner", Label: "Input planner", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: skill selection (right hand and aura, left hand) runs through the input planner: the fork's select verb when the connected d2d advertises input.select in caps, else the bound key, else the popup click, and every path is verified by reading the hand back; the path used is counted in the minute line as input=verb/key/click. 0 (default): the legacy popup path, unchanged. Off until the Proving Ground scenario and a live check pass on the devp14 image (docs/process/sandbox-devp14.md; question Q13 of the Fable review is open)."},
	}
}
