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
	"d2lod-bot/internal/vault"
)

// Reveal paid in gold (docs/vault.md, "Reveal paid in gold"). A gold reveal is an ORDER: the bot buys Scrolls of Identify
// at the vendor of the town it visits (real gold, above the repair and potion-kit reserve) and reads them on the vault
// items through the ordinary Reveal flow in the same visit. No virtual gold exists, so the price is the vendor price of
// the scrolls that are bought and used up.
//
//	queued -> (town visit) buying -> revealing -> done | partial | cancelled

// Order states.
const (
	OrderQueued    = "queued"    // waits for a town visit
	OrderBuying    = "buying"    // the visit is buying scrolls
	OrderRevealing = "revealing" // scrolls are in the pack, the items are being revealed
	OrderDone      = "done"
	OrderPartial   = "partial" // ended with items left: gold ran short, the cap was reached or an item failed
	OrderCancelled = "cancelled"
)

// ScrollShop is the vendor side of the order: the town's buy-identify errand and the purse.
type ScrollShop interface {
	// Price is the vendor price of one Scroll of Identify, 0 when unknown.
	Price() int
	// Purse reads the pack gold, the stash gold and the gold the town guards keep back (repair, potion kit, merc).
	Purse(ctx context.Context) (purse, stash, reserve int, err error)
	// Buy buys up to n scrolls with gold above the reserve (taking gold out of the stash when the pack is short). It
	// returns without error when it could buy fewer; the caller reads the pack to see how many.
	Buy(ctx context.Context, n int) error
}

// GoldOrder is the one gold reveal order of a hero. It is stored next to the vault log so a restart keeps it.
type GoldOrder struct {
	ID     string   `json:"id"`
	IDs    []string `json:"ids"`             // the records the player named, reveal order
	Max    int      `json:"max"`             // gold the order may spend
	Price  int      `json:"price"`           // gold of one scroll when ordered
	State  string   `json:"state"`           // Order*
	Asked  []string `json:"asked,omitempty"` // handed to the Reveal queue
	Done   int      `json:"done"`            // revealed
	Failed int      `json:"failed"`          // asked and not revealed (reason on the record)
	Bought int      `json:"bought"`          // scrolls bought for this order
	Spent  int      `json:"spent"`           // gold spent on them
	Note   string   `json:"note,omitempty"`  // why the order ended early, in plain words
	Ended  int64    `json:"ended,omitempty"` // unix time the order ended; the page forgets it after goldShown
	Stop   bool     `json:"stop,omitempty"`  // the rest cannot be paid: finish what is queued, then end as partial
}

// goldShown is how long the page keeps showing a finished order.
const goldShown = 30 * time.Minute

// Active reports whether the order still moves.
func (o *GoldOrder) Active() bool {
	return o != nil && (o.State == OrderQueued || o.State == OrderBuying || o.State == OrderRevealing)
}

// GoldView is the order as the page shows it.
type GoldView struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Total  int    `json:"total"`
	Done   int    `json:"done"`
	Failed int    `json:"failed"`
	Bought int    `json:"bought"`
	Spent  int    `json:"spent"`
	Max    int    `json:"max"`
	Note   string `json:"note,omitempty"`
	Line   string `json:"line"`
}

// Reasons of the gold order.
const (
	ReasonGoldOff    = "reveal paid in gold is off (loot.vaultRevealGold is 0)"
	ReasonGoldNoShop = "this bot cannot buy scrolls"
	ReasonGoldShort  = "short of gold: the repair and potion kit come first"
	ReasonGoldCap    = "the gold cap of the order is spent"
)

func (k *Keeper) goldOn() bool { return k.Shop != nil && k.GoldOn != nil && k.GoldOn() }

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

// saveOrder writes the order through a temp file so a crash never leaves half a file. A failed write is returned: the
// order then lives on in memory only and the caller says so.
func (k *Keeper) saveOrder(o *GoldOrder) error {
	k.mem = o
	path := k.orderPath()
	if path == "" {
		return nil
	}
	k.loaded = path
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

func sameSet(a, b []string) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	sort.Strings(x)
	sort.Strings(y)
	return slices.Equal(x, y)
}

