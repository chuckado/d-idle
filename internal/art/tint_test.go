package art

import (
	"bytes"
	"errors"
	"image/png"
	"testing"
)

func pixels(t *testing.T, s *Store, key string) (b []byte, rgb [2][4][3]uint8) {
	t.Helper()
	b, err := s.PNG(key)
	if err != nil {
		t.Fatalf("%s: %v", key, err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			rgb[y][x] = [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8)}
		}
	}
	return b, rgb
}

func TestTintKeys(t *testing.T) {
	for k, want := range map[string]bool{
		"item-invtst": true, "item-invtst~t1c0": true, "item-invtst~t8c20": true, "item-invtst~t6c12": true,
		"item-invtst~t0c1": false, "item-invtst~t9c1": false, "item-invtst~t1c21": false, "item-invtst~t1c": false,
		"item-invtst~": false, "item-invtst~t1c01": false, "item-invtst~c1": false, "item-invtst~t1c1~t1c1": false,
		"item-invtst~T1C1": false, "item-invtst~t-1c1": false,
	} {
		if got := ValidKey(k); got != want {
			t.Errorf("ValidKey(%q) = %v, want %v", k, got, want)
		}
	}
}

// Two item kinds (a jewel-like icon in the base archive and an armor-like one in the expansion archive).
func TestTintRemap(t *testing.T) {
	s, err := Open(synthInstall(t), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, base := range []string{"invtst", "invexp"} {
		plainPNG, plain := pixels(t, s, "item-"+base)
		// set 1 colour 3: index i -> 1+(i+30)%250, via pal (i, 2, 3)
		_, c3 := pixels(t, s, "item-"+base+"~t1c3")
		_, c4 := pixels(t, s, "item-"+base+"~t1c4")
		if c3 == plain {
			t.Errorf("%s: tinted pixels equal the plain art", base)
		}
		if c3 == c4 {
			t.Errorf("%s: same base, different colour gave identical pixels", base)
		}
		if got := c3[0][1][0]; got != 1+(5+30)%250 {
			t.Errorf("%s: pixel (1,0) red = %d, want %d", base, got, 1+(5+30)%250)
		}
		if _, _, _, a := func() (uint32, uint32, uint32, uint32) {
			img, _ := png.Decode(bytes.NewReader(plainPNG))
			return img.At(0, 0).RGBA()
		}(); a != 0 {
			t.Error("transparent pixel changed")
		}
		// a different set file gives a different result for the same colour
		_, s5 := pixels(t, s, "item-"+base+"~t6c3")
		if s5 == c3 {
			t.Errorf("%s: different colormap set gave the same pixels", base)
		}
		// no tint = byte-identical to the plain art, also after tinted renders ran
		again, _ := pixels(t, s, "item-"+base)
		if !bytes.Equal(plainPNG, again) {
			t.Errorf("%s: plain art changed after tinting", base)
		}
	}
}

func TestTintFallback(t *testing.T) {
	s, err := Open(synthInstall(t), "")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	plain, _ := pixels(t, s, "item-invtst")
	// set 2 has no file in the synthetic install, set 3 and 4 are never tinted by the game: plain icon, no error
	for _, k := range []string{"item-invtst~t2c3", "item-invtst~t3c1", "item-invtst~t4c1"} {
		got, err := s.PNG(k)
		if err != nil || !bytes.Equal(got, plain) {
			t.Errorf("%s: err %v, plain fallback = %v", k, err, bytes.Equal(got, plain))
		}
	}
	// a missing base is still not found; a bad key still 404s
	if _, err := s.PNG("item-nope~t1c3"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing base: %v", err)
	}
	if _, err := s.PNG("item-invtst~t9c3"); !errors.Is(err, ErrBadKey) {
		t.Errorf("bad tint: %v", err)
	}
}

func TestColorMapShort(t *testing.T) {
	if _, err := ColorMap(make([]byte, 21*256-1), 0); err == nil {
		t.Error("short colormap accepted")
	}
	if _, err := ColorMap(make([]byte, 21*256), 21); err == nil {
		t.Error("colour 21 accepted")
	}
}

func TestTintDiskCacheOnePerKey(t *testing.T) {
	cache := t.TempDir()
	s, err := Open(synthInstall(t), cache)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a, _ := s.PNG("item-invtst~t1c3")
	b, _ := s.PNG("item-invtst~t1c3")
	if !bytes.Equal(a, b) {
		t.Error("cache returned different bytes")
	}
	s.mu.Lock()
	n := len(s.mem)
	s.mu.Unlock()
	if n != 1 {
		t.Errorf("%d memory entries after two reads of one key", n)
	}
}
