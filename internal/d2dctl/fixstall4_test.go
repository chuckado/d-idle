package d2dctl

import (
	"context"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

// campWorld is a fake player on the REAL camp grid (DRILLS7 fixture): "cmd move" and "cmd interact" walk a straight line
// at about 2 cells/s and stop at the first blocked subtile, like d2d's verbs (V: docs/fixstall.md, stops at an
// object; DRILLS8 timeline for the speed). The position advances lazily with the virtual clock.
type campWorld struct {
	rows       []string
	virt       *time.Time
	last       time.Time
	x, y       float64
	tx, ty     float64
	moving     bool
	npcX, npcY float64
	dropMoves  int // the first n "cmd move" legs do nothing (the leg did not free the player)
	moves      int
}

func (w *campWorld) free(x, y float64) bool {
	cx, cy := int(5*x), int(5*y)
	return cy >= 0 && cy < len(w.rows) && cx >= 0 && cx < len(w.rows[cy]) && w.rows[cy][cx] == '.'
}

func (w *campWorld) advance() {
	dt := w.virt.Sub(w.last).Seconds()
	w.last = *w.virt
	for dist := 2 * dt; w.moving && dist > 0; dist -= 0.05 {
		d := math.Hypot(w.tx-w.x, w.ty-w.y)
		if d < 0.05 {
			w.moving = false
			return
		}
		nx, ny := w.x+(w.tx-w.x)/d*0.05, w.y+(w.ty-w.y)/d*0.05
		if !w.free(nx, ny) {
			w.moving = false
			return
		}
		w.x, w.y = nx, ny
	}
}

func campWorldApproach(t *testing.T, startX, startY float64, dropMoves int) (*campWorld, error, time.Duration) {
	return campWorldApproachFull(t, startX, startY, dropMoves, nil)
}

func campWorldApproachHook(t *testing.T, startX, startY float64, seen func(string, time.Duration)) (*campWorld, error, time.Duration) {
	return campWorldApproachFull(t, startX, startY, 0, seen)
}

func campWorldApproachFull(t *testing.T, startX, startY float64, dropMoves int, seen func(string, time.Duration)) (*campWorld, error, time.Duration) {
	oldPause, oldNow := Pause, Now
	t.Cleanup(func() { Pause, Now = oldPause, oldNow })
	virt := time.Unix(0, 0)
	Pause = func(ctx context.Context, d time.Duration) error { virt = virt.Add(d); return ctx.Err() }
	Now = func() time.Time { return virt }
	rows := campGrid(t)
	w := &campWorld{rows: rows, virt: &virt, last: virt, x: startX, y: startY, npcX: 17.1, npcY: 18.3, dropMoves: dropMoves}
	grid := strings.Join(rows, "\n") + "\nok\n" // rows first, then the ok line
	_, c := startStub(t, map[string]string{"npcs": "Kashya\t64\t-148\t1\t7\t17.10\t18.30\nok\n", "debug grid 5": grid}, func(line string, s *stub) {
		w.advance()
		if seen != nil {
			seen(line, virt.Sub(time.Unix(0, 0)))
		}
		switch {
		case strings.HasPrefix(line, "cmd move"):
			w.moves++
			if w.dropMoves > 0 {
				w.dropMoves--
				w.moving = false
				break
			}
			f := strings.Fields(line)
			w.tx, _ = strconv.ParseFloat(f[2], 64)
			w.ty, _ = strconv.ParseFloat(f[3], 64)
			w.moving = true
		case line == "cmd interact 7":
			w.tx, w.ty, w.moving = w.npcX, w.npcY, true
		}
		s.rep["debug level"] = "ok 1 " + strconv.FormatFloat(w.x, 'f', 2, 64) + " " + strconv.FormatFloat(w.y, 'f', 2, 64) + " 20 12 0 0\n"
		if math.Hypot(w.x-w.npcX, w.y-w.npcY) < 2.5 {
			s.rep["menu"] = "Kashya\t265\t39\nbounty\t265\t72\nok\n"
		}
	})
	err := c.ApproachNPC(context.Background(), 7, 15*time.Second)
	return w, err, virt.Sub(time.Unix(0, 0))
}

// F13 (DRILLS8): from the H.pause stall cells beside the camp block Kashya must be reached on the real grid.
func TestApproachNPCCampStallCells(t *testing.T) {
	for _, c := range [][2]float64{{26.4, 20.5}, {25.9, 20.5}, {25.5, 22.5}, {27.7, 20.6}} {
		_, err, el := campWorldApproach(t, c[0], c[1], 0)
		if err != nil {
			t.Errorf("from %v: %v", c, err)
		} else if el > 20*time.Second {
			t.Errorf("from %v: %v of game time, want within 20 s", c, el)
		}
	}
}

// The pocket cell (an isolated free subtile, DRILLS8 t681) has a route from its nearest free neighbour. The fake world
// cannot leave it (it would have to cross the blocked subtiles, which the real player did: UNVERIFIED how), so this
// is a plan-level check only.
func TestPlanAroundFromPocket(t *testing.T) {
	if wps := planAround(campGrid(t), 26.1, 20.1, 17.1, 18.3, nil); len(wps) < 2 {
		t.Fatalf("no route from the pocket cell: %v", wps)
	}
}

// The first leg of the first route does not free the player (dropped by the fake): the second route does, inside the
// budget, and the first route starts within about 2 s of the stall.
func TestApproachNPCSecondRouteFrees(t *testing.T) {
	w, err, el := campWorldApproach(t, 26.4, 20.5, 1)
	if err != nil {
		t.Fatalf("not reached (player at %.1f,%.1f): %v", w.x, w.y, err)
	}
	if el > 25*time.Second {
		t.Fatalf("%v of game time", el)
	}
}

// Every leg of a route is a straight line over free cells: the route may not cut a corner of an object between two
// waypoints (F13 first cause).
func TestPlanAroundLegsAreFreeLines(t *testing.T) {
	rows := campGrid(t)
	free := func(x, y int) bool { return y >= 0 && y < len(rows) && x >= 0 && x < len(rows[y]) && rows[y][x] == '.' }
	for _, c := range [][2]float64{{26.4, 20.5}, {25.9, 20.5}, {25.5, 22.5}, {27.7, 20.6}} {
		wps := planAround(rows, c[0], c[1], 17.1, 18.3, nil)
		px, py := c[0], c[1]
		for i, wp := range wps {
			a := [2]float64{5 * px, 5 * py}
			b := [2]float64{5 * wp[0], 5 * wp[1]}
			if !lineFree(free, a, b, 0) {
				t.Errorf("from %v leg %d %v to %v crosses blocked subtiles", c, i, a, b)
			}
			px, py = wp[0], wp[1]
		}
	}
}

// The error text reports the final position, not the previous poll's.
func TestApproachNPCErrorReportsFinalPosition(t *testing.T) {
	oldPause, oldNow := Pause, Now
	t.Cleanup(func() { Pause, Now = oldPause, oldNow })
	virt := time.Unix(0, 0)
	Pause = func(ctx context.Context, d time.Duration) error { virt = virt.Add(d); return ctx.Err() }
	Now = func() time.Time { return virt }
	_, c := startStub(t, map[string]string{"npcs": "Kashya\t64\t-148\t1\t7\t60.0\t20.0\nok\n"}, func(line string, s *stub) {
		el := virt.Sub(time.Unix(0, 0)).Seconds()
		s.rep["debug level"] = "ok 1 " + itoa(int(20+el/10)) + ".00 20.00 80 80 0 0\n"
	})
	err := c.ApproachNPC(context.Background(), 7, 2*time.Second)
	if err == nil {
		t.Fatal("expected a failure")
	}
	msg := err.Error()
	i := strings.Index(msg, "player at ")
	var px, py float64
	if i < 0 {
		t.Fatalf("no position in %q", msg)
	}
	px, _ = strconv.ParseFloat(strings.Split(strings.TrimPrefix(msg[i:], "player at "), ",")[0], 64)
	_ = py
	if want := 20 + math.Floor(virt.Sub(time.Unix(0, 0)).Seconds()/10); math.Abs(px-want) > 1 {
		t.Fatalf("error says x=%.1f, the player is at about %.1f: %q", px, want, msg)
	}
}

// A player that walks for a second and then stops at the camp block is routed about 1.5 s after it stopped, not after
// the rest of the 3 s round plus the next round's check (F13 second cause: the first route started late).
func TestApproachNPCStopMidRoundRoutesQuickly(t *testing.T) {
	var firstMove time.Duration = -1
	w, err, _ := campWorldApproachHook(t, 27.7, 20.6, func(line string, at time.Duration) {
		if strings.HasPrefix(line, "cmd move") && firstMove < 0 {
			firstMove = at
		}
	})
	if err != nil {
		t.Fatalf("not reached (player at %.1f,%.1f): %v", w.x, w.y, err)
	}
	if firstMove < 0 || firstMove > 2500*time.Millisecond {
		t.Fatalf("first route leg at %v of game time, want within 2.5 s (the player stops at about 0.7 s)", firstMove)
	}
}
