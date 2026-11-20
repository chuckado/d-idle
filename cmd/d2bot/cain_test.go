package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dsrc"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/quest"

	"d2lod-bot/internal/talk"
	"d2lod-bot/internal/task"
)

// The Act 1 chain (bark, stones, portal, cage) and the Rescue cages must name only things the adapter knows.
func TestAct1CampaignNamesResolve(t *testing.T) {
	r := &recorder{talks: map[[2]string]bool{}, objects: map[string]bool{}, bosses: map[string]bool{}}
	c := task.Template("normal_act1", true)
	for _, e := range c.Entries {
		for _, n := range task.Handler(e.Quest) {
			_ = n.Step.Do(context.Background(), r)
		}
	}
	for _, o := range []string{"Tree of Inifuss", CairnStones, "Cain's Cage"} {
		if !r.objects[o] {
			t.Errorf("the Act 1 campaign never asked for %q", o)
		}
	}
	for _, k := range [][2]string{{task.NPCAkara, "bark"}, {task.NPCCharsi, "imbue:best"}} {
		if !r.talks[k] {
			t.Errorf("the Act 1 campaign never asked for %v", k)
		}
	}
	for k := range r.talks {
		if _, ok := talk.Paths[k]; !ok {
			t.Errorf("no talk path for %v", k)
		}
	}
	for o := range r.objects {
		if _, ok := objSpecs[o]; !ok && o != CairnStones {
			t.Errorf("object %q is not implemented", o)
		}
	}
	for _, o := range []string{"Tristram portal", "cage"} {
		if _, ok := objSpecs[o]; !ok {
			t.Errorf("object %q missing", o)
		}
	}
}

func TestOnceSetIsPerAreaAndPerObject(t *testing.T) {
	var s *onceSet
	s.add(1, 1) // nil is a no-op
	if s.has(1, 1) {
		t.Fatal("nil set remembers")
	}
	s = &onceSet{}
	s.add(25, 7)
	s.add(25, 8)
	if !s.has(25, 7) || !s.has(25, 8) || s.has(25, 9) || s.has(26, 7) {
		t.Fatal("per area and id")
	}
	s.add(38, 7) // new area forgets the old one
	if s.has(25, 7) || !s.has(38, 7) {
		t.Fatal("area change must reset")
	}
}

// fakeStones is the game of lab drill L1 (V): the order is a permutation of the five stones fixed at the first click; a
// click on the stone that is next lights it, every other click (wrong position, lit stone) is ignored and nothing
// resets; the fifth lit stone opens the portal. needKey false is the missing Key to the Cairn Stones: no click works.
type fakeStones struct {
	order     []int
	next      int
	lit       [5]bool
	clicks    int
	blind     bool // the lit state is not readable
	needKey   bool
	portalLag int // Portal answers false this many more times after the fifth stone
	portal    bool
}

func (f *fakeStones) Click(_ context.Context, i int) error {
	f.clicks++
	if f.needKey || f.next >= 5 || i != f.order[f.next] {
		return nil
	}
	f.lit[i] = true
	if f.next++; f.next == 5 {
		f.portal = true
	}
	return nil
}

func (f *fakeStones) Lit(context.Context) ([]bool, bool, error) {
	if f.blind {
		return nil, false, nil
	}
	return append([]bool(nil), f.lit[:]...), true, nil
}

func (f *fakeStones) Portal(context.Context) bool {
	if f.portal && f.portalLag > 0 {
		f.portalLag--
		return false
	}
	return f.portal
}

func permutations(n int) [][]int {
	if n == 1 {
		return [][]int{{0}}
	}
	var out [][]int
	for _, p := range permutations(n - 1) {
		for at := 0; at <= len(p); at++ {
			q := append(append(append([]int{}, p[:at]...), n-1), p[at:]...)
			out = append(out, q)
		}
	}
	return out
}

