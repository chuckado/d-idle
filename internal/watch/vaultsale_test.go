package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"d2lod-bot/internal/vault"
)

// postRaw posts a body to an answer route and decodes the reply into out; it returns the status.
func postRaw(t testing.TB, a *VaultAPI, route string, body, out any) int {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", route, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.PostHandler(AnswerOptions{}, route).ServeHTTP(w, r)
	_ = json.Unmarshal(w.Body.Bytes(), out)
	return w.Code
}

// fakeSaleActor is a stubActor that also is the instant sale: SaleOn, SaleQuote and SellNow answer from fields.
type fakeSaleActor struct {
	*stubActor
	on     bool
	quote  VaultSaleQuote
	sold   [][]string
	confrm []bool
}

func (a *fakeSaleActor) SaleOn(context.Context) (bool, string) { return a.on, "off" }
func (a *fakeSaleActor) SaleQuote(_ context.Context, ids []string) VaultSaleQuote {
	q := a.quote
	q.Asked = len(ids)
	return q
}
func (a *fakeSaleActor) SellNow(_ context.Context, ids []string, confirm bool) ([]VaultKeepResult, *VaultSale) {
	a.sold = append(a.sold, ids)
	a.confrm = append(a.confrm, confirm)
	out := make([]VaultKeepResult, len(ids))
	for i, id := range ids {
		out[i] = VaultKeepResult{ID: id, OK: i == 0, Code: "pinned", Why: "pinned items are never sold"}
		if i == 0 {
			out[i].Code, out[i].Why = "", ""
		}
	}
	return out, &VaultSale{ID: "s-1", State: "done", Items: 1, Done: 1, Gold: 30, StashAfter: 130, Line: "Sold 1 items for 30 gold. Stash gold: 130"}
}

func TestVaultSellRouteIsInstantWhenTheActorSaysSo(t *testing.T) {
	s := testVault(t, 6, 100)
	a := &fakeSaleActor{stubActor: &stubActor{st: s}, on: true, quote: VaultSaleQuote{Count: 1, Total: 30, Stash: 100, StashAfter: 130, ConfirmAt: 2000}}
	api := actAPI(s, a.stubActor)
	api.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Act: a})
	_, all := getVault(t, api, "/api/vault?limit=5")
	ids := []string{all.Items[0].ID, all.Items[1].ID}
	// no confirm: refused, nothing sold
	if code, _, pr := postVault(t, api, VaultSellRoute, vaultBody{IDs: ids}); code != 400 || pr.Code != "confirm" || len(a.sold) != 0 {
		t.Fatalf("confirm: %d %+v", code, pr)
	}
	code, rep, _ := postVault(t, api, VaultSellRoute, vaultBody{IDs: ids, Confirm: true, Ack: true})
	if code != 200 || len(rep.Done) != 1 || len(rep.Failed) != 1 || rep.Failed[0].Code != "pinned" || rep.Sale == nil || rep.Sale.StashAfter != 130 || !a.confrm[0] {
		t.Fatalf("instant: %d %+v", code, rep)
	}
	// the quote changes nothing and answers the sale's numbers
	var q VaultSaleQuote
	if c := postRaw(t, api, VaultSellQuote, vaultBody{IDs: ids}, &q); c != 200 || q.Total != 30 || q.StashAfter != 130 || q.Asked != 2 || len(a.sold) != 1 {
		t.Fatalf("quote: %d %+v sold %v", c, q, a.sold)
	}
	// a bot that says no (older d2d, knob 0): the old queue answers and the quote is refused
	a.on = false
	if code, rep, _ := postVault(t, api, VaultSellRoute, vaultBody{IDs: []string{all.Items[2].ID}, Confirm: true}); code != 200 || len(rep.Done) != 1 || rep.Sale != nil || len(a.sold) != 1 {
		t.Fatalf("queued: %d %+v", code, rep)
	}
}
