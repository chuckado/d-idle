package bulkid

import (
	"context"
	"errors"
	"slices"
	"testing"
)

// fakeMover holds two places with a cell capacity each. Failure injection: lie makes Move report success without
// moving; dieAfter makes every call after N game calls fail (a process that dies).
type fakeMover struct {
	pack, stash       []Item
	packCap, stashCap int
	lie               map[uint32]bool
	calls, dieAfter   int
}

var errDied = errors.New("died")

func (f *fakeMover) tick() error {
	f.calls++
	if f.dieAfter > 0 && f.calls > f.dieAfter {
		return errDied
	}
	return nil
}
func (f *fakeMover) Pack(context.Context) ([]Item, error) {
	return slices.Clone(f.pack), f.tick()
}
func (f *fakeMover) Stash(context.Context) ([]Item, error) {
	return slices.Clone(f.stash), f.tick()
}
func (f *fakeMover) Move(_ context.Context, id uint32, to string) error {
	if err := f.tick(); err != nil {
		return err
	}
	from, dst, capDst := &f.stash, &f.pack, f.packCap
	if to == PlaceStash {
		from, dst, capDst = &f.pack, &f.stash, f.stashCap
	}
	i := slices.IndexFunc(*from, func(it Item) bool { return it.ID == id })
	if i < 0 {
		return errors.New("no such item")
	}
	if f.lie[id] {
		return nil
	}
	if len(*dst) >= capDst {
		return ErrNoRoom
	}
	it := (*from)[i]
	*from = slices.Delete(*from, i, i+1)
	it.Place = to
	*dst = append(*dst, it)
	return nil
}

func world(nPack, nStash, packCap, stashCap int) *fakeMover {
	f := &fakeMover{packCap: packCap, stashCap: stashCap}
	for i := 1; i <= nPack; i++ {
		f.pack = append(f.pack, Item{ID: uint32(100 + i), Place: PlacePack})
	}
	for i := 1; i <= nStash; i++ {
		f.stash = append(f.stash, Item{ID: uint32(i), Place: PlaceStash})
	}
	return f
}

func ids(n int) []uint32 {
	var out []uint32
	for i := 1; i <= n; i++ {
		out = append(out, uint32(i))
	}
	return out
}

// total counts every item the hero owns: none may ever disappear.
func total(f *fakeMover) int { return len(f.pack) + len(f.stash) }

func TestStageInChunksAndVerifies(t *testing.T) {
	f := world(2, 25, 40, 100)
	var lines []string
	s := &Stager{M: f, J: &Journal{}, Chunk: 10, Logf: func(format string, a ...any) { lines = append(lines, format) }}
	res, err := s.StageIn(context.Background(), ids(25))
	if err != nil || len(res.Moved) != 25 || len(res.Left) != 0 {
		t.Fatalf("%+v %v", res, err)
	}
	if len(lines) != 3 { // 10 + 10 + 5
		t.Fatalf("batches logged: %d", len(lines))
	}
	rr, err := s.Return(context.Background())
	if err != nil || len(rr.Back) != 25 || len(s.J.IDs) != 0 || len(f.stash) != 25 || len(f.pack) != 2 {
		t.Fatalf("%+v %v journal %v", rr, err, s.J.IDs)
	}
}

// Must-fail: a full pack must not lose or overfill anything.
func TestFullPackNeverLosesItems(t *testing.T) {
	f := world(0, 12, 5, 100)
	s := &Stager{M: f, J: &Journal{}, Chunk: 4}
	res, err := s.StageIn(context.Background(), ids(12))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Moved) != 5 || len(res.Left) != 7 || res.Reason == "" {
		t.Fatalf("%+v", res)
	}
	if len(f.pack) > f.packCap || total(f) != 12 {
		t.Fatalf("pack %d total %d", len(f.pack), total(f))
	}
	if _, err := s.Return(context.Background()); err != nil || total(f) != 12 || len(f.stash) != 12 {
		t.Fatalf("after return: pack %d stash %d", len(f.pack), len(f.stash))
	}
}

// Must-fail: a mover that reports success but did not move is caught by the read-back; nothing is believed.
func TestLyingMoveIsNotBelieved(t *testing.T) {
	f := world(0, 6, 40, 100)
	f.lie = map[uint32]bool{3: true}
	s := &Stager{M: f, J: &Journal{}, Chunk: 6}
	res, _ := s.StageIn(context.Background(), ids(6))
	if slices.Contains(res.Moved, 3) || !slices.Contains(res.Left, 3) || res.Reason == "" {
		t.Fatalf("%+v", res)
	}
	if total(f) != 6 {
		t.Fatal("items lost")
	}
	if _, err := s.Return(context.Background()); err != nil || len(s.J.IDs) != 0 || len(f.stash) != 6 {
		t.Fatalf("journal %v stash %d", s.J.IDs, len(f.stash))
	}
}

func TestStashFullOnReturnKeepsItemsInThePack(t *testing.T) {
	f := world(0, 6, 40, 6)
	s := &Stager{M: f, J: &Journal{}, Chunk: 6}
	if _, err := s.StageIn(context.Background(), ids(6)); err != nil {
		t.Fatal(err)
	}
	for i := uint32(70); i < 76; i++ { // something filled the freed cells meanwhile
		f.stash = append(f.stash, Item{ID: i, Place: PlaceStash})
	}
	rr, _ := s.Return(context.Background())
	if len(rr.Kept) != 6 || rr.Reason == "" || len(s.J.IDs) != 6 || total(f) != 12 {
		t.Fatalf("%+v journal %v total %d", rr, s.J.IDs, total(f))
	}
}

// Restart at every call count: a new Stager with the persisted journal and a healthy hero must end with all 9 items
// in the stash or the pack, none gone, and Return must finish the job.
func TestRestartAtEveryCutPoint(t *testing.T) {
	for cut := 1; cut < 40; cut++ {
		f := world(1, 9, 40, 100)
		var saved []uint32
		j := &Journal{Save: func(v []uint32) error { saved = v; return nil }}
		f.dieAfter = cut
		s := &Stager{M: f, J: j, Chunk: 3}
		_, _ = s.StageIn(context.Background(), ids(9))
		f.dieAfter = 0 // the process restarts
		s2 := &Stager{M: f, J: &Journal{IDs: slices.Clone(saved)}, Chunk: 3}
		rr, err := s2.Return(context.Background())
		if err != nil || len(rr.Gone) != 0 || len(rr.Kept) != 0 {
			t.Fatalf("cut %d: %+v %v", cut, rr, err)
		}
		if total(f) != 10 || len(f.stash) != 9 || len(f.pack) != 1 || len(s2.J.IDs) != 0 {
			t.Fatalf("cut %d: pack %d stash %d journal %v", cut, len(f.pack), len(f.stash), s2.J.IDs)
		}
	}
}
