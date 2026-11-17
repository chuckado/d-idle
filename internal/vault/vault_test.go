package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)

// clock is deterministic so two runs of the same script give the same bytes.
type clock struct{ n int }

func (c *clock) now() time.Time { c.n++; return t0.Add(time.Duration(c.n) * time.Second) }

func fixed() time.Time { return t0 }

func mk(i int, code string, q, ilvl int) Item {
	return Item{Raw: fmt.Sprintf("%08x%02x", i, q), Code: code, Name: fmt.Sprintf("%s #%d", code, i), Quality: q, Ilvl: ilvl,
		Chain: []string{code, "weap"}, Found: Found{Area: "Burial Grounds", MonsterClass: "zombie"}}
}

func open(t testing.TB, dir, hero string, capN int) *Store {
	t.Helper()
	s, err := Open(dir, hero, Options{Capacity: func() int { return capN }, Now: fixed, NoSync: true})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func dump(s *Store) string { b, _ := json.Marshal(s.Dump()); return string(b) }

func TestAddDedupeAndGeneration(t *testing.T) {
	s := open(t, t.TempDir(), "sorc", 10)
	r, err := s.Add(mk(1, "wsp", 6, 14))
	if err != nil || !strings.HasPrefix(r.ID, "v-") || r.Hero != "sorc" || r.Value == 0 {
		t.Fatalf("add: %+v %v", r, err)
	}
	if _, err := s.Add(mk(1, "wsp", 6, 14)); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("dup: %v", err)
	}
	if err := s.Discard(r.ID, "junk"); err != nil {
		t.Fatal(err)
	}
	r2, err := s.Add(mk(1, "wsp", 6, 14)) // same bytes again after the first finished: a new generation
	if err != nil || r2.ID == r.ID || !strings.HasSuffix(r2.ID, "~1") {
		t.Fatalf("regen: %+v %v", r2, err)
	}
	for _, bad := range []Item{{}, {Raw: "zz", Code: "a"}, {Raw: "00", Code: "a", Quality: 99}} {
		if _, err := s.Add(bad); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid %+v: %v", bad, err)
		}
	}
}

func TestStateMachine(t *testing.T) {
	s := open(t, t.TempDir(), "sorc", 50)
	add := func(i int) string { r, _ := s.Add(mk(i, "cap", 4, 10)); return r.ID }
	a := add(1)
	mustErr := func(err error, want error) {
		t.Helper()
		if !errors.Is(err, want) {
			t.Fatalf("got %v want %v", err, want)
		}
	}
	mustErr(s.CompleteKeep(a), ErrState)
	mustErr(s.Sold(a, 5), ErrState)
	mustErr(s.Pin("nope"), ErrNotFound)
	must(t, s.Keep(a))
	mustErr(s.Keep(a), ErrState)
	mustErr(s.Discard(a, "x"), ErrState)
	mustErr(s.Pin(a), ErrState)
	must(t, s.CancelKeep(a))
	must(t, s.Keep(a))
	must(t, s.BeginKeep(a))
	mustErr(s.BeginKeep(a), ErrState) // never a second create
	mustErr(s.CancelKeep(a), ErrState)
	must(t, s.FailKeep(a, "nospace"))
	if r, _ := s.Get(a); r.State != Recorded || r.Reason != "nospace" || r.InFlight {
		t.Fatalf("%+v", r)
	}
	must(t, s.Keep(a))
	must(t, s.BeginKeep(a))
	must(t, s.CompleteKeep(a))
	mustErr(s.Keep(a), ErrState)
	b := add(2)
	must(t, s.MarkSell(b))
	must(t, s.FailSell(b, "shop closed"))
	must(t, s.MarkSell(b))
	must(t, s.Sold(b, 120))
	if r, _ := s.Get(b); r.State != Sold || r.Gold != 120 {
		t.Fatalf("%+v", r)
	}
	if st := s.Status(); st.Count != 0 {
		t.Fatalf("final states still count: %+v", st)
	}
}

