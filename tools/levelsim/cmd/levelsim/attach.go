package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/power"
)

// Replay inputs a bot log does not carry (docs/survival/power.md, card powerwire): the build whose attack skill the hero
// plays at each log line's clvl, and an ASSUMED weapon. The weapon flags are a sensitivity input (U), never a reading.
var (
	replayBuild  = flag.String("build", "", "plandata build name (e.g. paladin-normal30) whose main attack and skill level at each clvl feed the skill damage estimate in -shadow and -controller")
	replayWeapon = flag.String("weapon", "", "ASSUMED weapon damage MIN-MAX for the replay hero (U: the log has none; needed by weapon skills)")
	replayAR     = flag.Int("ar", 0, "ASSUMED attack rating for the replay hero when -weapon is set (U)")
)

// attachFor returns the per-hero hook for ShadowOpts, nil without -build.
func attachFor(m *power.Model) (func(h *power.Hero), error) {
	if *replayBuild == "" {
		return nil, nil
	}
	var lo, hi int
	if *replayWeapon != "" {
		a, b, ok := strings.Cut(*replayWeapon, "-")
		var e1, e2 error
		lo, e1 = strconv.Atoi(a)
		hi, e2 = strconv.Atoi(b)
		if !ok || e1 != nil || e2 != nil || hi < lo {
			return nil, fmt.Errorf("-weapon %q: want MIN-MAX", *replayWeapon)
		}
	}
	file := "build-" + *replayBuild + ".json"
	if _, err := power.SpecFromBuild(plandata.FS, file, 30, classAttacks); err != nil {
		return nil, err
	}
	return func(h *power.Hero) {
		spec, err := power.SpecFromBuild(plandata.FS, file, h.Level, classAttacks)
		if err != nil {
			return
		}
		h.Class = spec.Class
		if hi > 0 {
			h.DmgMin, h.DmgMax, h.AR = lo, hi, *replayAR
			h.SetQuality(power.TermOffence, power.Derived)
		}
		_ = m.Attach(h, spec) // an attack the data lacks keeps the constants, flagged by the estimate
	}, nil
}

// classAttacks is the combat-profile adapter of cmd/d2bot/powerwire.go (a main package cannot be imported).
func classAttacks(class string) ([]power.Band, bool, error) {
	var pick *combat.Profile
	for _, p := range combat.Profiles() {
		if strings.EqualFold(p.Class, class) && (pick == nil || (pick.Build != "" && p.Build == "")) {
			pf := p
			pick = &pf
		}
	}
	if pick == nil {
		return nil, false, fmt.Errorf("no combat profile for class %q", class)
	}
	bs := make([]power.Band, 0, len(pick.Bands))
	for _, b := range pick.Bands {
		bs = append(bs, power.Band{MinLevel: b.MinLevel, Requires: b.Requires, Attack: b.Attack})
	}
	return bs, pick.Melee != nil, nil
}
