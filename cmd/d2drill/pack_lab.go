package main

import (
	"context"
	"fmt"
	"strings"

	"d2lod-bot/internal/drill"
)

func (l *lab) PickitLog(ctx context.Context) (drill.PickInfo, []string, error) {
	mark := l.rec.Mark()
	info, err := l.Pickit(ctx)
	var lines []string
	for _, ln := range l.rec.Since(mark) {
		if strings.HasPrefix(ln, "loot:") {
			lines = append(lines, ln)
		}
	}
	return info, lines, err
}

func (l *lab) GroundCodes(ctx context.Context) ([]string, error) {
	g, err := l.be.Client.Ground(ctx)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, it := range g.Items {
		out = append(out, it.Code)
	}
	return out, nil
}

// FillPack gives and picks up items of code one at a time until the pack has no free 1x1 cell (at most n+20 gives).
// The engine's pickup also takes anything an earlier drill left on the ground (the pickit refuses plain items and the
// ground cannot be cleared), so the pack is judged by free cells, not by the count.
func (l *lab) FillPack(ctx context.Context, r *drill.Run, code string, n int) (int, error) {
	for i := 0; i < n+20; i++ {
		if free, err := l.PackFree(ctx, 1, 1); err == nil && !free {
			break
		}
		if err := l.GiveGround(ctx, r, code); err != nil {
			return i, err
		}
		if _, err := l.be.Client.Step(ctx, 20); err != nil {
			return i, err
		}
		if _, err := l.PickupRaw(ctx); err != nil {
			return i, err
		}
	}
	v, err := l.TView(ctx)
	if free, ferr := l.PackFree(ctx, 1, 1); ferr == nil && free {
		return -1, fmt.Errorf("the pack still has a free cell after %d gives", n+20)
	}
	return v.Pack[code], err
}

func (l *lab) PackFree(ctx context.Context, w, h int) (bool, error) {
	s, err := l.be.Env.Snapshot()
	if err != nil {
		return false, err
	}
	return l.tw.PackFits(s, w, h)
}

func (l *lab) MakeRoom(ctx context.Context, w, h int) (bool, string, error) {
	if err := l.OpenVendor(ctx, classGheed); err != nil {
		return false, "", err
	}
	mark := l.rec.Mark()
	ok, err := l.tw.MakeRoom(ctx, w, h, nil)
	var notes []string
	for _, ln := range l.rec.Since(mark) {
		if strings.Contains(ln, "make room") || strings.Contains(ln, "sell") {
			notes = append(notes, ln)
		}
	}
	_ = l.CloseAll(ctx)
	return ok, strings.Join(notes, " / "), err
}

func (l *lab) Transmute(ctx context.Context, codes []string) ([]string, error) {
	return l.tw.Transmute(ctx, codes)
}

func (l *lab) PurseNow(ctx context.Context) (int, error) {
	p, _, err := l.be.Client.Gold(ctx)
	return p, err
}

var _ drill.PackLab = (*lab)(nil)

const classGheed = 147
