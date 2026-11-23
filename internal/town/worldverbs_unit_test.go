package town_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/town"
	"d2lod-bot/internal/town/sim"
)

// Data-free tests of the vendor verbs (they need no game tables, unlike the sim rig of worldverbs_test.go).

type stubReader struct {
	d2.StateReader
	snap *d2.Snapshot
}

func (s stubReader) Snapshot() (*d2.Snapshot, error) { return s.snap, nil }

// stubIn is a controller that can only be asked for the world verbs; any click would panic on the nil Controller.
type stubIn struct {
	input.Controller
	buy, sell func() error
	asked     []string
}

func (s *stubIn) BuyAt(tab, x, y int) error { s.asked = append(s.asked, "buy"); return s.buy() }
func (s *stubIn) SellItem(id uint32) error  { s.asked = append(s.asked, "sell"); return s.sell() }
func (s *stubIn) TakeWaypoint(int) error    { return nil }
func (s *stubIn) EnterPortal(int) error     { return nil }
func (s *stubIn) PickupItem(uint32) error   { return nil }

type verbCase struct {
	t    *town.Town
	snap *d2.Snapshot
	in   *stubIn
}

func verbTown(cls d2.Class, on bool, caps ...string) *verbCase {
	snap := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: cls, Gold: 500, Level: 12}, Inventory: &d2.Inventory{}}
	in := &stubIn{}
	tune := env.DefaultTuning()
	e := &env.Env{ID: "t", Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Tune: &tune, In: in, Read: stubReader{snap: snap},
		Clock: &sim.Clock{T: time.Unix(1000, 0)}}
	e.Planner = &input.Planner{Enabled: func() bool { return on }, Caps: &input.BackendCaps{Verbs: caps}}
	return &verbCase{t: town.New(e, nil), snap: snap, in: in}
}

var vendorItem = d2.Unit{ID: 0x90001, Type: d2.TypeItem, X: 2, Y: 1, Item: &d2.Item{Page: 1}}

func TestBuyByVerb(t *testing.T) {
	ctx := context.Background()
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		// the verb works: the gold drops, the visit pays that and nothing is clicked
		c := verbTown(cls, true, input.CapBuy)
		c.in.buy = func() error { c.snap.Player.Gold -= 30; return nil }
		if paid, handled, err := c.t.BuyByVerbForTest(ctx, vendorItem, 500); !handled || err != nil || paid != 30 {
			t.Errorf("%v: verb buy: paid %d handled %v err %v", cls, paid, handled, err)
		}
		// knob off and no capability leave the click path alone, the verb is not even asked
		for name, c := range map[string]*verbCase{"knob off": verbTown(cls, false, input.CapBuy), "no cap": verbTown(cls, true)} {
			c.in.buy = func() error { t.Fatalf("%v %s: the verb was asked", cls, name); return nil }
			if _, handled, err := c.t.BuyByVerbForTest(ctx, vendorItem, 500); handled || err != nil {
				t.Errorf("%v %s: handled %v err %v", cls, name, handled, err)
			}
		}
		// a rule refusal is the answer: gold is ErrNoMoney, no click follows
		c = verbTown(cls, true, input.CapBuy)
		c.in.buy = func() error { return input.Refusal("gold") }
		if _, handled, err := c.t.BuyByVerbForTest(ctx, vendorItem, 500); !handled || !errors.Is(err, town.ErrNoMoney) {
			t.Errorf("%v: gold refusal: handled %v err %v", cls, handled, err)
		}
		c.in.buy = func() error { return input.Refusal("space") }
		if _, handled, err := c.t.BuyByVerbForTest(ctx, vendorItem, 500); !handled || err == nil || errors.Is(err, town.ErrNoMoney) {
			t.Errorf("%v: space refusal: handled %v err %v", cls, handled, err)
		}
		// a transport error or an older build falls back to the clicks
		c.in.buy = func() error { return errors.New("unknown verb buy") }
		if _, handled, err := c.t.BuyByVerbForTest(ctx, vendorItem, 500); handled || err != nil {
			t.Errorf("%v: fallback: handled %v err %v", cls, handled, err)
		}
		// sent but the gold never moved: an error, never a second purchase by click
		c.in.buy = func() error { return nil }
		if _, handled, err := c.t.BuyByVerbForTest(ctx, vendorItem, 500); !handled || err == nil {
			t.Errorf("%v: unverified verb: handled %v err %v", cls, handled, err)
		}
	}
}

func TestSellByVerb(t *testing.T) {
	ctx := context.Background()
	sold := func(s *d2.Snapshot) bool { return s.Player.Gold > 500 }
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		item := d2.Unit{ID: 7, Type: d2.TypeItem, Item: &d2.Item{}}
		c := verbTown(cls, true, input.CapSell)
		c.in.sell = func() error { c.snap.Player.Gold += 12; return nil }
		if handled, ok, err := c.t.SellByVerbForTest(ctx, item, sold); !handled || !ok || err != nil {
			t.Errorf("%v: verb sell: handled %v ok %v err %v", cls, handled, ok, err)
		}
		off := verbTown(cls, false, input.CapSell)
		off.in.sell = func() error { t.Fatalf("%v: asked with the knob off", cls); return nil }
		if handled, _, _ := off.t.SellByVerbForTest(ctx, item, sold); handled {
			t.Errorf("%v: knob off handled the sale", cls)
		}
		c = verbTown(cls, true, input.CapSell)
		c.in.sell = func() error { return input.Refusal("refused") }
		if handled, ok, err := c.t.SellByVerbForTest(ctx, item, sold); !handled || ok || err != nil {
			t.Errorf("%v: refusal: handled %v ok %v err %v", cls, handled, ok, err)
		}
		c.in.sell = func() error { return errors.New("transport") }
		if handled, _, _ := c.t.SellByVerbForTest(ctx, item, sold); handled {
			t.Errorf("%v: a transport error must fall back to the clicks", cls)
		}
	}
}
