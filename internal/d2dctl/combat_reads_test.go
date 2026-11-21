package d2dctl

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl/fake"
)

// Two classes: a Paladin with a Holy Shield mage-less army of none, a Sorceress; a Necromancer's golem for the minions.
const minionsNecro = `{"tick":120,"hero":1,"minions":[{"id":41,"name":"claygolem","owner":1,"skill":75,"life":330,"maxlife":330,"x":12.5,"y":9.5,"mode":"NU","here":true,"ttl":-1,"visible":true,"hidden":["ttl"]},{"id":42,"name":"skeleton","owner":1,"skill":70,"life":40,"maxlife":61,"x":13.5,"y":9.5,"mode":"A1","here":true,"ttl":300,"visible":true,"hidden":["ttl"]}]}`

const skillStatePaladin = `{"tick":200,"class":3,"left":97,"right":112,"skills":[` +
	`{"id":0,"name":"Attack","base":1,"level":1,"mana":0,"cooldown":0,"usable":true,"left":false,"right":false,"charges":0,"status":0,"visible":["id","name","base","level","mana","cooldown","usable","left","right","charges"],"hidden":["status"]},` +
	`{"id":97,"name":"Smite","base":3,"level":3,"mana":2,"cooldown":0,"usable":true,"left":true,"right":false,"charges":0,"status":0,"visible":["id","name","base","level","mana","cooldown","usable","left","right","charges"],"hidden":["status"]},` +
	`{"id":112,"name":"Blessed Hammer","base":5,"level":5,"mana":9,"cooldown":0,"usable":false,"left":false,"right":true,"charges":0,"status":1,"visible":["id","name","base","level","mana","cooldown","usable","left","right","charges"],"hidden":["status"]}]}`

const skillStateSorc = `{"tick":210,"class":1,"left":36,"right":59,"skills":[` +
	`{"id":36,"name":"Fire Bolt","base":4,"level":4,"mana":3,"cooldown":0,"usable":true,"left":true,"right":false,"charges":0,"status":0,"visible":["id","name","base","level","mana","cooldown","usable","left","right","charges"],"hidden":["status"]},` +
	`{"id":59,"name":"Blizzard","base":1,"level":1,"mana":23,"cooldown":18,"usable":false,"left":false,"right":true,"charges":0,"status":8,"visible":["id","name","base","level","mana","cooldown","usable","left","right","charges"],"hidden":["status"]}]}`

const combatLogSorc = `{"tick":300,"total":4,"enabled":true,"events":[` +
	`{"seq":2,"tick":290,"kind":"cast","source":1,"target":7,"skill":36,"by_hero":true,"amount":3,"detail":4,"visible":["kind","source","target","skill","by_hero"],"hidden":["amount","detail"]},` +
	`{"seq":3,"tick":295,"kind":"hit","source":1,"target":7,"skill":36,"by_hero":true,"amount":14,"detail":0,"visible":["kind","source","target","skill","by_hero"],"hidden":["amount","detail"]},` +
	`{"seq":4,"tick":295,"kind":"kill","source":1,"target":7,"skill":36,"by_hero":true,"amount":0,"detail":1,"visible":["kind","source","target","skill","by_hero"],"hidden":["amount","detail"]}]}`

