package bulkid

import (
	"fmt"
	"strconv"
	"strings"
)

// The Reveal bar of the Vault page (docs/vault.md, "Reveal paid in gold"): one line per way to pay, the default, and the
// items an option cannot cover. Pure: the page's server fills BarIn from the hero and the wallet, nothing is spent here.

// Payment ids of a bar option.
const (
	PayFree  = "free"  // Cain stands here, or scrolls and tome charges the hero already holds
	PayEther = "ether" // unlock from the affixes the vault holds, Ether charged per item
	PayGold  = "gold"  // the bot buys the missing scrolls at the next town visit
)

// GoldMax is the most items one gold order names (the POST body cap, internal/watch vaultMaxIDs).
const GoldMax = 150

// BarIn is what the quote reads.
type BarIn struct {
	Free     bool // Cain stands here
	Charges  int  // scrolls plus tome charges held, -1 unknown counts as none
	EtherPer int  // 0: Ether is not a payer
	EtherBal int
	// Gold: ScrollGold is the vendor price of one scroll (0 unknown). Purse and Stash are the hero's gold, Reserve the gold
	// the town guards keep back for the repair and the potion kit. GoldOn is knob loot.vaultRevealGold with a shop wired.
	GoldOn            bool
	ScrollGold        int
	Purse, Stash      int
	Reserve           int
	ConfirmAt         int // gold from which the order needs a confirming press (0: never)
	GoldWhy           string
	ActiveGoldOrderID string // an order is running: a second one is refused until it ends
}

// BarSkip groups the items an option cannot cover under one plain reason.
type BarSkip struct {
	N   int    `json:"n"`
	Why string `json:"why"`
}

// BarOption is one way to pay for the reveal of the items in the filter.
type BarOption struct {
	ID      string    `json:"id"`
	Label   string    `json:"label"`
	Line    string    `json:"line"`  // the one plain cost line of this option
	OK      bool      `json:"ok"`    // the option can reveal at least one item now
	Count   int       `json:"count"` // items it reveals
	IDs     []string  `json:"ids"`
	Cost    int       `json:"cost"`              // Ether or gold (0 for free)
	Confirm bool      `json:"confirm,omitempty"` // gold at or over the page's threshold
	Why     string    `json:"why,omitempty"`     // when not OK
	Skipped []BarSkip `json:"skipped,omitempty"`
}

// Bar is the quote for the N filtered items.
type Bar struct {
	Total   int         `json:"total"`
	Options []BarOption `json:"options"`
	Default string      `json:"default,omitempty"` // option id picked for the player
	Line    string      `json:"line"`              // "Ether: 29 of 665 · Gold: 5,800 (stash 7,200) · Scrolls: 12"
}

// Thousands writes n with a comma every three digits.
func Thousands(n int) string {
	s := strconv.Itoa(max(n, 0))
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func candIDs(cs []Cand, n int) []string {
	out := make([]string, 0, min(n, len(cs)))
	for _, c := range cs[:min(n, len(cs))] {
		out = append(out, c.ID)
	}
	return out
}

func skip(l []BarSkip, n int, why string) []BarSkip {
	if n > 0 {
		l = append(l, BarSkip{n, why})
	}
	return l
}

// MakeBar quotes the three payments. The default is the first option, in the order free, Ether, gold, that reveals every
// item; when none does, the one that reveals the most (the earlier one on a tie): the cheapest the hero can pay.
func MakeBar(cs []Cand, in BarIn) Bar {
	n := len(cs)
	b := Bar{Total: n, Options: []BarOption{}}
	held := max(in.Charges, 0)

	free := BarOption{ID: PayFree, Label: "Free", IDs: []string{}}
	switch {
	case in.Free:
		free.Count, free.Line = min(n, MassMax), "Cain: free"
	default:
		free.Count, free.Line = min(n, MassMax, held), fmt.Sprintf("Scrolls: %d", held)
	}
	free.OK, free.IDs = free.Count > 0, candIDs(cs, free.Count)
	if !free.OK {
		free.Why = "no scroll or tome charge, and Cain is not here"
	}
	free.Skipped = skip(nil, n-free.Count, free.Why)
	if free.Skipped != nil && free.Why == "" {
		free.Skipped[0].Why = "over the " + strconv.Itoa(MassMax) + " one press takes"
	}
	b.Options = append(b.Options, free)

	if in.EtherPer > 0 && !in.Free {
		sealed := 0
		var sids []string
		for _, c := range cs {
			if c.Sealed {
				sids = append(sids, c.ID)
			}
		}
		sealed = len(sids)
		cnt := min(MassMax, sealed, max(in.EtherBal, 0)/in.EtherPer)
		e := BarOption{ID: PayEther, Label: "Ether", Count: cnt, IDs: sids[:cnt], Cost: cnt * in.EtherPer, OK: cnt > 0}
		e.Line = fmt.Sprintf("Ether: %s of %s", Thousands(e.Cost), Thousands(max(in.EtherBal, 0)))
		switch {
		case sealed == 0:
			e.Why = "the vault holds no affixes for these items"
		case cnt == 0:
			e.Why = fmt.Sprintf("not enough Ether (%d each, you have %d)", in.EtherPer, max(in.EtherBal, 0))
		}
		e.Skipped = skip(skip(nil, n-sealed, "the vault holds no affixes for them"), sealed-cnt, "not enough Ether for them")
		b.Options = append(b.Options, e)
	}

	if in.GoldOn && !in.Free {
		g := BarOption{ID: PayGold, Label: "Gold", IDs: []string{}}
		room := min(n, GoldMax)
		avail := max(in.Purse+in.Stash-in.Reserve, 0)
		switch {
		case in.ActiveGoldOrderID != "":
			g.Why = "a gold reveal is already running"
		case in.ScrollGold <= 0:
			g.Why = "the scroll price is unknown"
		case in.GoldWhy != "":
			g.Why = in.GoldWhy
		default:
			g.Count = min(room, held+avail/in.ScrollGold)
			g.Cost = max(g.Count-held, 0) * in.ScrollGold
			g.OK = g.Count > 0
		}
		g.IDs = candIDs(cs, g.Count)
		g.Line = fmt.Sprintf("Gold: %s (stash %s)", Thousands(g.Cost), Thousands(in.Stash))
		g.Confirm = in.ConfirmAt > 0 && g.Cost >= in.ConfirmAt
		if g.OK {
			g.Skipped = skip(skip(nil, room-g.Count, fmt.Sprintf("not enough gold above the repair and potion reserve (%s kept)", Thousands(in.Reserve))), n-room, "over the "+strconv.Itoa(GoldMax)+" one order takes")
		} else if g.Why == "" {
			g.Why = fmt.Sprintf("not enough gold above the repair and potion reserve (%s kept)", Thousands(in.Reserve))
		}
		b.Options = append(b.Options, g)
	}

	best := -1
	for i, o := range b.Options {
		switch {
		case !o.OK:
		case best < 0 || o.Count > b.Options[best].Count && b.Options[best].Count < n:
			best = i
		}
	}
	if best >= 0 {
		b.Default = b.Options[best].ID
	}
	var lines []string
	for _, o := range b.Options {
		if o.ID != PayFree || in.Free || held > 0 {
			lines = append(lines, o.Line)
		}
	}
	if len(lines) == 0 {
		lines = append(lines, b.Options[0].Line)
	}
	b.Line = strings.Join(lines, " · ")
	return b
}
