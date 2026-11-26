package combat

import (
	"context"
	"errors"
	"time"

	"d2lod-bot/internal/d2"
)

// The reflex lane (Fable review 2026-11-25 P3 c, owner decision 2026-11-25): the bot's guard may send a belt drink itself
// while an action runs, without cancelling it. A potion key needs no stop and no panel; the old way (cancel, wait for the
// action to return, then drink) held the drink back for as long as the action took to notice the cancel.

// ErrNeedsAction is returned by ReflexDrink when the drink cannot be sent from the guard: no belt column of the kind
// (the pack path opens a panel) or a kind that is not a potion. The caller falls back to cancelling the action.
var ErrNeedsAction = errors.New("combat: drink needs the running action to stop")

// drunkAt is the last drink time of a kind.
func (f *Fighter) drunkAt(k PotionKind) (time.Time, bool) {
	f.dmu.Lock()
	defer f.dmu.Unlock()
	at, ok := f.lastDrink[k]
	return at, ok
}

// noteDrunk records a potion drunk now. A healing potion also starts its over-time effect.
func (f *Fighter) noteDrunk(k PotionKind, tier int, class d2.Class) {
	f.dmu.Lock()
	defer f.dmu.Unlock()
	now := f.now()
	f.lastDrink[k] = now
	if k == PotionHealing {
		f.heal = f.heal.drink(now, tier, class)
	}
}

// healNow is a copy of the working healing effect.
func (f *Fighter) healNow() healEffect {
	f.dmu.Lock()
	defer f.dmu.Unlock()
	return f.heal
}

// ReflexDrink sends the belt potion for a (DrinkLife or DrinkRejuv) and reports whether the safety action is handled:
// true after a key was sent and also when a potion is still working and nothing better is in the belt (the same
// "wait" the fighter's own drink makes), false with ErrNeedsAction when only the running action can do it. It serialises
// with the fighter's own drink, so two goroutines never drink the same potion twice.
func (f *Fighter) ReflexDrink(ctx context.Context, a SafetyAction) (bool, error) {
	if a != DrinkLife && a != DrinkRejuv {
		return false, ErrNeedsAction
	}
	if !f.sendMu.TryLock() {
		return true, nil // the fighter is drinking right now
	}
	defer f.sendMu.Unlock()
	err := f.drinkBelt(ctx, a)
	if errors.Is(err, errNoBeltColumn) {
		return false, ErrNeedsAction
	}
	return err == nil, err
}
