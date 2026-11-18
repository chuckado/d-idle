package danger

// gapFactor is the level-gap penalty of a kill's XP: gap = clvl - monster level. Gap 5 or less pays in full.
// Two points are V from areamon-table.csv (Blood Moor, mlvl 1: gap 7 pays 13.4 of 22 = 0.61, gap 11 pays 0.4 of 22 =
// 0.02); the rest of the curve is an interpolation (U) so XPByClvl is a guide, not a number a gate may use.
func gapFactor(gap int) float64 {
	table := []float64{1, 1, 1, 1, 1, 1, 0.80, 0.61, 0.42, 0.27, 0.12, 0.02}
	switch {
	case gap < 0:
		return 1 // below-level monsters pay full here; the real rule also trims big negative gaps (U, not used by the bands)
	case gap >= len(table):
		return 0.01
	}
	return table[gap]
}
