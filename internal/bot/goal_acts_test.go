package bot

import (
	"testing"

	"d2lod-bot/internal/d2"
)

// An earlier act is not reachable by walking: a hero in Act 2 or 3 is never planned into the Act 1 zones of the level plan
// (card qdmatrix). Class-free: the act is the only input; both classes are read the same way.
func TestAllowedActsDropsEarlierActs(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, c := range []struct {
			cfg  []int
			act  int // 0-based as the game stores it
			want map[int]bool
		}{
			{[]int{1}, 0, map[int]bool{1: true}},
			{[]int{1}, 1, map[int]bool{2: true}},
			{[]int{1, 2}, 1, map[int]bool{2: true}},
			{[]int{1, 2, 3}, 1, map[int]bool{2: true, 3: true}},
		} {
			r := &Runner{cfg: Config{Acts: c.cfg}}
			got := r.allowedActs(&d2.Snapshot{Player: &d2.Player{Class: class, Act: c.act}})
			if len(got) != len(c.want) {
				t.Errorf("%v cfg %v act %d: got %v, want %v", class, c.cfg, c.act+1, got, c.want)
				continue
			}
			for a := range c.want {
				if !got[a] {
					t.Errorf("%v cfg %v act %d: got %v, want %v", class, c.cfg, c.act+1, got, c.want)
				}
			}
		}
	}
}
