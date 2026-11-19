package vaultkeep

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

var t0 = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

// classes: the queue never looks at the class; two heroes run every scenario to prove no rule is a Sorceress one.
var classes = []string{"sorceress", "paladin"}

// flaky wraps the fake and injects verb failures and a clock that moves per create.
type flaky struct {
	*d2dctl.VaultFake
	calls     int
	failAt    map[int]error // nth ItemCreate (1-based) fails with this error before the fake sees it
	clock     *time.Time
	step      time.Duration
	readBack  func(norm string) string // rewrites what ItemFull(id) reports
	vaultErr  error
	removedID []int
}

func (f *flaky) ItemCreate(ctx context.Context, h string, o d2dctl.CreateOpts) (d2dctl.CreateResult, error) {
	f.calls++
	if f.clock != nil {
		*f.clock = f.clock.Add(f.step)
	}
	if err := f.failAt[f.calls]; err != nil {
		return d2dctl.CreateResult{}, err
	}
	return f.VaultFake.ItemCreate(ctx, h, o)
}

func (f *flaky) ItemFull(ctx context.Context, id int) ([]d2dctl.VaultRecord, error) {
	recs, err := f.VaultFake.ItemFull(ctx, id)
	if f.readBack != nil && id >= 0 {
		for i := range recs {
			recs[i].Norm = f.readBack(recs[i].Norm)
		}
	}
	return recs, err
}

func (f *flaky) ItemVault(ctx context.Context, id int) (d2dctl.VaultRecord, error) {
	if f.vaultErr != nil {
		return d2dctl.VaultRecord{}, f.vaultErr
	}
	f.removedID = append(f.removedID, id)
	return f.VaultFake.ItemVault(ctx, id)
}

type rig struct {
	t     *testing.T
	dir   string
	hero  string
	st    *vault.Store
	api   *flaky
	k     *Keeper
	now   time.Time
	saves int
	saveE error
	notes []string
	ready string
	on    bool
	per   int
	tools *fakeTools
	vend  *fakeVendor
	ch    *chaos
}

func newRig(t *testing.T, class string) *rig {
	r := &rig{t: t, dir: t.TempDir(), hero: class, now: t0, on: true, per: 20}
	r.api = &flaky{VaultFake: d2dctl.NewVaultFake(class), clock: &r.now, step: 0}
	r.open()
	return r
}

// open (re)opens the store from disk and builds a fresh Keeper: a bot restart. The fake game survives it (the d2d keeps
// running), the Keeper's memory does not.
func (r *rig) open() {
	if r.st != nil {
		_ = r.st.Close()
	}
	st, err := vault.OpenHero(r.dir, r.hero, vault.Options{Capacity: func() int { return 100 }, NoSync: true, Now: func() time.Time { return t0 }})
	if err != nil {
		r.t.Fatal(err)
	}
	r.st = st
	r.k = &Keeper{
		Store: func() *vault.Store { return r.st }, API: r.api,
		Enabled: func() bool { return r.on }, Ready: func() string { return r.ready },
		Backup: func(context.Context) error { r.notes = append(r.notes, "backup"); return nil },
		Save:   func(context.Context) error { r.saves++; return r.saveE },
		Notice: func(s string) { r.notes = append(r.notes, s) }, Now: func() time.Time { return r.now },
		PerVisit: func() int { return r.per },
	}
	if r.ch == nil {
		r.ch = &chaos{limit: 1 << 30}
	}
	r.ch.reset()
	r.k.API = chaosAPI{r.api, r.ch}
	if r.tools != nil {
		r.k.Tools = chaosTools{r.tools, r.ch}
	}
	if r.vend != nil {
		r.k.Vendor = chaosVendor{r.vend, r.ch}
	}
}

func (r *rig) add(i int, code string, q int) string {
	r.t.Helper()
	rec := d2dctl.FakeRecord(i, code, q, 30+i, fmt.Sprintf("roll%d", i))
	it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: code, Name: fmt.Sprintf("%s %d", code, i), Quality: q, Ilvl: 30 + i, Identified: q <= 3})
	if err != nil {
		r.t.Fatal(err)
	}
	return it.ID
}

