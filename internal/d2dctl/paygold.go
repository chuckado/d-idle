package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/cheat"
)

// `debug paygold <amount> <orderid>`: the vault gold Reveal's ONE game write (ADR 0029, amending ADR 0014 for a player-paid
// vault cost on an explicit player order). The fork (private lab binary, flag --vault-pay, docs/control_channel.md there)
// takes <amount> from the STASH gold only (never the purse or the pack), once per order id, and refuses a measured run.
// Here: the client of that verb, written to the cheat ledger BEFORE it is sent with the reason `vault gold reveal order <id>`,
// and an in-memory stand-in (PayFake) for tests. The bot's own play never calls it: only the vault API with the answer
// listener's auth does (internal/vaultkeep EnqueueRevealGold), and never in a measured run.
//
//	debug paygold <amount> <orderid>   {"ok":true,"paid":N,"stash_before":A,"stash_after":B,"order":"id"}
//	                                   a repeated id answers paid:0 already:true; err ... on a refusal
//
// V (fork test paygold_d2d.py on the builder, see docs/vault.md): the reply shapes and every refusal; U: nothing here ran
// against a live bot yet.

// CapVaultPaygold is the capability name a --vault-pay fork build lists in `caps`.
const CapVaultPaygold = "vault.paygold"

// Refusal kinds of PayGold (errors.Is).
var (
	ErrPayOff      = errors.New("d2dctl: paygold is off (d2d runs without --vault-pay)")
	ErrPayMeasured = errors.New("d2dctl: paygold is refused in a measured run")
	ErrPayShort    = errors.New("d2dctl: the stash holds less than the order costs")
	ErrPayAmount   = errors.New("d2dctl: paygold needs a positive whole amount")
)

// PayResult is the fork's answer: what was paid and the stash before and after.
type PayResult struct {
	Paid        int    `json:"paid"`
	StashBefore int    `json:"stash_before"`
	StashAfter  int    `json:"stash_after"`
	Order       string `json:"order"`
	Already     bool   `json:"already"` // the order id was paid before: nothing changed now
}

// GoldPayer is the surface the vault keeper codes against: *Pay (a fork) and *PayFake (tests).
type GoldPayer interface {
	PayGold(ctx context.Context, amount int, order string) (PayResult, error)
}

