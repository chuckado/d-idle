package points

import (
	"fmt"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/plandata"
)

// The build cards (card buildcards, review E6): one card per catalogue build of the hero's class, the plan tables that
// implement it behind it as variants. The text and ratings come from plandata's catalogue; the desk only joins them with
// the build files the wiring lists (Config.Builds), so a table the catalogue names but the leveler cannot load is not offered.

// Variant is a plan table behind a card: the name SetBuild takes, its label and tier. The detail view adds what it needs
// and the first levels.
type Variant struct {
	Name    string       `json:"name"`
	Label   string       `json:"label"`
	Tier    string       `json:"tier"`
	Needs   []Need       `json:"needs,omitempty"`
	Preview []PreviewRow `json:"preview,omitempty"`
}

// BuildCard is the small card: identity, the four ratings and the variants.
type BuildCard struct {
	plandata.CardView
	Variants []Variant `json:"variants"`
}

// BuildDetail is the detail view: all seven ratings, summary, pros and cons, and the variants with their previews.
type BuildDetail struct {
	plandata.DetailView
	Variants []Variant `json:"variants"`
}

func variantsOf(tables []string, have []BuildInfo, full bool) []Variant {
	out := []Variant{}
	for _, t := range tables {
		for _, b := range have {
			if b.Name != t {
				continue
			}
			v := Variant{Name: b.Name, Label: b.Label, Tier: b.Tier}
			if full {
				v.Needs, v.Preview = b.Needs, b.Preview
			}
			out = append(out, v)
		}
	}
	return out
}

// Cards lists the cards of a class. An unknown class is an error naming it; there is no fallback to another class.
func (d *Desk) Cards(class string) ([]BuildCard, error) {
	cat, err := plandata.Catalog()
	if err != nil {
		return nil, err
	}
	bs, err := cat.ForClass(class)
	if err != nil {
		return nil, err
	}
	have := d.buildsOf(class)
	out := []BuildCard{}
	for _, b := range bs {
		out = append(out, BuildCard{CardView: b.Card(), Variants: variantsOf(b.Tables, have, false)})
	}
	return out, nil
}

// BuildDetail is one card's detail, fetched when the player opens it. A build of another class is refused by name.
func (d *Desk) BuildDetail(id string) (BuildDetail, error) {
	var p *d2.Player
	if d.c.Player != nil {
		p = d.c.Player()
	}
	if p == nil {
		return BuildDetail{}, fmt.Errorf("the hero is not in a game yet")
	}
	cat, err := plandata.Catalog()
	if err != nil {
		return BuildDetail{}, err
	}
	b, err := cat.Build(id)
	if err != nil {
		return BuildDetail{}, err
	}
	if !strings.EqualFold(b.Class, p.Class.String()) {
		return BuildDetail{}, fmt.Errorf("build %q is for the %s, this hero is a %s", id, b.Class, p.Class.String())
	}
	return BuildDetail{DetailView: b.Detail(), Variants: variantsOf(b.Tables, d.buildsOf(p.Class.String()), true)}, nil
}
