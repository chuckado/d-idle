package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sort"
	"strings"

	"d2lod-bot/internal/d2map"
)

// stitchGroups lists the levels the game draws on one automap layer (Levels.txt "Layer" column, shared by these
// levels; game.exe keeps one cell store per layer, so every explored room of a layer is on the map whichever of its
// levels the player stands in). V: layer values read from Levels.txt, the per-layer store from the decompiled
// automap (docs/stream/map-contiguity.md). Every other level has a layer of its own, so its map stands alone.
var stitchGroups = [][]int{
	{1, 2, 3, 4, 5, 6, 7, 17, 26, 27, 28, 32, 33}, // act 1: town, the wilderness, Burial Grounds, Monastery to Cathedral
	{40, 41, 42, 43, 44, 45},                      // act 2: Lut Gholein and the desert
	{75, 76, 77, 78, 79, 80, 81, 82, 83},          // act 3: Kurast Docks to Travincal
	{103, 104, 105, 106},                          // act 4: Fortress, Outer Steppes, Plains of Despair, City of the Damned
	{107, 108},                                    // act 4: River of Flame, Chaos Sanctuary
	{109, 110, 111, 112},                          // act 5: Harrogath to Arreat Plateau
}

// stitchMembers is every level on the layer of level, in level order (just level itself when it stands alone).
func stitchMembers(level int) []int {
	for _, g := range stitchGroups {
		for _, n := range g {
			if n == level {
				return g
			}
		}
	}
	return []int{level}
}

// stitchID names the layer group of a level: the lowest level number of its members.
func stitchID(level int) int { return stitchMembers(level)[0] }

// maxStitchCells caps the composite (one pixel per subtile); beyond it only the first level is drawn.
const maxStitchCells = 4096 * 4096

// stitchPart is one level's grid and where it sits: Origin is the act-absolute subtile of the grid's frame (zero when
// the grid already holds absolute subtiles).
type stitchPart struct {
	No     int
	Origin d2map.Point
	Grid   *d2map.Grid
}

// stitchParts lists the known levels of the player's layer in level order (a fixed order, so a crossing does not change
// the picture: where two grids both know a cell, each reaches one tile past its edge, the lower level number wins).
// lookup returns nil for a level never visited.
func stitchParts(cur int, lookup func(int) *d2map.LevelMap) []stitchPart {
	var out []stitchPart
	for _, n := range stitchMembers(cur) {
		if lm := lookup(n); lm != nil && lm.Grid != nil && lm.Grid.KnownCount() > 0 {
			out = append(out, stitchPart{n, lm.Origin, lm.Grid})
		}
	}
	return out
}

// stitchKey changes whenever the composite would: a member appears, grows or moves.
func stitchKey(parts []stitchPart) string {
	var b strings.Builder
	ps := append([]stitchPart(nil), parts...)
	sort.Slice(ps, func(i, j int) bool { return ps[i].No < ps[j].No })
	for _, p := range ps {
		fmt.Fprintf(&b, "%d:%d:%d,%d;", p.No, p.Grid.Version(), p.Origin.X, p.Origin.Y)
	}
	return b.String()
}

// stitchBounds is the box of all parts in act-absolute subtiles (empty when none has cells).
func stitchBounds(parts []stitchPart) d2map.Rect {
	var x0, y0, x1, y1 int
	first := true
	for _, p := range parts {
		b := p.Grid.Bounds()
		if b.W <= 0 || b.H <= 0 {
			continue
		}
		ax, ay := p.Origin.X+b.X, p.Origin.Y+b.Y
		if first {
			x0, y0, x1, y1, first = ax, ay, ax+b.W, ay+b.H, false
			continue
		}
		x0, y0, x1, y1 = min(x0, ax), min(y0, ay), max(x1, ax+b.W), max(y1, ay+b.H)
	}
	if first {
		return d2map.Rect{}
	}
	return d2map.Rect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
}

var stitchPal = color.Palette{color.NRGBA{}, color.NRGBA{28, 32, 40, 255}, color.NRGBA{120, 130, 150, 255}}

// renderStitch draws the parts into one image in act-absolute subtiles and returns the PNG with its box. Known cells
// only: walkable dark grey, blocked light, unknown transparent; a cell known to an earlier (lower numbered) part is not overdrawn.
func renderStitch(parts []stitchPart) ([]byte, d2map.Rect) {
	r := stitchBounds(parts)
	if r.W*r.H > maxStitchCells && len(parts) > 1 {
		return renderStitch(parts[:1])
	}
	if r.W <= 0 || r.H <= 0 || r.W*r.H > maxStitchCells {
		return nil, d2map.Rect{}
	}
	img := image.NewPaletted(image.Rect(0, 0, r.W, r.H), stitchPal)
	for _, p := range parts {
		b := p.Grid.Bounds()
		for y := 0; y < b.H; y++ {
			for x := 0; x < b.W; x++ {
				fl, ok := p.Grid.Flags(b.X+x, b.Y+y)
				if !ok {
					continue
				}
				px, py := p.Origin.X+b.X+x-r.X, p.Origin.Y+b.Y+y-r.Y
				if img.ColorIndexAt(px, py) != 0 {
					continue
				}
				if fl&d2map.BlockWalk != 0 {
					img.SetColorIndex(px, py, 2)
				} else {
					img.SetColorIndex(px, py, 1)
				}
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes(), r
}
