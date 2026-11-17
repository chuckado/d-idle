// Package vaultkeep is the Keep queue of the Ethereal Vault (docs/vault.md, section Keep; slice V5 of
// docs/design/loot-vault-2026-11-16.md): it turns the records the player marked Keep into real items in the hero's stash
// at the next safe moment in town, exactly as dropped, and records the outcome in the vault log.
//
// Order per item (every step is a log line of the store, so a crash anywhere resumes, never creates twice):
//
//	Keep (keepq) -> BeginKeep -> item create (key=vk-<id>) -> read-back (item full == recorded bytes) -> save -> kept
//
// The create is logged BEFORE it is sent. A record that is still in flight after a restart, a failed save or a transport
// error is resolved first, by looking at the hero (Keeper.present), never by creating again.
package vaultkeep

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// Reasons shown on cards and rows; each is a specific cause, never a generic failure.
const (
	ReasonStashFull = "the stash is full: free space or upgrade"
	ReasonPackFull  = "the backpack is full"
	ReasonDead      = "the hero is dead"
	ReasonNoGame    = "not in a game"
	ReasonVerbs     = "this d2d build has no vault verbs (start it with --vault-verbs)"
	ReasonBusy      = "the game is busy"
	ReasonBadData   = "the game does not accept this item's data"
	ReasonVersion   = "the item was recorded by another game version"
	ReasonReadback  = "the created item differs from the recorded one"
	ReasonDup       = "the game already holds an item created with this key but it is not in the hero"
	ReasonOff       = "Keep is switched off (loot.vaultKeep 0, loot.vault 0 or a measured run)"
	ReasonBackup    = "the save backup failed"
	ReasonSave      = "the hero could not be saved"
	reasonPrefix    = "keep: "
	defaultPerVisit = 20
	defaultBudget   = 8 * time.Second
	keyPrefix       = "vk-"
)

// Reason strips the log prefix from a stored reason.
func Reason(stored string) string { return strings.TrimPrefix(stored, reasonPrefix) }

func stored(why string) string { return reasonPrefix + why }

// Keeper is the Keep queue and its executor for one hero. Zero values of the optional fields are usable; Store and API
// are required. All callbacks run unlocked and may block for one verb round trip.
type Keeper struct {
	Store    func() *vault.Store         // the hero's vault (nil: none yet)
	API      d2dctl.VaultAPI             // the fork verbs (a VaultFake offline)
	Enabled  func() bool                 // loot.vault and loot.vaultKeep on and no measured run; nil counts as on
	Ready    func() string               // "" when the hero is alive, in town, no blocking panel, cursor free; else the reason
	Backup   func(context.Context) error // before the first create of a session (the fork writes its .bak); required
	Save     func(context.Context) error // persists the created items (the fork's save verb with read-back); required
	Notice   func(string)                // a Needs-you info line, once per state change
	Now      func() time.Time
	PerVisit func() int    // loot.vaultKeepPerVisit
	Budget   time.Duration // most wall time of one Run before it yields (default 8 s)

	mu       sync.Mutex
	backedUp bool
	tripped  string // a read-back mismatch switches creating off for the process
	info     Info
	tripWant bool
	noticed  map[string]bool
	caps     d2dctl.VaultCaps // cached for capsTTL: the page polls, the socket is shared
	capsAt   time.Time
	capsOK   bool
}

const capsTTL = 60 * time.Second

// vaultCaps answers from the cache when it is fresh. With fetch false it never touches the socket (the page's poll).
func (k *Keeper) vaultCaps(ctx context.Context, fetch bool) (d2dctl.VaultCaps, bool, error) {
	k.mu.Lock()
	c, at, ok := k.caps, k.capsAt, k.capsOK
	k.mu.Unlock()
	if ok && k.now().Sub(at) < capsTTL || !fetch {
		return c, ok, nil
	}
	c, err := k.API.VaultCaps(ctx)
	if err != nil {
		return c, false, err
	}
	k.mu.Lock()
	k.caps, k.capsAt, k.capsOK = c, k.now(), true
	k.mu.Unlock()
	return c, true, nil
}

