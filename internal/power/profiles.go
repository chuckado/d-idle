// Package power is the character side of survival slice S-B (docs/survival/design-2026-11-14.md sections 4 to 6): the
// archetype profiles and resolver, the power estimate with a quality flag per term, and the margin of an encounter
// against a danger profile. It only computes and logs; nothing here changes what the bot does.
package power

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed profiles.json
var profilesJSON []byte

// Exposure says how the hero is reached (design 6.2).
type Exposure struct {
	Mode          string  `json:"mode"` // contact | range | screened
	ContactCap    int     `json:"contactCap"`
	CloseSeconds  float64 `json:"closeSeconds"`
	ScreenShare   float64 `json:"screenShare"`
	DefenceCounts bool    `json:"defenceCounts"`
	SecondPool    string  `json:"secondPool"`
}

// Offence says what limits the damage the hero deals.
type Offence struct {
	Mode           string  `json:"mode"` // sustained | mana_burst | pet
	AreaFactor     string  `json:"areaFactor"`
	Resource       string  `json:"resource"`
	RampSeconds    float64 `json:"rampSeconds"`
	ImmunityBlocks bool    `json:"immunityBlocks"`
}

// Archetype is one preset over the two axes.
type Archetype struct {
	ID            string             `json:"id"`
	Exposure      Exposure           `json:"exposure"`
	Offence       Offence            `json:"offence"`
	HazardWeights map[string]float64 `json:"hazardWeights"`
	Escape        struct {
		PortalInterruptible bool   `json:"portalInterruptible"`
		MobilitySkill       string `json:"mobilitySkill"`
	} `json:"escape"`
	Uncertainty struct {
		Derived float64 `json:"derived"`
		Default float64 `json:"default"`
	} `json:"uncertainty"`
}

// Dial is one risk budget row (design 5.2): the margin to enter a zone and to stay in it.
type Dial struct {
	ID    string  `json:"id"`
	Enter float64 `json:"enter"`
	Stay  float64 `json:"stay"`
}

// Params are the model constants; every one carries a tag (U unless a source is named) in Tags.
type Params struct {
	ChickenPct          float64           `json:"chickenPct"`          // share of max life the reflex layer keeps in reserve
	PotionHeal          float64           `json:"potionHeal"`          // life one healing potion restores when its tier is not known
	PotionEffectSeconds float64           `json:"potionEffectSeconds"` // one potion effect at a time (R4)
	SwingSeconds        float64           `json:"swingSeconds"`        // seconds per weapon swing when the weapon speed is not read
	DefaultDpsPerClvl   float64           `json:"defaultDpsPerClvl"`   // kill speed per clvl when no damage is readable (flagged default)
	MercDpsPerLevel     float64           `json:"mercDpsPerLevel"`     // merc dps per merc level (hireling table not read yet)
	MercHold            float64           `json:"mercHold"`            // share of melee attackers a living merc holds
	ElemShare           float64           `json:"elemShare"`           // share of an elemental monster's damage that the resist applies to
	EliteLifeMul        float64           `json:"eliteLifeMul"`
	EliteDmgMul         float64           `json:"eliteDmgMul"`
	ManaPerSecond       float64           `json:"manaPerSecond"` // mana a casting hero spends per second
	ManaFloor           float64           `json:"manaFloor"`     // kill speed share left once the mana pool is empty
	RefDefence          int               `json:"refDefence"`    // hero defence when it is not read (flagged default)
	MaxMargin           float64           `json:"maxMargin"`
	Tags                map[string]string `json:"tags"`
}

// Profiles is the embedded data file.
type Profiles struct {
	Schema       string            `json:"schema"`
	SkillKinds   map[string]string `json:"skillKinds"`
	Archetypes   []Archetype       `json:"archetypes"`
	Dials        []Dial            `json:"dials"`
	DefaultDial  string            `json:"defaultDial"`
	ElementIndex map[string]int    `json:"elementIndex"` // monstats element type -> index into Hero.Resist
	Params       Params            `json:"params"`
}

// Load parses and checks the shipped data.
func Load() (*Profiles, error) { return parse(profilesJSON) }

func parse(raw []byte) (*Profiles, error) {
	var p Profiles
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("power: profiles: %w", err)
	}
	if len(p.Archetypes) == 0 || len(p.Dials) == 0 {
		return nil, fmt.Errorf("power: profiles: no archetypes or dials")
	}
	for _, d := range p.Dials {
		if d.Stay > d.Enter {
			return nil, fmt.Errorf("power: dial %q stay %.2f above enter %.2f", d.ID, d.Stay, d.Enter)
		}
	}
	if _, err := p.Dial(p.DefaultDial); err != nil {
		return nil, err
	}
	for k := range p.Params.tagged() {
		if p.Params.Tags[k] == "" {
			return nil, fmt.Errorf("power: param %q has no tag", k)
		}
	}
	return &p, nil
}

func (p Params) tagged() map[string]bool {
	m := map[string]bool{}
	b, _ := json.Marshal(p)
	var raw map[string]any
	_ = json.Unmarshal(b, &raw)
	for k := range raw {
		if k != "tags" {
			m[k] = true
		}
	}
	return m
}

// Archetype returns a preset by id; an unknown id is an error naming it.
func (p *Profiles) Archetype(id string) (*Archetype, error) {
	for i := range p.Archetypes {
		if p.Archetypes[i].ID == id {
			return &p.Archetypes[i], nil
		}
	}
	return nil, fmt.Errorf("power: unknown archetype %q", id)
}

// Dial returns a risk budget row by id; an unknown id is an error naming it.
func (p *Profiles) Dial(id string) (Dial, error) {
	for _, d := range p.Dials {
		if d.ID == id {
			return d, nil
		}
	}
	return Dial{}, fmt.Errorf("power: unknown dial %q", id)
}

// ArchetypeIDs lists the presets, sorted.
func (p *Profiles) ArchetypeIDs() []string {
	var ids []string
	for _, a := range p.Archetypes {
		ids = append(ids, a.ID)
	}
	sort.Strings(ids)
	return ids
}