func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRevealStatusAndPins(t *testing.T) {
	s := open(t, t.TempDir(), "pala", 8) // pin limit 25% = 2
	var ids []string
	for i := 0; i < 4; i++ {
		r, _ := s.Add(mk(i, "ltp", 6, 20))
		ids = append(ids, r.ID)
	}
	must(t, s.Pin(ids[0]))
	must(t, s.Pin(ids[1]))
	if err := s.Pin(ids[2]); !errors.Is(err, ErrPinCap) {
		t.Fatal(err)
	}
	must(t, s.Pin(ids[0])) // pinning twice is fine
	must(t, s.Unpin(ids[1]))
	if r, _ := s.Get(ids[0]); r.Status() != "pinned" {
		t.Fatal(r.Status())
	}
	id := mk(2, "ltp", 6, 20)
	id.Identified, id.Raw, id.Name = true, "ffff0000", "Doom Plate"
	must(t, s.Reveal(ids[2], id))
	if r, _ := s.Get(ids[2]); r.Status() != "revealed" || r.Raw != "ffff0000" || r.ID != ids[2] {
		t.Fatalf("%+v", r)
	}
	if err := s.Reveal(ids[3], id); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("reveal onto an existing item: %v", err)
	}
}

func TestCapacityTrimOrderProtectedAndPins(t *testing.T) {
	capN := 4
	c := &clock{}
	s, _ := Open(t.TempDir(), "sorc", Options{Capacity: func() int { return capN }, Now: c.now, NoSync: true})
	rare, _ := s.Add(mk(1, "wsp", 6, 40)) // protected (rare)
	low, _ := s.Add(mk(2, "cap", 2, 5))   // normal, lowest value
	mag1, _ := s.Add(mk(3, "cap", 4, 30))
	mag2, _ := s.Add(mk(4, "cap", 4, 10))
	must(t, s.Pin(mag1.ID))
	// full: the new item displaces the lowest value record, not the pinned or the rare
	if _, err := s.Add(mk(5, "cap", 4, 20)); err != nil {
		t.Fatal(err)
	}
	if r, _ := s.Get(low.ID); r.State != Trimmed {
		t.Fatalf("lowest not trimmed: %+v", r)
	}
	arch, _ := s.Archived()
	if len(arch) != 1 || arch[0].ID != low.ID {
		t.Fatalf("archive before delete: %+v", arch)
	}
	// next trim: magic ilvl 10 (mag2) before the rest; rare stays
	s.Add(mk(6, "cap", 4, 25))
	if r, _ := s.Get(mag2.ID); r.State != Trimmed {
		t.Fatalf("%+v", r)
	}
	// now live = rare(protected), mag1(pinned), 5, 6 : trim 5 then 6 then only protected/pinned remain
	s.Add(mk(7, "cap", 4, 26))
	s.Add(mk(8, "cap", 4, 27))
	for i := 9; i < 12; i++ {
		s.Add(mk(i, "cap", 4, 30+i))
	}
	if r, _ := s.Get(rare.ID); r.State != Recorded {
		t.Fatalf("protected rare trimmed: %+v", r)
	}
	if r, _ := s.Get(mag1.ID); r.State != Recorded || !r.Pin {
		t.Fatalf("pinned trimmed: %+v", r)
	}
	// make everything unpinned-trimmable protected: fill with rares only
	capN = 2
	s2 := open(t, t.TempDir(), "sorc", 2)
	s2.Add(mk(1, "wsp", 6, 40))
	p, _ := s2.Add(mk(2, "wsp", 7, 40))
	_ = p
	if _, err := s2.Add(mk(3, "cap", 4, 1)); !errors.Is(err, ErrFull) {
		t.Fatalf("expected ErrFull, got %v", err)
	}
	if st := s2.Status(); !st.Full || st.Count != 2 {
		t.Fatalf("%+v", st)
	}
}

