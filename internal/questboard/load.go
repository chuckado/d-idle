package questboard

import questboarddata "d2lod-bot/data/questboard"

// Load parses and validates the embedded definitions (Acts 1 to 5). A bad file is an error naming
// the field, never a partial board.
func Load() (*Board, error) {
	return Parse(questboarddata.Families, questboarddata.Act1, questboarddata.Act2, questboarddata.Act3, questboarddata.Act4, questboarddata.Act5)
}
