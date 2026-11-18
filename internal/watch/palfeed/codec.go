// Package palfeed is the prototype of a lossless frame feed for the watch viewer (docs/video-feed.md): the d2d frame
// ring's RGBA frames, sent as tile deltas against the last frame a client holds, deflated with the stdlib, drawn on a
// canvas by web/palfeed.js. Off unless the host wires it in (knob watch.paletteFeed, default off); the MJPEG path is the
// fallback and is not touched.
//
// Wire message (little endian), one per frame:
//
//	0  "PF"  2 u8 version (1)  3 u8 flags (1 key frame)  4 u16 width  6 u16 height  8 u8 tile edge in pixels  9 u8 0
//	10 u32 seq  14 u32 base seq (0 for a key frame)  18 u32 game ms  22 u32 payload length
//	26 i8 dx  27 i8 dy (the scroll the encoder found: the view moved by (dx,dy) since the base; 0,0 for a key frame)
//	28 payload
//
// The payload is raw DEFLATE (RFC 1951, what DecompressionStream("deflate-raw") and zlib.inflateRaw read) of
//
//	key frame: every tile's RGB bytes, tile after tile (row-major), each tile row left-neighbour filtered (byte minus
//	  the byte 3 to the left, the first pixel raw); alpha is dropped, d2d draws opaque.
//	delta: a mode per tile, 2 bits each, row-major, 4 tiles per byte (tile 0 in the low bits), then the RGB bytes of
//	  every tile of mode 1 or 2, tile after tile. Let S(x,y) be the base pixel at (x-dx, y-dy), black outside the frame.
//	  mode 0: same as the base at the same place (no data); mode 1: base XOR data; mode 2: S XOR data;
//	  mode 3: exactly S (no data). The scroll modes exist because the whole isometric view moves while the hero walks.
package palfeed

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Version    = 1
	HeaderSize = 28
	FlagKey    = 1

	// MaxShift is the largest scroll per frame the encoder looks for, in pixels per axis.
	MaxShift = 32
)

// ErrNeedKey: the message is a delta against a frame this decoder does not hold; reconnect or ask for a key frame.
var ErrNeedKey = errors.New("palfeed: delta against a frame we do not hold")

// Frame is one drawn frame: RGBA, W*H*4 bytes (alpha ignored on the wire).
type Frame struct {
	W, H   int
	Pix    []byte
	Seq    uint32
	GameMS uint32
	ref    *frameRef // hub bookkeeping; nil for a frame that is not pooled
}

// Encoder turns frames into messages. Not safe for concurrent use; one per client (it holds the deflate state).
type Encoder struct {
	Tile    int  // tile edge in pixels, default 16
	NoShift bool // never look for a scroll (the measurement uses it for the comparison)
	buf     bytes.Buffer
	raw     []byte
	fw      *flate.Writer
	sref    []byte // one shifted tile row block
	last    [2]int // the shift of the previous frame pair
}

func (e *Encoder) tile() int {
	if e.Tile <= 0 || e.Tile > 255 {
		return 16
	}
	return e.Tile
}

