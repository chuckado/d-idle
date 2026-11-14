package danger

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/data"
)

// fixtureDB is a tiny stand-in for the extracted tables; the numbers are made up (real values: real_test.go).
func fixtureDB() *data.DB {
	mon := func(key string, ranged bool, grp int, boss bool) *data.Monster {
		m := &data.Monster{Key: key, Enabled: true, Killable: true, Ranged: ranged, Boss: boss, MinGrp: grp, MaxGrp: grp + 2, Run: 6}
		for d := 0; d < 3; d++ {
			m.HPMin[d], m.HPMax[d] = 10*(d+1), 20*(d+1)
			m.AC[d] = 20 * (d + 1)
			m.A1Min[d], m.A1Max[d], m.A1TH[d] = 4*(d+1), 8*(d+1), 60*(d+1)
			m.Exp[d] = 30 * (d + 1)
			m.Level[d] = []int{5, 40, 70}[d]
		}
		m.Res[0][data.Cold] = 100
		return m
	}
	scale := &data.MonLevel{HP: [3]int{100, 100, 100}, DM: [3]int{100, 100, 100}, TH: [3]int{100, 100, 100}, AC: [3]int{100, 100, 100}, XP: [3]int{100, 100, 100}}
	ml := make([]*data.MonLevel, 90)
	for i := range ml {
		ml[i] = scale
	}
	lv := func(id int, name string, mlvl [3]int, pool ...string) *data.Level {
		l := &data.Level{ID: id, DrlgType: 3, Key: name, Name: name, MonLvl: mlvl, Mon: pool, NMon: pool, MonDen: [3]int{1000, 1000, 1000},
			MonUMin: [3]int{1, 1, 1}, MonUMax: [3]int{3, 3, 3}}
		l.SizeX, l.SizeY = [3]int{100, 100, 100}, [3]int{100, 100, 100}
		return l
	}
	boss := mon("andariel", false, 1, true)
	smithBase := mon("smith", false, 1, true)
	return &data.DB{
		Monsters:     []*data.Monster{mon("goatman", false, 3, false), mon("archer", true, 1, false), boss, smithBase},
		MonLevels:    ml,
		SuperUniques: []*data.SuperUnique{{Key: "The Smith", Name: "The Smith", Class: "smith"}},
		Levels: []*data.Level{
			lv(6, "Black Marsh", [3]int{6, 41, 71}, "goatman", "archer"),
			lv(3, "Cold Plains", [3]int{2, 37, 68}, "goatman"),
			{ID: 1, Key: "Town", Name: "Town"},
		},
	}
}

const testBosses = `{"schema":"danger.bosses/1","evidence":"docs/research/boss1-trace.md","bosses":[
 {"key":"smith","name":"The Smith","superunique":"The Smith","area":6,"mods":["extra_strong"],"dmgPct":235},
 {"key":"andariel","name":"Andariel","monster":"andariel","area":3}]}`

func fixtureTable(t *testing.T, ov *Overrides) *Table {
	t.Helper()
	tb, err := build(fixtureDB(), DefaultParams(), ov, []byte(testBosses))
	if err != nil {
		t.Fatal(err)
	}
	return tb
}

