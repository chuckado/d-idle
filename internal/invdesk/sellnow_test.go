package invdesk

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

type fakeSeller struct {
	calls int
	msg   string
	err   error
	left  func() // runs during the sale (clear the marks, ...)
}

func (f *fakeSeller) SellMarked(ctx context.Context) (string, error) {
	f.calls++
	if f.left != nil {
		f.left()
	}
	return f.msg, f.err
}

var bothClasses = []d2.Class{d2.Paladin, d2.Sorceress}

func sellRig(t *testing.T, class d2.Class, area int) (*Desk, *fakeSeller, *d2.Snapshot) {
	d, _, snap := rig(t, class, area)
	fs := &fakeSeller{msg: "sold 1 of 1 marked"}
	d.SetSeller(fs)
	return d, fs, snap
}

func TestSellNowNothingMarkedIsRefused(t *testing.T) {
	for _, class := range bothClasses {
		d, _, _ := sellRig(t, class, town1)
		_, err := d.RequestSell()
		if code(err) != "nothing" || err.Error() != NothingMarked {
			t.Fatalf("%v: %v", class, err)
		}
		if len(d.View().Queue) != 0 {
			t.Errorf("%v: queued anyway", class)
		}
		bare, _, _ := rig(t, class, town1) // no seller installed
		if _, err := bare.RequestSell(); code(err) != "off" {
			t.Errorf("no backend: %v", err)
		}
	}
}

func TestSellNowInTownSellsAtOnce(t *testing.T) {
	for _, class := range bothClasses {
		d, fs, _ := sellRig(t, class, town1)
		if err := d.SetMark(1, MarkSell); err != nil {
			t.Fatal(err)
		}
		if _, err := d.RequestSell(); err != nil {
			t.Fatal(err)
		}
		if d.Pending() != 1 {
			t.Fatalf("pending %d", d.Pending())
		}
		out := d.RunPending(context.Background())
		v := d.View()
		if fs.calls != 1 || len(out) != 1 || len(v.Queue) != 0 || len(v.Recent) != 1 || v.Recent[0].State != StateDone || v.Recent[0].Result != fs.msg {
			t.Fatalf("%v: calls %d out %v queue %+v recent %+v", class, fs.calls, out, v.Queue, v.Recent)
		}
		if d.View().Equipment["rhand"] != nil || len(d.View().Pins) != 0 {
			t.Errorf("%v: a sale pinned or equipped something", class)
		}
	}
}

func TestSellNowFromTheFieldIsAnErrandThatWaitsForTown(t *testing.T) {
	for _, class := range bothClasses {
		d, fs, snap := sellRig(t, class, field)
		if err := d.SetMark(1, MarkSell); err != nil {
			t.Fatal(err)
		}
		if _, err := d.RequestSell(); err != nil {
			t.Fatal(err)
		}
		if d.Pending() != 0 || d.ErrandsWaiting() != 1 || d.ErrandKind() != KindSell {
			t.Fatalf("field: pending %d errands %d kind %q", d.Pending(), d.ErrandsWaiting(), d.ErrandKind())
		}
		if got := d.RunPending(context.Background()); len(got) != 0 || fs.calls != 0 {
			t.Fatalf("ran in the field: %v", got)
		}
		if q := d.View().Queue; q[0].Reason != WaitsForTown {
			t.Errorf("queue card %+v", q)
		}
		snap.Player.Area = town1 // the errand trip brought the hero home
		if got := d.RunPending(context.Background()); len(got) != 1 || fs.calls != 1 {
			t.Fatalf("in town: %v calls %d", got, fs.calls)
		}
	}
}

func TestSellNowOnlySellsMarkedAndNeverSellItems(t *testing.T) {
	d, _, _ := sellRig(t, d2.Paladin, town1)
	// item 4 is a quest item: SetMark refuses it, and an unmarked pack item does not count
	if err := d.SetMark(4, MarkSell); code(err) != "never" {
		t.Fatal(err)
	}
	if err := d.SetMark(2, MarkKeep); err != nil {
		t.Fatal(err)
	}
	if _, err := d.RequestSell(); code(err) != "nothing" {
		t.Errorf("keep or never marks counted: %v", err)
	}
}

