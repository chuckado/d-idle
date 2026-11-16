package danger

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"slices"
)

//go:embed areaunits.json
var areaUnitsJSON []byte

// AreaUnits is the measured room population of game.exe (docs/survival/units-fit-2026-11-15.md): units and units per
// screen per area for one difficulty, and the per-type rate fit for areas it lacks. Only Normal, Acts 1 and 2 are
// measured; Nightmare, Hell and Acts 3 to 5 fall back to the Levels.txt formula for units.
type AreaUnits struct {
	Diff       string
	rows       map[int]areaRow
	fit        map[int]areaFit // by Levels DrlgType
	screenTile float64
}

type areaRow struct {
	Area      int     `json:"area"`
	DrlgType  int     `json:"drlgType"`
	MonDen    int     `json:"monDen"`
	Tiles     int     `json:"tiles"`
	UnitsMean float64 `json:"unitsMean"`
	PerScreen float64 `json:"perScreen"`
}

type areaFit struct{ a, b float64 }

type areaUnitsFile struct {
	Schema string `json:"schema"`
	Diff   string `json:"diff"`
	Fit    struct {
		Maze struct {
			A float64 `json:"lnUnitsPerTile"`
			B float64 `json:"lnDenCoef"`
		} `json:"maze"`
		Preset struct {
			A float64 `json:"lnUnitsPerTile"`
		} `json:"preset"`
		Outdoor struct {
			A float64 `json:"lnUnitsPerTile"`
		} `json:"outdoor"`
		ScreenTiles float64 `json:"screenTiles"`
	} `json:"fit"`
	Rows []areaRow `json:"rows"`
}

// LoadAreaUnits parses the shipped areaunits.json; an unknown schema or difficulty is an error.
func LoadAreaUnits() (*AreaUnits, error) {
	var f areaUnitsFile
	if err := json.Unmarshal(areaUnitsJSON, &f); err != nil {
		return nil, fmt.Errorf("danger: areaunits.json: %w", err)
	}
	if f.Schema != "danger.areaunits/1" {
		return nil, fmt.Errorf("danger: areaunits.json: unknown schema %q", f.Schema)
	}
	if !slices.Contains(DiffNames[:], f.Diff) {
		return nil, fmt.Errorf("danger: areaunits.json: unknown difficulty %q", f.Diff)
	}
	if f.Fit.ScreenTiles <= 0 {
		return nil, fmt.Errorf("danger: areaunits.json: no screenTiles")
	}
	u := &AreaUnits{Diff: f.Diff, rows: map[int]areaRow{}, screenTile: f.Fit.ScreenTiles,
		fit: map[int]areaFit{1: {f.Fit.Maze.A, f.Fit.Maze.B}, 2: {a: f.Fit.Preset.A}, 3: {a: f.Fit.Outdoor.A}}}
	for _, r := range f.Rows {
		u.rows[r.Area] = r
	}
	return u, nil
}

// Row returns the measured units and units per screen of an area at a difficulty; false when not measured.
func (u *AreaUnits) Row(area int, diff string) (units, perScreen float64, ok bool) {
	if u == nil || diff != u.Diff {
		return 0, 0, false
	}
	r, ok := u.rows[area]
	return r.UnitsMean, r.PerScreen, ok
}

// Rate is units per tile of the fit for a Levels DrlgType at a MonDen (the fit is the tiles model: every generated
// room is 8 x 8 tiles). An unknown type is an error.
func (u *AreaUnits) Rate(drlg, monDen int) (float64, error) {
	f, ok := u.fit[drlg]
	if !ok {
		return 0, fmt.Errorf("danger: areaunits: no fit for level type %d", drlg)
	}
	x := f.a
	if f.b != 0 {
		if monDen <= 0 {
			return 0, fmt.Errorf("danger: areaunits: level type %d needs MonDen over 0", drlg)
		}
		x += f.b * math.Log(float64(monDen))
	}
	return math.Exp(x), nil
}

// PerScreen is the units on one screen of an unmeasured area: the fit rate times the screen's tiles. It does not need
// the area's real size (Levels.txt has none), which is what made the old value 10 to 30 times low in mazes.
func (u *AreaUnits) PerScreen(drlg, monDen int) (float64, error) {
	r, err := u.Rate(drlg, monDen)
	return r * u.screenTile, err
}
