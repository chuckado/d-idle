package main

import (
	"strings"
	"testing"
)

func TestSplitDetArgs(t *testing.T) {
	n, out, rest, err := splitDetArgs([]string{"-id", "town.resupply", "-n", "3", "-out", "o/x", "-class", "Paladin"})
	if err != nil || n != 3 || out != "o/x" || strings.Join(rest, " ") != "-id town.resupply -class Paladin" {
		t.Fatalf("%d %q %v %v", n, out, rest, err)
	}
	if _, _, _, err := splitDetArgs([]string{"-n", "1"}); err == nil {
		t.Fatal("a determinism run of one repetition compares nothing")
	}
	if _, _, _, err := splitDetArgs([]string{"-n"}); err == nil {
		t.Fatal("a flag without a value")
	}
}

func TestRunJournalFlagReachesExecOne(t *testing.T) {
	var sb strings.Builder
	if err := cmdRun([]string{"-dir", dir, "-dry-run", "-journal", "-id", "town.resupply", "-class", "Paladin", "-seed", "11", "-out", t.TempDir()}, &sb); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "exec-one") || !strings.Contains(sb.String(), "-journal") {
		t.Fatalf("dry run: %s", sb.String())
	}
	sb.Reset()
	if err := cmdRun([]string{"-dir", dir, "-dry-run", "-id", "town.resupply", "-class", "Paladin", "-seed", "11", "-out", t.TempDir()}, &sb); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sb.String(), "-journal") {
		t.Fatal("no journal unless asked")
	}
}
