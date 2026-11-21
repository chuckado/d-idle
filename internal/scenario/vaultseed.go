package scenario

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/vault"
)

// WornItem is one item the hero wears at the start (pre.worn). The runner gives it to the pack with the declared SETUP
// verb and wears it with the fork's `item equip` (a lab verb), so the item is worn in the saved hero the bot loads.
type WornItem struct {
	Code    string `json:"code"`
	Slot    string `json:"slot"` // head, neck, torso, rhand, lhand, rring, lring, belt, feet, gloves
	Quality int    `json:"quality,omitempty"`
	Unid    bool   `json:"unid,omitempty"`
}

// BodySlot maps a slot word to the game's body location (internal/d2, the words of the Inventory page).
var BodySlot = map[string]int{"head": d2.BodyHead, "neck": d2.BodyNeck, "torso": d2.BodyTorso, "rhand": d2.BodyRHand,
	"lhand": d2.BodyLHand, "rring": d2.BodyRRing, "lring": d2.BodyLRing, "belt": d2.BodyBelt, "feet": d2.BodyFeet, "gloves": d2.BodyGloves}

// VaultSeed is pre.vault_records: records written into the hero's Vault before the start, from real items the lab d2d
// gives, reads (`item full`) and takes off the hero (`item vault`), so every record carries real item bytes. Either
// Items (explicit) or Count (fillers) or both.
type VaultSeed struct {
	// Count adds that many filler records of Quality (a word: low, normal, superior, magic, or a quality number)
	// after Items; fillers are distinct real items, never pinned unless Pinned says so.
	Count   int       `json:"count,omitempty"`
	Quality string    `json:"quality,omitempty"`
	Pinned  int       `json:"pinned,omitempty"` // how many of the first records (Items first) are pinned
	Items   []VaultIt `json:"items,omitempty"`
	// Capacity is the vault capacity the scenario assumes; it must equal the table base because the runner cannot change
	// the bot's capacity (Ether steps are bought, not set): a different value makes the run BLOCKED, never silently wrong.
	Capacity int `json:"capacity,omitempty"`
}

// VaultIt is one explicit record: the given item's code, quality, whether it is identified, and pin.
type VaultIt struct {
	Code    string `json:"code"`
	Quality int    `json:"quality,omitempty"` // docs/forkcheats.md: 4 magic, 6 rare; 0 leaves the code's default
	// Unid makes the record unidentified (the item is given with the `unid` flag; its affixes are held sealed).
	Unid bool `json:"unid,omitempty"`
	// Sealed is the held affix text ("stat:param:value;...") written when the fork sent none for the unidentified item
	// (U: it may not), so a gold reveal has something to unseal.
	Sealed string `json:"sealed,omitempty"`
	Pin    bool   `json:"pin,omitempty"`
	N      int    `json:"n,omitempty"` // copies (default 1); copies of a magic or better item are distinct items
}

var qualityWords = map[string]int{"low": 1, "normal": 2, "superior": 3, "magic": 4, "rare": 6, "": 2}

// FillerQuality resolves VaultSeed.Quality to the verb's number.
func (v VaultSeed) FillerQuality() (int, error) {
	if n, err := strconv.Atoi(v.Quality); err == nil && n >= 1 && n <= 8 {
		return n, nil
	}
	q, ok := qualityWords[strings.ToLower(v.Quality)]
	if !ok {
		return 0, fmt.Errorf("vault_records quality %q (want low, normal, superior, magic, rare or 1..8)", v.Quality)
	}
	return q, nil
}

// Total is the number of records the seed writes.
func (v VaultSeed) Total() int {
	n := v.Count
	for _, it := range v.Items {
		n += max(it.N, 1)
	}
	return n
}

// Validate checks the seed alone; capacity is checked against the table base here so a wrong scenario is refused at load.
func (v VaultSeed) Validate() error {
	if v.Total() == 0 {
		return fmt.Errorf("vault_records seeds nothing (count or items)")
	}
	if v.Count < 0 || v.Pinned < 0 || v.Pinned > v.Total() {
		return fmt.Errorf("vault_records count %d pinned %d", v.Count, v.Pinned)
	}
	if _, err := v.FillerQuality(); err != nil {
		return err
	}
	for i, it := range v.Items {
		if it.Code == "" || len(it.Code) != 3 && len(it.Code) != 4 {
			return fmt.Errorf("vault_records.items[%d]: code %q", i, it.Code)
		}
		if it.Quality < 0 || it.Quality > 8 || it.N < 0 {
			return fmt.Errorf("vault_records.items[%d]: quality %d n %d", i, it.Quality, it.N)
		}
	}
	if v.Capacity != 0 {
		t, err := vault.DefaultTiers()
		if err != nil {
			return err
		}
		if v.Capacity != t.Base {
			return fmt.Errorf("vault_records.capacity %d: the runner cannot set it (the table base is %d; Ether steps change it)", v.Capacity, t.Base)
		}
	}
	return nil
}

