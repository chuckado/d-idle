package palfeed

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestCompare is the honest bandwidth comparison on a long recording (tools/videofeed/record2.py): JPEG at several
// qualities, the plain tile delta, the tile delta with the scroll search, a lossy variant of it (colour truncated to 6
// bits), and the hybrid (delta when it is not larger than the JPEG, else a JPEG that leaves the client's lossless base
// as it was, so the next delta is built from that older base: the cost of that is counted, not assumed).
// D2FEED_FRAMES=dir D2FEED_OUT=report.txt go test -run TestCompare -v ./internal/watch/palfeed
func TestCompare(t *testing.T) {
	dir := os.Getenv("D2FEED_FRAMES")
	if dir == "" {
		t.Skip("D2FEED_FRAMES not set")
	}
	files, _ := filepath.Glob(filepath.Join(dir, "f*.png"))
	sort.Strings(files)
	if len(files) < 10 {
		t.Fatalf("only %d frames", len(files))
	}
	scenes := readScenes(dir, len(files))
	methods := []string{"jpeg q72", "jpeg q50", "jpeg q35", "delta plain", "delta+scroll", "delta+scroll 6bit", "hybrid q72"}
	type acc struct{ b map[string][]int }
	by := map[string]*acc{}
	scrolling := map[string][2]int{} // scene -> frames with a scroll, frames
	add := func(scene string, m string, n int) {
		for _, k := range []string{scene, "ALL"} {
			a := by[k]
			if a == nil {
				a = &acc{b: map[string][]int{}}
				by[k] = a
			}
			a.b[m] = append(a.b[m], n)
		}
	}
	load := func(f string) Frame {
		fr, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		defer fr.Close()
		img, err := png.Decode(fr)
		if err != nil {
			t.Fatal(err)
		}
		b := img.Bounds()
		pix := make([]byte, b.Dx()*b.Dy()*4)
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				o := (y*b.Dx() + x) * 4
				pix[o], pix[o+1], pix[o+2], pix[o+3] = byte(r>>8), byte(g>>8), byte(bl>>8), 255
			}
		}
		return Frame{W: b.Dx(), H: b.Dy(), Pix: pix}
	}
	jpegLen := func(f *Frame, q int) int {
		var b bytes.Buffer
		_ = jpeg.Encode(&b, &image.RGBA{Pix: f.Pix, Stride: f.W * 4, Rect: image.Rect(0, 0, f.W, f.H)}, &jpeg.Options{Quality: q})
		return b.Len()
	}
	quant := func(f Frame) Frame {
		p := make([]byte, len(f.Pix))
		for i, v := range f.Pix {
			p[i] = v & 0xFC
		}
		f.Pix = p
		return f
	}
	plain, scroll, scrollQ, hyb := &Encoder{NoShift: true}, &Encoder{}, &Encoder{}, &Encoder{}
	var prev, prevQ, hbase *Frame
	for i, f := range files {
		cur := load(f)
		curQ := quant(cur)
		scene := sceneOf(scenes, i)
		if prev != nil {
			add(scene, "jpeg q72", jpegLen(&cur, 72))
			add(scene, "jpeg q50", jpegLen(&cur, 50))
			add(scene, "jpeg q35", jpegLen(&cur, 35))
			m, _ := plain.Encode(prev, &cur)
			add(scene, "delta plain", len(m))
			m, _ = scroll.Encode(prev, &cur)
			add(scene, "delta+scroll", len(m))
			c := scrolling[scene]
			c[1]++
			if int8(m[26]) != 0 || int8(m[27]) != 0 {
				c[0]++
			}
			scrolling[scene] = c
			m, _ = scrollQ.Encode(prevQ, &curQ)
			add(scene, "delta+scroll 6bit", len(m))
			dm, _ := hyb.Encode(hbase, &cur)
			j := jpegLen(&cur, 72)
			if len(dm) <= j {
				add(scene, "hybrid q72", len(dm))
				c := cur
				hbase = &c
			} else {
				add(scene, "hybrid q72", j)
			}
		} else {
			c := cur
			hbase = &c
		}
		p, pq := cur, curQ
		prev, prevQ = &p, &pq
	}
	var out bytes.Buffer
	var names []string
	for _, s := range scenes {
		names = append(names, s.name)
	}
	names = append(names, "ALL")
	fmt.Fprintf(&out, "frames %d, 25 Hz, bytes per frame (mean / p50 / p95) and MB/s at 25 fps (mean x 25)\n", len(files))
	mean := func(v []int) float64 {
		s := 0
		for _, x := range v {
			s += x
		}
		return float64(s) / float64(len(v))
	}
	seen := map[string]bool{}
	scen := map[string]map[string]float64{}
	for _, name := range names {
		a := by[name]
		if a == nil || seen[name] {
			continue
		}
		seen[name] = true
		sc := scrolling[name]
		if name == "ALL" {
			for _, c := range scrolling {
				sc[0] += c[0]
			}
		}
		fmt.Fprintf(&out, "\n%s (n %d, frames with a scroll found %d)\n", name, len(a.b["jpeg q72"]), sc[0])
		for _, m := range methods {
			v := a.b[m]
			fmt.Fprintf(&out, "  %-18s mean %8.0f  p50 %7d  p95 %7d  %6.2f MB/s\n", m, mean(v), pi(v, 50), pi(v, 95), mean(v)*25/1e6)
		}
		scen[name] = map[string]float64{}
		for _, m := range methods {
			scen[name][m] = mean(a.b[m]) * 25 / 1e6
		}
	}
	// weighted: walking share w of the time (town walk + field walk pooled by their own frame counts), 10 percent fight,
	// the rest standing (idle scenes pooled)
	pool := func(pred func(string) bool) map[string]float64 {
		r := map[string]float64{}
		for _, m := range methods {
			var sum float64
			var n int
			for name, a := range by {
				if name != "ALL" && pred(name) {
					sum += mean(a.b[m]) * float64(len(a.b[m]))
					n += len(a.b[m])
				}
			}
			if n > 0 {
				r[m] = sum / float64(n) * 25 / 1e6
			}
		}
		return r
	}
	walk := pool(func(s string) bool { return strings.Contains(s, "walk") })
	fight := pool(func(s string) bool { return s == "fight" })
	still := pool(func(s string) bool { return strings.Contains(s, "idle") })
	fmt.Fprintf(&out, "\nweighted MB/s (walking share w, fight 10 percent, standing the rest; pools: walk %v, fight, idle)\n", "town+field walk")
	fmt.Fprintf(&out, "  %-18s %7s %7s %7s %7s\n", "method", "w=30%", "w=50%", "w=70%", "w=77%")
	for _, m := range methods {
		fmt.Fprintf(&out, "  %-18s", m)
		for _, w := range []float64{0.3, 0.5, 0.7, 0.77} {
			f := min(0.1, 1-w)
			fmt.Fprintf(&out, " %7.2f", w*walk[m]+f*fight[m]+(1-w-f)*still[m])
		}
		fmt.Fprintln(&out)
	}
	t.Log("\n" + out.String())
	if p := os.Getenv("D2FEED_OUT"); p != "" {
		_ = os.WriteFile(p, out.Bytes(), 0o644)
	}
}
