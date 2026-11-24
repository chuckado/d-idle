package watch

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The vault API at the size of a real vault (capacity 400..2400): what one poll (limit=1, every 6 s) and one page cost on the
// server and on the wire. See docs/probes/vault-perf-2026-11-23.md.
func benchVault(b *testing.B, n int, path string) {
	b.Helper()
	s := testVault(b, n, n+10)
	h := massAPI(b, s, &stubActor{st: s, free: true}, true).Handler()
	b.ReportAllocs()
	b.ResetTimer()
	var size int
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		size = w.Body.Len()
	}
	b.ReportMetric(float64(size), "bytes")
}

func BenchmarkVaultPoll(b *testing.B) {
	for _, n := range []int{400, 2400} {
		b.Run(fmt.Sprint(n), func(b *testing.B) { benchVault(b, n, "/api/vault?limit=1") })
	}
}

func BenchmarkVaultPage(b *testing.B) {
	for _, n := range []int{400, 2400} {
		b.Run(fmt.Sprint(n), func(b *testing.B) { benchVault(b, n, "/api/vault?limit=200&facets=1") })
	}
}

func gzGet(h http.Handler, path string, hdr map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", path, nil)
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// Must-fail guards: the Vault's files are validated (304) and gzipped, a page of 200 rows stays small on the wire, a poll stays tiny.
func TestVaultWireWeight(t *testing.T) {
	s := testVault(t, 2400, 2500)
	h := massAPI(t, s, &stubActor{st: s, free: true}, true).Handler()
	gz := map[string]string{"Accept-Encoding": "gzip"}

	for _, p := range []string{"/watch/vault.js", "/watch/vault-logic.js", "/watch/vault.css", "/watch/vault.json"} {
		w := gzGet(h, p, gz)
		if w.Header().Get("Content-Encoding") != "gzip" || w.Header().Get("ETag") == "" || w.Header().Get("Cache-Control") == "no-store" {
			t.Errorf("%s: encoding %q etag %q cache %q", p, w.Header().Get("Content-Encoding"), w.Header().Get("ETag"), w.Header().Get("Cache-Control"))
		}
		raw := gzGet(h, p, nil)
		if raw.Header().Get("Content-Encoding") != "" || raw.Body.Len() <= w.Body.Len() {
			t.Errorf("%s: plain %d bytes, gzip %d", p, raw.Body.Len(), w.Body.Len())
		}
		if again := gzGet(h, p, map[string]string{"If-None-Match": w.Header().Get("ETag")}); again.Code != http.StatusNotModified || again.Body.Len() != 0 {
			t.Errorf("%s: revalidation answered %d with %d bytes", p, again.Code, again.Body.Len())
		}
	}

	page := gzGet(h, "/api/vault?limit=200&facets=1", gz)
	if page.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("the list is not gzipped")
	}
	wire := page.Body.Len()
	zr, err := gzip.NewReader(page.Body)
	if err != nil {
		t.Fatal(err)
	}
	plain, _ := io.ReadAll(zr)
	var v VaultView
	if err := json.Unmarshal(plain, &v); err != nil || len(v.Items) != 200 {
		t.Fatalf("decoded list: %v, %d items", err, len(v.Items))
	}
	t.Logf("200 rows: %d bytes plain, %d gzip", len(plain), wire)
	if len(plain) > 60_000 || wire > 8_000 {
		t.Errorf("a page of 200 rows weighs %d plain, %d gzip (ceilings 60000, 8000)", len(plain), wire)
	}
	if poll := gzGet(h, "/api/vault?limit=1", nil); poll.Body.Len() > 4_000 {
		t.Errorf("a poll weighs %d bytes (ceiling 4000)", poll.Body.Len())
	}
}

// A poll (every 6 s per open page) must not scale into milliseconds with the vault; the ceiling is ten times today's number.
func TestVaultPollCost(t *testing.T) {
	s := testVault(t, 2400, 2500)
	h := massAPI(t, s, &stubActor{st: s, free: true}, true).Handler()
	res := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gzGet(h, "/api/vault?limit=1", nil)
		}
	})
	t.Logf("poll at 2400 records: %d ns/op", res.NsPerOp())
	if res.NsPerOp() > 10_000_000 {
		t.Errorf("a poll costs %d ns at 2400 records (ceiling 10 ms)", res.NsPerOp())
	}
}