// GoldQuote is what the order would cost right now: the gold of the scrolls the hero lacks for n items.
func (k *Keeper) GoldQuote(ctx context.Context, n int) (cost, price int) {
	if k.Shop == nil {
		return 0, 0
	}
	price = k.Shop.Price()
	return max(n-max(k.charges(ctx), 0), 0) * price, price
}

// EnqueueRevealGold creates the gold order for ids. Idempotent: the same ids while the order runs answer Already; a
// different set while one runs is refused with code "active". maxGold caps what the bot may spend (0: the quote). A cost at
// or over the confirm threshold needs confirm. Refusals have a code: off, noshop, state, missing, active, cap, confirm.
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
	case !k.goldOn() && k.Shop == nil:
		return fail("noshop", ReasonGoldNoShop)
	case !k.goldOn():
		return fail("off", ReasonGoldOff)
	case st == nil:
		return fail("missing", "there is no vault for this hero yet")
	case len(ids) == 0 || len(ids) > bulkid.GoldMax:
		return fail("ids", fmt.Sprintf("name 1 to %d items", bulkid.GoldMax))
	}
	if o := k.order(); o.Active() {
		if sameSet(o.IDs, ids) {
			out := make([]Result, len(ids))
			for i, id := range ids {
				out[i] = Result{ID: id, OK: true, Already: true}
			}
			v := k.goldView(o)
			return out, &v
		}
		return fail("active", "a gold reveal is already running: cancel it first")
	}
	res = make([]Result, len(ids))
	var ok []string
	for i, id := range ids {
		res[i] = Result{ID: id}
		rec, found := st.Get(id)
		switch {
		case !found:
			res[i].Code, res[i].Why = "missing", "no such item"
		case rec.Identified:
			res[i].Code, res[i].Why = "state", "it is identified already"
		case rec.State != vault.Recorded && rec.State != vault.KeepQ:
			res[i].Code, res[i].Why = "state", "only an item in the vault can be revealed (it is "+string(rec.State)+")"
		case rec.RevQ:
			res[i].Code, res[i].Why = "queued", "a reveal is queued for it already"
		default:
			ok = append(ok, id)
		}
	}
	if len(ok) == 0 {
		return res, nil
	}
	cost, price := k.GoldQuote(ctx, len(ok))
	if price <= 0 {
		return fail("noshop", "the scroll price is unknown")
	}
	if t := k.goldConfirm(); t > 0 && cost >= t && !confirm {
		return fail("confirm", fmt.Sprintf("this spends up to %d gold on scrolls: confirm it", cost))
	}
	if maxGold <= 0 {
		maxGold = cost
	}
	if maxGold < price && cost > 0 {
		return fail("cap", fmt.Sprintf("the cap %d is below one scroll (%d)", maxGold, price))
	}
	o := &GoldOrder{ID: fmt.Sprintf("g-%d", k.now().UnixNano()), IDs: ok, Max: maxGold, Price: price, State: OrderQueued}
	k.omu.Lock()
	err := k.saveOrder(o)
	k.omu.Unlock()
	if err != nil {
		k.note("goldorder", "The gold reveal order could not be saved ("+err.Error()+"): it is kept until the bot restarts.")
	}
	for i := range res {
		if res[i].Code == "" {
			res[i].OK = true
		}
	}
	v := k.goldView(o)
	return res, &v
}

func (k *Keeper) goldConfirm() int {
	if k.GoldConfirm == nil {
		return 0
	}
	return k.GoldConfirm()
}

// CancelGold ends a running order: reveals not yet started are taken back, scrolls already bought stay in the pack.
func (k *Keeper) CancelGold() (GoldView, bool) {
	k.omu.Lock()
	defer k.omu.Unlock()
	o := k.orderLocked()
	if !o.Active() {
		if o == nil {
			return GoldView{}, false
		}
		return k.goldView(o), false
	}
	if st := k.store(); st != nil {
		for _, id := range o.Asked {
			if r, ok := st.Get(id); ok && r.RevQ && r.Stage == "" {
				_ = st.RevealCancel(id)
			}
		}
	}
	k.tally(o)
	o.State, o.Ended = OrderCancelled, k.now().Unix()
	o.Note = fmt.Sprintf("cancelled; %d revealed", o.Done)
	if o.Bought > 0 {
		o.Note += fmt.Sprintf(", %d bought scroll%s stay in the pack", o.Bought, plural(o.Bought))
	}
	_ = k.saveOrder(o)
	return k.goldView(o), true
}

