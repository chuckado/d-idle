package nav

import "errors"

// Sync reads the game once and advances the map tracker, so the planner sees the level and the rooms around the
// player's CURRENT position. Walks do this on every step; PathTo is pure planning and does not (it runs from the
// guard goroutine too, and the tracker is single-goroutine). A caller that planned before it ever walked after a
// position jump (debug goto, town portal arrival, waypoint arrival, a restart) calls Sync first: otherwise the grid
// is the old level's or empty and every point looks unreachable (DRILLS5 F4). Acting goroutine only.
func (n *Navigator) Sync() error {
	_, err := n.observe()
	if errors.Is(err, errLoading) {
		return ErrNotInGame
	}
	return err
}
