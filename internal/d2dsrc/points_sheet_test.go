package d2dsrc

import (
	"context"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/points"
)

// TestSheetMatchesDumpFixture: the points sheet built from a snapshot read through the real parser equals what the
// recorded dump says, for two classes (Sorceress late game, Necromancer fresh): attributes, base split, unspent points and
// the class base from the class table. The base split comes from `dump stats` through d2dctl.BaseAttrs.
func TestSheetMatchesDumpFixture(t *testing.T) {
	classes := []*data.CharClass{
		{Name: "Sorceress", Code: "sor", Str: 10, Dex: 25, Vit: 10, Int: 35, LifePerVit: 8, ManaPerEnergy: 8},
		{Name: "Necromancer", Code: "nec", Str: 15, Dex: 25, Vit: 15, Int: 25, LifePerVit: 8, ManaPerEnergy: 8},
	}
	for _, c := range []struct {
		file              string
		class             d2.Class
		base, now         [4]int
		statPts, skillPts int
		waiting           string
	}{
		{"dump_head_late.json", d2.Sorceress, [4]int{20, 25, 30, 60}, [4]int{23, 25, 35, 70}, 5, 0, "5 stat points waiting for you"},
		{"dump_head_necro.json", d2.Necromancer, [4]int{15, 25, 15, 25}, [4]int{15, 25, 15, 25}, 0, 1, "1 skill point waiting for you"},
	} {
		r := lateReplies(t)
		r[headSections] = readFile(t, c.file) + "\nok\n"
		r["dump stats"] = readFile(t, c.file) + "\nok\n"
		_, cl := startScript(t, r)
		res, err := New(cl, Options{Tables: fakeTables{}, Logf: quiet}).Read(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		p := res.Snap.Player
		base, ok, err := cl.BaseAttrs(context.Background())
		if err != nil || !ok {
			t.Fatalf("%s: BaseAttrs %v %v", c.file, ok, err)
		}
		var cc *data.CharClass
		for _, x := range classes {
			if x.Name == c.class.String() {
				cc = x
			}
		}
		s := points.BuildSheet(points.Input{Hero: p.Name, Player: p, Class: cc, Base: &base, Enabled: true, Verbs: points.Verbs{Stat: true, Skill: true}})
		if s.StatPoints != c.statPts || s.SkillPoints != c.skillPts || s.Waiting != c.waiting {
			t.Errorf("%s: points %d/%d waiting %q", c.file, s.StatPoints, s.SkillPoints, s.Waiting)
		}
		for i, row := range s.Stats {
			if row.Base != c.base[i] || row.Now != c.now[i] {
				t.Errorf("%s %s: base %d now %d, dump says %d and %d", c.file, row.Key, row.Base, row.Now, c.base[i], c.now[i])
			}
			if want := c.base[i] - map[string]int{"str": cc.Str, "dex": cc.Dex, "vit": cc.Vit, "ene": cc.Int}[row.Key]; row.Spent != want || row.ClassBase+row.Spent != row.Base {
				t.Errorf("%s %s: spent %d (class base %d), want %d", c.file, row.Key, row.Spent, row.ClassBase, want)
			}
		}
	}
}
