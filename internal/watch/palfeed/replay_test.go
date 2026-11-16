package palfeed

import (
	"bufio"
	"bytes"
	"compress/flate"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestReplay measures the codec on recorded d2d frames (tools/videofeed/record.py writes them; they are never
// committed). D2FEED_FRAMES=dir D2FEED_OUT=report.txt go test -run TestReplay -v ./internal/watch/palfeed
func TestReplay(t *testing.T) {
	dir := os.Getenv("D2FEED_FRAMES")
	if dir == "" {
		t.Skip("D2FEED_FRAMES not set")
	}
	files, _ := filepath.Glob(filepath.Join(dir, "f*.png"))
	sort.Strings(files)
	if len(files) < 10 {
		t.Fatalf("only %d frames in %s", len(files), dir)
	}
	scenes := readScenes(dir, len(files))
	frames := make([]Frame, len(files))
	alphaOK := true
	maxColours, colourSum := 0, 0
	over256 := 0
	for i, f := range files {
		fr, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(fr)
		fr.Close()
		if err != nil {
			t.Fatal(err)
		}
		b := img.Bounds()
		pix := make([]byte, b.Dx()*b.Dy()*4)
		colours := map[uint32]struct{}{}
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				o := (y*b.Dx() + x) * 4
				pix[o], pix[o+1], pix[o+2], pix[o+3] = byte(r>>8), byte(g>>8), byte(bl>>8), byte(a>>8)
				if a>>8 != 255 {
					alphaOK = false
				}
				colours[uint32(pix[o])<<16|uint32(pix[o+1])<<8|uint32(pix[o+2])] = struct{}{}
			}
		}
		frames[i] = Frame{W: b.Dx(), H: b.Dy(), Pix: pix, Seq: uint32(i + 1), GameMS: uint32(i * 40)}
		maxColours = max(maxColours, len(colours))
		colourSum += len(colours)
		if len(colours) > 256 {
			over256++
		}
	}
	if p := os.Getenv("D2FEED_MSGS"); p != "" { // the 25 fps wire stream (tile 16) for the node and browser decode benches
		var wire bytes.Buffer
		enc := &Encoder{Tile: 16}
		for i := range frames {
			var base *Frame
			if i > 0 {
				base = &frames[i-1]
			}
			m, err := enc.Encode(base, &frames[i])
			if err != nil {
				t.Fatal(err)
			}
			wire.Write([]byte{byte(len(m)), byte(len(m) >> 8), byte(len(m) >> 16), byte(len(m) >> 24)})
			wire.Write(m)
		}
		if err := os.WriteFile(p, wire.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	w := bufio.NewWriter(&out)
	fmt.Fprintf(w, "frames %d  %dx%d  alpha all 255: %v  unique RGB colours per frame: mean %d, max %d, frames over 256: %d\n",
		len(frames), frames[0].W, frames[0].H, alphaOK, colourSum/len(frames), maxColours, over256)

	type row struct {
		name         string
		bytes        []int
		encMs, decMs []float64
	}
	for _, tile := range []int{8, 16, 32} {
		for _, step := range []int{1, 2} { // 25 fps client, and a client that only takes every second frame (12.5 fps)
			enc := &Encoder{Tile: tile}
			var dec Decoder
			rows := map[string]*row{}
			var order []string
			add := func(scene string, n int, e, d float64) {
				r := rows[scene]
				if r == nil {
					r = &row{name: scene}
					rows[scene] = r
					order = append(order, scene)
				}
				r.bytes = append(r.bytes, n)
				r.encMs = append(r.encMs, e)
				r.decMs = append(r.decMs, d)
			}
			var keyBytes int
			for i := 0; i < len(frames); i += step {
				var base *Frame
				if i >= step {
					base = &frames[i-step]
				}
				t0 := time.Now()
				m, err := enc.Encode(base, &frames[i])
				e := time.Since(t0)
				if err != nil {
					t.Fatal(err)
				}
				if base == nil {
					keyBytes = len(m)
				}
				if i == 0 || base == nil {
					dec = Decoder{}
				}
				t1 := time.Now()
				if err := dec.Apply(m); err != nil {
					t.Fatal(err)
				}
				d := time.Since(t1)
				if !bytes.Equal(rgbOnly(dec.Frame.Pix), rgbOnly(frames[i].Pix)) {
					t.Fatalf("tile %d step %d frame %d: decoded pixels differ", tile, step, i)
				}
				if base != nil {
					add(sceneOf(scenes, i), len(m), ms(e), ms(d))
					add("ALL", len(m), ms(e), ms(d))
				}
			}
			fmt.Fprintf(w, "\ntile %d, every %d frame(s) (%g fps): key frame %d B\n", tile, step, 25/float64(step), keyBytes)
			fmt.Fprintf(w, "  %-18s %6s %9s %9s %9s %9s %9s %9s\n", "scene", "n", "B p50", "B p95", "B max", "enc p50", "enc p95", "dec p95")
			for _, s := range append(order[:0:0], order...) {
				r := rows[s]
				fmt.Fprintf(w, "  %-18s %6d %9d %9d %9d %8.2fms %8.2fms %8.2fms\n", r.name, len(r.bytes), pi(r.bytes, 50), pi(r.bytes, 95), pi(r.bytes, 100), pf(r.encMs, 50), pf(r.encMs, 95), pf(r.decMs, 95))
			}
		}
	}

	// reference points on the same frames: JPEG q72 (what /stream.mjpg sends) and one deflate of the whole-frame XOR.
	var jb, xb []int
	var jms, xms []float64
	for i := 1; i < len(frames); i += 3 {
		img := &image.RGBA{Pix: frames[i].Pix, Stride: frames[i].W * 4, Rect: image.Rect(0, 0, frames[i].W, frames[i].H)}
		var b bytes.Buffer
		t0 := time.Now()
		_ = jpeg.Encode(&b, img, &jpeg.Options{Quality: 72})
		jms = append(jms, ms(time.Since(t0)))
		jb = append(jb, b.Len())
		x := make([]byte, len(frames[i].Pix)/4*3)
		t0 = time.Now()
		for k, o := 0, 0; k < len(frames[i].Pix); k, o = k+4, o+3 {
			x[o], x[o+1], x[o+2] = frames[i].Pix[k]^frames[i-1].Pix[k], frames[i].Pix[k+1]^frames[i-1].Pix[k+1], frames[i].Pix[k+2]^frames[i-1].Pix[k+2]
		}
		var zb bytes.Buffer
		fw, _ := flate.NewWriter(&zb, flate.BestSpeed)
		fw.Write(x)
		fw.Close()
		xms = append(xms, ms(time.Since(t0)))
		xb = append(xb, zb.Len())
	}
	fmt.Fprintf(w, "\nreference, every 3rd frame: JPEG q72 bytes p50 %d p95 %d, encode p50 %.2f ms p95 %.2f ms (lossy)\n", pi(jb, 50), pi(jb, 95), pf(jms, 50), pf(jms, 95))
	fmt.Fprintf(w, "reference, every 3rd frame: whole-frame XOR + deflate level 1 bytes p50 %d p95 %d, time p50 %.2f ms p95 %.2f ms (no tiles)\n", pi(xb, 50), pi(xb, 95), pf(xms, 50), pf(xms, 95))
	w.Flush()
	t.Log("\n" + out.String())
	if p := os.Getenv("D2FEED_OUT"); p != "" {
		_ = os.WriteFile(p, out.Bytes(), 0o644)
	}
}

func rgbOnly(p []byte) []byte {
	o := make([]byte, 0, len(p)/4*3)
	for i := 0; i < len(p); i += 4 {
		o = append(o, p[i], p[i+1], p[i+2])
	}
	return o
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func pi(v []int, p int) int {
	s := append([]int(nil), v...)
	sort.Ints(s)
	return s[min(len(s)-1, (len(s)*p+99)/100-1+boolInt(p == 100))]
}

func pf(v []float64, p int) float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s[min(len(s)-1, max(0, (len(s)*p+99)/100-1))]
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type scene struct {
	first int
	name  string
}

func readScenes(dir string, n int) []scene {
	b, err := os.ReadFile(filepath.Join(dir, "scenes.txt"))
	if err != nil {
		return []scene{{0, "all"}}
	}
	var s []scene
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var i int
		var name string
		if k := strings.IndexByte(l, ' '); k > 0 {
			fmt.Sscanf(l[:k], "%d", &i)
			name = l[k+1:]
			s = append(s, scene{i, name})
		}
	}
	return s
}

func sceneOf(s []scene, i int) string {
	name := "all"
	for _, sc := range s {
		if i >= sc.first {
			name = sc.name
		}
	}
	return name
}
