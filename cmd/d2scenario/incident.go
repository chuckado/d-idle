package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"d2lod-bot/internal/scenario"
)

// cmdIncident turns incident files (the state the bot recorded at a death, internal/bot/incident.go) into scenario rows:
//
//	d2scenario incident [-class C] [-seed N] [-out DIR] FILE.json|DIR...
//
// With -out DIR the rows are written as DIR/incident-<id>.json (a scenario file), ready for `d2scenario run -dir DIR`; without
// it the scenario file is printed. Report-only tier: nothing here touches the default scenario directory or the merge gate.
func cmdIncident(args []string, w io.Writer) error {
	fs := newFlags("incident")
	class := fs.String("class", "", "class of the replay hero (default: the class recorded in the incident)")
	seed := fs.Int64("seed", 1, "map seed")
	out := fs.String("out", "", "directory to write one scenario file per incident (default: print)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("incident: give at least one incident file")
	}
	files, err := incidentFiles(fs.Args())
	if err != nil {
		return err
	}
	var all []scenario.Scenario
	for _, p := range files {
		in, err := scenario.LoadIncident(p)
		if err != nil {
			return err
		}
		s, err := in.Scenario(*class, *seed)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		all = append(all, s)
	}
	if *out == "" {
		return writeScenarioFile(w, all)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	for _, s := range all {
		path := filepath.Join(*out, s.ID+".json")
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		err = writeScenarioFile(f, []scenario.Scenario{s})
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
		fmt.Fprintln(w, path)
	}
	return nil
}

// incidentFiles expands a directory argument (the bot's incidents directory) to its *.json files, oldest first.
func incidentFiles(args []string) ([]string, error) {
	var out []string
	for _, a := range args {
		if st, err := os.Stat(a); err == nil && st.IsDir() {
			names, err := filepath.Glob(filepath.Join(a, "*.json"))
			if err != nil {
				return nil, err
			}
			if len(names) == 0 {
				return nil, fmt.Errorf("incident: no .json files in %s", a)
			}
			out = append(out, names...) // Glob sorts: unix seconds first
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

func writeScenarioFile(w io.Writer, ss []scenario.Scenario) error {
	e := json.NewEncoder(w)
	e.SetIndent("", " ")
	return e.Encode(scenario.File{Version: scenario.Version, Scenarios: ss})
}
