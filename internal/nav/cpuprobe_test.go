package nav

import (
	"context"
	"os"
	"testing"
	"time"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// townTour walks the simulated Rogue Encampment (recorded level data) between its services, the way a repair trip
// does. It is the scenario of docs/probes/bot-cpu-2026-11-16.md: BenchmarkTownTour reports bot CPU per simulated
// second of walking (the sim clock only advances when the navigator sleeps, so wall time is pure navigator CPU).
func townTour(tb testing.TB, file string) (s *sim, n *Navigator, stops []input.Point) {
	fh, err := os.Open("../d2map/testdata/" + file)
	if err != nil {
		tb.Fatal(err)
	}
	defer fh.Close()
	img, err := d2map.LoadImage(fh)
	if err != nil {
		tb.Fatal(err)
	}
	tr := d2map.NewTracker(img)
	st, err := tr.Update()
	if err != nil {
		tb.Fatal(err)
	}
	lm := tr.Current()
	for _, c := range [][2]int{{d2map.PresetNPC, 154}, {d2map.PresetObject, 267}, {d2map.PresetNPC, 148}, {d2map.PresetObject, 119}, {d2map.PresetNPC, 147}} {
		if p, ok := presetPos(lm, c[0], c[1]); ok {
			stops = append(stops, p)
		}
	}
	s = newSim(lm.Grid, input.Point{X: st.Pos.X, Y: st.Pos.Y}, &simLevel{no: 1, rect: lm.Rect})
	s.speed = 9.4
	s.stopShort = 5.5
	return s, s.navigator(), stops
}

func BenchmarkTownTour(b *testing.B) {
	for _, file := range townFixtures {
		b.Run(file, func(b *testing.B) {
			var simTime time.Duration
			var snaps, clicks int
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				s, n, stops := townTour(b, file)
				t0 := s.now
				b.StartTimer()
				for _, p := range stops {
					s.exact = &p
					if err := n.GoTo(context.Background(), p); err != nil {
						b.Fatal(err)
					}
				}
				simTime += s.now.Sub(t0)
				snaps += s.updates
				clicks += len(s.clicks)
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/simTime.Seconds()/1e6, "cpu-ms/sim-s")
			b.ReportMetric(float64(snaps)/simTime.Seconds(), "updates/sim-s")
			b.ReportMetric(float64(clicks)/simTime.Seconds(), "clicks/sim-s")
		})
	}
}
