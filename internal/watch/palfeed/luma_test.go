package palfeed

import (
	"bytes"
	"image"
	"image/jpeg"
	"math"
	"testing"
)

// meanLuma is the BT.601 luma (what JPEG's Y channel carries) averaged over a frame's RGB bytes.
func meanLuma(pix []byte) float64 {
	var s float64
	for i := 0; i+3 < len(pix); i += 4 {
		s += 0.299*float64(pix[i]) + 0.587*float64(pix[i+1]) + 0.114*float64(pix[i+2])
	}
	return s / float64(len(pix)/4)
}

// darkForest is a smooth, dark frame (deep shadows, a few bright lights) like a D2 outdoor scene.
func darkForest(w, h int) Frame {
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y*w + x) * 4
			base := 18 + 30*math.Sin(float64(x)/23)*math.Cos(float64(y)/17)
			if (x/40+y/40)%7 == 0 {
				base += 90 // a lit patch
			}
			pix[o], pix[o+1], pix[o+2], pix[o+3] = byte(base*0.9), byte(base), byte(base*0.7), 255
		}
	}
	return Frame{W: w, H: h, Pix: pix, Seq: 1}
}

// The brightness regression (docs/probes/viewer-brightness-2026-11-17.md): neither viewer path may lift or crush the
// picture. The lossless feed is exact, so its luma equals the reference; the JPEG path (full-range sRGB JFIF as Go writes
// it, q72 like d2dview) stays within one luma level of the reference mean.
func TestViewerPathsKeepLuminance(t *testing.T) {
	ref := darkForest(800, 600)
	want := meanLuma(ref.Pix)

	var dec Decoder
	msg, err := (&Encoder{}).Encode(nil, &ref)
	if err != nil {
		t.Fatal(err)
	}
	if err := dec.Apply(msg); err != nil {
		t.Fatal(err)
	}
	if got := meanLuma(dec.Frame.Pix); got != want {
		t.Errorf("palette feed mean luma %.3f, reference %.3f: the lossless path must be exact", got, want)
	}

	img := &image.RGBA{Pix: ref.Pix, Stride: ref.W * 4, Rect: image.Rect(0, 0, ref.W, ref.H)}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 72}); err != nil {
		t.Fatal(err)
	}
	d, err := jpeg.Decode(&b)
	if err != nil {
		t.Fatal(err)
	}
	out := image.NewRGBA(d.Bounds())
	for y := 0; y < ref.H; y++ {
		for x := 0; x < ref.W; x++ {
			out.Set(x, y, d.At(x, y))
		}
	}
	if got := meanLuma(out.Pix); math.Abs(got-want) > 1.0 {
		t.Errorf("JPEG q72 mean luma %.3f, reference %.3f: more than one level apart (range or matrix wrong)", got, want)
	}
}
