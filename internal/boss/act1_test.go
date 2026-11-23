package boss

import "testing"

// The Smith's anchor matches a read superunique preset by NAME (quests.layoutRead, `debug levelmonsters`): his table id
// is U (754 never appeared in the lab). Must fail without Name on the anchor; the script is class-free data.
func TestSmithAnchorResolvesByName(t *testing.T) {
	ref := smith().Anchor.Preset
	ps := []Preset{{Kind: PresetNPC, ID: 99999, Name: "TheSmith", Pos: Pt{X: 120, Y: 340}}}
	got, ok := resolvePreset(ref, ps)
	if !ok || got != (Pt{X: 120, Y: 340}) {
		t.Fatalf("resolved %v %v", got, ok)
	}
	if _, ok := resolvePreset(ref, []Preset{{Kind: PresetNPC, ID: 99999, Name: "Griswold", Pos: Pt{X: 1, Y: 1}}}); ok {
		t.Error("another name must not match")
	}
}
