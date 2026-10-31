package drill

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// simLab models d2d's belt as seen in the live probe: a ground potion goes to the lowest free belt box, else to the pack;
// the belt fill fills the lowest free box (the unfixed behaviour) or the plan column (byColumn); the drink takes the
// asked kind from the first column that holds it.
type simLab struct {
	w         *fakeWorld
	slots     int
	belt      map[int]string
	pack      map[string]int
	ground    []string
	life, mp  int
	byColumn  bool
	pickitOK  bool
	engineCap int // belt boxes the engine really has (0: slots)
}

func newSim() *simLab {
	return &simLab{w: &fakeWorld{}, slots: 4, belt: map[int]string{}, pack: map[string]int{}, life: 40, mp: 35, pickitOK: true}
}

func (s *simLab) cap() int {
	if s.engineCap > 0 {
		return s.engineCap
	}
	return s.slots
}
func (s *simLab) View(context.Context) (View, error) {
	v := View{Life: s.life, MaxLife: 40, Mana: s.mp, MaxMana: 35, Level: 1, BeltSlots: s.slots, Belt: map[int]string{}, Pack: map[string]int{}, Ground: append([]string(nil), s.ground...)}
	for b, c := range s.belt {
		v.Belt[b] = c
	}
	for c, n := range s.pack {
		v.Pack[c] = n
	}
	return v, nil
}
func (s *simLab) Tick(context.Context, int) error { return nil }
func (s *simLab) PlanCols(context.Context) ([4]Kind, error) {
	return sPlan, nil
}
func (s *simLab) GiveGround(_ context.Context, r *Run, code string) error {
	_, err := r.Cheat(context.Background(), "debug give "+code, "potion for the drill")
	s.ground = append(s.ground, code)
	return err
}
func (s *simLab) ClearPack(_ context.Context, r *Run) error {
	_, err := r.Cheat(context.Background(), "debug clearinv", "empty pack")
	s.pack = map[string]int{}
	return err
}
func (s *simLab) SetLife(_ context.Context, r *Run, p int) error {
	_, err := r.Cheat(context.Background(), "debug stat 6", "low life")
	s.life = p
	return err
}
func (s *simLab) SetMana(_ context.Context, r *Run, p int) error {
	_, err := r.Cheat(context.Background(), "debug stat 8", "low mana")
	s.mp = p
	return err
}
func (s *simLab) EquipBelt(_ context.Context, r *Run, code string) error {
	_, err := r.Cheat(context.Background(), "debug equip "+code+" 8", "belt type")
	s.slots = map[string]int{"lbl": 8, "mbl": 12}[code]
	return err
}
func (s *simLab) free() (int, bool) {
	for b := 0; b < s.cap(); b++ {
		if _, ok := s.belt[b]; !ok {
			return b, true
		}
	}
	return 0, false
}
func (s *simLab) PickupRaw(context.Context) (int, error) {
	n := len(s.ground)
	for _, c := range s.ground {
		if b, ok := s.free(); ok {
			s.belt[b] = c
		} else {
			s.pack[c]++
		}
	}
	s.ground = nil
	return n, nil
}
func (s *simLab) DrainBelt(context.Context) error { s.belt = map[int]string{}; return nil }

var sPlan = [4]Kind{"hp", "hp", "mp", "hp"}

func (s *simLab) FillBelt(context.Context) (FillInfo, error) {
	moved := 0
	for _, c := range []string{"hp1", "mp1"} {
		for s.pack[c] > 0 {
			var box int
			var ok bool
			if s.byColumn {
				for b := 0; b < s.cap(); b++ {
					if _, used := s.belt[b]; !used && sPlan[b%4] == PotionKind(c) {
						box, ok = b, true
						break
					}
				}
			} else {
				box, ok = s.free()
			}
			if !ok {
				break
			}
			s.belt[box] = c
			s.pack[c]--
			moved++
		}
	}
	return FillInfo{Moved: moved}, nil
}
func (s *simLab) Pickit(context.Context) (PickInfo, error) {
	if !s.pickitOK {
		return PickInfo{Seen: len(s.ground), Skipped: []string{"Minor Healing Potion: keep [C] potion already in the pack"}}, nil
	}
	n := len(s.ground)
	s.PickupRaw(context.Background())
	return PickInfo{Seen: n, Picked: []string{"hp1 (keep)"}}, nil
}
func (s *simLab) Drink(_ context.Context, k Kind) error {
	var box []int
	for b := range s.belt {
		box = append(box, b)
	}
	sort.Ints(box)
	for _, b := range box {
		if PotionKind(s.belt[b]) == k {
			delete(s.belt, b)
			if k == "hp" {
				s.life += 20
			} else {
				s.mp += 20
			}
			return nil
		}
	}
	return errors.New("no potion")
}

func runD1(t *testing.T, s *simLab, belts []string) map[string]Result {
	t.Helper()
	g := (&Runner{W: s.w, Repeat: 2}).RunGroup(context.Background(), D1(s, belts))
	out := map[string]Result{}
	for _, r := range g.Results {
		out[r.ID] = r
	}
	return out
}

func TestD1UnfixedBehaviourFails(t *testing.T) {
	s := newSim()
	s.pickitOK = false
	res := runD1(t, s, []string{"lbl"})
	if r := res["D1.P1"]; r.Verdict != Fail || !strings.Contains(r.Reason, "wrong columns") {
		t.Errorf("P1: %+v", r)
	}
	if r := res["D1.P7"]; r.Verdict != Fail || !strings.Contains(r.Reason, "already in the pack") {
		t.Errorf("P7: %+v", r)
	}
	if r := res["D1.P7a"]; r.Verdict != Pass || !strings.Contains(r.Reason, "belt") {
		t.Errorf("P7a: %+v", r)
	}
	if r := res["D1.P3"]; r.Verdict != Pass {
		t.Errorf("P3: %+v", r)
	}
	for _, id := range []string{"D1.P1s", "D1.P3m"} {
		if r := res[id]; r.Verdict != Blocked || r.Reason == "" {
			t.Errorf("%s: %+v", id, r)
		}
	}
}

func TestD1FixedBehaviourPasses(t *testing.T) {
	s := newSim()
	s.byColumn = true
	res := runD1(t, s, []string{"lbl", "mbl"})
	for _, id := range []string{"D1.P1", "D1.P7", "D1.P7a", "D1.P2", "D1.P3"} {
		if r := res[id]; r.Verdict != Pass || !r.Deterministic {
			t.Errorf("%s: %s", id, r.Line())
		}
	}
}

func TestD1P2CatchesReaderMismatch(t *testing.T) {
	s := newSim()
	s.engineCap = 4 // the engine belt has 4 boxes whatever the reader derives from the belt type
	r := runD1(t, s, []string{"lbl"})["D1.P2"]
	if r.Verdict != Fail || !strings.Contains(r.Reason, "reader says 8") {
		t.Fatalf("%s", r.Line())
	}
}

func TestD1CheatsAreLedgeredWithReasons(t *testing.T) {
	s := newSim()
	runD1(t, s, []string{"lbl"})
	if len(s.w.cheats) == 0 {
		t.Fatal("no cheats recorded")
	}
	for _, c := range s.w.cheats {
		if !strings.HasPrefix(c, "drills:D1.") {
			t.Errorf("actor missing: %s", c)
		}
	}
	_ = fmt.Sprint()
}
