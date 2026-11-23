package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// web/settings.js in a stub DOM (node, skipped without it; card uisettings): the Settings sheet's focus in, Escape, labels, reason lines,
// the Rules calls (revoke one, revoke all) and the Alerts settings.
func TestSettingsSheetDOM(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for src, dst := range map[string]string{"web/sheet.js": "sheet.cjs", "web/toast.js": "toast.cjs", "web/why.js": "why.cjs", "web/settings.js": "settings.cjs"} {
		b, err := webFS.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, dst), b, 0o600)
	}
	for _, f := range []string{"nav_dom.cjs", "settings_test.cjs"} {
		b, err := os.ReadFile(filepath.Join("testdata", f))
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, f), b, 0o600)
	}
	out, err := exec.Command(node, filepath.Join(dir, "settings_test.cjs"), dir).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	var res struct{ Fails []string }
	if err := json.Unmarshal(out[strings.Index(string(out), "{"):], &res); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, f := range res.Fails {
		t.Error(f)
	}
}

// The More sheet of the shared navigation holds a Settings entry, and every page that loads nav.js therefore reaches it (TestNavEveryPage
// checks that each page loads nav.js); nav.js opens the sheet in place and the script is served.
func TestNavSettingsEntry(t *testing.T) {
	c := loadNav(t)
	found := false
	for _, m := range c.More {
		if m.ID == "settings" && m.Label == "Settings" && m.Href != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("nav.json More has no Settings entry")
	}
	js, _ := webFS.ReadFile("web/nav.js")
	for _, want := range []string{"'settings'", "settings.js", "sheet.js", "toast.js"} {
		if !strings.Contains(string(js), want) {
			t.Errorf("nav.js does not mention %s", want)
		}
	}
	for _, name := range []string{"settings.js"} {
		if b, _ := webFS.ReadFile("web/" + name); len(b) == 0 {
			t.Errorf("web/%s missing from the embedded web", name)
		}
	}
	if !strings.Contains(string(mustRead(t, "http_view.go")), `"/watch/settings.js"`) {
		t.Error("the viewer does not serve /watch/settings.js")
	}
}

func mustRead(t *testing.T, p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// web/alerts-feed.js in node (skipped without it; card uialerts): off means no request; the sound or granted notifications turn the feed on.
func TestAlertsFeedJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "alertsfeed.cjs"), mustReadFS(t, "web/alerts-feed.js"), 0o600)
	os.WriteFile(filepath.Join(dir, "alertsfeed_test.cjs"), mustRead(t, "testdata/alertsfeed_test.cjs"), 0o600)
	out, err := exec.Command(node, filepath.Join(dir, "alertsfeed_test.cjs"), dir).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	var res struct{ Fails []string }
	if err := json.Unmarshal(out[strings.Index(string(out), "{"):], &res); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, f := range res.Fails {
		t.Error(f)
	}
}

func mustReadFS(t *testing.T, p string) []byte {
	b, err := webFS.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Settings > Rules links to the Loot rules sheet (the old Inventory entry) and the live screen opens it from ?lootrules=1 (card uiloot).
func TestSettingsLinksLootRules(t *testing.T) {
	if !strings.Contains(string(mustReadFS(t, "web/shell.js")), "qs.get('lootrules') === '1'") {
		t.Error("shell.js does not open the Loot rules sheet from ?lootrules=1")
	}
	if !strings.Contains(string(mustReadFS(t, "web/settings.js")), "/watch?live=1&lootrules=1") {
		t.Error("settings.js has no link to the Loot rules sheet")
	}
}

// The keyboard walk (card uikeys): Tab order, wrap, Shift+Tab, native controls, Escape and focus back on the Settings and More sheets;
// and no positive tabindex anywhere in the viewer's pages and scripts (the order is the document order).
func TestKeyboardWalk(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for src, dst := range map[string]string{"web/sheet.js": "sheet.cjs", "web/toast.js": "toast.cjs", "web/settings.js": "settings.cjs", "web/nav.js": "nav.cjs", "web/nav.json": "nav.json"} {
		os.WriteFile(filepath.Join(dir, dst), mustReadFS(t, src), 0o600)
	}
	for _, f := range []string{"nav_dom.cjs", "keywalk_test.cjs"} {
		os.WriteFile(filepath.Join(dir, f), mustRead(t, "testdata/"+f), 0o600)
	}
	out, err := exec.Command(node, filepath.Join(dir, "keywalk_test.cjs"), dir).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	var res struct{ Fails []string }
	if err := json.Unmarshal(out[strings.Index(string(out), "{"):], &res); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, f := range res.Fails {
		t.Error(f)
	}
}

func TestNoPositiveTabindex(t *testing.T) {
	bad := regexp.MustCompile(`tabindex="[1-9]|tabIndex\s*=\s*[1-9]|tabindex', '[1-9]`)
	for _, dirGlob := range []string{"web/*.js", "web/panels/*/*.js", "web/*.html", "*.go"} {
		files, _ := filepath.Glob(dirGlob)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			b, _ := os.ReadFile(f)
			if m := bad.Find(b); m != nil {
				t.Errorf("%s sets a positive tabindex (%s): the tab order is the document order", f, m)
			}
		}
	}
}
