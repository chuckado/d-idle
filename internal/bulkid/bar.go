package bulkid

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// The Reveal bar of the Vault page (docs/vault.md, "Reveal paid in gold"): one line per way to pay, the default, and the
// items an option cannot cover. Pure: the page's server fills BarIn from the hero and the wallet, nothing is spent here.

// Payment ids of a bar option.
const (
	PayFree  = "free"  // Cain stands here, or scrolls and tome charges the hero already holds
	PayEther = "ether" // unlock from the affixes the vault holds, Ether charged per item
	PayGold  = "gold"  // unlock from the held affixes, the price taken from the stash gold at once
)

// GoldMax is the most items one gold order names (the POST body cap, internal/watch vaultMaxIDs).
const GoldMax = 150

// ScopeAllUnrevealed is the gold order scope "Identify all": the server lists every unrevealed record of the whole vault
// itself when the order runs. GoldAllMax is the most records one such order takes (the order is stored, not posted).
const (
	ScopeAllUnrevealed = "all_unrevealed"
	GoldAllMax         = 1000
)

// BarIn is what the quote reads.
type BarIn struct {
	Free     bool // Cain stands here
	Charges  int  // scrolls plus tome charges held, -1 unknown counts as none
	EtherPer int  // 0: Ether is not a payer
	EtherBal int
	// Gold: ScrollGold is the price of one reveal, the vendor price of one scroll (0 unknown). Stash is the stash gold the
	// order pays from (never the purse). GoldOn is knob loot.vaultRevealGold with a payer wired.
	GoldOn     bool
	ScrollGold int
	Stash      int
	ConfirmAt  int // gold from which the order needs a confirming press (0: never)
	// Decode: the fork can decode an item's affixes from its own bytes (`item decode`), so a gold reveal does not need the
	// vault to hold them already: every unidentified candidate is gold-revealable (the order decodes it first).
	Decode            bool
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
		// Gold unlocks from the affixes the vault holds, exactly as Ether does (EnqueueRevealGold): the price of an item is the
		// vendor price of one Scroll of Identify, taken from the STASH gold at once (nothing is bought, no town visit).
		g := BarOption{ID: PayGold, Label: "Gold", IDs: []string{}}
		var sids []string
		for _, c := range cs {
			if c.Sealed || in.Decode {
				sids = append(sids, c.ID)
			}
		}
		sealed := len(sids)
		room := min(sealed, GoldMax)
		switch {
		case in.ActiveGoldOrderID != "":
			g.Why = "a gold reveal is already running"
		case in.ScrollGold <= 0:
			g.Why = "the scroll price is unknown"
		case in.GoldWhy != "":
			g.Why = in.GoldWhy
		case sealed == 0:
			g.Why = "the vault holds no affixes for these items"
		default:
			g.Count = min(room, max(in.Stash, 0)/in.ScrollGold)
			g.Cost = g.Count * in.ScrollGold
			g.OK = g.Count > 0
		}
		g.IDs = sids[:g.Count]
		g.Line = fmt.Sprintf("Gold: %s (stash %s)", Thousands(g.Cost), Thousands(in.Stash))
		g.Confirm = in.ConfirmAt > 0 && g.Cost >= in.ConfirmAt
		if g.OK {
			g.Skipped = skip(skip(skip(nil, n-sealed, "the vault holds no affixes for them"), room-g.Count, fmt.Sprintf("not enough gold in the stash (%s each)", Thousands(in.ScrollGold))), sealed-room, "over the "+strconv.Itoa(GoldMax)+" one order takes")
		} else if g.Why == "" {
			g.Why = fmt.Sprintf("not enough gold in the stash (%s each, you have %s)", Thousands(in.ScrollGold), Thousands(in.Stash))
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

// Only keeps the options named in ids (the Vault page's offered payments, knob loot.vaultRevealMethods) and works the default
// and the summary line out again; the other options are not hidden by the page but absent from the answer. ids nil keeps all.
func (b Bar) Only(ids []string) Bar {
	if ids == nil {
		return b
	}
	out := Bar{Total: b.Total, Options: []BarOption{}}
	for _, o := range b.Options {
		if slices.Contains(ids, o.ID) {
			out.Options = append(out.Options, o)
		}
	}
	best := -1
	var lines []string
	for i, o := range out.Options {
		lines = append(lines, o.Line)
		if o.OK && (best < 0 || o.Count > out.Options[best].Count && out.Options[best].Count < b.Total) {
			best = i
		}
	}
	if best >= 0 {
		out.Default = out.Options[best].ID
	}
	out.Line = strings.Join(lines, " · ")
	return out
}
