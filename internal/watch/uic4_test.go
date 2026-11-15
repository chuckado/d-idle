package watch

import (
	"os/exec"
	"testing"
)

// Review c2c3 R6, R7, R11 and R12 against the real scripts in one stub page: the pending outline and the TOWN mark are on the cells
// inventory.js builds, a request's answer repaints the grids, the confirm step of the menu focuses "Keep it", and a touch press or
// a drag never opens the menu. (Skipped without node; CI has it.)
func TestUIC4PendingMenuTouch(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	if out, err := exec.Command(node, "testdata/uic4/c4.js", "web/inventory.js", "web/inventory-actions.js").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}
