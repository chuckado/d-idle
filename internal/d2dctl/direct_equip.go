package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/cheat"
)

// Direct equip verbs (ADR 0029, amended 2026-11-25): the player's Equip creates the item straight into its body slot and
// hands back the unit it displaced, so a swap needs no town trip, no backpack room and no clicks. Fork side: the private
// lab fork (devp29, spec in the ops repo, docs/private/devp29-directequip.md). Only the vault API with the answer
// listener's auth reaches them; the bot's own play never does, and a measured run refuses them (here and in the fork).
//
//	item equipdirect <hex> <slot>      {"ok":true,"slot":"head","unit":N,"removed":["<hex>",...]}
//	item unequipdirect <slot>          {"ok":true,"slot":"head","removed":["<hex>"]}
//	item wearset <slot>=<hex> ...      {"ok":true,"results":[{one result as above, or "ok":false,"code","why"}...]}   up to 12
//
// U: the wire shapes above are the bot side's reading of the brief, written before the spec landed; every field name lives
// in parseDirect, so the spec's own shapes change that one function. The slot words are the desk's slot keys (head, torso,
// rhand ...). Nothing here ran against a fork yet.

// Capability names a devp29 fork lists in `caps`.
const (
	CapItemEquipDirect   = "item.equipdirect"
	CapItemUnequipDirect = "item.unequipdirect"
	CapItemWearSet       = "item.wearset"
)

// MaxWearSet is the most pieces one wearset call takes.
const MaxWearSet = 12

// Errors of the direct verbs.
var (
	ErrDirectOff      = errors.New("d2dctl: the direct equip verbs are not in this d2d build")
	ErrDirectMeasured = errors.New("d2dctl: direct equip is refused in a measured run")
)

// DirectRefused is a piece the engine would not wear: Code is class, level, str, dex, hands, slot, state or parse, Why the
// engine's sentence for the player.
type DirectRefused struct{ Code, Why string }

func (e *DirectRefused) Error() string { return "direct equip refused (" + e.Code + "): " + e.Why }

// DirectPiece is one wearset entry: the item bytes and the body slot.
type DirectPiece struct{ Hex, Slot string }

// DirectResult is one piece's outcome. Removed are the item bytes (norm hex) of every unit the piece pushed off the body
// (the displaced one, and the shield of a two-hander).
type DirectResult struct {
	OK      bool     `json:"ok"`
	Slot    string   `json:"slot"`
	Unit    int      `json:"unit"`
	Removed []string `json:"removed"`
	Code    string   `json:"code,omitempty"`
	Why     string   `json:"why,omitempty"`
}

// DirectEquipper is what the vault keeper codes against: *DirectEquip (a fork) and *DirectFake (tests).
type DirectEquipper interface {
	// DirectCaps says whether the build has the verbs (a probe of `caps`; false on an older image).
	DirectCaps(ctx context.Context) (bool, error)
	EquipDirect(ctx context.Context, hex, slot, order string) (DirectResult, error)
	UnequipDirect(ctx context.Context, slot, order string) (DirectResult, error)
	WearSet(ctx context.Context, pieces []DirectPiece, order string) ([]DirectResult, error)
}

// DirectEquip is the client side. Client, Ledger and Actor are required; Measured as in Pay.
type DirectEquip struct {
	C         *Client
	Ledger    *cheat.Ledger
	Actor     string
	Instance  string
	Character string
	Measured  func() bool
}

func (d *DirectEquip) ready(order string) error {
	if d == nil || d.C == nil || d.Ledger == nil || d.Actor == "" {
		return errors.New("d2dctl: DirectEquip needs a client, a ledger and an actor")
	}
	if measuredNow(d.Measured) {
		return ErrDirectMeasured
	}
	if !orderIDRe.MatchString(order) {
		return fmt.Errorf("d2dctl: direct equip order id %q: 1..64 of [A-Za-z0-9._:-]", order)
	}
	return nil
}

// DirectCaps reads `caps` and wants the three verbs.
func (d *DirectEquip) DirectCaps(ctx context.Context) (bool, error) {
	caps, err := d.C.Caps(ctx)
	if err != nil {
		return false, err
	}
	return HasCap(caps, CapItemEquipDirect) && HasCap(caps, CapItemWearSet), nil
}

