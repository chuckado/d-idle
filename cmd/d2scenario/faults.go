package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"d2lod-bot/internal/faultproxy"
	"d2lod-bot/internal/scenario"
)

//go:embed faults.json
var faultsJSON []byte

// faultsEnv names a fault set for every run of this process; the soak driver sets it. It beats scenario.faults.
const faultsEnv = "D2SCEN_FAULTS"

func faultSets() (map[string]faultproxy.Spec, error) {
	var f struct {
		Sets map[string]faultproxy.Spec `json:"sets"`
	}
	if err := json.Unmarshal(faultsJSON, &f); err != nil {
		return nil, fmt.Errorf("faults.json: %w", err)
	}
	for n, s := range f.Sets {
		if err := s.Validate(); err != nil {
			return nil, fmt.Errorf("faults.json set %s: %w", n, err)
		}
	}
	return f.Sets, nil
}

func faultSet(name string) (faultproxy.Spec, error) {
	sets, err := faultSets()
	if err != nil {
		return faultproxy.Spec{}, err
	}
	s, ok := sets[name]
	if !ok {
		var names []string
		for n := range sets {
			names = append(names, n)
		}
		sort.Strings(names)
		return faultproxy.Spec{}, fmt.Errorf("no fault set %q (have %s)", name, strings.Join(names, ", "))
	}
	return s, nil
}

// faultsFor is the set a run uses: the environment override, else the scenario's own.
func faultsFor(s scenario.Scenario) string {
	if v := os.Getenv(faultsEnv); v != "" {
		return v
	}
	return s.Faults
}

// startFaults puts the proxy in front of the game socket; the bot is given the returned path.
func startFaults(set, upstream, listen string) (*faultproxy.Proxy, error) {
	spec, err := faultSet(set)
	if err != nil {
		return nil, err
	}
	p := &faultproxy.Proxy{Spec: spec, Upstream: upstream, Listen: listen}
	if err := p.Start(); err != nil {
		return nil, err
	}
	return p, nil
}

// faultReport stops the proxy and writes its events next to the bot log as JSON lines.
func faultReport(set string, p *faultproxy.Proxy, logPath string) *scenario.FaultReport {
	ev := p.Stop()
	rep := &scenario.FaultReport{Set: set, Counts: map[string]int{}, Log: logPath}
	for _, e := range ev {
		rep.Counts[e.Kind]++
	}
	if f, err := os.Create(logPath); err == nil {
		w := bufio.NewWriter(f)
		for _, e := range ev {
			b, _ := json.Marshal(e)
			w.Write(append(b, '\n'))
		}
		w.Flush()
		f.Close()
	}
	return rep
}
