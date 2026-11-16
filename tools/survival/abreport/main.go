package main

import (
	"flag"
	"fmt"
	"os"
)

// usage: abreport [-base legacy] [-arm margin] [-horizon 120] [-xp-floor 1.0] [-min-seeds 3] DIR
// Exit 0 PASS, 1 FAIL, 3 INCONCLUSIVE, 2 bad input.
func main() {
	base := flag.String("base", "legacy", "base arm directory name under DIR")
	arm := flag.String("arm", "margin", "test arm directory name under DIR")
	horizon := flag.Int("horizon", 120, "virtual minutes each run must cover")
	floor := flag.Float64("xp-floor", 1.0, "share of the base arm's XP at the horizon the test arm must reach")
	minSeeds := flag.Int("min-seeds", 3, "pairs needed for PASS (the gate says 3; docs/ab-design.md says 8 for a claimed effect)")
	dslack := flag.Int("death-slack", 0, "pooled extra deaths the test arm may have")
	cslack := flag.Int("chicken-slack", 0, "pooled extra chickens the test arm may have")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: abreport [flags] DIR   (DIR/<arm>/<seed>.log)")
		os.Exit(2)
	}
	dir := flag.Arg(0)
	b, nb, err := Load(dir, *base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "abreport:", err)
		os.Exit(2)
	}
	t, nt, err := Load(dir, *arm)
	if err != nil {
		fmt.Fprintln(os.Stderr, "abreport:", err)
		os.Exit(2)
	}
	if len(b) == 0 || len(t) == 0 {
		fmt.Fprintf(os.Stderr, "abreport: no usable logs (base %d, test %d)\n", len(b), len(t))
		os.Exit(2)
	}
	for _, n := range append(nb, nt...) {
		fmt.Println("note:", n)
	}
	s := Compare(b, t, Options{Horizon: *horizon, XPFloor: *floor, MinSeeds: *minSeeds, DeathSlack: *dslack, ChickenSlack: *cslack})
	fmt.Print(s.Format())
	switch s.Verdict {
	case "PASS":
	case "FAIL":
		os.Exit(1)
	default:
		os.Exit(3)
	}
}
