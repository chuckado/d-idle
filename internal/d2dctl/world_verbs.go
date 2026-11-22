package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// The second set of input verbs of the private fork (apps/d2d/devctl/world_verbs.hpp, docs/process/input-verbs2.md): buy,
// sell, waypoint, portal, pickup. Each is the message a client sends (0x32, 0x33, 0x49, 0x13 on a town portal, 0x16);
// the fork names a refusal's cause BEFORE it sends anything (the engine's handlers drop a bad request silently). The
// reply is one JSON line then "ok"; a refusal is "err <code> <why>". They are ordinary input: no lab flag, caps lists
// input.buy, input.sell, input.waypoint, input.portal, input.pickup. U until a sandbox image carries them
// (fork branch didle/input-verbs2-20261121).

// Capability names a fork build advertises in "caps".
const (
	CapInputBuy      = "input.buy"
	CapInputSell     = "input.sell"
	CapInputWaypoint = "input.waypoint"
	CapInputPortal   = "input.portal"
	CapInputPickup   = "input.pickup"
)

// World verb refusal codes (the cause the fork names).
const (
	VerbErrClosed    = "closed"    // no vendor window / waypoint panel is open
	VerbErrStock     = "stock"     // no such stock index
	VerbErrGold      = "gold"      // not enough gold
	VerbErrSpace     = "space"     // no room in the backpack
	VerbErrRefused   = "refused"   // the vendor refused (or does not buy it)
	VerbErrItem      = "item"      // no such item / ground item
	VerbErrBelt      = "belt"      // a belt item cannot be sold
	VerbErrCursor    = "cursor"    // the hand is not empty
	VerbErrObject    = "object"    // not a waypoint
	VerbErrLevel     = "level"     // no waypoint in that level
	VerbErrHere      = "here"      // already at that waypoint
	VerbErrInactive  = "inactive"  // that waypoint is not active yet
	VerbErrNone      = "none"      // no such town portal
	VerbErrElsewhere = "elsewhere" // the portal stands in another level
	VerbErrFar       = "far"       // pickup beyond the caller's maximum
	VerbErrDead      = "dead"
)

// VerbRefusal is a rule refusal of a world verb: the verb understood the request and the game says no, so the click
// path would fail the same way (callers do not repeat it).
type VerbRefusal struct {
	Verb, Code, Why string
}

func (e *VerbRefusal) Error() string { return e.Verb + " refused (" + e.Code + "): " + e.Why }

// RuleRefusal marks the error as a game-rule refusal.
func (e *VerbRefusal) RuleRefusal() bool { return true }

// RefusalCode is the cause the fork named.
func (e *VerbRefusal) RefusalCode() string { return e.Code }

// IsVerbRefusal reports whether err is a *VerbRefusal and returns it.
func IsVerbRefusal(err error) (*VerbRefusal, bool) {
	var r *VerbRefusal
	ok := errors.As(err, &r)
	return r, ok
}

// worldDo sends one line and returns the reply's JSON. A "closed"-style refusal is a *VerbRefusal; an unknown verb (an
// older build), "not in a game" and a joined game are plain errors, so the caller falls back to the pointer path.
func (c *Client) worldDo(ctx context.Context, verb, line string, out any) error {
	r, err := c.Do(ctx, line)
	if err != nil {
		var ce *CmdError
		if errors.As(err, &ce) && !errors.Is(err, ErrBusy) && !errors.Is(err, ErrLineTooLong) {
			code, why, _ := strings.Cut(ce.Msg, " ")
			switch {
			case strings.HasPrefix(ce.Msg, "unknown"), code == "game", code == "joined", code == "usage":
				return err
			}
			return &VerbRefusal{Verb: verb, Code: code, Why: why}
		}
		return err
	}
	if out == nil {
		return nil
	}
	line1 := strings.TrimSpace(r.Rest)
	for _, l := range r.Lines {
		if strings.HasPrefix(strings.TrimSpace(l), "{") {
			line1 = strings.TrimSpace(l)
			break
		}
	}
	if err := json.Unmarshal([]byte(line1), out); err != nil {
		return fmt.Errorf("d2dctl: %s reply %q: %w", verb, line1, err)
	}
	return nil
}