// Info is the visible state of the queue for the progress line and the Needs-you row.
type Info struct {
	Queued   int    `json:"queued"`            // waiting, not started
	Creating int    `json:"creating"`          // started and not confirmed
	Waiting  string `json:"waiting,omitempty"` // why the queue does not move right now
	Full     bool   `json:"full,omitempty"`    // the stash (or pack) has no room: a Needs-you row
	Tripped  string `json:"tripped,omitempty"` // creating is off until restart, and why
	Done     int    `json:"done"`              // created in this process
}

func (k *Keeper) now() time.Time {
	if k.Now != nil {
		return k.Now()
	}
	return time.Now()
}

func (k *Keeper) enabled() bool {
	k.mu.Lock()
	t := k.tripped
	k.mu.Unlock()
	return t == "" && (k.Enabled == nil || k.Enabled())
}

func (k *Keeper) perVisit() int {
	if k.PerVisit != nil {
		if n := k.PerVisit(); n > 0 {
			return n
		}
	}
	return defaultPerVisit
}

func (k *Keeper) store() *vault.Store {
	if k.Store == nil {
		return nil
	}
	return k.Store()
}

func (k *Keeper) note(key, text string) {
	k.mu.Lock()
	if k.noticed == nil {
		k.noticed = map[string]bool{}
	}
	first := !k.noticed[key]
	k.noticed[key] = true
	k.mu.Unlock()
	if first && k.Notice != nil && text != "" {
		k.Notice(text)
	}
}

func (k *Keeper) clearNote(key string) {
	k.mu.Lock()
	delete(k.noticed, key)
	k.mu.Unlock()
}

func (k *Keeper) setWaiting(why string) {
	k.mu.Lock()
	k.info.Waiting = why
	k.mu.Unlock()
}

// Gate says whether a Keep press can be queued now. It refuses only what cannot become true by waiting for the next town
// visit: Keep off, a dead hero, a build without the verbs. Not being in town is not a refusal (the item queues).
func (k *Keeper) Gate(ctx context.Context) (ok bool, why string) { return k.gate(ctx, true) }

// Peek is Gate without any round trip: the cached capabilities (unknown counts as fine). The page's poll uses it.
func (k *Keeper) Peek() (ok bool, why string) { return k.gate(context.Background(), false) }

func (k *Keeper) gate(ctx context.Context, fetch bool) (ok bool, why string) {
	k.mu.Lock()
	t := k.tripped
	k.mu.Unlock()
	if t != "" {
		return false, "Keep stopped for this run: " + t
	}
	if !k.enabled() {
		return false, ReasonOff
	}
	if k.API != nil {
		if caps, known, err := k.vaultCaps(ctx, fetch); err == nil && known && !caps.ItemCreate {
			return false, ReasonVerbs
		}
	}
	if k.Ready != nil {
		if w := k.Ready(); w == ReasonDead {
			return false, ReasonDead
		}
	}
	return true, ""
}

// Result is the outcome of queueing one id.
type Result struct {
	ID      string
	OK      bool
	Already bool   // it was queued or kept already: a double press changes nothing
	Code    string // missing, state, target
	Why     string
}

// Enqueue queues records for creation (target stash or inv; "" is the stash). It is idempotent: a record that is queued or
// kept answers OK with Already. wantNow also asks the bot for an errand trip home (Keep now); it never creates anything
// by itself. The create happens in Run, at a safe moment.
func (k *Keeper) Enqueue(ctx context.Context, ids []string, target string, wantNow bool) []Result {
	st := k.store()
	out := make([]Result, 0, len(ids))
	for _, id := range ids {
		r := Result{ID: id}
		switch {
		case st == nil:
			r.Code, r.Why = "missing", "there is no vault for this hero yet"
		default:
			rec, found := st.Get(id)
			switch {
			case !found:
				r.Code, r.Why = "missing", "no such item"
			case rec.State == vault.KeepQ || rec.State == vault.Kept:
				r.OK, r.Already = true, true
			default:
				if err := st.KeepTo(id, target); err != nil {
					r.Code, r.Why = codeOf(err), err.Error()
				} else {
					r.OK = true
				}
			}
		}
		out = append(out, r)
	}
	if wantNow {
		for _, r := range out {
			if r.OK {
				k.mu.Lock()
				k.tripWant = true
				k.mu.Unlock()
				break
			}
		}
	}
	return out
}

