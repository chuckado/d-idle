package vaultkeep

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// fakeShop is the vendor side of a gold order: it sells identify scrolls at price into the rig's fakeTools and keeps the
// purse, the stash and the reserve the town guards hold back.
type fakeShop struct {
	r                    *rig
	price                int
	purse, stash, reserv int
	bought               int
	err                  error
}

func (s *fakeShop) Price() int { return s.price }
func (s *fakeShop) Purse(context.Context) (int, int, int, error) {
	return s.purse, s.stash, s.reserv, nil
}
func (s *fakeShop) Buy(_ context.Context, n int) error {
	if s.err != nil {
		return s.err
	}
	for i := 0; i < n && s.purse+s.stash-s.reserv >= s.price; i++ {
		if s.purse < s.price { // take the missing gold out of the stash
			d := s.price - s.purse
			s.stash, s.purse = s.stash-d, s.purse+d
		}
		s.purse -= s.price
		s.r.tools.charges++
		s.bought++
	}
	return nil
}

func (r *rig) wireShop() {
	if r.shop == nil {
		return
	}
	r.k.Shop = r.shop
	r.k.GoldOn = func() bool { return r.goldOn }
	r.k.GoldConfirm = func() int { return r.confirm }
	r.k.OrderFile = func() string { return filepath.Join(r.dir, "revealgold.json") }
}

func (r *rig) withShop(purse, stash, reserve, price int) {
	r.goldOn = true
	r.shop = &fakeShop{r: r, price: price, purse: purse, stash: stash, reserv: reserve}
	r.wireShop()
}

func (r *rig) unids(n int) []string {
	var ids []string
	for i := 1; i <= n; i++ {
		ids = append(ids, r.addUnid(i, "wsp"))
	}
	return ids
}

func (r *rig) revealed(ids []string) int {
	n := 0
	for _, id := range ids {
		if r.state(id).Identified {
			n++
		}
	}
	return n
}

func TestGoldRevealBuysScrollsAndRevealsInOneVisit(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(1)
			r.withShop(300, 700, 200, 80)
			ids := r.unids(5)
			res, v := r.k.EnqueueRevealGold(context.Background(), ids, 0, false)
			for _, x := range res {
				if !x.OK {
					t.Fatalf("refused: %+v", x)
				}
			}
			if v == nil || v.State != OrderQueued || v.Max != 4*80 {
				t.Fatalf("view %+v (4 missing scrolls at 80 is the cap)", v)
			}
			if r.k.Pending() == 0 || !r.k.GoldPending() {
				t.Fatal("the bot step must see the order")
			}
			// idempotent while it runs; a different set is refused
			if again, _ := r.k.EnqueueRevealGold(context.Background(), ids, 0, false); !again[0].OK || !again[0].Already {
				t.Fatalf("same ids must be Already: %+v", again[0])
			}
			if other, _ := r.k.EnqueueRevealGold(context.Background(), ids[:2], 0, false); other[0].Code != "active" {
				t.Fatalf("a second order must be refused: %+v", other[0])
			}
			if r.revealed(ids) != 0 || r.shop.bought != 0 {
				t.Fatal("nothing happens before the visit")
			}
			lines := r.run()
			if n := r.revealed(ids); n != 5 {
				t.Fatalf("revealed %d of 5; lines %v", n, lines)
			}
			if r.shop.bought != 4 || r.shop.purse+r.shop.stash != 1000-4*80 || r.tools.charges != 0 {
				t.Fatalf("bought %d gold left %d charges %d", r.shop.bought, r.shop.purse+r.shop.stash, r.tools.charges)
			}
			o := r.k.GoldView()
			if o == nil || o.State != OrderDone || o.Done != 5 || o.Spent != 320 || o.Bought != 4 {
				t.Fatalf("order %+v", o)
			}
			if !r.heroEmpty() {
				t.Fatal("temporary copies must leave the hero")
			}
		})
	}
}

func TestGoldRevealShortKeepsTheReserve(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withShop(400, 0, 300, 80) // 100 gold above the reserve: one scroll
			ids := r.unids(4)
			if res, _ := r.k.EnqueueRevealGold(context.Background(), ids, 4*80, false); !res[0].OK {
				t.Fatalf("%+v", res[0])
			}
			r.run()
			o := r.k.GoldView()
			if o.State != OrderPartial || o.Done != 1 || o.Bought != 1 || !strings.Contains(o.Line, "short of gold") {
				t.Fatalf("order %+v", o)
			}
			if got := r.revealed(ids); got != 1 {
				t.Fatalf("revealed %d", got)
			}
			if r.shop.purse+r.shop.stash < r.shop.reserv {
				t.Fatalf("the reserve was spent: %d < %d", r.shop.purse+r.shop.stash, r.shop.reserv)
			}
			if r.k.GoldPending() {
				t.Fatal("a finished order must not keep the bot busy")
			}
		})
	}
}

