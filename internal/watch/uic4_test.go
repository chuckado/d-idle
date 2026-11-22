package watch

import (
	"os"
	"os/exec"
	"strings"
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

// The page's first request waits for window load: the other scripts hear of the view by an event, and a view that came back before
// they had run left them undrawn (the half-drawn page the fit check found on a loaded server).
func TestInventoryFirstLoadWaitsForTheOtherScripts(t *testing.T) {
	b, err := os.ReadFile("web/inventory.js")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "window.addEventListener('load', start)") || strings.Contains(string(b), "\n  load(); artProbe();") {
		t.Error("inventory.js must start its first request from window load, after the page's other scripts have registered")
	}
}
