package d2dworld

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2map"
)

func (s *stub) LevelMap(context.Context) (*d2dctl.LevelMap, error) {
	s.n("levelmap")
	if s.lm == nil {
		return nil, errors.New("err unknown debug verb")
	}
	return s.lm, nil
}

// farTree is a level-1 layout with the Tree of Inifuss (class 30) far from the hero and its preset room 161.
func farTree() *d2dctl.LevelMap {
	return &d2dctl.LevelMap{Level: 1, Unit: "cell",
		Rooms:   []d2dctl.LayoutRoom{{X: 2, Y: 2, W: 8, H: 8, Preset: true, Def: 161}, {X: 20, Y: 20, W: 8, H: 8, Def: -1}},
		Objects: []d2dctl.LayoutObject{{Index: 1, Class: 12, Name: "Wooden Door", X: 1.0, Y: 0.2}, {Index: 7, Class: 30, Name: "Tree of Inifuss", X: 90.5, Y: 70.5, Kind: "tree"}}}
}

func presetOf(lm *d2map.LevelMap, class int) (d2map.Point, bool) {
	for _, p := range lm.Presets {
		if p.ID == class {
			return p.Pos, true
		}
	}
	return d2map.Point{}, false
}

func TestLayoutReadAddsFarObjectAndRoom(t *testing.T) {
	s := newStub()
	s.lm = farTree()
	w := New(s, Options{LayoutRead: func() bool { return true }})
	if _, err := w.Update(); err != nil {
		t.Fatal(err)
	}
	lm := w.Current()
	pos, ok := presetOf(lm, 30)
	if !ok || pos != (d2map.Point{X: 452, Y: 352}) { // OriginNone: subtile = floor(5 * cell)
		t.Fatalf("tree preset %v %v (presets %+v)", pos, ok, lm.Presets)
	}
	if n := len(lm.Presets); n != 3 { // waypoint, door (listed by presets, not doubled by the layout) and the tree
		t.Fatalf("presets %d: %+v", n, lm.Presets)
	}
	found := false
	for _, r := range lm.RoomRows {
		found = found || r.Type == 161
	}
	if !found {
		t.Fatalf("room 161 missing: %+v", lm.RoomRows)
	}
}

func TestLayoutReadOffOrAbsentChangesNothing(t *testing.T) {
	cases := map[string]func() *World{
		"knob off": func() *World {
			s := newStub()
			s.lm = farTree()
			return New(s, Options{LayoutRead: func() bool { return false }})
		},
		"knob unset": func() *World { s := newStub(); s.lm = farTree(); return New(s, Options{}) },
		"verb absent (stock d2d)": func() *World {
			return New(struct{ Verbs }{newStub()}, Options{LayoutRead: func() bool { return true }})
		},
		"verb errors": func() *World { return New(newStub(), Options{LayoutRead: func() bool { return true }}) },
	}
	for name, mk := range cases {
		w := mk()
		if _, err := w.Update(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		lm := w.Current()
		if _, ok := presetOf(lm, 30); ok || len(lm.Presets) != 2 || len(lm.RoomRows) != 0 {
			t.Fatalf("%s: presets %+v rooms %+v", name, lm.Presets, lm.RoomRows)
		}
	}
}

func TestLayoutReadNotSentWhenOff(t *testing.T) {
	s := newStub()
	s.lm = farTree()
	w := New(s, Options{})
	if _, err := w.Update(); err != nil {
		t.Fatal(err)
	}
	if s.calls["levelmap"] != 0 {
		t.Fatalf("verb sent with the knob off: %d", s.calls["levelmap"])
	}
}

func (s *stub) LevelMonsters(context.Context) (*d2dctl.LevelMonsters, error) {
	s.n("levelmonsters")
	if s.mons == nil {
		return nil, errors.New("err unknown debug verb")
	}
	return s.mons, nil
}

// countessRead is a level-1 read with a superunique spawn point far from the hero in a room not up yet.
func countessRead() *d2dctl.LevelMonsters {
	return &d2dctl.LevelMonsters{Level: 1, Unit: "cell", RoomsUp: []int{0}, Presets: []d2dctl.MonsterPreset{
		{Index: 1, Kind: "monster", ID: 12, Super: -1, Name: "zombie1", X: 3.5, Y: 3.5, Room: 0, Up: true},
		{Index: 2, Kind: "superunique", ID: 740, Super: 6, Name: "The Countess", X: 90.5, Y: 70.5, Room: 4}}}
}

func TestLayoutReadAddsSuperuniqueSpawnPresetWithName(t *testing.T) {
	s := newStub()
	s.mons = countessRead()
	w := New(s, Options{LayoutRead: func() bool { return true }})
	if _, err := w.Update(); err != nil {
		t.Fatal(err)
	}
	lm := w.Current()
	var got *d2map.Preset
	for i := range lm.Presets {
		if lm.Presets[i].Kind == d2map.PresetNPC {
			got = &lm.Presets[i]
		}
	}
	if got == nil || got.ID != 740 || got.Name != "The Countess" || got.Pos != (d2map.Point{X: 452, Y: 352}) {
		t.Fatalf("presets %+v", lm.Presets)
	}
	if n := len(lm.Presets); n != 3 { // waypoint, door and her spawn: the plain monster preset is not added
		t.Fatalf("presets %d: %+v", n, lm.Presets)
	}
	if _, err := w.Update(); err != nil || len(w.Current().Presets) != 3 {
		t.Fatalf("a second fetch doubled the preset: %v %+v", err, w.Current().Presets)
	}
}

func TestLayoutReadMonstersOffOrAbsentChangesNothing(t *testing.T) {
	cases := map[string]func() (*World, *stub){
		"knob off": func() (*World, *stub) { s := newStub(); s.mons = countessRead(); return New(s, Options{}), s },
		"verb errors": func() (*World, *stub) {
			s := newStub()
			return New(s, Options{LayoutRead: func() bool { return true }}), s
		},
		"other level": func() (*World, *stub) {
			s := newStub()
			s.mons = countessRead()
			s.mons.Level = 25
			return New(s, Options{LayoutRead: func() bool { return true }}), s
		},
	}
	for name, mk := range cases {
		w, s := mk()
		if _, err := w.Update(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if n := len(w.Current().Presets); n != 2 {
			t.Fatalf("%s: presets %+v", name, w.Current().Presets)
		}
		if name == "knob off" && s.calls["levelmonsters"] != 0 {
			t.Fatalf("verb sent with the knob off")
		}
	}
}
