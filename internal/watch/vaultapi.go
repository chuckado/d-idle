package watch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	item "d2lod-bot/internal/status"
	"d2lod-bot/internal/vault"
)

// The Ethereal Vault page and its routes (docs/vault.md, UI section; design docs/design/loot-vault-2026-11-16.md section 6, slice V4).
// GET /api/vault and GET /api/vault/item/<id> on the viewer; POST /api/vault/{pin,unpin,discard} on the answer listener with the
// same gate as the other answer routes. Read side first: no Keep, Reveal or Sell here (V5, V6), no wallet (V7). The desk is
// late-bound; without a desk, with the knob loot.vault at 0 or with no vault on disk yet the page still opens and says so, and
// every POST answers 409 code "off".

const (
	VaultRoute        = "/api/vault"
	VaultItemRoute    = "/api/vault/item/" // GET /api/vault/item/<id>
	VaultPinRoute     = "/api/vault/pin"
	VaultUnpinRoute   = "/api/vault/unpin"
	VaultDiscardRoute = "/api/vault/discard"
	VaultKeepRoute    = "/api/vault/keep"
	VaultRevealRoute  = "/api/vault/reveal"
	VaultSellRoute    = "/api/vault/sell"
	VaultCancelRoute  = "/api/vault/cancel"
	VaultPromptRoute  = "/api/vault/sellprompt" // POST {"dismiss":true}: mute the "short on gold" prompt
	vaultSchema       = "watch.vault/1"

	vaultPageLimit = 60  // rows per page when the client names none
	vaultMaxLimit  = 200 // rows per page at most
	vaultMaxIDs    = 150 // ids in one POST (the answer body is capped at 4 KiB: about 20 bytes an id); the page sends bigger selections in chunks
)

// VaultRoutes are the viewer paths (the page and its files, the two GET routes); VaultPostRoutes the answer ones.
var (
	VaultRoutes     = []string{VaultRoute, VaultItemRoute, "/watch/vault", "/watch/vault.js", "/watch/vault-logic.js", "/watch/vault.css", "/watch/vault.json"}
	VaultPostRoutes = []string{VaultPinRoute, VaultUnpinRoute, VaultDiscardRoute, VaultKeepRoute, VaultRevealRoute, VaultSellRoute, VaultCancelRoute, VaultPromptRoute}
)

// VaultDesk is what the routes need from the bot side (cmd/d2bot/vaultwire.go). Every field is a func read per request, so a
// hero switch or a knob change shows at once.
type VaultDesk struct {
	On    func() bool         // the knob loot.vault; nil counts as on
	Store func() *vault.Store // the active hero's vault, nil while there is none
	Tiers *vault.Tiers        // the capacity table, for the tier line
	Owned func() []string     // purchased capacity steps (V7); nil: none
	Keep  VaultKeeper         // the Keep queue (V5); nil: Keep is not part of this bot
	Act   VaultActor          // Reveal, Sell and the sell prompt (V6); nil: not part of this bot
	Tip   VaultTipFunc        // the tooltip facts of a row (base type, art, stat lines); nil: rows carry none
}

// VaultTip is what the inventory tooltip needs beyond the stored fields: the same fields status.Item carries (base facts, art
// key, kind, sell price, stat lines), so inventory-tip.js draws a vault row unchanged.
type VaultTip struct {
	Base  *item.ItemBase
	Art   string
	Tint  string
	Kind  string
	Sell  int
	Props []string
}

// VaultTipFunc builds a row's tooltip facts. props is "" for an unidentified record: the caller of this func never hands
// out what identify costs a charge for, and vaultRow drops Props of an unidentified row again after the call.
type VaultTipFunc func(r VaultTipIn) VaultTip

// VaultTipIn is what the tip source may read of a record. Props is empty for an unidentified record.
type VaultTipIn struct {
	Code                 string
	Quality              int
	Identified, Ethereal bool
	Props                string
	Pic, FileIndex       int
	Tint                 string
}

// VaultActor is Reveal and Sell as the routes see them (cmd/d2bot/vaultkeep.go adapts internal/vaultkeep).
type VaultActor interface {
	GateReveal(ctx context.Context) (ok bool, code, why string)
	EnqueueReveal(ctx context.Context, ids []string, now bool) []VaultKeepResult
	GateSell(ctx context.Context) (ok bool, code, why string)
	EnqueueSell(ctx context.Context, ids []string, now bool) []VaultKeepResult
	Cancel(ids []string) []VaultKeepResult
	DismissOffer()
	Act(ctx context.Context) VaultActView
}

