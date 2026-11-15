package watch

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The character tooltip is black in both themes, laid over the page (inventory-tip.css: the tip colour over --d-surface-0).
// Its text colours (tip tokens) must read on that panel at 4.5:1 or better in the dark and in the light theme; the
// requirement rows and the comparison deltas once used the page's --d-good, --d-bad and --d-warn, which are dark colours in
// the light theme (3.5, 2.9 and 3.5:1 on the panel; review c2c3 R9).
func TestTooltipContrastBothThemes(t *testing.T) {
	tok, err := os.ReadFile("web/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	dark := tokenBlock(t, string(tok), ":root {")
	light := merge(dark, tokenBlock(t, string(tok), `:root[data-theme="light"] {`))
	for name, vars := range map[string]map[string]string{"dark": dark, "light": light} {
		bg := tipPanel(t, vars)
		for _, k := range []string{"--d-tip-text", "--d-tip-dim", "--d-tip-bad", "--d-tip-good", "--d-tip-warn", "--d-tip-q-normal", "--d-tip-q-magic", "--d-tip-q-rare", "--d-tip-q-set", "--d-tip-q-unique", "--d-tip-q-crafted"} {
			if r := contrast(resolveHex(t, vars, k), bg); r < 4.5 {
				t.Errorf("%s theme: %s %s on the tooltip panel %s is %.2f:1 (need 4.5)", name, k, resolveHex(t, vars, k), bg, r)
			}
		}
	}
}

// tipPanel is the colour of the tooltip panel: --d-tip-bg (rgb(r g b / a)) over the lightest surface it can lie on.
func tipPanel(t *testing.T, vars map[string]string) string {
	m := regexp.MustCompile(`^rgb\((\d+) (\d+) (\d+) / ([\d.]+)\)$`).FindStringSubmatch(vars["--d-tip-bg"])
	if m == nil {
		t.Fatalf("--d-tip-bg is not rgb(r g b / a): %q", vars["--d-tip-bg"])
	}
	a, _ := strconv.ParseFloat(m[4], 64)
	under := strings.TrimPrefix(resolveHex(t, vars, "--d-surface-0"), "#")
	if len(under) == 3 {
		under = string([]byte{under[0], under[0], under[1], under[1], under[2], under[2]})
	}
	out := "#"
	for i := 0; i < 3; i++ {
		fg, _ := strconv.Atoi(m[i+1])
		bg, _ := strconv.ParseUint(under[2*i:2*i+2], 16, 8)
		out += fmt.Sprintf("%02x", int(float64(fg)*a+float64(bg)*(1-a)+.5))
	}
	return out
}

// No rule of the tooltip takes the page's state colours as a text colour: they follow the theme, the panel does not.
func TestTooltipUsesTipTokensForText(t *testing.T) {
	css, err := os.ReadFile("web/inventory-tip.css")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^\s*\.inv-tip[^{]*\{[^}]*\bcolor:\s*var\(--d-(good|bad|warn|info)\)`)
	if m := re.FindString(string(css)); m != "" {
		t.Errorf("a tooltip rule colours text with a page state token (use --d-tip-*): %s", strings.TrimSpace(m))
	}
}
