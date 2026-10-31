package town_test

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/town"
	"d2lod-bot/internal/town/hand"
)

// putInput is the d2d input: a belt put that lands the potion in belt box `box` (GX = box, GY 0, like the fork's dump).
type putInput struct {
	input.Controller
	r     *rig
	fail  []error // errors the next puts return, one each
	boxes []int
}

func (p *putInput) BeltPut(id uint32, box int) error {
	if len(p.fail) > 0 {
		err := p.fail[0]
		p.fail = p.fail[1:]
		return err
	}
	for i, u := range p.r.W.Items {
		if u.ID == id {
			p.r.W.Items[i].Item.Where, p.r.W.Items[i].Mode = d2.ContBelt, d2.ItemModeBelt
			p.r.W.Items[i].X, p.r.W.Items[i].Y = box, 0
			p.boxes = append(p.boxes, box)
			return nil
		}
	}
	return hand.ErrNoItem
}

func d2dRig(t *testing.T) (*rig, *putInput) {
	r := newRig(t)
	pi := &putInput{Controller: r.Env.In, r: r}
	r.Env.In = pi
	return r, pi
}

// LOG SIGNATURE (BELTFIX F1): a belt put fails once with "item not found" right after a re-entry. The belt must not be
// marked full: the same call retries after a fresh read, and the next call fills too.
func TestBeltPutItemNotFoundDoesNotStick(t *testing.T) {
	r, pi := d2dRig(t)
	addPack(r, "hp1", 4)
	pi.fail = []error{hand.ErrNoItem}
	rep, err := r.T.FillBelt(context.Background())
	if err != nil || rep.Moved < 2 || rep.Failed != 0 || rep.Moved != count(r.W, d2.ContBelt) {
		t.Fatalf("retry after a vanished id: %+v err %v belt %d", rep, err, count(r.W, d2.ContBelt))
	}

	// every put fails with item-not-found: nothing sticks, the next call (a later tick) fills
	r, pi = d2dRig(t)
	addPack(r, "hp1", 4)
	pi.fail = []error{hand.ErrNoItem, hand.ErrNoItem, hand.ErrNoItem, hand.ErrNoItem}
	_, _ = r.T.FillBelt(context.Background())
	if !r.T.NeedsBeltFill(mustSnap(t, r)) {
		t.Fatal("belt marked full after item-not-found")
	}
	rep, err = r.T.FillBelt(context.Background())
	if err != nil || rep.Moved < 2 {
		t.Fatalf("second call: %+v err %v", rep, err)
	}
}

// Another failure kind still sticks, and a new game (ForgetBelt) or a changed pack lifts the mark.
func TestBeltFullMarkResets(t *testing.T) {
	r, pi := d2dRig(t)
	addPack(r, "hp1", 4)
	boom := errors.New("boom")
	pi.fail = []error{boom, boom}
	if _, err := r.T.FillBelt(context.Background()); err == nil {
		t.Fatal("want the failure")
	}
	if r.T.NeedsBeltFill(mustSnap(t, r)) {
		t.Fatal("a real failure should stick until something changes")
	}
	r.T.ForgetBelt()
	if !r.T.NeedsBeltFill(mustSnap(t, r)) {
		t.Fatal("ForgetBelt did not reset the mark")
	}
	pi.fail = []error{boom, boom}
	_, _ = r.T.FillBelt(context.Background())
	if r.T.NeedsBeltFill(mustSnap(t, r)) {
		t.Fatal("mark not set again")
	}
	addPack(r, "hp1", 1) // the pack's potion set changed
	if !r.T.NeedsBeltFill(mustSnap(t, r)) {
		t.Fatal("a changed pack did not lift the mark")
	}
}

// BELTFIX F8: a mana potion is never put above a healing column, and a healing potion not above a mana column.
func TestBeltPutRespectsColumnKind(t *testing.T) {
	r, pi := d2dRig(t)
	r.W.Player.Level = 14
	// row 0: hp, mp, mp, mp (columns 0..3)
	r.W.Add("hp1", d2.ContBelt, 0, 0)
	for c := 1; c < 4; c++ {
		r.W.Add("mp1", d2.ContBelt, c, 0)
	}
	addPack(r, "hp1", 2)
	if _, err := r.T.FillBelt(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, b := range pi.boxes {
		if b%4 != 0 {
			t.Errorf("healing potion put in box %d (column %d), not above the healing column", b, b%4)
		}
	}
	// a belt whose only free boxes are above mana columns takes no healing potion
	r, pi = d2dRig(t)
	r.W.Add("hp1", d2.ContBelt, 0, 0)
	r.W.Add("hp1", d2.ContBelt, 4, 0)
	r.W.Add("hp1", d2.ContBelt, 8, 0)
	r.W.Add("hp1", d2.ContBelt, 12, 0)
	for c := 1; c < 4; c++ {
		r.W.Add("mp1", d2.ContBelt, c, 0)
	}
	addPack(r, "hp1", 1)
	_, _ = r.T.FillBelt(context.Background())
	for _, b := range pi.boxes {
		if b%4 != 0 {
			t.Errorf("healing potion put in box %d above a mana column", b)
		}
	}
}

// BELTFIX F3 + the empty-belt visit: a visit that starts with an empty belt and potions in the pack ends with a full
// belt, and potions bought by the shop are belted before the visit ends.
func TestVisitEndsWithBeltFilled(t *testing.T) {
	r, _ := d2dRig(t)
	dressed(r.W)
	addPack(r, "hp1", 4)
	if _, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true}); err != nil {
		t.Fatal(err)
	}
	if n := count(r.W, d2.ContBelt); n == 0 {
		t.Fatal("belt still empty after the visit")
	}
	left := 0
	for _, u := range r.W.Items {
		if u.Item.Where == d2.ContInventory && r.Env.Data.Items[u.Class].Code == "hp1" {
			left++
		}
	}
	if left > 0 && count(r.W, d2.ContBelt) < 4 {
		t.Errorf("%d healing potions left in the pack with %d belt slots used", left, count(r.W, d2.ContBelt))
	}
}

// BELTFIX F3: d2d purchases land in the pack; the belt step after the shop belts them in the same visit.
func TestShopPurchaseIsBelted(t *testing.T) {
	r, _ := d2dRig(t)
	r.W.NoAutoBelt = true
	dressed(r.W)
	rep, err := r.T.Chores(context.Background(), town.ChorePlan{Potions: true})
	if err != nil {
		t.Fatal(err)
	}
	if bought(rep) == 0 {
		t.Skipf("nothing bought: %+v", rep.Steps)
	}
	if count(r.W, d2.ContBelt) == 0 {
		t.Fatalf("bought %v but the belt is empty: %+v", rep.Bought, rep.Steps)
	}
}