// run writes the ledger line BEFORE the verb (an unwritable ledger stops it), then completes it with the reply.
func (d *DirectEquip) run(ctx context.Context, verb, order string, params map[string]any, line string, parse func(*Reply) ([]DirectResult, error)) ([]DirectResult, error) {
	params["verb"] = verb
	params["order"] = order
	e := cheat.Entry{Time: time.Now().UTC(), Actor: d.Actor, Instance: d.Instance, Character: d.Character, Kind: KindD2DDebug,
		Params: params, Reason: "player vault equip order " + order}
	if err := d.Ledger.Append(e); err != nil {
		return nil, err
	}
	r, err := d.C.Do(ctx, line)
	var res []DirectResult
	if err == nil {
		res, err = parse(r)
	} else {
		err = mapDirectErr(err)
	}
	e.Time, e.Success = time.Now().UTC(), err == nil
	ok := 0
	for _, x := range res {
		if x.OK {
			ok++
		}
	}
	e.After = map[string]any{"worn": ok, "pieces": len(res)}
	if err != nil {
		e.Error = err.Error()
	}
	if lerr := d.Ledger.Append(e); lerr != nil && err == nil {
		err = lerr
	}
	return res, err
}

func one(res []DirectResult, err error) (DirectResult, error) {
	if err != nil || len(res) == 0 {
		return DirectResult{}, err
	}
	if !res[0].OK {
		return res[0], &DirectRefused{Code: res[0].Code, Why: res[0].Why}
	}
	return res[0], nil
}

func (d *DirectEquip) EquipDirect(ctx context.Context, hex, slot, order string) (DirectResult, error) {
	if err := d.ready(order); err != nil {
		return DirectResult{}, err
	}
	if err := ValidHex(hex); err != nil {
		return DirectResult{}, err
	}
	return one(d.run(ctx, "item equipdirect", order, map[string]any{"slot": slot}, "item equipdirect "+hex+" "+slot, parseDirect))
}

func (d *DirectEquip) UnequipDirect(ctx context.Context, slot, order string) (DirectResult, error) {
	if err := d.ready(order); err != nil {
		return DirectResult{}, err
	}
	return one(d.run(ctx, "item unequipdirect", order, map[string]any{"slot": slot}, "item unequipdirect "+slot, parseDirect))
}

func (d *DirectEquip) WearSet(ctx context.Context, pieces []DirectPiece, order string) ([]DirectResult, error) {
	if err := d.ready(order); err != nil {
		return nil, err
	}
	if len(pieces) == 0 || len(pieces) > MaxWearSet {
		return nil, fmt.Errorf("d2dctl: wearset takes 1..%d pieces", MaxWearSet)
	}
	line := "item wearset"
	var slots []string
	for _, p := range pieces {
		if err := ValidHex(p.Hex); err != nil {
			return nil, err
		}
		line += " " + p.Slot + "=" + p.Hex
		slots = append(slots, p.Slot)
	}
	res, err := d.run(ctx, "item wearset", order, map[string]any{"slots": slots}, line, parseDirect)
	if err == nil && len(res) != len(pieces) {
		err = fmt.Errorf("d2dctl: wearset answered %d results for %d pieces", len(res), len(pieces))
	}
	return res, err
}

// parseDirect reads the verbs' JSON answer: one result, or {"results":[...]}.
func parseDirect(r *Reply) ([]DirectResult, error) {
	for _, l := range r.Lines {
		if !strings.HasPrefix(strings.TrimSpace(l), "{") {
			continue
		}
		var out struct {
			DirectResult
			Results []DirectResult `json:"results"`
		}
		if err := json.Unmarshal([]byte(l), &out); err != nil {
			return nil, fmt.Errorf("d2dctl: direct equip reply: %w", err)
		}
		if out.Results != nil {
			return out.Results, nil
		}
		return []DirectResult{out.DirectResult}, nil
	}
	return nil, errors.New("d2dctl: direct equip reply holds no answer")
}

