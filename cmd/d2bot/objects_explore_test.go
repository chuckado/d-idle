package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/nav"
	"d2lod-bot/internal/task"
)

type exploreClock struct{ now time.Time }

func (c *exploreClock) Now() time.Time { return c.now }

// fakeSearch: the object shows after appearsAt walks; each walk takes the whole budget. levelEnds makes the walk
// report ErrExplored from that walk on.
type fakeSearch struct {
	clk       *exploreClock
	walks     int
	appearsAt int
	levelEnds int
}

func (f *fakeSearch) explore(context.Context) error {
	f.walks++
	f.clk.now = f.clk.now.Add(objectExploreBudget)
	if f.levelEnds > 0 && f.walks >= f.levelEnds {
		return nav.ErrExplored
	}
	return context.DeadlineExceeded
}

func (f *fakeSearch) look() (*d2.Unit, error) {
	if f.appearsAt > 0 && f.walks >= f.appearsAt {
		return &d2.Unit{Class: 30}, nil
	}
	return nil, errors.New("object class 30 not found")
}

func slice(f *fakeSearch, tally *exploreTally) (*d2.Unit, error) {
	return exploreForObject(context.Background(), f.clk, tally, exploreKey{2, 30}, f.explore, f.look)
}

func TestExploreForObjectKeepsGoingUntilTheObjectShows(t *testing.T) {
	f := &fakeSearch{clk: &exploreClock{}, appearsAt: 5}
	tally := &exploreTally{}
	for i := 1; i <= 4; i++ {
		if _, err := slice(f, tally); task.Classify(err) != task.KindExploring {
			t.Fatalf("slice %d: want exploring, got %v", i, err)
		}
	}
	u, err := slice(f, tally)
	if err != nil || u == nil {
		t.Fatalf("slice 5: %v", err)
	}
	if len(tally.spent) != 0 {
		t.Fatal("a find must clear the tally")
	}
}

func TestExploreForObjectEndsAtTheTotalBudget(t *testing.T) {
	f := &fakeSearch{clk: &exploreClock{}}
	tally := &exploreTally{}
	var last error
	slices := 0
	for ; slices < 20; slices++ {
		_, last = slice(f, tally)
		if task.Classify(last) != task.KindExploring {
			break
		}
	}
	want := int(objectExploreTotal / objectExploreBudget) // 7.2: the 8th walk crosses the total
	if slices != want || last == nil {
		t.Fatalf("ended after %d exploring slices (%v), want %d", slices, last, want)
	}
}

func TestExploreForObjectParksAtOnceWhenTheLevelIsWalkedOut(t *testing.T) {
	f := &fakeSearch{clk: &exploreClock{}, levelEnds: 1}
	_, err := slice(f, &exploreTally{})
	if err == nil || task.Classify(err) == task.KindExploring {
		t.Fatalf("got %v", err)
	}
}

func TestExploreForObjectWithoutTallyIsOneWalk(t *testing.T) {
	f := &fakeSearch{clk: &exploreClock{}}
	_, err := slice(f, nil)
	if err == nil || task.Classify(err) == task.KindExploring {
		t.Fatalf("got %v", err)
	}
}
