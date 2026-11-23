package watch

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// E11a (docs/design/fable-review-2026-11-22/e11-deletion-plan.md): the pre-shell cards, their scripts and the old layout presets are gone.
// These tests fail while one of them is still in the tree or still named by shipped code, so a deleted file cannot creep back in.
var e11aDeleted = []string{
	"web/app.js", "web/unlockui.js", "web/alerts.js",
	"web/layouts/classic.json", "web/layouts/simple.json", "web/layouts/standard.json", "web/layouts/debug.json",
	"web/layouts/mockup.json", "web/layouts/live.json", "web/layouts/command.json",
}

func TestE11aDeletedFilesStayGone(t *testing.T) {
	for _, f := range e11aDeleted {
		if _, err := webFS.ReadFile(f); err == nil {
			t.Errorf("%s is still embedded: E11a deletes it", f)
		}
	}
}

// Nothing the viewer ships (HTML, JS, CSS, JSON, non-test Go) may still name a deleted file or route.
func TestE11aNoReferenceToDeleted(t *testing.T) {
	ban := regexp.MustCompile(`/watch/app\.js|web/app\.js|unlockui|UnlockUIRoute|AlertsUIRoute|/watch/alerts\.js|web/alerts\.js|layouts/(classic|simple|standard|debug|mockup|live|command)\b|id="(sheet|scrim)"|class="d-compat"|data-lid=`)
	check := func(path string) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Error(err)
			return
		}
		if m := ban.Find(b); m != nil {
			t.Errorf("%s still names %q (deleted in E11a)", path, m)
		}
	}
	for _, g := range []string{"web/*.html", "web/*.js", "web/*.css", "web/panels/*/*", "*.go"} {
		files, _ := filepath.Glob(g)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			check(f)
		}
	}
}