// BuyResult is a successful buy: what was bought (Code), what it cost and the purse left (carried gold).
type BuyResult struct {
	Code      string `json:"code"`
	Stock     int    `json:"stock"`
	Tab       int    `json:"tab"`
	Price     int    `json:"price"`
	GoldAfter int    `json:"gold_after"`
}

// VerbBuy buys stock item stock of tab (-1: the window's own tab) from the open vendor window. fill is the shift-click
// on a multibuy item.
func (c *Client) VerbBuy(ctx context.Context, stock, tab int, fill bool) (BuyResult, error) {
	line := fmt.Sprintf("buy %d", stock)
	if tab >= 0 || fill {
		if tab < 0 {
			tab = 0
		}
		line += fmt.Sprintf(" %d", tab)
	}
	if fill {
		line += " fill"
	}
	var r BuyResult
	return r, c.worldDo(ctx, "buy", line, &r)
}

// SellResult is a successful sale.
type SellResult struct {
	ID    int    `json:"id"`
	Code  string `json:"code"`
	Price int    `json:"price"`
}

// VerbSell sells item id (carried, worn or held) to the open vendor window.
func (c *Client) VerbSell(ctx context.Context, id uint32) (SellResult, error) {
	var r SellResult
	return r, c.worldDo(ctx, "sell", fmt.Sprintf("sell %d", id), &r)
}

// WaypointResult is a successful waypoint take.
type WaypointResult struct {
	Object int `json:"object"`
	Level  int `json:"level"`
	From   int `json:"from"`
}

// VerbWaypoint travels from the OPEN waypoint panel to the waypoint of level (object < 0: the panel's own object).
func (c *Client) VerbWaypoint(ctx context.Context, level, object int) (WaypointResult, error) {
	line := fmt.Sprintf("waypoint %d", level)
	if object >= 0 {
		line += fmt.Sprintf(" %d", object)
	}
	var r WaypointResult
	return r, c.worldDo(ctx, "waypoint", line, &r)
}

// WorldPortal is one town portal the fork lists.
type WorldPortal struct {
	Which int     `json:"which"`
	Level int     `json:"level"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Here  bool    `json:"here"`
}

// VerbPortals lists the town portals (an empty list: none opened).
func (c *Client) VerbPortals(ctx context.Context) ([]WorldPortal, error) {
	var r struct {
		Portals []WorldPortal `json:"portals"`
	}
	err := c.worldDo(ctx, "portal", "portal", &r)
	return r.Portals, err
}

// PortalResult is a successful portal take: the hero walks to the portal and goes through.
type PortalResult struct {
	Which int     `json:"which"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Dist  float64 `json:"dist"`
}

// VerbPortal enters town portal which (the message a click on it sends). It primes the walk like CmdInteract does.
func (c *Client) VerbPortal(ctx context.Context, which int) (PortalResult, error) {
	if err := c.autoKick(ctx); err != nil {
		return PortalResult{}, err
	}
	var r PortalResult
	return r, c.worldDo(ctx, "portal", fmt.Sprintf("portal %d", which), &r)
}

// PickupResult is a successful pickup request: the engine walks to the item and takes it.
type PickupResult struct {
	ID   int     `json:"id"`
	Code string  `json:"code"`
	Dist float64 `json:"dist"`
}

// VerbPickup picks up ground item id. max > 0 refuses ("far") an item farther than max cells; 0 lets the engine walk.
func (c *Client) VerbPickup(ctx context.Context, id uint32, max float64) (PickupResult, error) {
	line := fmt.Sprintf("pickup %d", id)
	if max > 0 {
		line += fmt.Sprintf(" %g", max)
	}
	var r PickupResult
	return r, c.worldDo(ctx, "pickup", line, &r)
}

// StockIndex is the index the buy verb wants for the stock item at cell (x, y) of tab: its place among that tab's items
// in the dump (the dump lists a tab's items in stock order).
func (v Vendor) StockIndex(tab, x, y int) (int, bool) {
	n := 0
	for _, it := range v.Items {
		if it.Tab != tab {
			continue
		}
		if it.X == x && it.Y == y {
			return n, true
		}
		n++
	}
	return 0, false
}
