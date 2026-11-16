package support

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)

// seeded holds one secret of every class; none of them may survive Redact.
var seeded = map[string]string{
	"home":    "opened /home/alice/.d2lod/instances/i1/logs/x",
	"winhome": `C:\Users\Bob\AppData\x`,
	"ip":      "peer 192.168.4.20 and 8.8.4.4",
	"ip6":     "peer fe80:0:0:0:1:2:3:4",
	"email":   "owner jane.doe@example.org wrote",
	"host":    "fetch https://updates.example-owner.test/x and nas.lan",
	"bearer":  "Authorization: Bearer abcdef0123456789",
	"kv":      `watch_token=hunter2hunter2 and "passphrase": "open sesame"`,
	"query":   "GET /watch?token=ZZtopSecret99",
	"packurl": "GET /p/Abcdef0123456789xyz/pack.tar",
	"uuid":    "id 123e4567-e89b-12d3-a456-426614174000",
	"cdkey":   "key ABCD1-2EFGH-3IJKL-4MN",
	"keyfile": "read default.key and my.key",
	"save":    "loading Mychar.d2s now",
	"pem":     "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASC\n-----END PRIVATE KEY-----",
	"long":    "tok Xk3jD9sLq2Wv8Zr1Tn5Yb7Hc4Gf6Mp0QaUe",
	"hex":     "tok 0123456789abcdef0123456789abcdef0123",
}

func TestRedactSeeded(t *testing.T) {
	needles := []string{"alice", "Bob", "192.168", "8.8.4.4", "fe80", "jane.doe", "example-owner", "nas.lan", "abcdef0123456789", "hunter2", "sesame", "ZZtopSecret", "Abcdef0123456789xyz",
		"123e4567", "ABCD1-2EFGH", "default.key", "my.key", "Mychar", "MIIEvQ", "Xk3jD9sLq2Wv8Zr1Tn5Yb7Hc4Gf6Mp0QaUe", "0123456789abcdef0123456789abcdef0123"}
	var all []string
	for _, v := range seeded {
		all = append(all, v)
	}
	out := Redact(strings.Join(all, "\n"))
	for _, n := range needles {
		if strings.Contains(out, n) {
			t.Errorf("%q survived redaction:\n%s", n, out)
		}
	}
	if f := Scan("x", out); len(f) != 0 {
		t.Errorf("scan finds redacted text dirty: %+v", f)
	}
}

func TestRedactKeepsDiagnostics(t *testing.T) {
	s := "12:00:01 area=Blood Moor life=212/300 hit monsters=7 path=internal/watch/supportapi.go"
	if got := Redact(s); got != s {
		t.Errorf("harmless line changed: %q", got)
	}
}

func TestExtraLiterals(t *testing.T) {
	out := Redact("running on venusbox as kid at /srv/d2data", "venusbox", "kid", "/srv/d2data")
	if strings.Contains(out, "venusbox") || strings.Contains(out, "/srv/d2data") {
		t.Fatal(out)
	}
	if f := Scan("a", "still venusbox here", "venusbox"); len(f) != 1 || f[0].Class != "local" || strings.Contains(f[0].Context, "venusbox") {
		t.Fatalf("%+v", f)
	}
}

func TestScanFindsUnredacted(t *testing.T) {
	for k, v := range seeded {
		if f := Scan("f", v); len(f) == 0 {
			t.Errorf("scan missed %s", k)
		} else {
			for _, x := range f {
				if x.Class != "pem" && strings.Contains(x.Context, "hunter2") {
					t.Errorf("finding echoes a secret: %+v", x)
				}
			}
		}
	}
}

func inputs(files ...Input) Inputs { return Inputs{Files: files, Now: now} }

func unzip(t *testing.T, b []byte) map[string]string {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]string{}
	for _, f := range zr.File {
		r, _ := f.Open()
		d, _ := io.ReadAll(r)
		m[f.Name] = string(d)
	}
	return m
}

func TestBuildAllowlistAndRedaction(t *testing.T) {
	b, err := Build(inputs(
		Input{"log.txt", []byte("2026-11-15T11:59:00Z started from /home/alice/bot\n")},
		Input{"knobs.json", []byte(`{"combat.kite": 3, "watch_token": "abc", "nested": {"apiKey": "k", "ok": 1}}`)},
		Input{".env", []byte("D2BOT_SUPPORT_TOKEN=secret")},
		Input{"default.key", []byte("cdkey")},
		Input{"screenshot.png", []byte("\x89PNG\r\n\x1a\nxx")},
		Input{"save/Mychar.d2s", []byte{0x55, 0xAA, 0x55, 0xAA, 1}},
	), Options{})
	if err != nil || !b.Sendable() {
		t.Fatalf("%v %+v", err, b)
	}
	m := unzip(t, b.Zip)
	if len(m) != 3 || m["MANIFEST.txt"] == "" {
		t.Fatalf("files: %v", m)
	}
	if strings.Contains(m["log.txt"], "alice") || strings.Contains(m["knobs.json"], "abc") || strings.Contains(m["knobs.json"], "apiKey") || !strings.Contains(m["knobs.json"], "combat.kite") {
		t.Fatalf("%v", m)
	}
	if len(b.Dropped) != 2 {
		t.Fatalf("dropped: %v", b.Dropped)
	}
	for _, f := range b.Files {
		if strings.Contains(f.Name, "screenshot") || strings.Contains(f.Name, "save") {
			t.Fatalf("optional file in the default bundle: %v", b.Files)
		}
	}
}

