// Package ether is the account-level meta wallet of the Ethereal Vault (V7, design docs/design/loot-vault-2026-11-16.md
// section 2.6, docs/ether.md). ETHER is not hero gold: it lives in one append-only ledger file per profile, so it
// survives hero switches and runs. The balance and the owned capacity steps are never stored on their own: both are
// replayed from the ledger, so a crash can never leave them out of step with it.
package ether

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/vault"
)

// FileName is the ledger inside the profile's currency directory.
const FileName = "ether.jsonl"

var (
	ErrInsufficient  = errors.New("ether: balance too low")
	ErrInUse         = errors.New("ether: the vault holds more than the lower tier fits")
	ErrOutOfOrder    = errors.New("ether: capacity steps are bought and refunded in table order")
	ErrNothingToBuy  = errors.New("ether: every capacity step is owned")
	ErrNothingToUndo = errors.New("ether: no capacity step is owned")
	ErrUnknownSource = errors.New("ether: unknown source")
	ErrBadCredit     = errors.New("ether: bad credit")
	ErrCorrupt       = errors.New("ether: ledger corrupt")
)

// Entry kinds.
const (
	Credit = "credit"
	Spend  = "spend"
	Refund = "refund"
)

// Entry is one ledger line.
type Entry struct {
	Seq    int    `json:"seq"`
	TS     string `json:"ts"`
	Kind   string `json:"kind"`
	Amount int    `json:"amount"`
	Source string `json:"source,omitempty"` // credit: the earn source id
	Key    string `json:"key"`              // idempotency key, unique over the whole ledger
	Step   string `json:"step,omitempty"`   // spend or refund of a capacity step
	Note   string `json:"note,omitempty"`
}

// Options configure a Wallet. Tiers and Sources default to the embedded tables.
type Options struct {
	Tiers   *vault.Tiers
	Sources *Sources
	// Content is the number of records the active hero's vault holds; the refund rule reads it. nil counts as 0.
	Content func() int
	Now     func() time.Time
}

// Wallet is the ledger plus its replayed state. Safe for concurrent use within one process.
type Wallet struct {
	path string
	opt  Options

	mu      sync.Mutex
	entries []Entry
	keys    map[string]int // key to index in entries
	balance int
	owned   map[string]int // step id to price paid
	refunds map[string]int // step id to number of refunds so far (part of the next spend key)
	torn    int
	cap     atomic.Int64 // capacity from the owned steps, kept for Capacity (see there)
}

