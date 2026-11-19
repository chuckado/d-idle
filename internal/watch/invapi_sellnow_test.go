package watch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/invdesk"
	invstatus "d2lod-bot/internal/status"
)

// sellInv puts the real desk behind the fake of the other route tests.
type sellInv struct {
	*fakeInv
	d *invdesk.Desk
}

func (s *sellInv) RequestSell() (any, error) { return s.d.RequestSell() }

type seller struct{}

func (seller) SellMarked(context.Context) (string, error) { return "sold", nil }

func sellDesk(t *testing.T, class d2.Class, mark bool) *invdesk.Desk {
	t.Helper()
	snap := &d2.Snapshot{Player: &d2.Player{Class: class, Area: 2}, Inventory: &d2.Inventory{Items: []d2.Unit{
		{ID: 1, Class: 10, Item: &d2.Item{Where: d2.ContInventory, Quality: 4, ILvl: 5}},
	}}}
	d, err := invdesk.New(invdesk.Config{
		Path:      filepath.Join(t.TempDir(), invdesk.StateFile),
		Snap:      func() *d2.Snapshot { return snap },
		Inventory: func() invstatus.InventoryData { return invstatus.InventoryData{} },
		Now:       func() time.Time { return time.Unix(1_800_000_000, 0) },
	})
	if err != nil {
		t.Fatal(err)
	}
	d.SetSeller(seller{})
	if mark {
		if err := d.SetMark(1, invdesk.MarkSell); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func sellGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if os.Getenv("UPDATE") != "" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Errorf("%s differs\n got: %s\nwant: %s", path, got, want)
	}
}

// The sell-now route: the gates of the other inventory routes, 503 until the desk is bound, code off on a desk without
// the part, 409 with the reason when nothing is marked, and the queued request as the reply. Goldens for both classes.
func TestInventorySellNowRoute(t *testing.T) {
	auth := map[string]string{"Authorization": "Bearer tok"}
	opts := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}

	unbound := NewInvAPI()
	if w := invPost(unbound.PostHandler(opts), RouteSellNow, `{}`, auth); w.Code != 503 || !strings.Contains(w.Body.String(), `"off"`) {
		t.Errorf("unbound: %d %s", w.Code, w.Body)
	}
	plain := NewInvAPI()
	plain.Bind(&fakeInv{})
	if w := invPost(plain.PostHandler(opts), RouteSellNow, `{}`, auth); w.Code != 409 || !strings.Contains(w.Body.String(), `"off"`) {
		t.Errorf("no sale part: %d %s", w.Code, w.Body)
	}
	found := false
	for _, r := range InvPostRoutes {
		found = found || r == RouteSellNow
	}
	if !found {
		t.Error("route missing from InvPostRoutes")
	}

	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		api := NewInvAPI()
		api.Bind(&sellInv{fakeInv: &fakeInv{}, d: sellDesk(t, class, false)})
		post := api.PostHandler(opts)
		if w := invPost(post, RouteSellNow, `{}`, nil); w.Code != 401 {
			t.Errorf("no token: %d", w.Code)
		}
		if w := invPost(post, RouteSellNow, `{"bogus":1}`, auth); w.Code != 400 {
			t.Errorf("unknown field: %d", w.Code)
		}
		w := invPost(post, RouteSellNow, `{}`, auth)
		if w.Code != 409 {
			t.Fatalf("%v nothing marked: %d %s", class, w.Code, w.Body)
		}
		sellGolden(t, "inventory_sellnow_nothing.json", w.Body.String())

		api2 := NewInvAPI()
		api2.Bind(&sellInv{fakeInv: &fakeInv{}, d: sellDesk(t, class, true)})
		w = invPost(api2.PostHandler(opts), RouteSellNow, `{}`, auth)
		if w.Code != 200 {
			t.Fatalf("%v: %d %s", class, w.Code, w.Body)
		}
		sellGolden(t, "inventory_sellnow_queued_"+class.String()+".json", w.Body.String())
	}
}