// Encode returns the message taking a client that holds base (nil: none) to cur. A nil base, or a base of another size,
// gives a key frame. The result is valid until the next call.
func (e *Encoder) Encode(base, cur *Frame) ([]byte, error) {
	if cur == nil || len(cur.Pix) != cur.W*cur.H*4 || cur.W <= 0 || cur.H <= 0 || cur.W > 65535 || cur.H > 65535 {
		return nil, errors.New("palfeed: bad frame")
	}
	key := base == nil || base.W != cur.W || base.H != cur.H || len(base.Pix) != len(cur.Pix)
	t := e.tile()
	tx, ty := (cur.W+t-1)/t, (cur.H+t-1)/t
	var dx, dy int
	e.raw = e.raw[:0]
	if key {
		for j := 0; j < ty; j++ {
			for i := 0; i < tx; i++ {
				x0, y0 := i*t, j*t
				w, h := min(t, cur.W-x0), min(t, cur.H-y0)
				for y := y0; y < y0+h; y++ {
					e.raw = appendSubRGB(e.raw, cur.Pix[(y*cur.W+x0)*4:(y*cur.W+x0+w)*4])
				}
			}
		}
	} else {
		if !e.NoShift {
			dx, dy = estimateShift(base, cur, e.last)
			e.last = [2]int{dx, dy}
		}
		modes := make([]byte, (tx*ty+3)/4)
		e.raw = append(e.raw, modes...) // placeholder, patched below
		for j := 0; j < ty; j++ {
			for i := 0; i < tx; i++ {
				x0, y0 := i*t, j*t
				w, h := min(t, cur.W-x0), min(t, cur.H-y0)
				if tileEqual(base.Pix, cur.Pix, cur.W, x0, y0, w, h) {
					continue
				}
				mode := 1
				if dx != 0 || dy != 0 {
					mode = e.shiftedMode(base, cur, x0, y0, w, h, dx, dy)
				}
				n := j*tx + i
				modes[n>>2] |= byte(mode) << (2 * (n & 3))
				switch mode {
				case 1:
					for y := y0; y < y0+h; y++ {
						e.raw = appendXorRGB(e.raw, base.Pix[(y*cur.W+x0)*4:], cur.Pix[(y*cur.W+x0)*4:(y*cur.W+x0+w)*4])
					}
				case 2:
					for y := y0; y < y0+h; y++ {
						sr := shiftedRow(e.sref[:0], base, x0-dx, y-dy, w)
						e.sref = sr
						e.raw = appendXorRGB(e.raw, sr, cur.Pix[(y*cur.W+x0)*4:(y*cur.W+x0+w)*4])
					}
				}
			}
		}
		copy(e.raw, modes)
	}
	e.buf.Reset()
	e.buf.Write(make([]byte, HeaderSize))
	if e.fw == nil {
		e.fw, _ = flate.NewWriter(&e.buf, flate.BestSpeed)
	} else {
		e.fw.Reset(&e.buf)
	}
	if _, err := e.fw.Write(e.raw); err != nil {
		return nil, err
	}
	if err := e.fw.Close(); err != nil {
		return nil, err
	}
	m := e.buf.Bytes()
	m[0], m[1], m[2] = 'P', 'F', Version
	if key {
		m[3] = FlagKey
	}
	binary.LittleEndian.PutUint16(m[4:], uint16(cur.W))
	binary.LittleEndian.PutUint16(m[6:], uint16(cur.H))
	m[8] = byte(t)
	binary.LittleEndian.PutUint32(m[10:], cur.Seq)
	if !key {
		binary.LittleEndian.PutUint32(m[14:], base.Seq)
	}
	binary.LittleEndian.PutUint32(m[18:], cur.GameMS)
	binary.LittleEndian.PutUint32(m[22:], uint32(len(m)-HeaderSize))
	m[26], m[27] = byte(int8(dx)), byte(int8(dy))
	return m, nil
}

// shiftedMode picks the mode (1, 2 or 3) of a tile that differs from the unshifted base.
func (e *Encoder) shiftedMode(base, cur *Frame, x0, y0, w, h, dx, dy int) int {
	nz0, nz1 := 0, 0
	for y := y0; y < y0+h; y++ {
		c := cur.Pix[(y*cur.W+x0)*4 : (y*cur.W+x0+w)*4]
		b := base.Pix[(y*cur.W+x0)*4:]
		sr := shiftedRow(e.sref[:0], base, x0-dx, y-dy, w)
		e.sref = sr
		for i := 0; i < len(c); i += 4 {
			if c[i] != b[i] || c[i+1] != b[i+1] || c[i+2] != b[i+2] {
				nz0++
			}
			if c[i] != sr[i] || c[i+1] != sr[i+1] || c[i+2] != sr[i+2] {
				nz1++
			}
		}
	}
	switch {
	case nz1 == 0:
		return 3
	case nz1 < nz0:
		return 2
	}
	return 1
}

// shiftedRow appends w RGBA pixels of base row sy starting at column sx (black where outside the frame).
func shiftedRow(dst []byte, base *Frame, sx, sy, w int) []byte {
	n := w * 4
	if cap(dst) < n {
		dst = make([]byte, 0, n)
	}
	dst = dst[:n]
	if sy < 0 || sy >= base.H {
		clear(dst)
		return dst
	}
	row := base.Pix[sy*base.W*4 : (sy+1)*base.W*4]
	lo, hi := max(sx, 0), min(sx+w, base.W)
	if lo >= hi {
		clear(dst)
		return dst
	}
	clear(dst[:(lo-sx)*4])
	copy(dst[(lo-sx)*4:], row[lo*4:hi*4])
	clear(dst[(hi-sx)*4:])
	return dst
}