func TestTwoDifficulties(t *testing.T) {
	tb := fixtureTable(t, nil)
	n, err := tb.Find(6, "normal")
	if err != nil {
		t.Fatal(err)
	}
	h, err := tb.Find(6, "hell")
	if err != nil {
		t.Fatal(err)
	}
	if n.MLvl != 6 || h.MLvl != 71 {
		t.Fatalf("mlvl %d / %d", n.MLvl, h.MLvl)
	}
	if h.Monsters[0].Life <= n.Monsters[0].Life || h.Monsters[0].Dps <= n.Monsters[0].Dps {
		t.Errorf("hell must be harder: %+v vs %+v", h.Monsters[0], n.Monsters[0])
	}
	if len(n.Immune) != 1 || n.Immune[0] != "cold" || len(h.Immune) != 0 {
		t.Errorf("immune normal %v hell %v", n.Immune, h.Immune)
	}
	if _, err := tb.Find(1, "normal"); err == nil {
		t.Error("town has no profile")
	}
	var share float64
	for _, m := range n.Monsters {
		share += m.Share
	}
	if share < 0.999 || share > 1.001 {
		t.Errorf("shares sum %v", share)
	}
	if n.Monsters[1].Reach != "ranged" || n.Encounters.Bad.Ranged == 0 || n.Encounters.Bad.N <= n.Encounters.Typical.N {
		t.Errorf("encounters %+v monsters %+v", n.Encounters, n.Monsters)
	}
	c, _ := tb.Find(3, "normal")
	if c.Encounters.Bad.Ranged != 0 || c.Pay.XPByClvl["c1"] <= c.Pay.XPByClvl["c16"] {
		t.Errorf("cold plains %+v", c)
	}
}

func TestBosses(t *testing.T) {
	tb := fixtureTable(t, nil)
	s, err := tb.FindBoss("smith", "normal")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := tb.FindBoss("andariel", "normal")
	if s.Kind != "superunique" || a.Kind != "act" || s.Life != 4*a.Life || s.XP != 5*a.XP || s.Level != a.Level+3 {
		t.Errorf("smith %+v andariel %+v", s, a)
	}
	if s.Dmg <= a.Dmg*2 {
		t.Errorf("extra strong: %v vs %v", s.Dmg, a.Dmg)
	}
	z, _ := tb.Find(6, "normal")
	if len(z.Bosses) != 1 || z.Bosses[0] != "smith" {
		t.Errorf("zone bosses %v", z.Bosses)
	}
	bad := strings.Replace(testBosses, `"andariel","name"`, `"nobody","name"`, 1)
	bad = strings.Replace(bad, `"monster":"andariel"`, `"monster":"nobody"`, 1)
	if _, err := build(fixtureDB(), DefaultParams(), nil, []byte(bad)); err == nil || !strings.Contains(err.Error(), "nobody") {
		t.Errorf("unknown boss monster must error naming it, got %v", err)
	}
}

func TestOverrides(t *testing.T) {
	row := `{"id":"bm-pack","area":6,"diff":"normal","path":"encounters.bad.n","value":25,"evidence":"docs/research/areamon-trace.md","tag":"V","expires":"2027-01-01"}`
	ov, err := ParseOverrides([]byte(`{"schema":"danger.overrides/1","rows":[` + row + `]}`))
	if err != nil {
		t.Fatal(err)
	}
	ov.Now = func() time.Time { return time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC) }
	tb := fixtureTable(t, ov)
	z, _ := tb.Find(6, "normal")
	if z.Encounters.Bad.N != 25 || len(tb.Overrides) != 1 || z.Source["override:bm-pack"] != "V" {
		t.Errorf("override not applied: %+v %v", z.Encounters.Bad, tb.Overrides)
	}
	if err := ov.CheckEvidence("../.."); err != nil {
		t.Error(err)
	}
	ov.Rows[0].Evidence = "docs/research/does-not-exist.md"
	if err := ov.CheckEvidence("../.."); err == nil {
		t.Error("missing evidence file must fail")
	}
	ov.Now = func() time.Time { return time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC) }
	if len(ov.Expired()) != 1 {
		t.Error("row should be expired")
	}
	if tb := fixtureTable(t, ov); len(tb.Overrides) != 0 {
		t.Error("expired row applied")
	}
}

