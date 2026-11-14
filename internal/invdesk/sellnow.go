package invdesk

import (
	"context"
	"fmt"

	"d2lod-bot/internal/d2"
)

// Sell now (API schema 10, docs/inventory-actions.md): a town errand that names no item. It makes the bot walk home at
// the next safe moment (errandTripStep) and run the sale of the items the player marked "sell"; nothing else is sold, so
// there is no new decision about what to sell. The request ends when the sale step ran.
const KindSell = "sell"

// NothingMarked is the reason of a sell-now request while no backpack item is marked for sale.
const NothingMarked = "nothing is marked for sale"

// Seller is the town's sale of the marked items (*town.Town through the adapter in cmd/d2bot). It runs in town and may
// return a "fight is on" error (isBusy) or a cancelled context to postpone the request.
type Seller interface {
	// SellMarked sells the marked items at the act's vendors and returns a sentence such as "sold 3 of 4 marked". It
	// returns an error when nothing could be sold.
	SellMarked(ctx context.Context) (string, error)
}

// SetSeller installs the sale step; until then sell now is refused with code "off".
func (d *Desk) SetSeller(s Seller) { d.mu.Lock(); d.sell = s; d.mu.Unlock() }

// markedForSale counts the backpack items marked sell that may be sold. Callbacks run unlocked (see SetMark).
func (d *Desk) markedForSale(s *d2.Snapshot) int {
	if s == nil || s.Inventory == nil {
		return 0
	}
	d.mu.Lock()
	block := d.c.SellBlock
	d.mu.Unlock()
	n := 0
	for _, u := range s.Inventory.Container(d2.ContInventory) {
		if u.Item != nil && d.MarkOf(u) == MarkSell && (block == nil || block(u) == "") {
			n++
		}
	}
	return n
}

// RequestSell queues the sell-now errand. Refused when nothing is marked (code "nothing"), when one is waiting already
// and when the bot has no sale step.
func (d *Desk) RequestSell() (Request, error) {
	d.mu.Lock()
	sl := d.sell
	d.mu.Unlock()
	if sl == nil {
		return Request{}, refuse("off", "selling on request is not available on this bot")
	}
	s := d.snap()
	if s == nil || s.Player == nil || s.Inventory == nil {
		return Request{}, refuse("waiting", "waiting for the first look at the hero")
	}
	n := d.markedForSale(s)
	if n == 0 {
		return Request{}, refuse("nothing", NothingMarked)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.queuedLocked(0, KindSell) {
		return Request{}, refuse("queued", "a sale is already waiting: cancel it first")
	}
	return d.enqueueLocked(Request{Kind: KindSell, Name: fmt.Sprintf("Sell marked items (%d)", n), Count: n})
}

// runSell performs the sale; the caller settles the request. It re-checks the marks: the player may have cleared them
// while the request waited for the trip home.
func (d *Desk) runSell(ctx context.Context, sl Seller) (string, error) {
	if sl == nil {
		return "", fmt.Errorf("selling is not available")
	}
	if d.markedForSale(d.snap()) == 0 {
		return "", fmt.Errorf("%s", NothingMarked)
	}
	return sl.SellMarked(ctx)
}
