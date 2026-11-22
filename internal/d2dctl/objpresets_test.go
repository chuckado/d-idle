package d2dctl

import "testing"

func TestParseLevelMonstersObjPresets(t *testing.T) {
	line := `{"level":5,"unit":"cell","rooms_up":[0],"presets":[],"monsters":[],"objpresets":[{"index":9,"class":119,"name":"Waypoint","kind":"waypoint","x":90.5,"y":70.5,"room":4,"built":false,"npc":-1}]}`
	lm, err := ParseLevelMonsters(&Reply{Lines: []string{line}})
	if err != nil || len(lm.ObjPresets) != 1 {
		t.Fatalf("%v %+v", err, lm)
	}
	if o := lm.ObjPresets[0]; o.Class != 119 || o.Kind != "waypoint" || o.X != 90.5 || o.Y != 70.5 || o.Room != 4 || o.Built || o.Npc != -1 {
		t.Fatalf("%+v", o)
	}
	// An older fork has no objpresets: the field is empty, not an error.
	old, err := ParseLevelMonsters(&Reply{Lines: []string{`{"level":5,"unit":"cell","rooms_up":[],"presets":[],"monsters":[]}`}})
	if err != nil || len(old.ObjPresets) != 0 {
		t.Fatalf("%v %+v", err, old)
	}
}
