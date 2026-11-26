package main

import (
	"fmt"
	"io"
	"slices"

	"d2lod-bot/internal/arbiter"
)

const defaultOutcomes = "tools/provingground/arbiter-outcomes"

// cmdArbiterReplay scores the arbiter against the chain by OUTCOME (docs/arbiter.md, exit rule by outcome):
//
//	d2scenario arbiter-replay [-strict] FILE.json...
//
// Each file holds decision points (an incident pre-state or a recorded divergence) and the data rules that say which
// invariant a choice would hit. Hard and soft come from invariants.json. Report-only tier: the exit code is 0 unless
// -strict is given and a family's verdict fails.
func cmdArbiterReplay(args []string, w io.Writer) error {
	fs := newFlags("arbiter-replay")
	strict := fs.Bool("strict", false, "exit 1 when a family has cases and does not pass")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("arbiter-replay: give at least one outcome file (see %s)", defaultOutcomes)
	}
	pol := loadInvariantPolicy()
	hard := func(n string) bool { return slices.Contains(pol.Hard, n) }
	var all []arbiter.CaseScore
	for _, p := range fs.Args() {
		f, err := arbiter.LoadOutcomes(p)
		if err != nil {
			return err
		}
		for _, c := range f.Cases {
			s := f.Score(c, arbiter.DefaultConfig())
			all = append(all, s)
			fmt.Fprintf(w, "%-40s chain=%-9s hits=%v | arbiter=%-9s hits=%v | %s\n", s.ID, s.Chain, s.ChainHits, s.Arbiter, s.ArbiterHits, s.ArbiterIntent)
		}
	}
	failed := false
	fmt.Fprintln(w, "verdicts (hard, soft hits chain vs arbiter):")
	for _, v := range arbiter.Judge(all, hard) {
		fmt.Fprintf(w, "  %-9s cases=%d hard %d vs %d soft %d vs %d pass=%v (%s)\n", v.Family, v.Cases, v.ChainHard, v.ArbHard, v.ChainSoft, v.ArbSoft, v.Pass, v.Note)
		failed = failed || !v.Pass
	}
	if *strict && failed {
		return exitErr{code: 1, msg: "arbiter-replay: a family does not pass"}
	}
	return nil
}
