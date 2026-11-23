package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"d2lod-bot/internal/scenario"
)

// splitDetArgs takes -n and -out out of the argument list and returns the rest unchanged, to be passed to run.
func splitDetArgs(args []string) (n int, out string, rest []string, err error) {
	n, out = 2, "provingground-out/det"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--n", "-out", "--out":
			if i+1 >= len(args) {
				return 0, "", nil, fmt.Errorf("%s needs a value", args[i])
			}
			if args[i][len(args[i])-1] == 'n' {
				if n, err = strconv.Atoi(args[i+1]); err != nil || n < 2 {
					return 0, "", nil, fmt.Errorf("-n wants a number of at least 2")
				}
			} else {
				out = args[i+1]
			}
			i++
		default:
			rest = append(rest, args[i])
		}
	}
	return
}

// cmdDeterminismRun runs the selected scenarios N times (run -journal, one directory per repetition) and compares every
// repetition with the first: verdict and assertion digest are the hard rule, the state digest is drift, and when a
// d2ddet binary sits next to this one the journals of the first differing pair are compared too (the journal says which
// layer moved first: a bot command at equal state, or a game reply under equal commands; docs/d2d-determinism.md).
// Runs are sequential on purpose: the lab slots are the scarce resource. It writes <out>/determinism.txt as well.
func cmdDeterminismRun(args []string, w io.Writer) error {
	n, out, rest, err := splitDetArgs(args)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	rep, err := os.Create(filepath.Join(out, "determinism.txt"))
	if err != nil {
		return err
	}
	defer rep.Close()
	w = io.MultiWriter(w, rep)
	var sets [][]scenario.Result
	for i := 1; i <= n; i++ {
		d := filepath.Join(out, fmt.Sprintf("rep%d", i))
		ra := append([]string{"-journal", "-out", d}, rest...)
		fmt.Fprintf(w, "repetition %d of %d -> %s\n", i, n, d)
		// A red or blocked run is a result, not an infrastructure failure here: the comparison is what we want.
		if err := cmdRun(ra, io.Discard); err != nil {
			if _, ok := err.(exitErr); !ok {
				return err
			}
		}
		rs, err := readResults(filepath.Join(d, "results.json"))
		if err != nil {
			return err
		}
		sets = append(sets, rs)
	}
	bad := false
	for i := 1; i < n; i++ {
		ok, vd, dd, missing := scenario.CheckDeterminism(sets[0], sets[i])
		fmt.Fprintf(w, "\nrepetition 1 against %d\n", i+1)
		for _, d := range vd {
			fmt.Fprintf(w, "  DIFFERS (hard) %s: %s vs %s (%s)\n", d.Key, d.A, d.B, d.Detail)
		}
		for _, d := range dd {
			fmt.Fprintf(w, "  drift (state digest, not a failure) %s: %s vs %s\n", d.Key, d.A, d.B)
		}
		for _, m := range missing {
			fmt.Fprintf(w, "  missing in one repetition: %s\n", m)
		}
		if ok && len(dd) == 0 {
			fmt.Fprintln(w, "  identical verdicts, assertion digests and state digests")
		}
		bad = bad || !ok
		compareJournals(w, sets[0], sets[i])
	}
	if bad {
		return exitErr{3, "NOT DETERMINISTIC (a flaky scenario is a failed scenario)"}
	}
	return nil
}

// compareJournals runs `d2ddet compare -ignore state` on the journals of the runs the two repetitions share, when the
// binary is next to this one. A missing binary or journal is said, never skipped silently.
func compareJournals(w io.Writer, a, b []scenario.Result) {
	self, _ := os.Executable()
	bin := filepath.Join(filepath.Dir(self), "d2ddet")
	if _, err := os.Stat(bin); err != nil {
		fmt.Fprintf(w, "  journals not compared: no d2ddet next to %s (go build -o <dir>/d2ddet ./cmd/d2ddet)\n", filepath.Base(self))
		return
	}
	bm := map[string]scenario.Result{}
	for _, r := range b {
		bm[r.Key()] = r
	}
	for _, ra := range a {
		rb, ok := bm[ra.Key()]
		if !ok || ra.Journal == "" || rb.Journal == "" {
			fmt.Fprintf(w, "  journal of %s: not written in both repetitions (run -journal; BLOCKED runs have none)\n", ra.Key())
			continue
		}
		o, err := exec.Command(bin, "compare", "-ignore", "state", ra.Journal, rb.Journal).CombinedOutput()
		fmt.Fprintf(w, "  d2ddet compare %s (err: %v)\n%s", ra.Key(), err, indent(string(o)))
	}
}

func indent(s string) string {
	var b strings.Builder
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		b.WriteString("    " + l + "\n")
	}
	return b.String()
}
