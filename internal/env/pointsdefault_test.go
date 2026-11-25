package env

import "testing"

// User decision 2026-11-25: a fresh hero's points are the player's to place (docs/points.md); the default stays 0 and the
// Needs-you row tells how to hand them to the bot (a locked build or resource.pointsDecisions 1).
func TestPointsStayWithThePlayerByDefault(t *testing.T) {
	if DefaultResourceTuning().BotSpendsPoints() {
		t.Fatal("the bot spends points by default")
	}
	r := DefaultResourceTuning()
	r.PointsDecisions = 1
	if !r.BotSpendsPoints() {
		t.Fatal("1 must let the bot spend")
	}
}
