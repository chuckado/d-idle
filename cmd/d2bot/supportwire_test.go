package main

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/paths"
	"d2lod-bot/internal/support"
)

// The desk gathers a seeded instance (secrets in the log, a key file and a token file next to it) and the whole chain
// leaves none of them in the zip.
func TestSupportDeskCollectsClean(t *testing.T) {
	root := t.TempDir()
	lay := paths.Layout{Root: root}
	logDir := filepath.Join(root, "instances", "i1", "logs")
	os.MkdirAll(logDir, 0o755)
	os.WriteFile(filepath.Join(logDir, "d2bot.log"), []byte("2099-01-01T00:00:00Z INFO start data="+root+" peer 10.1.2.3\n2099-01-01T00:00:01Z ERROR api token=SeedSecret99 failed for kid@example.org\n"), 0o600)
	os.MkdirAll(filepath.Join(root, "tokens"), 0o700)
	os.WriteFile(filepath.Join(root, "tokens", "watch"), []byte("SeedWatchToken"), 0o600)
	os.WriteFile(filepath.Join(root, "default.key"), []byte("SEEDCDKEY"), 0o600)
	state := filepath.Join(root, "state")
	os.MkdirAll(state, 0o755)
	os.WriteFile(filepath.Join(state, "deaths.jsonl"), []byte("{\"area\":\"x\"}\nnot json\n{\"area\":\"y\"}\n"), 0o600)
	d := &supportDesk{l: ladderOptions{instance: "i1", supportLayout: lay, supportKnobs: func() any { return map[string]float64{"combat.kite": 3} }}, stateDir: state, client: http.DefaultClient}
	local := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/glance" {
			io.WriteString(w, `{"hero":"x","watch_token":"SeedWatchToken"}`)
			return
		}
		http.NotFound(w, r)
	})
	in, err := d.Collect(support.Options{}, local)
	if err != nil {
		t.Fatal(err)
	}
	b, err := support.Build(in, support.Options{})
	if err != nil || !b.Sendable() {
		t.Fatalf("%v %+v", err, b)
	}
	zr, _ := zip.NewReader(bytes.NewReader(b.Zip), int64(len(b.Zip)))
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
		r, _ := f.Open()
		data, _ := io.ReadAll(r)
		for _, bad := range []string{"SeedSecret99", "SeedWatchToken", "SEEDCDKEY", root, "10.1.2.3", "kid@example.org"} {
			if strings.Contains(string(data), bad) {
				t.Errorf("%s contains %q", f.Name, bad)
			}
		}
	}
	for _, want := range []string{"log.txt", "errors.txt", "glance.json", "knobs.json", "version.json", "system.json", "deaths.json", "MANIFEST.txt"} {
		if !names[want] {
			t.Errorf("missing %s in %v", want, names)
		}
	}
	for n := range names {
		if strings.Contains(n, "key") || strings.Contains(n, "token") || strings.Contains(n, "screenshot") || strings.HasPrefix(n, "save/") {
			t.Errorf("unexpected file %s", n)
		}
	}
}
