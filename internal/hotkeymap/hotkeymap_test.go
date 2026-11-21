package hotkeymap

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/agency"
)

func ch(name string) *agency.Choice { return &agency.Choice{Skill: name} }

func chs(names ...string) []agency.Choice {
	var out []agency.Choice
	for _, n := range names {
		out = append(out, agency.Choice{Skill: n})
	}
	return out
}

type class struct {
	name string
	plan agency.Plan
	ids  map[string]int
	want []string // skill per slot, in order
}

func classes() []class {
	return []class{
		{"Paladin", agency.Plan{Main: ch("Zeal"), Attack: ch("Zeal"), Secondary: ch("Fist of the Heavens"), Movement: chs("Charge"),
			Buffs: chs("Holy Shield"), Aura: ch("Might")},
			map[string]int{"Zeal": 101, "Fist of the Heavens": 121, "Charge": 107, "Holy Shield": 117, "Might": 98, "Book of Townportal": 220},
			[]string{"Zeal", "Fist of the Heavens", "Charge", "Holy Shield", "Might", "Book of Townportal"}},
		{"Sorceress", agency.Plan{Main: ch("Fire Ball"), Attack: ch("Fire Ball"), Secondary: ch("Blizzard"), Movement: chs("Teleport"),
			Buffs: chs("Frozen Armor", "Thunder Storm")},
			map[string]int{"Fire Ball": 47, "Blizzard": 59, "Teleport": 54, "Frozen Armor": 40, "Thunder Storm": 57, "Book of Townportal": 220},
			[]string{"Fire Ball", "Blizzard", "Teleport", "Frozen Armor", "Thunder Storm", "Book of Townportal"}},
	}
}

func resolver(ids map[string]int) func(string) (int, bool) {
	return func(n string) (int, bool) { id, ok := ids[n]; return id, ok }
}

type fakeBinder struct {
	table    [Slots]int
	writable int
	drop     map[int]bool // slots whose write is silently lost (the game ignored the key)
	binds    []int
}

func newFake(writable int) *fakeBinder {
	f := &fakeBinder{writable: writable, drop: map[int]bool{}}
	for i := range f.table {
		f.table[i] = Unbound
	}
	return f
}
func (f *fakeBinder) Writable() int { return f.writable }
func (f *fakeBinder) Bind(_ context.Context, slot, id int) error {
	f.binds = append(f.binds, slot)
	if !f.drop[slot] {
		f.table[slot] = id
	}
	return nil
}
func (f *fakeBinder) Read(context.Context) ([Slots]int, error) { return f.table, nil }

func TestGenerateOrderTwoClasses(t *testing.T) {
	for _, c := range classes() {
		m, err := Generate(c.plan, Options{Resolve: resolver(c.ids), TownPortal: "Book of Townportal"})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if len(m.Slots) != len(c.want) {
			t.Fatalf("%s: %d slots, want %d: %+v", c.name, len(m.Slots), len(c.want), m.Slots)
		}
		for i, s := range m.Slots {
			if s.Skill != c.want[i] || s.Index != i || s.ID != c.ids[c.want[i]] {
				t.Fatalf("%s: slot %d = %+v, want %s", c.name, i, s, c.want[i])
			}
		}
		if m.Slots[len(m.Slots)-1].Role != RoleTP {
			t.Fatalf("%s: town portal not last", c.name)
		}
	}
}

func TestGenerateRefusesUnknownSkillAndNoResolver(t *testing.T) {
	for _, c := range classes() {
		ids := map[string]int{}
		for k, v := range c.ids {
			ids[k] = v
		}
		delete(ids, c.want[1]) // the second element is not in the table
		_, err := Generate(c.plan, Options{Resolve: resolver(ids)})
		if err == nil || !strings.Contains(err.Error(), c.want[1]) {
			t.Fatalf("%s: err %v, want one naming %q", c.name, err, c.want[1])
		}
		if _, err := Generate(c.plan, Options{}); err == nil {
			t.Fatalf("%s: no resolver must be an error", c.name)
		}
	}
}

