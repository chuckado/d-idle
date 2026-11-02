package resource

import (
	"math"
	"testing"

	"d2lod-bot/internal/data"
)

func mean(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

// The stock scenarios use the traced Blood Moor population: mean life 4.2 (monster-stats-trace section 4).
func TestBloodMoorLifeMatchesTrace(t *testing.T) {
	if m := mean(BloodMoorLife); math.Abs(m-4.2)/4.2 > 0.03 {
		t.Errorf("stock life mean %.2f, trace 4.2", m)
	}
	db := &data.DB{MonLevels: []*data.MonLevel{nil, {Level: 1, HP: [3]int{100, 100, 100}}}}
	fallen := &data.Monster{HPMin: [3]int{1}, HPMax: [3]int{4}}
	rat := &data.Monster{HPMin: [3]int{1}, HPMax: [3]int{5}}
	zombie := &data.Monster{HPMin: [3]int{7}, HPMax: [3]int{12}}
	got := LifeCycle(db, data.Normal, 1, []*data.Monster{fallen, rat, zombie}, []int{10, 4, 4})
	if len(got) != 18 || math.Abs(mean(got)-mean(BloodMoorLife)) > 0.01 {
		t.Errorf("LifeCycle %v mean %.2f", got, mean(got))
	}
}

// B1: gold per kill in the stock scenario is the plain-kill model at monster level 1 (0.39), not 6.
func TestSimGoldFollowsModel(t *testing.T) {
	w := &world{sc: BloodMoorSorc(), mHP: []float64{1}}
	for i := 0; i < 1000; i++ {
		w.mHP = []float64{1}
		w.hit(5)
	}
	if want := 1000 * data.GoldPerKill(data.KillPlain, 1); math.Abs(float64(w.gold)-want) > 1 || w.gold > 400 {
		t.Errorf("1000 kills gave %d gold, model %.0f", w.gold, want)
	}
	deep := &world{sc: Scenario{MonLevel: 5}}
	for i := 0; i < 1000; i++ {
		deep.mHP = []float64{1}
		deep.hit(5)
	}
	if deep.gold < 5*w.gold {
		t.Errorf("gold must scale with monster level: %d vs %d", deep.gold, w.gold)
	}
}
