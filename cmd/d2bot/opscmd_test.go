package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/ops"
	"d2lod-bot/internal/questboard"
)

type cliDesk struct{ calls []string }

func (c *cliDesk) View() questboard.View {
	return questboard.View{Enabled: true, Difficulty: "normal", Quests: []questboard.QuestRow{{ID: "den",
		Chunks: []questboard.ChunkRow{{ID: "c1", State: questboard.Done}}}}}
}
func (c *cliDesk) rep(s string) questboard.Reply {
	c.calls = append(c.calls, s)
	return questboard.Reply{OK: true, State: "accepted"}
}
func (c *cliDesk) Start(q, ch string) questboard.Reply  { return c.rep("start " + q) }
func (c *cliDesk) HandIn(q, ch string) questboard.Reply { return c.rep("handin " + q) }
func (c *cliDesk) Cancel(q, ch string) questboard.Reply { return c.rep("cancel " + q) }
func (c *cliDesk) Retry(q, ch string) questboard.Reply  { return c.rep("retry " + q) }
func (c *cliDesk) SetAuto(q string, on bool) questboard.Reply {
	return c.rep("auto " + q)
}
func (c *cliDesk) Record(questboard.AuditEvent) {}

func TestOpsCLI(t *testing.T) {
	dir := t.TempDir()
	d := &cliDesk{}
	env := map[string]string{"D2BOT_OPERATOR": "1"}
	stop, err := ops.Start(dir, func() ops.Desk { return d }, func(k string) string { return env[k] }, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	sock := filepath.Join(dir, ops.SockName)
	run := func(args ...string) (int, string) {
		var b bytes.Buffer
		code, err := opsMain(append(args, "--sock", sock), &b)
		if err != nil {
			b.WriteString(err.Error())
		}
		return code, b.String()
	}
	if c, o := run("quest", "retry", "den", "--why", "overnight"); c != 0 || !strings.Contains(o, "ok=true") {
		t.Fatalf("%d %s", c, o)
	}
	if c, o := run("quest", "retry", "den", "--why", "again"); c != 1 || !strings.Contains(o, "code=cooldown") {
		t.Fatalf("%d %s", c, o)
	}
	if c, o := run("quest", "start", "den"); c != 1 || !strings.Contains(o, "code=why") {
		t.Fatalf("no why: %d %s", c, o)
	}
	if c, o := run("quest", "sell", "den", "--why", "x"); c != 2 {
		t.Fatalf("sell: %d %s", c, o)
	}
	if c, _ := run("quest", "auto", "den", "--why", "x"); c != 2 {
		t.Fatalf("auto without on/off: %d", c)
	}
	if c, o := run("quest", "auto", "den", "--on", "--why", "x"); c != 0 {
		t.Fatalf("%d %s", c, o)
	}
	if c, o := run("quest", "status"); c != 0 || !strings.Contains(o, "den/c1\tdone") {
		t.Fatalf("%d %s", c, o)
	}
	if c, o := run("quest", "wait", "den", "--timeout", "1"); c != 0 || !strings.Contains(o, "RESULT outcome=finished") {
		t.Fatalf("%d %s", c, o)
	}
}

func TestStartOperatorOffByDefault(t *testing.T) {
	t.Setenv("D2BOT_OPERATOR", "")
	dir := t.TempDir()
	startOperator(dir, questboard.NewDesk(nil, questboard.DeskConfig{}), func(string, ...any) {})
	if matches, _ := filepath.Glob(filepath.Join(dir, "operator.sock*")); len(matches) != 0 {
		t.Fatalf("files without the env: %v", matches)
	}
}