func (r *rig) enq(ids ...string) []Result {
	res := r.k.Enqueue(context.Background(), ids, "", false)
	for _, x := range res {
		if !x.OK {
			r.t.Fatalf("enqueue %s: %s", x.ID, x.Why)
		}
	}
	return res
}

func (r *rig) run() []string { return r.k.Run(context.Background()) }

func (r *rig) state(id string) vault.Rec {
	rec, ok := r.st.Get(id)
	if !ok {
		r.t.Fatalf("no record %s", id)
	}
	return rec
}

func TestQueueCreateReadback(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			a, b := r.add(1, "wsp", 6), r.add(2, "uap", 7)
			r.enq(a, b)
			if got := r.k.Info(); got.Queued != 2 || r.k.Pending() != 2 {
				t.Fatalf("info %+v", got)
			}
			if line := r.k.ProgressLine(); line != "Keeping 2 items: arrives at your next town visit" {
				t.Fatalf("progress %q", line)
			}
			r.run()
			for _, id := range []string{a, b} {
				rec := r.state(id)
				if rec.State != vault.Kept || !strings.HasPrefix(rec.Reason, "stash ") {
					t.Fatalf("%s: %s %q", id, rec.State, rec.Reason)
				}
			}
			// identical: what the hero carries is byte for byte what was recorded
			have, _ := r.api.VaultFake.ItemFull(context.Background(), -1)
			if len(have) != 2 || have[0].Norm != r.state(a).Raw || have[1].Norm != r.state(b).Raw {
				t.Fatalf("carried %v", have)
			}
			if st, _ := r.api.VaultStats(context.Background()); st.Creates != 2 || r.saves != 1 {
				t.Fatalf("creates %d saves %d", st.Creates, r.saves)
			}
			if r.k.ProgressLine() != "" || r.k.Pending() != 0 {
				t.Fatal("queue should be empty")
			}
		})
	}
}

func TestUnidentifiedStaysUnidentified(t *testing.T) {
	r := newRig(t, "paladin")
	id := r.add(1, "wsp", 6) // rare, unidentified in the vault: identified false is part of the recorded bytes
	rec, _ := r.st.Get(id)
	if rec.Identified {
		t.Fatal("fixture should be unidentified")
	}
	r.enq(id)
	r.run()
	have, _ := r.api.VaultFake.ItemFull(context.Background(), -1)
	if len(have) != 1 || have[0].Identified || have[0].Norm != rec.Raw {
		t.Fatalf("carried %+v", have)
	}
}

func TestDoublePressCreatesOnce(t *testing.T) {
	for _, class := range classes {
		r := newRig(t, class)
		id := r.add(1, "wsp", 6)
		first := r.enq(id)
		second := r.enq(id)
		if first[0].Already || !second[0].Already {
			t.Fatalf("already flags %v %v", first, second)
		}
		r.run()
		third := r.enq(id) // pressed again after it was kept
		if !third[0].Already {
			t.Fatal("kept record should answer already")
		}
		r.run()
		if st, _ := r.api.VaultStats(context.Background()); st.Creates != 1 {
			t.Fatalf("creates %d", st.Creates)
		}
	}
}

// A restart between BeginKeep and the create: the item is absent, the next Run queues it again and creates it once.
func TestRestartAfterBeginBeforeCreate(t *testing.T) {
	r := newRig(t, "sorceress")
	id := r.add(1, "wsp", 6)
	r.enq(id)
	if err := r.st.BeginKeep(id); err != nil {
		t.Fatal(err)
	}
	r.open()
	if !r.state(id).InFlight {
		t.Fatal("expected in flight after restart")
	}
	lines := r.run()
	if r.state(id).State != vault.Kept {
		t.Fatalf("state %s lines %v", r.state(id).State, lines)
	}
	if st, _ := r.api.VaultStats(context.Background()); st.Creates != 1 {
		t.Fatalf("creates %d", st.Creates)
	}
}

