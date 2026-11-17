package buildinfo

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The dev update ships a tree to the worker host, where a copied tree has no .git and a stale mirror's .git names the WRONG commit
// (the page once showed the coordinator checkout's old commit, "a756ec2-dirty"). So tools/dev-update.sh computes the build id of
// the tree it ships and stamps it with -X internal/buildinfo.Version through dev-update-host.sh --stamp.
func TestDevUpdateStampsTheShippedCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	script, err := os.ReadFile("../../tools/dev-update.sh")
	if err != nil {
		t.Fatal(err)
	}
	s := string(script)
	i := strings.Index(s, "stamp_of() {")
	j := strings.Index(s[i:], "\n}\n")
	if i < 0 || j < 0 {
		t.Fatal("stamp_of not found in tools/dev-update.sh")
	}
	fn := s[i : i+j+3]
	dir := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q")
	os.WriteFile(dir+"/a.txt", []byte("one\n"), 0o644)
	run("add", "a.txt")
	run("commit", "-qm", "one")
	first := run("rev-parse", "--short=7", "HEAD")
	os.WriteFile(dir+"/a.txt", []byte("two\n"), 0o644)
	run("commit", "-qam", "two")
	second := run("rev-parse", "--short=7", "HEAD")
	stamp := func(args ...string) string {
		out, err := exec.Command("bash", append([]string{"-c", fn + "\nstamp_of \"$@\"", "x"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("stamp_of %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	if got := stamp(dir); got != second {
		t.Fatalf("clean tree: %q, want %q", got, second)
	}
	if got := stamp(dir, first); got != first {
		t.Fatalf("a named commit: %q, want %q (no -dirty for a ref)", got, first)
	}
	os.WriteFile(dir+"/a.txt", []byte("three\n"), 0o644)
	if got := stamp(dir); got != second+"-dirty" {
		t.Fatalf("modified tracked file: %q, want %q", got, second+"-dirty")
	}
	os.WriteFile(dir+"/new.txt", []byte("untracked\n"), 0o644)
	run("checkout", "a.txt")
	if got := stamp(dir); got != second {
		t.Fatalf("an untracked file must not make it dirty: %q", got)
	}
	// the host half passes the stamp on to the linker, and the image build can too
	host, _ := os.ReadFile("../../tools/dev-update-host.sh")
	if !strings.Contains(string(host), "--stamp) stamp=$2") || !strings.Contains(string(host), "-X d2lod-bot/internal/buildinfo.Version=$stamp") {
		t.Fatal("tools/dev-update-host.sh does not stamp buildinfo.Version from --stamp")
	}
	if !strings.Contains(s, `host+=(--stamp "$stamp")`) {
		t.Fatal("tools/dev-update.sh does not pass --stamp to the host")
	}
	ctx, _ := os.ReadFile("../../deploy/d2d-container/context.sh")
	if !strings.Contains(string(ctx), "-X d2lod-bot/internal/buildinfo.Version=$BUILD_STAMP") {
		t.Fatal("context.sh cannot stamp the build id")
	}
}
