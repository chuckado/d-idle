package watch

import (
	"strings"
	"testing"
)

// The phone is a page: no panel list on the phone tabs may stay an inner scroll box (the 2026-11-21 report, Quests tab).
func TestPhoneTabsAreNotScrollTraps(t *testing.T) {
	css, _ := webFS.ReadFile("web/utilities.css")
	for _, want := range []string{`.d-main[data-cmode="phone"] .d-slottabs :is(.qb-box, .qb-list`, `overscroll-behavior: auto`, `data-cmode="phone"] { block-size: auto`} {
		if !strings.Contains(string(css), want) {
			t.Errorf("utilities.css: missing %q", want)
		}
	}
}
