package watch

import (
	"os"
	"os/exec"
	"testing"
)

// The area-map slot panel in a stub DOM: DOM goldens per state and the fit-whole rule (testdata/areamap/areamap_dom.js).
func TestAreaMapPanelDOMAndFit(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	cmd := exec.Command(node, "testdata/areamap/areamap_dom.js", "web/panels")
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}
