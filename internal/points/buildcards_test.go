package points

import (
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/plandata"
)

func cardDesk(t *testing.T, class d2.Class, tables ...string) *Desk {
	t.Helper()
	cl := strings.ToLower(class.String())
	return NewDesk(Config{
		Hero: "Testy", Dir: t.TempDir(),
		Builds: func() []BuildInfo {
			var out []BuildInfo
			for _, n := range tables {
				out = append(out, BuildInfo{Name: n, Label: strings.ReplaceAll(n, "_", " "), Class: cl, Tier: BuildSupported})
			}
			return out
		},
		Player: func() *d2.Player { return &d2.Player{Class: class, Level: 12} },
	})
}

// One card per catalogue build, keyed by the catalogue id, with the real tables behind it as variants; the small card holds
// Defense and three other ratings, each with a text equivalent and the honest source word.
func TestCardsThreeClasses(t *testing.T) {
	for _, c := range []struct {
		class  d2.Class
		tables []string
		ids    []string
		first  int // variants of the first card
	}{
		{d2.Paladin, []string{"paladin-normal30", "paladin-normal30-v2", "paladin_hammerdin_guide", "paladin_hammerdin_endgame90", "paladin-zealot-lite"}, []string{"paladin_hammerdin", "paladin_zeal_start"}, 4},
		{d2.Sorceress, []string{"sorceress_normal30", "primary-blizzard"}, []string{"sorceress_blizzard", "sorceress_fire", "sorceress_lightning"}, 2},
		{d2.Necromancer, []string{"necromancer_summoner_normal30"}, []string{"necromancer_summoner", "necromancer_bone", "necromancer_poison_nova"}, 1},
	} {
		cards, err := cardDesk(t, c.class, c.tables...).Cards(c.class.String())
		if err != nil {
			t.Fatalf("%s: %v", c.class, err)
		}
		var got []string
		for _, k := range cards {
			got = append(got, k.ID)
			if len(k.Ratings) != 4 || k.Ratings[0].Key != plandata.RatingDefense {
				t.Errorf("%s %s: small card must be Defense plus three: %+v", c.class, k.ID, k.Ratings)
			}
			for _, r := range k.Ratings {
				if !strings.HasPrefix(r.Text, r.Label+": ") || !strings.Contains(r.Text, " of 5 stars, estimate") {
					t.Errorf("%s %s: rating text %q", c.class, k.ID, r.Text)
				}
			}
			if k.Variants == nil {
				t.Errorf("%s %s: variants must be a list, not null", c.class, k.ID)
			}
		}
		if strings.Join(got, ",") != strings.Join(c.ids, ",") {
			t.Errorf("%s cards %v, want %v", c.class, got, c.ids)
		}
		if n := len(cards[0].Variants); n != c.first {
			t.Errorf("%s: first card has %d variants, want %d", c.class, n, c.first)
		}
	}
}

func TestCardsUnknownClassIsAnError(t *testing.T) {
	d := cardDesk(t, d2.Paladin, "paladin-normal30")
	if _, err := d.Cards("Bard"); err == nil || !strings.Contains(err.Error(), "Bard") {
		t.Fatalf("unknown class must be an error naming it, got %v", err)
	}
}

func TestBuildDetail(t *testing.T) {
	d := cardDesk(t, d2.Paladin, "paladin-normal30", "paladin_hammerdin_guide")
	det, err := d.BuildDetail("paladin_hammerdin")
	if err != nil {
		t.Fatal(err)
	}
	if len(det.Ratings) != 7 || len(det.Pros) != 3 || len(det.Cons) != 3 || det.Summary == "" {
		t.Errorf("detail shape: %d ratings, %d pros, %d cons", len(det.Ratings), len(det.Pros), len(det.Cons))
	}
	if len(det.Variants) != 2 || len(det.Tables) != 4 {
		t.Errorf("variants %d (the two listed), tables %d (the catalogue's four)", len(det.Variants), len(det.Tables))
	}
	if _, err := d.BuildDetail("sorceress_fire"); err == nil || !strings.Contains(err.Error(), "sorceress_fire") {
		t.Errorf("another class must be refused by name: %v", err)
	}
	if _, err := d.BuildDetail("nope"); err == nil || !strings.Contains(err.Error(), "nope") {
		t.Errorf("unknown id must be an error naming it: %v", err)
	}
}

func TestBuildViewCarriesCards(t *testing.T) {
	v := cardDesk(t, d2.Sorceress, "sorceress_normal30").BuildView()
	if len(v.Cards) != 3 || v.CardsErr != "" || v.Class != "Sorceress" {
		t.Errorf("cards %d err %q class %q", len(v.Cards), v.CardsErr, v.Class)
	}
}