// VaultActView is the state of Reveal and Sell in GET /api/vault: the price shown before the click, what is queued, the
// sentences under the toolbar and the opt-in prompt.
type VaultActView struct {
	Reveal VaultRevealView `json:"reveal"`
	Sell   VaultSellView   `json:"sell"`
	Prompt *VaultPrompt    `json:"prompt,omitempty"`
}

type VaultRevealView struct {
	Enabled bool   `json:"enabled"`
	Why     string `json:"why,omitempty"` // when not enabled: shown on the button (code noscroll offers the Buy action)
	Code    string `json:"code,omitempty"`
	Charges int    `json:"charges"` // identify scrolls and tome charges in the pack, -1 unknown
	Free    bool   `json:"free"`    // Cain stands here: Reveal costs nothing
	PerItem int    `json:"perItem"`
	Queued  int    `json:"queued"`
	Working int    `json:"working"`
	Done    int    `json:"done"`
	Line    string `json:"line,omitempty"`
}

type VaultSellView struct {
	Enabled bool   `json:"enabled"`
	Why     string `json:"why,omitempty"`
	Code    string `json:"code,omitempty"`
	Queued  int    `json:"queued"`
	Working int    `json:"working"`
	Sold    int    `json:"sold"`
	Gold    int    `json:"gold"` // earned in this process
	Line    string `json:"line,omitempty"`
}

// VaultPrompt is the opt-in "short on gold: sell these?" offer; it exists only while loot.vaultSellPrompt is on.
type VaultPrompt struct {
	Short int      `json:"short"`
	IDs   []string `json:"ids"`
	Gold  int      `json:"gold"`
	Names []string `json:"names"`
}

// VaultKeeper is the Keep queue as the routes see it (cmd/d2bot/vaultkeep.go adapts internal/vaultkeep).
type VaultKeeper interface {
	// Gate says whether a press can be queued; code is a short word (off, dead, verbs, stopped) the page can branch on.
	Gate(ctx context.Context) (ok bool, code, why string)
	Enqueue(ctx context.Context, ids []string, target string, now bool) []VaultKeepResult
	Progress() VaultKeepView
}

// VaultKeepResult is the outcome of queueing one id.
type VaultKeepResult struct {
	ID      string
	OK      bool
	Already bool
	Code    string
	Why     string
}

// VaultKeepView is the queue's state in GET /api/vault: counts, the progress sentence, and why nothing moves.
type VaultKeepView struct {
	Enabled  bool   `json:"enabled"`
	Why      string `json:"why,omitempty"` // when not enabled: the reason, shown on the button
	Queued   int    `json:"queued"`
	Creating int    `json:"creating"`
	Done     int    `json:"done"`
	Full     bool   `json:"full,omitempty"`
	Line     string `json:"line,omitempty"`
}

type VaultAPI struct {
	mu sync.Mutex
	d  *VaultDesk
}

func NewVaultAPI() *VaultAPI { return &VaultAPI{} }

func (a *VaultAPI) Bind(d VaultDesk) { a.mu.Lock(); a.d = &d; a.mu.Unlock() }

func (a *VaultAPI) desk() *VaultDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// quality ids of the game (value.json documents them) to the words the inventory page and its tooltip use.
var vaultQuality = map[int]string{1: "low", 2: "normal", 3: "superior", 4: "magic", 5: "set", 6: "rare", 7: "unique", 8: "crafted"}

func qualityWord(q int) string {
	if w, ok := vaultQuality[q]; ok {
		return w
	}
	return "normal"
}

func qualityID(w string) (int, bool) {
	if n, err := strconv.Atoi(w); err == nil {
		_, ok := vaultQuality[n]
		return n, ok
	}
	for id, name := range vaultQuality {
		if name == w {
			return id, true
		}
	}
	return 0, false
}

// vaultKind is one entry of the "Kind" select: a set of quality ids and/or item code prefixes, in web/vault.json (data, not code).
type vaultKind struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Quality    []int    `json:"quality,omitempty"`
	CodePrefix []string `json:"codePrefix,omitempty"`
}

type vaultTexts struct {
	Kinds []vaultKind       `json:"kinds"`
	Types map[string]string `json:"types"` // item type code to its label; a code without one shows as itself
}

func vaultData() vaultTexts {
	var t vaultTexts
	if b, err := webFS.ReadFile("web/vault.json"); err == nil {
		_ = json.Unmarshal(b, &t)
	}
	return t
}

