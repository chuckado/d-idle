package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"d2lod-bot/internal/d2dsoakstat"
)

// cmdBands prints the clvl-band split of minute logs and, with -baselines, the per-band XP/h verdict of the Gate 1
// clause (spec soak-gates "Gate 1 deaths and progression"). Exit 0 unless a band is FAIL (exit 3, like compare);
// ADVISORY bands never fail.
func cmdBands(args []string) error {
	fs := flag.NewFlagSet("bands", flag.ExitOnError)
	spec := fs.String("bands", d2dsoakstat.DefaultBands, "clvl bands LO-HI,...")
	hours := fs.Float64("hours", 0, "fixed run length of the log (d2d virtual hours); 0 = play time from the timestamps")
	bl := fs.String("baselines", "", "baseline file(s), comma separated; their band tables are pooled")
	asJSON := fs.Bool("json", false, "print JSON")
	exitZero := fs.Bool("exit-zero", false, "exit 0 whatever the verdict")
	fs.Parse(args)
	bands, err := d2dsoakstat.ParseBands(*spec)
	if err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("no input")
	}
	var rows []d2dsoakstat.BaselineRun
	if *bl != "" {
		for _, p := range strings.Split(*bl, ",") {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			b, err := d2dsoakstat.ReadBaselines(f)
			f.Close()
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			rows = append(rows, b.Runs...)
		}
	}
	type bandOut struct {
		d2dsoakstat.BandStat
		Base    *d2dsoakstat.BandBase `json:"baseline,omitempty"`
		Verdict string                `json:"verdict,omitempty"`
		Tol     float64               `json:"tol,omitempty"`
		Gap     float64               `json:"gap,omitempty"`
	}
	type runOut struct {
		Run   string    `json:"run"`
		Bands []bandOut `json:"bands"`
	}
	var outs []runOut
	failed := false
	for _, p := range fs.Args() {
		ms, name, err := loadChained(strings.Split(p, "+"))
		if err != nil {
			return err
		}
		ro := runOut{Run: name}
		for _, s := range d2dsoakstat.Bands(ms, bands, *hours) {
			bo := bandOut{BandStat: s}
			if len(rows) > 0 {
				bb := d2dsoakstat.BandBaseline(rows, s.Band)
				bo.Base = &bb
				bo.Verdict, bo.Tol, bo.Gap = d2dsoakstat.BandJudge(s.XPPerHour, s.Hours, bb, 1)
				failed = failed || bo.Verdict == d2dsoakstat.BandBelow
			}
			ro.Bands = append(ro.Bands, bo)
		}
		outs = append(outs, ro)
	}
	if *asJSON {
		if err := json.NewEncoder(os.Stdout).Encode(outs); err != nil {
			return err
		}
	} else {
		for _, ro := range outs {
			fmt.Printf("%s\n%-7s %6s %5s %8s %7s %6s %6s", ro.Run, "band", "hours", "min", "xp/h", "kills/h", "deaths", "open")
			if len(rows) > 0 {
				fmt.Printf(" | %2s %6s %8s %6s %6s %7s %7s  %s", "n", "hours", "base xp/h", "cv", "tol", "gap", "", "verdict")
			}
			fmt.Println()
			for _, b := range ro.Bands {
				fmt.Printf("%-7s %6.2f %5d %8.0f %7.0f %6d %6v", b.Band, b.Hours, b.Minutes, b.XPPerHour, b.KillsPerH, b.Deaths, b.Open)
				if b.Base != nil {
					fmt.Printf(" | %2d %6.2f %9.0f %5.0f%% %5.0f%% %+6.0f%% %7s  %s", b.Base.N, b.Base.Hours, b.Base.Mean, 100*b.Base.CV, 100*b.Tol, 100*b.Gap, "", b.Verdict)
				}
				fmt.Println()
			}
		}
	}
	if failed && !*exitZero {
		return exitError{Code: d2dsoakstat.ExitCode(d2dsoakstat.Fail)}
	}
	return nil
}
