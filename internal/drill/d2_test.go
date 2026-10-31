package drill

import (
	"context"
	"strings"
	"testing"
)

// simPortal models the d2d portal facts the live D2 run saw: a scroll or a tome charge is spent at the click, the
// object stands at the player's spot, a second cast closes the first, a scroll picked up with a tome goes into the tome,
// the player's walk-in works only when enter is set.
type simPortal struct {
	w       *fakeWorld
	scrolls int
	tome    int // -1 none
	ground  []string
	area    int
	x       int
	portal  *int // x of the open portal in the field, nil none
	enter   bool
	mode    int
	tick    int
	badBusy bool // the verb ignores a busy player
}

func newSimPortal() *simPortal {
	return &simPortal{w: &fakeWorld{}, tome: -1, area: townArea, enter: true}
}

func (s *simPortal) view() PortalView {
	v := PortalView{Area: s.area, InTown: s.area == townArea, X: s.x, Y: 240, Mode: s.mode, Tome: s.tome, Scrolls: s.scrolls, Tick: s.tick}
	if s.tome >= 0 {
		v.TomeID = 7
	}
	for i := 0; i < s.scrolls; i++ {
		v.ScrollIDs = append(v.ScrollIDs, 100+i)
	}
	if s.portal != nil {
		v.Any = 1
		v.PortalPos = [][2]int{{*s.portal, 243}}
		if d := *s.portal - s.x; d < 25 && d > -25 {
			v.Near = 1
		}
	}
	return v
}
func (s *simPortal) PView(context.Context) (PortalView, error) { return s.view(), nil }
func (s *simPortal) Tick(_ context.Context, n int) error {
	s.tick++
	if n > 1 {
		s.mode = 0
	}
	return nil
}
func (s *simPortal) ClearPack(_ context.Context, r *Run) error {
	_, err := r.Cheat(context.Background(), "debug clearinv", "empty pack")
	s.scrolls, s.tome = 0, -1
	return err
}
func (s *simPortal) GiveGround(_ context.Context, r *Run, code string) error {
	_, err := r.Cheat(context.Background(), "debug give "+code, "item for the drill")
	s.ground = append(s.ground, code)
	return err
}
func (s *simPortal) PickupRaw(context.Context) (int, error) {
	n := len(s.ground)
	for _, c := range s.ground {
		switch {
		case c == "tbk":
			s.tome = 1
		case s.tome >= 0:
			s.tome++ // a scroll goes into the tome
		default:
			s.scrolls++
		}
	}
	s.ground = nil
	return n, nil
}
func (s *simPortal) Goto(_ context.Context, r *Run, level int) error {
	_, err := r.Cheat(context.Background(), "debug goto 2", "field")
	s.area, s.x = level, 140
	return err
}
func (s *simPortal) PortalList(context.Context) (string, error) { return "list", nil }
func (s *simPortal) cast(tome bool) {
	if tome {
		s.tome--
	} else {
		s.scrolls--
	}
	x := s.x
	s.portal = &x
}
func (s *simPortal) RunPortal(context.Context) BotCall {
	c := BotCall{ClickTick: -1}
	switch {
	case s.scrolls > 0:
		c.Rights, c.Clicks, c.Keys, c.Ticks, c.ClickTick = 1, 1, 2, 32, 22
		if s.mode == modeSC && !s.badBusy {
			c.Ticks, c.ClickTick = 42, 37
		}
		s.cast(false)
	case s.tome > 0:
		c.Rights, c.Clicks, c.Keys, c.Ticks, c.ClickTick = 1, 1, 2, 32, 22
		s.cast(true)
	default:
		c.Err, c.NoPortal = "combat: no town portal scroll or tome", true
	}
	return c
}
func (s *simPortal) RunToTown(context.Context) BotCall {
	if s.enter && s.portal != nil {
		s.area = townArea
		return BotCall{}
	}
	return BotCall{Err: "town: area -1 not reached within 15s", Ticks: 1584}
}
func (s *simPortal) RunBack(context.Context, int) BotCall {
	s.area = field
	return BotCall{Log: []string{"back to field: portal in town id=1"}}
}
func (s *simPortal) EnterViaVerb(context.Context) error {
	if s.enter {
		s.area = townArea
		return nil
	}
	return context.DeadlineExceeded
}
func (s *simPortal) RawCast(context.Context, int) (int, error) {
	if s.tome > 0 {
		s.cast(true)
	} else {
		s.cast(false)
	}
	return 8, nil
}
func (s *simPortal) CastSkill(context.Context, int, float64, float64) error {
	s.mode = modeSC
	return nil
}
func (s *simPortal) Walk(_ context.Context, dx, _ float64) error { s.x += int(dx * 5); return nil }

func runD2(s *simPortal) map[string]Result {
	g := (&Runner{W: s.w, Repeat: 2}).RunGroup(context.Background(), D2(s))
	out := map[string]Result{}
	for _, r := range g.Results {
		out[r.ID] = r
	}
	return out
}

func TestD2Passes(t *testing.T) {
	res := runD2(newSimPortal())
	for _, id := range []string{"D2.a", "D2.t", "D2.a2", "D2.b", "D2.c", "D2.e", "D2.h", "D2.f"} {
		if r := res[id]; r.Verdict != Pass || !r.Deterministic {
			t.Errorf("%s: %s", id, r.Line())
		}
	}
	if r := res["D2.d"]; r.Verdict != Blocked || !strings.Contains(r.Reason, "debug hit") {
		t.Errorf("D2.d: %s", r.Line())
	}
}

func TestD2NoWalkInFailsAndBlocks(t *testing.T) {
	s := newSimPortal()
	s.enter = false
	res := runD2(s)
	if r := res["D2.t"]; r.Verdict != Fail || !strings.Contains(r.Reason, "did not arrive") {
		t.Errorf("D2.t: %s", r.Line())
	}
	if r := res["D2.g"]; r.Verdict != Fail {
		t.Errorf("D2.g: %s", r.Line())
	}
	if r := res["D2.f"]; r.Verdict != Blocked {
		t.Errorf("D2.f: %s", r.Line())
	}
	if r := res["D2.a"]; r.Verdict != Pass {
		t.Errorf("the verb drill must not depend on the walk-in: %s", r.Line())
	}
}

func TestD2BusyVerbThatDoesNotWaitFails(t *testing.T) {
	s := newSimPortal()
	s.badBusy = true
	if r := runD2(s)["D2.e"]; r.Verdict != Fail || !strings.Contains(r.Reason, "did not wait") {
		t.Errorf("D2.e: %s", r.Line())
	}
}
