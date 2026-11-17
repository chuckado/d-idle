package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/vault"
)

// testVault opens a vault in a temp dir with room for cap records. Item i is deterministic: the code cycles over four bases
// (one a rune), quality over 1..8, ilvl over 1..89, sockets over 0..3, every 7th one ethereal, every 5th unidentified.
func testVault(t testing.TB, n, capacity int) *vault.Store {
	t.Helper()
	s, err := vault.OpenHero(t.TempDir(), "Sorc", vault.Options{Capacity: func() int { return capacity }, NoSync: true,
		Now: func() time.Time { return time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC) }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	codes := []struct {
		code, name string
		chain      []string
	}{
		{"cap", "Cap", []string{"helm", "armo"}}, {"wsp", "War Sword", []string{"swor", "weap"}}, {"r05", "Thul Rune", []string{"rune", "misc"}}, {"rin", "Ring", []string{"ring", "misc"}},
	}
	for i := 0; i < n; i++ {
		c := codes[i%4]
		it := vault.Item{Hero: "Sorc", Raw: fmt.Sprintf("%08x%08x", i, i*7919), Code: c.code, Name: fmt.Sprintf("%s %d", c.name, i), Quality: 1 + i%8, Ilvl: 1 + i%89,
			Identified: i%5 != 0, Chain: c.chain, Sockets: i % 4, Eth: i%7 == 0,
			Found: vault.Found{Time: time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Second), Area: "Burial Grounds"}}
		it.Value = it.Quality*100 + it.Ilvl
		if _, err := s.Add(it); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}
	return s
}

func vaultAPI(s *vault.Store, on bool) *VaultAPI {
	tiers, _ := vault.DefaultTiers()
	a := NewVaultAPI()
	a.Bind(VaultDesk{On: func() bool { return on }, Store: func() *vault.Store { return s }, Tiers: tiers})
	return a
}

func getVault(t testing.TB, a *VaultAPI, path string) (int, VaultView) {
	t.Helper()
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	var v VaultView
	if w.Code == 200 {
		if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
			t.Fatalf("%s: %v\n%s", path, err, w.Body)
		}
	}
	return w.Code, v
}

