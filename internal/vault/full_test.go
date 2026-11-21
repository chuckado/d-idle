package vault

import (
	"errors"
	"testing"
)

// Decision 2026-11-21: a full vault removes nothing and does not record new drops; loot.vaultAutoTrim (AutoTrim)
// switches the old trim back on. Two heroes (Sorceress and Paladin vaults).
func TestFullVaultNeverTrimsByDefault(t *testing.T) {
	for _, hero := range []string{"Sorc-Frost", "Pala-Hammer"} {
		t.Run(hero, func(t *testing.T) {
			s, err := Open(t.TempDir(), hero, Options{Capacity: func() int { return 4 }, Now: fixed, NoSync: true})
			must(t, err)
			var ids []string
			for i := 1; i <= 4; i++ {
				r, err := s.Add(mk(i, "cap", 2, 5))
				must(t, err)
				ids = append(ids, r.ID)
			}
			if _, err := s.Add(mk(9, "wsp", 7, 80)); !errors.Is(err, ErrFull) {
				t.Fatalf("a worthwhile drop at capacity must be refused, got %v", err)
			}
			if st := s.Status(); st.Count != 4 {
				t.Fatalf("count %d", st.Count)
			}
			for _, id := range ids {
				if r, _ := s.Get(id); r.State != Recorded {
					t.Fatalf("%s left the vault: %s", id, r.State)
				}
			}
			if a, _ := s.Archived(); len(a) != 0 {
				t.Fatalf("archive not empty: %d", len(a))
			}
			must(t, s.Miss(mk(9, "wsp", 7, 80)))
			must(t, s.Miss(mk(10, "cap", 4, 30)))
			m := s.MissedDrops()
			if m.Count != 2 || len(m.Best) != 2 || m.Best[0].Quality != 7 {
				t.Fatalf("missed %+v", m)
			}
			must(t, s.Close())
			s2, err := Open(s.dir, hero, Options{Capacity: func() int { return 4 }, Now: fixed, NoSync: true})
			must(t, err)
			if m := s2.MissedDrops(); m.Count != 2 {
				t.Fatalf("missed not persisted: %+v", m)
			}
			must(t, s2.ClearMissed())
			if s2.MissedDrops().Count != 0 {
				t.Fatal("clear")
			}
		})
	}
}

func TestAutoTrimOnReproducesOldTrim(t *testing.T) {
	for _, hero := range []string{"Sorc-Frost", "Pala-Hammer"} {
		t.Run(hero, func(t *testing.T) {
			s, err := Open(t.TempDir(), hero, Options{AutoTrim: func() bool { return true }, Capacity: func() int { return 3 }, Now: fixed, NoSync: true})
			must(t, err)
			low, _ := s.Add(mk(1, "cap", 2, 5))
			s.Add(mk(2, "cap", 4, 30))
			s.Add(mk(3, "cap", 4, 20))
			if _, err := s.Add(mk(9, "cap", 4, 25)); err != nil {
				t.Fatal(err)
			}
			if r, _ := s.Get(low.ID); r.State != Trimmed {
				t.Fatalf("lowest not trimmed: %+v", r)
			}
		})
	}
}

func TestSellMarksWaitForConfirm(t *testing.T) {
	s := open(t, t.TempDir(), "Pala-Hammer", 10)
	a, _ := s.Add(mk(1, "cap", 2, 5))
	b, _ := s.Add(mk(2, "cap", 2, 5))
	must(t, s.MarkSell(a.ID))
	must(t, s.OrderSell(b.ID))
	if m := s.SellMarked(); len(m) != 1 || m[0].ID != a.ID {
		t.Fatalf("marked %v", m)
	}
	n, err := s.ConfirmSell()
	if err != nil || n != 1 || len(s.SellMarked()) != 0 {
		t.Fatalf("confirm %d %v", n, err)
	}
	must(t, s.Close())
	s2 := open(t, s.dir, "Pala-Hammer", 10)
	if r, _ := s2.Get(a.ID); !r.Ordered || r.State != SellQ {
		t.Fatalf("not persisted %+v", r)
	}
}
