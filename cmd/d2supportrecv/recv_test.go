package main

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const tok = "a-long-random-upload-token-123456"

func mkZip(t *testing.T, names ...string) []byte {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte("hello"))
	}
	zw.Close()
	return b.Bytes()
}

func setup(t *testing.T, mod func(*config)) (*server, config) {
	dir := t.TempDir()
	tf := filepath.Join(dir, "tokens")
	os.WriteFile(tf, []byte("# comment\n"+hashToken(tok)+" pack\n"), 0o600)
	c := config{dir: filepath.Join(dir, "inbox"), tokensFile: tf, quota: 1 << 30, maxFiles: 100, tokenRate: 100, ipRate: 100, failRate: 100}
	if mod != nil {
		mod(&c)
	}
	return newServer(c), c
}

func post(s *server, body []byte, auth, ct string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/support/upload", bytes.NewReader(body))
	if auth != "" {
		r.Header.Set("Authorization", auth)
	}
	r.Header.Set("Content-Type", ct)
	r.RemoteAddr = "203.0.113.5:4000"
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestUploadOK(t *testing.T) {
	s, c := setup(t, nil)
	w := post(s, mkZip(t, "log.txt", "save/character1.d2s"), "Bearer "+tok, "application/zip")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	code := strings.Split(w.Body.String(), `"`)[3]
	if !codeRx.MatchString(code) {
		t.Fatal(code)
	}
	st, err := os.Stat(filepath.Join(c.dir, code+".zip"))
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatal(err, st)
	}
}

