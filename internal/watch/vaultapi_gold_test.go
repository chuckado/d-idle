package watch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/vault"
)

// sealedVault is testVault whose unidentified records hold their affixes sealed (a gold reveal unlocks those only).
func sealedVault(t testing.TB, n, capacity int) *vault.Store {
	t.Helper()
	s, err := vault.OpenHero(t.TempDir(), "Sorc", vault.Options{Capacity: func() int { return capacity }, NoSync: true,
		Now: func() time.Time { return time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC) }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	for i := 0; i < n; i++ {
		it := vault.Item{Hero: "Sorc", Raw: fmt.Sprintf("%08x%08x", i, i*7919), Code: "wsp", Name: fmt.Sprintf("War Sword %d", i), Quality: 7, Ilvl: 30 + i%50,
			Chain: []string{"swor", "weap"}, Found: vault.Found{Time: time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Second), Area: "Burial Grounds"}}
		it.Value = 100 + i
		it.Sealed = "7:0:20;93:0:10"
		if _, err := s.Add(it); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}
	return s
}

// goldStub is a stubActor whose gold order settles in the call, like the real keeper: queued is never seen by the page.
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
	if g.info.ConfirmAt > 0 && len(ids)*g.info.ScrollGold >= g.info.ConfirmAt && !confirm {
		for _, id := range ids {
			out = append(out, VaultKeepResult{ID: id, Code: "confirm", Why: "confirm it"})
		}
		return out, nil
	}
	cost := len(ids) * g.info.ScrollGold
	g.info.Stash -= cost
	g.order = &VaultGoldOrder{ID: "g-1", State: "revealed", Total: len(ids), Want: len(ids), Done: len(ids), Paid: cost, StashAfter: g.info.Stash, Line: "done"}
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
	s := sealedVault(t, 40, 200)
	g := &goldStub{stubActor: &stubActor{st: s, charges: 0}, info: VaultGoldInfo{On: true, ScrollGold: 80, Stash: 7200, ConfirmAt: 5000}}
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
	if gold == nil || !gold.OK || gold.Count == 0 || gold.Cost != gold.Count*80 || gold.Line != fmt.Sprintf("Gold: %s (stash 7,200)", bulkid.Thousands(gold.Cost)) {
		t.Fatalf("gold option %+v of %+v", gold, b)
	}
	ids := gold.IDs
	code, rep, _ := postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: ids, MaxGold: gold.Cost})
	if code != 200 || len(rep.Done) != len(ids) || g.order == nil || g.order.State != "revealed" {
		t.Fatalf("post: %d %+v", code, rep)
	}
	// the page sees the finished order and the stash gold after the payment
	_, v = getVault(t, api, "/api/vault?limit=5")
	if o := v.Act.Reveal.Gold.Order; o == nil || o.State != "revealed" || o.StashAfter != 7200-gold.Cost || v.Act.Reveal.Gold.Stash != 7200-gold.Cost {
		t.Fatalf("order not shown: %+v", v.Act.Reveal.Gold)
	}
	if code, _, _ = postVault(t, api, VaultRevealGoldCancelRoute, vaultBody{}); code != 409 {
		t.Fatalf("a paid order has nothing to cancel: %d", code)
	}
	// over the threshold without confirm
	g.order = nil
	g.info.ConfirmAt = 100
	if code, rep, _ = postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: ids}); code != 409 || len(rep.Failed) == 0 || rep.Failed[0].Code != "confirm" {
		t.Fatalf("confirm: %d %+v", code, rep)
	}
}

// Gold needs no gate: it works while the hero is dead or the game is not ready (an Ether tap is the same).
func TestVaultGoldRouteIgnoresTheRevealGate(t *testing.T) {
	s := sealedVault(t, 5, 100)
	g := &goldStub{stubActor: &stubActor{st: s, revGate: "dead"}, info: VaultGoldInfo{On: true, ScrollGold: 80, Stash: 7200}}
	api := massAPI(t, s, g.stubActor, true)
	api.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Act: g, MassFilter: api.desk().MassFilter, Quality: api.desk().Quality})
	_, v := getVault(t, api, "/api/vault?limit=5")
	if code, rep, _ := postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: []string{v.Items[0].ID}}); code != 200 || len(rep.Done) != 1 {
		t.Fatalf("gold with the gate closed: %d %+v", code, rep)
	}
}

func TestVaultGoldRouteWithoutAPayer(t *testing.T) {
	s := testVault(t, 3, 100)
	api := massAPI(t, s, &stubActor{st: s}, true)
	if code, _, pr := postVault(t, api, VaultRevealGoldRoute, vaultBody{IDs: []string{"x"}}); code != 409 || pr.Code != "off" {
		t.Fatalf("%d %+v", code, pr)
	}
}
