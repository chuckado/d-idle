package bot

import (
	"log/slog"
	"strings"
	"sync"
	"testing"

	"d2lod-bot/internal/d2"
)

type tomeCatalog struct{ fakeCatalog }

func (c tomeCatalog) Item(class int) (ItemMeta, bool) {
	if class == 6 {
		return ItemMeta{Name: "Tome of Town Portal", GoalType: "book", W: 1, H: 2}, true
	}
	return c.fakeCatalog.Item(class)
}

type statsReader struct {
	d2.StateReader
	st map[uint32]d2.Stats
}

func (r statsReader) UnitStats(a uint32) (d2.Stats, error) {
	if s, ok := r.st[a]; ok {
		return s, nil
	}
	return nil, d2.ErrUnsupported
}

func hasTP(t *testing.T, charges int, scrolls int) bool {
	g := newRig(t)
	g.r.d.Catalog = tomeCatalog{}
	st := map[uint32]d2.Stats{}
	if charges > 0 {
		st[100] = d2.Stats{70: int32(charges)}
	}
	g.r.e.Read = statsReader{st: st}
	g.w.set(func(w *world) {
		w.items = append(w.items, d2.Unit{ID: 900, Addr: 100, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}})
		for i := 0; i < scrolls; i++ {
			w.items = append(w.items, d2.Unit{ID: uint32(901 + i), Addr: uint32(200 + i), Type: d2.TypeItem, Class: 5, Item: &d2.Item{Where: d2.ContInventory}})
		}
	})
	s, _ := g.w.snapshot()
	return g.r.resInput(s).HasTP
}

// F5: an empty Tome of Town Portal is not a portal; a tome with charges or a scroll is.
func TestHasTPCountsCharges(t *testing.T) {
	if hasTP(t, 0, 0) {
		t.Error("empty tome counted as a portal")
	}
	if !hasTP(t, 0, 2) {
		t.Error("empty tome plus scrolls: scrolls are a portal source")
	}
	if !hasTP(t, 3, 0) {
		t.Error("tome with 3 charges is a portal source")
	}
}

type logBuf struct {
	mu sync.Mutex
	b  strings.Builder
}

func (l *logBuf) Write(p []byte) (int, error) { l.mu.Lock(); defer l.mu.Unlock(); return l.b.Write(p) }
func (l *logBuf) dumps() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Count(l.b.String(), "inv-tp-dump")
}

// The dump fires on the first look, on an area change and on a verdict change, and not otherwise.
func TestInvTPDump(t *testing.T) {
	g := newRig(t)
	g.r.d.Catalog = tomeCatalog{}
	st := map[uint32]d2.Stats{100: {70: 3}}
	g.r.e.Read = statsReader{st: st}
	buf := &logBuf{}
	g.r.e.Log = slog.New(slog.NewTextHandler(buf, nil))
	g.w.set(func(w *world) {
		w.items = append(w.items, d2.Unit{ID: 900, Addr: 100, Type: d2.TypeItem, Class: 6, Item: &d2.Item{Where: d2.ContInventory}})
	})
	look := func() {
		s, _ := g.w.snapshot()
		g.r.resInput(s)
	}
	step := func(name string, want int, f func()) {
		t.Helper()
		if f != nil {
			f()
		}
		look()
		if got := buf.dumps(); got != want {
			t.Fatalf("%s: %d dump lines, want %d", name, got, want)
		}
	}
	step("first", 1, nil)
	step("same area, same verdict", 1, nil)
	step("again", 1, nil)
	step("area change", 2, func() { g.w.set(func(w *world) { w.area = 2 }) })
	step("verdict change", 3, func() { delete(st, 100) })
	step("steady", 3, nil)
	if out := buf.b.String(); !strings.Contains(out, "raw70=err:") || !strings.Contains(out, "hastp=false") {
		t.Errorf("dump lacks the raw stat read or the verdict:\n%s", out)
	}
}
