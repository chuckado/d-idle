package drill

import (
	"context"
	"strings"
	"testing"
)

// fakeGame is a GameLab whose methods the tests override one by one; any other call panics on the nil interface, which
// the runner reports as a FAIL, so a test only has to implement what its drill calls.
type fakeGame struct {
	GameLab
	w      *fakeWorld
	gold   int
	dur    [2]int
	knob   int
	opened []int
	stash  int
}

func (f *fakeGame) TView(context.Context) (TownView, error) {
	return TownView{InTown: true, Gold: f.gold, Stash: f.stash, Life: 40, Max: 40, Pack: map[string]int{"isc": 1}, DurCur: f.dur[0], DurMax: f.dur[1]}, nil
}
func (f *fakeGame) Tick(context.Context, int) error { return nil }
func (f *fakeGame) ClearPack(_ context.Context, r *Run) error {
	_, err := r.Cheat(context.Background(), "debug clearinv", "empty the pack")
	return err
}
func (f *fakeGame) SetGold(_ context.Context, r *Run, n int) error {
	f.gold = n
	_, err := r.Cheat(context.Background(), "debug stat 14", "gold")
	return err
}
func (f *fakeGame) Wear(_ context.Context, r *Run) error {
	f.dur[0] = 1
	_, err := r.Cheat(context.Background(), "debug wear", "wear")
	return err
}
func (f *fakeGame) RepairKnob() int { return f.knob }
func (f *fakeGame) Chores(_ context.Context, plan string) ChoreInfo {
	if 100*f.dur[0]/f.dur[1] < f.knob && f.gold > 0 { // repair at a price of 10
		f.dur[0], f.gold = f.dur[1], f.gold-10
		return ChoreInfo{Plan: plan, Steps: []string{"repair ran"}, Repaired: true, Spent: 10}
	}
	return ChoreInfo{Plan: plan, Steps: []string{"repair skipped"}}
}
func (f *fakeGame) OpenNPC(_ context.Context, class int) OpenInfo {
	f.opened = append(f.opened, class)
	return OpenInfo{Class: class, Flag: []string{"npc-dialog"}, Ticks: 20}
}
func (f *fakeGame) CloseAll(context.Context) error { return nil }

func run1(t *testing.T, l GameLab, id string) Result {
	t.Helper()
	r := &Runner{W: &fakeWorld{}, Repeat: 2}
	var d Drill
	for _, x := range D3(l) {
		if x.ID == id {
			d = x
		}
	}
	if d.ID == "" {
		t.Fatalf("no drill %s", id)
	}
	return r.RunGroup(context.Background(), []Drill{d}).Results[0]
}

func TestD3StartingGoldIsRecorded(t *testing.T) {
	res := run1(t, &fakeGame{}, "D3.g0")
	if res.Verdict != Pass || !strings.Contains(res.Observed, "purse 0") || !res.Deterministic {
		t.Fatalf("%+v", res)
	}
}

func TestD3NPCsOpenAndClose(t *testing.T) {
	f := &fakeGame{}
	res := run1(t, f, "D3.n")
	if res.Verdict != Pass || len(f.opened) != 10 { // 5 NPCs, twice
		t.Fatalf("%+v opened %v", res, f.opened)
	}
}

func TestD3RepairRunsBelowTheKnobAndIsBlockedAboveIt(t *testing.T) {
	hi := run1(t, &fakeGame{gold: 500, dur: [2]int{20, 20}, knob: 30}, "D3.r")
	if hi.Verdict != Pass || !strings.Contains(hi.Reason, "10 gold") {
		t.Fatalf("default knob: %+v", hi)
	}
	low := run1(t, &fakeGame{gold: 500, dur: [2]int{20, 20}, knob: 3}, "D3.r")
	if low.Verdict != Blocked || !strings.Contains(low.Reason, "repair knob") {
		t.Fatalf("knob 3: %+v", low)
	}
}

func TestEveryGroupBlockedDrillNamesItsMissingPiece(t *testing.T) {
	for _, g := range [][]Drill{D3(nil), D4(nil), D5(nil), D6(nil), D7(nil), D8(nil), D9(nil), H(nil)} {
		seen := map[string]bool{}
		for _, d := range g {
			if seen[d.ID] {
				t.Errorf("duplicate id %s", d.ID)
			}
			seen[d.ID] = true
			if d.Do == nil && len(d.Needs) == 0 {
				t.Errorf("%s has neither a body nor a missing piece", d.ID)
			}
			for _, n := range d.Needs {
				if len(n) < 20 {
					t.Errorf("%s: missing piece %q is not specific", d.ID, n)
				}
			}
		}
	}
}
