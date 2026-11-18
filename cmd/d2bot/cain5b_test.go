package main

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/quest"
	"d2lod-bot/internal/saves"
)

// Fixtures of the lab drill L2 (docs/probes/search-for-cain-l2-2026-11-18.md, V on the d2d lab build).
const cainL2 = "../../docs/probes/cainl2/testdata"

// rowsOf returns the tab separated fields of every line of a fixture that starts with the prefix.
func rowsOf(t *testing.T, name, prefix string) [][]string {
	t.Helper()
	f, err := os.Open(filepath.Join(cainL2, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), prefix) {
			out = append(out, strings.Split(sc.Text(), "\t"))
		}
	}
	return out
}

type wordSrc uint16

func (w wordSrc) Word(int, string) (uint16, bool) { return uint16(w), true }
func (wordSrc) Evidence() quest.Evidence          { return quest.EvMemory }

// The quest word of Search for Cain through the reader the bot uses: arrival 0x0010 is started, the open cage 0x2012 is
// "reward pending" (the free_cain evidence, the saveQuestReader maps it to Cleared), the report 0x2011 is done. The
// words are the fixtures' (V), not literals of this test.
func TestCainWordsFromFixtures(t *testing.T) {
	word := func(name string, last bool) uint16 {
		rows := rowsOf(t, name, "4\t0x")
		if len(rows) == 0 {
			t.Fatalf("%s: no record 4 row", name)
		}
		r := rows[0]
		if last {
			r = rows[len(rows)-1]
		}
		v, err := strconv.ParseUint(strings.TrimPrefix(r[1], "0x"), 16, 16)
		if err != nil {
			t.Fatal(err)
		}
		return uint16(v)
	}
	cases := []struct {
		fix     string
		last    bool
		status  quest.Status
		pending bool
	}{
		{"a1.txt", false, quest.Started, false}, // arrival
		{"g1.txt", true, quest.Started, true},   // cage open
		{"r1.txt", true, quest.Done, false},     // reported to Akara
	}
	for _, c := range cases {
		w := word(c.fix, c.last)
		info := quest.NewReader(wordSrc(w)).Quest(0, "search_for_cain")
		if info.Status != c.status || info.Pending != c.pending {
			t.Errorf("%s word 0x%04x: status %v pending %v, want %v %v", c.fix, w, info.Status, info.Pending, c.status, c.pending)
		}
		if got := w&saves.QRewardPending != 0; got != c.pending {
			t.Errorf("%s word 0x%04x: bit 1 = %v, want %v", c.fix, w, got, c.pending)
		}
	}
}

// The gibbet row of `debug objects` is class quest.ObjCainGibbet, NU before and ON after; the approach cell of the
// adapter is within a cell of the point the drill walked to (25, 27), and about two cells from the cage.
func TestCageApproachFromFixtures(t *testing.T) {
	find := func(name string) []string {
		for _, r := range rowsOf(t, name, "") {
			if len(r) >= 7 && r[4] == strconv.Itoa(quest.ObjCainGibbet) {
				return r
			}
		}
		t.Fatalf("%s: no class %d row", name, quest.ObjCainGibbet)
		return nil
	}
	before, after := find("a1.txt"), find("e1.txt")
	if before[6] != "NU" || after[6] != "ON" {
		t.Fatalf("gibbet state %s -> %s, want NU -> ON", before[6], after[6])
	}
	cx, _ := strconv.ParseFloat(before[1], 64)
	cy, _ := strconv.ParseFloat(before[2], 64)
	off := objApproach[quest.ObjCainGibbet]
	ax := (float64(d2dctl.SubTile(cx)+off.X) + 0.5) / 5
	ay := (float64(d2dctl.SubTile(cy)+off.Y) + 0.5) / 5
	if math.Hypot(ax-25, ay-27) > 1 {
		t.Errorf("approach cell (%.1f, %.1f) is not near the walked-to (25, 27)", ax, ay)
	}
	if d := math.Hypot(ax-cx, ay-cy); d < 1.5 || d > 3 {
		t.Errorf("approach is %.1f cells from the cage, want about 2", d)
	}
	if objSpecs["Cain's Cage"].class != quest.ObjCainGibbet {
		t.Fatal("the cage spec must resolve by the gibbet class, not by an index")
	}
}
