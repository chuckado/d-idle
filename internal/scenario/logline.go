package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"
)

// LogLine is one bot log line kept for a log_line assertion: Text is `<msg> k=v ...` (attributes sorted), T the line's
// place in the run in virtual minutes.
type LogLine struct {
	T    float64 `json:"t"`
	Text string  `json:"text"`
}

// maxLogLines caps what one run keeps (a pattern that matches every line must not bloat the result file).
const maxLogLines = 500

func (a Assert) validateLogLine() error {
	if a.Pattern == "" {
		return fmt.Errorf("log_line needs pattern")
	}
	if _, err := regexp.Compile(a.Pattern); err != nil {
		return fmt.Errorf("log_line pattern: %w", err)
	}
	if a.Min < 0 || a.Max != nil && *a.Max < a.Min {
		return fmt.Errorf("log_line min %d max bounds", a.Min)
	}
	if a.FromMin < 0 || a.ToMin < 0 || a.ToMin != 0 && a.ToMin <= a.FromMin {
		return fmt.Errorf("log_line window %v to %v", a.FromMin, a.ToMin)
	}
	return nil
}

// CollectLog reads the bot's JSON log (one object per line) and returns the lines that match any log_line pattern of as.
// The virtual minute of a line is its wall-clock share of the log's span times virtualMin: a LINEAR MAPPING (U: lockstep
// does not pace the wall clock evenly, so a window is a coarse bound, not an exact time). ok is false when no line of the
// log parses as a bot log line (a log that is not the bot's): the fact is then not observed.
func CollectLog(as []Assert, log []byte, virtualMin float64) (lines []LogLine, ok bool) {
	var res []*regexp.Regexp
	for _, a := range as {
		if a.Kind == ALogLine {
			if re, err := regexp.Compile(a.Pattern); err == nil {
				res = append(res, re)
			}
		}
	}
	type rec struct {
		t    time.Time
		text string
	}
	var all []rec
	for _, l := range bytes.Split(log, []byte("\n")) {
		if len(l) == 0 || l[0] != '{' {
			continue
		}
		var m map[string]any
		if json.Unmarshal(l, &m) != nil {
			continue
		}
		msg, _ := m["msg"].(string)
		ts, _ := m["time"].(string)
		t, err := time.Parse(time.RFC3339Nano, ts)
		if msg == "" || err != nil {
			continue
		}
		var keys []string
		for k := range m {
			if k != "msg" && k != "time" && k != "level" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		text := msg
		for _, k := range keys {
			text += fmt.Sprintf(" %s=%v", k, m[k])
		}
		all = append(all, rec{t, text})
	}
	if len(all) == 0 {
		return nil, false
	}
	t0, t1 := all[0].t, all[len(all)-1].t
	for _, r := range all {
		for _, re := range res {
			if re.MatchString(r.text) {
				v := 0.0
				if span := t1.Sub(t0); span > 0 {
					v = float64(r.t.Sub(t0)) / float64(span) * virtualMin
				}
				lines = append(lines, LogLine{T: v, Text: r.text})
				break
			}
		}
		if len(lines) >= maxLogLines {
			break
		}
	}
	return lines, true
}

func evalLogLine(a Assert, st State) AssertResult {
	lo := a.Min
	if lo == 0 && (a.Max == nil || *a.Max > 0) {
		lo = 1 // min unset means one line; absence is written max 0
	}
	want := fmt.Sprintf("log lines /%s/ >= %d", a.Pattern, lo)
	if a.Max != nil {
		want += fmt.Sprintf(" and <= %d", *a.Max)
	}
	if a.FromMin > 0 || a.ToMin > 0 {
		want += fmt.Sprintf(" in virtual min %g..%g", a.FromMin, a.ToMin)
	}
	if !st.has("log") {
		return unobs(want, "the bot log")
	}
	re := regexp.MustCompile(a.Pattern) // validated at load
	n, first := 0, ""
	for _, l := range st.LogLines {
		if l.T < a.FromMin || a.ToMin > 0 && l.T > a.ToMin || !re.MatchString(l.Text) {
			continue
		}
		if n == 0 {
			first = l.Text
		}
		n++
	}
	got := fmt.Sprintf("%d lines", n)
	if first != "" {
		if len(first) > 160 {
			first = first[:160] + "..."
		}
		got += ", first: " + first
	}
	return AssertResult{OK: n >= lo && (a.Max == nil || n <= *a.Max), Want: want, Got: got}
}
