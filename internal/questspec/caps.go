package questspec

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Tier is how a backend can read a fact. "none" means it cannot.
type Tier string

const (
	TierMemory Tier = "memory"
	TierSave   Tier = "save"
	TierProxy  Tier = "proxy"
	TierNone   Tier = "none"
)

type Reader struct {
	Tier Tier `json:"tier"`
	// Via is the backend verb the read needs (a lab-only verb makes the fact unreadable on the shipped build).
	Via string `json:"via,omitempty"`
	// Volatile marks a fact the bot only observes in the running game (a proxy): it is lost on restart.
	Volatile bool   `json:"volatile,omitempty"`
	Note     string `json:"note,omitempty"` // "RCn: ..." names the documented root cause
}

// Caps is what one backend can read and do. Readers are keyed by fact name,
// or "name/lastarg" for a more specific entry (object_state/lit).
type Caps struct {
	Schema  string `json:"schema"`
	Backend string `json:"backend"`
	// Verbs are the control-channel verbs the backend answers (the sandbox build's SHIPPED set, or the lab set).
	Verbs []string `json:"verbs"`
	// Knowledge lists how a target can become known to the bot: preset_list (rows of debug presets), preset_near
	// (a row appears only within about 15 cells of the hero), room_rows (the room rows of debug presets).
	Knowledge     []string          `json:"knowledge"`
	KnowledgeNote string            `json:"knowledge_note,omitempty"`
	Readers       map[string]Reader `json:"readers"`
	ItemMatch     string            `json:"item_match"` // code | name
	ItemMatchNote string            `json:"item_match_note,omitempty"`
	ProgressRules []string          `json:"progress_rules"`
	ProgressNote  string            `json:"progress_note,omitempty"`
}

func LoadCaps(data []byte) (*Caps, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var c Caps
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("questcaps: %w", err)
	}
	if c.Schema != CapsSchema || c.Backend == "" || (c.ItemMatch != "code" && c.ItemMatch != "name") {
		return nil, fmt.Errorf("questcaps: bad schema, backend or item_match")
	}
	for k, r := range c.Readers {
		switch r.Tier {
		case TierMemory, TierSave, TierProxy, TierNone:
		default:
			return nil, fmt.Errorf("questcaps: reader %s: tier %q", k, r.Tier)
		}
	}
	return &c, nil
}

func (c *Caps) reader(p Pred) (Reader, bool) {
	if n := len(p.Args); n > 0 {
		if r, ok := c.Readers[p.Name+"/"+p.Args[n-1]]; ok {
			return r, true
		}
	}
	r, ok := c.Readers[p.Name]
	return r, ok
}

func (c *Caps) readable(p Pred) bool {
	r, ok := c.reader(p)
	return ok && r.Tier != TierNone && (r.Via == "" || contains(c.Verbs, r.Via))
}

func (c *Caps) volatile(p Pred) bool {
	r, ok := c.reader(p)
	return ok && r.Volatile
}

// missingVerb returns the verb a fact needs but the backend lacks, or "".
func (c *Caps) missingVerb(p Pred) string {
	if r, ok := c.reader(p); ok && r.Tier != TierNone && r.Via != "" && !contains(c.Verbs, r.Via) {
		return r.Via
	}
	return ""
}
