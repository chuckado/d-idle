package main

import (
	"bytes"
	"image/png"
	"testing"

	"d2lod-bot/internal/d2map"
)

// lvl builds a level map whose grid is rows of ASCII at level-relative subtile (0,0), placed at an act tile origin.
func lvl(no, otx, oty int, rows ...string) *d2map.LevelMap {
	return &d2map.LevelMap{No: no, Origin: d2map.Point{X: 5 * otx, Y: 5 * oty}, Grid: d2map.ParseASCII(0, 0, rows...)}
}

type world map[int]*d2map.LevelMap

func (w world) look(no int) *d2map.LevelMap { return w[no] }

// at takes one poll in level no with the player at level-relative tile (x, y).
func at(f *Feed, w world, no int, x, y float64) Snap {
	s := Snap{X: x, Y: y, Level: no}
	f.stitch(&s, w[no], w.look)
	return s
}

func decode(t *testing.T, f *Feed) (func(x, y int) int, [4]int) {
	t.Helper()
	b, _, r := f.Map()
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return func(x, y int) int { // palette index of the pixel at act-absolute subtile (x, y)
		c := img.At(x-r[0], y-r[1])
		_, _, _, a := c.RGBA()
		if a == 0 {
			return 0
		}
		if r, _, _, _ := c.RGBA(); r>>8 > 100 {
			return 2
		}
		return 1
	}, r
}

func TestStitchGroupsAreDisjointAndAsInLevelsTxt(t *testing.T) {
	seen := map[int]int{}
	for gi, g := range stitchGroups {
		for _, n := range g {
			if o, dup := seen[n]; dup {
				t.Errorf("level %d in groups %d and %d", n, o, gi)
			}
			seen[n] = gi
		}
	}
	same := func(a, b int) bool { return stitchID(a) == stitchID(b) }
	// V: Levels.txt gives these one Layer
	for _, p := range [][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 6}, {6, 7}, {1, 17}, {7, 26}, {26, 27}, {27, 32}, {32, 33}, {40, 41}, {41, 45}, {75, 83}, {103, 106}, {107, 108}, {109, 112}} {
		if !same(p[0], p[1]) {
			t.Errorf("levels %d and %d should be stitched", p[0], p[1])
		}
	}
	// V: caves, crypts, towers, Tristram, the Act 2 sewers and tombs, and the act borders have a layer of their own
	for _, p := range [][2]int{{2, 8}, {3, 9}, {17, 18}, {17, 19}, {6, 20}, {1, 38}, {1, 39}, {40, 47}, {45, 55}, {106, 107}, {83, 100}, {112, 113}, {1, 40}} {
		if same(p[0], p[1]) {
			t.Errorf("levels %d and %d must not be stitched", p[0], p[1])
		}
	}
	if len(stitchMembers(8)) != 1 || stitchID(8) != 8 {
		t.Error("a lone level is its own group")
	}
}

// Blood Moor (2) and Cold Plains (3) touch: Cold Plains starts where Blood Moor ends (2 tiles = 10 subtiles).
func TestStitchTransitionBloodMoorColdPlains(t *testing.T) {
	w := world{
		2: lvl(2, 100, 200, "..........", "..........", ".........."),
		3: lvl(3, 102, 200, "##########", "##########", "##########"),
	}
	f := &Feed{}
	// Blood Moor first: only its own map.
	s := at(f, world{2: w[2]}, 2, 1.9, 1)
	if s.MapRect != [4]int{500, 1000, 10, 3} || s.Stitch != 1 {
		t.Fatalf("before: %+v", s)
	}
	// Step across the border: the act-frame position is continuous (0.2 tile), both maps are drawn, the box covers both.
	a := at(f, w, 2, 1.9, 1)
	b := at(f, w, 3, 0.1, 1)
	if d := b.X - a.X; d < 0.19 || d > 0.21 || b.Y != a.Y {
		t.Errorf("position jumped by %v,%v", d, b.Y-a.Y)
	}
	if b.MapRect != [4]int{500, 1000, 20, 3} || b.Stitch != a.Stitch {
		t.Errorf("after: %+v", b)
	}
	px, _ := decode(t, f)
	// the border cells: Blood Moor's last column (walkable) and Cold Plains' first (blocked) are neighbours
	if px(509, 1001) != 1 || px(510, 1001) != 2 {
		t.Errorf("border cells %d %d, want 1 2", px(509, 1001), px(510, 1001))
	}
	// no flicker: crossing back and forth again changes neither the map version nor its box
	v := b.MapVer
	for _, c := range []Snap{at(f, w, 2, 1.95, 1), at(f, w, 3, 0.05, 1), at(f, w, 3, 3, 1)} {
		if c.MapVer != v || c.MapRect != b.MapRect {
			t.Errorf("map changed on a crossing: %+v", c)
		}
	}
}

