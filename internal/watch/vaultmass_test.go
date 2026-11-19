package watch

import (
	"encoding/json"
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
