package questboard

import (
	"os"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/power"
	"d2lod-bot/internal/survival"
)

func field(area int, name string, mlvl, melee, ranged int, life, dmg float64) *danger.Zone {
	mon := func(reach string) danger.Monster {
		return danger.Monster{Key: reach, Share: 0.5, Life: int(life), Def: 30, ToHit: 60 + 5*mlvl, Dmg: dmg, Reach: reach}
	}
	return &danger.Zone{Area: area, Name: name, Diff: "normal", MLvl: mlvl, Monsters: []danger.Monster{mon("melee"), mon("ranged")},
		Encounters: danger.Encounters{
			Typical: danger.Encounter{N: melee + ranged, Melee: melee, Ranged: ranged},
			Bad:     danger.Encounter{N: 2 * (melee + ranged), Melee: 2 * melee, Ranged: 2 * ranged},
			Elite:   danger.Encounter{N: melee + ranged + 1, Melee: melee + 1, Ranged: ranged, Leader: "unique"}}}
}

func measured(h power.Hero) power.Hero {
	for _, term := range []string{power.TermLife, power.TermDefence, power.TermBlock, power.TermResist, power.TermOffence, power.TermPotions, power.TermMerc} {
		h.SetQuality(term, power.Measured)
	}
	return h
}

func paladin(lvl int, merc bool) power.Hero {
	return measured(power.Hero{Class: "Paladin", Level: lvl, MaxLife: 120 + 12*lvl, MaxMana: 40, Defence: 60 + 8*lvl, BlockPct: 30, AR: 60 + 12*lvl,
		DmgMin: 3 + lvl, DmgMax: 8 + 2*lvl, DmgPct: 20, Potions: 6, Resist: [4]int{10, 10, 10, 10}, Merc: power.Merc{Hired: merc, Alive: merc, Level: lvl}})
}

func sorceress(lvl int, merc bool) power.Hero {
	return measured(power.Hero{Class: "Sorceress", Level: lvl, MaxLife: 60 + 9*lvl, MaxMana: 60 + 10*lvl, Defence: 20, MeasuredDPS: 3 + 1.5*float64(lvl), Potions: 6,
		Merc: power.Merc{Hired: merc, Alive: merc, Level: lvl}})
}

// The zones are made up (as in internal/survival tests; the real tables need the install). Black Marsh is sized so a
// clvl 12 Paladin without a merc lands between the stay and enter lines, the logged outcome of
// docs/survival/README.md (a clvl 12 Paladin with a living merc was fine; the deaths came with pack size and
// no merc). The Sorceress rows are regression rows of the same zones, not logged outcomes.
var (
	coldPlains = field(2, "Cold Plains", 5, 2, 1, 40, 8)
	blackMarsh = field(4, "Black Marsh", 9, 4, 1, 60, 15)
)

func thresholds(t *testing.T) *Thresholds {
	t.Helper()
	th, err := LoadThresholds()
	if err != nil {
		t.Fatal(err)
	}
	return th
}

