package watch

import (
	"context"
	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/vault"
	"testing"
)

// goldStub is a stubActor that runs one gold order in memory.
type goldStub struct {
	*stubActor
	info   VaultGoldInfo
	order  *VaultGoldOrder
	posted int
}

func (g *goldStub) GoldInfo(context.Context) VaultGoldInfo {
	i := g.info
	i.Order = g.order
	return i
}
func (g *goldStub) EnqueueRevealGold(_ context.Context, ids []string, max int, confirm bool) ([]VaultKeepResult, *VaultGoldOrder) {
	g.posted++
	var out []VaultKeepResult
	if g.order != nil && g.order.State == "queued" {
		for _, id := range ids {
			out = append(out, VaultKeepResult{ID: id, OK: true, Already: true})
		}
		return out, g.order
	}
	if g.info.ConfirmAt > 0 && len(ids)*g.info.ScrollGold >= g.info.ConfirmAt && !confirm {
		for _, id := range ids {
			out = append(out, VaultKeepResult{ID: id, Code: "confirm", Why: "confirm it"})
		}
		return out, nil
	}
	g.order = &VaultGoldOrder{ID: "g-1", State: "queued", Total: len(ids), Max: max, Line: "queued"}
	for _, id := range ids {
		out = append(out, VaultKeepResult{ID: id, OK: true})
	}
	return out, g.order
}
func (g *goldStub) CancelGold() (VaultGoldOrder, bool) {
	if g.order == nil || g.order.State != "queued" {
		return VaultGoldOrder{}, false
	}
	g.order.State = "cancelled"
	return *g.order, true
}

func TestVaultGoldBarAndOrderRoutes(t *testing.T) {
	s := testVault(t, 120, 200)
	g := &goldStub{stubActor: &stubActor{st: s, charges: 2}, info: VaultGoldInfo{On: true, ScrollGold: 80, Purse: 500, Stash: 7200, Reserve: 200, ConfirmAt: 2000}}
	api := massAPI(t, s, g.stubActor, true)
	api.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Act: g, MassFilter: api.desk().MassFilter, Quality: api.desk().Quality, ScrollGold: 80})
	_, v := getVault(t, api, "/api/vault?limit=5")
	b := v.Act.Reveal.Bar
	if b == nil || b.Total == 0 || v.Act.Reveal.Gold == nil || !v.Act.Reveal.Gold.On {
		t.Fatalf("no bar: %+v %+v", v.Act.Reveal.Bar, v.Act.Reveal.Gold)
	}
	var gold *bulkid.BarOption
	for i := range b.Options {
		if b.Options[i].ID == "gold" {
			gold = &b.Options[i]
		}
	}
	if gold == nil || !gold.OK || gold.Count != min(b.Total, 150) || gold.Cost != (gold.Count-2)*80 {
		t.Fatalf("gold option %+v of %+v", gold, b)
	}
	ids := gold.IDs
	code, rep, _ := postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: ids, MaxGold: gold.Cost})
	if code != 200 || len(rep.Done) != len(ids) || g.order == nil {
		t.Fatalf("post: %d %+v", code, rep)
	}
	// the page sees the order and the bar refuses a second one
	_, v = getVault(t, api, "/api/vault?limit=5")
	if v.Act.Reveal.Gold.Order == nil || v.Act.Reveal.Gold.Order.State != "queued" {
		t.Fatalf("order not shown: %+v", v.Act.Reveal.Gold)
	}
	for _, o := range v.Act.Reveal.Bar.Options {
		if o.ID == "gold" && o.OK {
			t.Fatalf("a running order must close the gold option: %+v", o)
		}
	}
	if code, rep, _ = postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: ids}); code != 200 || len(rep.Already) != len(ids) {
		t.Fatalf("idempotent post: %d %+v", code, rep)
	}
	if code, _, _ = postVault(t, api, VaultRevealGoldCancelRoute, vaultBody{}); code != 200 || g.order.State != "cancelled" {
		t.Fatalf("cancel: %d %+v", code, g.order)
	}
	if code, _, _ = postVault(t, api, VaultRevealGoldCancelRoute, vaultBody{}); code != 409 {
		t.Fatalf("a second cancel has nothing to cancel: %d", code)
	}
	// over the threshold without confirm
	g.order = nil
	g.info.ConfirmAt = 100
	if code, rep, _ = postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: ids}); code != 409 || len(rep.Failed) == 0 || rep.Failed[0].Code != "confirm" {
		t.Fatalf("confirm: %d %+v", code, rep)
	}
}

func TestVaultGoldRouteWithoutAShop(t *testing.T) {
	s := testVault(t, 3, 100)
	api := massAPI(t, s, &stubActor{st: s}, true)
	if code, _, pr := postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: []string{"x"}}); code != 409 || pr.Code != "off" {
		t.Fatalf("%d %+v", code, pr)
	}
}