func TestVaultListPagingAndFilters(t *testing.T) {
	s := testVault(t, 10_000, 12_000)
	a := vaultAPI(s, true)
	code, v := getVault(t, a, "/api/vault?facets=1")
	if code != 200 || v.Schema != "watch.vault/1" || v.Total != 10_000 || len(v.Items) != vaultPageLimit || !v.Exists || v.Hero != "Sorc" {
		t.Fatalf("first page: %d total %d items %d exists %v hero %q", code, v.Total, len(v.Items), v.Exists, v.Hero)
	}
	if v.Items[0].Name != "Ring 9999" { // newest first
		t.Errorf("newest first: %q", v.Items[0].Name)
	}
	if len(v.Kinds) < 3 || len(v.Types) == 0 {
		t.Errorf("facets: %d kinds %d types", len(v.Kinds), len(v.Types))
	}
	if v.Capacity.Count != 10_000 || v.Capacity.Capacity != 12_000 || v.Capacity.Level != "warn" || v.Capacity.Percent != 83 {
		t.Errorf("capacity %+v", v.Capacity)
	}
	// a page past the first is the next rows, not the same ones
	_, p2 := getVault(t, a, "/api/vault?offset=60&limit=60")
	if p2.Items[0].ID == v.Items[0].ID || p2.Items[0].Name != "Ring 9939" {
		t.Errorf("second page starts at %q", p2.Items[0].Name)
	}
	// every filter against a count worked out by hand from the generator above
	count := func(pred func(i int) bool) int {
		n := 0
		for i := 0; i < 10_000; i++ {
			if pred(i) {
				n++
			}
		}
		return n
	}
	for _, c := range []struct {
		name, q string
		want    int
	}{
		{"quality unique", "quality=unique", count(func(i int) bool { return 1+i%8 == 7 })},
		{"quality words and ids", "quality=rare,4", count(func(i int) bool { q := 1 + i%8; return q == 6 || q == 4 })},
		{"kind unique", "kind=unique", count(func(i int) bool { return 1+i%8 == 7 })},
		{"kind rune", "kind=rune", count(func(i int) bool { return i%4 == 2 })},
		{"kind set and quality chip share nothing", "kind=set&quality=rare", 0},
		{"ilvl range", "ilvlMin=20&ilvlMax=30", count(func(i int) bool { l := 1 + i%89; return l >= 20 && l <= 30 })},
		{"sockets", "sockets=3", count(func(i int) bool { return i%4 == 3 })},
		{"no sockets", "sockets=0", count(func(i int) bool { return i%4 == 0 })},
		{"ethereal", "eth=1", count(func(i int) bool { return i%7 == 0 })},
		{"unidentified", "identified=0", count(func(i int) bool { return i%5 == 0 })},
		{"type from the base chain", "type=swor", count(func(i int) bool { return i%4 == 1 })},
		{"name search", "q=war+sword+99", count(func(i int) bool { return i%4 == 1 && strings.HasPrefix(fmt.Sprint(i), "99") })},
		{"pinned", "pinned=1", 0},
		{"combined", "kind=unique&ilvlMin=50&eth=1&identified=1", count(func(i int) bool { return 1+i%8 == 7 && 1+i%89 >= 50 && i%7 == 0 && i%5 != 0 })},
	} {
		code, v := getVault(t, a, "/api/vault?limit=1&"+c.q)
		if code != 200 || v.Total != c.want {
			t.Errorf("%s (%s): %d total %d, want %d", c.name, c.q, code, v.Total, c.want)
		}
	}
	// sorts
	for sort, ok := range map[string]func(a, b VaultItem) bool{
		"quality": func(a, b VaultItem) bool { return a.Q >= b.Q },
		"ilvl":    func(a, b VaultItem) bool { return a.Ilvl >= b.Ilvl },
		"name":    func(a, b VaultItem) bool { return strings.ToLower(a.Name) <= strings.ToLower(b.Name) },
		"value":   func(a, b VaultItem) bool { return a.Value >= b.Value },
	} {
		_, v := getVault(t, a, "/api/vault?limit=200&sort="+sort)
		for i := 1; i < len(v.Items); i++ {
			if !ok(v.Items[i-1], v.Items[i]) {
				t.Fatalf("sort %s: %q before %q", sort, v.Items[i-1].Name, v.Items[i].Name)
			}
		}
	}
	_, old := getVault(t, a, "/api/vault?sort=old&limit=1")
	if old.Items[0].Name != "Cap 0" {
		t.Errorf("oldest first: %q", old.Items[0].Name)
	}
}

func TestVaultQueryRefusals(t *testing.T) {
	a := vaultAPI(testVault(t, 20, 100), true)
	for _, q := range []string{"quality=shiny", "kind=nope", "sort=random", "limit=9999", "offset=-1", "eth=maybe", "state=lost", "sockets=9", "q=" + strings.Repeat("x", 65)} {
		if code, _ := getVault(t, a, "/api/vault?"+q); code != 400 {
			t.Errorf("%s: %d, want 400", q, code)
		}
	}
}

// Paging at 10k must stay far under the page's poll: a filtered, sorted page in well under 50 ms on a small box.
func TestVaultPagingPerformance(t *testing.T) {
	s := testVault(t, 10_000, 12_000)
	a := vaultAPI(s, true)
	for _, q := range []string{"", "?sort=value", "?sort=name&offset=5000", "?kind=unique&sort=ilvl", "?q=ring+9&sort=quality", "?type=swor&ilvlMin=10&facets=1"} {
		start := time.Now()
		for i := 0; i < 10; i++ {
			getVault(t, a, "/api/vault"+q)
		}
		per := time.Since(start) / 10
		t.Logf("GET /api/vault%-40s %v", q, per)
		if per > 150*time.Millisecond { // a loose ceiling for a shared box; the measured numbers are in docs/vault.md
			t.Errorf("%s took %v per request", q, per)
		}
	}
}