// VaultItem is a record as the page draws it: the shape of an inventory item where the tooltip needs it (name, code, quality
// word, identified, ethereal, sockets, level), so inventory-tip.js draws it unchanged, plus the vault's own fields. An
// unidentified record shows only what the game shows: the stored name is the as-displayed one and nothing else of the item
// is sent (no property text exists in the store for it).
type VaultItem struct {
	ID         string         `json:"id"`
	Code       string         `json:"code"`
	Name       string         `json:"name"`
	Quality    string         `json:"quality"`
	Q          int            `json:"q"`
	Ilvl       int            `json:"ilvl"`
	Level      int            `json:"level"`
	Identified bool           `json:"identified"`
	Base       *item.ItemBase `json:"base,omitempty"` // tooltip facts, as in status.Item
	Art        string         `json:"art,omitempty"`
	Tint       string         `json:"tint,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	SellPrice  int            `json:"sellPrice,omitempty"` // the vendor price the tooltip shows (the row's own "sell" is the Sell errand's state)
	Props      []string       `json:"props,omitempty"`     // stat lines of an IDENTIFIED row only (docs/vault.md, identify stays a step)
	Ethereal   bool           `json:"ethereal,omitempty"`
	Sockets    int            `json:"sockets,omitempty"`
	Pin        bool           `json:"pin,omitempty"`
	State      string         `json:"state"`
	Status     string         `json:"status"`
	Value      int            `json:"value"`
	Found      string         `json:"found,omitempty"` // RFC 3339
	Area       string         `json:"area,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	Keep       string         `json:"keep,omitempty"` // queued, creating, kept or failed: the Keep queue's word for the row
	Where      string         `json:"where,omitempty"`
	Reveal     string         `json:"reveal,omitempty"`       // queued, working or failed: the Reveal errand's word for the row
	RevealWhy  string         `json:"revealReason,omitempty"` // why a reveal failed
	Sell       string         `json:"sell,omitempty"`         // queued, selling, sold or failed
	SellWhy    string         `json:"sellReason,omitempty"`
}

const (
	keepReasonPrefix   = "keep: " // vaultkeep stores failure reasons with this prefix; the row shows the sentence
	revealReasonPrefix = "reveal: "
	sellReasonPrefix   = "sell: "
)

func vaultRow(r *vault.Rec, tip VaultTipFunc) VaultItem {
	name := r.Name
	if name == "" {
		name = r.Code
	}
	v := VaultItem{ID: r.ID, Code: r.Code, Name: name, Quality: qualityWord(r.Quality), Q: r.Quality, Ilvl: r.Ilvl, Level: r.Ilvl,
		Identified: r.Identified, Ethereal: r.Eth, Sockets: r.Sockets, Pin: r.Pin, State: string(r.State), Status: r.Status(),
		Value: r.Value, Area: r.Found.Area, Reason: r.Reason}
	if !r.Found.Time.IsZero() {
		v.Found = r.Found.Time.UTC().Format(time.RFC3339)
	}
	if tip != nil {
		props := ""
		if r.Identified {
			props = r.Props
		}
		t := tip(VaultTipIn{Code: r.Code, Quality: r.Quality, Identified: r.Identified, Ethereal: r.Eth, Props: props, Pic: r.Pic, FileIndex: r.FileIndex, Tint: r.Tint})
		v.Base, v.Art, v.Tint, v.Kind, v.SellPrice = t.Base, t.Art, t.Tint, t.Kind, t.Sell
		if r.Identified {
			v.Props = t.Props
		}
	}
	switch {
	case r.State == vault.KeepQ && r.InFlight:
		v.Keep = "creating"
	case r.State == vault.KeepQ:
		v.Keep = "queued"
	case r.State == vault.Kept:
		v.Keep = "kept"
	case r.State == vault.Recorded && strings.HasPrefix(r.Reason, keepReasonPrefix):
		v.Keep = "failed"
	}
	if v.Keep != "" {
		v.Reason, v.Where = strings.TrimPrefix(r.Reason, keepReasonPrefix), r.Where
	}
	switch {
	case (r.State == vault.Recorded || r.State == vault.KeepQ) && r.RevQ && r.Stage != "":
		v.Reveal = "working"
	case (r.State == vault.Recorded || r.State == vault.KeepQ) && r.RevQ:
		v.Reveal = "queued"
	case r.State == vault.Recorded && r.Stage == vault.StageRevCleanup:
		v.Reveal = "working"
	case r.State == vault.Recorded && strings.HasPrefix(r.Reason, revealReasonPrefix):
		v.Reveal, v.RevealWhy = "failed", strings.TrimPrefix(r.Reason, revealReasonPrefix)
		v.Reason = v.RevealWhy
	case r.State == vault.KeepQ && strings.HasPrefix(r.Reason, revealReasonPrefix):
		v.Reveal, v.RevealWhy = "failed", strings.TrimPrefix(r.Reason, revealReasonPrefix)
	}
	switch {
	case r.State == vault.SellQ && r.Stage != "":
		v.Sell = "selling"
	case r.State == vault.SellQ:
		v.Sell = "queued"
	case r.State == vault.Sold:
		v.Sell = "sold"
		if strings.HasPrefix(r.Reason, sellReasonPrefix) {
			v.SellWhy, v.Reason = strings.TrimPrefix(r.Reason, sellReasonPrefix), strings.TrimPrefix(r.Reason, sellReasonPrefix)
		}
	case r.State == vault.Recorded && strings.HasPrefix(r.Reason, sellReasonPrefix):
		v.Sell, v.SellWhy = "failed", strings.TrimPrefix(r.Reason, sellReasonPrefix)
		v.Reason = v.SellWhy
	}
	if r.State == vault.Sold {
		v.Value = r.Gold // the row of a sold item shows what it fetched
	}
	return v
}

