package danger

// Params are the model constants of the generator. Each has a tag; S-B fits the I and U ones against the fight logs.
type Params struct {
	FPS          float64 `json:"fps"`          // game frames per second (V: 25)
	AttackFrames float64 `json:"attackFrames"` // one attack plus the AI delay (I: boss1-trace section "Exposure model", 32 ticks)
	RefDefense   int     `json:"refDefense"`   // defence of the reference hero for Dps (U: the hero's defence is not read yet)
	ScreenTiles  float64 `json:"screenTiles"`  // tiles on one screen (U)
	// UnitsPerDensityGroup converts MonDen x mean pack size into units per clear, per Levels DrlgType (1 maze, 2 preset,
	// 3 outdoor). Fitted to the V run population of docs/research/areamon-table.csv (36 Act 1 areas, median per type):
	// outdoor 0.1735 (median error 20%), preset 0.0272 (18%), maze 0.0453 (55%: the maze fill depends on the rooms the
	// generator places, which the tables do not show). The sizes in Levels.txt are not the real areas (every outdoor row
	// says 80x80, every maze 200x200), so they are not used. A type missing here is an error.
	UnitsPerDensityGroup map[int]float64 `json:"unitsPerDensityGroup"`
	BadScreens           float64         `json:"badScreens"` // bad pack = typical + this many screens of units (I)
	FastRun              int             `json:"fastRun"`    // run velocity from which a monster is flagged fast (U)
	ClvlBands            []int           `json:"clvlBands"`  // clvl columns of Pay.XPByClvl (as areamon-table.csv)
}

// DefaultParams is the shipped parameter set.
func DefaultParams() Params {
	return Params{
		FPS: 25, AttackFrames: 32, RefDefense: 60, ScreenTiles: 280,
		UnitsPerDensityGroup: map[int]float64{1: 0.0453, 2: 0.0272, 3: 0.1735}, BadScreens: 1, FastRun: 10,
		ClvlBands: []int{1, 3, 5, 8, 12, 16},
	}
}