func TestShrinkNeverDeletes(t *testing.T) {
	capN := 6
	dir := t.TempDir()
	c := &clock{}
	s, _ := Open(dir, "sorc", Options{Capacity: func() int { return capN }, Now: c.now, NoSync: true})
	var ids []string
	for i := 0; i < 6; i++ {
		r, _ := s.Add(mk(i, "cap", 4, i))
		ids = append(ids, r.ID)
	}
	capN = 3 // shrink
	if st := s.Status(); !st.Over {
		t.Fatalf("%+v", st)
	}
	if _, err := s.Add(mk(99, "cap", 4, 50)); !errors.Is(err, ErrOverCapacity) {
		t.Fatal(err)
	}
	if got, _ := s.Query(Filter{}); len(got) != 6 {
		t.Fatalf("shrink deleted: %d", len(got))
	}
	arch, _ := s.Archived()
	if len(arch) != 0 {
		t.Fatal("shrink archived something")
	}
	must(t, s.Discard(ids[0], "x"))
	must(t, s.Discard(ids[1], "x"))
	if _, err := s.Add(mk(99, "cap", 4, 50)); !errors.Is(err, ErrOverCapacity) { // 4 > 3 still over
		t.Fatal(err)
	}
	must(t, s.Discard(ids[2], "x")) // 3 == cap: exact full, a new insert trims one
	if _, err := s.Add(mk(99, "cap", 4, 50)); err != nil {
		t.Fatal(err)
	}
	if st := s.Status(); st.Count != 3 {
		t.Fatalf("%+v", st)
	}
}

func TestTiers(t *testing.T) {
	ti, err := DefaultTiers()
	if err != nil {
		t.Fatal(err)
	}
	if ti.Capacity(nil, 0) != ti.Base || ti.Capacity([]string{"v1"}, 0) != ti.Base+ti.Steps[0].Add {
		t.Fatal("tiers")
	}
	all := []string{}
	for _, s := range ti.Steps {
		all = append(all, s.ID)
	}
	if ti.Capacity(all, 0) != ti.Max || ti.Capacity([]string{"gone"}, 0) != ti.Base || ti.Capacity(nil, 7) != 7 {
		t.Fatal("max/unknown/override")
	}
	if n, ok := ti.Next([]string{"v1"}); !ok || n.ID != "v2" {
		t.Fatal(n)
	}
	if _, ok := ti.Next(all); ok {
		t.Fatal("next after all")
	}
	if _, err := ParseTiers([]byte(`{"base":0}`)); err == nil {
		t.Fatal("bad table accepted")
	}
	c, err := ti.CapacityFrom(fakeAcct{[]string{"v1", "v2"}}, 0)
	if err != nil || c != ti.Base+ti.Steps[0].Add+ti.Steps[1].Add {
		t.Fatal(c, err)
	}
	if c, _ := ti.CapacityFrom(nil, 0); c != ti.Base {
		t.Fatal(c)
	}
}

type fakeAcct struct{ own []string }

func (f fakeAcct) OwnedSteps() ([]string, error) { return f.own, nil }

func TestKeepCrashRecoveryNoDoubleCreate(t *testing.T) {
	dir := t.TempDir()
	s := open(t, dir, "sorc", 20)
	a, _ := s.Add(mk(1, "wsp", 6, 14))
	b, _ := s.Add(mk(2, "wsp", 6, 14))
	must(t, s.Keep(a.ID))
	must(t, s.Keep(b.ID))
	must(t, s.BeginKeep(a.ID))
	must(t, s.BeginKeep(b.ID))
	// process dies here: no Close, no checkpoint
	s2 := open(t, dir, "sorc", 20)
	creates := 0
	stash := map[string]bool{a.Hash: true} // a's create happened, b's did not
	done, requeued, err := s2.Recover(func(r Rec) (bool, error) { return stash[r.Hash], nil })
	if err != nil || len(done) != 1 || len(requeued) != 1 || done[0] != a.ID || requeued[0] != b.ID {
		t.Fatalf("%v %v %v", done, requeued, err)
	}
	// the requeued one goes through the normal path exactly once
	must(t, s2.BeginKeep(b.ID))
	creates++
	must(t, s2.CompleteKeep(b.ID))
	// a second recover is a no-op; a third process sees both kept
	if d, r, _ := s2.Recover(func(Rec) (bool, error) { t.Fatal("asked again"); return false, nil }); len(d)+len(r) != 0 {
		t.Fatal("recover not idempotent")
	}
	s3 := open(t, dir, "sorc", 20)
	for _, id := range []string{a.ID, b.ID} {
		if r, _ := s3.Get(id); r.State != Kept || r.InFlight {
			t.Fatalf("%+v", r)
		}
	}
	if creates != 1 {
		t.Fatal(creates)
	}
	// a probe error leaves the record in flight
	c, _ := s3.Add(mk(3, "wsp", 6, 14))
	must(t, s3.Keep(c.ID))
	must(t, s3.BeginKeep(c.ID))
	if _, _, err := s3.Recover(func(Rec) (bool, error) { return false, errors.New("game gone") }); err == nil {
		t.Fatal("probe error swallowed")
	}
	if r, _ := s3.Get(c.ID); !r.InFlight {
		t.Fatal("in-flight mark lost on probe error")
	}
}

