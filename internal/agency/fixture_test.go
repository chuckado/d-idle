package agency

import "d2lod-bot/internal/data"

// fixtureRules is a hand-written subset of skills.txt (reqlevel, reqskill). U: written from memory of the game, not
// read from the game tables (none are on this box); the numbers only have to be self-consistent for the tests.
func fixtureRules() Rules {
	rows := []struct {
		n   string
		lvl int
		req []string
	}{
		{"Fire Bolt", 1, nil}, {"Charged Bolt", 1, nil}, {"Ice Bolt", 1, nil}, {"Frozen Armor", 1, nil},
		{"Fire Ball", 12, []string{"Fire Bolt"}}, {"Ice Blast", 12, []string{"Ice Bolt"}}, {"Lightning", 12, []string{"Charged Bolt"}},
		{"Glacial Spike", 18, []string{"Ice Blast"}}, {"Blizzard", 24, []string{"Glacial Spike"}}, {"Chain Lightning", 30, []string{"Lightning"}},
		{"Sacrifice", 1, nil}, {"Might", 1, nil}, {"Smite", 6, []string{"Sacrifice"}}, {"Zeal", 12, []string{"Sacrifice", "Smite"}},
		{"Holy Fire", 6, []string{"Might"}}, {"Holy Bolt", 6, nil}, {"Blessed Hammer", 18, []string{"Holy Bolt"}},
		{"Prayer", 1, nil}, {"Thorns", 1, nil}, {"Blessed Aim", 6, nil}, {"Concentration", 18, []string{"Blessed Aim"}},
		{"Resist Fire", 1, nil}, {"Defiance", 6, nil},
		{"Bash", 1, nil}, {"Double Swing", 12, []string{"Bash"}}, {"Frenzy", 24, []string{"Double Swing"}},
		{"Leap Attack", 6, nil}, {"Whirlwind", 30, []string{"Double Swing", "Leap Attack"}},
		{"Raise Skeleton", 1, nil}, {"Skeleton Mastery", 1, []string{"Raise Skeleton"}}, {"Teeth", 1, nil},
		{"Bone Armor", 6, []string{"Teeth"}}, {"Bone Spear", 24, []string{"Teeth", "Bone Armor"}}, {"Clay Golem", 6, []string{"Raise Skeleton"}},
	}
	r := Rules{Skills: map[string]data.Skill{}, MaxQuest: 12}
	for _, x := range rows {
		s := data.Skill{Key: x.n, ReqLevel: x.lvl, MaxLevel: 20}
		copy(s.ReqSkills[:], x.req)
		r.Skills[x.n] = s
	}
	return r
}
