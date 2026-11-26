package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"d2lod-bot/internal/scenario"
)

//go:embed invariants.json
var invariantsJSON []byte

type invariantPolicy struct {
	Hard       []string `json:"hard"`
	Soft       []string `json:"soft"`
	ReportOnly []string `json:"report_only"`
}

func loadInvariantPolicy() invariantPolicy {
	var p invariantPolicy
	if err := json.Unmarshal(invariantsJSON, &p); err != nil {
		panic("invariants.json: " + err.Error())
	}
	return p
}

// invariantCounts counts the `invariant <name>` WARN lines of the bot log (the monitor's one line per hit, invariants.go in internal/bot).
func invariantCounts(log []byte) map[string]int {
	out := map[string]int{}
	for _, l := range bytes.Split(log, []byte("\n")) {
		if len(l) == 0 || l[0] != '{' || !bytes.Contains(l, []byte(`"invariant `)) {
			continue
		}
		var m struct {
			Msg string `json:"msg"`
		}
		if json.Unmarshal(l, &m) == nil && strings.HasPrefix(m.Msg, "invariant ") {
			out[strings.TrimPrefix(m.Msg, "invariant ")]++
		}
	}
	return out
}

// invariantAsserts turns the hits of a run into assert results: a hard row not in a report-only scenario fails the run.
func (p invariantPolicy) invariantAsserts(id string, hits map[string]int) []scenario.AssertResult {
	hard := map[string]bool{}
	for _, n := range p.Hard {
		hard[n] = true
	}
	reportOnly := false
	for _, pre := range p.ReportOnly {
		if strings.HasPrefix(id, pre) {
			reportOnly = true
		}
	}
	var names []string
	for n := range hits {
		names = append(names, n)
	}
	sort.Strings(names)
	var out []scenario.AssertResult
	for _, n := range names {
		if !hard[n] {
			continue
		}
		r := scenario.AssertResult{Kind: "invariant " + n, OK: reportOnly, Want: "0 hits", Got: fmt.Sprintf("%d hits", hits[n])}
		if reportOnly {
			r.Why = "report only"
		}
		out = append(out, r)
	}
	return out
}