func zoneReady(t *testing.T, th *Thresholds, b *Board, class, arch, dial string, h power.Hero, z *danger.Zone, k float64) Readiness {
	t.Helper()
	m, err := power.NewModel(danger.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	a, err := m.P.Archetype(arch)
	if err != nil {
		t.Fatal(err)
	}
	zm, err := m.Zone(h, a, z, dial)
	if err != nil {
		t.Fatal(err)
	}
	r, err := th.Assess(AssessInput{Dial: dial, Hero: HeroFrom(h), Encounters: []Encounter{EncounterFrom(z.Name, "zone", zm.Bad, k)}})
	if err != nil {
		t.Fatal(err)
	}
	if string(r.Word) != strings.ReplaceAll(zm.Word(), " ", "_") && k == 1 {
		t.Fatalf("%s: questboard word %q differs from power.ZoneMargin.Word %q", z.Name, r.Word, zm.Word())
	}
	return r
}

func TestWordsMatchLoggedOutcomesOnTwoClasses(t *testing.T) {
	th := thresholds(t)
	cases := []struct {
		name  string
		class string
		arch  string
		mk    func(int, bool) power.Hero
		zone  *danger.Zone
		lvl   int
		merc  bool
		want  Word
	}{
		// Paladin: logged outcomes (V docs/survival/README.md; Cold Plains 1.5 s per kill, no potions)
		{"paladin cold plains clvl 4", "paladin", "bruiser", paladin, coldPlains, 4, true, ReadySafe},
		{"paladin cold plains clvl 5", "paladin", "bruiser", paladin, coldPlains, 5, true, ReadySafe},
		{"paladin black marsh clvl 4", "paladin", "bruiser", paladin, blackMarsh, 4, true, ReadyTooHard},
		{"paladin black marsh clvl 12 merc", "paladin", "bruiser", paladin, blackMarsh, 12, true, ReadySafe},
		{"paladin black marsh clvl 12 no merc", "paladin", "bruiser", paladin, blackMarsh, 12, false, ReadyRisky},
		// Sorceress: regression rows of the same made-up zones
		{"sorceress cold plains clvl 5", "sorceress", "caster", sorceress, coldPlains, 5, true, ReadyTooHard},
		{"sorceress cold plains clvl 12", "sorceress", "caster", sorceress, coldPlains, 12, true, ReadySafe},
		{"sorceress black marsh clvl 12 merc", "sorceress", "caster", sorceress, blackMarsh, 12, true, ReadyTooHard},
		{"sorceress black marsh clvl 12 no merc", "sorceress", "caster", sorceress, blackMarsh, 12, false, ReadyTooHard},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := zoneReady(t, th, nil, c.class, c.arch, "balanced", c.mk(c.lvl, c.merc), c.zone, 1)
			if r.Word != c.want || r.Unknown || r.Sentence == "" {
				t.Fatalf("word %q (want %q), unknown %v, sentence %q", r.Word, c.want, r.Unknown, r.Sentence)
			}
			if !c.merc && !hasReason(r, ReasonNoMerc) {
				t.Errorf("no merc reason in %+v", r.Reasons)
			}
			if strings.ContainsAny(r.Sentence, "0123456789") && !hasReason(r, ReasonLevelUnder) {
				t.Errorf("a number in the sentence without the legacy level reason: %q", r.Sentence)
			}
		})
	}
}

func hasReason(r Readiness, code string) bool {
	for _, x := range r.Reasons {
		if x.Code == code {
			return true
		}
	}
	return false
}

func TestDialsShiftTheWord(t *testing.T) {
	th := thresholds(t)
	// a margin of 1.3 (low = best): bold says safe (enter 1.2), balanced risky, cautious risky (stay 1.6 -> too hard)
	e := []Encounter{{Name: "Blood Raven", Kind: "boss", Best: 1.3, Low: 1.3, Known: true}}
	for dial, want := range map[string]Word{"bold": ReadySafe, "balanced": ReadyRisky, "cautious": ReadyTooHard} {
		r, err := th.Assess(AssessInput{Dial: dial, Hero: Hero{MercHired: true, MercAlive: true, Potions: 6}, Encounters: e})
		if err != nil || r.Word != want {
			t.Errorf("%s: %q %v, want %q", dial, r.Word, err, want)
		}
	}
	if _, err := th.Assess(AssessInput{Dial: "reckless"}); err == nil || !strings.Contains(err.Error(), `"reckless"`) {
		t.Errorf("unknown dial error = %v", err)
	}
}

func TestDialsMirrorThePowerProfiles(t *testing.T) {
	th := thresholds(t)
	p, err := power.Load()
	if err != nil {
		t.Fatal(err)
	}
	for id, l := range th.Dials {
		d, err := p.Dial(id)
		if err != nil || d.Enter != l.Enter || d.Stay != l.Stay {
			t.Errorf("dial %s: data %+v, power %+v %v", id, l, d, err)
		}
	}
}

func TestLedgerCorrectionMakesBlackMarshHarder(t *testing.T) {
	th := thresholds(t)
	sp, err := survival.LoadParams()
	if err != nil {
		t.Fatal(err)
	}
	l := survival.NewLedger(sp)
	key := survival.ZoneKey{Area: blackMarsh.Area, Diff: blackMarsh.Diff}
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	h := paladin(12, true)
	if w := zoneReady(t, th, nil, "paladin", "bruiser", "balanced", h, blackMarsh, 1).Word; w != ReadySafe {
		t.Fatalf("before the ledger: %q", w)
	}
	for i := 0; i < 2; i++ { // two merc deaths in the zone (V docs/survival/ledger.md: the second quote is refused)
		l.Observe(key, now, survival.Event{Kind: survival.EvMercDeath}, survival.PowerOf(h))
	}
	l.Observe(key, now, survival.Event{Kind: survival.EvDeath}, survival.PowerOf(h))
	k := l.KZoneFor(key, now, false) * l.KHeroFor(false)
	if k <= 1 {
		t.Fatalf("k = %v, want a correction above 1 after deaths", k)
	}
	if w := zoneReady(t, th, nil, "paladin", "bruiser", "balanced", h, blackMarsh, k).Word; w == ReadySafe {
		t.Fatalf("after merc deaths and a death (k %.2f) the word is still safe", k)
	}
}