// A restart after the create and before the save and the kept line: the item is in the hero, so Run confirms and never
// creates a second one.
func TestRestartAfterCreateBeforeKept(t *testing.T) {
	for _, class := range classes {
		r := newRig(t, class)
		id := r.add(1, "wsp", 6)
		r.enq(id)
		rec := r.state(id)
		_ = r.st.BeginKeep(id)
		if _, err := r.api.VaultFake.ItemCreate(context.Background(), rec.Raw, d2dctl.CreateOpts{Key: keyPrefix + id}); err != nil {
			t.Fatal(err)
		}
		r.open()
		r.run()
		if r.state(id).State != vault.Kept {
			t.Fatalf("state %s", r.state(id).State)
		}
		if st, _ := r.api.VaultStats(context.Background()); st.Creates != 1 {
			t.Fatalf("creates %d, a restart must not create twice", st.Creates)
		}
		r.run() // and a further visit changes nothing
		if st, _ := r.api.VaultStats(context.Background()); st.Creates != 1 {
			t.Fatal("second visit created again")
		}
	}
}

// The save fails after the create: nothing is confirmed, nothing is created again, the next Run saves and confirms.
func TestSaveFailureThenRecover(t *testing.T) {
	r := newRig(t, "paladin")
	id := r.add(1, "wsp", 6)
	r.enq(id)
	r.saveE = errors.New("disk busy")
	lines := r.run()
	if r.state(id).State != vault.KeepQ || !r.state(id).InFlight || !strings.Contains(strings.Join(lines, "|"), "save failed") {
		t.Fatalf("state %+v lines %v", r.state(id), lines)
	}
	if r.k.Info().Waiting != ReasonSave {
		t.Fatalf("waiting %q", r.k.Info().Waiting)
	}
	r.saveE = nil
	r.run()
	if r.state(id).State != vault.Kept {
		t.Fatal("not confirmed after a good save")
	}
	if st, _ := r.api.VaultStats(context.Background()); st.Creates != 1 {
		t.Fatalf("creates %d", st.Creates)
	}
}

func TestStashFullKeepsQueuedThenResumes(t *testing.T) {
	for _, class := range classes {
		r := newRig(t, class)
		r.api.StashCols, r.api.StashRows = 2, 1 // two cells
		var ids []string
		for i := 1; i <= 3; i++ {
			ids = append(ids, r.add(i, "rin", 7))
		}
		r.enq(ids...)
		r.run()
		if r.state(ids[0]).State != vault.Kept || r.state(ids[1]).State != vault.Kept {
			t.Fatal("first two should fit")
		}
		third := r.state(ids[2])
		if third.State != vault.KeepQ || third.InFlight || Reason(third.Reason) != ReasonStashFull {
			t.Fatalf("third %+v", third)
		}
		in := r.k.Info()
		if !in.Full || r.k.NeedsText() != "Stash full: free space or upgrade. 1 kept item(s) wait." {
			t.Fatalf("info %+v needs %q", in, r.k.NeedsText())
		}
		if !strings.Contains(r.k.ProgressLine(), "stash is full") {
			t.Fatal(r.k.ProgressLine())
		}
		r.run() // a second visit with the stash still full
		noted := 0
		for _, n := range r.notes {
			if strings.Contains(n, "stash is full") {
				noted++
			}
		}
		if noted != 1 { // one notice per episode, not one per visit
			t.Fatalf("notices %d: %v", noted, r.notes)
		}
		// the player frees a cell
		r.api.Carried = r.api.Carried[1:]
		r.run()
		if r.state(ids[2]).State != vault.Kept || r.k.NeedsText() != "" {
			t.Fatalf("third %+v", r.state(ids[2]))
		}
	}
}

func TestVerbFailureMidBatchPartialSuccess(t *testing.T) {
	r := newRig(t, "sorceress")
	var ids []string
	for i := 1; i <= 3; i++ {
		ids = append(ids, r.add(i, "wsp", 6))
	}
	r.api.failAt = map[int]error{2: &d2dctl.VaultRefused{Code: d2dctl.VaultErrVersion, Why: "other codec"}}
	r.enq(ids...)
	r.run()
	if r.state(ids[0]).State != vault.Kept || r.state(ids[2]).State != vault.Kept {
		t.Fatal("the good ones must be kept")
	}
	bad := r.state(ids[1])
	if bad.State != vault.Recorded || Reason(bad.Reason) != ReasonVersion {
		t.Fatalf("failed one %+v", bad)
	}
	if st, _ := r.api.VaultStats(context.Background()); st.Creates != 2 {
		t.Fatalf("creates %d", st.Creates)
	}
}

