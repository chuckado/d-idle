package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/ether"
	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/watch"
)

// The Ethereal Vault page (internal/watch/web/vault.js, docs/vault.md): the real VaultAPI over real vault stores filled with a
// frozen set of drops, one store per page state. The page's query vault= picks the state (the API calls carry it as the
// Referer, like chars= for the Characters page):
//
//	pop (default) 140 items in 400 slots: every quality, long names, unidentified, pinned, ethereal, socketed, runes
//	warn 140 in 160 (87 percent), full 140 in 140, over 140 in 100 (a shrink), empty (a vault with nothing in it), none (no vault yet),
//	off (loot.vault is 0: read only, items shown), error (the API answers 500), big 1200 items
var (
	vaultOnce sync.Once
	vaultAPIs = map[string]*watch.VaultAPI{}
)

type vaultSet struct {
	n, capacity, shrinkTo int
	on                    bool
	none                  bool
	// the Ether wallet of the page (the eth-* states): the balance left after buying `owned` steps. eth false: no wallet, the old placeholder
	eth            bool
	balance, owned int
	keep           string // "": no Keep queue; ok, dead (the gate refuses: the hero is dead), full (a queue waiting for stash room)
	preset         bool   // queued, creating, kept and failed rows are put in the store
	act            string // "": no Reveal or Sell; ok (3 charges), none (no scroll), cain (free), dead (the gate refuses)
	actPreset      bool   // reveal and sell rows in every state, and the "short on gold" prompt
}

var vaultModes = map[string]vaultSet{
	"pop": {n: 140, capacity: 400, on: true}, "warn": {n: 140, capacity: 160, on: true}, "full": {n: 140, capacity: 140, on: true},
	"over": {n: 140, capacity: 400, shrinkTo: 100, on: true}, "empty": {n: 0, capacity: 400, on: true}, "none": {none: true, on: true},
	"off": {n: 140, capacity: 400, on: false}, "big": {n: 1200, capacity: 2400, on: true},
	// the Keep queue (slice V5): its own stores, so a press on one page never changes another
	"keeping":   {n: 140, capacity: 400, on: true, keep: "ok", preset: true},
	"keeppress": {n: 140, capacity: 400, on: true, keep: "ok"},
	"keepdead":  {n: 140, capacity: 400, on: true, keep: "dead"},
	"keepfull":  {n: 140, capacity: 400, on: true, keep: "full", preset: true},
	// Reveal and Sell (slice V6)
	"revealing":   {n: 140, capacity: 400, on: true, keep: "ok", act: "ok", actPreset: true},
	"revealpress": {n: 140, capacity: 400, on: true, keep: "ok", act: "ok"},
	"revealnone":  {n: 140, capacity: 400, on: true, keep: "ok", act: "none"},
	"revealcain":  {n: 140, capacity: 400, on: true, keep: "ok", act: "cain"},
	"revealdead":  {n: 140, capacity: 400, on: true, keep: "ok", act: "dead"},
	// the wallet states (V7b): enough Ether to buy, short, one step owned (the refund fits), one step owned and the vault too full for the lower tier, every step owned
	"eth": {n: 140, on: true, eth: true, balance: 35}, "eth-short": {n: 140, on: true, eth: true, balance: 5},
	"eth-owned": {n: 140, on: true, eth: true, balance: 15, owned: 1}, "eth-nofit": {n: 450, on: true, eth: true, balance: 15, owned: 1},
	"eth-max": {n: 140, on: true, eth: true, balance: 12, owned: 6},
}

// vaultEthers holds the wallet of each wallet state, keyed like vaultAPIs; they answer /api/ether and the two POST routes.
var vaultEthers = map[string]*watch.EtherAPI{}

