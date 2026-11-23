package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Run is a row of subtile cells y, x0..x1 inclusive, level-relative like the grid.
type Run [3]int

// WarpTrigger is one warp's trigger area as the engine's own predicates give it ("debug warptriggers", private fork,
// K1 / ADR 0022; read-only). A warp is taken by a CLICK inside its tile unit's LvlWarp Select box (Click: the subtile
// cells whose click takes it) and the walk ends within Reach of the unit (the cells from which that click is taken
// at once); stepping over the tile alone takes nothing. Truncated: a run touches the fork's scan edge (the area is
// larger than the scan, so the lists are not the whole set).
type WarpTrigger struct {
	Index     int    `json:"index"`
	To        int    `json:"to"`
	Unit      [2]int `json:"unit"`
	Select    [4]int `json:"select"`
	Click     []Run  `json:"click"`
	Reach     []Run  `json:"reach"`
	Truncated bool   `json:"truncated"`
}

// WarpTriggers is the reply of "debug warptriggers" for the CURRENT level; index i is warp i of "debug warps".
type WarpTriggers struct {
	Level     int           `json:"level"`
	Unit      string        `json:"unit"`
	Scan      int           `json:"scan"`
	ReachRule int           `json:"reach_rule"`
	Warps     []WarpTrigger `json:"warps"`
}

// ErrNoWarpTriggers is returned when the reply is not the verb's JSON line.
var ErrNoWarpTriggers = errors.New("d2dctl: no warptriggers reply")

// ParseWarpTriggers reads the JSON line of a "debug warptriggers" reply.
func ParseWarpTriggers(r *Reply) (*WarpTriggers, error) {
	if r == nil || len(r.Lines) == 0 {
		return nil, ErrNoWarpTriggers
	}
	var w WarpTriggers
	if err := json.Unmarshal([]byte(r.Lines[0]), &w); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoWarpTriggers, err)
	}
	return &w, nil
}

// WarpTriggers asks the fork for the trigger areas of the current level's warps. An "err" reply (stock d2d or a fork
// without the verb) comes back as the error; callers keep the guessed trigger bounds.
func (c *Client) WarpTriggers(ctx context.Context) (*WarpTriggers, error) {
	r, err := c.Debug(ctx, "warptriggers")
	if err != nil {
		return nil, err
	}
	return ParseWarpTriggers(r)
}