// Every one of the 120 orders opens the portal within the bound, with the lit state readable (15 clicks) and without
// it (25), and a lagging portal object does not fail the search.
func TestSolveCairnAllOrders(t *testing.T) {
	perms := permutations(5)
	if len(perms) != 120 {
		t.Fatalf("%d orders", len(perms))
	}
	worst := map[bool]int{}
	for _, blind := range []bool{false, true} {
		for _, p := range perms {
			f := &fakeStones{order: p, blind: blind, portalLag: 2}
			if err := solveCairn(context.Background(), f, 5); err != nil || !f.portal {
				t.Fatalf("order %v blind %v: %v", p, blind, err)
			}
			if f.clicks > worst[blind] {
				worst[blind] = f.clicks
			}
		}
	}
	if worst[false] > maxStoneClicks || worst[true] > maxBlindClicks {
		t.Fatalf("worst clicks %v", worst)
	}
	// the known bound is tight (the reverse order needs 5+4+3+2+1); blind it is 5+5+5+5+3 before the portal shows
	if worst[false] != maxStoneClicks || worst[true] != 23 {
		t.Fatalf("worst clicks %v", worst)
	}
}

// A wrong click resets nothing: no stone is ever clicked again once lit (the old model re-clicked the prefix).
func TestSolveCairnNeverReclicksLitStones(t *testing.T) {
	for _, p := range permutations(5) {
		rc := &recClicks{fakeStones: fakeStones{order: p}}
		if err := solveCairn(context.Background(), rc, 5); err != nil {
			t.Fatal(err)
		}
		if rc.reclicked != 0 {
			t.Fatalf("order %v: %d clicks on an already lit stone", p, rc.reclicked)
		}
	}
}

type recClicks struct {
	fakeStones
	reclicked int
}

func (r *recClicks) Click(ctx context.Context, i int) error {
	if r.lit[i] {
		r.reclicked++
	}
	return r.fakeStones.Click(ctx, i)
}

// Without the Key to the Cairn Stones every click is ignored: the search stops after one cycle with a clear error that
// names the lit set and the key, not after 15 clicks and not silently.
func TestSolveCairnWithoutKeyFailsClearly(t *testing.T) {
	f := &fakeStones{order: []int{0, 1, 2, 3, 4}, needKey: true}
	err := solveCairn(context.Background(), f, 5)
	if err == nil || !strings.Contains(err.Error(), "Key to the Cairn Stones") || !strings.Contains(err.Error(), "00000") {
		t.Fatalf("err = %v", err)
	}
	if f.clicks != 5 {
		t.Fatalf("%d clicks, want one cycle of 5", f.clicks)
	}
}

// Without a lit state and without a key the blind search ends at its bound.
func TestSolveCairnBlindWithoutKeyStopsAtBound(t *testing.T) {
	f := &fakeStones{order: []int{0, 1, 2, 3, 4}, needKey: true, blind: true}
	err := solveCairn(context.Background(), f, 5)
	if err == nil || !strings.Contains(err.Error(), "25 clicks") || f.clicks != maxBlindClicks {
		t.Fatalf("err = %v clicks %d", err, f.clicks)
	}
}

// A portal that is there already (a restart in the same game) needs no click.
func TestSolveCairnPortalAlreadyOpen(t *testing.T) {
	f := &fakeStones{order: []int{0, 1, 2, 3, 4}, portal: true}
	if err := solveCairn(context.Background(), f, 5); err != nil || f.clicks != 0 {
		t.Fatalf("err %v clicks %d", err, f.clicks)
	}
}

// All stones lit but no portal ever: an error, not a loop.
func TestSolveCairnAllLitNoPortal(t *testing.T) {
	f := &fakeStones{order: []int{0, 1, 2, 3, 4}}
	f.lit = [5]bool{true, true, true, true, true}
	if err := solveCairn(context.Background(), f, 5); err == nil || !strings.Contains(err.Error(), "no portal") {
		t.Fatalf("err = %v", err)
	}
}

