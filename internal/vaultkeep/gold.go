package vaultkeep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// Reveal paid in gold (docs/vault.md, "Reveal paid in gold"; ADR 0029 amending ADR 0014). A gold reveal is an ORDER the
// player gives through the vault API: the price of each item (the vendor price of one Scroll of Identify) is subtracted from
// the STASH gold at once by the fork's `debug paygold` verb (internal/d2dctl Pay: ledgered, once per order id, never the
// purse, refused in a measured run), and the items are unlocked from the affixes the vault already holds, exactly as an
// Ether reveal does. No town visit, no scroll, no tome charge, no game step. The bot never starts one by itself.
//
//	queued -> paid -> revealed | partial | failed | cancelled     (seconds)
//
// The order is saved before the payment so a restart or a lost reply is settled again with the same order id (the fork
// answers a repeated id with paid 0 and changes nothing, so a gold reveal is never charged twice).

// Order states.
const (
	OrderQueued    = "queued" // items unlocked, payment not yet confirmed
	OrderPaid      = "paid"   // the fork took the gold
	OrderRevealed  = "revealed"
	OrderPartial   = "partial" // some of the player's items were revealed, the stash paid for no more
	OrderFailed    = "failed"  // nothing was paid and nothing stays unlocked; Note says why
	OrderCancelled = "cancelled"
)

// GoldPayer is the payment side of the order: the reveal price, the stash gold and the fork's paygold verb.
type GoldPayer interface {
	// Price is the gold one reveal costs: the vendor price of one Scroll of Identify, 0 when unknown.
	Price() int
	// Stash reads the stash gold (a fresh read; the order never pays from the purse).
	Stash(ctx context.Context) (int, error)
	// PayGold takes amount out of the stash gold once per order id (d2dctl.Pay in the bot, d2dctl.PayFake in tests).
	PayGold(ctx context.Context, amount int, order string) (d2dctl.PayResult, error)
}

// GoldOrder is the one gold reveal order of a hero. It is stored next to the vault log so a restart keeps it.
type GoldOrder struct {
	ID         string   `json:"id"`
	IDs        []string `json:"ids"`                  // the records the stash pays for, the player's list order
	Want       int      `json:"want"`                 // the records the player asked for that could be revealed
	Price      int      `json:"price"`                // gold of one reveal when ordered
	State      string   `json:"state"`                // Order*
	Done       int      `json:"done"`                 // revealed
	Paid       int      `json:"paid"`                 // gold taken from the stash
	StashAfter int      `json:"stashAfter,omitempty"` // the stash gold the fork answered after the payment
	Note       string   `json:"note,omitempty"`       // why the order ended early, in plain words
	Ended      int64    `json:"ended,omitempty"`      // unix time the order ended; the page forgets it after goldShown
}

// goldShown is how long the page keeps showing a finished order.
const goldShown = 30 * time.Minute

// Active reports whether the order still moves (a payment is not settled yet).
func (o *GoldOrder) Active() bool {
	return o != nil && (o.State == OrderQueued || o.State == OrderPaid)
}

// GoldView is the order as the page shows it.
type GoldView struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	Total      int    `json:"total"` // items the stash pays for
	Want       int    `json:"want"`  // items the player asked for
	Done       int    `json:"done"`
	Paid       int    `json:"paid"`
	StashAfter int    `json:"stashAfter"` // stash gold after the payment (0 until paid)
	Note       string `json:"note,omitempty"`
	Line       string `json:"line"`
}

// Reasons of the gold order.
const (
	ReasonGoldOff      = "reveal paid in gold is off (loot.vaultRevealGold is 0)"
	ReasonGoldNoPayer  = "this bot cannot pay a reveal in gold"
	ReasonGoldMeasured = "a reveal paid in gold is never done in a measured run"
	ReasonGoldNoFlag   = "this d2d build cannot take the gold (start it with --vault-pay)"
)

func (k *Keeper) goldOn() bool { return k.Payer != nil && k.GoldOn != nil && k.GoldOn() }

func (k *Keeper) measured() bool { return k.Measured != nil && k.Measured() }

// orderPath is the file of the hero's order ("" when no place is wired: the order lives in memory).
func (k *Keeper) orderPath() string {
	if k.OrderFile == nil {
		return ""
	}
	return k.OrderFile()
}

// order returns the stored order (nil when none). Callers hold no lock.
func (k *Keeper) order() *GoldOrder {
	k.omu.Lock()
	defer k.omu.Unlock()
	return k.orderLocked()
}

func (k *Keeper) orderLocked() *GoldOrder {
	path := k.orderPath()
	if path == "" {
		return k.mem
	}
	if k.loaded != path {
		k.mem, k.loaded = nil, path
		if b, err := os.ReadFile(path); err == nil {
			var o GoldOrder
			if json.Unmarshal(b, &o) == nil && o.ID != "" {
				k.mem = &o
			}
		}
	}
	return k.mem
}

