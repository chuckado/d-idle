package invdesk

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/d2"
)

// The fast lane (docs/inventory-actions.md, "Instant actions"): where the backend offers the item verbs, a request the
// player gives runs at once through them instead of waiting for the bot's goal step and a safe moment. The POST that
// queued the request waits a bounded time for the verb and answers with the final state; anything the lane cannot do
// now stays in the queue unchanged. Nothing here writes game memory and nothing here clicks: a verb that cannot run
// leaves the request to the queue path, which keeps the click path.

// How values added by the lane: HowFast is a request the lane did through the verbs; HowQueue is only ever the answer
// of a request that fell back to the queue (it is not stored).
const (
	HowFast  = "fast"
	HowQueue = "queue"
)

// Request.Wait values.
const (
	WaitSafe = "safe" // the bot's next safe moment
	WaitTown = "town" // the next town visit (stash, take, gold, and the identify of a stash item)
)

// CapItemUnequip is the capability name of the unequip verb (d2dctl.CapItemUnequip).
const CapItemUnequip = "item.unequip"

// DefaultFastWait bounds how long a POST waits for the verb (a verb answers "queued" and the end state is read back from
// fresh snapshots, up to Settles x 100 ms).
const DefaultFastWait = 4 * time.Second

// FastVerbs is the whole item verb surface (d2dctl.ItemVerbs plus unequip; cmd/d2bot adapts a client). It extends both
// ItemVerbs and MoverVerbs, so one serialised instance serves the bot's executors and the lane.
type FastVerbs interface {
	ItemVerbs
	MoverVerbs
	ItemUnequip(ctx context.Context, id uint32) (noop bool, err error)
}

// VerbRefusal is a rule refusal that carries the verb's code (the vocabulary of docs/design/item-verbs.md section 5).
type VerbRefusal struct{ Code, Why string }

func (e *VerbRefusal) Error() string     { return e.Why }
func (e *VerbRefusal) RuleRefusal() bool { return true }

type coded interface{ refusal() (code, why string) }

func (e *VerbRefusal) refusal() (string, string) { return e.Code, e.Why }

// sentKey carries a *bool that Serial sets once it holds the channel and is about to write the verb: from then on the
// verb may have been applied whatever the answer says (a deadline inside the call, a lost reply).
type sentKey struct{}

func markSent(ctx context.Context) {
	if p, ok := ctx.Value(sentKey{}).(*bool); ok {
		*p = true
	}
}

// afterOK wraps an error that came after the verb was sent (the read-back ran out of time): it is final, never retried.
type afterOK struct{ error }

// errVerify: the verb answered ok but a fresh snapshot does not show the change (reported failed, never retried).
var errVerify = errors.New("the item verb answered ok but the game did not show the change")

// Channel is the one owner slot of the d2d control channel for item verbs. The bot's executors and the fast lane reach
// the verbs only through Serial, so two verbs are never in flight together and a lane request cannot interleave with a
// bot verb. It is a token, not a mutex held across desks: every holder releases before it returns.
type Channel struct{ tok chan struct{} }

// NewChannel makes a free channel.
func NewChannel() *Channel { return &Channel{tok: make(chan struct{}, 1)} }