// GoldView reads the order for the page (nil when none).
func (k *Keeper) GoldView() *GoldView {
	k.omu.Lock()
	defer k.omu.Unlock()
	o := k.orderLocked()
	if o == nil {
		return nil
	}
	if o.Active() {
		k.tally(o)
	} else if o.Ended > 0 && k.now().Sub(time.Unix(o.Ended, 0)) > goldShown {
		return nil
	}
	v := k.goldView(o)
	return &v
}

// goldView words the state; the caller holds omu or owns o.
func (k *Keeper) goldView(o *GoldOrder) GoldView {
	v := GoldView{ID: o.ID, State: o.State, Total: len(o.IDs), Done: o.Done, Failed: o.Failed, Bought: o.Bought, Spent: o.Spent, Max: o.Max, Note: o.Note}
	switch o.State {
	case OrderQueued:
		v.Line = fmt.Sprintf("Gold reveal of %d item%s: queued, the bot buys the scrolls at your next town visit", v.Total, plural(v.Total))
	case OrderBuying:
		v.Line = fmt.Sprintf("Gold reveal: buying scrolls (%d of %d revealed)", o.Done, v.Total)
	case OrderRevealing:
		v.Line = fmt.Sprintf("Gold reveal: revealing, %d of %d done, %d gold spent", o.Done, v.Total, o.Spent)
	case OrderDone:
		v.Line = fmt.Sprintf("Gold reveal done: %d item%s revealed for %d gold", o.Done, plural(o.Done), o.Spent)
	default:
		v.Line = fmt.Sprintf("Gold reveal %s: %d of %d revealed, %d gold spent", o.State, o.Done, v.Total, o.Spent)
		if o.Note != "" {
			v.Line += ". " + o.Note
		}
	}
	return v
}

// tally recounts done and failed from the records the order asked for.
func (k *Keeper) tally(o *GoldOrder) {
	st := k.store()
	if st == nil {
		return
	}
	o.Done, o.Failed = 0, 0
	for _, id := range o.Asked {
		r, ok := st.Get(id)
		switch {
		case !ok:
			o.Failed++
		case r.Identified:
			o.Done++
		case !r.RevQ && r.Stage == "":
			o.Failed++
		}
	}
}

// pendingIDs are the order's records that are not yet handed to the Reveal queue and can still be revealed.
func (k *Keeper) goldPending(o *GoldOrder, st *vault.Store) []string {
	asked := map[string]bool{}
	for _, id := range o.Asked {
		asked[id] = true
	}
	var out []string
	for _, id := range o.IDs {
		if asked[id] {
			continue
		}
		if r, ok := st.Get(id); ok && !r.Identified && !r.RevQ && (r.State == vault.Recorded || r.State == vault.KeepQ) {
			out = append(out, id)
		}
	}
	return out
}

// GoldPending is true while an order needs the bot in town (the bot step then runs the keeper).
func (k *Keeper) GoldPending() bool {
	o := k.order()
	return o.Active() && k.goldOn()
}

