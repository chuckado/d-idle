package env

import "testing"

// Questy bed 2026-11-25: a fresh clvl 8 Paladin with no locked build had 69 max life and every stat point unspent, because the
// default banked the points for a player who was not there. The default is the bot spending from its plan table; a locked
// build's desk still owns them (cmd/d2bot/pointswire.go) and 0 banks them again.
func TestPointsAreSpentByDefault(t *testing.T) {
	if !DefaultResourceTuning().BotSpendsPoints() {
		t.Fatal("the default leaves points to a player who is not asked")
	}
	r := DefaultResourceTuning()
	r.PointsDecisions = 0
	if r.BotSpendsPoints() {
		t.Fatal("0 must bank the points for the player")
	}
}