func TestUnansweredCreateStaysInFlightAndIsLookedAt(t *testing.T) {
	r := newRig(t, "paladin")
	id := r.add(1, "wsp", 6)
	r.api.failAt = map[int]error{1: context.DeadlineExceeded}
	r.enq(id)
	r.run()
	if !r.state(id).InFlight || r.state(id).State != vault.KeepQ {
		t.Fatalf("%+v", r.state(id))
	}
	// the next visit: the item is not there, so it is queued again and created once
	r.run()
	if r.state(id).State != vault.Kept {
		t.Fatalf("%+v", r.state(id))
	}
	if st, _ := r.api.VaultStats(context.Background()); st.Creates != 1 {
		t.Fatalf("creates %d", st.Creates)
	}
}

func TestOrderIsInsertionOrder(t *testing.T) {
	r := newRig(t, "sorceress")
	a, b, c := r.add(1, "wsp", 6), r.add(2, "wsp", 6), r.add(3, "wsp", 6)
	r.enq(c, a, b) // pressed in another order: the queue still follows the vault order
	r.run()
	var got []string
	for _, it := range r.api.Carried {
		got = append(got, it.Rec.Norm)
	}
	want := []string{r.state(a).Raw, r.state(b).Raw, r.state(c).Raw}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order %d", i)
		}
	}
}

func TestPerVisitCapAndBudget(t *testing.T) {
	r := newRig(t, "paladin")
	for i := 1; i <= 5; i++ {
		r.enq(r.add(i, "wsp", 6))
	}
	r.per = 3
	r.run()
	if n := r.k.Pending(); n != 2 {
		t.Fatalf("pending %d after the cap", n)
	}
	r.per = 20
	r.run()
	if r.k.Pending() != 0 {
		t.Fatal("rest should follow")
	}

	// a slow game: each create takes 3 s of the fake clock, the budget is 5 s, so a Run yields after two
	r2 := newRig(t, "sorceress")
	r2.api.step = 3 * time.Second
	r2.k.Budget = 5 * time.Second
	for i := 1; i <= 5; i++ {
		r2.enq(r2.add(i, "wsp", 6))
	}
	r2.run()
	if n := r2.k.Pending(); n != 3 {
		t.Fatalf("pending %d, want 3 (two created inside the budget)", n)
	}
}

func TestNotReadyAndOffCreateNothing(t *testing.T) {
	r := newRig(t, "sorceress")
	id := r.add(1, "wsp", 6)
	r.enq(id)
	r.ready = "not in town"
	r.run()
	if r.state(id).State != vault.KeepQ || r.k.Info().Waiting != "not in town" {
		t.Fatalf("%+v %+v", r.state(id), r.k.Info())
	}
	if !strings.Contains(r.k.ProgressLine(), "waiting, not in town") {
		t.Fatal(r.k.ProgressLine())
	}
	r.ready = ""
	r.on = false // measured run or kill switch
	r.run()
	if r.state(id).State != vault.KeepQ {
		t.Fatal("created while off")
	}
	if ok, why := r.k.Gate(context.Background()); ok || why != ReasonOff {
		t.Fatalf("gate %v %q", ok, why)
	}
	if st, _ := r.api.VaultStats(context.Background()); st.Creates != 0 {
		t.Fatal("creates while off")
	}
}

func TestDeadHeroRefusedAtGateAndNeverCreated(t *testing.T) {
	r := newRig(t, "paladin")
	id := r.add(1, "wsp", 6)
	r.ready = ReasonDead
	if ok, why := r.k.Gate(context.Background()); ok || why != ReasonDead {
		t.Fatalf("gate %v %q", ok, why)
	}
	r.enq(id) // an enqueue before death stays queued
	r.run()
	if r.state(id).State != vault.KeepQ {
		t.Fatal("created for a dead hero")
	}
	// the fork also refuses: ready says fine but the game says dead
	r.ready = ""
	r.api.Dead = true
	r.run()
	rec := r.state(id)
	if rec.State != vault.KeepQ || rec.InFlight || Reason(rec.Reason) != ReasonDead {
		t.Fatalf("%+v", rec)
	}
}

