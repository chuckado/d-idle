package saves

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixture writes a synthetic character (never a real save) with the usual companion files.
func fixture(t *testing.T, dir, name string) {
	t.Helper()
	b := build(name, 0xE7DB542B, func(h []byte) { h[0x2B] = 11; h[0x28] = 1; h[0xA8] = 0; h[0xA9] = 0x81 })
	f, err := Verify(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SetStat(13, 123456); err != nil {
		t.Fatal(err)
	}
	if err := f.SetStat(14, 777); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		name + ".d2s": f.Bytes(), name + ".key": []byte("keys"), name + ".map": []byte("0c0000000100000042"),
		name + ".ma0": []byte("a0"), name + ".ma1": nil, "Other.d2s": []byte("not ours"),
	}
	for fn, c := range files {
		if err := os.WriteFile(filepath.Join(dir, fn), c, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

var libT0 = time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

func snapOne(t *testing.T) (src, lib string, m Manifest) {
	t.Helper()
	src, lib = t.TempDir(), filepath.Join(t.TempDir(), "lib")
	fixture(t, src, "Fresha")
	m, err := LibSnapshot(src, lib, "Fresha", SnapInfo{BotCommit: "abc123", SourceRun: "r1", SourceInst: "i5", LedgerStatus: "honest"}, Options{AtMenu: true}, libT0)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestLibSnapshotManifest(t *testing.T) {
	_, lib, m := snapOne(t)
	if m.ID != "Fresha-20261102T100000Z" || m.Name != "Fresha" || m.ClassName != "sorceress" || m.Level != 11 ||
		m.XP != 123456 || m.Gold != 777 || m.Difficulty != "nightmare" || m.Act != 2 || m.MapID != 0xE7DB542B ||
		m.LedgerStatus != "honest" || m.BotCommit != "abc123" || m.KitHash == "" {
		t.Fatalf("manifest %+v", m)
	}
	if len(m.Files) != 5 { // d2s key map ma0 ma1; Other.d2s is another character
		t.Fatalf("files %v", m.Files)
	}
	got, err := LibList(lib)
	if err != nil || len(got) != 1 || got[0].ID != m.ID {
		t.Fatalf("list %v %v", got, err)
	}
	if _, _, err := LibVerify(lib, m.ID); err != nil {
		t.Fatal(err)
	}
}

func TestLibSnapshotRefusals(t *testing.T) {
	src, lib, m := snapOne(t)
	ok := Options{AtMenu: true}
	cases := []struct {
		name string
		fn   func() error
	}{
		{"in game", func() error {
			_, err := LibSnapshot(src, lib, "Fresha", SnapInfo{LedgerStatus: "honest"}, Options{}, libT0)
			return err
		}},
		{"existing id", func() error {
			_, err := LibSnapshot(src, lib, "Fresha", SnapInfo{ID: m.ID, LedgerStatus: "honest"}, ok, libT0)
			return err
		}},
		{"no status", func() error { _, err := LibSnapshot(src, lib, "Fresha", SnapInfo{ID: "x"}, ok, libT0); return err }},
		{"bad id", func() error {
			_, err := LibSnapshot(src, lib, "Fresha", SnapInfo{ID: "../x", LedgerStatus: "honest"}, ok, libT0)
			return err
		}},
		{"missing", func() error {
			_, err := LibSnapshot(src, lib, "Nobody", SnapInfo{LedgerStatus: "honest"}, ok, libT0)
			return err
		}},
		{"corrupt source", func() error {
			p := filepath.Join(src, "Fresha.d2s")
			b, _ := os.ReadFile(p)
			b[0x30] ^= 1
			os.WriteFile(p, b, 0o644)
			_, err := LibSnapshot(src, lib, "Fresha", SnapInfo{ID: "y", LedgerStatus: "honest"}, ok, libT0)
			return err
		}},
	}
	for _, c := range cases {
		if err := c.fn(); err == nil {
			t.Errorf("%s: no error", c.name)
		}
	}
}

func TestLibVerifyDetects(t *testing.T) {
	for _, c := range []struct {
		name string
		mut  func(dir, id string)
	}{
		{"flipped key byte", func(d, id string) { os.WriteFile(filepath.Join(d, "Fresha.key"), []byte("keyz"), 0o644) }},
		{"missing file", func(d, id string) { os.Remove(filepath.Join(d, "Fresha.ma0")) }},
		{"extra file", func(d, id string) { os.WriteFile(filepath.Join(d, "Fresha.ma2"), nil, 0o644) }},
		{"manifest id", func(d, id string) {
			b, _ := os.ReadFile(filepath.Join(d, ManifestName))
			os.WriteFile(filepath.Join(d, ManifestName), []byte(strings.Replace(string(b), id, "other", 1)), 0o644)
		}},
	} {
		_, lib, m := snapOne(t)
		c.mut(filepath.Join(lib, m.ID), m.ID)
		if _, _, err := LibVerify(lib, m.ID); err == nil {
			t.Errorf("%s: verify passed", c.name)
		}
	}
	_, lib, m := snapOne(t)
	d2 := filepath.Join(lib, m.ID, "Fresha.d2s")
	b, _ := os.ReadFile(d2)
	b[0x30] ^= 1 // corrupt .d2s and keep the manifest hash honest: the parse must still fail
	os.WriteFile(d2, b, 0o644)
	mb, _ := os.ReadFile(filepath.Join(lib, m.ID, ManifestName))
	mb = []byte(strings.Replace(string(mb), m.Files["Fresha.d2s"], sum(b), 1))
	os.WriteFile(filepath.Join(lib, m.ID, ManifestName), mb, 0o644)
	if _, _, err := LibVerify(lib, m.ID); err == nil {
		t.Error("corrupt .d2s with a matching hash passed")
	}
}

func TestRename(t *testing.T) {
	b := build("Fresha", 9, nil)
	for _, n := range []string{"Freshb", "Ab", "Fresh_ab", "Fresh-ab", "Abcdefghijklmno"} {
		out, err := Rename(b, n)
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		f, err := Verify(out)
		if err != nil || f.Name() != n || len(out) != len(b) {
			t.Fatalf("%s: %v %v", n, f, err)
		}
	}
	for _, n := range []string{"", "A", "Fresha1", "_ab", "ab_", "a b", "Abcdefghijklmnop", "a-b-c", "../x"} {
		if _, err := Rename(b, n); err == nil {
			t.Errorf("%q accepted", n)
		}
	}
	// only the name and the checksum differ
	out, _ := Rename(b, "Freshb")
	diff := 0
	for i := range b {
		if b[i] != out[i] && !(i >= 0x0C && i < 0x10) && !(i >= 0x14 && i < 0x24) {
			diff++
		}
	}
	if diff != 0 {
		t.Fatalf("%d unexpected byte changes", diff)
	}
	if _, err := Rename(b[:50], "Freshb"); err == nil {
		t.Fatal("corrupt input accepted")
	}
}

func TestPlace(t *testing.T) {
	src, lib, m := snapOne(t)
	dst := t.TempDir()
	p, err := Place(lib, m.ID, dst, "Freshb", false)
	if err != nil {
		t.Fatal(err)
	}
	if p.Provenance != "copy-of-honest" || len(p.Files) != 5 || p.SourceD2S == p.PlacedD2S {
		t.Fatalf("%+v", p)
	}
	f, err := VerifyFile(filepath.Join(dst, "Freshb.d2s"))
	if err != nil || f.Name() != "Freshb" || f.Level() != 11 {
		t.Fatalf("placed d2s %v %v", f, err)
	}
	for ext, want := range map[string]string{".key": "keys", ".map": "0c0000000100000042", ".ma0": "a0"} {
		if b, _ := os.ReadFile(filepath.Join(dst, "Freshb"+ext)); string(b) != want {
			t.Errorf("%s = %q", ext, b)
		}
	}
	ents, _ := os.ReadDir(dst)
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".") || !strings.HasPrefix(e.Name(), "Freshb.") {
			t.Errorf("stray file %s", e.Name())
		}
	}
	// the source prefix is untouched
	if b, _ := os.ReadFile(filepath.Join(src, "Fresha.key")); string(b) != "keys" {
		t.Error("source changed")
	}
	// the placed character re-snapshots to the same kit and level under its new name
	m2, err := LibSnapshot(dst, filepath.Join(t.TempDir(), "l2"), "Freshb", SnapInfo{LedgerStatus: "honest"}, Options{AtMenu: true}, libT0)
	if err != nil || m2.KitHash != m.KitHash || m2.XP != m.XP || m2.MapID != m.MapID {
		t.Fatalf("re-snapshot %+v %v", m2, err)
	}
}

func TestPlaceRefusals(t *testing.T) {
	_, lib, m := snapOne(t)
	for _, c := range []struct {
		name string
		pre  func(dst string)
		as   string
		want error
	}{
		{"existing d2s", func(d string) { os.WriteFile(filepath.Join(d, "Freshb.d2s"), []byte("x"), 0o644) }, "Freshb", ErrExists},
		{"existing other ext, other case", func(d string) { os.WriteFile(filepath.Join(d, "FRESHB.ma3"), nil, 0o644) }, "Freshb", ErrExists},
		{"same name", nil, "fresha", ErrSameName},
		{"bad name", nil, "Fresh1", ErrBadName},
	} {
		dst := t.TempDir()
		if c.pre != nil {
			c.pre(dst)
		}
		before, _ := os.ReadDir(dst)
		if _, err := Place(lib, m.ID, dst, c.as, false); !errors.Is(err, c.want) {
			t.Errorf("%s: %v", c.name, err)
		}
		if after, _ := os.ReadDir(dst); len(after) != len(before) {
			t.Errorf("%s: target changed", c.name)
		}
	}
	if _, err := Place(lib, m.ID, filepath.Join(t.TempDir(), "nope"), "Freshb", false); err == nil {
		t.Error("missing target dir accepted")
	}
	if _, err := Place(lib, "ghost", t.TempDir(), "Freshb", false); err == nil {
		t.Error("missing snapshot accepted")
	}
}

func TestPlaceCheatedSource(t *testing.T) {
	src, lib := t.TempDir(), t.TempDir()
	fixture(t, src, "Fresha")
	m, _ := LibSnapshot(src, lib, "Fresha", SnapInfo{LedgerStatus: "cheated"}, Options{AtMenu: true}, libT0)
	if _, err := Place(lib, m.ID, t.TempDir(), "Freshb", false); err == nil {
		t.Fatal("cheated source placed")
	}
	p, err := Place(lib, m.ID, t.TempDir(), "Freshb", true)
	if err != nil || p.Provenance != "copy-of-cheated" {
		t.Fatalf("%+v %v", p, err)
	}
}

func TestPlaceRollsBack(t *testing.T) {
	_, lib, m := snapOne(t)
	dst := t.TempDir()
	os.Mkdir(filepath.Join(dst, ".Freshb.ma0.place"), 0o755) // staging file cannot be written
	if _, err := Place(lib, m.ID, dst, "Freshb", false); err == nil {
		t.Fatal("expected failure")
	}
	os.Remove(filepath.Join(dst, ".Freshb.ma0.place"))
	if ents, _ := os.ReadDir(dst); len(ents) != 0 {
		t.Fatalf("leftovers: %v", ents)
	}
}