// VaultCapacity is the meter. Level is ok, warn (80 percent), full (100) or over (a shrink left more than fits).
type VaultCapacity struct {
	Count    int    `json:"count"`
	Capacity int    `json:"capacity"`
	Pins     int    `json:"pins"`
	PinLimit int    `json:"pinLimit"`
	Percent  int    `json:"percent"`
	Level    string `json:"level"`
	Over     int    `json:"over,omitempty"` // how many records to discard before recording resumes
}

type VaultUpgrade struct {
	Available bool        `json:"available"`
	Currency  string      `json:"currency"`
	Reason    string      `json:"reason"`
	Next      *vault.Step `json:"next,omitempty"`
	Max       int         `json:"max"`
}

func capacityOf(st vault.Status) VaultCapacity {
	c := VaultCapacity{Count: st.Count, Capacity: st.Capacity, Pins: st.Pins, PinLimit: st.PinLimit, Level: "ok"}
	if st.Capacity > 0 {
		c.Percent = st.Count * 100 / st.Capacity
	}
	switch {
	case st.Over:
		c.Level, c.Over = "over", st.Count-st.Capacity
	case st.Full || c.Percent >= 100:
		c.Level = "full"
	case c.Percent >= 80:
		c.Level = "warn"
	}
	return c
}

type vaultFacet struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	N     int    `json:"n"`
}

// VaultView is the reply of GET /api/vault.
type VaultView struct {
	Schema    string         `json:"schema"`
	Mode      string         `json:"mode"`   // "on", or "off" while loot.vault is 0 (read only)
	Exists    bool           `json:"exists"` // a vault is on disk for the hero
	Hero      string         `json:"hero,omitempty"`
	Total     int            `json:"total"`
	Offset    int            `json:"offset"`
	Limit     int            `json:"limit"`
	Items     []VaultItem    `json:"items"`
	Capacity  VaultCapacity  `json:"capacity"`
	Upgrade   VaultUpgrade   `json:"upgrade"`
	Kinds     []vaultKind    `json:"kinds,omitempty"`
	Types     []vaultFacet   `json:"types,omitempty"`
	Anomalies int            `json:"anomalies,omitempty"`
	Keep      *VaultKeepView `json:"keep,omitempty"` // schema additive: absent when the bot has no Keep queue
	Act       *VaultActView  `json:"act,omitempty"`  // Reveal and Sell (V6); absent when the bot has neither
}

type vaultQuery struct {
	f      vault.Filter
	facets bool
}

