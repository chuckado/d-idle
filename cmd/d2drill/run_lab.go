package main

import (
	"context"
	"fmt"
	"math"

	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
)

var _ drill.RunLab = (*lab)(nil)

// Leg is the fixrun proof leg: debug goto 2, no natural monsters, then a move command of 40 subtiles through the
// bot's controller (it presses the run key by its policy), stepped one tick at a time with the position and the raw
// stamina read each tick. The direction is the first of eight that gets the character 30 subtiles away; it is
// remembered so that every leg of a drill run uses the same one.
func (l *lab) Leg(ctx context.Context, r *drill.Run, name string) drill.RunLeg {
	leg := drill.RunLeg{Name: name, Dir: -1}
	ctl, c := l.be.Ctl, l.be.Client
	if err := l.Goto(ctx, r, 2); err != nil {
		leg.Err = "goto: " + err.Error()
		return leg
	}
	if _, err := r.Cheat(ctx, "debug kill", "no natural monster may interrupt the timed leg"); err != nil {
		leg.Err = "kill: " + err.Error()
		return leg
	}
	_ = l.Tick(ctx, 30)
	dirs := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if l.runDir >= 0 {
		dirs = []int{l.runDir}
	}
	fpos := func() (x, y float64, err error) {
		li, err := c.Level(ctx)
		return li.X * 5, li.Y * 5, err
	}
	for _, d := range dirs {
		sx, sy, err := fpos()
		if err != nil {
			leg.Err = "pos: " + err.Error()
			return leg
		}
		start, err := ctl.Pos()
		if err != nil {
			leg.Err = "pos: " + err.Error()
			return leg
		}
		a := float64(d) * math.Pi / 4
		tgt := input.Point{X: start.X + int(math.Round(40*math.Cos(a))), Y: start.Y + int(math.Round(40*math.Sin(a)))}
		leg = drill.RunLeg{Name: name, Dir: d}
		leg.StamStart = l.stam(ctx)
		if err := ctl.Move(tgt); err != nil {
			leg.Err = "move: " + err.Error()
			continue
		}
		px, py, idle := sx, sy, 0
		for t := 0; t < 160; t++ {
			if _, err := c.Step(ctx, 1); err != nil {
				leg.Err = "step: " + err.Error()
				break
			}
			leg.Ticks++
			cx, cy, err := fpos()
			if err != nil {
				leg.Err = "pos: " + err.Error()
				break
			}
			dl := math.Hypot(cx-px, cy-py)
			if dl > 0.01 {
				if leg.Moving == 0 {
					leg.Lag = leg.Ticks
				}
				leg.Path += dl
				leg.Moving++
				idle = 0
				if len(leg.Moves) < 12 {
					leg.Moves = append(leg.Moves, fmt.Sprintf("%.3f", dl))
				}
			} else if idle++; idle > 12 && leg.Moving > 0 {
				break
			}
			px, py = cx, cy
			if math.Hypot(float64(tgt.X)-cx, float64(tgt.Y)-cy) <= 2 {
				break
			}
		}
		leg.Net = math.Hypot(px-sx, py-sy)
		leg.StamEnd = l.stam(ctx)
		if leg.Net >= 30 {
			l.runDir = d
			leg.Err = ""
			break
		}
		if leg.Err == "" {
			leg.Err = fmt.Sprintf("direction %d: net %.1f subtiles only", d, leg.Net)
		}
		if err := l.Goto(ctx, r, 2); err != nil {
			leg.Err = "goto: " + err.Error()
			return leg
		}
	}
	_, _ = r.Cheat(ctx, "debug goto 1", "back to town after the timed leg")
	_ = l.Tick(ctx, 25)
	if leg.Moving > 0 {
		leg.PerTick = leg.Path / float64(leg.Moving)
	}
	ri := ctl.RunInfo()
	leg.Believed, leg.Keys, leg.Resets, leg.WalkingRaw = ri.Believed, ri.Presses, ri.Resets, ri.Walking
	return leg
}

func (l *lab) SetRunPolicy(on bool) {
	if on {
		l.be.Ctl.Run = l.runPol
		return
	}
	if l.be.Ctl.Run != nil {
		l.runPol = l.be.Ctl.Run
	}
	l.be.Ctl.Run = nil
}

func (l *lab) stam(ctx context.Context) string {
	st, _, err := l.be.Client.State(ctx)
	if err != nil {
		return "?"
	}
	p, ok := st.Pair("stamina")
	if !ok {
		return "none"
	}
	return fmt.Sprintf("%d/%d", p.Cur, p.Max)
}

// Restart leaves and enters the game as the cycle drill does (Cycler.LeaveGame and EnterGame; OnInGame resets the run belief).
func (l *lab) Restart(ctx context.Context) error {
	if err := l.be.Cycler.LeaveGame(ctx); err != nil {
		return fmt.Errorf("leave: %w", err)
	}
	if err := l.be.Cycler.EnterGame(ctx, ""); err != nil {
		return fmt.Errorf("enter: %w", err)
	}
	_, _ = l.be.Client.Step(ctx, 30)
	return nil
}
