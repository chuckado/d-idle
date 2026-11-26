package arbiter

import "testing"

func TestDecisionExplained(t *testing.T) {
	if _, ok := (Decision{}).Explained(); ok {
		t.Fatal("empty decision explained")
	}
	d := Decision{
		Chosen:   &Intent{Source: SrcTrip, Kind: "town", Reason: "belt low"},
		Held:     true,
		Rejected: []Rejected{{Intent{Source: SrcQuest, Kind: "step"}, "lower class"}, {Intent{Source: SrcLevel, Kind: "farm"}, "needs area"}},
	}
	e, ok := d.Explained()
	if !ok || e.Intent != "trip/town" && e.Intent != SrcTrip+"/town" || e.Family != "chores" || e.Reason != "belt low" || !e.Held {
		t.Fatalf("explained %+v", e)
	}
	if len(e.Displaced) != 2 || e.Displaced[0].Intent != SrcQuest+"/step" || e.Displaced[0].Why != "lower class" {
		t.Fatalf("displaced %+v", e.Displaced)
	}
}
