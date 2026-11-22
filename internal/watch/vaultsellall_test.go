package watch

import (
	"context"
	"testing"

	"d2lod-bot/internal/vault"
)

// Sell all over HTTP (scope all_sellable): the page sends an order id and never a list.

type fakeAllActor struct {
	*fakeSaleActor
	orders  []string
	code    string
	queued  []string
	ids     []string
	sellErr string
}

func (a *fakeAllActor) SellAllQuote(context.Context) VaultSaleQuote {
	return VaultSaleQuote{Asked: 30, Count: 20, Total: 400, Stash: 100, StashAfter: 500, Reasons: []VaultSaleReason{{Code: "pinned", N: 6}, {Code: "order", N: 4}}}
}
func (a *fakeAllActor) SellAll(_ context.Context, id string) (string, string, *VaultSale) {
	a.orders = append(a.orders, id)
	if a.code != "" {
		return a.code, "refused: " + a.code, nil
	}
	return "", "", &VaultSale{ID: "sa-" + id, State: "done", Items: 20, Done: 20, Gold: 400, StashAfter: 500, All: true, Line: "Sold 20 for 400 gold; 10 skipped: 6 pinned, 4 in an order"}
}
func (a *fakeAllActor) SellAllIDs() []string { return a.ids }

func TestVaultSellAllRoute(t *testing.T) {
	s := testVault(t, 6, 100)
	sa := &fakeSaleActor{stubActor: &stubActor{st: s}, on: true}
	_, all := getVault(t, actAPI(s, sa.stubActor), "/api/vault?limit=5")
	a := &fakeAllActor{fakeSaleActor: sa, ids: []string{all.Items[0].ID, all.Items[1].ID}}
	api := actAPI(s, sa.stubActor)
	api.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Act: a})

	// no ids may come with the scope; the scope must be the known one; a confirm is needed
	if code, _, pr := postVault(t, api, VaultSellRoute, vaultBody{Scope: "all_sellable", IDs: []string{"a"}, Confirm: true, OrderID: "o1"}); code != 400 || pr.Code != "scope" || len(a.orders) != 0 {
		t.Fatalf("ids with the scope: %d %+v", code, pr)
	}
	if code, _, pr := postVault(t, api, VaultSellRoute, vaultBody{Scope: "everything", Confirm: true, OrderID: "o1"}); code != 400 || pr.Code != "scope" {
		t.Fatalf("unknown scope: %d %+v", code, pr)
	}
	if code, _, pr := postVault(t, api, VaultSellRoute, vaultBody{Scope: "all_sellable", OrderID: "o1"}); code != 400 || pr.Code != "confirm" || len(a.orders) != 0 {
		t.Fatalf("no confirm: %d %+v", code, pr)
	}
	// the quote changes nothing
	var q VaultSaleQuote
	if c := postRaw(t, api, VaultSellQuote, vaultBody{Scope: "all_sellable"}, &q); c != 200 || q.Count != 20 || q.Total != 400 || len(a.orders) != 0 {
		t.Fatalf("quote: %d %+v", c, q)
	}
	// the order: no ids reach the bot, the order id does
	code, rep, _ := postVault(t, api, VaultSellRoute, vaultBody{Scope: "all_sellable", Confirm: true, OrderID: "o1"})
	if code != 200 || !rep.OK || rep.Sale == nil || !rep.Sale.All || rep.Sale.Done != 20 || len(a.orders) != 1 || a.orders[0] != "o1" || len(sa.sold) != 0 {
		t.Fatalf("order: %d %+v orders %v", code, rep, a.orders)
	}
	// a refusal of the bot (measured run, active sale) is a 409 with its code
	a.code = "measured"
	if code, _, pr := postVault(t, api, VaultSellRoute, vaultBody{Scope: "all_sellable", Confirm: true, OrderID: "o2"}); code != 409 || pr.Code != "measured" {
		t.Fatalf("refusal: %d %+v", code, pr)
	}
	// the fork lacks the verbs: the same list goes to the old town-visit queue
	a.code = ""
	sa.on = false
	code, rep, _ = postVault(t, api, VaultSellRoute, vaultBody{Scope: "all_sellable", Confirm: true, OrderID: "o3"})
	if code == 409 || rep.Sale != nil || len(rep.Done)+len(rep.Failed) != 2 {
		t.Fatalf("queue: %d %+v", code, rep)
	}
	if len(a.orders) != 2 { // o1 and o2 only: the queue path never made an instant order
		t.Fatalf("orders %v", a.orders)
	}
}
