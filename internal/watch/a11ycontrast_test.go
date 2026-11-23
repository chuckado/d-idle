package watch

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// WCAG 2.1 AA for the shared tokens in both themes (card uia11y, docs/ui-design-system.md "Accessibility"): text tokens on the page and
// card surfaces at 4.5:1; the "on" colours on their fills at 4.5:1; focus ring, control borders and the accent at 3:1 (non-text UI).
// The numbers come from web/tokens.css only: a token that drifts under the line fails here, per theme and pair.
func TestTokenContrastAA(t *testing.T) {
	tok, err := os.ReadFile("web/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	dark := tokenBlock(t, string(tok), ":root {")
	light := merge(dark, tokenBlock(t, string(tok), `:root[data-theme="light"] {`))
	surfaces := []string{"--d-bg", "--d-surface-0", "--d-surface-1", "--d-surface-2", "--d-surface-3"}
	text := []string{"--d-text", "--d-text-2", "--d-accent-text", "--d-good", "--d-warn", "--d-bad", "--d-info"}
	pairs := []struct {
		fg, bg string
		min    float64
	}{{"--d-on-accent", "--d-accent", 4.5}, {"--d-on-good", "--d-good", 4.5}, {"--d-on-warn", "--d-warn", 4.5}, {"--d-on-bad", "--d-bad", 4.5}, {"--d-on-info", "--d-info", 4.5}}
	for _, s := range surfaces[:3] {
		pairs = append(pairs, struct {
			fg, bg string
			min    float64
		}{"--d-focus", s, 3}, struct {
			fg, bg string
			min    float64
		}{"--d-line-control", s, 3})
	}
	for _, f := range text {
		for _, s := range surfaces {
			pairs = append(pairs, struct {
				fg, bg string
				min    float64
			}{f, s, 4.5})
		}
	}
	for name, vars := range map[string]map[string]string{"dark": dark, "light": light} {
		for _, p := range pairs {
			if _, ok := vars[p.fg]; !ok {
				continue
			}
			if r := contrast(resolveHex(t, vars, p.fg), resolveHex(t, vars, p.bg)); r < p.min {
				t.Errorf("%s theme: %s on %s is %.2f:1 (need %.1f)", name, p.fg, p.bg, r, p.min)
			}
		}
	}
	// the checker itself: two close greys fail, black on white passes
	if contrast("#777777", "#808080") >= 4.5 || contrast("#000000", "#ffffff") < 20 {
		t.Error("the contrast function is wrong")
	}
}

// Every stylesheet that moves something with a literal duration (an animation or a transition that does not take its time from the
// --d-dur-* tokens, which are 0 ms under prefers-reduced-motion) says what happens when the reader asks for no motion (card uia11y).
func TestMotionHonoursReducedMotion(t *testing.T) {
	files, _ := filepath.Glob("web/*.css")
	more, _ := filepath.Glob("web/panels/*/*.css")
	tok, _ := os.ReadFile("web/tokens.css")
	globalReduce := strings.Contains(string(tok), "prefers-reduced-motion: reduce") && strings.Contains(string(tok), "animation-duration: .01ms; animation-iteration-count: 1;")
	moving := regexp.MustCompile(`(?m)^[^/\n]*\b(animation|transition):[^;}]*`)
	for _, f := range append(files, more...) {
		b, _ := os.ReadFile(f)
		s := string(b)
		for _, m := range moving.FindAllString(s, -1) {
			if strings.Contains(m, "var(--d-dur") || strings.Contains(m, "transition: none") || strings.Contains(m, "animation: none") {
				continue
			}
			if !strings.Contains(s, "prefers-reduced-motion") && !globalReduce {
				t.Errorf("%s moves with a literal duration and has no prefers-reduced-motion rule: %s", f, strings.TrimSpace(m))
			}
		}
	}
}
