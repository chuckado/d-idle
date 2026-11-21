package watch

import (
	"context"
	"net/http/httptest"
	"testing"
)

type fakeEquip struct {
	on     bool
	orders map[string]string // id -> slot
}

func (f *fakeEquip) Gate(context.Context) (bool, string, string) {
	if !f.on {
		return false, "off", "one-tap Equip is switched off"
	}
	return true, "", ""
}
func (f *fakeEquip) Preview(id string) (any, bool) {
	if id == "none" {
		return nil, false
	}
	return map[string]any{"enabled": f.on, "wearable": true, "slots": []any{}}, true
}
func (f *fakeEquip) Enqueue(_ context.Context, id, slot string, _ bool) VaultKeepResult {
	if slot == "" {
		slot = "auto"
	}
	if f.orders == nil {
		f.orders = map[string]string{}
	}
	if old, ok := f.orders[id]; ok {
		if old == slot {
			return VaultKeepResult{ID: id, OK: true, Already: true}
		}
		return VaultKeepResult{ID: id, Code: "busy", Why: "other slot"}
	}
	f.orders[id] = slot
	return VaultKeepResult{ID: id, OK: true}
}
func (f *fakeEquip) Cancel(id string) VaultKeepResult {
	delete(f.orders, id)
	return VaultKeepResult{ID: id, OK: true}
}

func equipDesk(t *testing.T, on bool) (*VaultAPI, *fakeEquip) {
	s := testVault(t, 10, 100)
	a := vaultAPI(s, true)
	fe := &fakeEquip{on: on}
	d := a.desk()
	d.Equip = fe
	return a, fe
}

func TestVaultEquipRoutes(t *testing.T) {
	a, fe := equipDesk(t, true)
	_, v := getVault(t, a, "/api/vault?limit=2")
	if v.Equip == nil || !v.Equip.Enabled {
		t.Fatalf("flag %+v", v.Equip)
	}
	id := v.Items[0].ID
	code, rep, _ := postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{id}, "slot": "left"})
	if code != 200 || !rep.OK || fe.orders[id] != "left" {
		t.Fatalf("%d %+v", code, rep)
	}
	code, rep, _ = postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{id}, "slot": "left"})
	if code != 200 || len(rep.Already) != 1 {
		t.Fatalf("idempotent: %d %+v", code, rep)
	}
	code, _, pr := postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{id}, "slot": "right"})
	if code != 409 {
		t.Fatalf("a second slot for the same item: %d %+v", code, pr)
	}
	if code, _, pr = postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{id}, "slot": "up"}); code != 400 || pr.Code != "slot" {
		t.Fatalf("bad slot: %d %+v", code, pr)
	}
	if code, _, pr = postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{id, "x"}}); code != 400 {
		t.Fatalf("two ids: %d %+v", code, pr)
	}
	if code, rep, _ = postVault(t, a, VaultEquipCancel, map[string]any{"ids": []string{id}}); code != 200 || fe.orders[id] != "" {
		t.Fatalf("cancel: %d %+v", code, rep)
	}
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", VaultEquipGet+id, nil))
	if w.Code != 200 {
		t.Fatalf("get %d", w.Code)
	}
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", VaultEquipGet+"none", nil))
	if w.Code != 404 {
		t.Fatalf("missing %d", w.Code)
	}
}

func TestVaultEquipOffRefuses(t *testing.T) {
	a, fe := equipDesk(t, false)
	_, v := getVault(t, a, "/api/vault?limit=2")
	if v.Equip == nil || v.Equip.Enabled {
		t.Fatalf("flag %+v", v.Equip)
	}
	code, _, pr := postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{v.Items[0].ID}})
	if code != 409 || pr.Code != "off" || len(fe.orders) != 0 {
		t.Fatalf("%d %+v", code, pr)
	}
}

func TestVaultWithoutEquipHasNoFlag(t *testing.T) {
	s := testVault(t, 3, 100)
	a := vaultAPI(s, true)
	_, v := getVault(t, a, "/api/vault?limit=2")
	if v.Equip != nil {
		t.Fatal("no Equip on this bot: no flag")
	}
	if code, _, pr := postVault(t, a, VaultEquipRoute, map[string]any{"ids": []string{v.Items[0].ID}}); code != 409 || pr.Code != "off" {
		t.Fatalf("%d %+v", code, pr)
	}
}
