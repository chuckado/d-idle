// Command lumastat prints the mean luma and the p50/p95 of PNG frames as they come out of each stage of the viewer path:
// the reference (what d2d drew), the lossless palette feed (encode + decode), and the JPEG stage of d2dview (q72, full and
// half size). Brightness probe, docs/probes/viewer-brightness-2026-11-17.md. Usage: lumastat frame.png ...
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"math"
	"os"
	"sort"

	"d2lod-bot/internal/watch/palfeed"
)

type stats struct{ mean, lin, p50, p95 float64 }

func measure(pix []byte) stats {
	var sum, lin float64
	hist := make([]int, 256)
	n := len(pix) / 4
	for i := 0; i < len(pix); i += 4 {
		y := 0.299*float64(pix[i]) + 0.587*float64(pix[i+1]) + 0.114*float64(pix[i+2])
		sum += y
		hist[int(y+0.5)]++
		lin += 0.2126*toLinear(pix[i]) + 0.7152*toLinear(pix[i+1]) + 0.0722*toLinear(pix[i+2])
	}
	q := func(p float64) float64 {
		want, acc := int(p*float64(n)), 0
		for v, c := range hist {
			acc += c
			if acc > want {
				return float64(v)
			}
		}
		return 255
	}
	return stats{sum / float64(n), lin / float64(n), q(0.5), q(0.95)}
}

func toLinear(v byte) float64 {
	c := float64(v) / 255
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func rgba(img image.Image) *image.RGBA {
	out := image.NewRGBA(img.Bounds())
	draw.Draw(out, out.Rect, img, img.Bounds().Min, draw.Src)
	return out
}

func jpegRT(img *image.RGBA, q int) *image.RGBA {
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: q}); err != nil {
		panic(err)
	}
	d, err := jpeg.Decode(&b)
	if err != nil {
		panic(err)
	}
	return rgba(d)
}

func main() {
	sort.Strings(os.Args[1:])
	for _, path := range os.Args[1:] {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		src, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		ref := rgba(src)
		fr := &palfeed.Frame{W: ref.Rect.Dx(), H: ref.Rect.Dy(), Pix: ref.Pix, Seq: 1}
		msg, err := (&palfeed.Encoder{}).Encode(nil, fr)
		if err != nil {
			panic(err)
		}
		var dec palfeed.Decoder
		if err := dec.Apply(msg); err != nil {
			panic(err)
		}
		for _, row := range []struct {
			name string
			pix  []byte
		}{{"reference", ref.Pix}, {"palfeed", dec.Frame.Pix}, {"jpeg q72", jpegRT(ref, 72).Pix}} {
			s := measure(row.pix)
			fmt.Printf("%-28s %-10s meanY %6.2f linY %.4f p50 %3.0f p95 %3.0f\n", path, row.name, s.mean, s.lin, s.p50, s.p95)
		}
	}
}