// EstimateShift looks for the scroll between two frames: the (dx,dy) within MaxShift for which the most sampled pixels
// of cur equal the base pixel at (x-dx, y-dy). It answers 0,0 when the frames already agree (a still view) or no shift
// is clearly better. hint is the shift of the previous frame pair (a walk scrolls at a steady rate), tried first.
func EstimateShift(base, cur *Frame) (int, int) { return estimateShift(base, cur, [2]int{}) }

func lattice(w, h, step int) [][2]int {
	var pts [][2]int
	for y := step / 2; y < h; y += step {
		for x := step / 2; x < w; x += step {
			pts = append(pts, [2]int{x, y})
		}
	}
	return pts
}

func shiftScore(base, cur *Frame, pts [][2]int, dx, dy int) int {
	n := 0
	for _, p := range pts {
		sx, sy := p[0]-dx, p[1]-dy
		if sx < 0 || sy < 0 || sx >= base.W || sy >= base.H {
			continue
		}
		a, b := (p[1]*cur.W+p[0])*4, (sy*base.W+sx)*4
		if cur.Pix[a] == base.Pix[b] && cur.Pix[a+1] == base.Pix[b+1] && cur.Pix[a+2] == base.Pix[b+2] {
			n++
		}
	}
	return n
}

func estimateShift(base, cur *Frame, hint [2]int) (int, int) {
	fine := lattice(cur.W, cur.H, 17)
	z := shiftScore(base, cur, fine, 0, 0)
	if z*10 >= len(fine)*9 {
		return 0, 0
	}
	best, bdx, bdy := z, 0, 0
	try := func(dx, dy int) {
		if abs(dx) > MaxShift || abs(dy) > MaxShift || (dx == 0 && dy == 0) {
			return
		}
		if s := shiftScore(base, cur, fine, dx, dy); s > best {
			best, bdx, bdy = s, dx, dy
		}
	}
	if hint != [2]int{} { // the last shift, then its neighbours: a steady walk usually lands here
		try(hint[0], hint[1])
		if best*10 >= len(fine)*7 {
			return bdx, bdy
		}
	}
	coarse := lattice(cur.W, cur.H, 31)
	type cand struct{ s, dx, dy int }
	var top [4]cand
	for dy := -MaxShift; dy <= MaxShift; dy++ {
		for dx := -MaxShift; dx <= MaxShift; dx++ {
			s := shiftScore(base, cur, coarse, dx, dy)
			for k := range top { // keep the four best
				if s > top[k].s {
					copy(top[k+1:], top[k:len(top)-1])
					top[k] = cand{s, dx, dy}
					break
				}
			}
		}
	}
	for _, c := range top {
		if c.s > 0 {
			try(c.dx, c.dy)
		}
	}
	if best < z+len(fine)/10 {
		return 0, 0
	}
	return bdx, bdy
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func tileEqual(a, b []byte, stride, x0, y0, w, h int) bool {
	for y := y0; y < y0+h; y++ {
		o := (y*stride + x0) * 4
		if !bytes.Equal(a[o:o+w*4], b[o:o+w*4]) {
			return false
		}
	}
	return true
}

func appendSubRGB(dst, rgba []byte) []byte {
	var pr, pg, pb byte
	for i := 0; i+3 < len(rgba); i += 4 {
		r, g, b := rgba[i], rgba[i+1], rgba[i+2]
		dst = append(dst, r-pr, g-pg, b-pb)
		pr, pg, pb = r, g, b
	}
	return dst
}

func appendXorRGB(dst, base, rgba []byte) []byte {
	for i := 0; i+3 < len(rgba); i += 4 {
		dst = append(dst, rgba[i]^base[i], rgba[i+1]^base[i+1], rgba[i+2]^base[i+2])
	}
	return dst
}

// Decoder holds the frame a client shows and applies messages to it. The Go decoder exists for tests and tools; the
// browser uses web/palfeed.js, which must stay byte-compatible (the cross tests in palfeed_test.go and
// tools/videofeed/palfeed.test.js share fixtures).
type Decoder struct {
	Frame Frame
	have  bool
	raw   bytes.Buffer
	spare []byte
	sref  []byte
}

// Apply decodes one message into d.Frame. A delta whose base is not d.Frame.Seq returns ErrNeedKey and changes nothing;
// after any other error the decoder must be fed a key frame (reconnect).
func (d *Decoder) Apply(m []byte) error {
	if len(m) < HeaderSize || m[0] != 'P' || m[1] != 'F' || m[2] != Version {
		return errors.New("palfeed: bad header")
	}
	key := m[3]&FlagKey != 0
	w, h := int(binary.LittleEndian.Uint16(m[4:])), int(binary.LittleEndian.Uint16(m[6:]))
	t := int(m[8])
	seq, base := binary.LittleEndian.Uint32(m[10:]), binary.LittleEndian.Uint32(m[14:])
	n := int(binary.LittleEndian.Uint32(m[22:]))
	dx, dy := int(int8(m[26])), int(int8(m[27]))
	if w == 0 || h == 0 || t == 0 || len(m) != HeaderSize+n {
		return errors.New("palfeed: bad message")
	}
	if !key && (!d.have || d.Frame.W != w || d.Frame.H != h || d.Frame.Seq != base) {
		return ErrNeedKey
	}
	d.raw.Reset()
	if _, err := io.Copy(&d.raw, io.LimitReader(flate.NewReader(bytes.NewReader(m[HeaderSize:])), int64(w*h*3+w*h/4+1024))); err != nil {
		return fmt.Errorf("palfeed: inflate: %w", err)
	}
	raw := d.raw.Bytes()
	tx, ty := (w+t-1)/t, (h+t-1)/t
	var pix []byte
	pos, nm := 0, 0
	if key {
		pix = make([]byte, w*h*4)
	} else {
		nm = (tx*ty + 3) / 4
		if len(raw) < nm {
			return errors.New("palfeed: short payload")
		}
		pos = nm
		if cap(d.spare) < w*h*4 {
			d.spare = make([]byte, w*h*4)
		}
		pix = d.spare[:w*h*4]
		copy(pix, d.Frame.Pix)
	}
	old := &Frame{W: w, H: h, Pix: d.Frame.Pix}
	for j := 0; j < ty; j++ {
		for i := 0; i < tx; i++ {
			x0, y0 := i*t, j*t
			tw, th := min(t, w-x0), min(t, h-y0)
			mode := 1
			if !key {
				k := j*tx + i
				mode = int(raw[k>>2]>>(2*(k&3))) & 3
				if mode == 0 {
					continue
				}
			}
			if mode != 3 && pos+tw*th*3 > len(raw) {
				return errors.New("palfeed: short payload")
			}
			for y := y0; y < y0+th; y++ {
				o := (y*w + x0) * 4
				var ref []byte
				switch mode {
				case 1:
					if !key {
						ref = old.Pix[o : o+tw*4]
					}
				case 2, 3:
					d.sref = shiftedRow(d.sref[:0], old, x0-dx, y-dy, tw)
					ref = d.sref
				}
				switch {
				case mode == 3:
					copy(pix[o:o+tw*4], ref)
				case key:
					var pr, pg, pb byte
					for x := 0; x < tw; x++ {
						pr, pg, pb = pr+raw[pos], pg+raw[pos+1], pb+raw[pos+2]
						pos += 3
						pix[o+4*x], pix[o+4*x+1], pix[o+4*x+2], pix[o+4*x+3] = pr, pg, pb, 255
					}
				default:
					for x := 0; x < tw; x++ {
						pix[o+4*x], pix[o+4*x+1], pix[o+4*x+2], pix[o+4*x+3] = ref[4*x]^raw[pos], ref[4*x+1]^raw[pos+1], ref[4*x+2]^raw[pos+2], 255
						pos += 3
					}
				}
			}
		}
	}
	if !key {
		d.spare = d.Frame.Pix[:0:cap(d.Frame.Pix)]
		if cap(d.spare) < w*h*4 {
			d.spare = nil
		}
	}
	d.Frame = Frame{W: w, H: h, Pix: pix, Seq: seq, GameMS: binary.LittleEndian.Uint32(m[18:])}
	d.have = true
	return nil
}
