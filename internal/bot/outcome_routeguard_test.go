package bot

import (
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
)

// The route guard hop and the guard exit's walk home refuse as typed outcomes (knob bot.typedOutcome); two classes.
// Must fail while routeGuardHop returns the plain gate error.
func TestRouteGuardHopIsTypedRefusal(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		g := campaignRig(t, &fakeQuests{}, 14)
		g.w.set(func(w *world) { w.area = 3; w.class = class })
		g.r.poll()
		s := g.r.latest()
		err := g.r.routeGuardHop(s, 25) // Forgotten Tower: guarded, no merc, clvl 14 < 15
		var ref *Refused
		if !errors.As(err, &ref) || !errors.Is(err, errAreaGated) || ref.Need != "route_closed" || ref.Wake == "" {
			t.Fatalf("%v: %v", class, err)
		}
		g.r.e.Tune.Bot.TypedOutcome = 0
		if err := g.r.routeGuardHop(s, 25); errors.As(err, &ref) || !errors.Is(err, errAreaGated) {
			t.Fatalf("%v knob off: %v", class, err)
		}
	}
}

func TestHomeWalkErrKeepsARefusalTyped(t *testing.T) {
	g := campaignRig(t, &fakeQuests{}, 14)
	g.r.poll()
	portal := errors.New("town: no town portal available")
	ref := g.r.refuse(errors.New("area not allowed now: x (3): route guard (y)"))
	var got *Refused
	if err := homeWalkErr(portal, ref); !errors.As(err, &got) {
		t.Fatalf("refusal lost: %v", err)
	}
	if err := homeWalkErr(portal, errors.New("no route")); errors.As(err, &got) || !errors.Is(err, portal) {
		t.Fatalf("a fault must stay joined with the portal error: %v", err)
	}
}