// script is a fixed mix of every event type; each step appends exactly one log line.
func script() []func(s *Store) error {
	var ids [8]string
	add := func(k int, code string, q, ilvl int) func(*Store) error {
		return func(s *Store) error {
			r, err := s.Add(mk(k, code, q, ilvl))
			if err == nil {
				ids[k] = r.ID
			}
			return err
		}
	}
	idv := func(k int) string { return "v-" + hashOf(mk(k, "x", 0, 0).Raw) }
	_ = idv
	return []func(*Store) error{
		add(0, "wsp", 6, 14), add(1, "cap", 4, 10), add(2, "cap", 2, 3), add(3, "ltp", 7, 30),
		func(s *Store) error { return s.Pin(ids[0]) },
		func(s *Store) error { return s.Keep(ids[1]) },
		func(s *Store) error { return s.BeginKeep(ids[1]) },
		func(s *Store) error { return s.CompleteKeep(ids[1]) },
		func(s *Store) error { return s.MarkSell(ids[2]) },
		func(s *Store) error { return s.Sold(ids[2], 33) },
		add(4, "cap", 4, 12),
		func(s *Store) error { return s.Discard(ids[4], "meh") },
		add(5, "cap", 3, 12), add(6, "cap", 4, 13), // 4 live: wsp, ltp, 5, 6 => full at capacity 4
		add(7, "cap", 4, 14), // trims the lowest
		func(s *Store) error { return s.Unpin(ids[0]) },
		func(s *Store) error { return s.Keep(ids[3]) },
		func(s *Store) error { return s.CancelKeep(ids[3]) },
	}
}

func TestCrashBetweenEverySteps(t *testing.T) {
	steps := script()
	// reference run, no crash, state after each step
	ref := open(t, t.TempDir(), "sorc", 4)
	var want []string
	want = append(want, dump(ref))
	for i, st := range steps {
		if err := st(ref); err != nil {
			t.Fatalf("reference step %d: %v", i, err)
		}
		want = append(want, dump(ref))
	}
	for _, partial := range []bool{false, true} {
		for k := range steps {
			dir := t.TempDir()
			s := open(t, dir, "sorc", 4)
			for i := 0; i < k; i++ {
				if err := steps[i](s); err != nil {
					t.Fatal(err)
				}
			}
			if k%3 == 0 {
				must(t, s.Checkpoint()) // some crashes happen with a snapshot behind the log
			}
			s.hook = func(line []byte) ([]byte, error) {
				if partial {
					return line[:len(line)/2], errors.New("crash")
				}
				return nil, errors.New("crash")
			}
			if err := steps[k](s); err == nil {
				t.Fatalf("step %d did not crash", k)
			}
			// the process is dead; a new one replays
			s2 := open(t, dir, "sorc", 4)
			if got := dump(s2); got != want[k] {
				t.Fatalf("partial=%v crash at %d: replay differs from the state before the step", partial, k)
			}
			if st := s2.Status(); st.Repaired != partial {
				t.Fatalf("repaired=%v partial=%v", st.Repaired, partial)
			}
			// resume: the remaining steps (including the crashed one) land on the reference end state
			for i := k; i < len(steps); i++ {
				if err := steps[i](s2); err != nil {
					t.Fatalf("resume step %d: %v", i, err)
				}
			}
			if got := dump(s2); got != want[len(steps)] {
				t.Fatalf("partial=%v crash at %d: resumed end state differs", partial, k)
			}
			// replay equals live
			must(t, s2.Close())
			if got := dump(open(t, dir, "sorc", 4)); got != want[len(steps)] {
				t.Fatal("replay after close differs")
			}
		}
	}
}

