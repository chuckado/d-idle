package scenario

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Assertion kinds that read the Vault of the hero after the run (the vault store on disk, the bot's GET /api/vault and
// the vault journal), and the reply fields of an api call.
const (
	AVaultLog      = "vault_log"             // pattern, min (default 1), max: lines of the vault journal (see VaultJournal)
	AVaultCount    = "vault_count"           // min, max: live records (recorded, keep queued, sell queued)
	AVaultPresent  = "vault_records_present" // which "seeded" (default): every seeded record is still live
	AVaultRevealed = "vault_revealed"        // min, max: live records that are revealed (identified, or unsealed by a paid reveal)
	AVaultSold     = "vault_sold"            // min, max: records sold by the instant sale; with index: their gold added up equals the reply's sale.gold, and that equals sale.stashAfter - sale.stashBefore
	AVaultMissed   = "vault_missed"          // min, max: missed.count of GET /api/vault
	AVaultAPI      = "vault_api"             // path, equals or min or max: a field of the last GET /api/vault sample
	vaultWhichSeed = "seeded"
)

// VaultRecFact is one record of the vault after the run.
type VaultRecFact struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	State      string `json:"state"`
	Pin        bool   `json:"pin,omitempty"`
	Identified bool   `json:"identified,omitempty"`
	Unlocked   bool   `json:"unlocked,omitempty"`
	Seeded     bool   `json:"seeded,omitempty"`
	Gold       int    `json:"gold,omitempty"` // a sold record: what the sale paid for it
}

// VaultFacts is what the runner read of the Vault: the store's records, the journal of what happened DURING the run (the
// seeding is not in it) and the last GET /api/vault sample of the bot.
type VaultFacts struct {
	Records []VaultRecFact  `json:"records"`
	Journal []string        `json:"journal,omitempty"`
	API     json.RawMessage `json:"api,omitempty"`
}

