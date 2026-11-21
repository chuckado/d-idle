package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/ether"
	"d2lod-bot/internal/scenario"
	"d2lod-bot/internal/vault"
)

// fakeGame is a hero as far as the seeding sees it: debug give adds an item, item vault removes it, item equip wears it.
type fakeGame struct {
	next    int
	items   []*fakeItem
	lines   []string
	dupes   bool // every give of a code rolls the same bytes
	noEquip bool // the equip verb answers ok and does nothing
	noGive  bool // the give verb answers ok and creates nothing
}

type fakeItem struct {
	id      int
	code    string
	quality int
	unid    bool
	loc     int // 0 pack, 1 worn
	roll    int
}

func (g *fakeGame) Do(_ context.Context, line string) (*d2dctl.Reply, error) {
	g.lines = append(g.lines, line)
	f := strings.Fields(line)
	switch {
	case len(f) >= 3 && f[0] == "debug" && f[1] == "give":
		if g.noGive {
			break
		}
		g.next++
		it := &fakeItem{id: 100 + g.next, code: f[2], quality: 2, roll: g.next}
		if g.dupes {
			it.roll = 0
		}
		for _, w := range f[3:] {
			switch {
			case w == "unid":
				it.unid = true
			case strings.HasPrefix(w, "q"):
				it.quality, _ = strconv.Atoi(w[1:])
				// a magic item of a later give rolls other affixes even when dupes is on
				if it.quality == 4 {
					it.roll = g.next
				}
			}
		}
		g.items = append(g.items, it)
	case len(f) == 3 && f[0] == "item" && f[1] == "vault":
		id, _ := strconv.Atoi(f[2])
		for i, it := range g.items {
			if it.id == id {
				g.items = append(g.items[:i], g.items[i+1:]...)
				break
			}
		}
	case len(f) == 4 && f[0] == "item" && f[1] == "equip":
		id, _ := strconv.Atoi(f[2])
		for _, it := range g.items {
			if it.id == id && !g.noEquip {
				it.loc = 1
			}
		}
	}
	return &d2dctl.Reply{Rest: "ok"}, nil
}

func (g *fakeGame) Step(context.Context, int) (d2dctl.ClockInfo, error) {
	return d2dctl.ClockInfo{}, nil
}

func (g *fakeGame) Items(context.Context) ([]d2dctl.Item, *d2dctl.Reply, error) {
	var out []d2dctl.Item
	for _, it := range g.items {
		out = append(out, d2dctl.Item{Code: it.code, Attrs: map[string]string{"id": strconv.Itoa(it.id), "loc": strconv.Itoa(it.loc)}})
	}
	return out, nil, nil
}

func (g *fakeGame) ItemFull(_ context.Context, _ int) ([]d2dctl.VaultRecord, error) {
	var out []d2dctl.VaultRecord
	for _, it := range g.items {
		r := d2dctl.FakeRecord(it.id, it.code, it.quality, 12, fmt.Sprint("roll", it.roll))
		r.Identified = !it.unid
		r.Props = "res-fire:0:" + strconv.Itoa(10+it.roll)
		out = append(out, r)
	}
	return out, nil
}

func chainStub() (chainFunc, error) {
	return func(code string) ([]string, string) { return []string{"helm", "armo"}, "Name of " + code }, nil
}

func seedScenario(class string, pre scenario.Pre) scenario.Scenario {
	return scenario.Scenario{ID: "seedtest", Pre: pre, Matrix: scenario.Matrix{Classes: []string{class}, Seeds: []int64{11}}}
}

func TestApplySeedsVaultAndWorn(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		t.Run(class, func(t *testing.T) {
			dir := t.TempDir()
			g := &fakeGame{}
			s := seedScenario(class, scenario.Pre{
				Worn:  []scenario.WornItem{{Code: "cap", Slot: "head", Quality: 2}},
				Vault: &scenario.VaultSeed{Count: 2, Pinned: 1, Items: []scenario.VaultIt{{Code: "cap", Quality: 4, Unid: true}, {Code: "rin", Quality: 4}}},
			})
			su, err := scenario.BuildSetup(s, class, 11)
			if err != nil {
				t.Fatal(err)
			}
			hero := scenario.HeroName(s.ID, class, 11)
			ledger := filepath.Join(dir, "ledger.jsonl")
			out, err := applySeeds(context.Background(), g, su, ledger, dir, chainStub, hero)
			if err != nil {
				t.Fatal(err)
			}
			if len(out.VaultIDs) != 4 || out.Seq < 5 {
				t.Fatalf("ids %v seq %d (4 adds and one pin make 5 events)", out.VaultIDs, out.Seq)
			}
			// the worn helm stays on the hero, the vault records left it
			if len(g.items) != 1 || g.items[0].loc != 1 || g.items[0].code != "cap" {
				t.Fatalf("hero carries %+v, want one worn cap", g.items)
			}
			st, err := vault.OpenHero(filepath.Join(dir, "vaults"), hero, vault.Options{})
			if err != nil {
				t.Fatal(err)
			}
			defer st.Close()
			r0, ok := st.Get(out.VaultIDs[0])
			if !ok || r0.Identified || r0.Sealed == "" || r0.Props != "" || r0.Code != "cap" || len(r0.Chain) == 0 {
				t.Fatalf("unidentified record 0 = %+v", r0)
			}
			r1, _ := st.Get(out.VaultIDs[1])
			if !r1.Identified || r1.Props == "" {
				t.Fatalf("identified record 1 = %+v", r1)
			}
			if r0.Pin != true || r1.Pin {
				t.Fatalf("pin: record 0 %v record 1 %v, want only the first", r0.Pin, r1.Pin)
			}
			lf, _ := os.ReadFile(ledger)
			for _, want := range []string{"item equip 101 1", "SETUP: scenario seedtest (" + class + ", seed 11)", "item vault"} {
				if !strings.Contains(string(lf), want) {
					t.Errorf("ledger lacks %q:\n%s", want, lf)
				}
			}
		})
	}
}

