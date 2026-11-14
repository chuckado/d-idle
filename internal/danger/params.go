package danger

// Params are the model constants of the generator. Each has a tag; S-B fits the I and U ones against the fight logs.
type Params struct {
	FPS          float64 `json:"fps"`          // game frames per second (V: 25)
	AttackFrames float64 `json:"attackFrames"` // one attack plus the AI delay (I: boss1-trace section "Exposure model", 32 ticks)
	RefDefense   int     `json:"refDefense"`   // defence of the reference hero for Dps (U: the hero's defence is not read yet)
	ScreenTiles  float64 `json:"screenTiles"`  // tiles on one screen (U)
	// UnitsPerDensityTile converts MonDen x tiles of a level into units per clear (U: fitted to areamon-table.csv by
	// TestAreamonDiff; the real room fill depends on the rooms the level generator places).
	UnitsPerDensityTile float64 `json:"unitsPerDensityTile"`
	BadScreens          float64 `json:"badScreens"` // bad pack = typical + this many screens of units (I)
	FastRun             int     `json:"fastRun"`    // run velocity from which a monster is flagged fast (U)
	ClvlBands           []int   `json:"clvlBands"`  // clvl columns of Pay.XPByClvl (as areamon-table.csv)
}

// DefaultParams is the shipped parameter set.
func DefaultParams() Params {
	return Params{
		FPS: 25, AttackFrames: 32, RefDefense: 60, ScreenTiles: 280,
		UnitsPerDensityTile: 0.001, BadScreens: 1, FastRun: 10,
		ClvlBands: []int{1, 3, 5, 8, 12, 16},
	}
}