func TestAuth(t *testing.T) {
	s, c := setup(t, nil)
	z := mkZip(t, "a.txt")
	for _, a := range []string{"", "Bearer ", "Bearer nope", "Basic " + tok, tok} {
		if w := post(s, z, a, "application/zip"); w.Code != 401 {
			t.Errorf("auth %q: %d", a, w.Code)
		}
	}
	if ents, _ := os.ReadDir(c.dir); len(ents) != 0 {
		t.Fatal("nothing may be stored without a token")
	}
	// revocation: the line goes, the next request is refused without a restart
	if w := post(s, z, "Bearer "+tok, "application/zip"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(c.tokensFile, []byte("\n"), 0o600)
	os.Chtimes(c.tokensFile, time.Now().Add(time.Minute), time.Now().Add(time.Minute))
	if w := post(s, z, "Bearer "+tok, "application/zip"); w.Code != 401 {
		t.Fatal("revoked token still works", w.Code)
	}
	os.Remove(c.tokensFile)
	if w := post(s, z, "Bearer "+tok, "application/zip"); w.Code != 401 {
		t.Fatal("no token file must refuse everyone")
	}
}

func TestContentAndSize(t *testing.T) {
	s, _ := setup(t, nil)
	if w := post(s, mkZip(t, "a"), "Bearer "+tok, "text/plain"); w.Code != 415 {
		t.Error(w.Code)
	}
	if w := post(s, []byte("not a zip at all"), "Bearer "+tok, "application/zip"); w.Code != 400 {
		t.Error(w.Code)
	}
	if w := post(s, append([]byte("PK\x03\x04"), make([]byte, 100)...), "Bearer "+tok, "application/zip"); w.Code != 400 {
		t.Error(w.Code)
	}
	if w := post(s, make([]byte, maxZip+10), "Bearer "+tok, "application/zip"); w.Code != 413 {
		t.Error(w.Code)
	}
	if r := httptest.NewRecorder(); true {
		s.ServeHTTP(r, httptest.NewRequest("GET", "/support/upload", nil))
		if r.Code != 405 {
			t.Error(r.Code)
		}
	}
}

func TestPathTraversalNames(t *testing.T) {
	s, c := setup(t, nil)
	for _, n := range []string{"../evil", "a/../../evil", "/abs", `a\b`, "a//b", "dir/", "./x"} {
		if w := post(s, mkZip(t, n), "Bearer "+tok, "application/zip"); w.Code != 400 {
			t.Errorf("%q accepted: %d", n, w.Code)
		}
	}
	if ents, _ := os.ReadDir(c.dir); len(ents) != 0 {
		t.Fatalf("stored: %v", ents)
	}
	// the stored name is only ever the generated code
	w := post(s, mkZip(t, "ok.txt"), "Bearer "+tok, "application/zip")
	ents, _ := os.ReadDir(c.dir)
	if w.Code != 200 || len(ents) != 1 || !codeRx.MatchString(strings.TrimSuffix(ents[0].Name(), ".zip")) {
		t.Fatal(w.Code, ents)
	}
}

func TestEntryLimit(t *testing.T) {
	s, _ := setup(t, nil)
	var names []string
	for i := 0; i < maxEntries+1; i++ {
		names = append(names, "f"+string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	if w := post(s, mkZip(t, names...), "Bearer "+tok, "application/zip"); w.Code != 400 {
		t.Fatal(w.Code)
	}
}

func TestRateLimits(t *testing.T) {
	now := time.Now()
	s, _ := setup(t, func(c *config) { c.ipRate, c.tokenRate, c.failRate = 4, 2, 2; c.now = func() time.Time { return now } })
	z := mkZip(t, "a")
	if post(s, z, "Bearer "+tok, "application/zip").Code != 200 || post(s, z, "Bearer "+tok, "application/zip").Code != 200 {
		t.Fatal("first two")
	}
	if w := post(s, z, "Bearer "+tok, "application/zip"); w.Code != 429 {
		t.Fatal("per-token limit", w.Code)
	}
	// the per-IP budget (4) is spent after one more request; a bad token also counts
	post(s, z, "Bearer bad", "application/zip")
	if w := post(s, z, "Bearer "+tok, "application/zip"); w.Code != 429 {
		t.Fatal("per-ip limit", w.Code)
	}
	now = now.Add(2 * time.Hour) // windows end
	if w := post(s, z, "Bearer "+tok, "application/zip"); w.Code != 200 {
		t.Fatal("window did not reset", w.Code)
	}
}

func TestFailRate(t *testing.T) {
	s, _ := setup(t, func(c *config) { c.failRate = 2 })
	z := mkZip(t, "a")
	codes := []int{}
	for i := 0; i < 4; i++ {
		codes = append(codes, post(s, z, "Bearer bad", "application/zip").Code)
	}
	if codes[0] != 401 || codes[1] != 401 || codes[2] != 429 {
		t.Fatal(codes)
	}
}

func TestQuotaPurgeOldestFirst(t *testing.T) {
	s, c := setup(t, func(c *config) { c.quota = 5000; c.maxFiles = 100 })
	os.MkdirAll(c.dir, 0o700)
	old := []string{"AAAAAAAA", "BBBBBBBB", "CCCCCCCC"}
	for i, n := range old {
		p := filepath.Join(c.dir, n+".zip")
		os.WriteFile(p, make([]byte, 2000), 0o600)
		tm := time.Now().Add(time.Duration(i-10) * time.Hour)
		os.Chtimes(p, tm, tm)
	}
	os.WriteFile(filepath.Join(c.dir, "notes.txt"), make([]byte, 9000), 0o600) // not ours: never purged, never counted
	w := post(s, mkZip(t, "a"), "Bearer "+tok, "application/zip")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	if _, err := os.Stat(filepath.Join(c.dir, "AAAAAAAA.zip")); err == nil {
		t.Error("oldest file survived")
	}
	if _, err := os.Stat(filepath.Join(c.dir, "CCCCCCCC.zip")); err != nil {
		t.Error("newest old file was purged")
	}
	if _, err := os.Stat(filepath.Join(c.dir, "notes.txt")); err != nil {
		t.Error("foreign file purged")
	}
	// the max-files limit works the same way, and the new upload itself is never purged
	s2, c2 := setup(t, func(c *config) { c.maxFiles = 2 })
	var last string
	for i := 0; i < 4; i++ {
		last = strings.Split(post(s2, mkZip(t, "a"), "Bearer "+tok, "application/zip").Body.String(), `"`)[3]
		time.Sleep(15 * time.Millisecond)
	}
	ents, _ := os.ReadDir(c2.dir)
	if len(ents) != 2 {
		t.Fatalf("files: %d", len(ents))
	}
	if _, err := os.Stat(filepath.Join(c2.dir, last+".zip")); err != nil {
		t.Error("newest upload purged")
	}
}

func TestProxyAddress(t *testing.T) {
	s, _ := setup(t, func(c *config) { c.trustProxy = true; c.ipRate = 1 })
	z := mkZip(t, "a")
	mk := func(ip string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/", bytes.NewReader(z))
		r.Header.Set("Authorization", "Bearer "+tok)
		r.Header.Set("Content-Type", "application/zip")
		r.Header.Set("CF-Connecting-IP", ip)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	if mk("198.51.100.1").Code != 200 || mk("198.51.100.2").Code != 200 || mk("198.51.100.1").Code != 429 {
		t.Fatal("per-client limits through the proxy header")
	}
}

func TestNoLogging(t *testing.T) {
	src, _ := os.ReadFile("recv.go")
	if strings.Contains(string(src), "log.") {
		t.Fatal("recv.go must not log")
	}
	var _ http.Handler = (*server)(nil)
}
