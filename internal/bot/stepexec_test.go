package bot

import (
	"context"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/task"
)

const scrollClass = 900

// scrollCatalog knows the Scroll of Inifuss (base code bks) on top of the shared fake.
type scrollCatalog struct{ fakeCatalog }

func (c scrollCatalog) Item(class int) (ItemMeta, bool) {
	if class == scrollClass {
		return ItemMeta{Name: "Scroll of Inifuss", Code: "bks", W: 1, H: 2}, true
	}
	return c.fakeCatalog.Item(class)
}

// packLoot puts the scroll in the pack on its n-th call (the drop lands a moment after the click); never when n < 0.
type packLoot struct {
	fakeLoot
	n, calls int
}

func (l *packLoot) PickUp(ctx context.Context, o PickOpts) (PickReport, error) {
	l.w.call("pickup")
	l.calls++
	if l.n >= 0 && l.calls == l.n {
		l.w.set(func(w *world) {
			w.items = append(w.items, d2.Unit{ID: 77, Type: d2.TypeItem, Class: scrollClass, Item: &d2.Item{Where: d2.ContInventory}})
		})
		return PickReport{Picked: 1}, nil
	}
	return PickReport{}, nil
}

func execRig(t *testing.T, class d2.Class, dropAt int) (taskActs, *packLoot, *fakeObjects) {
	g := newRig(t)
	g.w.class = class
	g.r.d.Catalog = scrollCatalog{}
	g.r.e.Tune.Quests.StepExec = 1
	l := &packLoot{fakeLoot: fakeLoot{w: g.w}, n: dropAt}
	objs := &fakeObjects{}
	g.r.d.Task = &TaskDeps{Objects: objs}
	g.r.d.Loot = l
	return taskActs{r: g.r}, l, objs
}

var execClasses = []d2.Class{d2.Sorceress, d2.Paladin}

// The executor clicks the tree once, picks up until the scroll is in the PACK (not until the looter says so), and the
// pseudo item of the click is set.
func TestStepExecTakesTheScrollOnTwoClasses(t *testing.T) {
	for _, cl := range execClasses {
		a, l, objs := execRig(t, cl, 3)
		if err := a.Interact(context.Background(), TreeObject); err != nil {
			t.Fatalf("class %v: %v", cl, err)
		}
		if len(objs.clicked) != 1 || objs.clicked[0] != TreeObject || l.calls != 3 {
			t.Fatalf("class %v: clicked %v pickups %d", cl, objs.clicked, l.calls)
		}
	}
}

// Must fail: a looter that says "picked" while the pack never gets the scroll is not Done; the post is the pack.
func TestStepExecWaitsForThePackNotForTheLooter(t *testing.T) {
	for _, cl := range execClasses {
		a, _, objs := execRig(t, cl, -1)
		a.r.d.Loot = liar{fakeLoot{w: a.r.d.Loot.(*packLoot).w}}
		err := a.Interact(context.Background(), TreeObject)
		if task.Classify(err) != task.KindNotFound || err == nil || !strings.Contains(err.Error(), "not observed") {
			t.Fatalf("class %v: got %v", cl, err)
		}
		if len(objs.clicked) != 3 { // tries 3 of the data row
			t.Fatalf("class %v: clicks %d, want the row's 3 tries", cl, len(objs.clicked))
		}
	}
}

type liar struct{ fakeLoot }

func (l liar) PickUp(context.Context, PickOpts) (PickReport, error) { return PickReport{Picked: 1}, nil }

// With the knob at 0 the old handler runs (its own pick-up loop of treeDropPolls looks), the executor is not used.
func TestStepExecOffKeepsTheOldHandler(t *testing.T) {
	a, l, _ := execRig(t, d2.Sorceress, -1)
	a.r.e.Tune.Quests.StepExec = 0
	if err := a.Interact(context.Background(), TreeObject); task.Classify(err) != task.KindNotFound {
		t.Fatalf("got %v", err)
	}
	if l.calls != treeDropPolls {
		t.Fatalf("polls %d, want the old bound %d", l.calls, treeDropPolls)
	}
}
