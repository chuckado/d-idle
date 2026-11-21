package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// MonsterPreset is one monster-type preset unit of the level's DS1 rooms in "debug levelmonsters" (private fork, K1 /
// ADR 0022). Kind is monster, superunique or place; a superunique's spawn point (the Countess in Tower Cellar 5) is
// one. Super is its SuperUniques row (-1 else), X and Y level cells, Room the room index, Up whether that room has
// come into play (its monsters exist only then; the preset is known from the level's build).
type MonsterPreset struct {
	Index int     `json:"index"`
	Kind  string  `json:"kind"`
	ID    int     `json:"id"`
	Super int     `json:"super"`
	Name  string  `json:"name"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Room  int     `json:"room"`
	Up    bool    `json:"up"`
}

// LevelMonster is a monster the engine has made so far on the level (the same table walk as the `monsters` verb).
type LevelMonster struct {
	ID      int     `json:"id"`
	Type    int     `json:"type"`
	Name    string  `json:"name"`
	Display string  `json:"display"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Boss    string  `json:"boss"`
	Super   int     `json:"super"`
	Mode    string  `json:"mode"`
	Life    int     `json:"life"`
	Alive   bool    `json:"alive"`
}

// LevelMonsters is the reply of "debug levelmonsters": a READ of what the engine already generated for the CURRENT
// level; it never makes a run CHEATED (ADR 0022). Rooms populate as the player nears them, so Monsters lacks those of
// rooms not Up; Presets does not.
type LevelMonsters struct {
	Level    int             `json:"level"`
	Unit     string          `json:"unit"`
	RoomsUp  []int           `json:"rooms_up"`
	Presets  []MonsterPreset `json:"presets"`
	Monsters []LevelMonster  `json:"monsters"`
}

// ErrNoLevelMonsters is returned when the reply is not the verb's JSON line.
var ErrNoLevelMonsters = errors.New("d2dctl: no levelmonsters reply")

// ParseLevelMonsters reads the JSON line of a "debug levelmonsters" reply.
func ParseLevelMonsters(r *Reply) (*LevelMonsters, error) {
	if r == nil || len(r.Lines) == 0 {
		return nil, ErrNoLevelMonsters
	}
	var lm LevelMonsters
	if err := json.Unmarshal([]byte(r.Lines[0]), &lm); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoLevelMonsters, err)
	}
	return &lm, nil
}

// LevelMonsters asks the fork for the current level's monster presets and made monsters (read-only). An "err" reply
// (stock d2d or a fork without the verb) comes back as the error; callers fall back to exploring.
func (c *Client) LevelMonsters(ctx context.Context) (*LevelMonsters, error) {
	r, err := c.Debug(ctx, "levelmonsters")
	if err != nil {
		return nil, err
	}
	return ParseLevelMonsters(r)
}
