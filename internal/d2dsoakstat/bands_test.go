package d2dsoakstat

import (
	"math"
	"os"
	"testing"
	"time"
)

func mkMinutes(clvls []int, xps []uint64, kills []int, gapAfter int) []Minute {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var ms []Minute
	at := t0
	for i := range clvls {
		if i > 0 {
			at = at.Add(time.Minute)
			if i == gapAfter {
				at = at.Add(30 * time.Minute)
			}
		}
		ms = append(ms, Minute{At: at, Clvl: clvls[i], XP: xps[i], Kills: kills[i]})
	}
	return ms
}

func TestParseBands(t *testing.T) {
	bs, err := ParseBands(DefaultBands)
	if err != nil || len(bs) != 5 || bs[4].String() != "15-20" {
		t.Fatalf("%v %v", bs, err)
	}
	for _, bad := range []string{"", "3", "5-3", "1-4,4-7", "a-b"} {
		if _, err := ParseBands(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestBandsTileAndGap(t *testing.T) {
	// clvl 1 x3 lines, 2 x2, 5 x3 (a 30 min gap before line 6); xp rises 100 per line.
	ms := mkMinutes([]int{1, 1, 1, 2, 2, 5, 5, 5}, []uint64{0, 100, 200, 300, 400, 500, 600, 700}, []int{0, 1, 2, 3, 4, 5, 6, 7}, 6)
	bs, _ := ParseBands("1-4,5-7")
	got := Bands(ms, bs, 0)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	// band 1-4: lines 0..5 (line 5 is the first above 4): 500 XP, 5 minutes.
	if got[0].XP != 500 || got[0].Minutes != 5 || got[0].Open || math.Abs(got[0].Hours-5.0/60) > 1e-9 {
		t.Errorf("band 1-4 %+v", got[0])
	}
	// band 5-7: lines 5..7, the 30 min gap sits before line 6 so it counts as one minute: 2 min.
	if got[1].XP != 200 || !got[1].Open || math.Abs(got[1].Hours-2.0/60) > 1e-9 {
		t.Errorf("band 5-7 %+v", got[1])
	}
	// fixed total hours: evenly spaced over 7 intervals.
	g2 := Bands(ms, bs, 7)
	if math.Abs(g2[0].Hours-5) > 1e-9 {
		t.Errorf("fixed hours %+v", g2[0])
	}
}

func TestBandsSkipsUnreached(t *testing.T) {
	ms := mkMinutes([]int{6, 6, 7}, []uint64{0, 10, 20}, []int{0, 0, 0}, -1)
	bs, _ := ParseBands("1-4,5-7,8-10")
	got := Bands(ms, bs, 0)
	if len(got) != 1 || got[0].Band != "5-7" {
		t.Fatalf("%+v", got)
	}
}

func TestBandBaselineAndJudge(t *testing.T) {
	rows := []BaselineRun{
		{Run: Run{Name: "r1"}, Bands: []BandStat{{Band: "11-14", Hours: 2, XPPerHour: 10000, Deaths: 1}, {Band: "5-7", Hours: 2, XPPerHour: 9000}}},
		{Run: Run{Name: "r2"}, Exclude: "overlap", Bands: []BandStat{{Band: "11-14", Hours: 2, XPPerHour: 12000}}},
		{Run: Run{Name: "r3"}, Bands: []BandStat{{Band: "11-14", Hours: 0.2, XPPerHour: 1}, {Band: "11-14", Hours: 1, XPPerHour: 11000}}},
	}
	b := BandBaseline(rows, "11-14")
	if b.N != 3 || b.Mean != 11000 || b.Deaths != 1 || b.CV != BandCVFloor && b.CV < 0.0909 {
		t.Fatalf("%+v", b)
	}
	// N=3 -> tol = 1.96*sqrt(.14^2 + .14^2/3) = 31.7 %
	v, tol, gap := BandJudge(11000*0.65, 2, b, 1)
	if v != BandBelow || math.Abs(tol-0.3168) > 0.001 || gap > -0.34 {
		t.Errorf("%s %v %v", v, tol, gap)
	}
	if v, _, _ := BandJudge(11000*0.75, 2, b, 1); v != BandOK {
		t.Errorf("inside tolerance: %s", v)
	}
	if v, _, _ := BandJudge(100000, 2, b, 1); v != BandOK {
		t.Errorf("above baseline: %s", v)
	}
	// one baseline run only: advisory whatever the gap
	if v, _, _ := BandJudge(1, 2, BandBaseline(rows, "5-7"), 1); v != BandAdvisory {
		t.Errorf("n=1: %s", v)
	}
	if v, _, _ := BandJudge(1, 2, BandBaseline(rows, "1-4"), 1); v != BandAdvisory {
		t.Errorf("empty: %s", v)
	}
	// too little d2d exposure in the band
	if v, _, _ := BandJudge(1, 0.1, b, 1); v != BandAdvisory {
		t.Errorf("short: %s", v)
	}
}

// The committed baseline files carry the band tables of card baselinefree.
func TestCommittedBaselineBands(t *testing.T) {
	var rows []BaselineRun
	for _, p := range []string{"../../tools/d2dsoak/baselines.json", "../../tools/d2dsoak/baselines-early.json"} {
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		b, err := ReadBaselines(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Counted()) >= len(b.Runs) && p == "../../tools/d2dsoak/baselines.json" {
			t.Errorf("a14 partial must be excluded from the means")
		}
		for _, r := range b.Runs {
			if r.Backend == "" || r.RunID == "" && r.Name != "soaky-fresh-1-7" {
				t.Errorf("%s lacks provenance", r.Name)
			}
		}
		rows = append(rows, b.Runs...)
	}
	if b := BandBaseline(rows, "11-14"); b.N != 4 {
		t.Errorf("clvl 11-14 baseline runs = %d, want 4 (a14-full, a16, a18, g2-8h): %v", b.N, b.Runs)
	}
	for _, band := range []string{"1-4", "5-7", "8-10", "15-20"} {
		if b := BandBaseline(rows, band); b.N != 1 {
			t.Errorf("band %s has %d runs, want 1 (Fresha chain, G2)", band, b.N)
		}
	}
}