func TestCrashBetweenArchiveAndLog(t *testing.T) {
	dir := t.TempDir()
	s := open(t, dir, "sorc", 4)
	r, _ := s.Add(mk(1, "cap", 4, 4))
	s.hook = func([]byte) ([]byte, error) { return nil, errors.New("crash") }
	if err := s.Discard(r.ID, "x"); err == nil {
		t.Fatal("no crash")
	}
	s2 := open(t, dir, "sorc", 4)
	if g, _ := s2.Get(r.ID); g.State != Recorded {
		t.Fatalf("a discard that never reached the log took effect: %+v", g)
	}
	must(t, s2.Discard(r.ID, "x")) // retried: archive has the record twice on disk, once logically
	if a, _ := s2.Archived(); len(a) != 1 {
		t.Fatalf("archive entries: %d", len(a))
	}
}

func TestSnapshotIsOnlyACache(t *testing.T) {
	dir := t.TempDir()
	s := open(t, dir, "sorc", 100)
	for i := 0; i < 30; i++ {
		s.Add(mk(i, "cap", 4, i))
	}
	must(t, s.Checkpoint())
	for i := 30; i < 40; i++ { // tail after the snapshot
		s.Add(mk(i, "cap", 4, i))
	}
	want := dump(s)
	s.f.Close()
	if got := dump(open(t, dir, "sorc", 100)); got != want {
		t.Fatal("snapshot + tail differs")
	}
	for _, damage := range []func(string){
		func(p string) { os.Remove(p) },
		func(p string) { os.WriteFile(p, []byte("{not json"), 0o644) },
		func(p string) {
			b, _ := os.ReadFile(p)
			os.WriteFile(p, bytes.Replace(b, []byte(`"lastLine":"`), []byte(`"lastLine":"00`), 1), 0o644)
		},
	} {
		damage(filepath.Join(dir, IndexFile))
		if got := dump(open(t, dir, "sorc", 100)); got != want {
			t.Fatal("damaged snapshot changed the state")
		}
	}
	// a snapshot from another hero is ignored
	s2 := open(t, t.TempDir(), "pala", 100)
	s2.Add(mk(1, "cap", 4, 1))
	must(t, s2.Checkpoint())
	b, _ := os.ReadFile(filepath.Join(s2.dir, IndexFile))
	os.WriteFile(filepath.Join(dir, IndexFile), b, 0o644)
	if got := dump(open(t, dir, "sorc", 100)); got != want {
		t.Fatal("foreign snapshot used")
	}
}

func TestRandomOpsReplayEquals(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		dir := t.TempDir()
		s := open(t, dir, "sorc", 15)
		var ids []string
		for n := 0; n < 300; n++ {
			switch rng.Intn(9) {
			case 0, 1, 2:
				r, err := s.Add(mk(rng.Intn(60), []string{"cap", "wsp", "ltp"}[rng.Intn(3)], 1+rng.Intn(7), rng.Intn(60)))
				if err == nil {
					ids = append(ids, r.ID)
				}
			default:
				if len(ids) == 0 {
					continue
				}
				id := ids[rng.Intn(len(ids))]
				switch rng.Intn(8) {
				case 0:
					s.Pin(id)
				case 1:
					s.Unpin(id)
				case 2:
					s.Keep(id)
				case 3:
					s.BeginKeep(id)
				case 4:
					s.CompleteKeep(id)
				case 5:
					s.Discard(id, "r")
				case 6:
					s.MarkSell(id)
				case 7:
					s.Sold(id, 1)
				}
			}
			if st := s.Status(); st.Count > st.Capacity {
				t.Fatalf("seed %d: over capacity without a shrink: %+v", seed, st)
			}
			if n%97 == 0 {
				s.Checkpoint()
			}
		}
		want := dump(s)
		s.f.Close()
		s2 := open(t, dir, "sorc", 15)
		if dump(s2) != want || s2.Status().Anomalies != 0 {
			t.Fatalf("seed %d: replay differs", seed)
		}
		// live counters equal a recount
		n, pins := 0, 0
		for _, r := range s2.Dump() {
			if r.State.Live() {
				n++
			}
			if r.Pin {
				pins++
			}
		}
		if st := s2.Status(); st.Count != n || st.Pins != pins {
			t.Fatalf("seed %d: counters %+v vs %d/%d", seed, st, n, pins)
		}
	}
}

