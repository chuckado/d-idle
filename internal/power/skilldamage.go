package power

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// skilldamage.json is the DECOMP29 data (docs/research/decomp29-skilldamage.md): game.exe 1.14d's own damage, mana and
// rate functions run in the emulator (V run), AnimData frame counts (D). Schema "power.skilldamage/<major>"; the loader
// accepts major 1 and ignores fields it does not know, so the file may grow additively.
//
//go:embed skilldamage.json
var skillDamageJSON []byte

const skillDamageMajor = "power.skilldamage/1"

// Synergy sets of the data: the skill alone, every named synergy at 1 point, every named synergy at 20 points.
const (
	setNone = "none"
	setAll1 = "all1"
	setAll  = "all20"
)

type damageRange struct {
	Min []float64 `json:"min"`
	Max []float64 `json:"max"`
}

// Skill is one entry of the data file (the fields the estimator reads).
type Skill struct {
	Name        string                  `json:"-"`
	ID          int                     `json:"id"`
	Class       string                  `json:"class"`
	Anim        string                  `json:"anim"`
	SeqTrans    string                  `json:"seqtrans"`
	RateKind    string                  `json:"rate_kind"` // cast | attack | attack_sq
	Kind        string                  `json:"kind"`
	SrcDam      *float64                `json:"srcdam"` // weapon share in 128ths; nil: the skill does not use the weapon
	DelayFrames *float64                `json:"delay_frames"`
	Mana        []float64               `json:"mana"`
	Elem        map[string]damageRange  `json:"elem"`
	Phys        map[string]damageRange  `json:"phys"`
	Calc        map[string][][4]float64 `json:"calc"`
	CalcDesc    [4]*string              `json:"calc_desc"`
	Levels      []int                   `json:"levels"`
}

type animFrames struct {
	Frames int `json:"frames"`
}

// SkillData is the loaded file.
type SkillData struct {
	Schema      string
	TickSeconds float64
	Skills      map[string]*Skill
	frames      map[string]map[string]map[string]animFrames // class code, mode, weapon class
	attack      map[int][]int                               // weapon speed modifier -> rate by IAS
	cast        []int
}

// LoadSkillData parses the embedded file.
func LoadSkillData() (*SkillData, error) { return parseSkillData(skillDamageJSON) }

func parseSkillData(raw []byte) (*SkillData, error) {
	var f struct {
		Schema string `json:"schema"`
		Units  struct {
			TickSeconds float64 `json:"tick_seconds"`
		} `json:"units"`
		Rate struct {
			AttackWSM0 []int            `json:"attack_wsm0"`
			ByWSM      map[string][]int `json:"attack_by_wsm"`
			Cast       []int            `json:"cast"`
		} `json:"rate"`
		Frames map[string]map[string]map[string]animFrames `json:"frames"`
		Skills map[string]*Skill                           `json:"skills"`
	}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&f); err != nil { // unknown fields allowed: the file is additive
		return nil, fmt.Errorf("power: skilldamage: %w", err)
	}
	if f.Schema != skillDamageMajor {
		return nil, fmt.Errorf("power: skilldamage: schema %q, want %q", f.Schema, skillDamageMajor)
	}
	if f.Units.TickSeconds <= 0 || len(f.Rate.AttackWSM0) == 0 || len(f.Rate.Cast) == 0 || len(f.Skills) == 0 {
		return nil, fmt.Errorf("power: skilldamage: missing tick length, rate arrays or skills")
	}
	d := &SkillData{Schema: f.Schema, TickSeconds: f.Units.TickSeconds, Skills: f.Skills, frames: f.Frames,
		attack: map[int][]int{0: f.Rate.AttackWSM0}, cast: f.Rate.Cast}
	for k, v := range f.Rate.ByWSM {
		var w int
		if _, err := fmt.Sscanf(k, "%d", &w); err != nil {
			return nil, fmt.Errorf("power: skilldamage: weapon speed key %q", k)
		}
		d.attack[w] = v
	}
	for n, s := range f.Skills {
		s.Name = n
	}
	return d, nil
}

