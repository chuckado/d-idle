package data

import (
	"math"
	"testing"
)

// Numbers from docs/research/treasure-trace.md sections 4 and 5.
func TestGoldModel(t *testing.T) {
	for _, tc := range []struct {
		k    KillKind
		ml   int
		want float64
	}{
		{KillPlain, 1, 0.39}, // Blood Moor
		{KillPlain, 2, 0.85},
		{KillPlain, 5, 2.23},    // Dark Wood
		{KillPlain, 67, 30.7},   // Hell Blood Moor
		{KillCaster, 1, 0.3375}, // 0.1125 piles x 3.0
		{KillChampion, 1, 9.0},  // 0.6 piles x 5 x 3.0
		{KillUnique, 10, 0},
	} {
		if got := GoldPerKill(tc.k, tc.ml); math.Abs(got-tc.want) > 0.02*math.Max(tc.want, 1) {
			t.Errorf("kind %d mlvl %d: %.3f want %.3f", tc.k, tc.ml, got, tc.want)
		}
	}
	if s := SupplyOf(KillSuperUnique); s.Items != 6 || SupplyOf(KillUnique).Items != 5 || SupplyOf(KillChampion).Items != 3 || SupplyOf(KillPlain).Items != 0.375 {
		t.Errorf("items %+v", s)
	}
}

func TestPileAtLeastAndMinWorth(t *testing.T) {
	for _, tc := range []struct {
		ml, n int
		want  float64
	}{{1, 10, 0}, {2, 10, 0.20}, {3, 10, 0.53}, {4, 10, 0.70}, {5, 10, 0.80}, {8, 10, 0.95}} {
		if got := PileAtLeast(tc.ml, tc.n); math.Abs(got-tc.want) > 0.01 {
			t.Errorf("mlvl %d >= %d: %.3f want %.2f", tc.ml, tc.n, got, tc.want)
		}
	}
	// the bar for 50 percent of the piles: never above the flat 10 once piles are big, and below 10 in the early areas
	for ml, want := range map[int]int{1: 3, 2: 7, 3: 10, 5: 17} {
		got := PileMinWorth(ml, 0.5)
		if want < 10 && got != want || want >= 10 && got < 10 {
			t.Errorf("mlvl %d: bar %d want %d", ml, got, want)
		}
		if min(got, 10) < 10 && PileAtLeast(ml, got) < 0.5 {
			t.Errorf("mlvl %d: bar %d lets %.2f through", ml, got, PileAtLeast(ml, got))
		}
	}
	if PileMinWorth(1, 0) != 0 {
		t.Error("share 0 must disable the bar")
	}
}
