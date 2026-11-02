package d2dsoakstat

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DefaultBands are the clvl bands of the Gate 1 XP clause (spec soak-gates, card baselinefree).
const DefaultBands = "1-4,5-7,8-10,11-14,15-20"

// Band is an inclusive clvl range.
type Band struct{ Lo, Hi int }

// String is the band key used in JSON and reports, "LO-HI".
func (b Band) String() string { return fmt.Sprintf("%d-%d", b.Lo, b.Hi) }

// ParseBands reads "1-4,5-7,..." (inclusive, ascending, no overlap).
func ParseBands(s string) ([]Band, error) {
	var out []Band
	for _, p := range strings.Split(s, ",") {
		lo, hi, ok := strings.Cut(strings.TrimSpace(p), "-")
		a, e1 := strconv.Atoi(lo)
		z, e2 := strconv.Atoi(hi)
		if !ok || e1 != nil || e2 != nil || a < 1 || z < a {
			return nil, fmt.Errorf("bad clvl band %q (want LO-HI)", p)
		}
		if n := len(out); n > 0 && a <= out[n-1].Hi {
			return nil, fmt.Errorf("clvl band %q overlaps the previous one", p)
		}
		out = append(out, Band{a, z})
	}
	return out, nil
}

// BandStat is the metrics of one run inside one clvl band. A band segment runs from the first minute line at clvl >= Lo
// to the first line above Hi (inclusive), so the level-up that leaves the band is charged to it and the segments of
// neighbouring bands share one boundary line and tile the run exactly.
type BandStat struct {
	Band      string  `json:"band"`
	Hours     float64 `json:"hours"`
	Minutes   int     `json:"minutes"`
	XP        int64   `json:"xp"`
	Kills     int     `json:"kills"`
	Deaths    int     `json:"deaths"`
	XPPerHour float64 `json:"xp_per_hour"`
	KillsPerH float64 `json:"kills_per_hour"`
	// Open is true when the run ended inside the band (the band was not left).
	Open bool `json:"open,omitempty"`
}

// BandHours is the play time of ms[i..j]: with total > 0 the lines are evenly spaced over total hours (d2d virtual
// clock), else the timestamps are summed and a gap above 150 s (a restart or pause) counts as one minute, so restart
// downtime is not play time (the definition of the early baseline, docs/d2d-baselines-early.md section 1).
func BandHours(ms []Minute, i, j int, total float64) float64 {
	if j <= i {
		return 0
	}
	if total > 0 {
		return total * float64(j-i) / float64(len(ms)-1)
	}
	h := 0.0
	for k := i + 1; k <= j; k++ {
		d := ms[k].At.Sub(ms[k-1].At).Seconds()
		if d > 150 || d < 0 {
			d = 60
		}
		h += d / 3600
	}
	return h
}

// Bands splits the run into the clvl bands that it entered (bands never reached are absent).
func Bands(ms []Minute, bands []Band, total float64) []BandStat {
	var out []BandStat
	for _, b := range bands {
		i := -1
		for k, m := range ms {
			if m.Clvl >= b.Lo {
				i = k
				break
			}
		}
		if i < 0 || ms[i].Clvl > b.Hi {
			continue
		}
		j, open := len(ms)-1, true
		for k := i; k < len(ms); k++ {
			if ms[k].Clvl > b.Hi {
				j, open = k, false
				break
			}
		}
		st := BandStat{Band: b.String(), Minutes: j - i, Open: open}
		st.Hours = BandHours(ms, i, j, total)
		st.XP = int64(ms[j].XP) - int64(ms[i].XP)
		for k := i + 1; k <= j; k++ {
			if d := ms[k].Kills - ms[k-1].Kills; d > 0 {
				st.Kills += d
			}
			if d := ms[k].Deaths - ms[k-1].Deaths; d > 0 {
				st.Deaths += d
			}
		}
		if st.Hours > 0 && !math.IsNaN(st.Hours) {
			st.XPPerHour = float64(st.XP) / st.Hours
			st.KillsPerH = float64(st.Kills) / st.Hours
		}
		out = append(out, st)
	}
	return out
}