func TestVaultEmptyOffAndNoVault(t *testing.T) {
	// no desk at all: the page opens and the API says off
	a := NewVaultAPI()
	if code, _ := getVault(t, a, "/api/vault"); code != 503 {
		t.Errorf("no desk: %d", code)
	}
	// a desk with no vault yet: an empty list, the base capacity, the placeholder upgrade
	tiers, _ := vault.DefaultTiers()
	a.Bind(VaultDesk{Store: func() *vault.Store { return nil }, Tiers: tiers})
	code, v := getVault(t, a, "/api/vault")
	if code != 200 || v.Exists || len(v.Items) != 0 || v.Capacity.Capacity != tiers.Base || v.Mode != "on" {
		t.Errorf("no vault: %d %+v", code, v)
	}
	if v.Upgrade.Available || v.Upgrade.Next == nil || v.Upgrade.Next.ID != "v1" || v.Upgrade.Currency != "Ether" {
		t.Errorf("upgrade placeholder: %+v", v.Upgrade)
	}
	// loot.vault at 0 with items on disk: readable, mode off
	s := testVault(t, 30, 100)
	code, v = getVault(t, vaultAPI(s, false), "/api/vault")
	if code != 200 || v.Mode != "off" || v.Total != 30 {
		t.Errorf("off: %d mode %q total %d", code, v.Mode, v.Total)
	}
	// the item route on an unknown id
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault/item/v-nope", nil))
	if w.Code != 404 {
		t.Errorf("unknown item: %d", w.Code)
	}
}