// parseVaultQuery turns the query string into a vault.Filter; a bad value is a 400 naming it. kinds come from web/vault.json.
func parseVaultQuery(q map[string][]string, kinds []vaultKind) (vaultQuery, error) {
	var o vaultQuery
	get := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	num := func(k string, lo, hi int) (int, error) {
		s := get(k)
		if s == "" {
			return 0, nil
		}
		n, err := strconv.Atoi(s)
		if err != nil || n < lo || n > hi {
			return 0, errors.New("bad " + k)
		}
		return n, nil
	}
	flag := func(k string) (*bool, error) {
		switch get(k) {
		case "":
			return nil, nil
		case "1", "true":
			t := true
			return &t, nil
		case "0", "false":
			f := false
			return &f, nil
		}
		return nil, errors.New("bad " + k)
	}
	f := &o.f
	var err error
	if f.Offset, err = num("offset", 0, 1<<30); err != nil {
		return o, err
	}
	if f.Limit, err = num("limit", 0, vaultMaxLimit); err != nil {
		return o, err
	}
	if f.Limit == 0 {
		f.Limit = vaultPageLimit
	}
	if f.MinIlvl, err = num("ilvlMin", 0, 200); err != nil {
		return o, err
	}
	if f.MaxIlvl, err = num("ilvlMax", 0, 200); err != nil {
		return o, err
	}
	if get("sockets") != "" {
		n, err := num("sockets", 0, 6)
		if err != nil {
			return o, err
		}
		f.Sockets = &n
	}
	if f.Eth, err = flag("eth"); err != nil {
		return o, err
	}
	if f.Identified, err = flag("identified"); err != nil {
		return o, err
	}
	if f.Pinned, err = flag("pinned"); err != nil {
		return o, err
	}
	o.facets = get("facets") == "1"
	if s := get("q"); len(s) > 64 {
		return o, errors.New("q is longer than 64 characters")
	} else {
		f.Name = s
	}
	if s := get("type"); len(s) > 16 {
		return o, errors.New("bad type")
	} else {
		f.Type = s
	}
	for _, w := range strings.Split(get("quality"), ",") {
		if w == "" {
			continue
		}
		id, ok := qualityID(w)
		if !ok {
			return o, errors.New("bad quality " + w)
		}
		f.Quality = append(f.Quality, id)
	}
	if k := get("kind"); k != "" {
		found := false
		for _, kd := range kinds {
			if kd.ID == k {
				f.CodePrefix = kd.CodePrefix
				if len(kd.Quality) > 0 {
					f.Quality = intersect(f.Quality, kd.Quality)
					if len(f.Quality) == 0 {
						f.Quality = []int{-1} // the kind and the quality chips share no quality: nothing matches
					}
				}
				found = true
			}
		}
		if !found {
			return o, errors.New("bad kind " + k)
		}
	}
	for _, s := range strings.Split(get("state"), ",") {
		switch vault.State(s) {
		case "":
		case vault.Recorded, vault.KeepQ, vault.Kept, vault.SellQ, vault.Sold, vault.Discarded, vault.Trimmed:
			f.States = append(f.States, vault.State(s))
		default:
			return o, errors.New("bad state " + s)
		}
	}
	switch s := get("sort"); s {
	case "", "new":
		f.Order = "new"
	case "old", "value", "ilvl", "quality", "name":
		f.Order = s
	default:
		return o, errors.New("bad sort " + s)
	}
	return o, nil
}

func intersect(a, b []int) []int {
	if len(a) == 0 {
		return b
	}
	var out []int
	for _, x := range a {
		for _, y := range b {
			if x == y {
				out = append(out, x)
			}
		}
	}
	return out
}

// keepAction is the card's Keep button: possible for a live recorded item while the queue accepts presses.
func (d *VaultDesk) keepAction(ctx context.Context, r *vault.Rec) VaultAction {
	switch {
	case d.Keep == nil:
		return VaultAction{Reason: "this bot has no Keep queue"}
	case !d.on():
		return VaultAction{Reason: "the vault is read only while loot.vault is 0"}
	case r.State == vault.KeepQ:
		return VaultAction{Reason: "already queued: arrives at your next town visit"}
	case r.State == vault.Kept:
		return VaultAction{Reason: "already in your stash"}
	case r.State != vault.Recorded:
		return VaultAction{Reason: "the item is " + string(r.State)}
	}
	if ok, _, why := d.Keep.Gate(ctx); !ok {
		return VaultAction{Reason: why}
	}
	return VaultAction{OK: true}
}

// revealAction is the card's Reveal button. Cost and the missing-scroll reason come from the gate and the price view.
func (d *VaultDesk) revealAction(ctx context.Context, r *vault.Rec) VaultAction {
	switch {
	case d.Act == nil:
		return VaultAction{Reason: "this bot has no Reveal"}
	case !d.on():
		return VaultAction{Reason: "the vault is read only while loot.vault is 0"}
	case r.Identified:
		return VaultAction{Reason: "it is identified already"}
	case r.RevQ:
		return VaultAction{Reason: "already queued: arrives at your next town visit"}
	case r.State != vault.Recorded && r.State != vault.KeepQ:
		return VaultAction{Reason: "the item is " + string(r.State)}
	}
	if ok, _, why := d.Act.GateReveal(ctx); !ok {
		return VaultAction{Reason: why}
	}
	if av := d.Act.Act(ctx); !av.Reveal.Free && av.Reveal.Charges == 0 {
		return VaultAction{Reason: "no scroll; buy on request"}
	}
	return VaultAction{OK: true}
}

// sellAction is the card's Sell button: a live recorded item that is not pinned.
func (d *VaultDesk) sellAction(ctx context.Context, r *vault.Rec) VaultAction {
	switch {
	case d.Act == nil:
		return VaultAction{Reason: "this bot has no Sell"}
	case !d.on():
		return VaultAction{Reason: "the vault is read only while loot.vault is 0"}
	case r.State == vault.SellQ:
		return VaultAction{Reason: "already queued: sold at your next town visit"}
	case r.Pin:
		return VaultAction{Reason: "pinned items are never sold: unpin it first"}
	case r.State != vault.Recorded:
		return VaultAction{Reason: "the item is " + string(r.State)}
	}
	if ok, _, why := d.Act.GateSell(ctx); !ok {
		return VaultAction{Reason: why}
	}
	return VaultAction{OK: true}
}

