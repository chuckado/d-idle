package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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