func readClient(t *testing.T, fix fake.Fixtures) *Client {
	t.Helper()
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	srv, err := fake.New(path, fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	c := New(path, Options{Timeout: 2 * time.Second})
	t.Cleanup(func() { c.Close() })
	return c
}

func TestMinionsNecromancer(t *testing.T) {
	fix := fake.Sample()
	fix.Minions = minionsNecro
	m, err := readClient(t, fix).Minions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m.Hero != 1 || len(m.Minions) != 2 {
		t.Fatalf("%+v", m)
	}
	g := m.Minions[0]
	if g.ID != 41 || g.Name != "claygolem" || g.Owner != m.Hero || g.Skill != 75 || g.Life != 330 || !g.Visible || g.TTL != -1 {
		t.Fatalf("golem %+v", g)
	}
	if !g.IsHidden("ttl") || g.IsHidden("life") {
		t.Fatalf("marks: ttl hidden, life visible: %+v", g.Hidden)
	}
	if s := m.Minions[1]; s.Mode != "A1" || s.TTL != 300 || s.Life >= s.MaxLife {
		t.Fatalf("skeleton %+v", s)
	}
}

func TestSkillStateTwoClasses(t *testing.T) {
	for _, tc := range []struct {
		name, json         string
		class, left, right int
		id                 int
		mana               int
		usable             bool
		cooldown           int
		hiddenStatus       int
	}{
		{"paladin", skillStatePaladin, 3, 97, 112, 112, 9, false, 0, 1},
		{"sorceress", skillStateSorc, 1, 36, 59, 59, 23, false, 18, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fix := fake.Sample()
			fix.SkillState = tc.json
			s, err := readClient(t, fix).SkillState(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if s.Class != tc.class || s.Left != tc.left || s.Right != tc.right {
				t.Fatalf("%+v", s)
			}
			row, ok := s.Skill(tc.id)
			if !ok || row.Mana != tc.mana || row.Usable != tc.usable || row.Cooldown != tc.cooldown || row.Status != tc.hiddenStatus || !row.Right {
				t.Fatalf("row %+v ok=%v", row, ok)
			}
			if !row.IsHidden("status") || row.IsHidden("mana") || row.IsHidden("cooldown") {
				t.Fatalf("marks: status hidden, mana and cooldown visible: %+v", row.Hidden)
			}
			if _, ok := s.Skill(9999); ok {
				t.Fatal("an unknown skill must not be found")
			}
		})
	}
}

func TestCombatLogKillAndDamage(t *testing.T) {
	fix := fake.Sample()
	fix.CombatLog = combatLogSorc
	l, err := readClient(t, fix).CombatLog(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if !l.Enabled || l.Total != 4 || len(l.Events) != 3 || l.Count("kill") != 1 || l.Count("hit") != 1 || l.Count("miss") != 0 {
		t.Fatalf("%+v", l)
	}
	if got := l.Since(3); len(got) != 1 || got[0].Kind != "kill" || !got[0].ByHero || got[0].Detail != 1 {
		t.Fatalf("since 3: %+v", got)
	}
	hit := l.Events[1]
	if hit.Amount != 14 || !hit.IsHidden("amount") || hit.IsHidden("kind") {
		t.Fatalf("hit marks %+v", hit)
	}
}

// A fork started without --read-combat (or stock d2d) answers err: the error names the verb and keeps the server's words; never an empty list.
func TestCombatReadsRefusedNamesTheVerb(t *testing.T) {
	c := readClient(t, fake.Sample())
	ctx := context.Background()
	if m, err := c.Minions(ctx); err == nil || m != nil || !errors.Is(err, ErrCombatRead) || !strings.Contains(err.Error(), "debug minions") || !strings.Contains(err.Error(), "--read-combat") {
		t.Fatalf("minions: %+v %v", m, err)
	}
	if s, err := c.SkillState(ctx); err == nil || s != nil || !strings.Contains(err.Error(), "debug skillstate") {
		t.Fatalf("skillstate: %+v %v", s, err)
	}
	if l, err := c.CombatLog(ctx, 0); err == nil || l != nil || !strings.Contains(err.Error(), "debug combatlog") {
		t.Fatalf("combatlog: %+v %v", l, err)
	}
}

func TestCombatReadsMalformedReply(t *testing.T) {
	for verb, parse := range map[string]func(*Reply) error{
		"minions":    func(r *Reply) error { _, err := ParseMinions(r); return err },
		"skillstate": func(r *Reply) error { _, err := ParseSkillState(r); return err },
		"combatlog":  func(r *Reply) error { _, err := ParseCombatLog(r); return err },
	} {
		for _, r := range []*Reply{nil, {}, {Lines: []string{"not json"}}} {
			err := parse(r)
			if err == nil || !errors.Is(err, ErrCombatRead) || !strings.Contains(err.Error(), "debug "+verb) {
				t.Fatalf("%s %+v: %v", verb, r, err)
			}
		}
	}
}