func (d *VaultDesk) on() bool { return d.On == nil || d.On() }

func (d *VaultDesk) store() *vault.Store {
	if d.Store == nil {
		return nil
	}
	return d.Store()
}

// upgradeOf is the placeholder of the capacity meter: the next tier's gain and price from the table, never available until the
// Ether wallet exists (V7).
func (d *VaultDesk) upgradeOf() VaultUpgrade {
	u := VaultUpgrade{Currency: "Ether", Reason: "needs the vault switched on (loot.vault) for the Ether wallet"}
	if d.Tiers == nil {
		return u
	}
	u.Max = d.Tiers.Max
	var owned []string
	if d.Owned != nil {
		owned = d.Owned()
	}
	if s, ok := d.Tiers.Next(owned); ok {
		u.Next = &s
	} else {
		u.Reason = "every capacity step is owned"
	}
	return u
}

func (d *VaultDesk) view(ctx context.Context, q vaultQuery, kinds vaultTexts) VaultView {
	v := VaultView{Schema: vaultSchema, Mode: "on", Items: []VaultItem{}, Offset: q.f.Offset, Limit: q.f.Limit, Upgrade: d.upgradeOf()}
	if !d.on() {
		v.Mode = "off"
	}
	if q.facets {
		v.Kinds = kinds.Kinds
	}
	if d.Keep != nil {
		kv := d.Keep.Progress()
		if !d.on() {
			kv.Enabled, kv.Why = false, "the vault is read only while loot.vault is 0"
		}
		v.Keep = &kv
	}
	if d.Act != nil {
		av := d.Act.Act(ctx)
		if !d.on() {
			av.Reveal.Enabled, av.Reveal.Why = false, "the vault is read only while loot.vault is 0"
			av.Sell.Enabled, av.Sell.Why = false, "the vault is read only while loot.vault is 0"
			av.Prompt = nil
		}
		v.Act = &av
	}
	st := d.store()
	if st == nil {
		if d.Tiers != nil {
			v.Capacity = capacityOf(vault.Status{Capacity: d.Tiers.Capacity(nil, 0)})
		}
		return v
	}
	v.Exists, v.Hero = true, st.Hero()
	recs, total := st.Query(q.f)
	v.Total = total
	for i := range recs {
		v.Items = append(v.Items, vaultRow(&recs[i], d.Tip))
	}
	s := st.Status()
	v.Capacity, v.Anomalies = capacityOf(s), s.Anomalies
	if q.facets {
		for id, n := range st.Facets().Types {
			label := kinds.Types[id]
			if label == "" {
				label = id
			}
			v.Types = append(v.Types, vaultFacet{ID: id, Label: label, N: n})
		}
		sortFacets(v.Types)
	}
	return v
}

func sortFacets(l []vaultFacet) {
	for i := 1; i < len(l); i++ {
		for j := i; j > 0 && (l[j].Label < l[j-1].Label); j-- {
			l[j], l[j-1] = l[j-1], l[j]
		}
	}
}

// VaultDetail is the reply of GET /api/vault/item/<id>: the row plus where it was found, the card's lines and what the
// buttons may do. Keep and Reveal are listed as unavailable with their slice, so the page words them from here.
type VaultDetail struct {
	Schema  string                 `json:"schema"`
	Item    VaultItem              `json:"item"`
	Where   map[string]string      `json:"where,omitempty"`
	Lines   []string               `json:"lines"`
	Actions map[string]VaultAction `json:"actions"`
}

type VaultAction struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