var orderIDRe = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,64}$`)

// Pay is the client side of the verb. Client, Ledger and Actor are required; Measured reports a measured run (default: the
// run environment variable D2BOT_RUN is set), checked before the ledger line and before anything is sent.
type Pay struct {
	C         *Client
	Ledger    *cheat.Ledger
	Actor     string
	Instance  string
	Character string
	Measured  func() bool
}

func measuredNow(f func() bool) bool {
	if f != nil {
		return f()
	}
	return strings.TrimSpace(os.Getenv(cheat.RunEnv)) != ""
}

// PayGold pays amount gold out of the stash for the order. The ledger line is written first (an unwritable ledger stops the
// payment) and completed with the reply. A repeated order id is a no-op that answers Already.
func (p *Pay) PayGold(ctx context.Context, amount int, order string) (PayResult, error) {
	if p == nil || p.C == nil || p.Ledger == nil || p.Actor == "" {
		return PayResult{}, errors.New("d2dctl: Pay needs a client, a ledger and an actor")
	}
	switch {
	case measuredNow(p.Measured):
		return PayResult{}, ErrPayMeasured
	case amount <= 0:
		return PayResult{}, ErrPayAmount
	case !orderIDRe.MatchString(order):
		return PayResult{}, fmt.Errorf("d2dctl: paygold order id %q: 1..64 of [A-Za-z0-9._:-]", order)
	}
	verb := fmt.Sprintf("debug paygold %d %s", amount, order)
	e := cheat.Entry{Time: time.Now().UTC(), Actor: p.Actor, Instance: p.Instance, Character: p.Character, Kind: KindD2DDebug,
		Params: map[string]any{"verb": "debug paygold", "amount": amount, "order": order}, Reason: "vault gold reveal order " + order}
	if err := p.Ledger.Append(e); err != nil { // before anything is sent
		return PayResult{}, err
	}
	r, err := p.C.Do(ctx, verb)
	var res PayResult
	if err == nil {
		res, err = parsePay(r)
	} else {
		err = mapPayErr(err)
	}
	e.Time, e.Success = time.Now().UTC(), err == nil
	e.After = map[string]any{"paid": res.Paid, "stash_before": res.StashBefore, "stash_after": res.StashAfter, "already": res.Already}
	if err != nil {
		e.Error = err.Error()
	}
	if lerr := p.Ledger.Append(e); lerr != nil && err == nil {
		err = lerr
	}
	return res, err
}

func parsePay(r *Reply) (PayResult, error) {
	for _, l := range r.Lines {
		if !strings.HasPrefix(strings.TrimSpace(l), "{") {
			continue
		}
		var out struct {
			OK bool `json:"ok"`
			PayResult
		}
		if err := json.Unmarshal([]byte(l), &out); err != nil {
			return PayResult{}, fmt.Errorf("d2dctl: paygold reply: %w", err)
		}
		if !out.OK {
			return PayResult{}, errors.New("d2dctl: paygold was not accepted")
		}
		return out.PayResult, nil
	}
	return PayResult{}, errors.New("d2dctl: paygold reply holds no answer")
}

// mapPayErr turns the fork's `err ...` lines into the kinds above.
func mapPayErr(err error) error {
	var ce *CmdError
	if !errors.As(err, &ce) || errors.Is(err, ErrBusy) || errors.Is(err, ErrLineTooLong) {
		return err
	}
	msg := ce.Msg
	switch {
	case strings.HasPrefix(msg, "vault pay off") || strings.HasPrefix(msg, "lab verbs off"):
		return fmt.Errorf("%w: %s", ErrPayOff, msg)
	case strings.HasPrefix(msg, "measured run"):
		return fmt.Errorf("%w: %s", ErrPayMeasured, msg)
	case strings.HasPrefix(msg, "stash"):
		return fmt.Errorf("%w: %s", ErrPayShort, msg)
	case strings.HasPrefix(msg, "amount"):
		return fmt.Errorf("%w: %s", ErrPayAmount, msg)
	case strings.HasPrefix(msg, "unknown"):
		return fmt.Errorf("%w: %s", ErrPayOff, msg)
	}
	return err
}

// PayFake is the in-memory stand-in of a --vault-pay fork: a stash, a purse that never changes, the paid order ids and the
// same refusals. Tests set Stash and read it back.
type PayFake struct {
	mu       sync.Mutex
	Stash    int
	Purse    int  // never touched
	FlagOff  bool // d2d runs without --vault-pay
	Limit    int  // the stash limit CreditGold enforces (0: StashLimit)
	Measured bool // the fork's measured-run marker
	Fail     error
	Calls    []string // "<amount> <order>" of every request that reached the fork
	paid     map[string]PayResult
	credited map[string]CreditResult
}

// PayGold mirrors the verb.
func (f *PayFake) PayGold(_ context.Context, amount int, order string) (PayResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, fmt.Sprintf("%d %s", amount, order))
	switch {
	case f.FlagOff:
		return PayResult{}, ErrPayOff
	case f.Measured:
		return PayResult{}, ErrPayMeasured
	case f.Fail != nil:
		return PayResult{}, f.Fail
	case !orderIDRe.MatchString(order):
		return PayResult{}, fmt.Errorf("d2dctl: paygold order id %q", order)
	}
	if prev, ok := f.paid[order]; ok {
		return PayResult{Paid: 0, StashBefore: f.Stash, StashAfter: f.Stash, Order: prev.Order, Already: true}, nil
	}
	switch {
	case amount <= 0:
		return PayResult{}, ErrPayAmount
	case amount > f.Stash:
		return PayResult{}, ErrPayShort
	}
	res := PayResult{Paid: amount, StashBefore: f.Stash, StashAfter: f.Stash - amount, Order: order}
	f.Stash -= amount
	if f.paid == nil {
		f.paid = map[string]PayResult{}
	}
	f.paid[order] = res
	return res, nil
}

var (
	_ GoldPayer = (*Pay)(nil)
	_ GoldPayer = (*PayFake)(nil)
)
