package d2dinput

import "fmt"

// The fork's second set of input verbs (docs/process/input-verbs2.md) as input.WorldVerbs. Nothing here decides whether
// to use them: callers ask the planner (input.planner, the backend's caps) and keep their legacy path as the fallback.

// verbOn says whether the planner is on and the connected d2d lists the capability.
func (c *Controller) verbOn(capName string) bool {
	return c.Planner.On() && c.Planner.HasCap(capName)
}

// BuyAt buys the stock item at cell (x, y) of vendor tab tab. The stock index is the item's place in its tab of "dump
// vendor"; an item the dump does not list is a plain error (the caller clicks instead).
func (c *Controller) BuyAt(tab, x, y int) error {
	ctx := c.ctx()
	v, err := c.C.Vendor(ctx)
	if err != nil {
		return err
	}
	idx, ok := v.StockIndex(tab, x, y)
	if !ok {
		return fmt.Errorf("d2dinput: no stock item at tab %d cell %d,%d", tab, x, y)
	}
	_, err = c.C.VerbBuy(ctx, idx, tab, false)
	return err
}

// SellItem sells carried item id to the open vendor window.
func (c *Controller) SellItem(id uint32) error {
	_, err := c.C.VerbSell(c.ctx(), id)
	return err
}

// TakeWaypoint travels from the open waypoint panel to the waypoint of area level.
func (c *Controller) TakeWaypoint(level int) error {
	_, err := c.C.VerbWaypoint(c.ctx(), level, -1)
	return err
}

// EnterPortal walks to town portal which and goes through.
func (c *Controller) EnterPortal(which int) error {
	_, err := c.C.VerbPortal(c.ctx(), which)
	return err
}

// PickupItem picks up ground item id; the engine walks to it.
func (c *Controller) PickupItem(id uint32) error {
	_, err := c.C.VerbPickup(c.ctx(), id, 0)
	return err
}