func detailOf(r *vault.Rec, tip VaultTipFunc, readOnly bool, keep, reveal, sell VaultAction) VaultDetail {
	it := vaultRow(r, tip)
	var lines []string
	if !it.Identified {
		lines = append(lines, "Unidentified: the properties are hidden until it is revealed")
	}
	lines = append(lines, "Item level "+strconv.Itoa(r.Ilvl))
	if r.Sockets > 0 {
		lines = append(lines, "Sockets: "+strconv.Itoa(r.Sockets))
	}
	if r.Eth {
		lines = append(lines, "Ethereal")
	}
	w := map[string]string{}
	if !r.Found.Time.IsZero() {
		w["time"] = r.Found.Time.UTC().Format(time.RFC3339)
	}
	if r.Found.Area != "" {
		w["area"] = r.Found.Area
	}
	if r.Found.Zone != "" {
		w["zone"] = r.Found.Zone
	}
	if r.Found.MonsterClass != "" {
		w["monster"] = r.Found.MonsterClass
	}
	live := r.State == vault.Recorded
	off := func(ok bool, why string) VaultAction {
		switch {
		case readOnly:
			return VaultAction{Reason: "the vault is read only while loot.vault is 0"}
		case !live:
			return VaultAction{Reason: "the item is " + string(r.State)}
		case !ok:
			return VaultAction{Reason: why}
		}
		return VaultAction{OK: true}
	}
	return VaultDetail{Schema: vaultSchema, Item: it, Where: w, Lines: lines, Actions: map[string]VaultAction{
		"pin":     off(!r.Pin, "already pinned"),
		"unpin":   off(r.Pin, "not pinned"),
		"discard": off(true, ""),
		"keep":    keep,
		"reveal":  reveal,
		"sell":    sell,
	}}
}

// Handler serves the page, its files and the two GET routes.
func (a *VaultAPI) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc(VaultRoute, func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the vault is not available on this bot"})
			return
		}
		td := vaultData()
		q, err := parseVaultQuery(r.URL.Query(), td.Kinds)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "query", Error: err.Error()})
			return
		}
		writeJSON(w, 200, d.view(r.Context(), q, td))
	})
	m.HandleFunc(VaultItemRoute, func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		var st *vault.Store
		if d != nil {
			st = d.store()
		}
		id := strings.TrimPrefix(r.URL.Path, VaultItemRoute)
		if st == nil || id == "" {
			writeJSON(w, http.StatusNotFound, policyReply{Code: "missing", Error: "no such item"})
			return
		}
		rec, ok := st.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, policyReply{Code: "missing", Error: "no such item"})
			return
		}
		writeJSON(w, 200, detailOf(&rec, d.Tip, !d.on(), d.keepAction(r.Context(), &rec), d.revealAction(r.Context(), &rec), d.sellAction(r.Context(), &rec)))
	})
	static := func(path, file, ctype string) {
		m.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			b, err := webFS.ReadFile("web/" + file)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			pactHdr(w, ctype)
			w.Write(b)
		})
	}
	static("/watch/vault", "vault.html", "text/html; charset=utf-8")
	static("/watch/vault.js", "vault.js", "text/javascript; charset=utf-8")
	static("/watch/vault-logic.js", "vault-logic.js", "text/javascript; charset=utf-8")
	static("/watch/vault.css", "vault.css", "text/css; charset=utf-8")
	static("/watch/vault.json", "vault.json", "application/json; charset=utf-8")
	return readOnly(m)
}

type vaultBody struct {
	IDs     []string `json:"ids"`
	Confirm bool     `json:"confirm"`
	Target  string   `json:"target,omitempty"`  // keep: stash (default) or inv
	Now     bool     `json:"now,omitempty"`     // keep, reveal, sell: also ask the bot to walk home for it
	Dismiss bool     `json:"dismiss,omitempty"` // sellprompt: mute the prompt
}

type vaultFail struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

type vaultReply struct {
	OK      bool        `json:"ok"`
	Done    []string    `json:"done"`
	Already []string    `json:"already,omitempty"` // keep: pressed before, nothing changed
	Failed  []vaultFail `json:"failed,omitempty"`
}

func vaultCode(err error) string {
	switch {
	case errors.Is(err, vault.ErrNotFound):
		return "missing"
	case errors.Is(err, vault.ErrState):
		return "state"
	case errors.Is(err, vault.ErrPinCap):
		return "pincap"
	}
	return "internal"
}

