package main

import (
	"context"
	"testing"
)

// The scan visits the tombs in order from the Canyon and stops in the one with the orifice; a hero already in a tomb with the
// orifice stays. Must fail while the Canyon is searched for the orifice (V drill.act2.duriel: "object class 152 not found
// (the whole level was walked)" after 12 minutes in the Canyon).
func TestFindTrueTombStopsAtTheTombWithTheOrifice(t *testing.T) {
	for _, truth := range []int{66, 69, 72} {
		cur := canyonOfMagi
		var visits []int
		exit := func(_ context.Context, to int) error { cur = to; visits = append(visits, to); return nil }
		err := findTrueTomb(context.Background(), func() int { return cur }, exit, func() bool { return cur == truth })
		if err != nil || cur != truth {
			t.Fatalf("truth %d: err %v ended in %d", truth, err, cur)
		}
		want := 2*(truth-66) + 1 // each wrong tomb is entered and left, then the right one
		if len(visits) != want {
			t.Fatalf("truth %d: visits %v", truth, visits)
		}
	}
	cur := 68
	err := findTrueTomb(context.Background(), func() int { return cur }, func(context.Context, int) error { t.Fatal("moved"); return nil }, func() bool { return true })
	if err != nil {
		t.Fatalf("already in the tomb: %v", err)
	}
	cur = 40
	if err := findTrueTomb(context.Background(), func() int { return cur }, func(context.Context, int) error { return nil }, func() bool { return false }); err == nil {
		t.Fatal("a scan from town was accepted")
	}
}