// Acquire waits for the channel or the context.
func (c *Channel) Acquire(ctx context.Context) error {
	select {
	case c.tok <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees the channel; call it once per successful Acquire.
func (c *Channel) Release() { <-c.tok }

// Serial runs every verb of V while holding the channel.
type Serial struct {
	Ch *Channel
	V  FastVerbs
}

func (s Serial) hold(ctx context.Context) (func(), error) {
	if err := s.Ch.Acquire(ctx); err != nil {
		return nil, err
	}
	markSent(ctx)
	return s.Ch.Release, nil
}

func (s Serial) Caps(ctx context.Context) ([]string, error) {
	rel, err := s.hold(ctx)
	if err != nil {
		return nil, err
	}
	defer rel()
	return s.V.Caps(ctx)
}

func (s Serial) ItemEquip(ctx context.Context, id uint32, slot int) (bool, error) {
	rel, err := s.hold(ctx)
	if err != nil {
		return false, err
	}
	defer rel()
	return s.V.ItemEquip(ctx, id, slot)
}

func (s Serial) ItemUnequip(ctx context.Context, id uint32) (bool, error) {
	rel, err := s.hold(ctx)
	if err != nil {
		return false, err
	}
	defer rel()
	return s.V.ItemUnequip(ctx, id)
}

func (s Serial) ItemMove(ctx context.Context, id uint32, panel string, col, row int) (bool, error) {
	rel, err := s.hold(ctx)
	if err != nil {
		return false, err
	}
	defer rel()
	return s.V.ItemMove(ctx, id, panel, col, row)
}

func (s Serial) ItemIdentify(ctx context.Context, id uint32) (bool, error) {
	rel, err := s.hold(ctx)
	if err != nil {
		return false, err
	}
	defer rel()
	return s.V.ItemIdentify(ctx, id)
}

func (s Serial) ItemDrop(ctx context.Context, id uint32) error {
	rel, err := s.hold(ctx)
	if err != nil {
		return err
	}
	defer rel()
	return s.V.ItemDrop(ctx, id)
}

func (s Serial) GoldStash(ctx context.Context, amount int, deposit bool) (int, int, error) {
	rel, err := s.hold(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer rel()
	return s.V.GoldStash(ctx, amount, deposit)
}

var _ FastVerbs = Serial{}

// FastLane is the lane's configuration (Desk.SetFast). Verbs should be a Serial over the shared Channel.
type FastLane struct {
	Verbs FastVerbs
	On    func() bool // the knob invdesk.verbs; nil: off
	// Instant is the knob invdesk.instant (default 0): the lane has its own switch because the verbs may be on for the bot's
	// own executors long before the instant lane has been seen working live. nil: off.
	Instant func() bool
	// Named says the process is a named run (D2BOT_RUN: gate, soak, drill): the lane is off there (the crutch rule, user
	// decision pending, docs/design/item-verbs.md section 10). nil: not named.
	Named func() bool
	// Hold says why something else owns the input right now ("" = free): a cursor item, a safety action, a click path
	// outside the desk. nil: nothing holds it.
	Hold func(s *d2.Snapshot) string
	// Fresh reads a snapshot taken now; Pause waits between polls. nil Fresh: trust the reply.
	Fresh   func() (*d2.Snapshot, error)
	Pause   func(ctx context.Context, d time.Duration) error
	Wait    time.Duration // bound of the POST; 0 = DefaultFastWait
	Settles int           // polls of 100 ms; 0 = 25
	Now     func() time.Time

	reading atomic.Bool
	mu      sync.Mutex
	caps    map[string]bool
	capsAt  time.Time
}

// SetFast installs the lane; nil removes it (every request queues, as before).
func (d *Desk) SetFast(f *FastLane) { d.mu.Lock(); d.fast = f; d.mu.Unlock() }

func (f *FastLane) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	return time.Now()
}

// why says what keeps the lane closed at all ("" = open). It does not look at the verbs or the request.
func (f *FastLane) why(s *d2.Snapshot) string {
	switch {
	case f.Instant == nil || !f.Instant():
		return "instant actions are off (knob invdesk.instant)"
	case f.Verbs == nil || f.On == nil || !f.On():
		return "item verbs are off (knob invdesk.verbs)"
	case f.Named != nil && f.Named():
		return "a named run: instant actions are off"
	case s == nil || s.Player == nil:
		return "waiting for the first look at the hero"
	case s.Player.MaxLife > 0 && s.Player.Life <= 0:
		return "the hero is dead"
	}
	if f.Hold != nil {
		return f.Hold(s)
	}
	return ""
}

// capSet reads the capability list at most once a minute; a transport error is not cached and gives no capability.
func (f *FastLane) capSet(ctx context.Context) map[string]bool {
	f.mu.Lock()
	if f.caps != nil && f.now().Sub(f.capsAt) < time.Minute {
		c := f.caps
		f.mu.Unlock()
		return c
	}
	f.mu.Unlock()
	list, err := f.Verbs.Caps(ctx)
	if err != nil {
		return nil
	}
	set := map[string]bool{}
	for _, c := range list {
		set[c] = true
	}
	f.mu.Lock()
	f.caps, f.capsAt = set, f.now()
	f.mu.Unlock()
	return set
}

// cachedCaps is the capability list as last read, and whether it is still fresh (a minute).
func (f *FastLane) cachedCaps() (map[string]bool, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.caps, f.caps != nil && f.now().Sub(f.capsAt) < time.Minute
}

// refreshAsync reads the capability list in the background (at most one read at a time) for the view.
func (f *FastLane) refreshAsync() {
	if !f.reading.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer f.reading.Store(false)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		f.capSet(ctx)
	}()
}

func (f *FastLane) dropCaps() {
	f.mu.Lock()
	f.caps, f.capsAt = map[string]bool{}, f.now()
	f.mu.Unlock()
}

// capFor names the capability a request kind needs; "" = this kind is never instant (an errand).
func capFor(r Request) string {
	switch r.kind() {
	case KindEquip:
		return CapItemEquip
	case KindUnequip:
		return CapItemUnequip
	case KindTrash:
		return CapItemDrop
	case KindIdentify:
		return CapItemIdentify
	case KindStash, KindTake:
		return CapItemMove
	case KindGold:
		return CapGoldStash
	}
	return ""
}

// needsTown: the stash verbs only work at the chest, which is in town (the verb's own rule says "far" elsewhere).
func needsTown(r Request) bool { return r.errand() }

// waitOf says what a queued request waits for.
func waitOf(r Request, inTown bool) string {
	if r.State != StateQueued {
		return ""
	}
	if r.errand() && !inTown {
		return WaitTown
	}
	return WaitSafe
}

// InstantView is what would run at once right now (API schema 9). Why says why nothing would ("" = the lane is open);
// Done counts the requests this process finished, by how.
type InstantView struct {
	Enabled  bool           `json:"enabled"` // the knob invdesk.instant; false: the UI uses the queue path
	Equip    bool           `json:"equip"`
	Unequip  bool           `json:"unequip"`
	Move     bool           `json:"move"` // the backend has the pack move verb; the desk has no pack-to-pack request yet (docs)
	Identify bool           `json:"identify"`
	Drop     bool           `json:"drop"`
	Stash    bool           `json:"stash"`
	Take     bool           `json:"take"`
	Gold     bool           `json:"gold"`
	Why      string         `json:"why,omitempty"`
	Done     map[string]int `json:"done,omitempty"`
}

func (d *Desk) doingLocked() bool {
	for _, r := range d.st.Queue {
		if r.State == StateDoing {
			return true
		}
	}
	return false
}

func (d *Desk) instantView(f *FastLane, s *d2.Snapshot) *InstantView {
	if f == nil {
		return nil
	}
	d.mu.Lock()
	v := &InstantView{}
	if len(d.done) > 0 {
		v.Done = map[string]int{}
		for k, n := range d.done {
			v.Done[k] = n
		}
	}
	busy := d.doingLocked()
	v.Enabled = f.Instant != nil && f.Instant()
	stash := d.stashSvLocked() != nil
	d.mu.Unlock()
	if v.Why = f.why(s); v.Why == "" && busy {
		v.Why = "an action is running"
	}
	if v.Why != "" {
		return v
	}
	caps, fresh := f.cachedCaps()
	if !fresh {
		f.refreshAsync() // the view never waits for the backend
	}
	if caps == nil {
		v.Why = "reading what the backend offers"
		return v
	}
	if len(caps) == 0 {
		v.Why = "the backend does not offer the item verbs"
		return v
	}
	town := s.Player.InTown()
	v.Equip, v.Unequip, v.Identify, v.Drop = caps[CapItemEquip], caps[CapItemUnequip], caps[CapItemIdentify], caps[CapItemDrop]
	v.Move = caps[CapItemMove]
	v.Stash = caps[CapItemMove] && town && stash
	v.Take = v.Stash
	v.Gold = caps[CapGoldStash] && town && stash
	return v
}

// The request entries: the checks and the queueing are the unexported request* functions; these add the lane.

func (d *Desk) RequestEquip(id uint32, slot string) (Request, error) {
	return d.lane(d.requestEquip(id, slot))
}
func (d *Desk) RequestTrash(id uint32, confirm bool) (Request, error) {
	return d.lane(d.requestTrash(id, confirm))
}
func (d *Desk) RequestUnequip(slot string) (Request, error) { return d.lane(d.requestUnequip(slot)) }
func (d *Desk) RequestIdentify(id uint32) (Request, error)  { return d.lane(d.requestIdentify(id)) }

// RequestStash queues moving backpack item id into the stash (instant at the chest, else it waits for the next town visit).
func (d *Desk) RequestStash(id uint32) (Request, error) { return d.lane(d.requestStash(id)) }

// RequestTake queues taking stash item id out into the backpack (same lane rule as RequestStash).
func (d *Desk) RequestTake(id uint32) (Request, error) { return d.lane(d.requestTake(id)) }

// RequestGold queues a gold move with the stash (same lane rule).
func (d *Desk) RequestGold(deposit bool, amount int) (Request, error) {
	return d.lane(d.requestGold(deposit, amount))
}

// lane takes a freshly queued request and tries the fast lane. A refusal of the checks passes through untouched.
func (d *Desk) lane(r Request, err error) (Request, error) {
	if err != nil {
		return r, err
	}
	d.mu.Lock()
	f := d.fast
	d.mu.Unlock()
	if f == nil {
		return r, nil
	}
	return d.runFast(f, r)
}

func queued(r Request, inTown bool) Request {
	r.How, r.Wait = HowQueue, waitOf(r, inTown)
	return r
}

// runFast is the whole fast path of one request. It returns the final state when the verb ran, else the queued request
// (How "queue", Wait says for what). A rule refusal comes back as a *Refusal with the verb's code.
func (d *Desk) runFast(f *FastLane, r Request) (Request, error) {
	s := d.snap()
	inTown := s != nil && s.Player != nil && s.Player.InTown()
	capName := capFor(r)
	if capName == "" || f.why(s) != "" || (needsTown(r) && !inTown) {
		return queued(r, inTown), nil
	}
	wait := f.Wait
	if wait <= 0 {
		wait = DefaultFastWait
	}
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	if !f.capSet(ctx)[capName] {
		return queued(r, inTown), nil
	}
	d.mu.Lock()
	i := -1
	for j, x := range d.st.Queue {
		if x.ID == r.ID && x.kind() == r.kind() && x.State == StateQueued {
			i = j
		}
	}
	if i < 0 || d.doingLocked() { // the bot took it, or the bot is running another request
		d.mu.Unlock()
		return queued(r, inTown), nil
	}
	d.st.Queue[i].State = StateDoing
	_ = d.saveLocked()
	d.mu.Unlock()

	// A panic below (a verb, a read-back) must not leave the request "doing": that would stop the whole queue.
	settled := false
	defer func() {
		if settled {
			return
		}
		d.mu.Lock()
		defer d.mu.Unlock()
		if j := d.findLocked(r.ID, r.kind()); j >= 0 {
			d.st.Queue = append(d.st.Queue[:j:j], d.st.Queue[j+1:]...)
		}
		r.State, r.Reason, r.Code, r.How = StateFailed, "the instant action failed unexpectedly; check the item and ask again", "internal", HowFast
		d.recentLocked(r)
		_ = d.saveLocked()
	}()

	var sent bool
	result, err := f.exec(context.WithValue(ctx, sentKey{}, &sent), r)
	var hx, hy int
	if err == nil && r.kind() == KindTrash {
		if s := d.snap(); s != nil && s.Player != nil {
			hx, hy = s.Player.X, s.Player.Y
		}
	}
	code, why, final := classify(err)
	// Once the verb was sent it may have been applied whatever the answer says. A gold move or a drop is not idempotent:
	// repeating it through the queue would move the gold twice, so a lost answer is final (the rule of recover()).
	if err != nil && !final && sent && (r.kind() == KindGold || r.kind() == KindTrash) {
		code, why, final = "timeout", "no answer in time and the action may have happened; check the item and ask again", true
	}
	settled = true

	d.mu.Lock()
	defer d.mu.Unlock()
	j := d.findLocked(r.ID, r.kind())
	if err != nil && !final { // the verb did not run (or is safe to repeat): the queue path takes it
		if j >= 0 {
			d.st.Queue[j].State = StateQueued
			_ = d.saveLocked()
		}
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			f.dropCaps()
		}
		return queued(r, inTown), nil
	}
	if j >= 0 {
		d.st.Queue = append(d.st.Queue[:j:j], d.st.Queue[j+1:]...)
	}
	r.How = HowFast
	if err != nil {
		r.State, r.Reason, r.Code = StateFailed, why, code
		d.recentLocked(r)
		_ = d.saveLocked()
		return r, refuse(code, "%s", why)
	}
	r.State, r.Result = StateDone, result
	switch r.kind() {
	case KindTrash:
		d.noteTrashedLocked(r, hx, hy)
	case KindUnequip:
		d.dropPinLocked(r.Slot)
	case KindEquip:
		d.pinLocked(Pin{Slot: r.Slot, ID: r.ID, Name: r.Name, Since: d.c.Now().Unix()})
	}
	d.recentLocked(r)
	_ = d.saveLocked()
	return r, nil
}

