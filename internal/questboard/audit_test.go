package questboard

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuditRingIsBounded(t *testing.T) {
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err := Open(b, t.TempDir(), "Testy")
	if err != nil {
		t.Fatal(err)
	}
	a := AuditOf(st)
	if AuditOf(st) != a {
		t.Fatal("one ring per file")
	}
	for i := 0; i < AuditMax+37; i++ {
		if err := a.Append(AuditEvent{At: time.Unix(int64(i), 0).UTC(), Action: "start", Quest: fmt.Sprintf("q%d", i), Path: "/p", Remote: "lan", OK: true}); err != nil {
			t.Fatal(err)
		}
	}
	raw, _ := os.ReadFile(a.path)
	if n := strings.Count(string(raw), "\n"); n != AuditMax {
		t.Fatalf("lines: %d, want %d", n, AuditMax)
	}
	r := a.Recent(3)
	if len(r) != 3 || r[0].Quest != fmt.Sprintf("q%d", AuditMax+36) || r[2].Quest != fmt.Sprintf("q%d", AuditMax+34) {
		t.Fatalf("newest first: %+v", r)
	}
	if got := a.Recent(0); len(got) != AuditMax || got[AuditMax-1].Quest != "q37" {
		t.Fatalf("oldest kept: %d", len(got))
	}
	if !strings.HasSuffix(a.path, "Testy.audit.jsonl") {
		t.Fatal(a.path)
	}
}

func TestAuditNilStore(t *testing.T) {
	if AuditOf(nil) != nil {
		t.Fatal("no store, no ring")
	}
}