// putOrder stores a copy of o and writes it through a temp file so a crash never leaves half a file. A failed write is
// returned: the order then lives on in memory only and the caller says so.
func (k *Keeper) putOrder(o GoldOrder) error {
	k.omu.Lock()
	defer k.omu.Unlock()
	k.orderLocked() // load the file of this hero first
	k.mem = &o
	path := k.orderPath()
	if path == "" {
		return nil
	}
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (k *Keeper) saveOrNote(o GoldOrder) {
	if err := k.putOrder(o); err != nil {
		k.note("goldorder", "The gold reveal order could not be saved ("+err.Error()+"): it is kept until the bot restarts.")
	}
}

func sameSet(a, b []string) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	sort.Strings(x)
	sort.Strings(y)
	return slices.Equal(x, y)
}

func goldShort(price, stash int) string {
	return fmt.Sprintf("short of gold: one reveal costs %s gold, the stash holds %s", bulkid.Thousands(price), bulkid.Thousands(stash))
}

// EnqueueRevealGold gives the gold order for ids and settles it before it returns: the stash pays for as many of the items
// as it can, in the order the player named them, and the paid items are revealed in the same call. maxGold caps what the
// order may cost (0: no cap). A cost at or over the confirm threshold needs confirm. Refusals have a code: off, noshop,
// measured, state, missing, nosealed, active, gold (the stash is short or the cap too low), confirm, payoff, pay.
func (k *Keeper) EnqueueRevealGold(ctx context.Context, ids []string, maxGold int, confirm bool) (res []Result, view *GoldView) {
	fail := func(code, why string) ([]Result, *GoldView) {
		out := make([]Result, len(ids))
		for i, id := range ids {
			out[i] = Result{ID: id, Code: code, Why: why}
		}
		return out, nil
	}
	st := k.store()
	switch {
	case k.Payer == nil:
		return fail("noshop", ReasonGoldNoPayer)
	case !k.goldOn():
		return fail("off", ReasonGoldOff)
	case k.measured():
		return fail("measured", ReasonGoldMeasured)
	case st == nil:
		return fail("missing", "there is no vault for this hero yet")
	case len(ids) == 0 || len(ids) > bulkid.GoldMax:
		return fail("ids", fmt.Sprintf("name 1 to %d items", bulkid.GoldMax))
	}
	k.pmu.Lock()
	defer k.pmu.Unlock()
	if o := k.order(); o.Active() {
		if !sameSet(o.IDs, ids) {
			return fail("active", "a gold reveal is still being paid: wait a moment")
		}
		lines := k.settleGold(ctx, *o) // a repeated press settles it again; the same order id is never charged twice
		_ = lines
		out := make([]Result, len(ids))
		for i, id := range ids {
			out[i] = Result{ID: id, OK: true, Already: true}
		}
		v := k.goldView(k.order())
		return out, &v
	}
	res = make([]Result, len(ids))
	var ok []string
	at := map[string]int{}
	for i, id := range ids {
		res[i] = Result{ID: id}
		rec, found := st.Get(id)
		switch {
		case !found:
			res[i].Code, res[i].Why = "missing", "no such item"
		case rec.Unlocked: // a repeated press after a paid reveal: nothing changes, nothing is charged
			res[i].OK, res[i].Already = true, true
		case rec.Identified:
			res[i].Code, res[i].Why = "state", "it is identified already"
		case rec.State != vault.Recorded && rec.State != vault.KeepQ:
			res[i].Code, res[i].Why = "state", "only an item in the vault can be revealed (it is "+string(rec.State)+")"
		case rec.Sealed == "":
			res[i].Code, res[i].Why = "nosealed", ReasonNoSealed
		default:
			ok = append(ok, id)
			at[id] = i
		}
	}
	if len(ok) == 0 {
		return res, nil
	}
	price := k.Payer.Price()
	if price <= 0 {
		for _, id := range ok {
			res[at[id]].Code, res[at[id]].Why = "noshop", "the reveal price is unknown"
		}
		return res, nil
	}
	stash, err := k.Payer.Stash(ctx)
	if err != nil {
		for _, id := range ok {
			res[at[id]].Code, res[at[id]].Why = "state", "the stash gold could not be read: "+err.Error()
		}
		return res, nil
	}
	n := min(len(ok), stash/price)
	why := goldShort(price, stash)
	if maxGold > 0 && maxGold/price < n {
		was := n
		n = maxGold / price
		why = fmt.Sprintf("the gold cap %s pays for %d of %d items", bulkid.Thousands(maxGold), n, was)
		if n == 0 {
			why = fmt.Sprintf("the gold cap %s is below one reveal (%s)", bulkid.Thousands(maxGold), bulkid.Thousands(price))
		}
	}
	if n == 0 {
		for _, id := range ok {
			res[at[id]].Code, res[at[id]].Why = "gold", why
		}
		return res, nil
	}
	if t := k.goldConfirm(); t > 0 && n*price >= t && !confirm {
		for _, id := range ok {
			res[at[id]].Code, res[at[id]].Why = "confirm", fmt.Sprintf("this pays %s gold from your stash: confirm it", bulkid.Thousands(n*price))
		}
		return res, nil
	}
	o := GoldOrder{ID: fmt.Sprintf("g-%d", k.now().UnixNano()), IDs: slices.Clone(ok[:n]), Want: len(ok), Price: price, State: OrderQueued}
	k.saveOrNote(o)
	k.settleGold(ctx, o)
	final := k.order()
	for _, id := range ok[:n] {
		r := &res[at[id]]
		if final != nil && slices.Contains(final.IDs, id) && (final.State == OrderRevealed || final.State == OrderPartial) {
			r.OK = true
			continue
		}
		r.Code, r.Why = "pay", "the payment did not go through"
		if final != nil && final.Note != "" {
			r.Why = final.Note
		}
		if final != nil && final.State == OrderQueued {
			r.Why = "the payment is not confirmed yet; it is settled again at the next step"
		}
	}
	for _, id := range ok[n:] {
		res[at[id]].Code, res[at[id]].Why = "gold", why
	}
	if final != nil && final.State == OrderFailed {
		for _, id := range ok[:n] {
			res[at[id]].Code = payCode(final.Note)
		}
	}
	if final != nil {
		v := k.goldView(final)
		return res, &v
	}
	return res, nil
}

