package watch

import (
	"context"
	"encoding/binary"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/charsel"
)

// The Characters routes run over a real charsel.Manager on a temp save folder with two header-only saves (a Paladin and a
// Sorceress), so the goldens are the shape the page will read. Fixtures are only added, never edited (UPDATE_GOLDEN=1
// writes missing ones).

func charSave(t *testing.T, dir, name string, class, level byte) {
	b := make([]byte, 0x40)
	binary.LittleEndian.PutUint32(b[0x00:], 0xAA55AA55)
	binary.LittleEndian.PutUint32(b[0x04:], 0x60)
	copy(b[0x14:], name)
	b[0x24], b[0x28], b[0x2B] = 0x20, class, level
	binary.LittleEndian.PutUint32(b[0x30:], 1763200000)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".d2s"), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func charRig(t *testing.T, measured bool) (*CharactersAPI, *charsel.Manager) {
	root := t.TempDir()
	save := filepath.Join(root, "save")
	charSave(t, save, "Holy", 3, 24)
	charSave(t, save, "Zapp", 1, 9)
	now := func() time.Time { return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC) }
	m := &charsel.Manager{Store: &charsel.Store{SaveDir: save, ArchiveDir: filepath.Join(root, "arch"), StateRoot: filepath.Join(root, "inst", "characters"), Now: now},
		Inst: filepath.Join(root, "inst"), Active: func() string { return "Holy" }, Supervised: true, Measured: measured, Now: now,
		Leave: func(context.Context) error { return nil }, Verify: func(string) error { return nil }}
	os.MkdirAll(m.Inst, 0o755)
	api := NewCharactersAPI()
	api.Bind(m)
	return api, m
}

func charGolden(t *testing.T, name, got string) {
	path := filepath.Join("testdata", "characters", name+".json")
	if _, err := os.Stat(path); err != nil && os.Getenv("UPDATE_GOLDEN") != "" {
		os.MkdirAll(filepath.Dir(path), 0o755)
		os.WriteFile(path, []byte(got), 0o644)
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Errorf("%s:\nwant %s\ngot  %s", name, want, got)
	}
}

func TestCharactersGoldens(t *testing.T) {
	auth := map[string]string{"Authorization": "Bearer tok"}
	o := AnswerOptions{Token: "tok"}
	get := func(api *CharactersAPI) string {
		w := httptest.NewRecorder()
		api.Handler().ServeHTTP(w, httptest.NewRequest("GET", CharactersRoute, nil))
		if w.Code != 200 {
			t.Fatalf("GET %d %s", w.Code, w.Body)
		}
		return w.Body.String()
	}
	post := func(api *CharactersAPI, route, body string, want int) string {
		w := invPost(api.PostHandler(o, route), route, body, auth)
		if w.Code != want {
			t.Fatalf("%s %s: %d (want %d) %s", route, body, w.Code, want, w.Body)
		}
		return w.Body.String()
	}
	api, m := charRig(t, false)
	charGolden(t, "list-two-classes", get(api))
	charGolden(t, "post-delete-zapp", post(api, CharactersDeleteRoute, `{"name":"Zapp","confirm":"Zapp"}`, 200))
	charGolden(t, "list-after-delete", get(api))
	charGolden(t, "post-restore-zapp", post(api, CharactersRestoreRoute, `{"archive":"Zapp@20261115T120000Z"}`, 200))
	charGolden(t, "refused-delete-active", post(api, CharactersDeleteRoute, `{"name":"Holy","confirm":"Holy"}`, 409))
	charGolden(t, "refused-delete-confirm", post(api, CharactersDeleteRoute, `{"name":"Zapp","confirm":"zapp"}`, 400))
	charGolden(t, "refused-create-exists", post(api, CharactersRoute, `{"name":"zapp","class":"sorceress"}`, 409))
	charGolden(t, "refused-create-name", post(api, CharactersRoute, `{"name":"Zapp9","class":"sorceress"}`, 400))
	charGolden(t, "post-switch-zapp", post(api, CharactersSwitchRoute, `{"name":"Zapp"}`, 200))
	charGolden(t, "list-switching", get(api))
	charGolden(t, "refused-switch-twice", post(api, CharactersSwitchRoute, `{"name":"Zapp"}`, 409))
	if in, ok := charsel.ReadIntent(m.Inst); !ok || in.To != "Zapp" {
		t.Errorf("intent %+v", in)
	}

	mapi, _ := charRig(t, true)
	charGolden(t, "list-measured", get(mapi))
	charGolden(t, "refused-measured-run", post(mapi, CharactersSwitchRoute, `{"name":"Zapp"}`, 409))
	charGolden(t, "refused-create-hardcore", post(func() *CharactersAPI { a, _ := charRig(t, false); return a }(), CharactersRoute, `{"name":"Newb","class":"paladin","hardcore":true}`, 409))
}

func TestCharactersRouteGates(t *testing.T) {
	api := NewCharactersAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	auth := map[string]string{"Authorization": "Bearer tok"}
	for _, r := range CharactersPostRoutes {
		if w := invPost(api.PostHandler(o, r), r, `{}`, auth); w.Code != 503 {
			t.Errorf("%s unbound: %d", r, w.Code)
		}
	}
	w := httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest("GET", CharactersRoute, nil))
	if w.Code != 503 {
		t.Errorf("GET unbound: %d", w.Code)
	}
	api, _ = charRig(t, false)
	for name, c := range map[string]struct {
		route, body string
		hdr         map[string]string
		want        int
	}{
		"no token":      {CharactersSwitchRoute, `{"name":"Zapp"}`, nil, 401},
		"foreign":       {CharactersSwitchRoute, `{"name":"Zapp"}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}, 403},
		"unknown field": {CharactersSwitchRoute, `{"name":"Zapp","x":1}`, auth, 400},
		"wrong type":    {CharactersRoute, `{"name":"Newb","class":"paladin","hardcore":"yes"}`, auth, 400},
		"foreign field": {CharactersSwitchRoute, `{"name":"Zapp","confirm":"Zapp"}`, auth, 400},
		"restore name":  {CharactersRestoreRoute, `{"archive":"A@B","name":"Zapp"}`, auth, 400},
		"path archive":  {CharactersRestoreRoute, `{"archive":"../x"}`, auth, 400},
	} {
		if w := invPost(api.PostHandler(o, c.route), c.route, c.body, c.hdr); w.Code != c.want {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
	w = httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest("POST", CharactersRoute, strings.NewReader("{}")))
	if w.Code != 405 {
		t.Errorf("viewer accepts POST: %d", w.Code)
	}
}
