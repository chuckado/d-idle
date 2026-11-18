package watch

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	item "d2lod-bot/internal/status"
	"d2lod-bot/internal/vault"
)

const secretAffix = "SECRET-AFFIX-TEXT"

// leakyTip is a tooltip source that words its props line from whatever it is handed and also returns it when the row is
// unidentified: the API must not let that reach the payload.
func leakyTip(code string, q int, identified, eth bool, props string) VaultTip {
	t := VaultTip{Base: &item.ItemBase{Name: code, MinDam: 1, MaxDam: 2}, Art: "inv" + code, Sell: 7}
	t.Props = []string{secretAffix + " " + props, secretAffix + " always"}
	return t
}

func tipVault(t *testing.T) (*VaultAPI, string, string) {
	t.Helper()
	s, err := vault.OpenHero(t.TempDir(), "Sorc", vault.Options{Capacity: func() int { return 100 }, NoSync: true,
		Now: func() time.Time { return time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC) }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	add := func(raw, code string, id bool) string {
		r, err := s.Add(vault.Item{Hero: "Sorc", Raw: raw, Code: code, Name: code, Quality: 4, Ilvl: 30, Identified: id, Props: "7:0:20;", Chain: []string{"weap"}})
		if err != nil {
			t.Fatal(err)
		}
		return r.ID
	}
	// two kinds: a weapon and a jewel, each identified and not
	idW, unW := add("aa01", "zvb", true), add("aa02", "zvb", false)
	idJ, unJ := add("bb01", "jew", true), add("bb02", "jew", false)
	_ = idW
	_ = idJ
	tiers, _ := vault.DefaultTiers()
	a := NewVaultAPI()
	a.Bind(VaultDesk{On: func() bool { return true }, Store: func() *vault.Store { return s }, Tiers: tiers, Tip: leakyTip})
	return a, unW, unJ
}

// An unidentified magic item never exposes affix text anywhere in the API payload (list or detail), even when the tip source
// hands it out; an identified one shows its lines. Weapon and jewel.
func TestVaultUnidentifiedNeverExposesAffixes(t *testing.T) {
	a, unW, unJ := tipVault(t)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault", nil))
	var v VaultView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Items) != 4 {
		t.Fatalf("%d rows", len(v.Items))
	}
	for _, it := range v.Items {
		if it.Base == nil || it.Art == "" || it.SellPrice != 7 {
			t.Errorf("%s: tooltip facts missing: %+v", it.ID, it)
		}
		if it.Identified && len(it.Props) != 2 {
			t.Errorf("%s: identified row has %d stat lines", it.ID, len(it.Props))
		}
		if !it.Identified && len(it.Props) != 0 {
			t.Errorf("%s: unidentified row carries %v", it.ID, it.Props)
		}
	}
	// the identified rows got the stored props handed in, the unidentified ones got none
	if !strings.Contains(w.Body.String(), secretAffix+" 7:0:20;") {
		t.Error("identified props did not reach the tip source")
	}
	for _, id := range []string{unW, unJ} {
		d := httptest.NewRecorder()
		a.Handler().ServeHTTP(d, httptest.NewRequest("GET", "/api/vault/item/"+id, nil))
		if d.Code != 200 || strings.Contains(d.Body.String(), secretAffix) || strings.Contains(d.Body.String(), "7:0:20") {
			t.Errorf("detail of unidentified %s: %d %s", id, d.Code, d.Body)
		}
	}
	// nothing of an unidentified record's text anywhere: cut the identified rows out and scan the rest of the body
	var raw struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &raw)
	for i, row := range raw.Items {
		if string(row["identified"]) == "false" {
			b, _ := json.Marshal(row)
			if strings.Contains(string(b), secretAffix) || strings.Contains(string(b), "7:0:20") || row["props"] != nil {
				t.Errorf("row %d leaks: %s", i, b)
			}
		}
	}
}

// The payload stays bounded: a page of rows, not the whole vault, carries the facts.
func TestVaultTipPayloadPaged(t *testing.T) {
	s := testVault(t, 300, 400)
	a := vaultAPI(s, true)
	d := a.desk()
	d.Tip = func(code string, q int, id, eth bool, props string) VaultTip {
		return VaultTip{Base: &item.ItemBase{Name: code, MinAC: 5, MaxAC: 9, ReqStr: 20}, Art: "inv" + code, Sell: 9}
	}
	a.Bind(*d)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/vault?limit=60", nil))
	if w.Code != 200 || w.Body.Len() > 60*700 {
		t.Fatalf("%d bytes for a page of 60", w.Body.Len())
	}
}
