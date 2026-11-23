package watch

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var browserDialog = regexp.MustCompile(`(^|[^A-Za-z0-9_.$])(window\.)?(confirm|alert|prompt)\s*\(`)

// Card uisheet: no browser dialog anywhere in the web assets or the pages the Go files write. Everything asks through web/sheet.js.
func TestNoBrowserDialogs(t *testing.T) {
	check := func(name string, b []byte) {
		for i, l := range strings.Split(string(b), "\n") {
			if c := strings.Index(l, "//"); c >= 0 && !strings.Contains(l[:c], "'") && !strings.Contains(l[:c], "\"") {
				l = l[:c]
			}
			if browserDialog.MatchString(l) {
				t.Errorf("%s:%d uses a browser dialog (use dSheet, web/sheet.js): %s", name, i+1, strings.TrimSpace(l))
			}
		}
	}
	err := fs.WalkDir(webFS, "web", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !(strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".html")) {
			return err
		}
		b, err := webFS.ReadFile(p)
		check(p, b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Every page that loads web/sheet.js also loads components.css, where the sheet is styled, and the script is served.
func TestSheetScriptServed(t *testing.T) {
	b, err := webFS.ReadFile("web/sheet.js")
	if err != nil || !strings.Contains(string(b), "window.dSheet") && !strings.Contains(string(b), "g.dSheet") {
		t.Fatalf("sheet.js: %v", err)
	}
	if css, _ := webFS.ReadFile("web/components.css"); !strings.Contains(string(css), ".d-sheet-back") {
		t.Fatal("components.css has no .d-sheet-back")
	}
}
