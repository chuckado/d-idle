package town_test

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/town"
)

// The fork's buy and sell verbs (docs/process/input-verbs2.md) replace the vendor clicks only behind the knob
// input.planner AND the backend's input.buy / input.sell capability; a rule refusal is never repeated by a click, any
// other failure falls back to the clicks. The fake engine below does what the fork's handlers do (the sim world's own
// right-click buy, a removed item and gold for a sale).

type verbEngine struct {
	r           *rig
	raw         input.Controller
	fail        map[string]error
	buys, sells int
}

func (v *verbEngine) BuyAt(tab, x, y int) error {
	if err := v.fail["buy"]; err != nil {
		return err
	}
	v.buys++
	p := v.r.W.Lay.Vendor.Cell(x, y)
	return v.raw.Click(input.Right, p.X, p.Y)
}

func (v *verbEngine) SellItem(id uint32) error {
	if err := v.fail["sell"]; err != nil {
		return err
	}
	v.sells++
	w := v.r.W
	for i, u := range w.Items {
		if u.ID == id {
			w.Items = append(w.Items[:i], w.Items[i+1:]...)
			w.Player.Gold += 7
			return nil
		}
	}
	return input.Refusal("item")
}
func (v *verbEngine) TakeWaypoint(int) error  { return errors.New("unused") }
func (v *verbEngine) EnterPortal(int) error   { return errors.New("unused") }
func (v *verbEngine) PickupItem(uint32) error { return errors.New("unused") }

// rightCounter counts the bot's own right-clicks and offers the verbs.
type rightCounter struct {
	input.Controller
	v      *verbEngine
	rights int
}

func (c *rightCounter) Click(b input.Button, x, y int, mods ...input.Mod) error {
	if b == input.Right {
		c.rights++
	}
	return c.Controller.Click(b, x, y, mods...)
}
func (c *rightCounter) WorldVerbs() (input.WorldVerbs, bool) { return c.v, true }

func verbRig(t *testing.T, cls d2.Class, on bool, caps ...string) (*rig, *verbEngine, *rightCounter) {
	t.Helper()
	r := earlyRig(t, cls, 14)
	r.W.Player.Gold = 3000
	dressed(r.W)
	v := &verbEngine{r: r, raw: r.Env.In, fail: map[string]error{}}
	c := &rightCounter{Controller: r.Env.In, v: v}
	r.Env.In = c
	r.Env.Planner = &input.Planner{Enabled: func() bool { return on }, Caps: &input.BackendCaps{Verbs: caps}}
	return r, v, c
}

func TestBuyVerbReplacesTheRightClick(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		base, _, bc := verbRig(t, cls, false)
		want, err := base.T.Chores(context.Background(), town.ChorePlan{Potions: true})
		if err != nil || len(want.Bought) == 0 || bc.rights == 0 {
			t.Fatalf("%v: baseline bought %+v rights %d err %v", cls, want.Bought, bc.rights, err)
		}
		r, v, c := verbRig(t, cls, true, input.CapBuy)
		got, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true})
		if err != nil {
			t.Fatal(err)
		}
		if v.buys == 0 || c.rights != 0 {
			t.Errorf("%v: %d buy verbs, %d bot right-clicks, want verbs and no clicks", cls, v.buys, c.rights)
		}
		for k, n := range want.Bought {
			if got.Bought[k] != n {
				t.Errorf("%v: bought %v with the verb, %v with clicks", cls, got.Bought, want.Bought)
				break
			}
		}
		if r.W.Player.Gold != base.W.Player.Gold {
			t.Errorf("%v: gold %d with the verb, %d with clicks", cls, r.W.Player.Gold, base.W.Player.Gold)
		}
	}
}

func TestBuyVerbOffWithoutKnobOrCapability(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for name, mk := range map[string]func() (*rig, *verbEngine, *rightCounter){
			"knob off": func() (*rig, *verbEngine, *rightCounter) { return verbRig(t, cls, false, input.CapBuy) },
			"no cap":   func() (*rig, *verbEngine, *rightCounter) { return verbRig(t, cls, true) },
		} {
			r, v, c := mk()
			if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true}); err != nil {
				t.Fatal(err)
			}
			if v.buys != 0 || c.rights == 0 {
				t.Errorf("%v %s: %d verbs, %d clicks: the legacy path must run unchanged", cls, name, v.buys, c.rights)
			}
		}
	}
}

func TestBuyVerbRefusalIsNotRepeatedByAClick(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, v, c := verbRig(t, cls, true, input.CapBuy)
		v.fail["buy"] = input.Refusal("space")
		rep, _ := r.T.Chores(context.Background(), town.ChorePlan{Potions: true})
		if c.rights != 0 || len(rep.Bought) != 0 {
			t.Errorf("%v: a refused buy was repeated by a click or counted: rights %d bought %+v", cls, c.rights, rep.Bought)
		}
	}
}

func TestBuyVerbTransportFailureFallsBackToClicks(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, v, c := verbRig(t, cls, true, input.CapBuy)
		v.fail["buy"] = errors.New("d2dctl: unknown verb buy") // an older build
		rep, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true})
		if err != nil || len(rep.Bought) == 0 || c.rights == 0 {
			t.Errorf("%v: no fallback: bought %+v rights %d err %v", cls, rep.Bought, c.rights, err)
		}
	}
}

func TestSellVerbSellsThePack(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, v, _ := verbRig(t, cls, true, input.CapSell)
		r.Env.Tune.Resource.GearDecisions = 0
		w := r.W
		w.Player.Gold = 40
		for i, code := range []string{"cap", "lbt", "qui", "qui"} {
			w.Add(code, d2.ContInventory, i*2, 0)
		}
		rep, err := r.T.Chores(context.Background(), town.ChorePlan{Sell: true})
		if err != nil {
			t.Fatal(err)
		}
		if n := count(w, d2.ContInventory); n != 0 || v.sells == 0 {
			t.Errorf("%v: %d items left, %d sell verbs, steps %+v", cls, n, v.sells, rep.Steps)
		}
	}
}

func TestSellVerbRefusalLeavesTheItem(t *testing.T) {
	for _, cls := range []d2.Class{d2.Paladin, d2.Sorceress} {
		r, v, _ := verbRig(t, cls, true, input.CapSell)
		r.Env.Tune.Resource.GearDecisions = 0
		v.fail["sell"] = input.Refusal("refused")
		w := r.W
		w.Player.Gold = 40
		w.Add("cap", d2.ContInventory, 0, 0)
		if _, err := r.T.Chores(context.Background(), town.ChorePlan{Sell: true}); err != nil {
			t.Fatal(err)
		}
		if n := count(w, d2.ContInventory); n != 1 || w.Player.Gold != 40 {
			t.Errorf("%v: a refused sale changed the pack (%d items) or the gold (%d)", cls, n, w.Player.Gold)
		}
	}
}