// VaultJournal words the event of the vault log (internal/vault) as one line `<word> <code> <id>`. The words are the
// ones the scenarios assert on: recorded, revealed, kept, sold, discarded, trimmed, pinned, queued-keep, queued-sell,
// and `skipped (full)` for the bot's own line when a new drop found the vault full.
func VaultJournal(ev, code, id string) (string, bool) {
	w, ok := map[string]string{"add": "recorded", "reveal": "revealed", "unseal": "revealed", "kept": "kept", "sold": "sold", "soldnow": "sold",
		"discard": "discarded", "pin": "pinned", "keepq": "queued-keep", "sellq": "queued-sell"}[ev]
	if !ok {
		return "", false
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s %s", w, code, id)), true
}

// JournalFromBotLog maps a bot log line to the journal line `skipped (full)` and `equipped ...`, nothing else.
func JournalFromBotLog(text string) (string, bool) {
	switch {
	case strings.Contains(text, "vault full: new records are stopped"):
		return "skipped (full)", true
	case strings.HasPrefix(text, "equip: ") && strings.Contains(text, " worn in "):
		return "equipped " + strings.TrimPrefix(text, "equip: "), true
	}
	return "", false
}

func (a Assert) validateVault() error {
	switch a.Kind {
	case AVaultLog:
		if a.Pattern == "" {
			return fmt.Errorf("vault_log needs pattern")
		}
		if _, err := regexp.Compile(a.Pattern); err != nil {
			return fmt.Errorf("vault_log pattern: %w", err)
		}
	case AVaultCount, AVaultMissed, AVaultRevealed, AVaultSold:
		if a.Max == nil && a.Min == 0 {
			return fmt.Errorf("%s needs min or max", a.Kind)
		}
	case AVaultPresent:
		if a.Which != "" && a.Which != vaultWhichSeed {
			return fmt.Errorf("vault_records_present which %q (want seeded)", a.Which)
		}
	case AVaultAPI:
		if a.Path == "" || a.Equals == nil && a.Max == nil && a.Min == 0 {
			return fmt.Errorf("vault_api needs path and equals, min or max")
		}
	}
	return nil
}

func evalVault(a Assert, st State) AssertResult {
	v := st.Vault
	if v == nil || !st.has("vault") {
		return unobs(a.Kind, "the vault")
	}
	switch a.Kind {
	case AVaultCount:
		n := 0
		for _, r := range v.Records {
			if r.State == "recorded" || r.State == "keepq" || r.State == "sellq" {
				n++
			}
		}
		return rangeResult(n, a.Min, a.Max, "live vault records")
	case AVaultSold:
		n, sum := 0, 0
		for _, r := range v.Records {
			if r.State == "sold" {
				n++
				sum += r.Gold
			}
		}
		res := rangeResult(n, a.Min, a.Max, "sold vault records")
		if a.Index == nil || !res.OK {
			return res
		}
		var reply *APIReply
		for i := range st.APIReplies {
			if st.APIReplies[i].Index == *a.Index {
				reply = &st.APIReplies[i]
			}
		}
		if reply == nil {
			return unobs(res.Want+", gold exact", "the api reply")
		}
		num := func(path string) (float64, bool) {
			x, ok := JSONPath(reply.Body, path)
			f, isNum := x.(float64)
			return f, ok && isNum
		}
		gold, g1 := num("sale.gold")
		after, g2 := num("sale.stashAfter")
		before, g3 := num("sale.stashBefore")
		want := fmt.Sprintf("%s, records' gold %d = sale.gold = stashAfter - stashBefore", res.Want, sum)
		if !g1 || !g2 || !g3 {
			return AssertResult{Want: want, Got: "the reply has no sale numbers: " + clip(reply.Body)}
		}
		exact := int(gold) == sum && int(after-before) == sum && sum > 0
		return AssertResult{OK: exact, Want: want, Got: fmt.Sprintf("%d sold, records %d, sale.gold %d, stash %d -> %d", n, sum, int(gold), int(before), int(after))}
	case AVaultRevealed:
		n := 0
		for _, r := range v.Records {
			if (r.State == "recorded" || r.State == "keepq" || r.State == "sellq") && (r.Identified || r.Unlocked) {
				n++
			}
		}
		return rangeResult(n, a.Min, a.Max, "revealed vault records")
	case AVaultPresent:
		var gone []string
		for _, r := range v.Records {
			if r.Seeded && r.State != "recorded" && r.State != "keepq" && r.State != "sellq" {
				gone = append(gone, r.Code+" "+r.State)
			}
		}
		seeded := 0
		for _, r := range v.Records {
			if r.Seeded {
				seeded++
			}
		}
		return AssertResult{OK: seeded > 0 && len(gone) == 0, Want: "every seeded record still live", Got: fmt.Sprintf("%d seeded, left: %v", seeded, gone)}
	case AVaultLog:
		re := regexp.MustCompile(a.Pattern)
		n := 0
		for _, l := range v.Journal {
			if re.MatchString(l) {
				n++
			}
		}
		min := a.Min
		if min == 0 && a.Max == nil {
			min = 1
		}
		return rangeResult(n, min, a.Max, "vault journal lines matching "+a.Pattern)
	case AVaultMissed:
		return apiNumber(v.API, "missed.count", a.Min, a.Max, "missed drops")
	case AVaultAPI:
		if len(v.API) == 0 {
			return unobs("api "+a.Path, "GET /api/vault")
		}
		val, ok := JSONPath(v.API, a.Path)
		if !ok {
			return AssertResult{Want: "api field " + a.Path, Got: "no such field"}
		}
		return compareField(val, a, "api "+a.Path)
	}
	return AssertResult{Want: a.Kind, Got: "unknown kind"}
}

func rangeResult(n, min int, max *int, what string) AssertResult {
	want := fmt.Sprintf(">= %d %s", min, what)
	if max != nil {
		want = fmt.Sprintf("%d..%d %s", min, *max, what)
	}
	return AssertResult{OK: n >= min && (max == nil || n <= *max), Want: want, Got: fmt.Sprintf("%d", n)}
}

func apiNumber(doc []byte, path string, min int, max *int, what string) AssertResult {
	if len(doc) == 0 {
		return unobs(what, "GET /api/vault")
	}
	val, ok := JSONPath(doc, path)
	f, isNum := val.(float64)
	if !ok || !isNum {
		return AssertResult{Want: what, Got: "no number at " + path}
	}
	return rangeResult(int(f), min, max, what)
}

// compareField checks one decoded JSON value against equals, min and max of the assertion.
func compareField(val any, a Assert, name string) AssertResult {
	want := name
	ok := true
	if a.Equals != nil {
		want += fmt.Sprintf(" = %v", a.Equals)
		ok = jsonEqual(val, a.Equals)
	}
	if a.Min != 0 || a.Max != nil && a.Equals == nil {
		f, isNum := val.(float64)
		want += fmt.Sprintf(" >= %d", a.Min)
		if a.Max != nil {
			want += fmt.Sprintf(" <= %d", *a.Max)
		}
		ok = ok && isNum && int(f) >= a.Min && (a.Max == nil || int(f) <= *a.Max)
	}
	return AssertResult{OK: ok, Want: want, Got: fmt.Sprintf("%v", val)}
}

// BuildJournal words the events of the vault log (log.jsonl, one JSON object per line) with seq above seedSeq (the seeding
// itself is not a result) and the vault lines of the bot's own log, in that order. A line of the log that does not parse
// is skipped (a torn tail).
func BuildJournal(storeLog []byte, seedSeq int64, botLog []byte) []string {
	var out []string
	code := map[string]string{}
	for _, l := range strings.Split(string(storeLog), "\n") {
		var e struct {
			Seq  int64  `json:"seq"`
			Ev   string `json:"ev"`
			ID   string `json:"id"`
			Item *struct {
				Code string `json:"code"`
			} `json:"item"`
		}
		if json.Unmarshal([]byte(l), &e) != nil {
			continue
		}
		if e.Item != nil {
			code[e.ID] = e.Item.Code
		}
		if e.Seq <= seedSeq {
			continue
		}
		if w, ok := VaultJournal(e.Ev, code[e.ID], e.ID); ok {
			out = append(out, w)
		}
	}
	for _, l := range strings.Split(string(botLog), "\n") {
		var m map[string]any
		if len(l) == 0 || l[0] != '{' || json.Unmarshal([]byte(l), &m) != nil {
			continue
		}
		msg, _ := m["msg"].(string)
		text := msg
		for k, v := range m {
			if s, ok := v.(string); ok && k == "err" {
				text += " err=" + s
			}
		}
		if w, ok := JournalFromBotLog(text); ok {
			out = append(out, w)
		}
	}
	return out
}

// NeedsVaultSample is true when an assertion reads the bot's GET /api/vault, so the runner must sample it.
func NeedsVaultSample(s Scenario) bool {
	for _, a := range s.Assert {
		if a.Kind == AVaultMissed || a.Kind == AVaultAPI {
			return true
		}
	}
	return false
}

// NeedsVault is true when any assertion reads the Vault.
func NeedsVault(s Scenario) bool {
	for _, a := range s.Assert {
		switch a.Kind {
		case AVaultLog, AVaultCount, AVaultPresent, AVaultMissed, AVaultAPI, AVaultRevealed, AVaultSold:
			return true
		}
	}
	return false
}
