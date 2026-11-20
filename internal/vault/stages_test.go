package vault

import (
	"errors"
	"testing"
)

// The reveal and sale stages (slice V6): logged, replayed, and guarding the transitions around them.
func TestRevealStagesReplay(t *testing.T) {
	dir := t.TempDir()
	s := open(t, dir, "sorc", 10)
	it := mk(1, "wsp", 4, 20)
	r, _ := s.Add(it)
	if err := s.RevealWant(r.ID); err != nil {
		t.Fatal(err)
	}
	for _, st := range []string{StageRevBegin, StageRevMade} {
		if err := s.SetStage(r.ID, st, 0); err != nil {
			t.Fatal(err)
		}
	}
	// the stage blocks the transitions that would strand the temporary copy
	if err := s.Discard(r.ID, "x"); !errors.Is(err, ErrState) {
		t.Fatalf("discard in flight: %v", err)
	}
	if err := s.Keep(r.ID); !errors.Is(err, ErrState) {
		t.Fatalf("keep in flight: %v", err)
	}
	if err := s.MarkSell(r.ID); !errors.Is(err, ErrState) {
		t.Fatalf("sell in flight: %v", err)
	}
	if err := s.RevealCancel(r.ID); !errors.Is(err, ErrState) {
		t.Fatalf("cancel in flight: %v", err)
	}
	idf := it
	idf.Raw, idf.Identified = it.Raw+"ff", true
	if err := s.Reveal(r.ID, idf); err != nil {
		t.Fatal(err)
	}
	// crash here: reopen replays the log; the cleanup stage must survive
	s.Close()
	s = open(t, dir, "sorc", 10)
	got, _ := s.Get(r.ID)
	if got.Raw != idf.Raw || !got.Identified || got.RevQ || got.Stage != StageRevCleanup || got.Hash == r.Hash {
		t.Fatalf("after reopen: %+v", got)
	}
	if q := s.RevealQueued(); len(q) != 1 {
		t.Fatalf("cleanup must stay listed: %d", len(q))
	}
	if err := s.SetStage(r.ID, "", 0); err != nil {
		t.Fatal(err)
	}
	if got, _ = s.Get(r.ID); got.Stage != "" || len(s.RevealQueued()) != 0 {
		t.Fatalf("idle again: %+v", got)
	}
	if err := s.RevealWant(r.ID); err == nil {
		t.Fatal("an identified record cannot be revealed again")
	}
}

func TestRevealFailAndKeepCombined(t *testing.T) {
	s := open(t, t.TempDir(), "sorc", 10)
	r, _ := s.Add(mk(1, "wsp", 4, 20))
	_ = s.RevealWant(r.ID)
	if err := s.KeepTo(r.ID, ""); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get(r.ID); got.State != KeepQ || !got.RevQ {
		t.Fatalf("keep and reveal wanted together: %+v", got)
	}
	if err := s.RevealFail(r.ID, "reveal: no scroll"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get(r.ID); got.RevQ || got.Reason != "reveal: no scroll" || got.State != KeepQ {
		t.Fatalf("fail clears the wish only: %+v", got)
	}
}

func TestSellStagesAndPinned(t *testing.T) {
	dir := t.TempDir()
	s := open(t, dir, "sorc", 10)
	a, _ := s.Add(mk(1, "wsp", 4, 20))
	b, _ := s.Add(mk(2, "rin", 4, 20))
	_ = s.Pin(b.ID)
	if err := s.MarkSell(b.ID); !errors.Is(err, ErrPinned) {
		t.Fatalf("pinned: %v", err)
	}
	if err := s.MarkSell(a.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Pin(a.ID); err == nil {
		t.Fatal("a queued sale cannot be pinned")
	}
	if err := s.SetStage(a.ID, StageSellBegin, 120); err != nil {
		t.Fatal(err)
	}
	if err := s.CancelSell(a.ID); !errors.Is(err, ErrState) {
		t.Fatalf("cancel in flight: %v", err)
	}
	s.Close()
	s = open(t, dir, "sorc", 10)
	got, _ := s.Get(a.ID)
	if got.Stage != StageSellBegin || got.GoldPre != 120 || got.Attempts != 1 || got.State != SellQ {
		t.Fatalf("replay: %+v", got)
	}
	if err := s.SoldNote(a.ID, 33, "sell: settled"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Get(a.ID)
	if got.State != Sold || got.Gold != 33 || got.Stage != "" || got.State.Live() {
		t.Fatalf("sold: %+v", got)
	}
	if n := s.Status().Count; n != 1 {
		t.Fatalf("a sold record leaves the live count: %d", n)
	}
	// a failed sale returns to the vault and may be queued again
	c, _ := s.Add(mk(3, "amu", 4, 20))
	_ = s.MarkSell(c.ID)
	_ = s.SetStage(c.ID, StageSellBegin, 5)
	if err := s.FailSell(c.ID, "sell: refused"); err != nil {
		t.Fatal(err)
	}
	if got, _ = s.Get(c.ID); got.State != Recorded || got.Stage != "" || got.GoldPre != 0 {
		t.Fatalf("failed sale: %+v", got)
	}
	if err := s.MarkSell(c.ID); err != nil {
		t.Fatal(err)
	}
}
