package task

import "testing"

func TestUnparkLiftsTheBlock(t *testing.T) {
	e := NewExecutor(nil, nil)
	e.blocked["den"] = "too weak"
	e.strikes["den"] = &strikeRec{parked: true}
	e.fails["quest.den/to-den"] = 3
	e.Unpark("den")
	if _, ok := e.Blocked("den"); ok || len(e.strikes) != 0 || len(e.fails) != 0 {
		t.Fatalf("blocked %v strikes %v fails %v", e.blocked, e.strikes, e.fails)
	}
}

// A chicken exit counts against its own, higher limit: three chickens in a minute do not park a quest (the Smith, live and
// lab 2026-11-23), eight do; three deaths still park it. Must fail without strikeChicken (it used the death limit).
func TestChickensHaveTheirOwnLimit(t *testing.T) {
	e := NewExecutor(nil, nil)
	for i := 1; i <= 7; i++ {
		if e.strikeChicken("tools", "chickened") {
			t.Fatalf("parked after %d chickens", i)
		}
	}
	if !e.strikeChicken("tools", "chickened") {
		t.Fatal("not parked after 8 chickens")
	}
	d := NewExecutor(nil, nil)
	for i := 1; i <= 2; i++ {
		d.strikeChicken("tools", "x")
		if d.strike("tools", "died") {
			t.Fatalf("death %d parked", i)
		}
	}
	if !d.strike("tools", "died") {
		t.Fatal("three deaths must park the group")
	}
	f := NewExecutor(nil, nil)
	f.Policy.MaxChickens = 2 // data, not a literal
	f.strikeChicken("tools", "x")
	if !f.strikeChicken("tools", "x") {
		t.Fatal("MaxChickens 2 parks at two")
	}
}