func TestVerbsMissing(t *testing.T) {
	r := newRig(t, "sorceress")
	id := r.add(1, "wsp", 6)
	r.enq(id)
	r.api.FlagOff = true
	r.run()
	if r.state(id).State != vault.KeepQ || r.k.Info().Waiting != ReasonVerbs {
		t.Fatalf("%+v", r.k.Info())
	}
	if ok, why := r.k.Gate(context.Background()); ok || why != ReasonVerbs {
		t.Fatalf("gate %v %q", ok, why)
	}
}

func TestBackupFirstAndFailClosed(t *testing.T) {
	r := newRig(t, "paladin")
	a, b := r.add(1, "wsp", 6), r.add(2, "wsp", 6)
	r.enq(a)
	backups := 0
	r.k.Backup = func(context.Context) error { backups++; return errors.New("no space on device") }
	r.run()
	if r.state(a).State != vault.KeepQ || r.state(a).InFlight || backups != 1 {
		t.Fatal("a failed backup must stop before any create")
	}
	if st, _ := r.api.VaultStats(context.Background()); st.Creates != 0 {
		t.Fatal("create without backup")
	}
	r.k.Backup = func(context.Context) error { backups++; return nil }
	r.run()
	r.enq(b)
	r.run()
	if backups != 2 { // one failure, one success, and none for the second batch
		t.Fatalf("backups %d", backups)
	}
}

func TestReadbackMismatchTripsAndUndoes(t *testing.T) {
	r := newRig(t, "sorceress")
	a, b := r.add(1, "wsp", 6), r.add(2, "wsp", 6)
	r.api.readBack = func(n string) string { return n[:len(n)-2] + "ff" }
	r.enq(a, b)
	r.run()
	bad := r.state(a)
	if bad.State != vault.Recorded || !strings.Contains(Reason(bad.Reason), ReasonReadback) {
		t.Fatalf("%+v", bad)
	}
	if len(r.api.removedID) != 1 || len(r.api.Carried) != 0 {
		t.Fatalf("copy not removed: %v", r.api.removedID)
	}
	if r.state(b).State != vault.KeepQ || r.k.Info().Tripped == "" {
		t.Fatal("the rest must wait: creating is off for the process")
	}
	if ok, _ := r.k.Gate(context.Background()); ok {
		t.Fatal("gate open after a mismatch")
	}
	// a record must never vanish: the vault still has it
	if _, ok := r.st.Get(a); !ok {
		t.Fatal("record lost")
	}
}

func TestPackTargetAndBadTarget(t *testing.T) {
	r := newRig(t, "paladin")
	id := r.add(1, "wsp", 6)
	if res := r.k.Enqueue(context.Background(), []string{id}, "moon", false); res[0].OK || res[0].Code != "target" {
		t.Fatalf("%+v", res)
	}
	if res := r.k.Enqueue(context.Background(), []string{"v-none"}, "", false); res[0].OK || res[0].Code != "missing" {
		t.Fatalf("%+v", res)
	}
	r.k.Enqueue(context.Background(), []string{id}, vault.WhereInv, true)
	if !r.k.TripWanted() {
		t.Fatal("Keep now should ask for a trip")
	}
	r.run()
	if r.api.Carried[0].Where != d2dctl.FakePack || !strings.HasPrefix(r.state(id).Reason, "backpack") {
		t.Fatalf("%+v %q", r.api.Carried, r.state(id).Reason)
	}
	if r.k.TripWanted() {
		t.Fatal("trip flag should clear")
	}
}

// The kill switch or the safe moment changing in the middle of a batch stops the next create.
func TestSwitchOffOrLeaveTownMidBatch(t *testing.T) {
	for _, class := range classes {
		for _, mode := range []string{"off", "ready"} {
			r := newRig(t, class)
			for i := 1; i <= 4; i++ {
				r.enq(r.add(i, "wsp", 6))
			}
			n := 0
			r.k.Backup = func(context.Context) error { return nil }
			inner := r.k.Ready
			r.k.Ready = func() string {
				n++
				if n > 2 { // the check at the start and before the first create pass, then the world changes
					if mode == "ready" {
						return "not in town"
					}
					r.on = false
				}
				return inner()
			}
			r.run()
			st, _ := r.api.VaultStats(context.Background())
			if st.Creates >= 4 || r.k.Pending() == 0 {
				t.Fatalf("%s/%s: creates %d pending %d, the batch did not stop", class, mode, st.Creates, r.k.Pending())
			}
		}
	}
}
