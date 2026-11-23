package watch

import (
	"regexp"
	"strings"
	"testing"
)

// The Return, Policy and Hall pages are on the shared frame and its components (card uiframe, review E12): no app.css, no old `.btn`,
// `.card`, `.hint`, `.badge`, `.empty` classes and no old colour aliases, so app.css can go with the pre-shell page. Each rule below
// fails on the page as it was before this card.
func TestReturnPolicyHallOnSharedFrame(t *testing.T) {
	pages := map[string]string{
		"return": string(RenderReturn(ReturnData{}, false)),
		"policy": string(RenderPolicy(PolicyView{})),
		"hall":   string(RenderHall(&HallView{})),
	}
	oldClass := regexp.MustCompile(`class="(?:[a-z-]+ )*(?:btn|card|hint|badge|empty)(?: [a-z-]+)*"|'(?:btn|card|hint|badge|empty)(?: [a-z]+)*'`)
	oldVar := regexp.MustCompile(`var\(--(?:bg|surface2?|line|text|muted|accent|good|warn|bad|info)\)`)
	for name, src := range pages {
		for _, want := range []string{"/watch/tokens.css", "/watch/components.css", "/watch/nav.css", "/watch/nav.js"} {
			if !strings.Contains(src, want) {
				t.Errorf("%s: does not load %s", name, want)
			}
		}
		if strings.Contains(src, "/watch/app.css") {
			t.Errorf("%s: still loads app.css", name)
		}
		if m := oldClass.FindString(src); m != "" {
			t.Errorf("%s: old class vocabulary %s (use d-btn, d-card, d-hint, d-badge, d-empty)", name, m)
		}
		if m := oldVar.FindString(src); m != "" {
			t.Errorf("%s: old colour alias %s (use the --d-* tokens)", name, m)
		}
	}
	if m := oldVar.FindString(returnCSS); m != "" {
		t.Errorf("return.css: old colour alias %s", m)
	}
	// the buttons the pages build carry the shared vocabulary
	for _, b := range regexp.MustCompile(`<button[^>]*>`).FindAllString(pages["return"]+pages["policy"], -1) {
		if !strings.Contains(b, `class="d-btn"`) {
			t.Errorf("a button outside the d-btn family: %s", b)
		}
	}
}
