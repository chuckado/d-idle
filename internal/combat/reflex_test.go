package combat

import (
	"context"
	"errors"
	"sync"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/goals"
)

// ReflexDrink is the drink the bot's guard sends without stopping the running action: belt only, never the pack.
func TestReflexDrinkSendsTheBeltPotion(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		t.Run(class.String(), func(t *testing.T) {
			h := newHarness(t)
			h.s.class = class
			h.s.life, h.s.maxL = 30, 100
			h.s.stock(itHP1, 0, 2)
			ok, err := h.f.ReflexDrink(context.Background(), DrinkLife)
			if !ok || err != nil {
				t.Fatalf("handled %v err %v", ok, err)
			}
			if n := len(h.kinds(goals.EvPotion)); n != 1 {
				t.Fatalf("%d potion events, want 1", n)
			}
		})
	}
}

func TestReflexDrinkLeavesThePackToTheAction(t *testing.T) {
	h := newHarness(t)
	h.s.life, h.s.maxL = 30, 100 // an empty belt
	ok, err := h.f.ReflexDrink(context.Background(), DrinkLife)
	if ok || !errors.Is(err, ErrNeedsAction) {
		t.Fatalf("handled %v err %v, want ErrNeedsAction", ok, err)
	}
	if ok, err := h.f.ReflexDrink(context.Background(), Chicken); ok || !errors.Is(err, ErrNeedsAction) {
		t.Fatalf("a chicken is not a reflex: %v %v", ok, err)
	}
}

// The guard goroutine reads Safety while the action drinks: the shared drink state is behind one mutex (run with -race).
func TestReflexDrinkAndSafetyShareTheDrinkState(t *testing.T) {
	h := newHarness(t)
	h.s.life, h.s.maxL = 30, 100
	h.s.stock(itHP1, 0, 4)
	snap := h.s.snapshot
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); h.f.ReflexDrink(context.Background(), DrinkLife) }()
		go func() {
			defer wg.Done()
			if s, err := snap(); err == nil {
				h.f.Safety(s)
			}
		}()
	}
	wg.Wait()
}