func TestLegacyFallbackWarnsNeverWalls(t *testing.T) {
	th := thresholds(t)
	b, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	q, _ := b.Quest("sisters_burial_grounds")
	for _, c := range []struct {
		class string
		lvl   int
		want  Word
	}{{"paladin", 7, ReadySafe}, {"paladin", 1, ReadyRisky}, {"sorceress", 8, ReadySafe}, {"sorceress", 2, ReadyRisky}} {
		floor, ok, err := b.Floor(q, c.class)
		if err != nil || !ok {
			t.Fatal(c.class, err)
		}
		r, err := th.Assess(AssessInput{Hero: Hero{Class: c.class, Level: c.lvl}, Floor: floor, HasFloor: ok})
		if err != nil || r.Word != c.want || !r.FromLegacy || r.Unknown || r.LegacyFloor != floor {
			t.Fatalf("%s clvl %d: %+v %v", c.class, c.lvl, r, err)
		}
		if r.Word == ReadyTooHard {
			t.Fatal("the legacy floor must never wall")
		}
		if !hasReason(r, ReasonLegacyOnly) {
			t.Errorf("no legacy_only reason: %+v", r.Reasons)
		}
	}
}

func TestSentenceWithLegacyFloorBesideMargin(t *testing.T) {
	th := thresholds(t)
	r, err := th.Assess(AssessInput{
		Hero:       Hero{Class: "paladin", Level: 15, Potions: 6},
		Encounters: []Encounter{{Name: "Andariel", Kind: "boss", Best: 1.3, Low: 1.3, Known: true}},
		Floor:      17, HasFloor: true})
	if err != nil {
		t.Fatal(err)
	}
	want := "You have no merc, are two levels under the recommended 17 and face a close fight against Andariel: risky"
	if r.Sentence != want || r.LegacyFloor != 17 || r.Word != ReadyRisky {
		t.Fatalf("sentence %q\nwant     %q\n%+v", r.Sentence, want, r)
	}
	if r.Debug == nil || r.Debug.Best != 1.3 {
		t.Fatalf("debug = %+v", r.Debug)
	}
}

func TestUnknownReadinessKeepsQuestAvailable(t *testing.T) {
	th := thresholds(t)
	r, err := th.Assess(AssessInput{Hero: Hero{Class: "paladin", Level: 30}, Encounters: []Encounter{{Name: "x", Known: false}}})
	if err != nil || !r.Unknown || r.Word != ReadyUnknown || !hasReason(r, ReasonNoData) {
		t.Fatalf("%+v %v", r, err)
	}
	b, _ := Load()
	in := Input{Facts: facts{}, Ready: map[string]Word{ChunkKey("den_of_evil", "clear_den"): r.Word}}
	views, err := Derive(b, in)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range views {
		if v.ID == "den_of_evil" && (v.State != Available || len(v.Offered) != 0) {
			t.Fatalf("den state %s offered %v, want available and not offered", v.State, v.Offered)
		}
	}
}

func TestOffenceUnreadAndKitReasons(t *testing.T) {
	th := thresholds(t)
	r, err := th.Assess(AssessInput{Diff: 1, Hero: Hero{MercHired: true, MercAlive: true, Potions: 2, LowestResist: -5, ResistKnown: true},
		Encounters: []Encounter{{Name: "Countess", Kind: "boss", Best: 3, Low: 1, Known: true, OffenceUnread: true}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{ReasonKitPotions, ReasonKitResist, ReasonOffenceUnread, ReasonMarginLow} {
		if !hasReason(r, code) {
			t.Errorf("missing %s in %+v", code, r.Reasons)
		}
	}
}

func TestReadinessJSONIsStrict(t *testing.T) {
	raw, err := os.ReadFile("../../data/questboard/readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseThresholds([]byte(strings.Replace(string(raw), `"minPotions": 4`, `"minPotions": 4, "bogus": 1`, 1))); err == nil {
		t.Error("an unknown field must be an error")
	}
	if _, err := ParseThresholds([]byte(strings.Replace(string(raw), `"bold": {"enter": 1.2, "stay": 1.0}`, `"bold": {"enter": 0.5, "stay": 1.0}`, 1))); err == nil {
		t.Error("enter below stay must be an error")
	}
}
