package watch

import (
	"context"
	"net/http"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/vault"
)

// Reveal paid in gold (docs/vault.md, "Reveal paid in gold"): the order routes and the Reveal bar the page draws.

const (
	VaultRevealGoldRoute       = "/api/vault/revealgold"        // POST {ids, maxGold, confirm}
	VaultRevealGoldCancelRoute = "/api/vault/revealgold/cancel" // POST {}
)

// VaultGoldOrder is the gold order as the page shows it (cmd/d2bot copies internal/vaultkeep.GoldView).
type VaultGoldOrder struct {
	ID     string `json:"id"`
	State  string `json:"state"` // queued, buying, revealing, done, partial, cancelled
	Total  int    `json:"total"`
	Done   int    `json:"done"`
	Failed int    `json:"failed"`
	Bought int    `json:"bought"`
	Spent  int    `json:"spent"`
	Max    int    `json:"max"`
	Note   string `json:"note,omitempty"`
	Line   string `json:"line"`
}

// VaultGoldInfo is what the bar needs of the hero's purse and the order, read per request.
type VaultGoldInfo struct {
	On         bool            `json:"on"` // loot.vaultRevealGold with a shop wired
	ScrollGold int             `json:"scrollGold"`
	Purse      int             `json:"purse"`
	Stash      int             `json:"stash"`
	Reserve    int             `json:"reserve"`
	ConfirmAt  int             `json:"confirmAt"`
	Order      *VaultGoldOrder `json:"order,omitempty"`
}

// VaultGoldActor is the optional part of a VaultActor that runs gold orders.
type VaultGoldActor interface {
	GoldInfo(ctx context.Context) VaultGoldInfo
	EnqueueRevealGold(ctx context.Context, ids []string, maxGold int, confirm bool) ([]VaultKeepResult, *VaultGoldOrder)
	CancelGold() (VaultGoldOrder, bool)
}

// bar quotes the Reveal bar for the filtered records.
func (d *VaultDesk) bar(ctx context.Context, recs []vault.Rec, r VaultRevealView) *bulkid.Bar {
	in := bulkid.BarIn{Free: r.Free, Charges: r.Charges, EtherPer: r.EtherPer, EtherBal: r.Balance}
	if ga, ok := d.Act.(VaultGoldActor); ok {
		gi := ga.GoldInfo(ctx)
		in.GoldOn, in.ScrollGold, in.Purse, in.Stash, in.Reserve, in.ConfirmAt = gi.On, gi.ScrollGold, gi.Purse, gi.Stash, gi.Reserve, gi.ConfirmAt
		if gi.Order != nil && (gi.Order.State == "queued" || gi.Order.State == "buying" || gi.Order.State == "revealing") {
			in.ActiveGoldOrderID = gi.Order.ID
		}
	}
	b := bulkid.MakeBar(bulkid.Candidates(d.MassFilter, recs, d.Quality), in)
	return &b
}

// goldRoute serves the order and its cancel. The gate is the same as Reveal's (the bot must be able to identify at all).
func (a *VaultAPI) goldRoute(w http.ResponseWriter, r *http.Request, d *VaultDesk, route string, b *vaultBody) {
	ga, ok := d.Act.(VaultGoldActor)
	if d.Act == nil || !ok {
		writeJSON(w, http.StatusConflict, policyReply{Code: "off", Error: "this bot cannot buy scrolls for a reveal"})
		return
	}
	type reply struct {
		vaultReply
		Order *VaultGoldOrder `json:"order,omitempty"`
	}
	if route == VaultRevealGoldCancelRoute {
		o, did := ga.CancelGold()
		rep := reply{vaultReply: vaultReply{OK: did, Done: []string{}}}
		if o.ID != "" {
			rep.Order = &o
		}
		code := 200
		if !did {
			code = http.StatusConflict
		}
		writeJSON(w, code, rep)
		return
	}
	if ok, code, why := d.Act.GateReveal(r.Context()); !ok {
		writeJSON(w, http.StatusConflict, policyReply{Code: code, Error: why})
		return
	}
	res, order := ga.EnqueueRevealGold(r.Context(), b.IDs, b.MaxGold, b.Confirm)
	rep := reply{vaultReply: vaultReply{Done: []string{}}, Order: order}
	for _, x := range res {
		switch {
		case x.OK && x.Already:
			rep.Already = append(rep.Already, x.ID)
			rep.Done = append(rep.Done, x.ID)
		case x.OK:
			rep.Done = append(rep.Done, x.ID)
		default:
			rep.Failed = append(rep.Failed, vaultFail{ID: x.ID, Code: x.Code, Error: x.Why})
		}
	}
	rep.OK = len(rep.Failed) == 0
	code := 200
	if len(rep.Done) == 0 {
		code = http.StatusConflict
	}
	writeJSON(w, code, rep)
}