// runGold (settle false) is the order's step of a town visit: buy what is missing, hand the items to the Reveal queue (runReveal does
// the rest in the same Run). It never spends below the town's reserve and never above the order's cap.
func (k *Keeper) runGold(ctx context.Context, settle bool) []string {
	k.omu.Lock()
	defer k.omu.Unlock()
	o := k.orderLocked()
	st := k.store()
	if !o.Active() || st == nil || !k.goldOn() {
		return nil
	}
	k.tally(o)
	pending := k.goldPending(o, st)
	if o.Stop {
		pending = nil
	}
	queued := 0 // asked and still waiting or working
	for _, id := range o.Asked {
		if r, ok := st.Get(id); ok && !r.Identified && (r.RevQ || r.Stage != "") {
			queued++
		}
	}
	finish := func(state, note string) []string {
		k.tally(o)
		o.State, o.Note, o.Ended = state, note, k.now().Unix()
		if err := k.saveOrder(o); err != nil {
			return []string{"gold reveal: order not saved: " + err.Error()}
		}
		return []string{"gold reveal " + state + ": " + k.goldView(o).Line}
	}
	if len(pending) == 0 && queued == 0 {
		switch {
		case o.Note != "":
			return finish(OrderPartial, o.Note)
		case o.Failed > 0:
			return finish(OrderPartial, fmt.Sprintf("%d item%s could not be revealed (the reason is on the item)", o.Failed, plural(o.Failed)))
		case o.Done < len(o.IDs):
			return finish(OrderPartial, fmt.Sprintf("%d item%s left the vault or were revealed another way", len(o.IDs)-o.Done, plural(len(o.IDs)-o.Done)))
		}
		return finish(OrderDone, "")
	}
	if settle { // after the visit's reveals: only close an order that is finished, buy and ask for nothing more
		return nil
	}
	if len(pending) == 0 {
		o.State = OrderRevealing
		_ = k.saveOrder(o)
		return nil
	}
	if !k.runGate(ctx) {
		return nil
	}
	free := k.free()
	batch := min(len(pending), k.perVisit())
	var lines []string
	n := batch
	if !free {
		held := max(k.charges(ctx), 0)
		if need := batch - held; need > 0 {
			o.State = OrderBuying
			_ = k.saveOrder(o)
			price := max(k.Shop.Price(), o.Price)
			afford := 0
			if price > 0 {
				afford = max(o.Max-o.Spent, 0) / price
			}
			buy := min(need, afford)
			if buy > 0 {
				p0, s0, _, perr := k.Shop.Purse(ctx)
				c0 := max(k.charges(ctx), 0)
				err := k.Shop.Buy(ctx, buy)
				p1, s1, _, perr2 := k.Shop.Purse(ctx)
				got := max(k.charges(ctx), 0) - c0
				if perr == nil && perr2 == nil {
					o.Spent += max(p0+s0-p1-s1, 0)
				} else {
					o.Spent += max(got, 0) * price
				}
				o.Bought += max(got, 0)
				switch {
				case errors.Is(err, ErrPostpone) || ctx.Err() != nil:
					_ = k.saveOrder(o)
					return append(lines, "gold reveal: scroll shopping postponed: "+fmt.Sprint(err))
				case err != nil && got == 0:
					lines = append(lines, "gold reveal: no scrolls bought: "+err.Error())
				}
				lines = append(lines, fmt.Sprintf("gold reveal: bought %d of %d scrolls for %d gold", got, buy, max(p0+s0-p1-s1, 0)))
			}
			held = max(k.charges(ctx), 0)
			n = min(batch, held)
			if n < batch {
				note := ReasonGoldShort
				if buy < need && afford < need {
					note = ReasonGoldCap
				}
				if n == 0 && queued == 0 {
					return append(lines, finish(OrderPartial, note)...)
				}
				// the rest of the order cannot be paid: reveal what the scrolls cover, then end as partial
				o.Note, o.Stop = note, true
			}
		}
	}
	for _, id := range pending[:n] {
		if err := st.RevealWant(id); err != nil {
			lines = append(lines, "gold reveal: "+id+": "+err.Error())
			continue
		}
		o.Asked = append(o.Asked, id)
	}
	o.State = OrderRevealing
	if err := k.saveOrder(o); err != nil {
		lines = append(lines, "gold reveal: order not saved: "+err.Error())
	}
	return lines
}

// GoldFacts is what the page's bar reads of the shop: whether gold orders are on, the scroll price, the purse and the
// reserve, and the confirm threshold.
type GoldFacts struct {
	On                                    bool
	Price, Purse, Stash, Reserve, Confirm int
}

// GoldFacts reads them; an unreadable purse counts as empty.
func (k *Keeper) GoldFacts(ctx context.Context) GoldFacts {
	f := GoldFacts{On: k.goldOn(), Confirm: k.goldConfirm()}
	if k.Shop == nil {
		return f
	}
	f.Price = k.Shop.Price()
	if p, s, r, err := k.Shop.Purse(ctx); err == nil {
		f.Purse, f.Stash, f.Reserve = p, s, r
	}
	return f
}
