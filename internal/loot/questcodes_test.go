package loot

import "testing"

// The Act 1 chain items were in no never-sell list (sandbox hero 2026-11-21): the data list must name them, and the Act 2
// and Act 3 to 5 lists must stay in it. Ordinary items are not quest items.
func TestQuestCodeCoversEveryAct(t *testing.T) {
	for _, c := range []string{"bks", "bkd", "hdm", "leg", "box", "msf", "vip", "hst", "tr1", "ass", "qf1", "mss", "hfh"} {
		if !QuestCode(c) {
			t.Errorf("%s is a quest item", c)
		}
	}
	for _, c := range []string{"", "tsc", "tbk", "cm1", "jew", "crs"} {
		if QuestCode(c) {
			t.Errorf("%q is no quest item", c)
		}
	}
}