func payCode(note string) string {
	switch note {
	case ReasonGoldMeasured:
		return "measured"
	case ReasonGoldNoFlag:
		return "payoff"
	}
	return "pay"
}

// settleGold runs an order from where it stands: unlock the items from the held affixes, take the gold, mark them revealed.
// The caller holds pmu. A refusal (stash short, measured run, flag off) takes the unlock back and ends the order failed; an
// unanswered request leaves it queued so the next settle repeats it with the same order id.
func (k *Keeper) settleGold(ctx context.Context, o GoldOrder) []string {
	st := k.store()
	if st == nil || !o.Active() {
		return nil
	}
	var lines []string
	end := func(state, note string) []string {
		o.State, o.Note, o.Ended = state, note, k.now().Unix()
		k.saveOrNote(o)
		return append(lines, "gold reveal "+state+": "+k.goldView(&o).Line)
	}
	if o.State == OrderQueued {
		var kept []string
		for _, id := range o.IDs {
			rec, found := st.Get(id)
			switch {
			case !found:
			case rec.Unlocked: // unlocked by an earlier try of this order
				kept = append(kept, id)
			default:
				if err := st.Unseal(id); err == nil {
					kept = append(kept, id)
				}
			}
		}
		o.IDs = kept
		if len(kept) == 0 {
			return end(OrderFailed, "none of the items could be unlocked any more")
		}
		cost := len(kept) * o.Price
		if k.measured() {
			k.reseal(st, kept)
			o.IDs = nil
			return end(OrderFailed, ReasonGoldMeasured)
		}
		pr, err := k.Payer.PayGold(ctx, cost, o.ID)
		switch {
		case err == nil:
			o.Paid, o.StashAfter = cost, pr.StashAfter
			o.State = OrderPaid
			k.saveOrNote(o)
		case errors.Is(err, d2dctl.ErrPayShort), errors.Is(err, d2dctl.ErrPayMeasured), errors.Is(err, d2dctl.ErrPayOff), errors.Is(err, d2dctl.ErrPayAmount):
			k.reseal(st, kept)
			o.IDs = nil
			note := "the gold was refused: " + err.Error()
			switch {
			case errors.Is(err, d2dctl.ErrPayMeasured):
				note = ReasonGoldMeasured
			case errors.Is(err, d2dctl.ErrPayOff):
				note = ReasonGoldNoFlag
			case errors.Is(err, d2dctl.ErrPayShort):
				note = "short of gold: the stash holds less than the order costs"
			}
			return end(OrderFailed, note)
		default: // unanswered: the gold may or may not have left the stash; the same order id settles it
			k.note("goldpay", "A gold reveal is waiting for the payment to be confirmed ("+err.Error()+"); it is tried again at the next step.")
			return append(lines, "gold reveal: payment unanswered, will repeat with the same order id: "+err.Error())
		}
	}
	k.clearNote("goldpay")
	revealed := 0
	for _, id := range o.IDs {
		if rec, ok := st.Get(id); ok && rec.Identified {
			revealed++
		}
	}
	o.Done = revealed
	k.mu.Lock()
	k.info.Revealed += revealed
	k.mu.Unlock()
	if o.Done < o.Want {
		return end(OrderPartial, fmt.Sprintf("%d of the %d items you asked for were revealed: the stash paid for those", o.Done, o.Want))
	}
	return end(OrderRevealed, "")
}

