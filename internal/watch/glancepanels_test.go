package watch

import (
	"encoding/json"
	"regexp"
	"testing"
)

// The cockpit's Needs you panel and the strip (progress, journey-strip, health), slice S3: DOM goldens and behaviour in a stub
// DOM (testdata/glancepanels/dom.js). Needs you: empty, one red, red and amber, five rows with the overflow line; progress:
// measuring, full, no hero; journey-strip: chapters 1, 3, 9, blocked, not reported; health: all ok, gear broken, two unknown.
func TestStripPanelDOMGoldens(t *testing.T) {
	runNode(t, "testdata/glancepanels/dom.js", "web/panels")
}

var stripPanels = []string{"needs-you", "progress", "journey-strip", "health"}

// The four panels keep the page rules (no timers, requests or queries of their own, text through textContent, no style
// writes) and declare themselves as opt-in cockpit slot panels reading the one /api/glance endpoint.
func TestStripPanelsFollowTheContract(t *testing.T) {
	bad := regexp.MustCompile(`setInterval|setTimeout|fetch\(|EventSource|document\.|innerHTML|\.style\b|window\.`)
	for _, id := range stripPanels {
		b, err := webFS.ReadFile("web/panels/" + id + "/panel.js")
		if err != nil {
			t.Fatal(err)
		}
		if m := bad.Find(b); m != nil {
			t.Errorf("%s: %q breaks the panel contract", id, m)
		}
		mb, _ := webFS.ReadFile("web/panels/" + id + "/panel.json")
		var m PanelManifest
		if err := json.Unmarshal(mb, &m); err != nil {
			t.Fatal(err)
		}
		if m.ID != id || !m.OptIn || !m.Slot || !m.CockpitOnly || m.Endpoint != GlanceRoute {
			t.Errorf("%s manifest: %+v", id, m)
		}
	}
}
