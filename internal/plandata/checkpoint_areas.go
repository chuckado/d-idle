package plandata

// checkpointAreas are the areas where a hero that is healthy and alone near the entrance drops a checkpoint portal
// (card checkpoint1, docs/design/danger-protocol-2026-11-25.md): the dangerous stretches between two waypoints, ids
// from levels.txt (S: the 1.14d table as the nav and route-guard code use them; not re-read live here). Data, not code:
// add an id with the reason to extend it.
var checkpointAreas = map[int]string{
	8:  "Den of Evil",
	17: "Crypt",
	19: "Mausoleum",
	20: "Forgotten Tower",
	25: "Tower Cellar 5 (Countess)",
	28: "Barracks",
	29: "Jail Level 1",
	32: "Inner Cloister",
	33: "Cathedral",
	34: "Catacombs Level 1",
	35: "Catacombs Level 2",
	36: "Catacombs Level 3",
	37: "Catacombs Level 4 (Andariel)",
	38: "Tristram",
}

// CheckpointArea reports whether area id is a checkpoint area and why.
func CheckpointArea(id int) (string, bool) {
	w, ok := checkpointAreas[id]
	return w, ok
}