// PostHandler serves pin, unpin and discard for one or more ids. Discard needs "confirm": true (the page asks first); the
// store archives the record before it leaves (docs/vault.md). A partly refused batch answers 200 with the failures listed;
// a batch where nothing worked answers 409.
func (a *VaultAPI) PostHandler(o AnswerOptions, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b vaultBody
		if !policyGate(o, w, r, &b) {
			return
		}
		d := a.desk()
		var st *vault.Store
		if d != nil && d.on() {
			st = d.store()
		}
		if st == nil {
			writeJSON(w, http.StatusConflict, policyReply{Code: "off", Error: "the vault is read only (loot.vault is 0 or there is no vault yet)"})
			return
		}
		if route == VaultPromptRoute {
			if d.Act == nil || !b.Dismiss {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "dismiss", Error: "send dismiss: true"})
				return
			}
			d.Act.DismissOffer()
			writeJSON(w, 200, vaultReply{OK: true, Done: []string{}})
			return
		}
		if len(b.IDs) == 0 || len(b.IDs) > vaultMaxIDs {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "ids", Error: "send 1 to " + strconv.Itoa(vaultMaxIDs) + " ids"})
			return
		}
		if route == VaultDiscardRoute && !b.Confirm {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "confirm", Error: "discard needs confirm: true"})
			return
		}
		if route == VaultKeepRoute {
			a.keep(w, r, d, &b)
			return
		}
		if route == VaultRevealRoute || route == VaultSellRoute || route == VaultCancelRoute {
			a.act(w, r, d, route, &b)
			return
		}
		rep := vaultReply{Done: []string{}}
		for _, id := range b.IDs {
			var err error
			switch route {
			case VaultPinRoute:
				err = st.Pin(id)
			case VaultUnpinRoute:
				err = st.Unpin(id)
			default:
				err = st.Discard(id, "player, vault page")
			}
			if err != nil {
				rep.Failed = append(rep.Failed, vaultFail{ID: id, Code: vaultCode(err), Error: err.Error()})
				continue
			}
			rep.Done = append(rep.Done, id)
		}
		rep.OK = len(rep.Failed) == 0
		code := 200
		if len(rep.Done) == 0 {
			code = http.StatusConflict
		}
		writeJSON(w, code, rep)
	})
}

// act queues Reveal or Sell, or cancels queued requests of any kind. Same shape as keep: the gate first, then the real reason
// of every refusal; the page reads the state back from GET. Sell needs confirm (it turns an item into gold).
func (a *VaultAPI) act(w http.ResponseWriter, r *http.Request, d *VaultDesk, route string, b *vaultBody) {
	if d.Act == nil {
		writeJSON(w, http.StatusConflict, policyReply{Code: "off", Error: "this bot has no Reveal or Sell"})
		return
	}
	var res []VaultKeepResult
	switch route {
	case VaultRevealRoute:
		if ok, code, why := d.Act.GateReveal(r.Context()); !ok {
			writeJSON(w, http.StatusConflict, policyReply{Code: code, Error: why})
			return
		}
		res = d.Act.EnqueueReveal(r.Context(), b.IDs, b.Now)
	case VaultSellRoute:
		if !b.Confirm {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "confirm", Error: "sell needs confirm: true"})
			return
		}
		if ok, code, why := d.Act.GateSell(r.Context()); !ok {
			writeJSON(w, http.StatusConflict, policyReply{Code: code, Error: why})
			return
		}
		res = d.Act.EnqueueSell(r.Context(), b.IDs, b.Now)
	default:
		res = d.Act.Cancel(b.IDs)
	}
	rep := vaultReply{Done: []string{}}
	for _, x := range res {
		switch {
		case x.OK && x.Already:
			rep.Already = append(rep.Already, x.ID)
			rep.Done = append(rep.Done, x.ID)
		case x.OK:
			rep.Done = append(rep.Done, x.ID)
		default:
			rep.Failed = append(rep.Failed, vaultFail{ID: x.ID, Code: x.Code, Error: x.Why})
		}
	}
	rep.OK = len(rep.Failed) == 0
	code := 200
	if len(rep.Done) == 0 {
		code = http.StatusConflict
	}
	writeJSON(w, code, rep)
}

// keep queues the ids (idempotent: a second press answers "already"), after the gate. The reply names the real reason of every
// failure; the page reads the state back from GET, so a toast never says more than what the store holds.
func (a *VaultAPI) keep(w http.ResponseWriter, r *http.Request, d *VaultDesk, b *vaultBody) {
	if d.Keep == nil {
		writeJSON(w, http.StatusConflict, policyReply{Code: "off", Error: "this bot has no Keep queue"})
		return
	}
	if b.Target != "" && b.Target != vault.WhereStash && b.Target != vault.WhereInv {
		writeJSON(w, http.StatusBadRequest, policyReply{Code: "target", Error: "target is stash or inv"})
		return
	}
	if ok, code, why := d.Keep.Gate(r.Context()); !ok {
		writeJSON(w, http.StatusConflict, policyReply{Code: code, Error: why})
		return
	}
	rep := vaultReply{Done: []string{}}
	for _, res := range d.Keep.Enqueue(r.Context(), b.IDs, b.Target, b.Now) {
		switch {
		case res.OK && res.Already:
			rep.Already = append(rep.Already, res.ID)
			rep.Done = append(rep.Done, res.ID)
		case res.OK:
			rep.Done = append(rep.Done, res.ID)
		default:
			rep.Failed = append(rep.Failed, vaultFail{ID: res.ID, Code: res.Code, Error: res.Why})
		}
	}
	rep.OK = len(rep.Failed) == 0
	code := 200
	if len(rep.Done) == 0 {
		code = http.StatusConflict
	}
	writeJSON(w, code, rep)
}
