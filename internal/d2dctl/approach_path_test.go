package d2dctl

import (
	"context"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A wall 3 cells wide (15 subtiles) between the player and the NPC: the blind 3.5-cell side step runs along it, the
// map route goes round and ends within standoff of the NPC.
func wallGrid() []string {
	rows := make([]string, 60)
	for y := range rows {
		r := []byte(strings.Repeat(".", 100))
		if y >= 20 && y < 60 {
			for x := 40; x < 55; x++ {
				r[x] = '#'
			}
		}
		rows[y] = string(r)
	}
	return rows
}

func TestPlanAroundWall(t *testing.T) {
	wps := planAround(wallGrid(), 6, 6, 14, 6, nil) // player (30,30) subtiles, NPC (70,30): behind the wall
	if len(wps) == 0 {
		t.Fatal("no route")
	}
	rows := wallGrid()
	for _, w := range wps {
		if rows[int(5*w[1])][int(5*w[0])] != '.' {
			t.Fatalf("waypoint %v inside the wall", w)
		}
	}
	last := wps[len(wps)-1]
	if d := (last[0]-14)*(last[0]-14) + (last[1]-6)*(last[1]-6); d > 2 {
		t.Fatalf("route ends at %v, not near the NPC", last)
	}
}

func TestPlanAroundAvoidsOtherNPC(t *testing.T) {
	rows := make([]string, 40)
	for i := range rows {
		rows[i] = strings.Repeat(".", 80)
	}
	// corridor 1 cell wide impossible here; just check no waypoint lies within the block radius of another NPC
	wps := planAround(rows, 2, 4, 14, 4, [][2]float64{{8, 4}})
	if len(wps) == 0 {
		t.Fatal("no route")
	}
	for _, w := range wps {
		if (w[0]-8)*(w[0]-8)+(w[1]-4)*(w[1]-4) < 0.64 {
			t.Fatalf("waypoint %v on the other NPC", w)
		}
	}
}

func TestPlanAroundNoRoute(t *testing.T) {
	rows := wallGrid()
	for y := range rows {
		rows[y] = rows[y][:40] + strings.Repeat("#", 15) + rows[y][55:]
	}
	if wps := planAround(rows, 6, 6, 14, 6, nil); len(wps) != 0 && wps[len(wps)-1][0] >= 8 {
		t.Fatalf("route through a closed wall: %v", wps)
	}
}

// ApproachNPC: a stall behind the wall is solved by the grid route, no blind side step needed.
func TestApproachNPCRoutesAroundTheWall(t *testing.T) {
	oldPause, oldNow := Pause, Now
	t.Cleanup(func() { Pause, Now = oldPause, oldNow })
	virt := time.Unix(0, 0)
	Pause = func(ctx context.Context, d time.Duration) error { virt = virt.Add(d); return ctx.Err() }
	Now = func() time.Time { return virt }
	x, y := 6.0, 6.0
	moves := 0
	grid := strings.Join(wallGrid(), "\n") + "\nok\n" // rows first, then the ok line (fixstall4: it was ok first, an empty grid)
	_, c := startStub(t, map[string]string{"npcs": "Kashya\t64\t-148\t1\t7\t14.0\t6.0\nok\n", "debug grid 5": grid}, func(line string, s *stub) {
		if strings.HasPrefix(line, "cmd move") {
			moves++
			f := strings.Fields(line)
			x, _ = strconv.ParseFloat(f[2], 64)
			y, _ = strconv.ParseFloat(f[3], 64)
		}
		s.rep["debug level"] = "ok 1 " + strconv.FormatFloat(x, 'f', 2, 64) + " " + strconv.FormatFloat(y, 'f', 2, 64) + " 20 12 0 0\n"
		if x > 12 && line == "cmd interact 7" {
			s.rep["menu"] = "Kashya\t265\t39\nbounty\t265\t72\nok\n"
		}
	})
	if err := c.ApproachNPC(context.Background(), 7, 15*time.Second); err != nil {
		t.Fatal(err)
	}
	if moves < 2 {
		t.Fatalf("expected several route legs, got %d moves", moves)
	}
}

// campGrid loads the REAL `debug grid 5` crop of the Rogue Encampment (DRILLS7, V live), x 60..160, y 60..120, into a
// full-size grid (everything outside the crop blocked).
func campGrid(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile("testdata/camp-grid.txt")
	if err != nil {
		t.Fatal(err)
	}
	const x0, y0 = 60, 60
	rows := make([][]byte, 121)
	for y := range rows {
		rows[y] = []byte(strings.Repeat("#", 161))
	}
	for _, ln := range strings.Split(string(b), "\n") {
		if ln == "" || ln[0] == '#' {
			continue
		}
		f := strings.SplitN(strings.TrimLeft(ln, " "), " ", 2)
		y, err := strconv.Atoi(f[0])
		if err != nil || len(f) < 2 {
			continue
		}
		copy(rows[y][x0:], strings.NewReplacer("P", ".", "K", ".", "A", ".").Replace(f[1])) // the markers sit on free cells
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = string(r)
	}
	return out
}

// F10 (DRILLS7): the three H.pause stall cells have no clearance-1 route; the adaptive planner finds one that ends
// within the standoff of Kashya (85,91 subtiles) on standable cells.
func TestPlanAroundRealCampStallCells(t *testing.T) {
	rows := campGrid(t)
	kx, ky := 85/5.0, 91/5.0
	for _, c := range [][2]float64{{25.9, 20.5}, {25.5, 22.5}, {26.4, 20.5}, {27.7, 20.6}} {
		wps := planAround(rows, c[0], c[1], kx, ky, nil)
		if len(wps) == 0 {
			t.Fatalf("no route from stall cell %v", c)
		}
		last := wps[len(wps)-1]
		if d := math.Hypot(5*last[0]-85, 5*last[1]-91); d > approachStandOff+1 {
			t.Errorf("from %v the route ends %.1f subtiles from Kashya", c, d)
		}
		if rows[int(5*last[1])][int(5*last[0])] != '.' {
			t.Errorf("from %v the route ends on a blocked cell", c)
		}
	}
	// clearance 1 still wins where it exists (the arrival cell)
	if planAroundClear(rows, 27.7, 20.6, kx, ky, nil, 1) == nil {
		t.Error("arrival cell lost its clearance-1 route")
	}
}