func TestOptionalFiles(t *testing.T) {
	in := inputs(Input{"screenshot.png", []byte("\x89PNG\r\n\x1a\nxx")}, Input{"save/Mychar.d2s", []byte{0x55, 0xAA, 0x55, 0xAA, 1}}, Input{"save/bad.d2s", []byte("nope")})
	b, _ := Build(in, Options{IncludeScreenshot: true, IncludeSave: true})
	m := unzip(t, b.Zip)
	if _, ok := m["screenshot.png"]; !ok {
		t.Fatal(m)
	}
	if _, ok := m["save/character1.d2s"]; !ok || len(m) != 3 {
		t.Fatalf("%v", m) // the character name is not in the path; the bad one is dropped
	}
	if len(b.Dropped) != 1 {
		t.Fatalf("%v", b.Dropped)
	}
}

func TestSaveOutsideSaveDirBlocks(t *testing.T) {
	b, _ := Build(inputs(Input{"log.txt", []byte("hi")}), Options{})
	if !b.Sendable() {
		t.Fatal("baseline")
	}
	// a collector that hands a save in under an allowed name must block the whole bundle
	b, _ = Build(inputs(Input{"errors.txt", []byte{0x55, 0xAA, 0x55, 0xAA, 'x'}}), Options{})
	if b.Sendable() || len(b.Findings) == 0 || b.Findings[0].Class != "save" || len(b.Zip) != 0 {
		t.Fatalf("%+v", b)
	}
}

func TestLogHoursAndCap(t *testing.T) {
	old := now.Add(-5 * time.Hour).Format(time.RFC3339)
	fresh := now.Add(-5 * time.Minute).Format(time.RFC3339)
	in := inputs(Input{"log.txt", []byte(old + " old line\n" + fresh + " fresh line\ncontinuation\n")})
	in.LogHours = 1
	b, _ := Build(in, Options{})
	log := unzip(t, b.Zip)["log.txt"]
	if strings.Contains(log, "old line") || !strings.Contains(log, "fresh line") || !strings.Contains(log, "continuation") {
		t.Fatal(log)
	}
	big := strings.Repeat("line of ordinary diagnostics here\n", 100000)
	b, _ = Build(inputs(Input{"log.txt", []byte(big)}), Options{})
	if !b.Sendable() || b.Files[len(b.Files)-1].Size > MaxTextFile && b.Files[0].Size > MaxTextFile {
		t.Fatal("cap")
	}
	if got := unzip(t, b.Zip)["log.txt"]; len(got) > MaxTextFile {
		t.Fatalf("log not capped: %d", len(got))
	}
}

func TestSendableCap(t *testing.T) {
	b := &Bundle{Zip: make([]byte, MaxZip+1)}
	if b.Sendable() {
		t.Fatal("over the cap")
	}
	b.Zip = b.Zip[:10]
	b.Findings = []Finding{{Class: "x"}}
	if b.Sendable() {
		t.Fatal("a finding blocks")
	}
	if (*Bundle)(nil).Sendable() {
		t.Fatal("nil")
	}
}

func TestSenderAuthAndCodes(t *testing.T) {
	var gotAuth, gotCT string
	var gotLen int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotCT = r.Header.Get("Authorization"), r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotLen = len(b)
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(401)
			return
		}
		io.WriteString(w, `{"code":"ABCD2345"}`)
	}))
	defer srv.Close()
	code, err := Sender{URL: srv.URL, Token: "good"}.Send(context.Background(), []byte("PK"))
	if err != nil || code != "ABCD2345" || gotAuth != "Bearer good" || gotCT != "application/zip" || gotLen != 2 {
		t.Fatal(code, err, gotAuth, gotCT, gotLen)
	}
	_, err = Sender{URL: srv.URL, Token: "bad"}.Send(context.Background(), []byte("PK"))
	if err == nil || strings.Contains(err.Error(), "bad") || strings.Contains(err.Error(), srv.URL) {
		t.Fatal(err)
	}
	if _, err = (Sender{}).Send(context.Background(), []byte("PK")); err == nil {
		t.Fatal("unconfigured must refuse")
	}
	if _, err = (Sender{URL: "http://example.org/x", Token: "t"}).Send(context.Background(), []byte("PK")); err == nil {
		t.Fatal("plain http to a non-loopback host must refuse")
	}
	// a redirect is never followed (the token would go along)
	red := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "http://127.0.0.1:1/x", 302) }))
	defer red.Close()
	if _, err = (Sender{URL: red.URL, Token: "t"}).Send(context.Background(), []byte("PK")); err == nil {
		t.Fatal("redirect")
	}
}

