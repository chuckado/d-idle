package bot

import (
	"context"
	"testing"

	"d2lod-bot/internal/task"
)

// dropLoot shows the item only after after polls (the tree animation), then reports it picked.
type dropLoot struct {
	fakeLoot
	w     *world
	after int
	polls int
}

func (l *dropLoot) PickUp(ctx context.Context, o PickOpts) (PickReport, error) {
	l.w.call("pickup")
	l.polls++
	if l.polls > l.after {
		return PickReport{Picked: 1}, nil
	}
	return PickReport{}, nil
}

func treeRig(t *testing.T, after int) (taskActs, *dropLoot) {
	g := newRig(t)
	l := &dropLoot{fakeLoot: fakeLoot{w: g.w}, w: g.w, after: after}
	g.r.d.Task = &TaskDeps{Objects: &fakeObjects{}}
	g.r.d.Loot = l
	return taskActs{r: g.r}, l
}

// RC4 of the 2026-11-17 diagnosis: the scroll appears a moment after the click; the pickup waits and retries, bounded.
func TestTreeWaitsForTheScrollToDrop(t *testing.T) {
	a, l := treeRig(t, 2)
	if err := a.Interact(context.Background(), TreeObject); err != nil {
		t.Fatal(err)
	}
	if l.polls != 3 {
		t.Fatalf("polls %d, want 3 (two empty, one pick)", l.polls)
	}
}

func TestTreeGivesUpWithAClearError(t *testing.T) {
	a, l := treeRig(t, 1000)
	err := a.Interact(context.Background(), TreeObject)
	if task.Classify(err) != task.KindNotFound {
		t.Fatalf("got %v", err)
	}
	if l.polls != treeDropPolls {
		t.Fatalf("polls %d, want the bound %d", l.polls, treeDropPolls)
	}
}