// Skill returns an entry by NAME; an unknown name is an error naming it.
func (d *SkillData) Skill(name string) (*Skill, error) {
	s, ok := d.Skills[name]
	if !ok {
		return nil, fmt.Errorf("power: skilldamage: no skill %q in the data", name)
	}
	return s, nil
}

// ClassCode is the three-letter class key of the frame table ("Paladin" -> "pal"); an unknown class is an error naming it.
func (d *SkillData) ClassCode(class string) (string, error) {
	c := strings.ToLower(class)
	if len(c) >= 3 {
		if _, ok := d.frames[c[:3]]; ok {
			return c[:3], nil
		}
	}
	return "", fmt.Errorf("power: skilldamage: no animation frames for class %q", class)
}

// Frames is the AnimData frame count of a class, animation mode and weapon class; a missing cell is an error naming it.
func (d *SkillData) Frames(class, mode, wclass string) (int, error) {
	code, err := d.ClassCode(class)
	if err != nil {
		return 0, err
	}
	fr, ok := d.frames[code][mode][wclass]
	if !ok || fr.Frames <= 0 {
		return 0, fmt.Errorf("power: skilldamage: no frames for class %q mode %q weapon class %q", class, mode, wclass)
	}
	return fr.Frames, nil
}

func ceilDiv(a, b int) int { return (a + b - 1) / b }

// AttackTicks is ceil(256 N / rate) with the rate of FUN_00623f50 (V run) at the given IAS and weapon speed modifier;
// a modifier the file has no row for is an error.
func (d *SkillData) AttackTicks(frames, ias, wsm int) (int, error) {
	row, ok := d.attack[wsm]
	if !ok {
		return 0, fmt.Errorf("power: skilldamage: no attack rate row for weapon speed %d", wsm)
	}
	return ceilDiv(256*frames, row[clamp(ias, 0, len(row)-1)]), nil
}

// CastTicks is ceil(256 N / rate) - 1 for a plain cast and without the - 1 for a sequence skill (V run, cast-trace).
func (d *SkillData) CastTicks(frames, fcr int, sequence bool) int {
	t := ceilDiv(256*frames, d.cast[clamp(fcr, 0, len(d.cast)-1)])
	if !sequence {
		t--
	}
	return t
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }

// at returns the value of a per-level array at a skill level, with synergy points p on every named synergy skill
// (piecewise linear between the sets none, all1 and all20; the damage is linear in the points).
func synergyAt(none, all1, all20 []float64, lvl int, p float64) float64 {
	i := lvl - 1
	if i < 0 || i >= len(none) {
		return 0
	}
	n, a, b := none[i], all1[i], all20[i]
	switch {
	case p <= 0:
		return n
	case p <= 1:
		return n + (a-n)*p
	}
	return a + (b-a)*(math.Min(p, 20)-1)/19
}

func rangeAvg(sets map[string]damageRange, lvl int, p float64) (float64, error) {
	n, ok0 := sets[setNone]
	a, ok1 := sets[setAll1]
	b, ok2 := sets[setAll]
	if !ok0 {
		return 0, nil // no such component (physical damage keeps only none and all20)
	}
	if !ok1 {
		a = n
	}
	if !ok2 {
		b = n
	}
	if lvl < 1 || lvl > len(n.Min) {
		return 0, fmt.Errorf("level %d outside the table (1..%d)", lvl, len(n.Min))
	}
	lo := synergyAt(n.Min, a.Min, b.Min, lvl, p)
	hi := synergyAt(n.Max, a.Max, b.Max, lvl, p)
	return (lo + hi) / 2 / 256, nil
}

