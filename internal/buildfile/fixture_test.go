package buildfile

import "strings"

// fixtureRules is a hand-written skill table (U: plausible reqlevels and prerequisites, NOT read from skills.txt;
// the repo has no extracted skills.txt to test against). ~33 skills of four classes.
type fixtureRules struct{}

var fixtureSkills = map[string]SkillInfo{}

func init() {
	add := func(class, name string, req int, pre ...string) {
		fixtureSkills[strings.ToLower(name)] = SkillInfo{Name: name, Class: class, ReqLevel: req, MaxLevel: 20, ReqSkills: pre}
	}
	add("sorceress", "Fire Bolt", 1)
	add("sorceress", "Charged Bolt", 1)
	add("sorceress", "Ice Bolt", 1)
	add("sorceress", "Frozen Armor", 1)
	add("sorceress", "Warmth", 1)
	add("sorceress", "Ice Blast", 6, "Ice Bolt")
	add("sorceress", "Frost Nova", 6, "Frozen Armor")
	add("sorceress", "Fire Ball", 18, "Fire Bolt")
	add("sorceress", "Glacial Spike", 18, "Ice Blast", "Frost Nova")
	add("sorceress", "Teleport", 18, "Frost Nova")
	add("sorceress", "Blizzard", 24, "Glacial Spike")
	add("paladin", "Might", 1)
	add("paladin", "Sacrifice", 1)
	add("paladin", "Defiance", 1)
	add("paladin", "Smite", 6, "Sacrifice")
	add("paladin", "Holy Bolt", 6)
	add("paladin", "Zeal", 12, "Sacrifice")
	add("paladin", "Charge", 12, "Sacrifice")
	add("paladin", "Blessed Aim", 12, "Might")
	add("paladin", "Holy Shield", 18, "Smite", "Charge", "Defiance")
	add("paladin", "Blessed Hammer", 18, "Holy Bolt")
	add("paladin", "Concentration", 18, "Blessed Aim")
	add("paladin", "Vigor", 24)
	add("paladin", "Fanaticism", 30, "Zeal")
	add("necromancer", "Raise Skeleton", 1)
	add("necromancer", "Teeth", 1)
	add("necromancer", "Amplify Damage", 6)
	add("necromancer", "Bone Armor", 12, "Teeth")
	add("necromancer", "Bone Spear", 24, "Teeth", "Bone Armor")
	add("barbarian", "Bash", 1)
	add("barbarian", "Howl", 6)
	add("barbarian", "Iron Skin", 6)
	add("barbarian", "Double Swing", 6, "Bash")
	add("barbarian", "Frenzy", 24, "Double Swing")
}

// start stats U (charstats.txt values from memory; not read here).
var fixtureStart = map[string][4]int{
	"sorceress": {10, 25, 10, 35}, "paladin": {25, 20, 25, 15},
	"necromancer": {15, 25, 15, 25}, "barbarian": {30, 20, 25, 10},
}

func (fixtureRules) Skill(name string) (SkillInfo, bool) {
	in, ok := fixtureSkills[strings.ToLower(name)]
	return in, ok
}
func (fixtureRules) StartStats(c string) ([4]int, bool) { s, ok := fixtureStart[c]; return s, ok }

// looseRules accepts any skill name of one class with reqlevel 1, no prerequisites, max 20: used to round-trip
// the repo's real tables, whose skills the fixture does not all cover. Exactness then rests on the when.clvl
// every FromRows goal carries, not on the rules.
type looseRules struct{ class string }

func (l looseRules) Skill(name string) (SkillInfo, bool) {
	return SkillInfo{Name: name, Class: l.class, ReqLevel: 1, MaxLevel: 20}, true
}
func (looseRules) StartStats(c string) ([4]int, bool) { s, ok := fixtureStart[c]; return s, ok }

// cyclicRules adds two skills that need each other.
type cyclicRules struct{ fixtureRules }

func (c cyclicRules) Skill(name string) (SkillInfo, bool) {
	switch strings.ToLower(name) {
	case "ping":
		return SkillInfo{Name: "Ping", Class: "paladin", ReqLevel: 1, ReqSkills: []string{"Pong"}}, true
	case "pong":
		return SkillInfo{Name: "Pong", Class: "paladin", ReqLevel: 1, ReqSkills: []string{"Ping"}}, true
	}
	return c.fixtureRules.Skill(name)
}
