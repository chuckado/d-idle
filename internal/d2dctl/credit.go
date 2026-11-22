package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"d2lod-bot/internal/cheat"
)

// `debug creditgold <amount> <orderid>`: the vault Sell's ONE game write (ADR 0029 sale amendment, proposed in the card result).
// The fork (private lab binary, flag --vault-pay, the same flag as paygold) ADDS <amount> to the STASH gold only (never the
// purse), once per order id, refuses a measured run and refuses (changing nothing) an amount past the stash limit of
// 2,500,000, naming the room that is left. Here: the client of that verb (ledger line BEFORE it is sent, reason `player vault
// sale order <id>`) and an in-memory stand-in. Only the vault API with the answer listener's auth reaches it, never the bot's
// own play, never in a measured run.
//
//	debug creditgold <amount> <orderid>   {"ok":true,"credited":N,"stash_before":A,"stash_after":B,"room":R,"order":"id"}
//	                                      a repeated id answers credited:0 already:true; err ... on a refusal
//
// And `item price <hex> [npc]` (read-only): what the engine's own sell function pays for the recorded bytes.

// CapVaultCreditgold is the capability name a --vault-pay fork build lists in `caps`.
const CapVaultCreditgold = "vault.creditgold"

// StashLimit is the most gold the stash holds (game.exe FUN_00623460; docs/research/price-trace.md Q4, V by disassembly).
const StashLimit = 2500000

// ErrCreditRoom is the fork's refusal of a credit that would pass the stash limit.
var ErrCreditRoom = errors.New("d2dctl: the stash has no room for that much gold")

// CreditResult is the fork's answer: what was added, the stash before and after, and the room left.
type CreditResult struct {
	Credited    int    `json:"credited"`
	StashBefore int    `json:"stash_before"`
	StashAfter  int    `json:"stash_after"`
	Room        int    `json:"room"`
	Order       string `json:"order"`
	Already     bool   `json:"already"`
}

// GoldCreditor is the surface the vault keeper codes against: *Pay (a fork) and *PayFake (tests).
type GoldCreditor interface {
	CreditGold(ctx context.Context, amount int, order string) (CreditResult, error)
}

// CreditGold adds amount gold to the stash for the order. The ledger line is written first (an unwritable ledger stops it).
func (p *Pay) CreditGold(ctx context.Context, amount int, order string) (CreditResult, error) {
	if p == nil || p.C == nil || p.Ledger == nil || p.Actor == "" {
		return CreditResult{}, errors.New("d2dctl: Pay needs a client, a ledger and an actor")
	}
	switch {
	case measuredNow(p.Measured):
		return CreditResult{}, ErrPayMeasured
	case amount <= 0:
		return CreditResult{}, ErrPayAmount
	case !orderIDRe.MatchString(order):
		return CreditResult{}, fmt.Errorf("d2dctl: creditgold order id %q: 1..64 of [A-Za-z0-9._:-]", order)
	}
	e := cheat.Entry{Time: time.Now().UTC(), Actor: p.Actor, Instance: p.Instance, Character: p.Character, Kind: KindD2DDebug,
		Params: map[string]any{"verb": "debug creditgold", "amount": amount, "order": order}, Reason: "player vault sale order " + order}
	if err := p.Ledger.Append(e); err != nil { // before anything is sent
		return CreditResult{}, err
	}
	r, err := p.C.Do(ctx, fmt.Sprintf("debug creditgold %d %s", amount, order))
	var res CreditResult
	if err == nil {
		res, err = parseCredit(r)
	} else {
		err = mapCreditErr(err)
	}
	e.Time, e.Success = time.Now().UTC(), err == nil
	e.After = map[string]any{"credited": res.Credited, "stash_before": res.StashBefore, "stash_after": res.StashAfter, "already": res.Already}
	if err != nil {
		e.Error = err.Error()
	}
	if lerr := p.Ledger.Append(e); lerr != nil && err == nil {
		err = lerr
	}
	return res, err
}

func parseCredit(r *Reply) (CreditResult, error) {
	for _, l := range r.Lines {
		if !strings.HasPrefix(strings.TrimSpace(l), "{") {
			continue
		}
		var out struct {
			OK bool `json:"ok"`
			CreditResult
		}
		if err := json.Unmarshal([]byte(l), &out); err != nil {
			return CreditResult{}, fmt.Errorf("d2dctl: creditgold reply: %w", err)
		}
		if !out.OK {
			return CreditResult{}, errors.New("d2dctl: creditgold was not accepted")
		}
		return out.CreditResult, nil
	}
	return CreditResult{}, errors.New("d2dctl: creditgold reply holds no answer")
}

