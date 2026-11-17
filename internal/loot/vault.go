package loot

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/lootrules"
	"d2lod-bot/internal/merc"
	"d2lod-bot/internal/town/hand"
	"d2lod-bot/internal/vault"
)

// The Ethereal Vault at the ground stage (docs/vault.md, slice V3 of docs/design/loot-vault-2026-11-16.md): for an item
// the Vault preset says `vault`, the looter walks next to it as for a pickup, PEEKS it (`ground full`), records it in the
// vault store and sends no pickup click: no pack space is used and the item stays on the floor. Default OFF (loot.vault 0);
// with 0 nothing here runs and every decision is byte-identical to today's.

// VaultRecorder is the part of *vault.Store the looter uses.
type VaultRecorder interface {
	Add(vault.Item) (*vault.Rec, error)
}

// VaultWiring connects a Looter to the fork verbs and the hero's vault store.
type VaultWiring struct {
	API    d2dctl.VaultAPI
	Store  func() (VaultRecorder, error) // opened lazily, once the hero is known
	Notice func(text string)             // a Needs-you info line (called once per state change); may be nil
}

// vaultBlockedFor is how long a full vault stays blocked before one more insert is tried (a discard may have freed room).
const vaultBlockedFor = 2 * time.Minute

// vaultKey names one floor unit: unit ids repeat from game to game, the spot tells them apart.
type vaultKey struct {
	id   uint32
	x, y int
}

func keyOf(u d2.Unit) vaultKey { return vaultKey{u.ID, u.X, u.Y} }

type vaultState struct {
	mu       sync.Mutex
	w        *VaultWiring
	checked  bool   // caps were read
	inert    string // why the feature is off for this process ("" = working)
	blocked  time.Time
	done     map[vaultKey]bool // recorded or duplicate: never walk to it again
	fell     map[vaultKey]bool // peek or record failed: today's pickup decision applies
	logged   map[string]bool
	recorded int
	dups     int
	fallback int
	eng      *Engine
	engAll   bool
	now      func() time.Time
	measured func() bool
}

func (v *vaultState) init() {
	if v.done == nil {
		v.done, v.fell, v.logged = map[vaultKey]bool{}, map[vaultKey]bool{}, map[string]bool{}
	}
}

// measuredRun is the existing measured-run detection (gates, soaks, baselines and drills set D2BOT_RUN; see
// cmd/d2bot/gearpolicy.go and invverbs.go).
func measuredRun() bool { return os.Getenv("D2BOT_RUN") != "" }

// UseVault wires the vault verbs and store. Without it loot.vault does nothing.
func (l *Looter) UseVault(w *VaultWiring) {
	l.vault.mu.Lock()
	l.vault.w = w
	l.vault.mu.Unlock()
}

// SetVaultSeams replaces the clock and the measured-run test (tests).
func (l *Looter) SetVaultSeams(now func() time.Time, measured func() bool) {
	l.vault.mu.Lock()
	l.vault.now, l.vault.measured = now, measured
	l.vault.mu.Unlock()
}

func (v *vaultState) clock() time.Time {
	if v.now != nil {
		return v.now()
	}
	return time.Now()
}

func (v *vaultState) isMeasured() bool {
	if v.measured != nil {
		return v.measured()
	}
	return measuredRun()
}

// once logs and returns true the first time key is seen.
func (l *Looter) vaultOnce(key, msg string, kv ...any) {
	v := &l.vault
	v.init()
	if v.logged[key] {
		return
	}
	v.logged[key] = true
	if l.e != nil && l.e.Log != nil {
		l.e.Log.Info(msg, kv...)
	}
}

func (l *Looter) vaultNotice(text string) {
	if w := l.vault.w; w != nil && w.Notice != nil {
		w.Notice(text)
	}
}

// vaultOn is the cheap gate: the knob, the wiring, the measured-run rule, a working fork. It never talks to the game.
func (l *Looter) vaultOn() bool {
	if l == nil || l.e == nil || l.e.Tune == nil || l.e.Tune.Loot.Vault == 0 {
		return false
	}
	v := &l.vault
	v.mu.Lock()
	defer v.mu.Unlock()
	v.init()
	switch {
	case v.w == nil:
		l.vaultOnce("nowire", "loot.vault is on but this backend has no vault verbs: pickups work as before")
		return false
	case v.isMeasured():
		l.vaultOnce("measured", "loot.vault forced off: a measured run is in progress (D2BOT_RUN)")
		return false
	}
	return v.inert == ""
}