// Validate checks one worn item.
func (w WornItem) Validate() error {
	if _, ok := BodySlot[w.Slot]; !ok {
		return fmt.Errorf("pre.worn: slot %q", w.Slot)
	}
	if w.Code == "" || w.Quality < 0 || w.Quality > 8 {
		return fmt.Errorf("pre.worn: code %q quality %d", w.Code, w.Quality)
	}
	return nil
}

// ---- templating of trigger.api ----

// Resolver substitutes {{...}} in an api call. VaultIDs are the ids of the seeded records in seed order (Items first,
// then fillers); Inventory is the body of GET /api/inventory read when the call is made (nil: not read).
//
//	{{vault.N.id}}                       the id of the N-th seeded record (a string)
//	{{pack.item[code=cap].id}}           the id of the first backpack item with that code (a number)
//	{{pack.item[name=cap,nth=1].id}}     the second one (nth counts from 0); name is a case-insensitive part of the item's
//	                                     name or base name (the inventory view's `code` is the numeric base class, not the 3-letter code)
//	{{worn.head.id}}                     the id of the item worn in that slot (a number)
type Resolver struct {
	VaultIDs  []string
	Inventory []byte
}

var tmplRe = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)
var packRe = regexp.MustCompile(`^pack\.item\[([^\]]+)\]\.id$`)

// HasTemplate reports whether any string of v (a body, a path) holds a placeholder.
func HasTemplate(v any) bool {
	found := false
	walk(v, func(s string) string { found = found || tmplRe.MatchString(s); return s })
	return found
}

// NeedsInventory reports whether a placeholder of v reads the inventory.
func NeedsInventory(v any) bool {
	need := false
	walk(v, func(s string) string {
		for _, m := range tmplRe.FindAllStringSubmatch(s, -1) {
			need = need || strings.HasPrefix(m[1], "pack.") || strings.HasPrefix(m[1], "worn.")
		}
		return s
	})
	return need
}

func walk(v any, f func(string) string) any {
	switch x := v.(type) {
	case string:
		return f(x)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = walk(e, f)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = walk(e, f)
		}
		return out
	}
	return v
}

// ResolveString substitutes the placeholders of one string (a path): every value becomes text.
func (r Resolver) ResolveString(s string) (string, error) {
	var first error
	out := tmplRe.ReplaceAllStringFunc(s, func(m string) string {
		v, err := r.lookup(tmplRe.FindStringSubmatch(m)[1])
		if err != nil && first == nil {
			first = err
		}
		return fmt.Sprint(v)
	})
	return out, first
}

// ResolveBody substitutes the placeholders of a JSON body. A string that is exactly one placeholder takes the value's own
// type (an item id of the inventory is a number); a placeholder inside a longer string becomes text.
func (r Resolver) ResolveBody(body map[string]any) (map[string]any, error) {
	if body == nil {
		return nil, nil
	}
	var first error
	out := walk(body, func(s string) string { return s })
	out = r.sub(out, &first)
	m, _ := out.(map[string]any)
	return m, first
}

func (r Resolver) sub(v any, first *error) any {
	switch x := v.(type) {
	case string:
		if m := tmplRe.FindStringSubmatch(x); m != nil && m[0] == x {
			val, err := r.lookup(m[1])
			if err != nil && *first == nil {
				*first = err
			}
			return val
		}
		s, err := r.ResolveString(x)
		if err != nil && *first == nil {
			*first = err
		}
		return s
	case map[string]any:
		for k, e := range x {
			x[k] = r.sub(e, first)
		}
	case []any:
		for i, e := range x {
			x[i] = r.sub(e, first)
		}
	}
	return v
}

