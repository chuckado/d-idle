// Command d2arbiter replays a bot log (JSON lines) through the intent arbiter offline and prints what it would have
// done (docs/arbiter.md). Usage: d2arbiter [-follow] [-trace] LOGFILE
package main

import (
	"flag"
	"fmt"
	"os"

	"d2lod-bot/internal/arbiter"
)

func main() {
	follow := flag.Bool("follow", false, "the simulated hero obeys the arbiter instead of the log")
	trace := flag.Bool("trace", false, "print the full decision trace")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: d2arbiter [-follow] [-trace] LOGFILE")
		os.Exit(2)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	res, err := arbiter.Replay(f, arbiter.DefaultConfig(), arbiter.ShadowConfig{LogGapSec: 60}, *follow)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("ticks %d, logged area changes %d, arbiter area changes (follow) %d, travel refusals %d\n", res.Ticks, res.AreaChanges, res.FollowMoves, res.Refused)
	fmt.Println("per family diverged/ticks:", res.FamilyCounts)
	for _, l := range res.Divergences {
		fmt.Println(l)
	}
	if *trace {
		for _, l := range res.Trace {
			fmt.Println(l)
		}
	}
}