// Rogue Encampment (1) and Blood Moor (2): the town's west gate leads into the field on the same layer.
func TestStitchFieldAndTown(t *testing.T) {
	w := world{
		1: lvl(1, 110, 200, "##########", "##########"),
		2: lvl(2, 108, 200, "..........", ".........."),
	}
	f := &Feed{}
	s := at(f, w, 1, 0.5, 1)
	if s.MapRect != [4]int{540, 1000, 20, 2} || s.Stitch != stitchID(2) {
		t.Fatalf("town: %+v", s)
	}
	px, _ := decode(t, f)
	if px(549, 1000) != 1 || px(550, 1000) != 2 {
		t.Errorf("the field must end where the town starts: %d %d", px(549, 1000), px(550, 1000))
	}
}

// Blood Moor (2) and the Den of Evil (8): a cave is a map of its own and the field is not drawn in it.
func TestStitchFieldToCaveIsNotStitched(t *testing.T) {
	w := world{
		2: lvl(2, 100, 200, "..........", ".........."),
		8: lvl(8, 400, 700, "####", "#..#", "####"),
	}
	f := &Feed{}
	field := at(f, w, 2, 1, 1)
	cave := at(f, w, 8, 1, 1)
	if cave.Stitch == field.Stitch || cave.Stitch != 8 {
		t.Errorf("stitch ids field %d cave %d", field.Stitch, cave.Stitch)
	}
	if cave.MapRect != [4]int{2000, 3500, 4, 3} {
		t.Errorf("the cave pane shows more than the cave: %+v", cave.MapRect)
	}
	px, r := decode(t, f)
	if r[2] != 4 || px(2001, 3501) != 1 {
		t.Errorf("cave map %v", r)
	}
	// back in the field: only the field, the cave memory stays for the next visit
	back := at(f, w, 2, 1, 1)
	if back.MapRect != field.MapRect || back.Stitch != field.Stitch {
		t.Errorf("field map after the cave: %+v", back)
	}
	// a never-visited neighbour has no map yet and does not break the pane
	if s := at(&Feed{}, world{3: lvl(3, 102, 200, "..")}, 3, 0, 0); s.MapRect != [4]int{510, 1000, 2, 1} {
		t.Errorf("single level: %+v", s)
	}
}

func TestStitchOverlapIsStableAcrossACrossing(t *testing.T) {
	// each grid reaches one cell into the other (column 505): the lower level number wins, whoever the player is in
	w := world{
		2: lvl(2, 100, 200, ".....#", "......"),
		3: lvl(3, 101, 200, "#.....", "######"),
	}
	f := &Feed{}
	a := at(f, w, 3, 0, 0)
	px, _ := decode(t, f)
	if px(505, 1001) != 1 || px(506, 1001) != 2 {
		t.Errorf("shared column: %d, Cold Plains' wall %d", px(505, 1001), px(506, 1001))
	}
	if b := at(f, w, 2, 5, 0); b.MapVer != a.MapVer {
		t.Error("the picture was rebuilt by the crossing")
	}
}

func TestPubCarriesStitch(t *testing.T) {
	p := pubOf(Snap{OK: true, Stitch: 1, X: 120.5, Y: 200})
	if p.Stitch != 1 || p.X != 120.5 {
		t.Errorf("%+v", p)
	}
}