func codeOf(err error) string {
	switch {
	case errors.Is(err, vault.ErrNotFound):
		return "missing"
	case errors.Is(err, vault.ErrState):
		return "state"
	case errors.Is(err, vault.ErrInvalid):
		return "target"
	}
	return "internal"
}

// Pending is the number of records waiting for a create (queued or in flight).
func (k *Keeper) Pending() int {
	if st := k.store(); st != nil {
		return len(st.KeepQueued())
	}
	return 0
}

// TripWanted reports whether the player pressed Keep now and the hero is not yet home; Run clears it.
func (k *Keeper) TripWanted() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.tripWant && k.Pending() > 0
}

// Info reads the visible state.
func (k *Keeper) Info() Info {
	k.mu.Lock()
	in := k.info
	in.Tripped = k.tripped
	k.mu.Unlock()
	in.Queued, in.Creating = 0, 0
	if st := k.store(); st != nil {
		for _, r := range st.KeepQueued() {
			if r.InFlight {
				in.Creating++
			} else {
				in.Queued++
			}
		}
	}
	return in
}

type created struct {
	rec   vault.Rec
	place string
}

// Run creates queued items, bounded by PerVisit and Budget, and returns one log line per notable event. It never blocks
// longer than Budget plus one verb round trip. Call it only at a safe moment in town; it checks Ready itself.
func (k *Keeper) Run(ctx context.Context) []string {
	st := k.store()
	if st == nil || len(st.KeepQueued()) == 0 {
		k.setWaiting("")
		return nil
	}
	if !k.enabled() {
		k.setWaiting(ReasonOff)
		return nil
	}
	if k.Ready != nil {
		if why := k.Ready(); why != "" {
			k.setWaiting(why)
			return nil
		}
	}
	if caps, _, err := k.vaultCaps(ctx, true); err != nil {
		k.setWaiting(ReasonBusy)
		return nil
	} else if !caps.ItemCreate || !caps.ItemFull {
		k.setWaiting(ReasonVerbs)
		k.note("verbs", "Keep is waiting: "+ReasonVerbs)
		return nil
	}
	k.clearNote("verbs")
	k.setWaiting("")
	k.mu.Lock()
	k.tripWant = false
	k.mu.Unlock()

	var lines []string
	var confirm []created
	deadline := k.now().Add(orDur(k.Budget, defaultBudget))
	stop := false

	// 1. what a restart or a failed save left in flight: look at the hero first
	var pack []d2dctl.VaultRecord
	havePack := false
	look := func() ([]d2dctl.VaultRecord, error) {
		if !havePack {
			p, err := k.API.ItemFull(ctx, -1)
			if err != nil {
				return nil, err
			}
			pack, havePack = p, true
		}
		return pack, nil
	}
	for _, r := range st.KeepQueued() {
		if !r.InFlight {
			continue
		}
		have, err := look()
		if err != nil {
			k.setWaiting(ReasonBusy)
			return append(lines, "keep: cannot read the hero, in-flight items stay as they are: "+err.Error())
		}
		if k.presentIn(st, r, have) {
			confirm = append(confirm, created{r, placeOf(r)})
		} else if err := st.RetryKeep(r.ID, ""); err != nil {
			lines = append(lines, "keep: "+r.ID+" requeue: "+err.Error())
		} else {
			lines = append(lines, "keep: "+r.ID+" was not created before the restart, queued again")
		}
	}

	// 2. create
	k.mu.Lock()
	k.info.Full = false // set again below when the stash refuses
	k.mu.Unlock()
	made := 0
	for _, r := range st.KeepQueued() {
		if stop || ctx.Err() != nil {
			break
		}
		if r.InFlight {
			continue // confirmed above, or unresolved: not ours to create again
		}
		if made >= k.perVisit() || !k.now().Before(deadline) {
			break
		}
		// the kill switch and the safe moment are judged again before every create, not only at the start
		if !k.enabled() {
			k.setWaiting(ReasonOff)
			break
		}
		if k.Ready != nil {
			if why := k.Ready(); why != "" {
				k.setWaiting(why)
				break
			}
		}
		if !k.backupOnce(ctx, &lines) {
			stop = true
			break
		}
		made++
		c, line, halt := k.createOne(ctx, st, r)
		if line != "" {
			lines = append(lines, line)
		}
		if c != nil {
			confirm = append(confirm, *c)
		}
		stop = halt
	}

	// 3. persist, then close the books
	if len(confirm) > 0 {
		if err := k.Save(ctx); err != nil {
			k.setWaiting(ReasonSave)
			k.note("save", "Keep: "+ReasonSave+" ("+err.Error()+"); the items stay in the queue's in-flight state and are confirmed after the next save")
			return append(lines, "keep: save failed, nothing confirmed: "+err.Error())
		}
		k.clearNote("save")
		for _, c := range confirm {
			if err := st.CompleteKeepAt(c.rec.ID, c.place); err != nil {
				lines = append(lines, "keep: "+c.rec.ID+" complete: "+err.Error())
				continue
			}
			k.mu.Lock()
			k.info.Done++
			k.mu.Unlock()
			lines = append(lines, fmt.Sprintf("keep: %s kept, %s", c.rec.Name, c.place))
		}
	}
	if !k.Info().Full {
		k.clearNote("full") // the next full stash is a new episode
	}
	return lines
}

