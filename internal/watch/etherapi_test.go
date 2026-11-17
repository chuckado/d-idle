package watch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"d2lod-bot/internal/ether"
)

func etherSrv(t *testing.T, count *int) (*EtherAPI, *http.ServeMux) {
	t.Helper()
	w, err := ether.Open(t.TempDir(), ether.Options{Content: func() int { return *count }})
	if err != nil {
		t.Fatal(err)
	}
	a := &EtherAPI{W: w}
	m := http.NewServeMux()
	m.Handle(EtherRoute, a.Get())
	m.Handle(UpgradeRoute, a.Upgrade())
	m.Handle(DowngradeRoute, a.Downgrade())
	return a, m
}

func etherDo(m http.Handler, method, path, body string) (int, map[string]any) {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, r)
	var out map[string]any
	json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestEtherAPIFlow(t *testing.T) {
	count := 100
	a, m := etherSrv(t, &count)

	code, v := etherDo(m, "GET", EtherRoute, "")
	if code != 200 || v["schema"] != etherSchema || v["balance"].(float64) != 0 || v["chip"] != "Ether 0" {
		t.Fatalf("get: %d %v", code, v)
	}
	if next := v["next"].(map[string]any); next["id"] != "v1" || next["canAfford"] != false || next["short"].(float64) != 20 {
		t.Fatalf("next: %v", next)
	}
	if code, v = etherDo(m, "POST", UpgradeRoute, ""); code != 402 || v["code"] != "insufficient" || v["need"].(float64) != 20 {
		t.Fatalf("short: %d %v", code, v)
	}
	if code, _ = etherDo(m, "GET", UpgradeRoute, ""); code != 405 {
		t.Fatalf("GET upgrade: %d", code)
	}
	if code, _ = etherDo(m, "POST", EtherRoute, ""); code != 405 {
		t.Fatalf("POST ether: %d", code)
	}
	a.W.Earn("quest", "q1", "hell", 50)

	code, v = etherDo(m, "POST", UpgradeRoute, `{"step":"v1"}`)
	if code != 200 || v["ok"] != true || v["capacity"].(float64) != 600 || v["balance"].(float64) != 30 {
		t.Fatalf("upgrade: %d %v", code, v)
	}
	code, v = etherDo(m, "POST", UpgradeRoute, `{"step":"v1"}`)
	if code != 200 || v["already"] != true || v["balance"].(float64) != 30 {
		t.Fatalf("double click: %d %v", code, v)
	}
	if code, v = etherDo(m, "POST", UpgradeRoute, `{"step":"v3"}`); code != 409 || v["code"] != "out-of-order" {
		t.Fatalf("skip: %d %v", code, v)
	}
	if code, _ = etherDo(m, "POST", UpgradeRoute, `{bad`); code != 400 {
		t.Fatalf("bad body: %d", code)
	}

	count = 601
	code, v = etherDo(m, "POST", DowngradeRoute, "")
	if code != 409 || v["code"] != "in-use" || v["count"].(float64) != 601 || v["fitsTier"].(float64) != 400 {
		t.Fatalf("in use: %d %v", code, v)
	}
	_, g := etherDo(m, "GET", EtherRoute, "")
	if d := g["downgrade"].(map[string]any); d["available"] != false || d["reason"] == "" {
		t.Fatalf("downgrade view: %v", d)
	}
	count = 399
	if code, v = etherDo(m, "POST", DowngradeRoute, ""); code != 200 || v["balance"].(float64) != 50 || v["capacity"].(float64) != 400 {
		t.Fatalf("downgrade: %d %v", code, v)
	}
	if code, v = etherDo(m, "POST", DowngradeRoute, ""); code != 409 || v["code"] != "nothing-to-refund" {
		t.Fatalf("again: %d %v", code, v)
	}
}
