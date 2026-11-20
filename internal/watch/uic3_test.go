package watch

import (
	"os/exec"
	"testing"
)

// uic3: the tooltip, the right-click menu, the gestures and the feedback states of the character canvas, with the
// refusal rules (a refused requirement or action never looks done) and the placement at the eight sizes.
func TestUIC3TooltipMenuGestures(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	if out, err := exec.Command(node, "testdata/uic3/c3.js", "web/inventory-tip.js", "web/inventory-actions.js").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := exec.Command(node, "testdata/uic3/sheet.js", "web/inventory-canvas.js").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}