func TestGenerateSummonerKeepsBothAndDedups(t *testing.T) {
	pl := agency.Plan{Main: ch("Raise Skeleton"), Attack: ch("Bone Spear"), Summons: chs("Raise Skeleton", "Clay Golem")}
	ids := map[string]int{"Raise Skeleton": 70, "Bone Spear": 84, "Clay Golem": 75}
	m, err := Generate(pl, Options{Resolve: resolver(ids)})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, s := range m.Slots {
		got = append(got, s.Skill)
	}
	if strings.Join(got, ",") != "Raise Skeleton,Bone Spear,Clay Golem" {
		t.Fatalf("got %v", got)
	}
}

func TestGenerateCapsAtSixteen(t *testing.T) {
	ids := map[string]int{}
	var names []string
	for i := 0; i < 20; i++ {
		n := string(rune('A' + i))
		ids[n] = i + 1
		names = append(names, n)
	}
	m, err := Generate(agency.Plan{Main: ch("A"), Buffs: chs(names[1:]...)}, Options{Resolve: resolver(ids)})
	if err != nil || len(m.Slots) != Slots {
		t.Fatalf("slots %d err %v", len(m.Slots), err)
	}
}

func TestApplyWritesAndReadsBackTwoClasses(t *testing.T) {
	for _, c := range classes() {
		m, err := Generate(c.plan, Options{Resolve: resolver(c.ids), TownPortal: "Book of Townportal"})
		if err != nil {
			t.Fatal(err)
		}
		f := newFake(8)
		res, err := Apply(context.Background(), f, m)
		if err != nil || res.Written != len(m.Slots) || res.Beyond != 0 {
			t.Fatalf("%s: %+v %v", c.name, res, err)
		}
		for _, s := range m.Slots {
			if f.table[s.Index] != s.ID {
				t.Fatalf("%s: slot %d holds %d", c.name, s.Index, f.table[s.Index])
			}
		}
		again, err := Apply(context.Background(), f, m) // idempotent: nothing is written twice
		if err != nil || again.Written != 0 || again.Already != len(m.Slots) {
			t.Fatalf("%s: second apply %+v %v", c.name, again, err)
		}
	}
}

// Must fail: the game ignores a write (the key did nothing), the read-back catches it, for both classes.
func TestApplyMustFailWhenReadBackDiffers(t *testing.T) {
	for _, c := range classes() {
		m, _ := Generate(c.plan, Options{Resolve: resolver(c.ids)})
		f := newFake(8)
		f.drop[2] = true
		_, err := Apply(context.Background(), f, m)
		if !errors.Is(err, ErrReadBack) || !strings.Contains(err.Error(), "slot 2") {
			t.Fatalf("%s: err %v, want ErrReadBack naming slot 2", c.name, err)
		}
	}
}

// Must fail: a slot that reads back as a different skill than the one written (another bind landed on it).
func TestApplyMustFailOnWrongSkillBack(t *testing.T) {
	for _, c := range classes() {
		m, _ := Generate(c.plan, Options{Resolve: resolver(c.ids)})
		f := &wrongBack{fakeBinder: newFake(8)}
		if _, err := Apply(context.Background(), f, m); !errors.Is(err, ErrReadBack) {
			t.Fatalf("%s: err %v", c.name, err)
		}
	}
}

type wrongBack struct{ *fakeBinder }

func (w *wrongBack) Bind(_ context.Context, slot, id int) error { w.table[slot] = id + 1; return nil }

func TestApplyCountsSlotsBeyondTheRoute(t *testing.T) {
	c := classes()[0]
	m, _ := Generate(c.plan, Options{Resolve: resolver(c.ids), TownPortal: "Book of Townportal"})
	f := newFake(4)
	res, err := Apply(context.Background(), f, m)
	if err != nil || res.Written != 4 || res.Beyond != len(m.Slots)-4 {
		t.Fatalf("%+v %v", res, err)
	}
	for i := 4; i < Slots; i++ {
		if f.table[i] != Unbound {
			t.Fatalf("slot %d written beyond the route", i)
		}
	}
}

func TestParseState(t *testing.T) {
	line := "screen=ingame hotkeys=101,65535,107,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535,65535 character_panel=0"
	got, err := ParseState(line)
	if err != nil || got[0] != 101 || got[1] != Unbound || got[2] != 107 {
		t.Fatalf("%v %v", got, err)
	}
	for _, bad := range []string{"screen=ingame", "hotkeys=1,2,3", "hotkeys=a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p", "hotkeys= x"} {
		if _, err := ParseState(bad); err == nil {
			t.Fatalf("%q must fail", bad)
		}
	}
}