// vaultName makes the i-th drop: a base and a quality that cycle, with a few fixed long and awkward names among them.
func vaultItem(i int, now time.Time) vault.Item {
	bases := []struct {
		code, name string
		chain      []string
	}{
		{"cap", "Cap", []string{"helm", "armo"}}, {"wsp", "War Sword", []string{"swor", "weap"}}, {"lrg", "Large Shield", []string{"shld", "armo"}},
		{"lgl", "Leather Gloves", []string{"glov", "armo"}}, {"rin", "Ring", []string{"ring", "misc"}}, {"amu", "Amulet", []string{"amul", "misc"}},
		{"bow", "Short Bow", []string{"bow", "weap"}}, {"hax", "Hand Axe", []string{"axe", "weap"}}, {"vbt", "Heavy Boots", []string{"boot", "armo"}}, {"jew", "Jewel", []string{"jewl", "misc"}},
	}
	b := bases[i%len(bases)]
	q := []int{2, 4, 4, 6, 3, 1, 4, 6, 7, 5, 2, 4}[i%12]
	it := vault.Item{Hero: "Sorc", Raw: fmt.Sprintf("%08x%08x", i+1, uint32((i+1)*2654435761)), Code: b.code, Name: b.name, Quality: q, Ilvl: 3 + (i*7)%80, Identified: i%6 != 1, Chain: b.chain,
		Sockets: []int{0, 0, 0, 1, 2, 0, 3, 0}[i%8], Eth: i%11 == 3,
		Found: vault.Found{Time: now.Add(-time.Duration(i*17+5) * time.Minute), Area: []string{"Blood Moor", "Cold Plains", "Burial Grounds", "Den of Evil", "Black Marsh"}[i%5]}}
	switch {
	case q == 4 && it.Identified:
		it.Name = []string{"Fine ", "Jagged ", "Fortified ", "Rugged "}[i%4] + b.name + []string{" of the Fox", " of Strength", " of Life", ""}[i%4]
	case q == 6 && it.Identified:
		it.Name = []string{"Doom Grip", "Gloom Plate", "Storm Eye", "Rune Spiral"}[i%4] + " (" + b.name + ")"
	case q == 7:
		it.Name = []string{"Stormshield", "Harlequin Crest", "Razortail", "Skin of the Vipermagi"}[i%4]
	case q == 5:
		it.Name = []string{"Civerb's Ward", "Hsarus' Iron Fist", "Angelic Mantle", "Tancred's Skull"}[i%4]
	}
	switch i {
	case 2:
		it.Name, it.Quality, it.Identified = "Stormshield, a very long unique shield name for the long names row", 7, true
	case 5:
		it.Code, it.Name, it.Quality, it.Chain, it.Sockets, it.Eth = "r05", "Thul Rune", 2, []string{"rune", "misc"}, 0, false
	case 6:
		it.Code, it.Name, it.Quality, it.Chain, it.Sockets, it.Eth = "r16", "Mal Rune", 2, []string{"rune", "misc"}, 0, false
	}
	if q == 4 && !it.Identified {
		it.Name = b.name // an unidentified magic item shows its base only
	}
	it.Value = q*100 + it.Ilvl
	return it
}

func newVaultFor(mode string, dir string) *watch.VaultAPI {
	set, ok := vaultModes[mode]
	if !ok {
		set = vaultModes["pop"]
	}
	tiers, err := vault.DefaultTiers()
	if err != nil {
		log.Fatalf("uishots: vault tiers: %v", err)
	}
	api := watch.NewVaultAPI()
	desk := watch.VaultDesk{On: func() bool { return set.on }, Tiers: tiers}
	if set.none {
		desk.Store = func() *vault.Store { return nil }
		api.Bind(desk)
		return api
	}
	capacity := set.capacity
	if set.shrinkTo > 0 {
		capacity = 1000
	}
	var wal *ether.Wallet
	var s *vault.Store
	if set.eth {
		var err error
		if wal, err = ether.Open(dir+"-ether", ether.Options{Tiers: tiers, Content: func() int { return s.Status().Count }}); err != nil {
			log.Fatalf("uishots: ether %s: %v", mode, err)
		}
		earned := set.balance
		for i := 0; i < set.owned; i++ {
			earned += tiers.Steps[i].Price
		}
		for i := 0; earned > 0; i++ { // the hook pays at most 100 per quest
			n := min(earned, 100)
			if _, err := wal.Earn("quest", fmt.Sprintf("q%d", i), "normal", n); err != nil {
				log.Fatalf("uishots: ether %s earn: %v", mode, err)
			}
			earned -= n
		}
		for i := 0; i < set.owned; i++ {
			if _, err := wal.Upgrade(""); err != nil {
				log.Fatalf("uishots: ether %s upgrade: %v", mode, err)
			}
		}
		capacity = wal.Capacity(0)
		desk.Owned = func() []string { o, _ := wal.OwnedSteps(); return o }
		vaultEthers[mode] = &watch.EtherAPI{W: wal}
	}
	now := time.Now()
	s, err = vault.OpenHero(dir, "Sorc", vault.Options{Capacity: func() int { return capacity }, NoSync: true})
	if err != nil {
		log.Fatalf("uishots: vault %s: %v", mode, err)
	}
	var ids []string
	for i := 0; i < set.n; i++ {
		r, err := s.Add(vaultItem(i, now))
		if err != nil {
			log.Fatalf("uishots: vault %s item %d: %v", mode, i, err)
		}
		ids = append(ids, r.ID)
	}
	for _, i := range []int{0, 3, 8, 12, 20} { // five pinned
		if i < len(ids) {
			s.Pin(ids[i])
		}
	}
	if set.shrinkTo > 0 {
		capacity = set.shrinkTo
	}
	if set.keep != "" {
		k := &fixtureKeeper{st: s, mode: set.keep}
		if set.preset {
			k.preset(ids)
		}
		desk.Keep = k
	}
	if set.act != "" {
		a := &fixtureActor{st: s, mode: set.act}
		if set.actPreset {
			a.preset(ids)
		}
		desk.Act = a
	}
	desk.Store = func() *vault.Store { return s }
	api.Bind(desk)
	return api
}

