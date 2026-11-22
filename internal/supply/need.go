package supply

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Line is one planned supply. Only the lines the first stages act on are modelled; repair, ammo, merc and keys join
// with the cards that need them.
type Line string

const (
	Healing      Line = "healing"
	Mana         Line = "mana"
	Rejuvenation Line = "rejuvenation"
	Portal       Line = "portal"
	GoldCarry    Line = "gold_carry" // gold the excursion itself needs in the purse
)

var lines = map[Line]bool{Healing: true, Mana: true, Rejuvenation: true, Portal: true, GoldCarry: true}

// Qty is one line of a need: below Min the hero does not leave, Target is what a visit fills to when gold allows.
type Qty struct {
	Min, Target int
	Max         int    // the most the hero can carry of this line (0 = unknown); a Min above it is clamped
	MinWhy      string // the rule that set the min, for the trace and the viewer
}

// Need is the supply need of one excursion.
type Need struct {
	Excursion string
	Profile   string
	Lines     map[Line]Qty
}

// Inputs are the build's own needs, read by the caller from the build profile and the belt plan (never from a class
// list here).
type Inputs struct {
	ManaNeed   int          // 0, 1 or 2: how much the build uses mana potions
	Capacity   map[Line]int // belt capacity plus pack buffer per line, 0 or absent = unknown
	BaseTarget map[Line]int // belt capacity plus pack buffer: the target when gold allows (section 3.2 layer 1)
}

type row struct {
	Min        int `json:"min"`
	Target     int `json:"target"`
	ManaNeedAt int `json:"manaNeedAtLeast"` // the line applies only to a build whose ManaNeed is at least this
}

//go:embed profiles.json
var profilesJSON []byte

var (
	profOnce sync.Once
	profs    map[string]map[Line]row
	profErr  error
)

func loadProfiles() (map[string]map[Line]row, error) {
	profOnce.Do(func() {
		var f struct {
			Profiles map[string]map[string]row `json:"profiles"`
		}
		if err := json.Unmarshal(profilesJSON, &f); err != nil {
			profErr = fmt.Errorf("supply: profiles.json: %w", err)
			return
		}
		profs = map[string]map[Line]row{}
		for name, ls := range f.Profiles {
			m := map[Line]row{}
			for l, r := range ls {
				if !lines[Line(l)] {
					profErr = fmt.Errorf("supply: profiles.json: profile %q names unknown line %q", name, l)
					return
				}
				if r.Min < 0 || r.Target < 0 {
					profErr = fmt.Errorf("supply: profiles.json: profile %q line %q has a negative number", name, l)
					return
				}
				m[Line(l)] = r
			}
			profs[name] = m
		}
	})
	return profs, profErr
}

// Profiles lists the profile names, sorted.
func Profiles() ([]string, error) {
	p, err := loadProfiles()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(p))
	for n := range p {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}

// NeedFor derives the need of an excursion from a profile row and the build's inputs. An unknown profile is an error
// that names it; there is no fallback profile.
func NeedFor(excursion, profile string, in Inputs) (Need, error) {
	p, err := loadProfiles()
	if err != nil {
		return Need{}, err
	}
	rows, ok := p[profile]
	if !ok {
		names, _ := Profiles()
		return Need{}, fmt.Errorf("supply: unknown profile %q (known: %s)", profile, strings.Join(names, ", "))
	}
	n := Need{Excursion: excursion, Profile: profile, Lines: map[Line]Qty{}}
	for l, r := range rows {
		if in.ManaNeed < r.ManaNeedAt {
			continue
		}
		q := Qty{Min: r.Min, Target: max(r.Target, r.Min, in.BaseTarget[l]), Max: in.Capacity[l], MinWhy: "profile " + profile}
		n.Lines[l] = q
	}
	return n, nil
}