// classify sorts a verb error: final = report it as the request's result; else the queue path takes over. The codes
// "cursor" (the hand is not empty) and "far" (not at the chest) are retry-later states, not refusals.
func classify(err error) (code, why string, final bool) {
	var c coded
	var rr RuleRefusal
	switch {
	case err == nil:
		return "", "", true
	case errors.Is(err, errVerify):
		return "verify", err.Error(), true
	case errors.As(err, new(*afterOK)):
		return "timeout", "the game did not show the change in time; check the item and ask again", true
	case errors.As(err, &c):
		code, why = c.refusal()
		return code, why, code != "cursor" && code != "far"
	case errors.As(err, &rr) && rr.RuleRefusal():
		return "refused", err.Error(), true
	}
	return "", err.Error(), false
}

// exec runs the verb of r and proves the end state from fresh snapshots.
func (f *FastLane) exec(ctx context.Context, r Request) (string, error) {
	v := f.Verbs
	switch r.kind() {
	case KindEquip:
		body := SlotBody[r.Slot]
		noop, err := v.ItemEquip(ctx, r.ID, body)
		if err != nil || noop {
			return "", err
		}
		return "", f.settle(ctx, func(s *d2.Snapshot) bool {
			u, ok := findUnit(s, r.ID)
			return ok && u.Item != nil && u.Item.Where == d2.ContEquipment && u.Item.Body == body
		})
	case KindUnequip:
		noop, err := v.ItemUnequip(ctx, r.ID)
		if err != nil || noop {
			return "", err
		}
		return "", f.settle(ctx, func(s *d2.Snapshot) bool {
			u, ok := findUnit(s, r.ID)
			return ok && u.Item != nil && u.Item.Where == d2.ContInventory
		})
	case KindTrash:
		if err := v.ItemDrop(ctx, r.ID); err != nil {
			return "", err
		}
		return "", f.settle(ctx, func(s *d2.Snapshot) bool { _, ok := findUnit(s, r.ID); return !ok })
	case KindIdentify:
		noop, err := v.ItemIdentify(ctx, r.ID)
		if err != nil || noop {
			return "", err
		}
		return "", f.settle(ctx, func(s *d2.Snapshot) bool {
			u, ok := findUnit(s, r.ID)
			return !ok || (u.Item != nil && u.Item.Identified())
		})
	case KindStash, KindTake:
		panel, want := "inv", d2.ContInventory
		if r.kind() == KindStash {
			panel, want = "stash", d2.ContStash
		}
		noop, err := v.ItemMove(ctx, r.ID, panel, -1, -1)
		if err != nil || noop {
			return "", err
		}
		return "", f.settle(ctx, func(s *d2.Snapshot) bool {
			u, ok := findUnit(s, r.ID)
			return ok && u.Item != nil && u.Item.Where == want
		})
	case KindGold:
		deposit := r.Op == OpDeposit
		bank, carried, err := v.GoldStash(ctx, r.Count, deposit)
		if err != nil {
			return "", err
		}
		if err := f.settle(ctx, func(s *d2.Snapshot) bool {
			return s.Player != nil && s.Player.StashGold == bank && s.Player.Gold == carried
		}); err != nil {
			return "", err
		}
		if deposit {
			return fmt.Sprintf("deposited %d gold", r.Count), nil
		}
		return fmt.Sprintf("withdrew %d gold", r.Count), nil
	}
	return "", errors.New("not an instant kind")
}

// settle polls fresh snapshots until ok; running out of polls is errVerify, a dead context is its own error.
func (f *FastLane) settle(ctx context.Context, ok func(*d2.Snapshot) bool) error {
	if f.Fresh == nil {
		return nil // no read-back wired: trust the reply
	}
	n := f.Settles
	if n == 0 {
		n = 25
	}
	for i := 0; i < n; i++ {
		if s, err := f.Fresh(); err == nil && s != nil && ok(s) {
			return nil
		}
		pause := f.Pause
		if pause == nil {
			pause = func(ctx context.Context, d time.Duration) error {
				select {
				case <-time.After(d):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
		if err := pause(ctx, 100*time.Millisecond); err != nil {
			return &afterOK{err}
		}
	}
	return errVerify
}
