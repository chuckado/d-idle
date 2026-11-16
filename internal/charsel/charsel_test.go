package charsel

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeD2S is a header-only image: enough for the list (the item section is never read), not for saves.Verify.
func fakeD2S(name string, d2sClass byte, level byte, status byte) []byte {
	b := make([]byte, 0x40)
	binary.LittleEndian.PutUint32(b[offSignature:], signature)
	binary.LittleEndian.PutUint32(b[offVersion:], version96)
	copy(b[offName:], name)
	b[offStatus], b[offClass], b[offLevel] = status, d2sClass, level
	binary.LittleEndian.PutUint32(b[offLastPlayed:], 1763200000)
	return b
}

type rig struct {
	t     *testing.T
	root  string
	st    *Store
	m     *Manager
	left  atomic.Int32
	exits chan int
	act   string
}

func newRig(t *testing.T) *rig {
	root := t.TempDir()
	r := &rig{t: t, root: root, exits: make(chan int, 4), act: "Holy"}
	r.st = &Store{SaveDir: filepath.Join(root, "save"), ArchiveDir: filepath.Join(root, "archive"), StateRoot: filepath.Join(root, "inst", "characters"),
		Now: func() time.Time { return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC) }}
	if err := os.MkdirAll(filepath.Join(root, "inst", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	r.put("Holy", 3, 24, flagExpansion) // Paladin
	r.put("Zapp", 1, 9, flagExpansion)  // Sorceress
	r.m = &Manager{Store: r.st, Inst: filepath.Join(root, "inst"), Active: func() string { return r.act }, Supervised: true,
		Leave: func(context.Context) error { r.left.Add(1); return nil }, Exit: func(c int) { r.exits <- c }, ExitCode: 6,
		Verify: func(string) error { return nil }, Now: r.st.Now}
	return r
}

func (r *rig) put(name string, class, level, status byte) {
	if err := os.MkdirAll(r.st.SaveDir, 0o755); err != nil {
		r.t.Fatal(err)
	}
	for ext, body := range map[string][]byte{"d2s": fakeD2S(name, class, level, status), "key": []byte("k-" + name), "ma0": []byte("m-" + name)} {
		if err := os.WriteFile(filepath.Join(r.st.SaveDir, name+"."+ext), body, 0o600); err != nil {
			r.t.Fatal(err)
		}
	}
}

func code(r *Refusal) string {
	if r == nil {
		return ""
	}
	return r.Code
}

func TestListReadsHeadersOnTwoClasses(t *testing.T) {
	r := newRig(t)
	v, ref := r.m.List()
	if ref != nil {
		t.Fatal(ref)
	}
	l := v.(ListView)
	if l.Schema != ListSchema || l.Active != "Holy" || len(l.Characters) != 2 {
		t.Fatalf("%+v", l)
	}
	h, z := l.Characters[0], l.Characters[1]
	if h.Name != "Holy" || h.Class != "paladin" || h.Level != 24 || !h.Active || !h.Expansion || h.Hardcore || h.LastPlayed == "" {
		t.Errorf("paladin row %+v", h)
	}
	if z.Name != "Zapp" || z.Class != "sorceress" || z.Level != 9 || z.Active {
		t.Errorf("sorceress row %+v", z)
	}
	if !l.Caps.Create || !l.Caps.Switch {
		t.Errorf("caps %+v", l.Caps)
	}
}

func TestDamagedAndUnknownClassAreListedNotHidden(t *testing.T) {
	r := newRig(t)
	os.WriteFile(filepath.Join(r.st.SaveDir, "Junk.d2s"), []byte("not a save"), 0o600)
	os.WriteFile(filepath.Join(r.st.SaveDir, "Odd.d2s"), fakeD2S("Odd", 42, 3, 0), 0o600)
	es, _, _ := r.st.List("")
	var junk, odd Entry
	for _, e := range es {
		switch e.Name {
		case "Junk":
			junk = e
		case "Odd":
			odd = e
		}
	}
	if junk.Damaged == "" || odd.Damaged == "" || odd.Class != "" {
		t.Errorf("junk %+v odd %+v (an unknown class byte must be named, never a Sorceress)", junk, odd)
	}
	if _, ref := r.m.Switch("Junk"); code(ref) != "damaged" {
		t.Errorf("switch to a damaged save: %v", ref)
	}
}

func TestNameRules(t *testing.T) {
	for n, ok := range map[string]bool{"Ab": true, "Holy-Man": true, "Holy_Man": true, "A": false, "": false, "Sixteenletters16": false, "Fifteenletters1": false,
		"Fifteenletters": true, "-Hi": false, "Hi-": false, "A-b-c": false, "A-b_c": false, "Hol y": false, "Holy1": false, "Hölzer": false, "../x": false} {
		if err := ValidName(n); (err == nil) != ok {
			t.Errorf("%q: %v want ok=%v", n, err, ok)
		}
	}
}

func TestArchiveDeleteRestoreRoundTrip(t *testing.T) {
	r := newRig(t)
	// parked state of Zapp travels with the archive
	os.MkdirAll(filepath.Join(r.st.StateRoot, "Zapp", "state"), 0o755)
	os.WriteFile(filepath.Join(r.st.StateRoot, "Zapp", "state", "tuning.json"), []byte(`{"z":1}`), 0o600)
	before, _ := os.ReadFile(filepath.Join(r.st.SaveDir, "Zapp.d2s"))

	if _, ref := r.m.Delete("Zapp", "zapp"); code(ref) != "confirm_mismatch" {
		t.Errorf("confirm must be the exact name: %v", ref)
	}
	if _, ref := r.m.Delete("Holy", "Holy"); code(ref) != "active" {
		t.Errorf("active delete: %v", ref)
	}
	if _, err := os.Stat(filepath.Join(r.st.SaveDir, "Holy.d2s")); err != nil {
		t.Fatal("the active character must still be there")
	}
	v, ref := r.m.Delete("Zapp", "Zapp")
	if ref != nil {
		t.Fatal(ref)
	}
	res := v.(Result)
	if res.Archive != "Zapp@20261115T120000Z" {
		t.Errorf("archive id %q", res.Archive)
	}
	for _, f := range []string{"Zapp.d2s", "Zapp.key", "Zapp.ma0"} {
		if _, err := os.Stat(filepath.Join(r.st.SaveDir, f)); err == nil {
			t.Errorf("%s still in the save folder", f)
		}
		if _, err := os.Stat(filepath.Join(r.st.ArchiveDir, "Zapp", "20261115T120000Z", f)); err != nil {
			t.Errorf("%s not archived", f)
		}
	}
	if _, err := os.Stat(filepath.Join(r.st.StateRoot, "Zapp")); err == nil {
		t.Error("parked state of the deleted character left behind")
	}
	lv, _ := r.m.List()
	l := lv.(ListView)
	if len(l.Characters) != 1 || len(l.Deleted) != 1 || l.Deleted[0].ID != res.Archive || l.Deleted[0].Class != "sorceress" || l.Deleted[0].Level != 9 {
		t.Errorf("list after delete: %+v", l)
	}
	if _, ref := r.m.Restore("../Zapp@x"); code(ref) != "bad_archive" {
		t.Errorf("path-like archive id: %v", ref)
	}
	if _, ref := r.m.Restore("Zapp@20000101T000000Z"); code(ref) != "not_found" {
		t.Errorf("unknown archive: %v", ref)
	}
	if _, ref := r.m.Restore(res.Archive); ref != nil {
		t.Fatal(ref)
	}
	after, _ := os.ReadFile(filepath.Join(r.st.SaveDir, "Zapp.d2s"))
	if string(after) != string(before) {
		t.Error("restored save differs")
	}
	if b, _ := os.ReadFile(filepath.Join(r.st.StateRoot, "Zapp", "state", "tuning.json")); string(b) != `{"z":1}` {
		t.Errorf("state not restored: %q", b)
	}
	if _, ref := r.m.Restore(res.Archive); code(ref) != "exists" {
		t.Errorf("restore over an existing character: %v", ref)
	}
	lv, _ = r.m.List()
	if l := lv.(ListView); len(l.Characters) != 2 || len(l.Deleted) != 0 || len(l.Characters[1].Archived) != 1 {
		t.Errorf("after restore the archive hangs off the character: %+v", l)
	}
}

func TestSwitchLeavesAndWritesIntent(t *testing.T) {
	r := newRig(t)
	if _, ref := r.m.Switch("Holy"); code(ref) != "already_active" {
		t.Errorf("%v", ref)
	}
	if _, ref := r.m.Switch("Nobody"); code(ref) != "not_found" {
		t.Errorf("%v", ref)
	}
	v, ref := r.m.Switch("zapp")
	if ref != nil {
		t.Fatal(ref)
	}
	if v.(Result).Switch.To != "Zapp" || r.left.Load() != 1 {
		t.Errorf("%+v left=%d", v, r.left.Load())
	}
	select {
	case c := <-r.exits:
		if c != 6 {
			t.Errorf("exit code %d", c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no exit")
	}
	in, ok := ReadIntent(r.m.Inst)
	if !ok || in.From != "Holy" || in.To != "Zapp" || in.Create || in.ClassIndex != -1 {
		t.Errorf("intent %+v", in)
	}
	if _, ref := r.m.Switch("Zapp"); code(ref) != "switch_in_progress" {
		t.Errorf("second press: %v", ref)
	}
	if _, ref := r.m.Delete("Zapp", "Zapp"); code(ref) != "switch_in_progress" {
		t.Errorf("delete during a switch: %v", ref)
	}
}

func TestSwitchRefusals(t *testing.T) {
	r := newRig(t)
	r.m.Measured = true
	for name, f := range map[string]func() *Refusal{
		"switch":  func() *Refusal { _, x := r.m.Switch("Zapp"); return x },
		"create":  func() *Refusal { _, x := r.m.Create("Newb", "paladin", false); return x },
		"delete":  func() *Refusal { _, x := r.m.Delete("Zapp", "Zapp"); return x },
		"restore": func() *Refusal { _, x := r.m.Restore("Zapp@20261115T120000Z"); return x },
	} {
		if c := code(f()); c != "measured_run" {
			t.Errorf("%s during a measured run: %q", name, c)
		}
	}
	if r.left.Load() != 0 {
		t.Error("a refused verb left the game")
	}
	r.m.Measured, r.m.Supervised = false, false
	if _, ref := r.m.Switch("Zapp"); code(ref) != "no_supervisor" {
		t.Errorf("%v", ref)
	}
	r.m.Supervised = true
	r.m.Leave = func(context.Context) error { return errors.New("d2d did not answer") }
	_, ref := r.m.Switch("Zapp")
	if code(ref) != "save_failed" || !strings.Contains(ref.Error, "nothing changed") {
		t.Errorf("%v", ref)
	}
	if _, ok := ReadIntent(r.m.Inst); ok {
		t.Error("a failed leave must not leave an intent behind")
	}
	r.m.Leave = nil
	if _, ref := r.m.Switch("Zapp"); ref == nil || ref.Status != 501 {
		t.Errorf("a backend that cannot leave: %v", ref)
	}
	r.m.Leave = func(context.Context) error { return nil }
	r.m.Verify = nil // the real check: a header-only image is not a valid save
	if _, ref := r.m.Switch("Zapp"); code(ref) != "corrupt" {
		t.Errorf("%v", ref)
	}
}

func TestCreateValidation(t *testing.T) {
	r := newRig(t)
	for _, c := range []struct{ name, class, want string }{
		{"A", "paladin", "bad_name"}, {"Holy1", "paladin", "bad_name"}, {"Newb", "warlock", "bad_class"},
		{"zapp", "sorceress", "exists"}, {"HOLY", "paladin", "exists"},
	} {
		if _, ref := r.m.Create(c.name, c.class, false); code(ref) != c.want {
			t.Errorf("%+v: %v", c, ref)
		}
	}
	if _, ref := r.m.Create("Newb", "paladin", true); code(ref) != "hardcore_unbuilt" {
		t.Errorf("hardcore: %v", ref)
	}
	v, ref := r.m.Create("Newb", "Sorceress", false)
	if ref != nil {
		t.Fatal(ref)
	}
	in, _ := ReadIntent(r.m.Inst)
	if !in.Create || in.To != "Newb" || in.ClassIndex != 4 || v.(Result).Action != "create" {
		t.Errorf("sorceress create: %+v", in)
	}
	ClearIntent(r.m.Inst)
	if _, ref := r.m.Create("Pally", "paladin", false); ref != nil {
		t.Fatal(ref)
	}
	if in, _ := ReadIntent(r.m.Inst); in.ClassIndex != 2 {
		t.Errorf("paladin d2d index: %+v", in)
	}
}

func TestSwapStateParksAndRestoresPerCharacter(t *testing.T) {
	inst := t.TempDir()
	w := func(rel, s string) {
		os.MkdirAll(filepath.Dir(filepath.Join(inst, rel)), 0o755)
		if err := os.WriteFile(filepath.Join(inst, rel), []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	rd := func(rel string) string { b, _ := os.ReadFile(filepath.Join(inst, rel)); return string(b) }
	w("goals.json", "holy-goals")
	w("ledger.json", "holy-ledger")
	w("state/tuning.json", "holy-tuning")
	w("state/onboarding.json", "shared")
	w("state/journey-v1.json", "holy-journey")
	if err := SwapState(inst, "Holy", "Zapp"); err != nil {
		t.Fatal(err)
	}
	if rd("goals.json") != "" || rd("state/tuning.json") != "" || rd("state/journey-v1.json") != "" {
		t.Error("Zapp starts with Holy's state")
	}
	if rd("state/onboarding.json") != "shared" {
		t.Error("shared state must stay")
	}
	w("goals.json", "zapp-goals")
	if err := SwapState(inst, "Zapp", "Holy"); err != nil {
		t.Fatal(err)
	}
	if rd("goals.json") != "holy-goals" || rd("ledger.json") != "holy-ledger" || rd("state/journey-v1.json") != "holy-journey" {
		t.Error("Holy's state did not come back")
	}
	if rd("characters/Zapp/goals.json") != "zapp-goals" {
		t.Error("Zapp's state is not parked")
	}
	if err := SwapState(inst, "Holy", "../evil"); err == nil {
		t.Error("a path-like target must be refused")
	}
	// a conflict puts everything back
	w("characters/Zapp/ledger.json", "z")
	w("goals.json", "live-goals-again")
	os.Remove(filepath.Join(inst, "characters/Zapp/goals.json"))
	w("characters/Zapp/goals.json", "zapp-goals")
	w("ledger.json", "live")
	os.MkdirAll(filepath.Join(inst, "characters", "Holy"), 0o755)
	w("characters/Holy/goals.json", "stale") // parking Holy would overwrite this
	if err := SwapState(inst, "Holy", "Zapp"); err == nil {
		t.Error("expected a conflict")
	}
	if rd("goals.json") != "live-goals-again" || rd("ledger.json") != "live" {
		t.Error("a failed swap must leave the live state in place")
	}
}

func TestStaleIntentIsDropped(t *testing.T) {
	r := newRig(t)
	if err := writeIntent(r.m.Inst, Intent{Schema: IntentSchema, From: "Holy", To: "Zapp", ClassIndex: -1, At: "2026-11-15T11:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	if _, ref := r.m.Delete("Zapp", "Zapp"); ref != nil { // the clock says 12:00, the intent is an hour old
		t.Errorf("a stale intent blocked a delete: %v", ref)
	}
	if _, ok := ReadIntent(r.m.Inst); ok {
		t.Error("stale intent not removed")
	}
}