// vaultFixture serves the Vault page's routes for the state the page asks for.
func vaultFixture() http.Handler {
	vaultOnce.Do(func() {
		dir, err := os.MkdirTemp("", "uishots-vault-")
		if err != nil {
			log.Fatalf("uishots: vault dir: %v", err)
		}
		for mode := range vaultModes {
			vaultAPIs[mode] = newVaultFor(mode, dir+"/"+mode)
		}
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode := "pop"
		if ref, err := url.Parse(r.Referer()); err == nil && ref.Query().Get("vault") != "" {
			mode = ref.Query().Get("vault")
		}
		if v := r.URL.Query().Get("vault"); v != "" {
			mode = v
		}
		if mode == "error" && strings.HasPrefix(r.URL.Path, "/api/vault") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"code":"internal","error":"the vault could not be read"}`))
			return
		}
		api := vaultAPIs[mode]
		if api == nil {
			api = vaultAPIs["pop"]
		}
		api.Handler().ServeHTTP(w, r)
	})
}

func mountVault(m *http.ServeMux) {
	h := vaultFixture()
	for _, p := range watch.VaultRoutes {
		m.Handle(p, h)
	}
	m.Handle(watch.EtherRoute, etherFixture(func(a *watch.EtherAPI) http.Handler { return a.Get() }))
	m.Handle(watch.UpgradeRoute, etherFixture(func(a *watch.EtherAPI) http.Handler { return a.Upgrade(watch.AnswerOptions{}) }))
	m.Handle(watch.DowngradeRoute, etherFixture(func(a *watch.EtherAPI) http.Handler { return a.Downgrade(watch.AnswerOptions{}) }))
	// the answer routes too (same origin here), so a press shows the real reply and the page reads the state back
	for _, p := range watch.VaultPostRoutes {
		route := p
		m.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			mode := "pop"
			if ref, err := url.Parse(r.Referer()); err == nil && ref.Query().Get("vault") != "" {
				mode = ref.Query().Get("vault")
			}
			api := vaultAPIs[mode]
			if api == nil {
				http.Error(w, "no such fixture", http.StatusNotFound)
				return
			}
			api.PostHandler(watch.AnswerOptions{}, route).ServeHTTP(w, r)
		})
	}
}

// etherFixture routes the wallet calls of the page to the wallet of its state (the Referer's vault=); a state without a wallet answers
// the way a bot with loot.vault 0 does (enabled false, and 409 off on a press).
func etherFixture(h func(*watch.EtherAPI) http.Handler) http.Handler {
	vaultFixture() // builds the states
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode := "pop"
		if ref, err := url.Parse(r.Referer()); err == nil && ref.Query().Get("vault") != "" {
			mode = ref.Query().Get("vault")
		}
		a := vaultEthers[mode]
		if a == nil {
			a = &watch.EtherAPI{}
		}
		h(a).ServeHTTP(w, r)
	})
}

// vaultDefect is, per name, the one defect put back on the Vault page. The fit pages vaultdefect-NAME must FAIL with the verdict
// mustFail names (fit.go); a pass there means the check is blind to it. The style ones live in defectCSS (fixtures_phone.go
// serves them), the script one here.
func init() {
	defectCSS["vault-clip"] = `.vt-count { overflow: hidden !important; white-space: nowrap !important; max-inline-size: 2ch !important; }`
	defectCSS["vault-wide"] = `.vt-grid { min-inline-size: 140vw !important; }`
	defectCSS["vault-under"] = `.vt-main { position: fixed !important; inset: 0 !important; }`
	defectCSS["vault-small"] = `.vt-input { min-block-size: 28px !important; block-size: 28px !important; }`
	defectCSS["vault-covered"] = `.vt-head { position: fixed !important; inset: 0 0 auto 0 !important; block-size: 300px !important; z-index: 99; background: var(--d-surface-3) !important; }`
	defectCSS["vault-sheet"] = `.vt-sheet-card { max-block-size: 120px !important; overflow: hidden !important; }`
	vaultJS["vault-noname"] = `new MutationObserver(function () { document.querySelectorAll('.vt [aria-label]').forEach(function (e) { e.removeAttribute('aria-label'); }); }).observe(document, { subtree: true, childList: true, attributes: true });`
	vaultJS["vault-enabled"] = `new MutationObserver(function () { document.querySelectorAll('[data-key="upgrade"]').forEach(function (e) { e.removeAttribute('aria-disabled'); }); }).observe(document, { subtree: true, childList: true, attributes: true });`
	vaultJS["vault-ethenabled"] = `new MutationObserver(function () { document.querySelectorAll('[data-key="upgrade"]').forEach(function (e) { if (e.getAttribute('aria-disabled') !== 'false') e.setAttribute('aria-disabled', 'false'); }); }).observe(document, { subtree: true, childList: true, attributes: true });`
}

var vaultJS = map[string]string{}

// fixtureKeeper is the Keep queue of the fixture: it queues through the real store like internal/vaultkeep does; nothing creates
// items here (the executor is exercised in internal/vaultkeep with the fake game).
type fixtureKeeper struct {
	st   *vault.Store
	mode string
}

// preset puts a row in every Keep state: two queued, one creating, one failed with a real reason and two kept.
func (k *fixtureKeeper) preset(ids []string) {
	for _, i := range []int{1, 2, 4, 6} {
		_ = k.st.KeepTo(ids[i], "")
	}
	_ = k.st.BeginKeep(ids[4])
	if k.mode == "full" {
		_ = k.st.RetryKeep(ids[1], "keep: the stash is full: free space or upgrade")
	}
	_ = k.st.FailKeep(ids[6], "keep: the game does not accept this item's data")
	for _, i := range []int{9, 10} {
		_ = k.st.KeepTo(ids[i], "")
		_ = k.st.BeginKeep(ids[i])
		_ = k.st.CompleteKeepAt(ids[i], "stash 2,3")
	}
}

func (k *fixtureKeeper) Gate(context.Context) (bool, string, string) {
	if k.mode == "dead" {
		return false, "dead", "The hero is dead; Keep works again once it is alive and in town."
	}
	return true, "", ""
}

func (k *fixtureKeeper) Enqueue(_ context.Context, ids []string, target string, _ bool) []watch.VaultKeepResult {
	var out []watch.VaultKeepResult
	for _, id := range ids {
		rec, ok := k.st.Get(id)
		switch {
		case !ok:
			out = append(out, watch.VaultKeepResult{ID: id, Code: "missing", Why: "no such item"})
		case rec.State == vault.KeepQ || rec.State == vault.Kept:
			out = append(out, watch.VaultKeepResult{ID: id, OK: true, Already: true})
		default:
			if err := k.st.KeepTo(id, target); err != nil {
				out = append(out, watch.VaultKeepResult{ID: id, Code: "state", Why: err.Error()})
			} else {
				out = append(out, watch.VaultKeepResult{ID: id, OK: true})
			}
		}
	}
	return out
}

func (k *fixtureKeeper) Progress() watch.VaultKeepView {
	q := k.st.KeepQueued()
	v := watch.VaultKeepView{Enabled: k.mode != "dead", Full: k.mode == "full"}
	for _, r := range q {
		if r.InFlight {
			v.Creating++
		} else {
			v.Queued++
		}
	}
	switch n := v.Queued + v.Creating; {
	case n == 0:
	case v.Full:
		v.Line = fmt.Sprintf("Keeping %d items: the stash is full, free space or upgrade", n)
	default:
		v.Line = fmt.Sprintf("Keeping %d items: arrives at your next town visit", n)
	}
	return v
}

// fixtureActor is Reveal and Sell of the fixture: it queues through the real store like internal/vaultkeep does; nothing
// identifies or sells here (the executors are exercised in internal/vaultkeep with the fake game).
type fixtureActor struct {
	st     *vault.Store
	mode   string
	prompt []string
}

// preset puts a row in every state: two reveals queued, one failed with a real reason, two sales queued, one sold, one refused.
func (a *fixtureActor) preset(ids []string) {
	for _, i := range []int{1, 7} {
		_ = a.st.RevealWant(ids[i])
	}
	_ = a.st.RevealWant(ids[13])
	_ = a.st.RevealFail(ids[13], "reveal: no scroll; buy on request")
	for _, i := range []int{2, 4} {
		_ = a.st.MarkSell(ids[i])
	}
	_ = a.st.MarkSell(ids[5])
	_ = a.st.Sold(ids[5], 120)
	_ = a.st.MarkSell(ids[6])
	_ = a.st.FailSell(ids[6], "sell: no vendor of this act would buy it")
	a.prompt = []string{ids[9], ids[10]}
}

func (a *fixtureActor) charges() int {
	switch a.mode {
	case "none", "dead":
		return 0
	case "cain":
		return 0
	}
	return 3
}

func (a *fixtureActor) GateReveal(context.Context) (bool, string, string) {
	if a.mode == "dead" {
		return false, "dead", "The hero is dead; Reveal works again once it is alive and in town."
	}
	return true, "", ""
}

func (a *fixtureActor) GateSell(ctx context.Context) (bool, string, string) { return a.GateReveal(ctx) }

func (a *fixtureActor) EnqueueReveal(_ context.Context, ids []string, _ bool) []watch.VaultKeepResult {
	var out []watch.VaultKeepResult
	for _, id := range ids {
		rec, ok := a.st.Get(id)
		switch {
		case !ok:
			out = append(out, watch.VaultKeepResult{ID: id, Code: "missing", Why: "no such item"})
		case rec.RevQ:
			out = append(out, watch.VaultKeepResult{ID: id, OK: true, Already: true})
		case rec.Identified:
			out = append(out, watch.VaultKeepResult{ID: id, Code: "state", Why: "it is identified already"})
		case a.mode != "cain" && a.charges() == 0:
			out = append(out, watch.VaultKeepResult{ID: id, Code: "noscroll", Why: "no scroll; buy on request"})
		default:
			if err := a.st.RevealWant(id); err != nil {
				out = append(out, watch.VaultKeepResult{ID: id, Code: "state", Why: err.Error()})
			} else {
				out = append(out, watch.VaultKeepResult{ID: id, OK: true})
			}
		}
	}
	return out
}

func (a *fixtureActor) EnqueueSell(_ context.Context, ids []string, _ bool) []watch.VaultKeepResult {
	var out []watch.VaultKeepResult
	for _, id := range ids {
		switch err := a.st.MarkSell(id); {
		case err == nil:
			out = append(out, watch.VaultKeepResult{ID: id, OK: true})
		case err == vault.ErrPinned:
			out = append(out, watch.VaultKeepResult{ID: id, Code: "pinned", Why: err.Error()})
		default:
			out = append(out, watch.VaultKeepResult{ID: id, Code: "state", Why: err.Error()})
		}
	}
	return out
}

func (a *fixtureActor) Cancel(ids []string) []watch.VaultKeepResult {
	var out []watch.VaultKeepResult
	for _, id := range ids {
		err := a.st.CancelSell(id)
		if err != nil {
			err = a.st.RevealCancel(id)
		}
		out = append(out, watch.VaultKeepResult{ID: id, OK: err == nil, Code: "state", Why: "nothing is queued for it"})
	}
	return out
}

func (a *fixtureActor) DismissOffer() { a.prompt = nil }

func (a *fixtureActor) Act(ctx context.Context) watch.VaultActView {
	v := watch.VaultActView{
		Reveal: watch.VaultRevealView{Charges: a.charges(), Free: a.mode == "cain", PerItem: 1},
		Sell:   watch.VaultSellView{},
	}
	v.Reveal.Enabled, v.Reveal.Code, v.Reveal.Why = a.GateReveal(ctx)
	if v.Reveal.Enabled {
		v.Reveal.Code, v.Reveal.Why = "", ""
		if a.mode == "none" {
			v.Reveal.Enabled, v.Reveal.Code, v.Reveal.Why = false, "noscroll", "no scroll; buy on request"
		}
	}
	v.Sell.Enabled, v.Sell.Code, v.Sell.Why = a.GateSell(ctx)
	for _, r := range a.st.RevealQueued() {
		if r.Stage != "" {
			v.Reveal.Working++
		} else {
			v.Reveal.Queued++
		}
	}
	v.Sell.Queued = len(a.st.SellQueued())
	if n := v.Reveal.Queued; n > 0 {
		v.Reveal.Line = fmt.Sprintf("Revealing %d items: at your next town visit", n)
	}
	if n := v.Sell.Queued; n > 0 {
		v.Sell.Line = fmt.Sprintf("Selling %d items: at your next town visit", n)
	}
	if len(a.prompt) > 0 {
		gold := 0
		var names []string
		for _, id := range a.prompt {
			if r, ok := a.st.Get(id); ok {
				gold, names = gold+r.Value, append(names, r.Name)
			}
		}
		v.Prompt = &watch.VaultPrompt{Short: 140, IDs: a.prompt, Gold: gold, Names: names}
	}
	return v
}
