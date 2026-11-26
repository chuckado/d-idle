package admit

import (
	"encoding/json"
	"fmt"
	"sync"

	"d2lod-bot/internal/plandata"
)

// Rows is internal/plandata/admission.json: the floors as data keyed by area and archetype.
type Rows struct {
	Roads      []Road       `json:"roads"`
	KnobFloors []KnobFloor  `json:"knobFloors"`
	QuestWalk  QuestWalkRow `json:"questWalk"`
	Poverty    struct {
		NoPotionGap int `json:"noPotionGap"`
		FromLevel   int `json:"fromLevel"`
	} `json:"poverty"`
	KitFloor struct {
		PotionsBelow int `json:"potionsBelow"`
		FromLevel    int `json:"fromLevel"`
	} `json:"kitFloor"`
	Arrival ArrivalRow `json:"arrival"`
	Lean struct {
		PotionsBelow int `json:"potionsBelow"`
		EntryLifePct int `json:"entryLifePct"`
	} `json:"lean"`
}

// Road is a dangerous leg: a level floor and a max life floor per archetype, and the kit. 0 = no rule.
type Road struct {
	ID         string         `json:"id"`
	Areas      []int          `json:"areas"`
	MinLevel   map[string]int `json:"minLevel"`
	MinMaxLife map[string]int `json:"minMaxLife"`
	Kit        struct {
		Healing int `json:"healing"`
		Portal  int `json:"portal"`
	} `json:"kit"`
}

// KnobFloor is a level floor whose number is an existing knob (named per archetype, filled in View.Knobs by the caller), so a
// knob value is not copied into data.
type KnobFloor struct {
	ID        string            `json:"id"`
	Areas     []int             `json:"areas"`
	LevelKnob map[string]string `json:"levelKnob"`
}

// QuestWalkRow is the kit every quest walk into a field needs, from FromLevel up to BelowLevel.
type QuestWalkRow struct {
	FromLevel  int `json:"fromLevel"`
	BelowLevel int `json:"belowLevel"`
	Healing    int `json:"healing"`
	Portal     int `json:"portal"`
}

// ArrivalRow is the crowd-on-arrival rule (card act2temple): in one of Areas, within WithinSec of arriving, CrowdN hostiles within
// Radius subtiles and life at or below LifePct leave at once (portal, no walk away from a crowd that stands at 1 to 2 subtiles).
// CrowdN 0 = off. The quest crowd override (combat.crowdAbandonQuestN) fights up to 40, so the row is the exception for the
// dungeons whose arrival nook is a known burst (Act 2 dungeons: the Claw Viper Temple stand-in died 1 s after arriving among
// 22 to 39 hostiles, V runs 20261126-110921 series).
type ArrivalRow struct {
	Areas     []int `json:"areas"`
	CrowdN    int   `json:"crowdN"`
	Radius    int   `json:"radius"`
	WithinSec int   `json:"withinSec"`
	LifePct   int   `json:"lifePct"`
}

// ArrivalFor returns the arrival row when area has one.
func (r Rows) ArrivalFor(area int) (ArrivalRow, bool) {
	a := r.Arrival
	if a.CrowdN <= 0 || !r.has(a.Areas, area) {
		return ArrivalRow{}, false
	}
	return a, true
}

var (
	once sync.Once
	rows Rows
	lerr error
)

// Load returns the embedded rows. An unreadable file is an error, never a silent default.
func Load() (Rows, error) {
	once.Do(func() {
		b, err := plandata.FS.ReadFile("admission.json")
		if err != nil {
			lerr = err
			return
		}
		if lerr = json.Unmarshal(b, &rows); lerr != nil {
			lerr = fmt.Errorf("admission.json: %w", lerr)
		}
	})
	return rows, lerr
}