func (k *Keeper) reseal(st *vault.Store, ids []string) {
	for _, id := range ids {
		_ = st.Reseal(id)
	}
}

func (k *Keeper) goldConfirm() int {
	if k.GoldConfirm == nil {
		return 0
	}
	return k.GoldConfirm()
}

// CancelGold ends an order whose payment is not confirmed yet (it takes the unlock back); a paid order cannot be undone.
func (k *Keeper) CancelGold() (GoldView, bool) {
	k.pmu.Lock()
	defer k.pmu.Unlock()
	o := k.order()
	if o == nil {
		return GoldView{}, false
	}
	if o.State != OrderQueued {
		return k.goldView(o), false
	}
	c := *o
	if st := k.store(); st != nil {
		k.reseal(st, c.IDs)
	}
	c.IDs = nil
	c.State, c.Ended, c.Note = OrderCancelled, k.now().Unix(), "cancelled before it was paid"
	k.saveOrNote(c)
	return k.goldView(&c), true
}

// GoldView reads the order for the page (nil when none).
func (k *Keeper) GoldView() *GoldView {
	k.omu.Lock()
	defer k.omu.Unlock()
	o := k.orderLocked()
	if o == nil {
		return nil
	}
	if !o.Active() && o.Ended > 0 && k.now().Sub(time.Unix(o.Ended, 0)) > goldShown {
		return nil
	}
	v := k.goldView(o)
	return &v
}

// goldView words the state; the caller owns o.
func (k *Keeper) goldView(o *GoldOrder) GoldView {
	v := GoldView{ID: o.ID, State: o.State, Total: len(o.IDs), Want: o.Want, Done: o.Done, Paid: o.Paid, StashAfter: o.StashAfter, Note: o.Note}
	n := max(len(o.IDs), o.Done)
	switch o.State {
	case OrderQueued:
		v.Line = fmt.Sprintf("Gold reveal of %d item%s: paying %s gold from the stash", n, plural(n), bulkid.Thousands(len(o.IDs)*o.Price))
	case OrderPaid:
		v.Line = fmt.Sprintf("Gold reveal: paid %s gold, revealing %d item%s", bulkid.Thousands(o.Paid), n, plural(n))
	case OrderRevealed:
		v.Line = fmt.Sprintf("Gold reveal done: %d item%s revealed for %s gold. Stash gold: %s", o.Done, plural(o.Done), bulkid.Thousands(o.Paid), bulkid.Thousands(o.StashAfter))
	case OrderPartial:
		v.Line = fmt.Sprintf("Gold reveal in part: %d of %d revealed for %s gold. Stash gold: %s. %s", o.Done, o.Want, bulkid.Thousands(o.Paid), bulkid.Thousands(o.StashAfter), o.Note)
	default:
		v.Line = "Gold reveal " + o.State
		if o.Note != "" {
			v.Line += ": " + o.Note
		}
	}
	return v
}

// GoldPending is true while an order has a payment to settle (the bot step then runs the keeper).
func (k *Keeper) GoldPending() bool {
	o := k.order()
	return o.Active() && k.goldOn()
}

// resumeGold settles an order a restart or a lost reply left half done.
func (k *Keeper) resumeGold(ctx context.Context) []string {
	k.pmu.Lock()
	defer k.pmu.Unlock()
	o := k.order()
	if !o.Active() || !k.goldOn() {
		return nil
	}
	return k.settleGold(ctx, *o)
}

// GoldFacts is what the page's bar reads of the payer: whether gold orders are on, the price of one reveal, the stash gold
// and the confirm threshold.
type GoldFacts struct {
	On                    bool
	Price, Stash, Confirm int
	Why                   string // why the bar offers no gold although it is on ("" when it can)
}

// GoldFacts reads them; an unreadable stash counts as empty.
func (k *Keeper) GoldFacts(ctx context.Context) GoldFacts {
	f := GoldFacts{On: k.goldOn() && !k.measured(), Confirm: k.goldConfirm()}
	if k.Payer == nil {
		return f
	}
	f.Price = k.Payer.Price()
	if k.API != nil {
		if caps, known, err := k.vaultCaps(ctx, false); err == nil && known && !caps.PayGold {
			f.Why = ReasonGoldNoFlag
		}
	}
	if s, err := k.Payer.Stash(ctx); err == nil {
		f.Stash = s
	}
	return f
}
