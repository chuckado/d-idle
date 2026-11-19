package watch

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The stack toast (.toast on --surface) once kept the old inverted colour (--bg on a dark surface): unreadable text
// and Dismiss button on the phone. Toast text must follow --text on --surface in both themes, at least 4.5:1.
func TestToastTextContrastBothThemes(t *testing.T) {
	tok, _ := os.ReadFile("web/tokens.css")
	app, _ := os.ReadFile("web/app.css")
	css := string(app)
	if m := regexp.MustCompile(`(?m)^\.toast \{[^}]*\}`).FindAllString(css, -1); len(m) != 1 || !strings.Contains(m[0], "color: var(--text)") {
		t.Fatalf("exactly one .toast rule, and it sets color: var(--text): %q", m)
	}
	dark := tokenBlock(t, string(tok), ":root {")
	light := tokenBlock(t, string(tok), `:root[data-theme="light"] {`)
	for name, vars := range map[string]map[string]string{"dark": dark, "light": merge(dark, light)} {
		fg, bg := resolveHex(t, vars, "--d-text"), resolveHex(t, vars, "--d-surface-1")
		if r := contrast(fg, bg); r < 4.5 {
			t.Errorf("%s theme: toast text %s on %s is %.2f:1", name, fg, bg, r)
		}
	}
}

func tokenBlock(t *testing.T, css, head string) map[string]string {
	i := strings.Index(css, head)
	if i < 0 {
		t.Fatalf("tokens.css lacks %s", head)
	}
	blk := css[i+len(head):]
	blk = blk[:strings.Index(blk, "}")]
	out := map[string]string{}
	for _, m := range regexp.MustCompile(`(--d-[\w-]+):\s*([^;]+);`).FindAllStringSubmatch(blk, -1) {
		out[m[1]] = strings.TrimSpace(m[2])
	}
	return out
}

func merge(base, over map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range over {
		out[k] = v
	}
	return out
}

func resolveHex(t *testing.T, vars map[string]string, name string) string {
	v := vars[name]
	for i := 0; i < 6 && strings.HasPrefix(v, "var("); i++ {
		v = vars[strings.TrimSuffix(strings.TrimPrefix(v, "var("), ")")]
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$|^#[0-9a-fA-F]{3}$`).MatchString(v) {
		t.Fatalf("%s does not resolve to a hex colour: %q", name, v)
	}
	return v
}

func contrast(a, b string) float64 {
	la, lb := lum(a), lum(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + .05) / (lb + .05)
}

func lum(h string) float64 {
	h = strings.TrimPrefix(h, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	var c [3]float64
	for i := range c {
		n, _ := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
		f := float64(n) / 255
		if f <= .03928 {
			c[i] = f / 12.92
		} else {
			c[i] = math.Pow((f+.055)/1.055, 2.4)
		}
	}
	return .2126*c[0] + .7152*c[1] + .0722*c[2]
}
