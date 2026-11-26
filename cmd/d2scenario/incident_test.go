package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCmdIncident(t *testing.T) {
	dir := t.TempDir()
	in := map[string]any{"version": "didle.incident/1", "kind": "death", "time": time.Date(2026, 11, 25, 3, 19, 0, 0, time.UTC),
		"areaName": "Stony Field", "class": "Paladin", "tags": []string{"SETUP"},
		"pre": map[string]any{"start": "field", "area": 4, "level": 8}, "facts": map[string]any{}}
	b, _ := json.Marshal(in)
	src := filepath.Join(dir, "i.json")
	os.WriteFile(src, b, 0o644)
	var out bytes.Buffer
	if err := cmdIncident([]string{"-class", "Sorceress", src}, &out); err != nil || !strings.Contains(out.String(), "incident.stony-field-") || !strings.Contains(out.String(), `"Sorceress"`) {
		t.Fatalf("%v %s", err, out.String())
	}
	out.Reset()
	if err := cmdIncident([]string{"-out", filepath.Join(dir, "rows"), src}, &out); err != nil {
		t.Fatal(err)
	}
	names, _ := filepath.Glob(filepath.Join(dir, "rows", "*.json"))
	if len(names) != 1 {
		t.Fatalf("%v", names)
	}
	if err := cmdIncident(nil, &out); err == nil {
		t.Fatal("no file accepted")
	}
}