func TestTwoBuildsOwnStores(t *testing.T) {
	root := t.TempDir()
	sorc, err := OpenHero(root, "Sorc-Frost", Options{NoSync: true})
	must(t, err)
	pala, err := OpenHero(root, "Pala-Hammer", Options{NoSync: true})
	must(t, err)
	// the same item bytes may live in both vaults: dedupe is per hero
	a, err := sorc.Add(mk(1, "ltp", 6, 30))
	must(t, err)
	b, err := pala.Add(mk(1, "ltp", 6, 30))
	must(t, err)
	if a.Hero == b.Hero || a.ID != b.ID {
		t.Fatalf("%+v %+v", a, b)
	}
	sorc.Add(mk(2, "orb", 4, 20))
	pala.Add(mk(3, "pal", 4, 20))
	must(t, sorc.Close())
	must(t, pala.Close())
	if _, err := os.Stat(filepath.Join(HeroDir(root, "Pala-Hammer"), LogFile)); err != nil {
		t.Fatal(err)
	}
	s2, _ := OpenHero(root, "Sorc-Frost", Options{NoSync: true})
	if got, n := s2.Query(Filter{}); n != 2 || got[0].Hero != "Sorc-Frost" {
		t.Fatalf("%+v", got)
	}
	// default capacity comes from the table
	if st := s2.Status(); st.Capacity != 400 || st.PinLimit != 100 {
		t.Fatalf("%+v", st)
	}
	for _, bad := range []string{"", "..", "a/b", `a\b`, "a:b"} {
		if _, err := OpenHero(root, bad, Options{}); err == nil {
			t.Fatalf("hero %q accepted", bad)
		}
	}
}

func TestQueryFilters(t *testing.T) {
	s := open(t, t.TempDir(), "sorc", 100)
	f := false
	tr := true
	one := 1
	a := mk(1, "wsp", 6, 40)
	a.Sockets, a.Eth, a.Name = 1, true, "Rare Short Sword"
	b := mk(2, "cap", 4, 10)
	b.Identified = true
	c := mk(3, "cap", 2, 5)
	for _, it := range []Item{a, b, c} {
		if _, err := s.Add(it); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		f    Filter
		want int
	}{
		{Filter{}, 3}, {Filter{Quality: []int{4, 6}}, 2}, {Filter{Code: "cap"}, 2}, {Filter{Type: "weap"}, 3},
		{Filter{MinIlvl: 10}, 2}, {Filter{MaxIlvl: 10}, 2}, {Filter{Sockets: &one}, 1}, {Filter{Eth: &tr}, 1},
		{Filter{Eth: &f}, 2}, {Filter{Identified: &tr}, 1}, {Filter{Name: "short SWORD"}, 1}, {Filter{Name: "zzz"}, 0},
		{Filter{Code: "cap", Quality: []int{2}}, 1}, {Filter{States: []State{Kept}}, 0},
	}
	for i, c := range cases {
		if _, n := s.Query(c.f); n != c.want {
			t.Fatalf("case %d %+v: got %d want %d", i, c.f, n, c.want)
		}
	}
	got, total := s.Query(Filter{Order: "value", Limit: 2})
	if total != 3 || len(got) != 2 || got[0].Code != "wsp" {
		t.Fatalf("%+v", got)
	}
	if got, _ := s.Query(Filter{Offset: 1, Limit: 1}); len(got) != 1 || got[0].Code != "cap" {
		t.Fatal(got)
	}
}

