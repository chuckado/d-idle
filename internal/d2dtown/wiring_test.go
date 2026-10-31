package d2dtown

import (
	"testing"

	"d2lod-bot/internal/d2dsrc"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/town"
)

// DRILLS7 F11: the constructor cmd/d2bot and cmd/d2drill build the town input with must offer the walking verbs.
var _ town.WalkingInteractor = townInput{}

func TestInputIsWalkingInteractor(t *testing.T) {
	in := Input(nil, nil)
	w, ok := in.(town.WalkingInteractor)
	if !ok {
		t.Fatal("d2dtown.Input does not implement town.WalkingInteractor: enterPortal would pre-walk with GoTo")
	}
	if !w.InteractWalks(input.Target{Type: town.TypeObject, ID: d2dsrc.PortalIDBase}) {
		t.Error("portal interact must walk")
	}
	if !w.InteractWalks(input.Target{ID: d2dsrc.NPCIDBase + 5}) {
		t.Error("NPC interact must walk")
	}
	if _, ok := in.(interface{ BeltPut(uint32, int) error }); !ok {
		t.Error("belt put missing")
	}
}
