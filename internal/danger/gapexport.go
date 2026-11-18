package danger

// GapFactor is the level-gap penalty of a kill's XP for gap = clvl - monster level (see gapFactor for its sources: two
// points V, the rest U). Exported for the survival controller's XP per minute.
func GapFactor(gap int) float64 { return gapFactor(gap) }
