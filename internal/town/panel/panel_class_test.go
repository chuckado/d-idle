package panel

import (
	"testing"

	"d2lod-bot/internal/d2"
)

func TestClassRowUnknownIsAnError(t *testing.T) {
	for c := d2.Amazon; c <= d2.Assassin; c++ {
		if r, err := classRow(c); err != nil || r == "" {
			t.Errorf("class %d: %q %v", int(c), r, err)
		}
	}
	if r, err := classRow(d2.Class(42)); err == nil || r != "" {
		t.Fatalf("unknown class got %q", r)
	}
	if _, err := NewLayout(nil, d2.Class(42), Pt{X: 800, Y: 600}, true); err == nil {
		t.Fatal("NewLayout must refuse an unknown class before reading data")
	}
}
