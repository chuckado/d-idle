package bot

import "testing"

func TestPointsGate(t *testing.T) {
	r := &Runner{}
	if !r.botMaySpendPoints() {
		t.Fatal("no gate: the bot spends as before")
	}
	on := false
	r.SetPointsGate(func() bool { return on })
	if r.botMaySpendPoints() {
		t.Error("gate closed: the bot banks")
	}
	on = true
	if !r.botMaySpendPoints() {
		t.Error("autopilot on: the bot spends")
	}
	r.SetPointsGate(nil)
	on = false
	if !r.botMaySpendPoints() {
		t.Error("gate removed")
	}
}