// edPercent is the damage percent slot of a weapon skill (the calc slot the data describes as damage percent), 0 when
// the skill has none.
func (s *Skill) edPercent(lvl int, p float64) float64 {
	for i, desc := range s.CalcDesc {
		if desc == nil {
			continue
		}
		dsc := strings.ToLower(*desc)
		if strings.Contains(dsc, "self") || !(strings.Contains(dsc, "dmg%") || strings.Contains(dsc, "damage %") || strings.Contains(dsc, "damage%")) {
			continue
		}
		col := func(set string) []float64 {
			rows := s.Calc[set]
			out := make([]float64, len(rows))
			for j, r := range rows {
				out[j] = r[i]
			}
			return out
		}
		return synergyAt(col(setNone), col(setAll1), col(setAll), lvl, p)
	}
	return 0
}

// multiStrike is true when a calc slot counts hits, bolts or targets per use (Zeal, Charged Bolt): the estimate counts
// one strike per use, a lower bound (U3 of the research doc).
func (s *Skill) multiStrike() bool {
	for _, desc := range s.CalcDesc {
		if desc == nil {
			continue
		}
		d := strings.ToLower(*desc)
		if strings.Contains(d, "hits") || strings.Contains(d, "bolts") || strings.Contains(d, "max targets") {
			return true
		}
	}
	return false
}

// Attack is what the estimator needs to know about the hero's main attack. Every gear field is "known" only when the
// caller read it; the zero value means not read and the result carries the matching flag.
type Attack struct {
	Skill         string  // by NAME, from the build
	Level         int     // skill level with points spent (items not included), 0 = not known
	SynergyPoints float64 // points on each named synergy skill (uniform); < 0 = not known
	FCR, IAS      int     // gear percent; read only when the matching Known flag is set
	FCRKnown      bool
	IASKnown      bool
	WeaponClass   string // AnimData weapon class of the equipped weapon ("1HS"), "" = not read
	WSM           int    // weapon speed modifier of the weapon, 0 when not read
	WeaponAvg     float64
	DmgPct        float64
}

// Flags of an OffenceDPS (each says a number was replaced by the table's first column or a named assumption).
const (
	FlagFCRDefault     = "fcr_default" // no cast rate from gear: 0
	FlagIASDefault     = "ias_default" // no attack rate from gear: 0
	FlagWeaponDefault  = "weapon_class_default"
	FlagWeaponUnread   = "weapon_unread"   // a weapon skill with no weapon damage read: no estimate
	FlagSynergyUnknown = "synergy_unknown" // synergy points not known: 0 (lower bound)
	FlagMastery        = "mastery_none"    // damage masteries and gear spell damage are not in the data
	FlagSequenceRate   = "sequence_rate_u" // attack_sq skills: the rate of a sequence animation is unverified (U1)
	FlagOneStrike      = "one_strike"      // multi-strike or multi-bolt skills counted as one strike per use
)

// OffenceDPS is the data-derived kill speed of the main attack against one target without resist (design 4).
type OffenceDPS struct {
	Skill      string   `json:"skill"`
	Level      int      `json:"level"`
	PerUse     float64  `json:"perUse"`     // damage per cast or swing at hit chance 1
	Seconds    float64  `json:"seconds"`    // seconds per use: the larger of the animation and the skill delay
	DPS        float64  `json:"dps"`        // PerUse / Seconds
	ManaPerSec float64  `json:"manaPerSec"` // mana of one use per Seconds
	NeedsHit   bool     `json:"needsHit"`   // a weapon skill rolls to hit; the caller multiplies by the hit chance
	Quality    Quality  `json:"-"`
	Flags      []string `json:"flags,omitempty"`
}

// Known is true when the estimate is usable (DPS above 0); otherwise Flags says why and nothing was invented.
func (o OffenceDPS) Known() bool { return o.DPS > 0 }

