package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCutReason(t *testing.T) {
	d := t.TempDir()
	cut := filepath.Join(d, "cut.log")
	ok := filepath.Join(d, "ok.log")
	_ = os.WriteFile(cut, []byte(`{"msg":"zone entry"}`+"\n"), 0o644)
	_ = os.WriteFile(ok, []byte(`{"msg":"virtual run length reached"}`+"\n"), 0o644)
	if cutReason(cut) == "" {
		t.Fatal("a log without the run-length line must be a cut run")
	}
	if cutReason(ok) != "" {
		t.Fatal("a log with the run-length line is a full run")
	}
}

func TestCollectProvHostRole(t *testing.T) {
	t.Setenv("D2D_HOST_ROLE", "builder")
	t.Setenv("D2LOD_BOT_COMMIT", "abc1234")
	p := collectProv(filepath.Join(t.TempDir(), "nobot"))
	if p.Host != "builder" || p.BotCommit != "abc1234" {
		t.Fatalf("got %+v", p)
	}
}