// mapCreditErr is mapPayErr plus the room refusal.
func mapCreditErr(err error) error {
	var ce *CmdError
	if errors.As(err, &ce) && strings.HasPrefix(ce.Msg, "room") {
		return fmt.Errorf("%w: %s", ErrCreditRoom, ce.Msg)
	}
	return mapPayErr(err)
}

// CreditGold mirrors the verb on the in-memory stash (limit StashLimit unless Limit is set).
func (f *PayFake) CreditGold(_ context.Context, amount int, order string) (CreditResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, fmt.Sprintf("credit %d %s", amount, order))
	limit := StashLimit
	if f.Limit > 0 {
		limit = f.Limit
	}
	switch {
	case f.FlagOff:
		return CreditResult{}, ErrPayOff
	case f.Measured:
		return CreditResult{}, ErrPayMeasured
	case f.Fail != nil:
		return CreditResult{}, f.Fail
	case !orderIDRe.MatchString(order):
		return CreditResult{}, fmt.Errorf("d2dctl: creditgold order id %q", order)
	}
	room := max(limit-f.Stash, 0)
	if prev, ok := f.credited[order]; ok {
		return CreditResult{Credited: 0, StashBefore: f.Stash, StashAfter: f.Stash, Room: room, Order: prev.Order, Already: true}, nil
	}
	switch {
	case amount <= 0:
		return CreditResult{}, ErrPayAmount
	case amount > room:
		return CreditResult{}, fmt.Errorf("%w: %d asked, room for %d", ErrCreditRoom, amount, room)
	}
	res := CreditResult{Credited: amount, StashBefore: f.Stash, StashAfter: f.Stash + amount, Room: room - amount, Order: order}
	f.Stash += amount
	if f.credited == nil {
		f.credited = map[string]CreditResult{}
	}
	f.credited[order] = res
	return res, nil
}

// ItemPrice is what the engine's sell function pays for one item's bytes at one vendor.
type ItemPrice struct {
	Sell   int    // the price of the bytes as recorded (an unidentified item sells for its base)
	SellID int    // the price of the same item once identified (quality extras and stat costs)
	Was    bool   // the recorded bytes are identified
	Npc    string // the vendor the price is for
}

// Pricer is the read-only price verb; *Client and *VaultFake implement it. A stock d2d or an older fork answers ErrVaultUnsupported.
type Pricer interface {
	ItemPrice(ctx context.Context, hex, npc string) (ItemPrice, error)
}

// ItemPrice asks the fork `item price <hex> [npc]` (npc "" = the fork's default vendor, Akara). Nothing is created and no game
// state changes (V once the fork test tests/itemprice_d2d.py has run: the state digest is equal before and after).
func (c *Client) ItemPrice(ctx context.Context, hexNorm, npc string) (ItemPrice, error) {
	if err := ValidHex(hexNorm); err != nil {
		return ItemPrice{}, err
	}
	line := "item price " + hexNorm
	if npc != "" {
		line += " " + npc
	}
	r, err := c.Do(ctx, line)
	if err != nil {
		return ItemPrice{}, mapVaultErr(err)
	}
	for _, l := range r.Lines {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), "price "); ok {
			return ParsePriceLine(rest)
		}
	}
	return ItemPrice{}, fmt.Errorf("d2dctl: item price answered no price line")
}

// ParsePriceLine parses what follows "price " in the verb's answer: `npc=<id> sell=<n> sell_id=<n> was=<0|1>`.
func ParsePriceLine(rest string) (ItemPrice, error) {
	var p ItemPrice
	seen := 0
	for _, f := range strings.Fields(rest) {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			continue
		}
		var n int
		switch k {
		case "npc":
			p.Npc = v
			seen++
			continue
		case "sell", "sell_id", "was":
			if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
				return ItemPrice{}, fmt.Errorf("d2dctl: item price field %s=%q", k, v)
			}
		default:
			continue
		}
		switch k {
		case "sell":
			p.Sell = n
		case "sell_id":
			p.SellID = n
		case "was":
			p.Was = n == 1
		}
		seen++
	}
	if seen < 4 || p.Sell < 0 || p.SellID < 0 {
		return ItemPrice{}, fmt.Errorf("d2dctl: item price line %q", rest)
	}
	return p, nil
}

var (
	_ GoldCreditor = (*Pay)(nil)
	_ GoldCreditor = (*PayFake)(nil)
	_ Pricer       = (*Client)(nil)
)
