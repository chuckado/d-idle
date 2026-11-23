package watch

import (
	"regexp"
	"strings"
	"testing"
)

var (
	reOffControl = regexp.MustCompile(`(?is)<(button|input|select)\b[^>]*(\sdisabled\b|aria-disabled="true")[^>]*>`)
	reDescribed  = regexp.MustCompile(`aria-describedby="([^"]+)"`)
)

// Card uireason (review C13): a control that is off in the static markup of a page names a reason line with aria-describedby,
// the line is on the page, and the title attribute is not the reason. The Vault, the quest panels and the live inventory bar build
// theirs in script; their node tests run web/why.js over the rendered DOM (vault_*.mjs, questui/quest_ui.mjs).
func TestStaticDisabledControlsHaveReasonLines(t *testing.T) {
	pages := map[string]string{"charPage": charPage, "skillsPage": skillsPage, "pointsPage": pointsPage}
	for _, f := range []string{"watch.html", "vault.html", "characters.html"} {
		b, err := webFS.ReadFile("web/" + f)
		if err != nil {
			t.Fatal(err)
		}
		pages[f] = string(b)
	}
	n := 0
	for name, src := range pages {
		for _, m := range reOffControl.FindAllString(src, -1) {
			n++
			d := reDescribed.FindStringSubmatch(m)
			if d == nil {
				t.Errorf("%s: %s is off with no aria-describedby reason line", name, m)
				continue
			}
			if !strings.Contains(src, `id="`+d[1]+`"`) {
				t.Errorf("%s: the reason line %q of %s is not on the page", name, d[1], m)
			}
			if strings.Contains(m, " title=") {
				t.Errorf("%s: %s carries a title (a title is not a reason)", name, m)
			}
		}
	}
	if n == 0 {
		t.Fatal("no off control found in any page: the test looks at nothing")
	}
}