// vaultCheck reads the fork's caps once; a build without the peek verb makes the feature inert for good (a Needs-you
// info line, no stall). A transport error leaves the check for the next call.
func (l *Looter) vaultCheck(ctx context.Context) {
	if !l.vaultOn() {
		return
	}
	v := &l.vault
	v.mu.Lock()
	if v.checked {
		v.mu.Unlock()
		return
	}
	api := v.w.API
	v.mu.Unlock()
	caps, err := api.VaultCaps(ctx)
	v.mu.Lock()
	defer v.mu.Unlock()
	switch {
	case err != nil: // not final: the next pickUp call asks again, a failing peek falls back per item
		l.vaultOnce("caps", "vault caps unreadable; pickups work as before for now", "err", err)
	case !caps.GroundFull:
		v.checked, v.inert = true, "verbs absent"
		l.vaultOnce("verbs", "loot.vault is inert: this d2d build lacks the vault verbs (start d2d with --vault-verbs)")
		l.vaultNotice("The vault is on, but this d2d build has no vault verbs. Looting works as before.")
	default:
		v.checked = true
	}
}

// VaultInfo is the visible state of the vault feature for the Needs-you row and the dashboard.
type VaultInfo struct {
	Active   bool   `json:"active"`
	Inert    string `json:"inert,omitempty"`   // why it is off ("verbs absent", "measured run")
	Blocked  bool   `json:"blocked,omitempty"` // over capacity: new records are stopped
	Recorded int    `json:"recorded"`
	Dups     int    `json:"dups"`
	Fallback int    `json:"fallback"` // peeks that failed and fell back to the pickup decision
}

// VaultInfo reports the state.
func (l *Looter) VaultInfo() VaultInfo {
	on := l.vaultOn()
	v := &l.vault
	v.mu.Lock()
	defer v.mu.Unlock()
	in := VaultInfo{Active: on, Inert: v.inert, Recorded: v.recorded, Dups: v.dups, Fallback: v.fallback}
	if !on && l.e != nil && l.e.Tune != nil && l.e.Tune.Loot.Vault != 0 && v.w != nil && v.isMeasured() {
		in.Inert = "measured run"
	}
	in.Blocked = !v.blocked.IsZero() && v.clock().Before(v.blocked)
	return in
}

func (l *Looter) vaultBlockedNow() bool {
	v := &l.vault
	v.mu.Lock()
	defer v.mu.Unlock()
	return !v.blocked.IsZero() && v.clock().Before(v.blocked)
}

func (l *Looter) vaultSeen(u d2.Unit) (done, fell bool) {
	v := &l.vault
	v.mu.Lock()
	defer v.mu.Unlock()
	v.init()
	k := keyOf(u)
	return v.done[k], v.fell[k]
}

// vaultEngine is the engine of the Vault preset, rebuilt when the record-everything toggle changes.
func (l *Looter) vaultEngine() *Engine {
	all := l.e.Tune.Loot.VaultAll != 0
	v := &l.vault
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.eng == nil || v.engAll != all {
		v.eng, v.engAll = NewEngine(lootrules.VaultPreset(all)), all
	}
	return v.eng
}

// VaultDecision is the ground decision under loot.vault 1: the Vault preset's verdict in the pickit's vocabulary, with
// vault=true when the preset says record it. ok=false (no pass, vault off, engine failed) keeps today's verdict.
func (p *ShadowPass) VaultDecision(u d2.Unit, stats d2.Stats) (vaulted bool, v Verdict, ok bool) {
	if p == nil || !p.vault || u.Item == nil {
		return false, v, false
	}
	it, err := p.item(ShadowGround, u, "", false, stats)
	if err == nil {
		it.Place = lootrules.PlaceGround
		eng := p.l.vaultEngine()
		var d lootrules.Decision
		if d, err = eng.decide(&it, p.cx); err == nil {
			if d.Action == lootrules.Vault {
				return true, Verdict{Action: Keep, Tier: d.Tier, Prio: d.Prio, Rule: d.Rule, Reason: d.Why}, true
			}
			return false, p.groundVerdict(d), true
		}
		eng.fallback(p.l.e.Log, err)
	}
	return false, v, false
}