// Every input of the security review (M1): Redact must remove the secret and Scan must see it when it is left in.
func TestReviewClasses(t *testing.T) {
	rows := []struct{ in, secret string }{
		{"passphrase: correct horse battery staple", "horse battery"},
		{"password = hunter two words", "two words"},
		{"pass=hunter2 pwd=hunter3 PW: hunter4", "hunter4"},
		{"key ABCD1234EFGH5678", "ABCD1234EFGH5678"},
		{`path="C:\\Users\\alice\\x"`, "alice"},
		{`Z:\home\alice\.wine\drive_c\game.exe`, "alice"},
		{`C:/users/alice/Saved Games/x`, "alice"},
		{`C:\Users\John Smith\AppData`, "Smith"},
		{"peer fe80::1 and 2001:db8::42", "db8"},
		{"http://alice-pc:8891/watch", "alice-pc"},
		{"myhouse.duckdns.org and nas.fritz.box", "myhouse"},
		{"build 0123456789abcdef0123456789abcdef01234567", "0123456789abcdef0123456789abcdef01234567"},
		{"tok AbCdEfGhIjKlMnOpQrStUvWxYz_-AbCd", "AbCdEfGhIjKlMnOpQrStUvWxYz"},
		{"tok AbCdEfGh1jKlMnOpQrStUvWxYz_-AbC", "AbCdEfGh1jKlMnOpQrStUvWxYz"},
		{"cd KJ8H2N4P6Q9RTVWXY3Z5ABCDEF", "KJ8H2N4P6Q9R"},
		{"cd KJ8H2N4P6Q9RTVWX", "KJ8H2N4P6Q9R"},
		{"/x?f=%2Fhome%2Falice%2Fsave", "alice"},
		{"user:pw1234567@host.example-owner.test", "pw1234567"},
	}
	for _, r := range rows {
		out := Redact(r.in)
		if strings.Contains(out, r.secret) {
			t.Errorf("not redacted %q -> %q", r.in, out)
		}
		if f := Scan("f", r.in); len(f) == 0 {
			t.Errorf("scan did not see %q", r.in)
		}
		if f := Scan("f", out); len(f) != 0 {
			t.Errorf("redacted %q still scans dirty: %+v", out, f)
		}
	}
}

// JSON escapes & as \u0026 and doubles backslashes; neither may hide a shape.
func TestJSONShapes(t *testing.T) {
	b, _ := Build(inputs(Input{"status.json", []byte(`{"link":"http://h.example-owner.test:8891/watch?x=1&t=abcd1234efgh","path":"C:\\Users\\alice\\AppData\\x","pp":"passphrase: a b c d"}`)}), Options{})
	if !b.Sendable() {
		t.Fatalf("%+v", b.Findings)
	}
	s := unzip(t, b.Zip)["status.json"]
	for _, bad := range []string{"abcd1234efgh", "alice", "example-owner", "a b c d", `\u0026`} {
		if strings.Contains(s, bad) {
			t.Errorf("%q in %s", bad, s)
		}
	}
}

// What Redact cannot see by shape, the known literals remove, and the scan blocks when one is left.
func TestKnownSecretsAsLiterals(t *testing.T) {
	tok := "plainlowercasetokenwithnodigitsatall" // no shape rule fires on it
	in := Inputs{Now: now, Extra: []string{tok, "/srv/home/kid", `C:\Users\kid`}, Files: []Input{{"log.txt", []byte("auth ok " + tok + " in /srv/home/kid and C:\\\\Users\\\\kid\\\\x\n")}}}
	b, _ := Build(in, Options{})
	log := unzip(t, b.Zip)["log.txt"]
	if !b.Sendable() || strings.Contains(log, tok) || strings.Contains(log, "kid") {
		t.Fatalf("%v %q", b.Findings, log)
	}
	if f := Scan("x", "leak "+tok, tok); len(f) != 1 || f[0].Class != "local" {
		t.Fatalf("%+v", f)
	}
}

func TestPreviewCoversEveryFile(t *testing.T) {
	big := strings.Repeat("a line of ordinary text\n", 2000)
	b, _ := Build(inputs(Input{"log.txt", []byte(big + "pass=hunter9\n")}, Input{"errors.txt", []byte("ERROR small\n")}), Options{})
	for _, f := range b.Files {
		if f.Name != "MANIFEST.txt" && (f.Excerpt == "" || f.Lines == 0) {
			t.Errorf("no excerpt for %s", f.Name)
		}
		if f.Name == "log.txt" && (f.Redactions != 1 || !strings.Contains(f.Excerpt, "not shown")) {
			t.Errorf("log: %+v", f)
		}
		if f.Name == "errors.txt" && f.Excerpt != "ERROR small\n" {
			t.Errorf("small file is shown whole: %q", f.Excerpt)
		}
	}
	var all strings.Builder
	for p := 0; ; p++ {
		s, more, ok := b.Page("log.txt", p)
		if !ok {
			t.Fatal("page", p)
		}
		all.WriteString(s)
		if !more {
			break
		}
	}
	if all.String() != unzip(t, b.Zip)["log.txt"] {
		t.Error("the pages are not the file that is in the zip")
	}
	if _, _, ok := b.Page("nope", 0); ok {
		t.Error("unknown file")
	}
}
