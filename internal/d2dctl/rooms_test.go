package d2dctl

import (
	"os"
	"strings"
	"testing"
)

// presetsReply reads the "debug presets" transcript of an L0 fixture (docs/probes/cainl0/testdata).
func presetsReply(t *testing.T, name string) *Reply {
	t.Helper()
	b, err := os.ReadFile("../../docs/probes/cainl0/testdata/" + name + ".txt")
	if err != nil {
		t.Fatal(err)
	}
	r := &Reply{}
	in := false
	for _, l := range strings.Split(string(b), "\n") {
		switch {
		case strings.HasPrefix(l, "> "):
			in = l == "> debug presets"
		case in && !strings.HasPrefix(l, " err="):
			r.Lines = append(r.Lines, l)
		}
	}
	return r
}

func TestParseRoom(t *testing.T) {
	r, err := ParseRoom("room\t161\t44\t44")
	if err != nil || r != (Room{Type: 161, X: 44, Y: 44}) {
		t.Fatalf("%+v %v", r, err)
	}
	for _, bad := range []string{"", "room 1 2", "0\t30\t44.5\t44.5\t0\tTree", "room x 1 2"} {
		if _, err := ParseRoom(bad); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
	if Rooms(nil) != nil {
		t.Fatal("nil reply has rooms")
	}
}

// V (L0 fixtures, five seeds): room 161 origin + 0.5 is where the tree row stands, whenever the tree row is listed.
func TestRoom161IsTheTreeRoomInTheFixtures(t *testing.T) {
	found := 0
	for _, f := range []string{"s1", "sd3", "sd5", "sd7", "sd9", "tr3", "tr5", "tr7"} {
		r := presetsReply(t, f)
		var room *Room
		for _, rm := range Rooms(r) {
			if rm.Type == 161 {
				rm := rm
				room = &rm
			}
		}
		for _, l := range r.Lines {
			p, err := ParsePreset(l)
			if err != nil || p.ObjectID != 30 {
				continue
			}
			if room == nil || p.X != room.X+0.5 || p.Y != room.Y+0.5 {
				t.Errorf("%s: tree %v,%v vs room %+v", f, p.X, p.Y, room)
			}
			found++
		}
	}
	if found < 5 {
		t.Fatalf("only %d trees checked", found)
	}
}
