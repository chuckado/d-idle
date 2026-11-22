package bulkid

import (
	"fmt"
	"strings"
	"testing"
)

func cands(n, sealed int) []Cand {
	var cs []Cand
	for i := 0; i < n; i++ {
		cs = append(cs, Cand{ID: fmt.Sprintf("v-%03d", i), Sealed: i < sealed})
	}
	return cs
}

func opt(t *testing.T, b Bar, id string) BarOption {
	t.Helper()
	for _, o := range b.Options {
		if o.ID == id {
			return o
		}
	}
	t.Fatalf("no option %q in %+v", id, b.Options)
	return BarOption{}
}

func TestBarDefaultsToTheCheapestPayerThatCoversAll(t *testing.T) {
	in := BarIn{Charges: 12, EtherPer: 1, EtherBal: 665, GoldOn: true, ScrollGold: 80, Stash: 7200, ConfirmAt: 2000}
	// 29 items, 29 sealed: scrolls cover 12, Ether covers 29, gold covers 29 at 80 each from the stash
	b := MakeBar(cands(29, 29), in)
	if b.Default != PayEther {
		t.Fatalf("default %q want ether: %+v", b.Default, b.Options)
	}
	if g := opt(t, b, PayGold); g.Cost != 29*80 || g.Count != 29 || !g.Confirm || g.Line != "Gold: 2,320 (stash 7,200)" {
		t.Fatalf("gold %+v", g)
	}
	// nothing sealed: Ether covers none, gold cannot unlock either (no held affixes), scrolls cover 12
	b = MakeBar(cands(29, 0), in)
	if g := opt(t, b, PayGold); g.OK || !strings.Contains(g.Why, "no affixes") || b.Default != PayFree {
		t.Fatalf("unsealed: gold %+v default %q", g, b.Default)
	}
	// Cain here: free covers all, no other option is offered
	in.Free = true
	b = MakeBar(cands(29, 0), in)
	if b.Default != PayFree || len(b.Options) != 1 || b.Options[0].Count != 29 {
		t.Fatalf("cain: %+v", b)
	}
}

func TestBarGoldFromTheStashCapAndConfirm(t *testing.T) {
	in := BarIn{GoldOn: true, ScrollGold: 100, Stash: 350, ConfirmAt: 2000}
	b := MakeBar(cands(10, 10), in)
	g := opt(t, b, PayGold)
	if g.Count != 3 || g.Cost != 300 || len(g.Skipped) != 1 || g.Skipped[0].N != 7 || !strings.Contains(g.Skipped[0].Why, "stash") {
		t.Fatalf("short gold: %+v", g)
	}
	if b.Default != PayGold || g.Line != "Gold: 300 (stash 350)" {
		t.Fatalf("a partial gold payer still beats nothing: %q %q", b.Default, g.Line)
	}
	in.Stash = 90000
	b = MakeBar(cands(200, 200), in)
	g = opt(t, b, PayGold)
	if g.Count != GoldMax || g.Cost != GoldMax*100 || !g.Confirm || g.Skipped[len(g.Skipped)-1].N != 50 {
		t.Fatalf("cap/confirm: %+v", g)
	}
	// an empty stash: not ok, says why, no default
	in.Stash = 50
	b = MakeBar(cands(10, 10), in)
	if g = opt(t, b, PayGold); g.OK || g.Why == "" || b.Default != "" {
		t.Fatalf("empty stash: %+v default %q", g, b.Default)
	}
}

func TestBarGoldOffOrRunning(t *testing.T) {
	in := BarIn{ScrollGold: 80, Stash: 5000}
	if b := MakeBar(cands(3, 3), in); len(b.Options) != 1 {
		t.Fatalf("gold off must not be offered: %+v", b.Options)
	}
	in.GoldOn, in.ActiveGoldOrderID = true, "o1"
	b := MakeBar(cands(3, 3), in)
	if g := opt(t, b, PayGold); g.OK || !strings.Contains(g.Why, "already running") {
		t.Fatalf("%+v", g)
	}
	if got := Thousands(1234567); got != "1,234,567" {
		t.Fatal(got)
	}
}

// With a fork that decodes (item decode) every unidentified candidate is gold-revealable, held affixes or not.
func TestBarGoldDecodeCoversSealedLessItems(t *testing.T) {
	in := BarIn{GoldOn: true, ScrollGold: 80, Stash: 7200, ConfirmAt: 2000}
	b := MakeBar(cands(29, 0), in)
	if g := opt(t, b, PayGold); g.OK {
		t.Fatalf("without decode nothing is revealable: %+v", g)
	}
	in.Decode = true
	b = MakeBar(cands(29, 0), in)
	if g := opt(t, b, PayGold); !g.OK || g.Count != 29 || g.Cost != 29*80 || len(g.Skipped) != 0 {
		t.Fatalf("with decode all 29: %+v", g)
	}
}

// Only keeps the offered payments and works the default and the line out again; nil keeps everything.
func TestBarOnlyKeepsTheOfferedPayments(t *testing.T) {
	in := BarIn{Charges: 12, EtherPer: 1, EtherBal: 665, GoldOn: true, ScrollGold: 80, Stash: 7200, Decode: true}
	b := MakeBar(cands(29, 0), in)
	if len(b.Only(nil).Options) != len(b.Options) {
		t.Fatal("nil keeps all")
	}
	g := b.Only([]string{PayGold})
	if len(g.Options) != 1 || g.Options[0].ID != PayGold || g.Default != PayGold || g.Line != "Gold: 2,320 (stash 7,200)" || g.Total != 29 {
		t.Fatalf("gold only: %+v", g)
	}
	// gold cannot pay (no stash gold): no default, the line still says Gold and no scroll or Cain word appears
	in.Stash = 0
	g = MakeBar(cands(3, 3), in).Only([]string{PayGold})
	if g.Default != "" || !strings.HasPrefix(g.Line, "Gold: 0 (stash 0)") || strings.Contains(strings.ToLower(g.Line), "scroll") {
		t.Fatalf("empty stash: %+v", g)
	}
}