func orDur(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

func placeOf(r vault.Rec) string {
	if r.Where == vault.WhereInv {
		return "backpack"
	}
	return "stash"
}

// presentIn reports whether the hero carries the record: more items with its bytes than kept records of the same bytes
// the vault already counts (two identical items are legitimate; one already kept must not stand in for this one).
func (k *Keeper) presentIn(st *vault.Store, r vault.Rec, have []d2dctl.VaultRecord) bool {
	n := 0
	for _, h := range have {
		if strings.EqualFold(h.Norm, r.Raw) {
			n++
		}
	}
	kept := 0
	for _, o := range st.Dump() {
		if o.State == vault.Kept && o.ID != r.ID && strings.EqualFold(o.Raw, r.Raw) {
			kept++
		}
	}
	return n > kept
}

// backupOnce takes the save backup before the first create of the session; a failure stops the batch (fail closed).
func (k *Keeper) backupOnce(ctx context.Context, lines *[]string) bool {
	k.mu.Lock()
	done := k.backedUp
	k.mu.Unlock()
	if done {
		return true
	}
	if k.Backup != nil {
		if err := k.Backup(ctx); err != nil {
			k.setWaiting(ReasonBackup)
			k.note("backup", "Keep is waiting: "+ReasonBackup+" ("+err.Error()+")")
			*lines = append(*lines, "keep: backup failed, no item created: "+err.Error())
			return false
		}
	}
	k.clearNote("backup")
	k.mu.Lock()
	k.backedUp = true
	k.mu.Unlock()
	return true
}

// createOne runs BeginKeep, the create and the read-back for one record. It returns the item to confirm after the save, a
// log line, and whether the batch must stop.
func (k *Keeper) createOne(ctx context.Context, st *vault.Store, r vault.Rec) (*created, string, bool) {
	fail := func(why string) (*created, string, bool) {
		if err := st.FailKeep(r.ID, stored(why)); err != nil {
			return nil, "keep: " + r.ID + " fail: " + err.Error(), true
		}
		return nil, fmt.Sprintf("keep: %s not kept: %s", r.Name, why), false
	}
	wait := func(why string, halt bool) (*created, string, bool) {
		_ = st.RetryKeep(r.ID, stored(why))
		k.setWaiting(why)
		return nil, fmt.Sprintf("keep: %s waits: %s", r.Name, why), halt
	}
	if err := d2dctl.ValidHex(r.Raw); err != nil {
		return fail(ReasonBadData)
	}
	if err := st.BeginKeep(r.ID); err != nil {
		return nil, "keep: " + r.ID + " begin: " + err.Error(), true
	}
	o := d2dctl.CreateOpts{Pack: r.Where == vault.WhereInv, Key: keyPrefix + r.ID}
	res, err := k.API.ItemCreate(ctx, r.Raw, o)
	if err != nil {
		switch {
		case errors.Is(err, d2dctl.ErrStashFull):
			why := ReasonStashFull
			if o.Pack {
				why = ReasonPackFull
			}
			k.mu.Lock()
			k.info.Full = true
			k.mu.Unlock()
			k.note("full", "Keep: "+why+". The item stays queued.")
			return wait(why, true)
		case errors.Is(err, d2dctl.ErrVaultState):
			return wait(deadOrNoGame(err), true)
		case errors.Is(err, d2dctl.ErrVaultOff), errors.Is(err, d2dctl.ErrVaultUnsupported):
			return wait(ReasonVerbs, true)
		case errors.Is(err, d2dctl.ErrBusy):
			return wait(ReasonBusy, true) // rejected before it ran: nothing was created
		case errors.Is(err, d2dctl.ErrBadHex):
			return fail(ReasonBadData)
		case errors.Is(err, d2dctl.ErrVaultVersion):
			return fail(ReasonVersion)
		case errors.Is(err, d2dctl.ErrVaultDup):
			have, lerr := k.API.ItemFull(ctx, -1)
			if lerr == nil && k.presentIn(st, r, have) {
				return &created{r, placeOf(r)}, "", false
			}
			if lerr != nil {
				return nil, "keep: " + r.ID + " dup, hero unreadable: " + lerr.Error(), true
			}
			return fail(ReasonDup)
		}
		var vr *d2dctl.VaultRefused
		if errors.As(err, &vr) {
			return fail(vr.Why)
		}
		// transport error or timeout: whether the item was created is unknown; it stays in flight and the next Run looks
		return nil, "keep: " + r.Name + " create unanswered, left in flight: " + err.Error(), true
	}

	back, err := k.API.ItemFull(ctx, res.ID)
	if err != nil || len(back) == 0 {
		return nil, fmt.Sprintf("keep: %s created (unit %d) but the read-back failed, left in flight", r.Name, res.ID), true
	}
	if !strings.EqualFold(back[0].Norm, r.Raw) {
		why := ReasonReadback
		if _, uerr := k.API.ItemVault(ctx, res.ID); uerr != nil {
			why += "; it could not be removed again, check the stash"
		} else {
			why += "; the new copy was removed again"
		}
		k.mu.Lock()
		k.tripped = "read-back mismatch"
		k.mu.Unlock()
		k.note("tripped", "Keep stopped for this run: a created item did not match the recorded one. Nothing else is created until the bot restarts.")
		_, line, _ := fail(why)
		return nil, line, true
	}
	place := fmt.Sprintf("%s %d,%d", placeOf(r), res.Col, res.Row)
	return &created{r, place}, "", false
}

func deadOrNoGame(err error) string {
	var vr *d2dctl.VaultRefused
	if errors.As(err, &vr) && vr.Code == d2dctl.VaultErrNoGame {
		return ReasonNoGame
	}
	return ReasonDead
}

// NeedsText is the Needs-you sentence, "" when nothing needs the player.
func (k *Keeper) NeedsText() string {
	in := k.Info()
	switch {
	case in.Tripped != "":
		return "Keep is stopped: " + in.Tripped + ". Restart the bot after checking the stash."
	case in.Full && in.Queued+in.Creating > 0:
		return fmt.Sprintf("Stash full: free space or upgrade. %d kept item(s) wait.", in.Queued+in.Creating)
	}
	return ""
}

// ProgressLine is the sentence under the toolbar of the vault page, "" when the queue is empty.
func (k *Keeper) ProgressLine() string {
	in := k.Info()
	n := in.Queued + in.Creating
	if n == 0 {
		return ""
	}
	s := fmt.Sprintf("Keeping %d item%s", n, plural(n))
	switch {
	case in.Full:
		return s + ": the stash is full, free space or upgrade"
	case in.Waiting != "":
		return s + ": waiting, " + in.Waiting
	}
	return s + ": arrives at your next town visit"
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