func (r Resolver) lookup(expr string) (any, error) {
	parts := strings.Split(expr, ".")
	switch {
	case len(parts) == 3 && parts[0] == "vault" && parts[2] == "id":
		n, err := strconv.Atoi(parts[1])
		if err != nil || n < 0 || n >= len(r.VaultIDs) {
			return nil, fmt.Errorf("template {{%s}}: the seed has %d vault records", expr, len(r.VaultIDs))
		}
		return r.VaultIDs[n], nil
	case len(parts) == 3 && parts[0] == "worn" && parts[2] == "id":
		return r.invID(expr, func(inv map[string]json.RawMessage) ([]json.RawMessage, error) {
			var eq map[string]json.RawMessage
			if json.Unmarshal(inv["equipment"], &eq) != nil || len(eq[parts[1]]) == 0 || string(eq[parts[1]]) == "null" {
				return nil, fmt.Errorf("nothing worn in %s", parts[1])
			}
			return []json.RawMessage{eq[parts[1]]}, nil
		}, "", "", 0)
	}
	if m := packRe.FindStringSubmatch(expr); m != nil {
		code, name, nth := "", "", 0
		for _, kv := range strings.Split(m[1], ",") {
			k, v, _ := strings.Cut(strings.TrimSpace(kv), "=")
			switch k {
			case "code":
				code = v
			case "name":
				name = strings.ToLower(v)
			case "nth":
				nth, _ = strconv.Atoi(v)
			default:
				return nil, fmt.Errorf("template {{%s}}: filter %q (want code, name, nth)", expr, k)
			}
		}
		if code == "" && name == "" {
			return nil, fmt.Errorf("template {{%s}}: no code or name filter", expr)
		}
		return r.invID(expr, func(inv map[string]json.RawMessage) ([]json.RawMessage, error) {
			var bp []json.RawMessage
			_ = json.Unmarshal(inv["backpack"], &bp)
			return bp, nil
		}, code, name, nth)
	}
	return nil, fmt.Errorf("template {{%s}}: unknown (vault.N.id, pack.item[code=..].id, worn.<slot>.id)", expr)
}

// invID finds the id of the nth item with code (any code when empty) among the items list returns.
func (r Resolver) invID(expr string, list func(map[string]json.RawMessage) ([]json.RawMessage, error), code, name string, nth int) (any, error) {
	if len(r.Inventory) == 0 {
		return nil, fmt.Errorf("template {{%s}}: the inventory could not be read", expr)
	}
	var inv map[string]json.RawMessage
	if err := json.Unmarshal(r.Inventory, &inv); err != nil {
		return nil, fmt.Errorf("template {{%s}}: inventory reply: %w", expr, err)
	}
	items, err := list(inv)
	if err != nil {
		return nil, fmt.Errorf("template {{%s}}: %w", expr, err)
	}
	seen := 0
	var codes []string
	var first string
	for _, raw := range items {
		var it struct {
			ID   float64 `json:"id"`
			Code string  `json:"code"`
			Name string  `json:"name"`
			Base struct {
				Name string `json:"name"`
			} `json:"base"`
		}
		if json.Unmarshal(raw, &it) != nil {
			continue
		}
		codes = append(codes, it.Code)
		if first == "" {
			first = string(raw)
			if len(first) > 300 {
				first = first[:300]
			}
		}
		if code != "" && it.Code != code || name != "" && !strings.Contains(strings.ToLower(it.Name), name) && !strings.Contains(strings.ToLower(it.Base.Name), name) {
			continue
		}
		if seen == nth {
			return int64(it.ID), nil
		}
		seen++
	}
	return nil, fmt.Errorf("template {{%s}}: no such item in the inventory (it lists %v; first %s; keys %s)", expr, codes, first, invKeys(inv))
}

// ---- reading JSON replies ----

// JSONPath reads a dotted path (a.b.0.c) out of a JSON document; ok is false when the path does not exist.
func JSONPath(doc []byte, path string) (v any, ok bool) {
	if json.Unmarshal(doc, &v) != nil {
		return nil, false
	}
	for _, p := range strings.Split(path, ".") {
		switch x := v.(type) {
		case map[string]any:
			if v, ok = x[p]; !ok {
				return nil, false
			}
		case []any:
			i, err := strconv.Atoi(p)
			if err != nil || i < 0 || i >= len(x) {
				return nil, false
			}
			v = x[i]
		default:
			return nil, false
		}
	}
	return v, true
}

// jsonEqual compares two decoded JSON values (numbers as float64, as encoding/json decodes them).
func jsonEqual(a, b any) bool {
	na, nb := norm(a), norm(b)
	return reflect.DeepEqual(na, nb)
}

func norm(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}

func invKeys(inv map[string]json.RawMessage) string {
	var k []string
	for name := range inv {
		k = append(k, name)
	}
	sort.Strings(k)
	return strings.Join(k, ",")
}