func TestApplySeedsMustFail(t *testing.T) {
	cases := []struct {
		name string
		g    *fakeGame
		pre  scenario.Pre
		want string
	}{
		{"equip verb that does nothing", &fakeGame{noEquip: true}, scenario.Pre{Worn: []scenario.WornItem{{Code: "cap", Slot: "head"}}}, "not worn"},
		{"give that creates nothing", &fakeGame{noGive: true}, scenario.Pre{Worn: []scenario.WornItem{{Code: "cap", Slot: "head"}}}, "no new cap"},
		{"vault give that creates nothing", &fakeGame{noGive: true}, scenario.Pre{Vault: &scenario.VaultSeed{Items: []scenario.VaultIt{{Code: "cap"}}}}, "no new cap"},
		{"a named duplicate is an error", &fakeGame{dupes: true}, scenario.Pre{Vault: &scenario.VaultSeed{Items: []scenario.VaultIt{{Code: "cap", N: 2}}}}, "identical item"},
	}
	for _, class := range []string{"Paladin", "Sorceress"} {
		for _, c := range cases {
			t.Run(class+"/"+c.name, func(t *testing.T) {
				su, err := scenario.BuildSetup(seedScenario(class, c.pre), class, 3)
				if err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				_, err = applySeeds(context.Background(), c.g, su, filepath.Join(dir, "l"), dir, chainStub, "PgX")
				if err == nil || !strings.Contains(err.Error(), c.want) {
					t.Fatalf("error %v, want it to mention %q", err, c.want)
				}
			})
		}
	}
}

func TestApplySeedsFillersAreDistinct(t *testing.T) {
	dir := t.TempDir()
	g := &fakeGame{dupes: true}
	n := len(scenario.FillerCodes) + 5 // past one round: the second round is magic
	su, err := scenario.BuildSetup(seedScenario("Sorceress", scenario.Pre{Vault: &scenario.VaultSeed{Count: n, Quality: "normal"}}), "Sorceress", 3)
	if err != nil {
		t.Fatal(err)
	}
	out, err := applySeeds(context.Background(), g, su, filepath.Join(dir, "l"), dir, chainStub, "PgFill")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, id := range out.VaultIDs {
		seen[id] = true
	}
	if len(out.VaultIDs) != n || len(seen) != n || len(g.items) != 0 {
		t.Fatalf("%d ids (%d distinct), hero carries %d items; want %d, %d, 0", len(out.VaultIDs), len(seen), len(g.items), n, n)
	}
}

func TestSeedEtherCreditsTheWallet(t *testing.T) {
	state := t.TempDir()
	if err := seedEther(state, 35); err != nil {
		t.Fatal(err)
	}
	w, err := ether.Open(filepath.Join(state, "currency"), ether.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if b, err := w.Balance(); err != nil || b != 35 {
		t.Fatalf("balance %d, %v; want 35", b, err)
	}
	if err := seedEther(t.TempDir(), 0); err != nil {
		t.Fatal(err)
	}
}

// templated calls: a vault id in the path and body, a backpack id (a number) in the body, and a call that cannot resolve.
func TestRunAPICallsResolvesPlaceholders(t *testing.T) {
	var mu sync.Mutex
	var got []string
	view := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"backpack":[{"id":7,"code":"isc"},{"id":42,"code":"cap"},{"id":43,"code":"cap"}],"equipment":{"head":{"id":9,"code":"cap"}}}`)
	}))
	defer view.Close()
	ans := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		got = append(got, r.Method+" "+r.URL.Path+" "+strings.TrimSpace(string(b)))
		mu.Unlock()
		io.WriteString(w, `{"ok":true,"state":"queued"}`)
	}))
	defer ans.Close()
	calls := []scenario.APICall{
		{Path: "/api/vault/equip/{{vault.1.id}}", Method: "GET"},
		{Path: "/api/vault/equip", Body: map[string]any{"ids": []any{"{{vault.0.id}}"}, "slot": "auto"}},
		{Path: "/api/inventory/equip", Body: map[string]any{"id": "{{pack.item[code=cap,nth=1].id}}", "slot": "head"}},
		{Path: "/api/vault/equip", Body: map[string]any{"ids": []any{"{{vault.5.id}}"}}},
	}
	var log strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	reps := runAPICalls(ctx, ans.URL, view.URL, "tok", calls, []string{"v-aa", "v-bb"}, &log)
	want := []string{
		"GET /api/vault/equip/v-bb null",
		`POST /api/vault/equip {"ids":["v-aa"],"slot":"auto"}`,
		`POST /api/inventory/equip {"id":43,"slot":"head"}`,
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("sent\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if len(reps) != 4 || reps[3].Code != "template" || reps[3].Status != 0 || !strings.Contains(reps[3].Reason, "vault.5.id") {
		t.Fatalf("the unresolved call must record a template reply, got %+v", reps)
	}
	if !strings.Contains(string(reps[0].Body), `"state":"queued"`) {
		t.Fatalf("reply body not kept: %q", reps[0].Body)
	}
}
