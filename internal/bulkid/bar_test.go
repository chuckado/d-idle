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
	in := BarIn{Charges: 12, EtherPer: 1, EtherBal: 665, GoldOn: true, ScrollGold: 80, Purse: 300, Stash: 7200, Reserve: 500, ConfirmAt: 2000}
	// 29 items, 29 sealed: scrolls cover 12, Ether covers 29, gold covers 29 at (29-12)*80
	b := MakeBar(cands(29, 29), in)
	if b.Default != PayEther {
		t.Fatalf("default %q want ether: %+v", b.Default, b.Options)
	}
	if g := opt(t, b, PayGold); g.Cost != 17*80 || g.Count != 29 || g.Confirm {
		t.Fatalf("gold %+v", g)
	}
	// nothing sealed: Ether covers none, scrolls 12, gold all 29: gold is the only full payer
	b = MakeBar(cands(29, 0), in)
	if b.Default != PayGold {
		t.Fatalf("default %q want gold", b.Default)
	}
	if !strings.Contains(b.Line, "Scrolls: 12") || !strings.Contains(b.Line, "Gold: 1,360 (stash 7,200)") {
		t.Fatalf("line %q", b.Line)
	}
	// Cain here: free covers all, no other option is offered
	in.Free = true
	b = MakeBar(cands(29, 0), in)
	if b.Default != PayFree || len(b.Options) != 1 || b.Options[0].Count != 29 {
		t.Fatalf("cain: %+v", b)
	}
}

func TestBarGoldRespectsReserveCapAndConfirm(t *testing.T) {
	in := BarIn{GoldOn: true, ScrollGold: 100, Purse: 1000, Stash: 0, Reserve: 650, ConfirmAt: 2000}
	b := MakeBar(cands(10, 0), in)
	g := opt(t, b, PayGold)
	if g.Count != 3 || g.Cost != 300 || len(g.Skipped) != 1 || g.Skipped[0].N != 7 || !strings.Contains(g.Skipped[0].Why, "reserve") {
		t.Fatalf("short gold: %+v", g)
	}
	if b.Default != PayGold {
		t.Fatalf("a partial gold payer still beats nothing: %q", b.Default)
	}
	in.Purse, in.Stash = 9000, 9000
	b = MakeBar(cands(200, 0), in)
	g = opt(t, b, PayGold)
	if g.Count != GoldMax || g.Cost != GoldMax*100 || !g.Confirm || g.Skipped[len(g.Skipped)-1].N != 50 {
		t.Fatalf("cap/confirm: %+v", g)
	}
	// below the reserve: not ok, says why, no default
	in.Purse, in.Stash = 100, 0
	b = MakeBar(cands(5, 0), in)
	if g = opt(t, b, PayGold); g.OK || g.Why == "" || b.Default != "" {
		t.Fatalf("broke: %+v default %q", g, b.Default)
	}
}

func TestBarGoldOffOrRunning(t *testing.T) {
	in := BarIn{ScrollGold: 80, Purse: 5000, Reserve: 0}
	if b := MakeBar(cands(3, 0), in); len(b.Options) != 1 {
		t.Fatalf("gold off must not be offered: %+v", b.Options)
	}
	in.GoldOn, in.ActiveGoldOrderID = true, "o1"
	b := MakeBar(cands(3, 0), in)
	if g := opt(t, b, PayGold); g.OK || !strings.Contains(g.Why, "already running") {
		t.Fatalf("%+v", g)
	}
	if got := Thousands(1234567); got != "1,234,567" {
		t.Fatal(got)
	}
}