var (
	errVaultRetry = errors.New("loot: vault state changed, listing again") // not a failure of the item
)

// vaultOne walks next to the cand like a pickup, peeks it and records it. It sends no pickup click. A nil error means the
// item is handled (recorded or already in the vault); errVaultRetry means the state changed (fell back to the pickup
// decision, or the vault is full) and the caller lists the candidates again.
func (l *Looter) vaultOne(ctx context.Context, c cand, o PickOpts) (recorded bool, err error) {
	t := &l.e.Tune.Loot
	if o.Abort != nil && o.Abort() {
		return false, errAborted
	}
	s, err := l.hand.Fresh()
	if err != nil {
		return false, err
	}
	u, on := groundItem(s, c.u.ID)
	if !on {
		return false, errGone
	}
	if d := math.Hypot(float64(u.X-s.Player.X), float64(u.Y-s.Player.Y)); d > float64(t.WalkDist) {
		if err := l.mv.GoTo(ctx, input.Point{X: u.X, Y: u.Y}); err != nil {
			return false, fmt.Errorf("walk: %w", err)
		}
	}
	if hand.PanelOpen(s.UI) {
		if err := l.hand.CloseAll(ctx); err != nil {
			return false, err
		}
	}
	v := &l.vault
	fall := func(why string, err error) (bool, error) {
		v.mu.Lock()
		v.init()
		v.fell[keyOf(u)] = true
		v.fallback++
		v.mu.Unlock()
		l.vaultOnce("fall:"+why, "vault: "+why+"; this item falls back to the pickup decision (logged once)", "item", c.name, "err", err)
		return false, errVaultRetry
	}
	recs, err := v.w.API.GroundFull(ctx, int(u.ID))
	switch {
	case ctx.Err() != nil:
		return false, ctx.Err()
	case errors.Is(err, d2dctl.ErrVaultUnsupported), errors.Is(err, d2dctl.ErrVaultOff):
		v.mu.Lock()
		v.inert = "verbs absent"
		v.mu.Unlock()
		l.vaultNotice("The vault is on, but d2d refused the vault verbs. Looting works as before.")
		return fall("the fork refuses the vault verbs", err)
	case err != nil:
		return fall("peek failed", err)
	}
	var rec *d2dctl.VaultRecord
	for i := range recs {
		if recs[i].UnitID == int(u.ID) {
			rec = &recs[i]
			break
		}
	}
	if rec == nil {
		return fall("peek returned no record", nil)
	}
	st, err := v.w.Store()
	if err != nil {
		return fall("vault store unavailable", err)
	}
	it := vault.Item{Raw: rec.Norm, Code: rec.Code, Name: c.name, Quality: rec.Quality, Ilvl: rec.ILvl, Identified: rec.Identified,
		Sockets: rec.Sockets, Eth: rec.Ethereal, Found: vault.Found{Time: v.clock().UTC(), Zone: "area " + strconv.Itoa(s.Player.Area)}}
	if b := l.base(u.Class); b != nil && l.e.Data != nil {
		it.Chain = merc.TypeChain(l.e.Data, b.Type)
	}
	_, err = st.Add(it)
	v.mu.Lock()
	defer v.mu.Unlock()
	v.init()
	switch {
	case err == nil:
		v.done[keyOf(u)] = true
		v.recorded++
		return true, nil
	case errors.Is(err, vault.ErrDuplicate):
		v.done[keyOf(u)] = true
		v.dups++
		return false, nil
	case errors.Is(err, vault.ErrOverCapacity), errors.Is(err, vault.ErrFull):
		first := v.blocked.IsZero() || !v.clock().Before(v.blocked)
		v.blocked = v.clock().Add(vaultBlockedFor)
		if first {
			l.vaultOnce("full:"+v.blocked.Format(time.RFC3339), "vault full: new records are stopped until room is made", "err", err)
			l.vaultNotice("The vault is full: new items are not recorded. Keep, sell or discard some to make room.")
		}
		return false, errVaultRetry
	}
	v.fell[keyOf(u)] = true
	v.fallback++
	l.vaultOnce("add", "vault: record failed; this item falls back to the pickup decision", "err", err)
	return false, errVaultRetry
}
