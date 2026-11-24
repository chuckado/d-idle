package nav

import (
	"testing"

	"d2lod-bot/internal/input"
)

// V live 2026-11-24 16:32 EST: the same refused click toward the stairs came back 215 times. From the second refusal the spot is blocked for planning, once. The rule reads no class.
func TestHeldClickBlocksTheSpotForPlanning(t *testing.T) {
	for _, name := range []string{"Paladin", "Necromancer"} {
		t.Run(name, func(t *testing.T) {
			_, n, o := gateNav(t, denAt(5, 10))
			var ex []input.Point // a read trigger area: the exit has no gated point (live "exits": null)
			to := denAt(7, 8)
			for i := 0; i < 4; i++ {
				from := denAt(5, 10)
				if _, err := n.hold(o.lm, from, to, ex, "inside a read trigger area"); err == nil {
					t.Fatalf("round %d: not refused", i)
				}
			}
			if len(n.hard) != 1 {
				t.Fatalf("%d planning blocks, want 1", len(n.hard))
			}
			if !n.hardNear(to, 4) {
				t.Fatal("the refused spot is not blocked")
			}
		})
	}
}