// Estimate computes the OffenceDPS of a hero's main attack. Errors name the skill or class the data lacks; a weapon
// skill without a read weapon returns a result with FlagWeaponUnread and DPS 0 rather than a made-up weapon.
func (d *SkillData) Estimate(class string, a Attack) (OffenceDPS, error) {
	sk, err := d.Skill(a.Skill)
	if err != nil {
		return OffenceDPS{}, err
	}
	out := OffenceDPS{Skill: a.Skill, Level: a.Level, Quality: Derived}
	if a.Level <= 0 {
		return out, fmt.Errorf("power: skilldamage: skill %q has no level", a.Skill)
	}
	if a.Level > len(sk.Levels) && len(sk.Levels) > 0 {
		out.Level = len(sk.Levels)
		a.Level = out.Level
	}
	p := a.SynergyPoints
	if p < 0 {
		p, out.Flags = 0, append(out.Flags, FlagSynergyUnknown)
	}
	el, err := rangeAvg(sk.Elem, a.Level, p)
	if err != nil {
		return out, fmt.Errorf("power: skilldamage: skill %q: %w", a.Skill, err)
	}
	ph, err := rangeAvg(sk.Phys, a.Level, p)
	if err != nil {
		return out, fmt.Errorf("power: skilldamage: skill %q: %w", a.Skill, err)
	}
	per := el + ph
	if sk.SrcDam != nil { // a weapon skill
		out.NeedsHit = true
		if a.WeaponAvg <= 0 {
			out.Quality = Default
			out.Flags = append(out.Flags, FlagWeaponUnread)
			return out, nil
		}
		ed := sk.edPercent(a.Level, p)
		per += a.WeaponAvg * *sk.SrcDam / 128 * (100 + ed + a.DmgPct) / 100
	} else if sk.Kind == "spell" || len(sk.Elem) > 0 {
		out.Flags = append(out.Flags, FlagMastery)
	}
	if per <= 0 {
		return out, fmt.Errorf("power: skilldamage: skill %q at level %d deals no damage in the data", a.Skill, a.Level)
	}
	mode := sk.SeqTrans
	if mode == "" {
		mode = sk.Anim
	}
	wc := a.WeaponClass
	if wc == "" {
		wc = "1HS"
		out.Flags = append(out.Flags, FlagWeaponDefault)
	}
	frames, err := d.Frames(class, mode, wc)
	if err != nil {
		return out, fmt.Errorf("power: skilldamage: skill %q: %w", a.Skill, err)
	}
	var ticks int
	if sk.RateKind == "cast" {
		if !a.FCRKnown {
			out.Flags = append(out.Flags, FlagFCRDefault)
		}
		ticks = d.CastTicks(frames, a.FCR, sk.Anim == "SQ")
	} else {
		if !a.IASKnown {
			out.Flags = append(out.Flags, FlagIASDefault)
		}
		if sk.RateKind == "attack_sq" {
			out.Flags = append(out.Flags, FlagSequenceRate)
		}
		if ticks, err = d.AttackTicks(frames, a.IAS, a.WSM); err != nil {
			return out, fmt.Errorf("power: skilldamage: skill %q: %w", a.Skill, err)
		}
	}
	if sk.multiStrike() {
		out.Flags = append(out.Flags, FlagOneStrike)
	}
	if sk.DelayFrames != nil {
		ticks = max(ticks, int(*sk.DelayFrames))
	}
	if ticks < 1 {
		return out, fmt.Errorf("power: skilldamage: skill %q gives %d ticks", a.Skill, ticks)
	}
	out.Seconds = float64(ticks) * d.TickSeconds
	out.PerUse = per
	out.DPS = per / out.Seconds
	if a.Level <= len(sk.Mana) {
		out.ManaPerSec = sk.Mana[a.Level-1] / out.Seconds
	}
	return out, nil
}

// SwingSeconds is the plain weapon swing of a class from the rate data (the replacement of the constant
// Params.SwingSeconds): one A1 animation at IAS 0 with the given weapon class and speed modifier.
func (d *SkillData) SwingSeconds(class, wclass string, wsm int) (float64, error) {
	fr, err := d.Frames(class, "A1", wclass)
	if err != nil {
		return 0, err
	}
	t, err := d.AttackTicks(fr, 0, wsm)
	if err != nil {
		return 0, err
	}
	return float64(t) * d.TickSeconds, nil
}