// mapDirectErr turns the fork's `err ...` lines into the kinds above.
func mapDirectErr(err error) error {
	var ce *CmdError
	if !errors.As(err, &ce) || errors.Is(err, ErrBusy) || errors.Is(err, ErrLineTooLong) {
		return err
	}
	msg := ce.Msg
	switch {
	case strings.HasPrefix(msg, "measured run"):
		return fmt.Errorf("%w: %s", ErrDirectMeasured, msg)
	case strings.HasPrefix(msg, "unknown") || strings.HasPrefix(msg, "lab verbs off") || strings.HasPrefix(msg, "vault verbs off"):
		return fmt.Errorf("%w: %s", ErrDirectOff, msg)
	}
	code, why, _ := strings.Cut(msg, " ")
	return &DirectRefused{Code: code, Why: why}
}

// DirectFake is the in-memory stand-in of a devp29 fork: a worn set by slot key holding item bytes, the refusal rules as a
// hook, the pieces a placement pushes off as a hook, the measured-run marker and a call log. Tests set Worn and Rule.
type DirectFake struct {
	mu       sync.Mutex
	FlagOff  bool                                  // an older image: no verbs
	Measured bool                                  // the fork's measured-run marker
	Fail     error                                 // every call answers this
	Worn     map[string]string                     // slot key -> item norm hex
	Rule     func(hex, slot string) *DirectRefused // class, level, strength, dexterity, hands (nil: everything fits)
	Pushes   func(hex, slot string) []string       // other slots a placement empties (a two-hander and the shield)
	Calls    []string                              // "<verb> <slot> <order>" per request that reached the fork
	next     int
}

func (f *DirectFake) DirectCaps(context.Context) (bool, error) { return !f.FlagOff, nil }

func (f *DirectFake) gate(verb, order string) error {
	switch {
	case f.FlagOff:
		return ErrDirectOff
	case f.Measured:
		return ErrDirectMeasured
	case f.Fail != nil:
		return f.Fail
	case !orderIDRe.MatchString(order):
		return fmt.Errorf("d2dctl: direct equip order id %q", order)
	}
	return nil
}

// place is one piece; the caller holds the lock.
func (f *DirectFake) place(hex, slot string) DirectResult {
	if f.Worn == nil {
		f.Worn = map[string]string{}
	}
	if f.Rule != nil {
		if r := f.Rule(hex, slot); r != nil {
			return DirectResult{Slot: slot, Code: r.Code, Why: r.Why}
		}
	}
	res := DirectResult{OK: true, Slot: slot, Removed: []string{}}
	take := func(s string) {
		if old, ok := f.Worn[s]; ok {
			res.Removed = append(res.Removed, old)
			delete(f.Worn, s)
		}
	}
	take(slot)
	if f.Pushes != nil {
		for _, s := range f.Pushes(hex, slot) {
			take(s)
		}
	}
	f.next++
	res.Unit = 500 + f.next
	f.Worn[slot] = hex
	return res
}

func (f *DirectFake) EquipDirect(_ context.Context, hex, slot, order string) (DirectResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, "equipdirect "+slot+" "+order)
	if err := f.gate("equipdirect", order); err != nil {
		return DirectResult{}, err
	}
	return one([]DirectResult{f.place(hex, slot)}, nil)
}

func (f *DirectFake) UnequipDirect(_ context.Context, slot, order string) (DirectResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, "unequipdirect "+slot+" "+order)
	if err := f.gate("unequipdirect", order); err != nil {
		return DirectResult{}, err
	}
	old, ok := f.Worn[slot]
	if !ok {
		return DirectResult{}, &DirectRefused{Code: "slot", Why: "nothing is worn there"}
	}
	delete(f.Worn, slot)
	return DirectResult{OK: true, Slot: slot, Removed: []string{old}}, nil
}

func (f *DirectFake) WearSet(_ context.Context, pieces []DirectPiece, order string) ([]DirectResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, fmt.Sprintf("wearset %d %s", len(pieces), order))
	if err := f.gate("wearset", order); err != nil {
		return nil, err
	}
	if len(pieces) == 0 || len(pieces) > MaxWearSet {
		return nil, fmt.Errorf("d2dctl: wearset takes 1..%d pieces", MaxWearSet)
	}
	out := make([]DirectResult, len(pieces))
	for i, p := range pieces {
		out[i] = f.place(p.Hex, p.Slot)
	}
	return out, nil
}

var (
	_ DirectEquipper = (*DirectEquip)(nil)
	_ DirectEquipper = (*DirectFake)(nil)
)