// Band baseline and the per-band verdict (spec soak-gates, Gate 1 XP clause; card baselinefree).
const (
	// BandMinHours is the exposure in a band below which a baseline run does not count for that band.
	BandMinHours = 0.5
	// BandHardRuns is the number of baseline runs a band needs before its XP/h verdict can be FAIL.
	BandHardRuns = 3
	// BandCVFloor is the assumed per-run CV of XP/h when the data cannot show it (audit section 3: 14 percent).
	BandCVFloor = 0.14
	bandZ       = 1.96
)

// BandBase is the baseline of one clvl band: the runs of the baseline files with at least BandMinHours inside it.
type BandBase struct {
	Band   string   `json:"band"`
	N      int      `json:"n"`
	Hours  float64  `json:"hours"`
	Deaths int      `json:"deaths"`
	Mean   float64  `json:"xp_per_hour_mean"` // mean of the per-run values (each run counts once)
	SD     float64  `json:"sd"`
	CV     float64  `json:"cv"`         // max(SD/Mean, BandCVFloor) when N >= 2, else the floor
	HalfW  float64  `json:"half_width"` // 95 % half width of the mean (Student t), 0 when N < 2
	Runs   []string `json:"runs"`
}

// BandBaseline pools the band segments of all baseline rows (rows excluded from the mean still count: the band
// tables are their purpose).
func BandBaseline(rows []BaselineRun, band string) BandBase {
	bb := BandBase{Band: band}
	var v []float64
	for _, r := range rows {
		for _, s := range r.Bands {
			if s.Band != band || s.Hours < BandMinHours {
				continue
			}
			v = append(v, s.XPPerHour)
			bb.Hours += s.Hours
			bb.Deaths += s.Deaths
			bb.Runs = append(bb.Runs, r.Name)
		}
	}
	bb.N = len(v)
	if bb.N == 0 {
		bb.CV = BandCVFloor
		return bb
	}
	st := stat(v)
	bb.Mean, bb.SD = st.Mean, st.SD
	bb.CV = BandCVFloor
	if bb.N >= 2 && bb.Mean > 0 {
		bb.CV = math.Max(bb.SD/bb.Mean, BandCVFloor)
		bb.HalfW = st.CIHalf95
	}
	return bb
}

// BandVerdict values.
const (
	BandAdvisory = "ADVISORY" // too few baseline runs (or d2d exposure) in the band: never a FAIL
	BandOK       = "PASS"
	BandBelow    = "FAIL"
)

// BandJudge is the verdict of a d2d run's XP/h in one band against its baseline. seeds is the number of d2d runs
// pooled in the value (1 for Gate 1). The tolerance is the 95 percent single-run interval of audit section 3
// (z 1.96, used one-sided, so it is conservative): tol = 1.96 * sqrt(CV^2/seeds + CV^2/N), relative to the baseline mean. FAIL only when the
// d2d value is below mean*(1-tol) and the baseline has at least BandHardRuns runs; a value above the baseline passes.
func BandJudge(xph, hours float64, b BandBase, seeds int) (verdict string, tol, gap float64) {
	if b.N == 0 || b.Mean <= 0 {
		return BandAdvisory, 0, 0
	}
	if seeds < 1 {
		seeds = 1
	}
	tol = bandZ * math.Sqrt(b.CV*b.CV/float64(seeds)+b.CV*b.CV/float64(b.N))
	gap = xph/b.Mean - 1
	switch {
	case b.N < BandHardRuns || hours < BandMinHours:
		return BandAdvisory, tol, gap
	case gap < -tol:
		return BandBelow, tol, gap
	}
	return BandOK, tol, gap
}
