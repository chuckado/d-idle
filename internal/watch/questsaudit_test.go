package watch

import (
	"bytes"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/support"
)

func TestRemoteClassAndAgent(t *testing.T) {
	for addr, want := range map[string]string{"127.0.0.1:5": "loopback", "[::1]:5": "loopback", "192.168.1.9:80": "lan", "10.0.0.2:1": "lan", "8.8.8.8:1": "other", "junk": "other"} {
		if got := remoteClass(addr); got != want {
			t.Errorf("%s: %s, want %s", addr, got, want)
		}
	}
	for ua, want := range map[string]string{"Mozilla/5.0 (X11) Gecko/20100101 Firefox/130.0": "firefox", "curl/8.1": "curl", "": "", "Mozilla/5.0 Chrome/120 Safari/537": "chrome", "weird": "other"} {
		if got := agentFamily(ua); got != want {
			t.Errorf("%q: %s, want %s", ua, got, want)
		}
	}
	if safeID("den_of_evil") != "den_of_evil" || safeID("a b\n") != "?" || safeID(strings.Repeat("a", 65)) != "?" {
		t.Error("safeID")
	}
}

func (r *questRig) doFrom(path, body, addr, ua string) {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.RemoteAddr = addr
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Origin", "http://ok.test")
	r.post.ServeHTTP(httptest.NewRecorder(), req)
}

func (r *questRig) audit() []questboard.AuditEvent {
	_, st := r.desk.ExecForTest()
	return questboard.AuditOf(st).Recent(0)
}

// A quest accepted and then cancelled over the API leaves an audit record that names the source class (the 2026-11-18
// mystery: a cancel with no trace). Must fail without the Record call in PostHandler.
func TestCancelLeavesAuditRecordWithSource(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	r := newQuestRig(t, "Paladin", 6)
	r.doFrom("/api/quests/den_of_evil/start", `{}`, "192.168.1.20:5555", "Mozilla/5.0 Firefox/130.0")
	r.tick(1)
	r.doFrom("/api/quests/den_of_evil/cancel", `{"confirm":true}`, "192.168.1.20:5555", "Mozilla/5.0 Firefox/130.0")
	ev := r.audit()
	if len(ev) != 2 {
		t.Fatalf("events: %+v", ev)
	}
	c := ev[0]
	if c.Action != "cancel" || c.Quest != "den_of_evil" || !c.OK || c.Already || c.Remote != "lan" || c.Agent != "firefox" || !c.Confirm || c.Path != "/api/quests/den_of_evil/cancel" || c.Status != 200 {
		t.Fatalf("cancel record: %+v", c)
	}
	if ev[1].Action != "start" || ev[1].Confirm {
		t.Fatalf("start record: %+v", ev[1])
	}
	out := buf.String()
	if !strings.Contains(out, "quests: action=cancel quest=den_of_evil") || !strings.Contains(out, "remote=lan agent=firefox confirm=true ok=true") {
		t.Fatalf("log: %s", out)
	}
	for _, bad := range []string{"tok", "192.168", "5555", "Mozilla", "Bearer"} {
		if strings.Contains(out, bad) {
			t.Errorf("log leaks %q: %s", bad, out)
		}
	}
	raw, _ := os.ReadFile(r.auditPath())
	for _, bad := range []string{"tok", "192.168", "Mozilla", "Bearer"} {
		if strings.Contains(string(raw), bad) {
			t.Errorf("audit file leaks %q", bad)
		}
	}
}

func (r *questRig) auditPath() string {
	_, st := r.desk.ExecForTest()
	return questboard.AuditPathOf(st)
}

// The board shows a stop for 24 hours and a repeated cancel (nothing to stop) is not a stop.
func TestStoppedNoteInView(t *testing.T) {
	r := newQuestRig(t, "Paladin", 6)
	r.doFrom("/api/quests/den_of_evil/start", `{}`, "127.0.0.1:1", "curl/8")
	r.tick(1)
	if strings.Contains(r.get(), `"stopped"`) {
		t.Fatal("stopped before any stop")
	}
	r.doFrom("/api/quests/den_of_evil/cancel", `{}`, "127.0.0.1:1", "curl/8")
	g := r.get()
	if !strings.Contains(g, `"stopped":[{"quest":"den_of_evil","questTitle":"Den of Evil"`) {
		t.Fatalf("no stopped row: %s", g)
	}
	r.doFrom("/api/quests/den_of_evil/cancel", `{}`, "127.0.0.1:1", "curl/8")
	if n := len(r.audit()); n != 3 {
		t.Fatalf("audit %d", n)
	}
	r.mu.Lock()
	r.now = r.now.Add(25 * 3600e9)
	r.mu.Unlock()
	if strings.Contains(r.get(), `"stopped"`) {
		t.Fatal("stop still shown after a day")
	}
}

// The audit file name is not in the support bundle allowlist, so a bundle can never carry it.
func TestAuditFileIsNotBundled(t *testing.T) {
	for _, n := range []string{"Testy.audit.jsonl", "quests/Testy.audit.jsonl", "audit.jsonl"} {
		if support.Allowed[n] {
			t.Errorf("%s is bundled", n)
		}
	}
}

// A refused call (bad token) is logged but never reaches the ring.
func TestRefusedCallNotAudited(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	r := newQuestRig(t, "Paladin", 6)
	req := httptest.NewRequest("POST", "/api/quests/den_of_evil/cancel", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong")
	r.post.ServeHTTP(httptest.NewRecorder(), req)
	if len(r.audit()) != 0 {
		t.Fatal("refused call in the ring")
	}
	if !strings.Contains(buf.String(), "quests: action=cancel") || strings.Contains(buf.String(), "wrong") {
		t.Fatalf("log: %s", buf.String())
	}
}
