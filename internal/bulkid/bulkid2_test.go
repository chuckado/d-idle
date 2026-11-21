package bulkid

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"d2lod-bot/internal/env"
)

// Must-fail: a process that dies between stage and return leaves nothing lost; the next start (any policy) puts the
// items back from the file journal.
func TestJournalRecoveryAfterRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bulkid.json")
	for _, build := range []string{"Sorceress", "Paladin"} {
		_ = os.Remove(path)
		cw := tripWorld(build)
		total0 := total(cw.fakeMover)
		j, err := FileJournal(path)
		if err != nil {
			t.Fatal(err)
		}
		tr := newTrip(t, cw, env.DefaultIdentifyTuning(), false)
		tr.S = &Stager{M: cw, J: j, Chunk: 10}
		// the process dies once the items are staged: every call after the stage fails
		cw.dieAfter = 2*len(cw.stash) + 2
		_, _ = tr.Run(context.Background())
		cw.dieAfter = 0
		if total(cw.fakeMover) != total0 {
			t.Fatalf("%s: items lost during the crash", build)
		}
		j2, err := FileJournal(path) // the restart reads the file
		if err != nil {
			t.Fatal(err)
		}
		if len(j2.IDs) == 0 {
			t.Fatalf("%s: items left the stash but the journal is empty", build)
		}
		tu := env.DefaultIdentifyTuning()
		tu.Policy = env.IdentifyManual // recovery does not depend on the policy
		tr2 := newTrip(t, cw, tu, false)
		tr2.S = &Stager{M: cw, J: j2, Chunk: 10}
		if _, err := tr2.Run(context.Background()); err != nil {
			t.Fatal(err)
		}
		if len(j2.IDs) != 0 || total(cw.fakeMover) != total0 {
			t.Fatalf("%s: journal %v total %d/%d", build, j2.IDs, total(cw.fakeMover), total0)
		}
		for _, it := range cw.pack {
			if it.Place == PlaceStash {
				t.Fatalf("%s: stash item %d stayed in the pack", build, it.ID)
			}
		}
	}
}

func TestDamagedJournalIsAnError(t *testing.T) {
	p := filepath.Join(t.TempDir(), "j.json")
	_ = os.WriteFile(p, []byte("{nope"), 0o644)
	if _, err := FileJournal(p); err == nil {
		t.Fatal("a damaged journal must not load as empty")
	}
}

func TestCainRescued(t *testing.T) {
	for w, want := range map[uint16]bool{0x0010: false, 0x2012: true, 0x2011: true, 0: false} {
		if CainRescued(w) != want {
			t.Fatalf("%#x", w)
		}
	}
}

type fakeScrolls struct {
	charges, gold, price int
	bought               int
	ident                []uint32
}

func (f *fakeScrolls) Charges(context.Context) int { return f.charges }
func (f *fakeScrolls) Gold(context.Context) int    { return f.gold }
func (f *fakeScrolls) Identify(_ context.Context, id uint32) error {
	if f.charges == 0 {
		return ErrNoTool
	}
	f.charges--
	f.ident = append(f.ident, id)
	return nil
}
func (f *fakeScrolls) Buy(_ context.Context, n int) (int, error) {
	f.bought += n
	f.charges += n
	f.gold -= n * f.price
	return n, nil
}

func picks(n int) []Pick {
	var p []Pick
	for i := 1; i <= n; i++ {
		p = append(p, Pick{Item: Item{ID: uint32(i)}, Via: ViaScroll})
	}
	return p
}

func scrollTu() env.IdentifyTuning {
	tu := env.DefaultIdentifyTuning()
	tu.Policy = env.IdentifyScrolls
	return tu
}

// Must-fail: nothing is bought unless town.identifyTome allows it.
func TestScrollsNeverBuyWithoutPermission(t *testing.T) {
	f := &fakeScrolls{charges: 1, gold: 5000, price: 80}
	rep := RunScrolls(context.Background(), scrollTu(), picks(3), 2, 80, false, f)
	if f.bought != 0 || rep.Identified != 1 {
		t.Fatalf("%+v bought %d", rep, f.bought)
	}
}

// Must-fail: a purchase never takes the purse below identify.goldFloor.
func TestScrollsGoldFloor(t *testing.T) {
	f := &fakeScrolls{gold: 600 + 80*2 + 10, price: 80}
	rep := RunScrolls(context.Background(), scrollTu(), picks(5), 5, 80, true, f)
	if f.bought != 2 || f.gold < 600 || rep.Identified != 2 {
		t.Fatalf("%+v bought %d gold %d", rep, f.bought, f.gold)
	}
	g := &fakeScrolls{gold: 650, price: 80}
	RunScrolls(context.Background(), scrollTu(), picks(3), 3, 80, true, g)
	if g.bought != 0 {
		t.Fatalf("bought %d with 50 gold above the floor", g.bought)
	}
}

// Must-fail: the per-trip cap bounds the charges used.
func TestScrollsPerTripCap(t *testing.T) {
	tu := scrollTu()
	tu.MaxScrolls = 2
	f := &fakeScrolls{charges: 20, gold: 9999}
	rep := RunScrolls(context.Background(), tu, picks(6), 0, 80, true, f)
	if rep.Spent != 2 || len(f.ident) != 2 || rep.Reason != WhyCap {
		t.Fatalf("%+v", rep)
	}
}

func TestScrollsOnlyUnderPolicy2(t *testing.T) {
	for _, pol := range []int{env.IdentifyManual, env.IdentifyCainFree, env.IdentifyNever} {
		tu := scrollTu()
		tu.Policy = pol
		f := &fakeScrolls{charges: 5, gold: 9999}
		if rep := RunScrolls(context.Background(), tu, picks(3), 0, 80, true, f); rep.Identified != 0 || len(f.ident) != 0 {
			t.Fatalf("policy %d identified", pol)
		}
	}
}

type fakeWallet struct {
	bal   int
	spent map[string]int
}

func (w *fakeWallet) Balance() (int, error) { return w.bal, nil }
func (w *fakeWallet) Spend(key string, n int) error {
	if _, ok := w.spent[key]; ok {
		return nil
	}
	if n > w.bal {
		return errors.New("short")
	}
	w.bal -= n
	w.spent[key] = n
	return nil
}

// Must-fail: Ether short refuses and spends nothing; Ether off (price 0) refuses; a repeated charge is one spend.
func TestEtherPayer(t *testing.T) {
	w := &fakeWallet{bal: 5, spent: map[string]int{}}
	price := 3
	p := &EtherPayer{W: w, Price: func() int { return price }}
	if err := p.Check(2); !errors.Is(err, ErrEtherShort) || w.bal != 5 {
		t.Fatalf("%v bal %d", err, w.bal)
	}
	if err := p.Charge("a"); err != nil || w.bal != 2 {
		t.Fatalf("%v bal %d", err, w.bal)
	}
	if err := p.Charge("a"); err != nil && !errors.Is(err, ErrEtherShort) {
		t.Fatal(err)
	}
	if len(w.spent) != 1 {
		t.Fatalf("%v", w.spent)
	}
	if err := p.Charge("b"); !errors.Is(err, ErrEtherShort) || len(w.spent) != 1 {
		t.Fatalf("short charge: %v %v", err, w.spent)
	}
	price = 0
	if err := p.Check(1); !errors.Is(err, ErrEtherOff) {
		t.Fatalf("%v", err)
	}
	if (*EtherPayer)(nil).Enabled() {
		t.Fatal("nil payer enabled")
	}
}
