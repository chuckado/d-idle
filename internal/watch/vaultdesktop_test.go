package watch

import (
	"regexp"
	"strings"
	"testing"
)

// User decision 2026-11-21: the Vault page does not show missed drops (the API keeps the counter), and it is one scrolling document.
func TestVaultPageHasNoMissedDropsBlockAndNoFrame(t *testing.T) {
	for _, f := range []string{"web/vault.js", "web/vault-logic.js", "web/vault.json", "web/vault.css"} {
		b, err := webFS.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		for _, bad := range []string{"missed-clear", "missRow", "missedTitle", "missedClear", "fullBannerMissed"} {
			if strings.Contains(s, bad) {
				t.Errorf("%s still renders missed drops (%q)", f, bad)
			}
		}
	}
	css, _ := webFS.ReadFile("web/vault.css")
	if regexp.MustCompile(`(^|[^-])block-size: 100dvh`).MatchString(strings.SplitN(string(css), "@layer utilities", 2)[0]) {
		t.Error("vault.css has a 100dvh frame outside the phone layer")
	}
}