func TestSolveCairnStopsOnCancelAndReadErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := solveCairn(ctx, &fakeStones{order: []int{0, 1, 2, 3, 4}}, 5); err == nil {
		t.Fatal("cancelled context ignored")
	}
	if err := solveCairn(context.Background(), &brokenStones{}, 5); err == nil || !strings.Contains(err.Error(), "left the view") {
		t.Fatalf("err = %v", err)
	}
	f := &fakeStones{order: []int{0, 1, 2, 3, 4}}
	if err := solveCairn(context.Background(), &shortStones{*f}, 5); err == nil {
		t.Fatal("a four stone read was accepted")
	}
}

type brokenStones struct{ fakeStones }

func (brokenStones) Lit(context.Context) ([]bool, bool, error) {
	return nil, false, errors.New("cairn stones: stone class 18 left the view")
}

type shortStones struct{ fakeStones }

func (s shortStones) Lit(context.Context) ([]bool, bool, error) { return make([]bool, 4), true, nil }

// liveStones reads the lit state from Unit.Mode: 0 untouched, above 0 lit, -1 on any stone is "not readable"; the
// stones are found by class, not by id. Two builds: the class tables are data (classes 17..21 for every character).
func TestLiveStonesLitFromMode(t *testing.T) {
	snap := &d2.Snapshot{Player: &d2.Player{Area: stonyFieldArea}}
	for i, c := range cairnClasses {
		snap.Objects = append(snap.Objects, d2.Unit{Type: d2.TypeObject, Class: c, ID: uint32(100 + 7*i), Mode: 0})
	}
	l := &liveStones{o: objectAdapter{e: &env.Env{Read: cainReader{snap: &snap}, Clock: cainClock{}}}}
	lit, known, err := l.Lit(context.Background())
	if err != nil || !known || litCount(lit) != 0 {
		t.Fatalf("fresh: %v %v %v", lit, known, err)
	}
	snap.Objects[3].Mode = 1
	snap.Objects[3].ID = 5 // ids shift as rooms load
	if lit, _, _ = l.Lit(context.Background()); !lit[3] || litCount(lit) != 1 {
		t.Fatalf("one lit: %v", lit)
	}
	snap.Objects[1].Mode = -1
	if _, known, _ = l.Lit(context.Background()); known {
		t.Fatal("a stone with no state read as known")
	}
	snap.Objects[1].Mode = 0
	snap.Objects = snap.Objects[:4]
	if _, _, err = l.Lit(context.Background()); err == nil {
		t.Fatal("a missing stone was not an error")
	}
	if l.Portal(context.Background()) {
		t.Fatal("portal without a red portal object")
	}
	snap.Objects = append(snap.Objects, d2.Unit{Type: d2.TypeObject, Class: quest.ObjRedPortal, ID: uint32(d2dsrc.PortalIDBase + 2), Mode: 1})
	if !l.Portal(context.Background()) {
		t.Fatal("red portal not seen")
	}
}

type cainReader struct {
	d2.StateReader
	snap **d2.Snapshot
}

func (s cainReader) Snapshot() (*d2.Snapshot, error) { return *s.snap, nil }

type cainClock struct{}

func (cainClock) Now() time.Time      { return time.Unix(0, 0) }
func (cainClock) Sleep(time.Duration) {}

// The stones step answers a wrong area as a prereq, and the arrival check names the level it landed in.
func TestCairnStonesWrongAreaIsPrereq(t *testing.T) {
	snap := &d2.Snapshot{Player: &d2.Player{Area: 1}}
	o := objectAdapter{e: &env.Env{Read: cainReader{snap: &snap}, Clock: cainClock{}}}
	err := o.cairnStones(context.Background())
	if task.Classify(err) != task.KindPrereq {
		t.Fatalf("err = %v kind %v", err, task.Classify(err))
	}
	snap.Player.Area = tristramArea
	if err := o.cairnStones(context.Background()); err != nil {
		t.Fatalf("already in Tristram: %v", err)
	}
}