func TestExportImport(t *testing.T) {
	src := open(t, t.TempDir(), "sorc", 100)
	for i := 0; i < 5; i++ {
		src.Add(mk(i, "cap", 4, i))
	}
	var buf bytes.Buffer
	must(t, src.Export(&buf))
	dst := open(t, t.TempDir(), "pala", 100)
	res, err := dst.Import(bytes.NewReader(buf.Bytes()))
	if err != nil || res.Added != 5 {
		t.Fatalf("%+v %v", res, err)
	}
	if res, _ = dst.Import(bytes.NewReader(buf.Bytes())); res.Duplicate != 5 || res.Added != 0 {
		t.Fatalf("%+v", res)
	}
	small := open(t, t.TempDir(), "pala", 3)
	if res, _ = small.Import(bytes.NewReader(buf.Bytes())); res.Added != 3 || res.Blocked != 2 {
		t.Fatalf("%+v", res)
	}
	if arch, _ := small.Archived(); len(arch) != 0 {
		t.Fatal("import trimmed")
	}
	for name, body := range map[string]string{
		"junk":    "{",
		"version": `{"version":9,"items":[]}`,
		"badhex":  `{"version":1,"items":[{"raw":"xx","code":"a"}]}`,
		"nocode":  `{"version":1,"items":[{"raw":"00"}]}`,
	} {
		if _, err := dst.Import(strings.NewReader(body)); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
	// one bad item rejects the whole file before anything is added
	before := dst.Status().Count
	dst.Import(strings.NewReader(`{"version":1,"items":[{"raw":"aa01","code":"cap","q":4},{"raw":"zz","code":"a"}]}`))
	if dst.Status().Count != before {
		t.Fatal("partial import")
	}
	big := `{"version":1,"items":[` + strings.Repeat(`{"raw":"aa","code":"c"},`, MaxImportItems+1) + `{"raw":"aa","code":"c"}]}`
	if _, err := dst.Import(strings.NewReader(big)); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
	if _, err := dst.Import(&zeros{n: MaxImportBytes + 10}); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
}

type zeros struct{ n int }

func (z *zeros) Read(p []byte) (int, error) {
	if z.n <= 0 {
		return 0, os.ErrClosed
	}
	n := min(len(p), z.n)
	for i := 0; i < n; i++ {
		p[i] = ' '
	}
	z.n -= n
	return n, nil
}

func TestPurgeArchiveOnlyTouchesArchive(t *testing.T) {
	dir := t.TempDir()
	c := &clock{}
	s, err := Open(dir, "sorc", Options{Capacity: func() int { return 100 }, Now: c.now, NoSync: true})
	must(t, err)
	a, _ := s.Add(mk(1, "cap", 4, 4))
	b, _ := s.Add(mk(2, "cap", 4, 4))
	must(t, s.Discard(a.ID, "x"))
	must(t, s.Discard(b.ID, "x"))
	n, err := s.PurgeArchive(t0.Add(time.Duration(c.n-1) * time.Second)) // only the older archive entry is before this
	must(t, err)
	if n < 1 {
		t.Fatalf("removed %d", n)
	}
	if r, ok := s.Get(a.ID); !ok || r.State != Discarded {
		t.Fatal("purge touched the log state")
	}
}

func TestValueTableData(t *testing.T) {
	v, err := DefaultValues()
	must(t, err)
	hi := &Item{Quality: 7, Ilvl: 50, Code: "ltp"}
	lo := &Item{Quality: 2, Ilvl: 50, Code: "cap"}
	if v.Of(hi) <= v.Of(lo) || v.Of(hi) != v.Of(hi) {
		t.Fatal("value order")
	}
	if !v.Protected(hi) || v.Protected(lo) || !v.Protected(&Item{Code: "r12", Quality: 2}) {
		t.Fatal("protected")
	}
	if !reflect.DeepEqual(hi.Chain, []string(nil)) {
		t.Fatal()
	}
}
