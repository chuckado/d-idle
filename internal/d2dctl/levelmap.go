package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// LayoutRoom is one room of "debug levelmap" (private fork, K1 / ADR 0022): X, Y, W, H in level cells. Def is the
// LvlPrest row of a preset room (the same number as the "room" rows of "debug presets"), -1 for a plain room.
type LayoutRoom struct {
	X      int  `json:"x"`
	Y      int  `json:"y"`
	W      int  `json:"w"`
	H      int  `json:"h"`
	Preset bool `json:"preset"`
	Def    int  `json:"def"`
}

// LayoutObject is one object of "debug levelmap". Class is the objects.txt id, X, Y are level cells, Kind one of
// shrine chest tree stone cage waypoint portal door object (from OperateFn), State the animation mode token.
type LayoutObject struct {
	Index  int     `json:"index"`
	Class  int     `json:"class"`
	Name   string  `json:"name"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Kind   string  `json:"kind"`
	State  string  `json:"state"`
	Hidden bool    `json:"hidden"`
}

// LayoutWarp is a level exit of "debug levelmap".
type LayoutWarp struct {
	Index int     `json:"index"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	To    int     `json:"to"`
}

// LayoutPortal is an open town portal of the player.
type LayoutPortal struct {
	Which int     `json:"which"`
	Level int     `json:"level"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

// LevelMap is the reply of "debug levelmap": what the engine already generated for the CURRENT level only. A READ:
// it changes nothing in the game and never makes a run CHEATED (ADR 0022). Tags: the object list equals the fork's own
// table walk (V, fork test tests/levelmap_d2d.py); the units (cells, as "debug presets") are U until a live capture.
type LevelMap struct {
	Level   int            `json:"level"`
	Unit    string         `json:"unit"`
	W       int            `json:"width"`
	H       int            `json:"height"`
	Rooms   []LayoutRoom   `json:"rooms"`
	Objects []LayoutObject `json:"objects"`
	Warps   []LayoutWarp   `json:"warps"`
	Portals []LayoutPortal `json:"portals"`
}

// ErrNoLevelMap is returned when the reply is not the verb's JSON line.
var ErrNoLevelMap = errors.New("d2dctl: no levelmap reply")

// ParseLevelMap reads the JSON line of a "debug levelmap" reply.
func ParseLevelMap(r *Reply) (*LevelMap, error) {
	if r == nil || len(r.Lines) == 0 {
		return nil, ErrNoLevelMap
	}
	var lm LevelMap
	if err := json.Unmarshal([]byte(r.Lines[0]), &lm); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoLevelMap, err)
	}
	return &lm, nil
}

// LevelMap asks the fork for the current level's room layout and objects (read-only). An "err" reply (stock d2d or a
// fork without the verb) comes back as the error; callers fall back to exploring.
func (c *Client) LevelMap(ctx context.Context) (*LevelMap, error) {
	r, err := c.Debug(ctx, "levelmap")
	if err != nil {
		return nil, err
	}
	return ParseLevelMap(r)
}
