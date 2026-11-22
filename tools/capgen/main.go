// Command capgen generates the views of the capability table (data/capabilities/capabilities.json) that other
// packages embed: internal/combat/classdata/skills.json and the skillKinds block of internal/power/profiles.json.
// With -check it only compares and exits 1 on a difference (CI); with -tiers it prints the computed tier of every
// class skill. Run from the repository root.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"

	"d2lod-bot/internal/delivery"
)

const (
	capsPath    = "data/capabilities/capabilities.json"
	provenPath  = "data/capabilities/proven.json"
	skillsPath  = "internal/combat/classdata/skills.json"
	profilePath = "internal/power/profiles.json"
	backend     = "d2d"
)

func main() {
	check := flag.Bool("check", false, "compare only; exit 1 when a generated file differs")
	tiers := flag.Bool("tiers", false, "print the computed tier of every class skill and exit")
	flag.Parse()
	if err := run(*check, *tiers); err != nil {
		fmt.Fprintln(os.Stderr, "capgen:", err)
		os.Exit(1)
	}
}

func run(check, tiers bool) error {
	cb, err := os.ReadFile(capsPath)
	if err != nil {
		return err
	}
	t, err := delivery.Load(cb)
	if err != nil {
		return err
	}
	if tiers {
		pb, err := os.ReadFile(provenPath)
		if err != nil {
			return err
		}
		proven, err := delivery.ProvenSkills(pb, backend)
		if err != nil {
			return err
		}
		tt := t.Tiers(proven)
		names := make([]string, 0, len(tt))
		for n := range tt {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Printf("%s\t%s\t%s\n", t.Skills[n].Class, n, tt[n])
		}
		return nil
	}
	skills, err := delivery.SkillsJSON(t)
	if err != nil {
		return err
	}
	old, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}
	prof, err := delivery.ProfilesJSON(old, t)
	if err != nil {
		return err
	}
	bad := false
	for _, f := range []struct {
		path string
		want []byte
	}{{skillsPath, skills}, {profilePath, prof}} {
		have, _ := os.ReadFile(f.path)
		if bytes.Equal(have, f.want) {
			continue
		}
		if check {
			fmt.Fprintf(os.Stderr, "capgen: %s differs from the table; run `go run ./tools/capgen`\n", f.path)
			bad = true
			continue
		}
		if err := os.WriteFile(f.path, f.want, 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", f.path)
	}
	if bad {
		return fmt.Errorf("generated files are stale")
	}
	return nil
}
