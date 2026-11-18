package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"d2lod-bot/internal/charsel"
	"d2lod-bot/internal/paths"
)

func TestSetArg(t *testing.T) {
	got := setArg([]string{"--managed", "--char", "Old", "--d2d-class=3", "--no-save"}, "--d2d-class", "-1")
	got = setArg(got, "--char", "New")
	if want := []string{"--managed", "--no-save", "--d2d-class", "-1", "--char", "New"}; !reflect.DeepEqual(got, want) {
		t.Errorf("%v", got)
	}
}

func switchRig(t *testing.T) (*switcher, string) {
	root := t.TempDir()
	o := options{instance: "i7", layout: paths.Layout{Root: root}, goalSpec: "campaign:normal"}
	ip := o.layout.MustInstance("i7")
	if err := os.MkdirAll(ip.State(), 0o755); err != nil {
		t.Fatal(err)
	}
	return &switcher{o: o, log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		cur: []string{"--managed", "--backend", "d2d", "--char", "Holy", "--d2d-class", "-1", "--goals", "campaign:normal", "--supervise"}}, ip.Dir()
}

func TestSwitcherLoadsAnotherCharacterAndParksState(t *testing.T) {
	w, dir := switchRig(t)
	os.WriteFile(filepath.Join(dir, "goals.json"), []byte("holy"), 0o600)
	os.MkdirAll(filepath.Join(dir, "characters", "Zapp"), 0o755)
	os.WriteFile(filepath.Join(dir, "characters", "Zapp", "goals.json"), []byte("zapp"), 0o600)
	writeIntentFor(t, dir, charsel.Intent{From: "Holy", To: "Zapp", ClassIndex: -1})
	if !w.apply() {
		t.Fatal("apply refused")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "goals.json")); string(b) != "zapp" {
		t.Errorf("goals %q", b)
	}
	if _, ok := charsel.ReadIntent(dir); ok {
		t.Error("intent not cleared")
	}
	a := w.args(1, 0)
	want := []string{"--managed", "--backend", "d2d", "--char", "Zapp", "--d2d-class", "-1", "--autostart", "--clean-start"}
	if !reflect.DeepEqual(a, want) {
		t.Errorf("args %v", a) // a character with parked goals does not get --goals again (it would queue a second campaign)
	}
}

func TestSwitcherCreateStartsFreshWithGoalsThenLoads(t *testing.T) {
	w, dir := switchRig(t)
	writeIntentFor(t, dir, charsel.Intent{From: "Holy", To: "Newb", Create: true, ClassIndex: 4})
	if !w.apply() {
		t.Fatal("apply refused")
	}
	a := w.args(1, 0)
	if !reflect.DeepEqual(a[len(a)-2:], []string{"--goals", "campaign:normal"}) || !slices.Contains(a, "4") {
		t.Errorf("create args %v", a)
	}
	// the save appears: the next start loads it instead of creating over it
	save := filepath.Join(dir, "d2d", "share", "saves")
	os.MkdirAll(save, 0o755)
	os.WriteFile(filepath.Join(save, "Newb.d2s"), []byte("x"), 0o600)
	a = w.args(2, 0)
	if !slices.Contains(a, "-1") || slices.Contains(a, "4") {
		t.Errorf("restart after the first save still creates: %v", a)
	}
}

