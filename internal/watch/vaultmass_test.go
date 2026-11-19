package watch

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/vault"
)

func massAPI(t *testing.T, s *vault.Store, a *stubActor, withFilter bool) *VaultAPI {
	t.Helper()
	tiers, _ := vault.DefaultTiers()
	d := VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Tiers: tiers, Act: a, Quality: func(q int) string {
		return [...]string{"?", "low", "normal", "superior", "magic", "set", "rare", "unique", "crafted"}[q]
	}, ScrollGold: 80}
	if withFilter {
		f, err := bulkid.DefaultFilter()
		if err != nil {
			t.Fatal(err)
		}
		d.MassFilter = f
	}
	api := NewVaultAPI()
	api.Bind(d)
	return api
}

// GET /api/vault carries the mass reveal preview with its cost sentence, and an unidentified item's text never appears.
func TestVaultViewMassRevealPreview(t *testing.T) {
	s := testVault(t, 12, 100)
	for _, c := range []struct {
		name string
		a    *stubActor
		mode string
	}{
		{"cain", &stubActor{st: s, free: true}, bulkid.ViaCain},
		{"scrolls", &stubActor{st: s, charges: 1}, bulkid.ViaScroll},
		{"none", &stubActor{st: s}, "none"},
	} {
		code, v := getVault(t, massAPI(t, s, c.a, true), "/api/vault?limit=5")
		if code != 200 || v.Act == nil || v.Act.Reveal.Mass == nil {
			t.Fatalf("%s: no preview", c.name)
		}
		m := v.Act.Reveal.Mass
		if m.Mode != c.mode || m.Line == "" || len(m.IDs) != m.Count {
			t.Fatalf("%s: %+v", c.name, m)
		}
		if c.mode == bulkid.ViaScroll && m.Count > 1 {
			t.Fatalf("scrolls: %d ids for 1 charge", m.Count)
		}
	}
	// No filter table wired: no preview, nothing else changes.
	_, v := getVault(t, massAPI(t, s, &stubActor{st: s, free: true}, false), "/api/vault?limit=5")
	if v.Act == nil || v.Act.Reveal.Mass != nil {
		t.Fatal("preview without a filter")
	}
	// Read only vault: no preview.
	api := massAPI(t, s, &stubActor{st: s, free: true}, true)
	_, v = getVault(t, api, "/api/vault?limit=5")
	b, _ := json.Marshal(v.Act.Reveal.Mass)
	if strings.Contains(string(b), "raw") {
		t.Fatalf("preview leaks: %s", b)
	}
}

// etherStub adds the Ether tap to stubActor: it unlocks sealed records like vaultkeep does and records that the gate was bypassed.
type etherStub struct {
	*stubActor
	taps int
}

func (e *etherStub) EnqueueRevealEther(_ context.Context, ids []string, _ bool) []VaultKeepResult {
	var out []VaultKeepResult
	for _, id := range ids {
		e.taps++
		if err := e.st.Unseal(id); err != nil {
			out = append(out, VaultKeepResult{ID: id, Code: "nosealed", Why: err.Error()})
			continue
		}
		out = append(out, VaultKeepResult{ID: id, OK: true})
	}
	return out
}

// The vault never sends sealed affixes (list, detail), the Ether tap works while the reveal gate is closed (no game
// needed), and the stat lines exist only after the unlock.
func TestVaultEtherRevealUnlocksOnlyOnPayment(t *testing.T) {
	s := testVault(t, 12, 100)
	_, un := getVault(t, massAPI(t, s, &stubActor{st: s}, false), "/api/vault?identified=0&limit=10")
	id := un.Items[0].ID
	rec, _ := s.Get(id)
	it := rec.Item
	it.Raw += "00"
	it.Sealed, it.Name = "SEALEDAFFIX", "sealed one"
	nr, err := s.Add(it)
	if err != nil {
		t.Fatal(err)
	}
	a := &etherStub{stubActor: &stubActor{st: s, revGate: "dead"}}
	tiers, _ := vault.DefaultTiers()
	api := NewVaultAPI()
	api.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Tiers: tiers, Act: a,
		Tip: func(in VaultTipIn) VaultTip { return VaultTip{Props: []string{in.Props}} }})
	w := httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault?identified=0&limit=50", nil))
	if strings.Contains(w.Body.String(), "SEALEDAFFIX") {
		t.Fatal("sealed affixes reached the page before payment")
	}
	if code, rep, _ := postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{nr.ID}, Ether: true}); code != 200 || len(rep.Done) != 1 || a.taps != 1 {
		t.Fatalf("ether tap with the gate closed: %d %+v", code, rep)
	}
	if code, _, pr := postVault(t, api, VaultRevealRoute, vaultBody{IDs: []string{nr.ID}}); code != 409 || pr.Code != "dead" {
		t.Fatalf("a plain tap still obeys the gate: %d %+v", code, pr)
	}
	w = httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault?identified=1&limit=50", nil))
	if !strings.Contains(w.Body.String(), "SEALEDAFFIX") {
		t.Fatal("the affix lines are missing after the unlock")
	}
}
