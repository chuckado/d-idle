package survival

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed params.json
var paramsJSON []byte

// StepDown is the observed cost that forces a zone out under one dial (design 5.2): that many chickens plus merc
// deaths, or that many deaths, within the window. 0 turns a row off.
//
// TypicalEnter, BadFloor and RiskValueMul are the pay-for-risk route of the dial (card survtune): a zone the bad-pack
// margin calls too tight to enter can still be entered when the typical pack is comfortable (typical margin at least
// TypicalEnter) and even the bad pack is not lost (best margin at least BadFloor, after the ledger correction). The hero
// then takes the occasional chicken for the pay; the zone's value is multiplied by RiskValueMul so a safe zone that pays
// about the same still wins. TypicalEnter 0 turns the route off for that dial.
type StepDown struct {
	Either       int     `json:"either"`
	Deaths       int     `json:"deaths"`
	TypicalEnter float64 `json:"typicalEnter"`
	BadFloor     float64 `json:"badFloor"`
	RiskValueMul float64 `json:"riskValueMul"`
}

// ArchTune overrides the yield and starvation constants for one archetype (profile id); 0 keeps the global value. Data,
// so a class or build is tuned in params.json and never in Go.
type ArchTune struct {
	YieldShrinkMinutes float64 `json:"yieldShrinkMinutes"`
	StarveCasts        float64 `json:"starveCasts"`
}

// Params are the controller constants; every one is tagged in Tags (all U today).
type Params struct {
	Schema              string              `json:"schema"`
	HalfLifeMinutes     float64             `json:"halfLifeMinutes"`
	HeroHalfLifeMinutes float64             `json:"heroHalfLifeMinutes"`
	HardHalfLifeMinutes float64             `json:"hardHalfLifeMinutes"`
	ShrinkKills         float64             `json:"shrinkKills"`
	ChickenMul          float64             `json:"chickenMul"`
	MercDeathMul        float64             `json:"mercDeathMul"`
	DeathMul            float64             `json:"deathMul"`
	KMin                float64             `json:"kMin"`
	KMax                float64             `json:"kMax"`
	KHeroUnreadMax      float64             `json:"kHeroUnreadMax"`
	PredTakenFloor      float64             `json:"predTakenFloorPerKill"`
	PowerChange         float64             `json:"powerChange"`
	WindowMinutes       float64             `json:"windowMinutes"`
	MercWindowMinutes   float64             `json:"mercWindowMinutes"`
	NoReviveMinutes     float64             `json:"noReviveMinutes"`
	SwitchMargin        float64             `json:"switchMargin"`
	StayMinutes         float64             `json:"stayMinutes"`
	LambdaMax           float64             `json:"lambdaMax"`
	MoveSecondsPerKill  float64             `json:"moveSecondsPerKill"`
	PotionPrice         float64             `json:"potionPrice"`
	ValuePerGold        float64             `json:"valuePerGold"`
	MercPriorMinutes    float64             `json:"mercPriorMinutes"`
	FreshRatio          float64             `json:"freshRatio"`
	FreshMinutes        float64             `json:"freshMinutes"`
	YieldMinMinutes     float64             `json:"yieldMinMinutes"`
	YieldShrinkMinutes  float64             `json:"yieldShrinkMinutes"`
	YieldMaxWeight      float64             `json:"yieldMaxWeight"`
	StarveCasts         float64             `json:"starveCasts"`
	StarveValueMul      float64             `json:"starveValueMul"`
	Archetypes          map[string]ArchTune `json:"archetypes"`
	Dials               map[string]StepDown `json:"dials"`
	Tags                map[string]string   `json:"tags"`
}

// LoadParams parses the embedded constants; a missing dial row or a constant that must be positive is an error.
func LoadParams() (*Params, error) {
	var p Params
	if err := json.Unmarshal(paramsJSON, &p); err != nil {
		return nil, fmt.Errorf("survival: params: %w", err)
	}
	for name, v := range map[string]float64{"halfLifeMinutes": p.HalfLifeMinutes, "heroHalfLifeMinutes": p.HeroHalfLifeMinutes,
		"hardHalfLifeMinutes": p.HardHalfLifeMinutes, "shrinkKills": p.ShrinkKills, "windowMinutes": p.WindowMinutes,
		"stayMinutes": p.StayMinutes, "kMax": p.KMax, "kHeroUnreadMax": p.KHeroUnreadMax, "mercWindowMinutes": p.MercWindowMinutes, "kMin": p.KMin} {
		if v <= 0 {
			return nil, fmt.Errorf("survival: params: %s must be positive", name)
		}
	}
	if p.PredTakenFloor < 0 {
		return nil, fmt.Errorf("survival: params: predTakenFloorPerKill must not be negative")
	}
	if p.YieldShrinkMinutes <= 0 || p.StarveCasts <= 0 || p.YieldMaxWeight < 0 || p.YieldMaxWeight > 1 || p.StarveValueMul <= 0 || p.StarveValueMul > 1 {
		return nil, fmt.Errorf("survival: params: yield and starvation constants out of range")
	}
	for _, d := range []string{"cautious", "balanced", "bold"} {
		r, ok := p.Dials[d]
		if !ok {
			return nil, fmt.Errorf("survival: params: no step-down row for dial %q", d)
		}
		if r.TypicalEnter < 0 || r.BadFloor < 0 || (r.TypicalEnter > 0 && (r.BadFloor <= 0 || r.RiskValueMul <= 0 || r.RiskValueMul > 1)) {
			return nil, fmt.Errorf("survival: params: dial %q: the risk route needs badFloor > 0 and riskValueMul in (0,1]", d)
		}
	}
	return &p, nil
}

// stepDown returns the row of a dial id; an unknown dial is an error naming it.
func (p *Params) stepDown(dial string) (StepDown, error) {
	s, ok := p.Dials[strings.ToLower(dial)]
	if !ok {
		return StepDown{}, fmt.Errorf("survival: unknown dial %q", dial)
	}
	return s, nil
}

// yieldShrink and starveCasts are the constants of an archetype ("" and unknown ids use the global ones).
func (p *Params) yieldShrink(arch string) float64 {
	if a := p.Archetypes[arch]; a.YieldShrinkMinutes > 0 {
		return a.YieldShrinkMinutes
	}
	return p.YieldShrinkMinutes
}

func (p *Params) starveLimit(arch string) float64 {
	if a := p.Archetypes[arch]; a.StarveCasts > 0 {
		return a.StarveCasts
	}
	return p.StarveCasts
}