func writeIntentFor(t *testing.T, dir string, in charsel.Intent) {
	t.Helper()
	in.Schema, in.At = charsel.IntentSchema, "2026-11-15T12:00:00Z"
	b := `{"schema":"` + in.Schema + `","from":"` + in.From + `","to":"` + in.To + `","classIndex":` + itoa(in.ClassIndex) + `,"at":"` + in.At + `"`
	if in.Create {
		b += `,"create":true`
	}
	if err := os.WriteFile(filepath.Join(dir, "charswitch.json"), []byte(b+"}"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

func TestSuperviseSwitchIsNotACrash(t *testing.T) {
	s, att, now := testSup([]int{exitSwitch, exitSwitch, exitSwitch, exitSwitch, 0}, 1) // max 1 restart: a crash loop would stop
	s.onSwitch = func() bool { return true }
	t0 := *now
	if c := s.loop(context.Background()); c != 0 || len(*att) != 5 {
		t.Fatalf("code %d attempts %v", c, *att)
	}
	if d := now.Sub(t0); d > 10*time.Second {
		t.Errorf("a switch slept %v", d)
	}
	// without an intent the exit is an ordinary non-zero code
	s2, att2, _ := testSup([]int{exitSwitch}, 2)
	s2.onSwitch = func() bool { return false }
	if c := s2.loop(context.Background()); c != exitSwitch || len(*att2) != 3 {
		t.Errorf("code %d attempts %v", c, *att2)
	}
}

func TestSaveFolderFindsTheActiveCharactersFolder(t *testing.T) {
	root := t.TempDir()
	a, b := filepath.Join(root, "a", "saves"), filepath.Join(root, "b", "saves")
	for _, d := range []string{a, b} {
		os.MkdirAll(d, 0o755)
	}
	os.WriteFile(filepath.Join(a, "One.d2s"), nil, 0o600)
	os.WriteFile(filepath.Join(a, "Two.d2s"), nil, 0o600)
	os.WriteFile(filepath.Join(b, "Holy.d2s"), nil, 0o600)
	if got := saveFolder([]string{root}, func() string { return "holy" })(); got != b {
		t.Errorf("active's folder: %s", got)
	}
	if got := saveFolder([]string{root}, func() string { return "" })(); got != a {
		t.Errorf("the folder with most saves: %s", got)
	}
	if got := saveFolder([]string{filepath.Join(root, "none")}, func() string { return "" })(); got != "" {
		t.Errorf("no folder: %q", got)
	}
}

func TestCharactersDeskGates(t *testing.T) {
	t.Setenv("D2BOT_RUN", "r1")
	ip := paths.Layout{Root: t.TempDir()}.MustInstance("i7")
	d := charactersDesk(ip, func() string { return "Holy" }, nil, nil, nil)
	if _, no := d.Switch("Zapp"); no == nil || no.Code != "measured_run" {
		t.Errorf("%v", no)
	}
}

func TestSwitcherRecordsActiveCharacterAndRollsBack(t *testing.T) {
	w, dir := switchRig(t)
	os.MkdirAll(filepath.Join(dir, "characters", "Zapp"), 0o755)
	writeIntentFor(t, dir, charsel.Intent{From: "Holy", To: "Zapp", ClassIndex: -1})
	if !w.apply() || charsel.ReadActive(dir) != "Zapp" {
		t.Fatalf("active %q", charsel.ReadActive(dir))
	}
	// a bad target (path-like) is refused: nothing changes, the record keeps the old hero
	writeIntentFor(t, dir, charsel.Intent{From: "Zapp", To: "../evil", ClassIndex: -1})
	if w.apply() {
		t.Fatal("bad target accepted")
	}
	if got := charsel.ReadActive(dir); got != "Zapp" {
		t.Errorf("active after a refused switch %q", got)
	}
	if _, ok := charsel.ReadIntent(dir); ok {
		t.Error("intent of a refused switch not cleared")
	}
	// no intent at all: exit 6 without a request restarts the old character
	if w.apply() {
		t.Error("apply without intent")
	}
}

func TestActiveFileRejectsBadNames(t *testing.T) {
	d := t.TempDir()
	if charsel.WriteActive(d, "a/b") == nil || charsel.ReadActive(d) != "" {
		t.Error("bad name recorded")
	}
	os.WriteFile(filepath.Join(d, charsel.ActiveFile), []byte("../x\n"), 0o600)
	if charsel.ReadActive(d) != "" {
		t.Error("bad file content accepted")
	}
}
