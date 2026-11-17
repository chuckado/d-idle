package questboard

import questboarddata "d2lod-bot/data/questboard"

// Load parses and validates the embedded definitions (Act 1 and Act 2 today). A bad file is an error naming
// the field, never a partial board.
func Load() (*Board, error) {
	return Parse(questboarddata.Families, questboarddata.Act1, questboarddata.Act2)
}