func TestGoldRevealCapAndNothingAffordable(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(0)
	r.withShop(5000, 0, 100, 80)
	ids := r.unids(6)
	r.k.EnqueueRevealGold(context.Background(), ids, 170, false) // two scrolls fit the cap
	r.run()
	if o := r.k.GoldView(); o.Bought != 2 || o.Spent != 160 || o.State != OrderPartial || !strings.Contains(o.Note, "cap") {
		t.Fatalf("%+v", o)
	}
	r2 := newRig(t, "paladin")
	r2.withTools(0)
	r2.withShop(100, 0, 100, 80) // nothing above the reserve
	ids = r2.unids(3)
	r2.k.EnqueueRevealGold(context.Background(), ids, 240, false)
	r2.run()
	if o := r2.k.GoldView(); o.State != OrderPartial || o.Bought != 0 || r2.shop.purse != 100 || r2.revealed(ids) != 0 {
		t.Fatalf("%+v purse %d", o, r2.shop.purse)
	}
}

func TestGoldRevealCainIsFree(t *testing.T) {
	r := newRig(t, "paladin")
	r.withTools(0)
	r.tools.cain = true
	r.withShop(500, 0, 100, 80)
	ids := r.unids(3)
	r.k.EnqueueRevealGold(context.Background(), ids, 0, false)
	r.run()
	if o := r.k.GoldView(); o.State != OrderDone || o.Spent != 0 || r.shop.bought != 0 || r.revealed(ids) != 3 {
		t.Fatalf("%+v bought %d", o, r.shop.bought)
	}
}

func TestGoldRevealRefusalsAndCancel(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(0)
	r.withShop(5000, 0, 100, 100)
	ids := r.unids(3)
	ctx := context.Background()
	r.goldOn = false
	if res, _ := r.k.EnqueueRevealGold(ctx, ids, 0, false); res[0].Code != "off" {
		t.Fatalf("knob 0: %+v", res[0])
	}
	r.goldOn, r.confirm = true, 250
	if res, _ := r.k.EnqueueRevealGold(ctx, ids, 0, false); res[0].Code != "confirm" {
		t.Fatalf("a 300 gold order needs confirm: %+v", res[0])
	}
	if res, _ := r.k.EnqueueRevealGold(ctx, ids, 0, true); !res[0].OK {
		t.Fatalf("confirmed: %+v", res[0])
	}
	v, ok := r.k.CancelGold()
	if !ok || v.State != OrderCancelled {
		t.Fatalf("cancel %+v %v", v, ok)
	}
	r.run()
	if r.shop.bought != 0 || r.revealed(ids) != 0 || r.k.GoldPending() {
		t.Fatal("a cancelled order buys and reveals nothing")
	}
	// a finished order is replaced by a new one
	if res, _ := r.k.EnqueueRevealGold(ctx, ids[:1], 0, true); !res[0].OK {
		t.Fatalf("after cancel: %+v", res[0])
	}
}

func TestGoldRevealSurvivesARestart(t *testing.T) {
	r := newRig(t, "paladin")
	r.withTools(0)
	r.withShop(2000, 0, 100, 80)
	ids := r.unids(2)
	r.k.EnqueueRevealGold(context.Background(), ids, 0, false)
	r.open() // the bot restarts: the order is read back from the file
	if !r.k.GoldPending() {
		t.Fatal("the order must survive a restart")
	}
	r.run()
	if o := r.k.GoldView(); o.State != OrderDone || r.revealed(ids) != 2 {
		t.Fatalf("%+v", o)
	}
}

func TestGoldRevealShopErrorBuysNothing(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(0)
	r.withShop(2000, 0, 100, 80)
	r.shop.err = errors.New("the vendor is not in reach")
	ids := r.unids(2)
	r.k.EnqueueRevealGold(context.Background(), ids, 0, false)
	lines := r.run()
	o := r.k.GoldView()
	if o.State != OrderPartial || o.Spent != 0 || r.revealed(ids) != 0 || !strings.Contains(strings.Join(lines, ";"), "vendor") {
		t.Fatalf("%+v %v", o, lines)
	}
}
