package nav

import "testing"

// There is no walk from the Valley of Snakes (45) to the Canyon of the Magi (46): 45 has one warp (to 58) and the Canyon
// stands alone in the layout (V lab 2026-11-23). A hint for it sent the Seven Tombs chunk to a dead end.
func TestNoHintToTheCanyon(t *testing.T) {
	for _, h := range act2Hints {
		if h[0] == 46 || h[1] == 46 {
			t.Errorf("hint %v leads to the Canyon of the Magi", h)
		}
	}
}
