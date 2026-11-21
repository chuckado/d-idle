package input

// The second set of fork input verbs (docs/process/input-verbs2.md): buy, sell, waypoint, portal, pickup. Each is the
// game's own client message, sent by the fork after it named any refusal. The bot adopts them only behind the knob
// input.planner (verb, then the verified legacy path), and only when the backend lists the capability.
const (
	CapBuy      = "input.buy"
	CapSell     = "input.sell"
	CapWaypoint = "input.waypoint"
	CapPortal   = "input.portal"
	CapPickup   = "input.pickup"
)

// WorldVerbs is the verb side of vendor, waypoint, portal and pickup actions. A rule refusal comes back as an error
// that implements RuleRefusal() bool (callers then do not retry the same request by clicking); every other error
// (transport, unknown verb on an older build, a joined game) is a reason to fall back.
type WorldVerbs interface {
	// BuyAt buys the stock item at cell (x, y) of vendor tab tab from the open vendor window.
	BuyAt(tab, x, y int) error
	// SellItem sells carried item id to the open vendor window.
	SellItem(id uint32) error
	// TakeWaypoint travels from the open waypoint panel to the waypoint of area level.
	TakeWaypoint(level int) error
	// EnterPortal walks to town portal which and goes through.
	EnterPortal(which int) error
	// PickupItem picks up ground item id (the engine walks to it).
	PickupItem(id uint32) error
}

// worldVerbsHolder lets a wrapper around a Controller (the d2d town's HoverFirst) pass the verbs through.
type worldVerbsHolder interface {
	WorldVerbs() (WorldVerbs, bool)
}

// AsWorldVerbs returns c's world verbs: c itself, or the one a wrapper holds.
func AsWorldVerbs(c Controller) (WorldVerbs, bool) {
	if h, ok := c.(worldVerbsHolder); ok {
		if v, ok := h.WorldVerbs(); ok {
			return v, true
		}
	}
	v, ok := c.(WorldVerbs)
	return v, ok
}

// IsRuleRefusal reports whether err is a game-rule refusal of a verb (not worth repeating by another path).
func IsRuleRefusal(err error) bool {
	for err != nil {
		if r, ok := err.(interface{ RuleRefusal() bool }); ok && r.RuleRefusal() {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// HasCap reports whether the backend lists the verb capability (the planner asks the connected build once).
func (p *Planner) HasCap(name string) bool {
	if p == nil {
		return false
	}
	return p.has(name)
}

// RefusalCode is the cause a verb named when it refused ("gold", "inactive", "closed", ...), "" for any other error.
func RefusalCode(err error) string {
	for err != nil {
		if r, ok := err.(interface{ RefusalCode() string }); ok {
			return r.RefusalCode()
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return ""
		}
		err = u.Unwrap()
	}
	return ""
}