func TestVaultCapacityLevels(t *testing.T) {
	for _, c := range []struct {
		n, cap int
		level  string
		over   int
	}{{10, 100, "ok", 0}, {80, 100, "warn", 0}, {100, 100, "full", 0}, {120, 100, "over", 20}} {
		st := vault.Status{Count: c.n, Capacity: c.cap, Over: c.n > c.cap, Full: c.n == c.cap}
		if got := capacityOf(st); got.Level != c.level || got.Over != c.over {
			t.Errorf("%d/%d: %+v", c.n, c.cap, got)
		}
	}
	// a shrink below the count through the live path: over capacity, nothing deleted, reads fine
	n := 50
	capacity := 100
	s, err := vault.OpenHero(t.TempDir(), "Sorc", vault.Options{Capacity: func() int { return capacity }, NoSync: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for i := 0; i < n; i++ {
		s.Add(vault.Item{Hero: "Sorc", Raw: fmt.Sprintf("%04x", i), Code: "cap", Name: "Cap", Quality: 4, Ilvl: 5, Identified: true})
	}
	capacity = 30
	_, v := getVault(t, vaultAPI(s, true), "/api/vault")
	if v.Capacity.Level != "over" || v.Capacity.Over != 20 || v.Total != 50 {
		t.Errorf("shrunk: %+v total %d", v.Capacity, v.Total)
	}
}

func TestVaultItemDetailAndUnidentified(t *testing.T) {
	s := testVault(t, 10, 100)
	a := vaultAPI(s, true)
	_, v := getVault(t, a, "/api/vault?identified=0")
	if len(v.Items) == 0 {
		t.Fatal("no unidentified item in the fixture")
	}
	id := v.Items[0].ID
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault/item/"+url.PathEscape(id), nil))
	var d VaultDetail
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &d) != nil {
		t.Fatalf("detail %d %s", w.Code, w.Body)
	}
	if d.Item.Identified || !strings.Contains(strings.Join(d.Lines, "|"), "Unidentified") || d.Where["area"] != "Burial Grounds" {
		t.Errorf("detail %+v", d)
	}
	if !d.Actions["pin"].OK || !d.Actions["discard"].OK || d.Actions["keep"].OK || !strings.Contains(d.Actions["keep"].Reason, "no Keep queue") || d.Actions["reveal"].OK {
		t.Errorf("actions %+v", d.Actions)
	}
	// the row is what the tooltip script reads: quality as a word, level, no property text for an unidentified record
	if d.Item.Quality == "" || d.Item.Level != d.Item.Ilvl {
		t.Errorf("tooltip fields %+v", d.Item)
	}
	if strings.Contains(w.Body.String(), `"raw"`) || strings.Contains(w.Body.String(), `"props"`) {
		t.Error("the detail leaks the raw item bytes or property text")
	}
	// read only: the actions say why
	w = httptest.NewRecorder()
	vaultAPI(s, false).Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault/item/"+url.PathEscape(id), nil))
	json.Unmarshal(w.Body.Bytes(), &d)
	if d.Actions["pin"].OK || !strings.Contains(d.Actions["pin"].Reason, "read only") {
		t.Errorf("read only actions %+v", d.Actions)
	}
}

func postVault(t testing.TB, a *VaultAPI, route string, body any) (int, vaultReply, policyReply) {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", route, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.PostHandler(AnswerOptions{}, route).ServeHTTP(w, r)
	var rep vaultReply
	var pr policyReply
	json.Unmarshal(w.Body.Bytes(), &rep)
	json.Unmarshal(w.Body.Bytes(), &pr)
	return w.Code, rep, pr
}

func TestVaultPinUnpinDiscard(t *testing.T) {
	s := testVault(t, 40, 100)
	a := vaultAPI(s, true)
	_, v := getVault(t, a, "/api/vault?limit=5")
	ids := []string{v.Items[0].ID, v.Items[1].ID}
	code, rep, _ := postVault(t, a, VaultPinRoute, vaultBody{IDs: ids})
	if code != 200 || !rep.OK || len(rep.Done) != 2 {
		t.Fatalf("pin: %d %+v", code, rep)
	}
	_, p := getVault(t, a, "/api/vault?pinned=1")
	if p.Total != 2 || p.Capacity.Pins != 2 || !p.Items[0].Pin || p.Items[0].Status != "pinned" {
		t.Errorf("pinned list %+v", p)
	}
	if code, rep, _ = postVault(t, a, VaultUnpinRoute, vaultBody{IDs: ids[:1]}); code != 200 || !rep.OK {
		t.Errorf("unpin %d %+v", code, rep)
	}
	// discard needs the confirm flag, then archives first and leaves the live list
	if code, _, pr := postVault(t, a, VaultDiscardRoute, vaultBody{IDs: ids}); code != 400 || pr.Code != "confirm" {
		t.Errorf("discard without confirm: %d %+v", code, pr)
	}
	if code, rep, _ = postVault(t, a, VaultDiscardRoute, vaultBody{IDs: ids, Confirm: true}); code != 200 || len(rep.Done) != 2 {
		t.Fatalf("discard: %d %+v", code, rep)
	}
	arch, err := s.Archived()
	if err != nil || len(arch) != 2 {
		t.Errorf("archive holds %d (%v), want 2", len(arch), err)
	}
	if _, v := getVault(t, a, "/api/vault"); v.Total != 38 {
		t.Errorf("live total %d, want 38", v.Total)
	}
	if _, v := getVault(t, a, "/api/vault?state=discarded"); v.Total != 2 {
		t.Errorf("discarded list %d", v.Total)
	}
	// a second discard of the same id is refused with a code, nothing else breaks
	code, rep, _ = postVault(t, a, VaultDiscardRoute, vaultBody{IDs: ids[:1], Confirm: true})
	if code != 409 || len(rep.Failed) != 1 || rep.Failed[0].Code != "state" {
		t.Errorf("repeat discard: %d %+v", code, rep)
	}
	// a partly refused batch is a 200 with the failures listed
	_, v = getVault(t, a, "/api/vault?limit=1")
	code, rep, _ = postVault(t, a, VaultPinRoute, vaultBody{IDs: []string{v.Items[0].ID, "v-missing"}})
	if code != 200 || rep.OK || len(rep.Done) != 1 || rep.Failed[0].Code != "missing" {
		t.Errorf("partial: %d %+v", code, rep)
	}
	// the id limit and the body gate
	big := make([]string, vaultMaxIDs+1)
	for i := range big {
		big[i] = fmt.Sprint("v-", i)
	}
	if code, _, pr := postVault(t, a, VaultPinRoute, vaultBody{IDs: big}); code != 400 || pr.Code != "ids" {
		t.Errorf("too many ids: %d %+v", code, pr)
	}
	r := httptest.NewRequest("POST", VaultPinRoute, strings.NewReader(`{"ids":["v-1"],"extra":1}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.PostHandler(AnswerOptions{}, VaultPinRoute).ServeHTTP(w, r)
	if w.Code != 400 {
		t.Errorf("unknown field: %d", w.Code)
	}
	// the biggest allowed batch fits the 4 KiB body cap
	b, _ := json.Marshal(vaultBody{IDs: func() []string {
		l := make([]string, vaultMaxIDs)
		for i := range l {
			l[i] = fmt.Sprintf("v-%016x~12", i)
		}
		return l
	}(), Confirm: true})
	if len(b) > maxBody {
		t.Errorf("a full batch is %d bytes, over the %d cap", len(b), maxBody)
	}
}

func TestVaultReadOnlyRefusesWrites(t *testing.T) {
	s := testVault(t, 5, 100)
	_, v := getVault(t, vaultAPI(s, true), "/api/vault?limit=1")
	for _, a := range []*VaultAPI{vaultAPI(s, false), vaultAPI(nil, true), NewVaultAPI()} {
		code, _, pr := postVault(t, a, VaultPinRoute, vaultBody{IDs: []string{v.Items[0].ID}})
		if code != 409 || pr.Code != "off" {
			t.Errorf("read only pin: %d %+v", code, pr)
		}
	}
	if _, p := getVault(t, vaultAPI(s, true), "/api/vault?pinned=1"); p.Total != 0 {
		t.Error("a refused pin changed the vault")
	}
}

func TestVaultViewerIsGetOnly(t *testing.T) {
	a := vaultAPI(testVault(t, 3, 10), true)
	for _, p := range []string{"/api/vault", "/api/vault/pin", "/watch/vault"} {
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, httptest.NewRequest("POST", p, nil))
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s on the viewer: %d", p, w.Code)
		}
	}
}

// stubKeeper queues through the store like internal/vaultkeep does, with a switchable gate.
type stubKeeper struct {
	st                *vault.Store
	gateCode, gateWhy string
	nowSeen           bool
}

func (k *stubKeeper) Gate(context.Context) (bool, string, string) {
	return k.gateCode == "", k.gateCode, k.gateWhy
}

func (k *stubKeeper) Enqueue(_ context.Context, ids []string, target string, now bool) []VaultKeepResult {
	k.nowSeen = k.nowSeen || now
	var out []VaultKeepResult
	for _, id := range ids {
		rec, ok := k.st.Get(id)
		switch {
		case !ok:
			out = append(out, VaultKeepResult{ID: id, Code: "missing", Why: "no such item"})
		case rec.State == vault.KeepQ:
			out = append(out, VaultKeepResult{ID: id, OK: true, Already: true})
		default:
			if err := k.st.KeepTo(id, target); err != nil {
				out = append(out, VaultKeepResult{ID: id, Code: "state", Why: err.Error()})
			} else {
				out = append(out, VaultKeepResult{ID: id, OK: true})
			}
		}
	}
	return out
}

func (k *stubKeeper) Progress() VaultKeepView {
	n := len(k.st.KeepQueued())
	v := VaultKeepView{Enabled: k.gateCode == "", Why: k.gateWhy, Queued: n}
	if n > 0 {
		v.Line = "Keeping " + strconv.Itoa(n) + " items: arrives at your next town visit"
	}
	return v
}

func keepAPI(s *vault.Store, k *stubKeeper) *VaultAPI {
	tiers, _ := vault.DefaultTiers()
	a := NewVaultAPI()
	a.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Tiers: tiers, Keep: k})
	return a
}

func TestVaultKeepRoute(t *testing.T) {
	s := testVault(t, 10, 100)
	k := &stubKeeper{st: s}
	a := keepAPI(s, k)
	_, v := getVault(t, a, "/api/vault?limit=5")
	ids := []string{v.Items[0].ID, v.Items[1].ID}

	code, rep, _ := postVault(t, a, VaultKeepRoute, vaultBody{IDs: ids, Now: true})
	if code != 200 || !rep.OK || len(rep.Done) != 2 || len(rep.Already) != 0 || !k.nowSeen {
		t.Fatalf("keep: %d %+v", code, rep)
	}
	// a double press changes nothing and says so
	if code, rep, _ = postVault(t, a, VaultKeepRoute, vaultBody{IDs: ids}); code != 200 || len(rep.Already) != 2 {
		t.Fatalf("double press: %d %+v", code, rep)
	}
	// the page reads the state back from GET
	_, q := getVault(t, a, "/api/vault?state=keepq")
	if q.Total != 2 || q.Items[0].Keep != "queued" || q.Keep == nil || q.Keep.Queued != 2 || !strings.Contains(q.Keep.Line, "Keeping 2 items") {
		t.Fatalf("state read back: %+v keep %+v", q.Items, q.Keep)
	}
	// partial: one id does not exist, the reply names it and the real reason
	_, rep, _ = postVault(t, a, VaultKeepRoute, vaultBody{IDs: []string{v.Items[2].ID, "v-nope"}})
	if rep.OK || len(rep.Done) != 1 || len(rep.Failed) != 1 || rep.Failed[0].Code != "missing" || rep.Failed[0].Error != "no such item" {
		t.Fatalf("partial: %+v", rep)
	}
	// a bad target is refused before anything is queued
	if code, _, pr := postVault(t, a, VaultKeepRoute, vaultBody{IDs: ids, Target: "moon"}); code != 400 || pr.Code != "target" {
		t.Fatalf("target: %d %+v", code, pr)
	}
	// the gate: dead hero, verbs missing and so on answer 409 with the code and the sentence
	k.gateCode, k.gateWhy = "dead", "the hero is dead"
	if code, _, pr := postVault(t, a, VaultKeepRoute, vaultBody{IDs: []string{v.Items[3].ID}}); code != 409 || pr.Code != "dead" || pr.Error != "the hero is dead" {
		t.Fatalf("gate: %d %+v", code, pr)
	}
	// the card's Keep action carries the same reason
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault/item/"+url.PathEscape(v.Items[3].ID), nil))
	var d VaultDetail
	json.Unmarshal(w.Body.Bytes(), &d)
	if d.Actions["keep"].OK || d.Actions["keep"].Reason != "the hero is dead" {
		t.Fatalf("detail keep %+v", d.Actions["keep"])
	}
	k.gateCode, k.gateWhy = "", ""
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault/item/"+url.PathEscape(v.Items[3].ID), nil))
	json.Unmarshal(w.Body.Bytes(), &d)
	if !d.Actions["keep"].OK {
		t.Fatalf("detail keep %+v", d.Actions["keep"])
	}
}

func TestVaultRowKeepStates(t *testing.T) {
	s := testVault(t, 4, 100)
	_, v := getVault(t, keepAPI(s, &stubKeeper{st: s}), "/api/vault?limit=4")
	id := v.Items[0].ID
	_ = s.KeepTo(id, "")
	_ = s.BeginKeep(id)
	if r := rowOf(t, s, id); r.Keep != "creating" {
		t.Fatalf("%+v", r)
	}
	_ = s.RetryKeep(id, "keep: the stash is full: free space or upgrade")
	if r := rowOf(t, s, id); r.Keep != "queued" || r.Reason != "the stash is full: free space or upgrade" {
		t.Fatalf("%+v", r)
	}
	_ = s.FailKeep(id, "keep: the game does not accept this item's data")
	if r := rowOf(t, s, id); r.Keep != "failed" || r.Reason != "the game does not accept this item's data" || r.State != "recorded" {
		t.Fatalf("%+v", r)
	}
	_ = s.KeepTo(id, "inv")
	_ = s.CompleteKeepAt(id, "backpack 2,1")
	if r := rowOf(t, s, id); r.Keep != "kept" || r.Reason != "backpack 2,1" || r.Where != "inv" {
		t.Fatalf("%+v", r)
	}
}

func rowOf(t *testing.T, s *vault.Store, id string) VaultItem {
	t.Helper()
	rec, ok := s.Get(id)
	if !ok {
		t.Fatal("missing")
	}
	return vaultRow(&rec)
}

// stubActor queues Reveal and Sell through the store like internal/vaultkeep does, with switchable gates.
type stubActor struct {
	st      *vault.Store
	revGate string // "" ok, else the refusal code (the reason is the code text)
	sellGat string
	charges int
	free    bool
	dismiss int
}

func (a *stubActor) GateReveal(context.Context) (bool, string, string) {
	if a.revGate != "" {
		return false, a.revGate, "reveal refused: " + a.revGate
	}
	return true, "", ""
}
func (a *stubActor) GateSell(context.Context) (bool, string, string) {
	if a.sellGat != "" {
		return false, a.sellGat, "sell refused: " + a.sellGat
	}
	return true, "", ""
}
func (a *stubActor) EnqueueReveal(_ context.Context, ids []string, _ bool) []VaultKeepResult {
	var out []VaultKeepResult
	for _, id := range ids {
		rec, ok := a.st.Get(id)
		switch {
		case !ok:
			out = append(out, VaultKeepResult{ID: id, Code: "missing", Why: "no such item"})
		case rec.RevQ:
			out = append(out, VaultKeepResult{ID: id, OK: true, Already: true})
		case rec.Identified:
			out = append(out, VaultKeepResult{ID: id, Code: "state", Why: "it is identified already"})
		case !a.free && a.charges <= 0:
			out = append(out, VaultKeepResult{ID: id, Code: "noscroll", Why: "no scroll; buy on request"})
		default:
			_ = a.st.RevealWant(id)
			out = append(out, VaultKeepResult{ID: id, OK: true})
		}
	}
	return out
}
func (a *stubActor) EnqueueSell(_ context.Context, ids []string, _ bool) []VaultKeepResult {
	var out []VaultKeepResult
	for _, id := range ids {
		if err := a.st.MarkSell(id); err != nil {
			code := "state"
			if errors.Is(err, vault.ErrPinned) {
				code = "pinned"
			}
			out = append(out, VaultKeepResult{ID: id, Code: code, Why: err.Error()})
			continue
		}
		out = append(out, VaultKeepResult{ID: id, OK: true})
	}
	return out
}
func (a *stubActor) Cancel(ids []string) []VaultKeepResult {
	var out []VaultKeepResult
	for _, id := range ids {
		err := a.st.CancelSell(id)
		out = append(out, VaultKeepResult{ID: id, OK: err == nil, Code: "state"})
	}
	return out
}
func (a *stubActor) DismissOffer() { a.dismiss++ }
func (a *stubActor) Act(context.Context) VaultActView {
	v := VaultActView{Reveal: VaultRevealView{Enabled: a.revGate == "", Charges: a.charges, Free: a.free, PerItem: 1}, Sell: VaultSellView{Enabled: a.sellGat == ""}}
	for _, r := range a.st.RevealQueued() {
		v.Reveal.Queued += map[bool]int{true: 1}[r.Stage == ""]
	}
	v.Sell.Queued = len(a.st.SellQueued())
	return v
}

func actAPI(s *vault.Store, a *stubActor) *VaultAPI {
	tiers, _ := vault.DefaultTiers()
	api := NewVaultAPI()
	api.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Tiers: tiers, Act: a})
	return api
}

func TestVaultRevealAndSellRoutes(t *testing.T) {
	s := testVault(t, 12, 100)
	a := &stubActor{st: s, charges: 2}
	api := actAPI(s, a)
	_, un := getVault(t, api, "/api/vault?identified=0&limit=10")
	_, idf := getVault(t, api, "/api/vault?identified=1&pinned=0&limit=10")
	if len(un.Items) < 2 || len(idf.Items) < 3 {
		t.Fatalf("fixture: %d unidentified, %d identified", len(un.Items), len(idf.Items))
	}
	// reveal: ok, idempotent, identified refused with the real reason, no scroll refused with the buy hint
	code, rep, _ := postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{un.Items[0].ID}, Now: true})
	if code != 200 || !rep.OK || len(rep.Done) != 1 {
		t.Fatalf("reveal %d %+v", code, rep)
	}
	if _, rep, _ = postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{un.Items[0].ID}}); len(rep.Already) != 1 {
		t.Fatalf("second press: %+v", rep)
	}
	_, rep, _ = postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{idf.Items[0].ID}})
	if rep.OK || len(rep.Failed) != 1 || rep.Failed[0].Code != "state" {
		t.Fatalf("identified: %+v", rep)
	}
	a.charges = 0
	_, rep, _ = postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{un.Items[1].ID}})
	if len(rep.Failed) != 1 || rep.Failed[0].Code != "noscroll" || !strings.Contains(rep.Failed[0].Error, "buy on request") {
		t.Fatalf("no scroll: %+v", rep)
	}
	a.revGate = "dead"
	if code, _, pr := postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{un.Items[1].ID}}); code != 409 || pr.Code != "dead" {
		t.Fatalf("gate: %d %+v", code, pr)
	}
	// sell: needs confirm; pinned refused by code
	if code, _, pr := postVault(t, api, VaultSellRoute, vaultBody{IDs: []string{idf.Items[0].ID}}); code != 400 || pr.Code != "confirm" {
		t.Fatalf("confirm: %d %+v", code, pr)
	}
	pin := idf.Items[1].ID
	if err := s.Pin(pin); err != nil {
		t.Fatal(err)
	}
	_, rep, _ = postVault(t, api, VaultSellRoute, vaultBody{IDs: []string{idf.Items[0].ID, pin}, Confirm: true})
	if len(rep.Done) != 1 || len(rep.Failed) != 1 || rep.Failed[0].Code != "pinned" {
		t.Fatalf("sell: %+v", rep)
	}
	// the page reads the state back from GET: row words, the price and the queue
	_, v := getVault(t, api, "/api/vault?limit=100&state=recorded,sellq")
	var sq, rq int
	for _, it := range v.Items {
		if it.Sell == "queued" {
			sq++
		}
		if it.Reveal == "queued" {
			rq++
		}
	}
	if sq != 1 || rq != 1 || v.Act == nil || v.Act.Sell.Queued != 1 || v.Act.Reveal.Queued != 1 || v.Act.Reveal.PerItem != 1 {
		t.Fatalf("read back: sell %d reveal %d act %+v", sq, rq, v.Act)
	}
	// dismiss
	if code, _, _ := postVault(t, api, VaultPromptRoute, vaultBody{Dismiss: true}); code != 200 || a.dismiss != 1 {
		t.Fatalf("dismiss %d %d", code, a.dismiss)
	}
	// detail actions carry the reasons
	w := httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault/item/"+url.PathEscape(pin), nil))
	var d VaultDetail
	json.Unmarshal(w.Body.Bytes(), &d)
	if d.Actions["sell"].OK || !strings.Contains(d.Actions["sell"].Reason, "never sold") {
		t.Fatalf("pinned card: %+v", d.Actions["sell"])
	}
}

func TestVaultActReadOnlyAndAbsent(t *testing.T) {
	s := testVault(t, 4, 100)
	ids := []string{"x"}
	// no actor: routes answer off
	if code, _, pr := postVault(t, vaultAPI(s, true), VaultRevealRoute, vaultBody{IDs: ids}); code != 409 || pr.Code != "off" {
		t.Fatalf("no actor: %d %+v", code, pr)
	}
	// knob off: read only
	tiers, _ := vault.DefaultTiers()
	a := NewVaultAPI()
	a.Bind(VaultDesk{On: func() bool { return false }, Store: func() *vault.Store { return s }, Tiers: tiers, Act: &stubActor{st: s, charges: 1}})
	if code, _, pr := postVault(t, a, VaultSellRoute, vaultBody{IDs: ids, Confirm: true}); code != 409 || pr.Code != "off" {
		t.Fatalf("read only: %d %+v", code, pr)
	}
	_, v := getVault(t, a, "/api/vault")
	if v.Act == nil || v.Act.Reveal.Enabled || v.Act.Sell.Enabled || !strings.Contains(v.Act.Sell.Why, "read only") {
		t.Fatalf("read only act view: %+v", v.Act)
	}
}