func TestSellNowFailsWhenTheMarksWereClearedMeanwhile(t *testing.T) {
	d, fs, _ := sellRig(t, d2.Sorceress, town1)
	_ = d.SetMark(1, MarkSell)
	if _, err := d.RequestSell(); err != nil {
		t.Fatal(err)
	}
	_ = d.SetMark(1, "")
	d.RunPending(context.Background())
	r := d.View().Recent
	if fs.calls != 0 || len(r) != 1 || r[0].State != StateFailed || r[0].Reason != NothingMarked {
		t.Fatalf("calls %d recent %+v", fs.calls, r)
	}
}

func TestSellNowFailureAndPostponement(t *testing.T) {
	d, fs, _ := sellRig(t, d2.Paladin, town1)
	_ = d.SetMark(1, MarkSell)
	_, _ = d.RequestSell()
	fs.err = errors.New("leveler: a fight is on") // isBusy: put back to queued, not failed
	d.RunPending(context.Background())
	if q := d.View().Queue; len(q) != 1 || q[0].State != StateQueued {
		t.Fatalf("not postponed: %+v", d.View())
	}
	fs.err = errors.New("no vendor would buy them")
	d.RunPending(context.Background())
	if r := d.View().Recent; len(r) != 1 || r[0].State != StateFailed || r[0].Reason != fs.err.Error() {
		t.Fatalf("recent %+v", r)
	}
}

func TestSellNowQueueLimitCancelAndDuplicate(t *testing.T) {
	d, _, _ := sellRig(t, d2.Paladin, field)
	_ = d.SetMark(1, MarkSell)
	if _, err := d.RequestSell(); err != nil {
		t.Fatal(err)
	}
	if _, err := d.RequestSell(); code(err) != "queued" {
		t.Errorf("duplicate: %v", err)
	}
	if err := d.Cancel(0, KindSell); err != nil {
		t.Fatal(err)
	}
	if v := d.View(); len(v.Queue) != 0 || v.Recent[0].State != StateCancelled {
		t.Errorf("cancel: %+v", v)
	}
	// the queue limit (8) still holds: fill it with buy-scroll errands' siblings, then sell is refused with "full"
	for i := 0; i < MaxQueue; i++ {
		d.mu.Lock()
		_, err := d.enqueueLocked(Request{ID: uint32(100 + i), Kind: KindTrash, Name: "x"})
		d.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.RequestSell(); code(err) != "full" {
		t.Errorf("limit: %v", err)
	}
	// a doing request cannot be cancelled
	d.mu.Lock()
	d.st.Queue[0].State = StateDoing
	d.mu.Unlock()
	if err := d.Cancel(100, KindTrash); code(err) != "busy" {
		t.Errorf("cancel doing: %v", err)
	}
}

func TestSellNowSurvivesARestartAsQueued(t *testing.T) {
	d, _, _ := sellRig(t, d2.Sorceress, field)
	_ = d.SetMark(1, MarkSell)
	_, _ = d.RequestSell()
	d2x, err := New(Config{Path: d.c.Path, Snap: d.c.Snap})
	if err != nil {
		t.Fatal(err)
	}
	if q := d2x.View().Queue; len(q) != 1 || q[0].Kind != KindSell || q[0].State != StateQueued {
		t.Fatalf("%+v", q)
	}
}

// Schema 10 golden: a sell-now request waiting for town and a done one, both classes.
func TestViewSchema10SellGolden(t *testing.T) {
	for _, class := range bothClasses {
		d, fs, snap := sellRig(t, class, town1)
		d.c.Now = func() time.Time { return time.Unix(1_800_000_000, 0) }
		_ = d.SetMark(1, MarkSell)
		_, _ = d.RequestSell()
		d.RunPending(context.Background())
		fs.msg = "sold 1 of 1 marked"
		snap.Player.Area = field
		if _, err := d.RequestSell(); err != nil {
			t.Fatal(err)
		}
		v := d.View()
		if v.Schema != 10 {
			t.Fatalf("schema %d", v.Schema)
		}
		b, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		b = append(b, '\n')
		path := filepath.Join("testdata", "view_schema10_sell_"+class.String()+".json")
		if os.Getenv("UPDATE") != "" {
			if err := os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(want) != string(b) {
			t.Errorf("%v: view differs from %s\n got: %s", class, path, b)
		}
	}
}