func TestOverrideRules(t *testing.T) {
	head := `{"schema":"danger.overrides/1","rows":[`
	cases := map[string]string{
		"no evidence": `{"id":"a","area":6,"diff":"normal","path":"density.units","value":1,"tag":"V","expires":"2027-01-01"}`,
		"bad tag":     `{"id":"a","area":6,"diff":"normal","path":"density.units","value":1,"evidence":"x","tag":"Z","expires":"2027-01-01"}`,
		"bad expiry":  `{"id":"a","area":6,"diff":"normal","path":"density.units","value":1,"evidence":"x","tag":"V","expires":"soon"}`,
		"bad diff":    `{"id":"a","area":6,"diff":"easy","path":"density.units","value":1,"evidence":"x","tag":"V","expires":"2027-01-01"}`,
		"unknown key": `{"id":"a","area":6,"diff":"normal","path":"density.units","value":1,"evidence":"x","tag":"V","expires":"2027-01-01","why":"y"}`,
	}
	for name, row := range cases {
		if _, err := ParseOverrides([]byte(head + row + `]}`)); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	var rows []string
	for i := 0; i < MaxPerClass+1; i++ {
		rows = append(rows, strings.Replace(`{"id":"rN","area":6,"diff":"normal","class":"paladin","path":"density.units","value":1,"evidence":"x","tag":"V","expires":"2027-01-01"}`, "rN", "r"+string(rune('a'+i)), 1))
	}
	if _, err := ParseOverrides([]byte(head + strings.Join(rows, ",") + `]}`)); err == nil {
		t.Error("six rows for one class must fail")
	}
	// a row that matches no area is an error, not a silent no-op
	ov, _ := ParseOverrides([]byte(head + `{"id":"a","area":999,"diff":"normal","path":"density.units","value":1,"evidence":"x","tag":"V","expires":"2999-01-01"}]}`))
	if _, err := build(fixtureDB(), DefaultParams(), ov, []byte(testBosses)); err == nil {
		t.Error("override for an unknown area must error")
	}
}

func TestShippedOverrides(t *testing.T) {
	ov, err := ShippedOverrides()
	if err != nil {
		t.Fatal(err)
	}
	if err := ov.CheckEvidence("../.."); err != nil {
		t.Error(err)
	}
	if e := ov.Expired(); len(e) > 0 {
		t.Errorf("expired override rows need review: %v", e)
	}
}

func TestHandler(t *testing.T) {
	h := Handler(fixtureTable(t, nil))
	get := func(url string) View {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
		var v View
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	if v := get(Route); len(v.Zones) != 2 {
		t.Errorf("list %+v", v)
	}
	if v := get(Route + "?area=Black+Marsh&difficulty=hell"); v.Zone == nil || v.Zone.Diff != "hell" || len(v.Bosses) != 1 {
		t.Errorf("zone %+v", v)
	}
	if v := get(Route + "?area=nowhere"); v.Error == "" {
		t.Error("unknown area must say so")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", Route, nil))
	if rec.Code != 405 {
		t.Errorf("POST %d", rec.Code)
	}
}

func TestBuildErrors(t *testing.T) {
	if _, err := Build(nil, DefaultParams(), nil); err == nil {
		t.Error("nil db")
	}
	if _, err := Build(fixtureDB(), Params{}, nil); err == nil {
		t.Error("empty params")
	}
}

func TestUnknownLevelType(t *testing.T) {
	db := fixtureDB()
	db.Levels[0].DrlgType = 9
	if _, err := build(db, DefaultParams(), nil, []byte(testBosses)); err == nil || !strings.Contains(err.Error(), "level type 9") {
		t.Errorf("unknown level type must error naming it, got %v", err)
	}
}

func TestShippedBossRows(t *testing.T) {
	var f bossFile
	if err := json.Unmarshal(bossesJSON, &f); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, b := range f.Bosses {
		if b.Key == "" || seen[b.Key] || (b.Monster == "") == (b.SuperUniq == "") || b.Area == 0 {
			t.Errorf("bad boss row %+v", b)
		}
		seen[b.Key] = true
	}
	if len(f.Bosses) != 10 {
		t.Errorf("%d boss rows, boss1-trace lists 10", len(f.Bosses))
	}
}