// Open reads or creates <dir>/ether.jsonl. A torn last line (crash mid-write) is dropped and the file cut back to
// the last whole line; damage anywhere else is an error that names the line and nothing is changed.
func Open(dir string, opt Options) (*Wallet, error) {
	var err error
	if opt.Tiers == nil {
		if opt.Tiers, err = vault.DefaultTiers(); err != nil {
			return nil, err
		}
	}
	if opt.Sources == nil {
		if opt.Sources, err = DefaultSources(); err != nil {
			return nil, err
		}
	}
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	w := &Wallet{path: filepath.Join(dir, FileName), opt: opt}
	if err := w.load(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Wallet) reset() {
	w.entries, w.keys, w.balance = nil, map[string]int{}, 0
	w.owned, w.refunds = map[string]int{}, map[string]int{}
	w.cap.Store(int64(w.opt.Tiers.Capacity(nil, 0)))
}

func (w *Wallet) load() error {
	w.reset()
	b, err := os.ReadFile(w.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	whole := b
	if i := bytes.LastIndexByte(b, '\n'); i >= 0 {
		whole = b[:i+1]
	} else {
		whole = nil
	}
	w.torn = len(b) - len(whole) // a write cut before its newline was never acknowledged
	for i, raw := range bytes.Split(bytes.TrimSuffix(whole, []byte("\n")), []byte("\n")) {
		if len(whole) == 0 {
			break
		}
		var e Entry
		if err := json.Unmarshal(raw, &e); err != nil {
			return fmt.Errorf("%w: line %d", ErrCorrupt, i+1)
		}
		if err := w.apply(e); err != nil {
			return fmt.Errorf("%w: line %d: %v", ErrCorrupt, i+1, err)
		}
	}
	if w.torn > 0 {
		return os.Truncate(w.path, int64(len(whole)))
	}
	return nil
}

// apply replays one entry into the derived state; it is the only place the state changes.
func (w *Wallet) apply(e Entry) error {
	if e.Seq != len(w.entries)+1 {
		return fmt.Errorf("seq %d after %d", e.Seq, len(w.entries))
	}
	if e.Key == "" || e.Amount < 0 {
		return fmt.Errorf("seq %d: empty key or negative amount", e.Seq)
	}
	if _, dup := w.keys[e.Key]; dup {
		return fmt.Errorf("seq %d: duplicate key %q", e.Seq, e.Key)
	}
	switch e.Kind {
	case Credit:
		w.balance += e.Amount
	case Spend:
		if e.Amount > w.balance {
			return fmt.Errorf("seq %d: spend %d over balance %d", e.Seq, e.Amount, w.balance)
		}
		w.balance -= e.Amount
		if e.Step != "" {
			w.owned[e.Step] = e.Amount
		}
	case Refund:
		w.balance += e.Amount
		if e.Step != "" {
			delete(w.owned, e.Step)
			w.refunds[e.Step]++
		}
	default:
		return fmt.Errorf("seq %d: kind %q", e.Seq, e.Kind)
	}
	w.keys[e.Key] = len(w.entries)
	w.entries = append(w.entries, e)
	w.cap.Store(int64(w.opt.Tiers.Capacity(w.ownedLocked(), 0)))
	return nil
}

// append validates, writes one fsynced line, then applies it. On a failed write the file is cut back and the
// state is untouched.
func (w *Wallet) append(e Entry) (Entry, error) {
	e.Seq = len(w.entries) + 1
	e.TS = w.opt.Now().Format(time.RFC3339) // local time: the daily cap resets at the player's midnight
	if err := w.dry(e); err != nil {
		return Entry{}, err
	}
	line, err := json.Marshal(e)
	if err != nil {
		return Entry{}, err
	}
	f, err := os.OpenFile(w.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return Entry{}, err
	}
	st, _ := f.Stat()
	_, err = f.Write(append(line, '\n'))
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		if st != nil {
			os.Truncate(w.path, st.Size())
		}
		return Entry{}, err
	}
	if err := w.apply(e); err != nil { // dry ran the same checks; this is a bug guard
		return Entry{}, err
	}
	return e, nil
}

// dry runs the apply checks on a copy of the numbers it needs.
func (w *Wallet) dry(e Entry) error {
	if _, dup := w.keys[e.Key]; dup {
		return fmt.Errorf("%w: key %q", ErrBadCredit, e.Key)
	}
	if e.Kind == Spend && e.Amount > w.balance {
		return ErrInsufficient
	}
	return nil
}

// Balance is the replayed balance; it is never negative.
func (w *Wallet) Balance() (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.balance, nil
}

// OwnedSteps implements vault.Account: the capacity steps bought and not refunded, in table order.
func (w *Wallet) OwnedSteps() ([]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ownedLocked(), nil
}

func (w *Wallet) ownedLocked() []string {
	var out []string
	for _, s := range w.opt.Tiers.Steps {
		if _, ok := w.owned[s.ID]; ok {
			out = append(out, s.ID)
		}
	}
	return out
}

// SpendKey implements vault.Wallet: a plain spend that is idempotent per key (a second call with the same key
// books nothing and returns nil) and returns ErrInsufficient, changing nothing, when the balance is short.
func (w *Wallet) SpendKey(key string, amount int) error {
	if key == "" || amount <= 0 {
		return fmt.Errorf("%w: spend key %q amount %d", ErrBadCredit, key, amount)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, dup := w.keys[key]; dup {
		return nil
	}
	_, err := w.append(Entry{Kind: Spend, Amount: amount, Key: key})
	return err
}

// Pay adapts the wallet to the vault.Wallet interface (Balance and Spend).
func (w *Wallet) Pay() vault.Wallet { return payer{w} }

type payer struct{ w *Wallet }

func (p payer) Balance() (int, error)         { return p.w.Balance() }
func (p payer) Spend(key string, n int) error { return p.w.SpendKey(key, n) }

var _ vault.Account = (*Wallet)(nil)
var _ vault.Wallet = payer{}

// Capacity is the effective vault capacity from the owned steps. It takes no lock: the vault store calls it while
// it holds its own lock, and the wallet reads the store's count (Options.Content) while it holds the wallet's, so a
// locked read here would be a lock order inversion.
func (w *Wallet) Capacity(override int) int {
	if override > 0 {
		return override
	}
	return int(w.cap.Load())
}

// ---- earning ----

// Result of an Earn call.
type Result struct {
	Credited int    `json:"credited"`
	Key      string `json:"key"`
	Already  bool   `json:"already,omitempty"` // the key was booked before: nothing paid
	Capped   bool   `json:"capped,omitempty"`  // the source's daily cap left nothing to pay
}

var difficulties = map[string]bool{"normal": true, "nm": true, "hell": true}

// Earn credits one event. sub names the event inside the source (a quest id, a monster name, a milestone id, a kill
// batch id); difficulty is required for sources with byDifficulty; hookAmount is used only by hook sources. A key
// that was booked before pays nothing (Already). Replaying a journal therefore books nothing the second time.
func (w *Wallet) Earn(source, sub, difficulty string, hookAmount int) (Result, error) {
	src, ok := w.opt.Sources.find(source)
	if !ok {
		return Result{}, fmt.Errorf("%w %q", ErrUnknownSource, source)
	}
	if sub == "" || strings.Contains(sub, "\n") {
		return Result{}, fmt.Errorf("%w: %s needs a sub key", ErrBadCredit, source)
	}
	key := source + "/" + sub
	if src.ByDifficulty {
		if !difficulties[difficulty] {
			return Result{}, fmt.Errorf("%w: %s needs a difficulty (normal, nm, hell), got %q", ErrBadCredit, source, difficulty)
		}
		key += "/" + difficulty
	}
	amt := src.Amount
	switch {
	case src.Hook:
		if hookAmount < 0 || hookAmount > src.MaxHook {
			return Result{}, fmt.Errorf("%w: %s amount %d outside 0..%d", ErrBadCredit, source, hookAmount, src.MaxHook)
		}
		amt = hookAmount
	case len(src.Difficulty) > 0:
		if amt = src.Difficulty[difficulty]; amt == 0 {
			return Result{}, fmt.Errorf("%w: %s has no pay for difficulty %q", ErrBadCredit, source, difficulty)
		}
	case len(src.Rows) > 0:
		if amt = src.Rows[sub]; amt == 0 {
			return Result{}, fmt.Errorf("%w: %s has no row %q", ErrBadCredit, source, sub)
		}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, dup := w.keys[key]; dup {
		return Result{Key: key, Already: true}, nil
	}
	if amt == 0 { // a hook source with a quest that pays nothing: book nothing
		return Result{Key: key}, nil
	}
	if src.DailyCap > 0 {
		left := src.DailyCap - w.earnedOn(source, w.opt.Now().Format("2006-01-02"))
		if left <= 0 {
			return Result{Key: key, Capped: true}, nil
		}
		amt = min(amt, left)
	}
	e, err := w.append(Entry{Kind: Credit, Amount: amt, Source: source, Key: key})
	if err != nil {
		return Result{}, err
	}
	return Result{Credited: e.Amount, Key: key}, nil
}

func (w *Wallet) earnedOn(source, day string) int {
	n := 0
	for _, e := range w.entries {
		if e.Kind == Credit && e.Source == source && strings.HasPrefix(e.TS, day) {
			n += e.Amount
		}
	}
	return n
}

// ---- capacity steps ----

// UpgradeResult is the reply of Upgrade and Downgrade.
type UpgradeResult struct {
	Step     string `json:"step"`
	Price    int    `json:"price"`
	Already  bool   `json:"already,omitempty"` // Upgrade: the step was owned; Downgrade: it was not
	Balance  int    `json:"balance"`
	Capacity int    `json:"capacity"`
}

// ShortError says how much is missing; it wraps ErrInsufficient.
type ShortError struct{ Need, Have int }

func (e *ShortError) Error() string {
	return fmt.Sprintf("%v: need %d more", ErrInsufficient, e.Need-e.Have)
}
func (e *ShortError) Unwrap() error { return ErrInsufficient }

// FitError says why a refund is refused; it wraps ErrInUse.
type FitError struct{ Count, Capacity int }

func (e *FitError) Error() string {
	return fmt.Sprintf("%v: %d records, the lower tier holds %d", ErrInUse, e.Count, e.Capacity)
}
func (e *FitError) Unwrap() error { return ErrInUse }

func (w *Wallet) result(step string, price int) UpgradeResult {
	return UpgradeResult{Step: step, Price: price, Balance: w.balance, Capacity: w.opt.Tiers.Capacity(w.ownedLocked(), 0)}
}

// Upgrade buys a capacity step: step "" means the next one. Steps are bought in table order. An owned step is a
// no-op (Already), so a double click books once. A short balance changes nothing.
func (w *Wallet) Upgrade(step string) (UpgradeResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	next, ok := w.opt.Tiers.Next(w.ownedLocked())
	if step != "" {
		if p, own := w.owned[step]; own {
			r := w.result(step, p)
			r.Already = true
			return r, nil
		}
		if !ok || next.ID != step {
			return UpgradeResult{}, fmt.Errorf("%w: next is %q, asked %q", ErrOutOfOrder, next.ID, step)
		}
	}
	if !ok {
		return UpgradeResult{}, ErrNothingToBuy
	}
	if next.Price > w.balance {
		return UpgradeResult{}, &ShortError{Need: next.Price, Have: w.balance}
	}
	key := fmt.Sprintf("vault/%s#%d", next.ID, w.refunds[next.ID])
	if _, err := w.append(Entry{Kind: Spend, Amount: next.Price, Key: key, Step: next.ID}); err != nil {
		return UpgradeResult{}, err
	}
	return w.result(next.ID, next.Price), nil
}

// Downgrade refunds the last owned step in full, only while the vault's content fits the lower tier; otherwise
// ErrInUse (FitError) and nothing changes. step "" means the last one; naming another step is out of order.
// Shrinking never deletes anything: the content is untouched either way.
func (w *Wallet) Downgrade(step string) (UpgradeResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	owned := w.ownedLocked()
	if len(owned) == 0 {
		if step != "" {
			r := w.result(step, 0)
			r.Already = true
			return r, nil
		}
		return UpgradeResult{}, ErrNothingToUndo
	}
	last := owned[len(owned)-1]
	if step != "" && step != last {
		if _, own := w.owned[step]; !own {
			r := w.result(step, 0)
			r.Already = true
			return r, nil
		}
		return UpgradeResult{}, fmt.Errorf("%w: only %q can be refunded", ErrOutOfOrder, last)
	}
	lower := w.opt.Tiers.Capacity(owned[:len(owned)-1], 0)
	if n := w.content(); n > lower {
		return UpgradeResult{}, &FitError{Count: n, Capacity: lower}
	}
	price := w.owned[last]
	key := fmt.Sprintf("vault/%s#%d/refund", last, w.refunds[last])
	if _, err := w.append(Entry{Kind: Refund, Amount: price, Key: key, Step: last}); err != nil {
		return UpgradeResult{}, err
	}
	return w.result(last, price), nil
}

func (w *Wallet) content() int {
	if w.opt.Content == nil {
		return 0
	}
	return w.opt.Content()
}

// ---- views ----

// Summary is what GET /api/ether and the wallet chips show.
type Summary struct {
	Currency  string         `json:"currency"`
	Balance   int            `json:"balance"`
	Recent    []Entry        `json:"recent"` // newest first, at most 20
	Owned     []string       `json:"owned"`
	Capacity  int            `json:"capacity"`
	Next      *NextStep      `json:"next,omitempty"`
	Downgrade DowngradeView  `json:"downgrade"`
	Sources   []SourceView   `json:"sources"`
	Earned    map[string]int `json:"earned"` // total credited per source id
}

// NextStep is the upgrade the capacity meter offers.
type NextStep struct {
	ID        string `json:"id"`
	Add       int    `json:"add"`
	Price     int    `json:"price"`
	CanAfford bool   `json:"canAfford"`
	Short     int    `json:"short,omitempty"` // how much more ETHER is needed
	Capacity  int    `json:"capacity"`        // capacity after buying it
}

// DowngradeView says whether the last step can be refunded now.
type DowngradeView struct {
	Available bool   `json:"available"`
	Step      string `json:"step,omitempty"`
	Refund    int    `json:"refund,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// SourceView is one earn rule for the "how to earn" list.
type SourceView struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	DailyCap    int    `json:"dailyCap,omitempty"`
	EarnedToday int    `json:"earnedToday"`
}

// Summary builds the view in one locked pass.
func (w *Wallet) Summary() Summary {
	w.mu.Lock()
	defer w.mu.Unlock()
	t := w.opt.Tiers
	owned := w.ownedLocked()
	s := Summary{Currency: w.opt.Sources.Currency, Balance: w.balance, Owned: owned, Capacity: t.Capacity(owned, 0),
		Recent: []Entry{}, Earned: map[string]int{}}
	if s.Owned == nil {
		s.Owned = []string{}
	}
	for i := len(w.entries) - 1; i >= 0 && len(s.Recent) < 20; i-- {
		s.Recent = append(s.Recent, w.entries[i])
	}
	for _, e := range w.entries {
		if e.Kind == Credit {
			s.Earned[e.Source] += e.Amount
		}
	}
	if n, ok := t.Next(owned); ok {
		ns := &NextStep{ID: n.ID, Add: n.Add, Price: n.Price, CanAfford: n.Price <= w.balance,
			Capacity: t.Capacity(append(append([]string{}, owned...), n.ID), 0)}
		if !ns.CanAfford {
			ns.Short = n.Price - w.balance
		}
		s.Next = ns
	}
	switch {
	case len(owned) == 0:
		s.Downgrade.Reason = "no capacity step is owned"
	default:
		last := owned[len(owned)-1]
		s.Downgrade.Step, s.Downgrade.Refund = last, w.owned[last]
		if lower, n := t.Capacity(owned[:len(owned)-1], 0), w.content(); n > lower {
			s.Downgrade.Reason = fmt.Sprintf("the vault holds %d records and the lower tier fits %d: discard %d first", n, lower, n-lower)
		} else {
			s.Downgrade.Available = true
		}
	}
	today := w.opt.Now().Format("2006-01-02")
	for _, r := range w.opt.Sources.Sources {
		s.Sources = append(s.Sources, SourceView{ID: r.ID, Label: r.Label, DailyCap: r.DailyCap, EarnedToday: w.earnedOn(r.ID, today)})
	}
	return s
}

// Entries is a copy of the whole ledger (audit, export).
func (w *Wallet) Entries() []Entry {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]Entry(nil), w.entries...)
}

// Repaired is the number of bytes of a torn last line dropped at Open.
func (w *Wallet) Repaired() int { return w.torn }
