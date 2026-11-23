package watch

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The allocator wording of cards allochandoff and the Points page (plan line, held, spent / unspent / waiting) wraps on a phone: the
// elements that carry it are plain hint paragraphs, and no rule makes a hint one line or hides its end (card uiloot; picture
// docs/ui/settings/build-plan-handoff-390.png, fit page builds-list at 390x844 and 1366x768).
func TestPlanWordingWrapsOnAPhone(t *testing.T) {
	for _, id := range []string{"bplan", "waiting"} {
		if !regexp.MustCompile(`<p class="d-hint" id="` + id + `"`).MatchString(pointsPage) {
			t.Errorf("#%s is not a plain d-hint paragraph", id)
		}
	}
	css, _ := os.ReadFile("web/components.css")
	for _, rule := range regexp.MustCompile(`\.d-hint[^{]*\{[^}]*\}`).FindAllString(string(css), -1) {
		for _, bad := range []string{"white-space: nowrap", "text-overflow", "overflow: hidden"} {
			if strings.Contains(rule, bad) {
				t.Errorf(".d-hint carries %q: the plan wording would be cut on a phone: %s", bad, rule)
			}
		}
	}
	js, _ := os.ReadFile("web/points.js")
	for _, want := range []string{"Spent by the build so far this run: ", "Unspent now: ", "' Waiting: '"} {
		if !strings.Contains(string(js), want) {
			t.Errorf("points.js lost the wording %q", want)
		}
	}
}
