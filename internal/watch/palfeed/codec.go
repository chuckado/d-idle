// Package palfeed is the prototype of a lossless frame feed for the watch viewer (docs/video-feed.md): the d2d frame
// ring's RGBA frames, sent as tile deltas against the last frame a client holds, deflated with the stdlib, drawn on a
// canvas by web/palfeed.js. Off unless the host wires it in (knob watch.paletteFeed, default off); the MJPEG path is the
// fallback and is not touched.
//
// Wire message (little endian), one per frame:
//
//	0  "PF"  2 u8 version (1)  3 u8 flags (1 key frame)  4 u16 width  6 u16 height  8 u8 tile edge in pixels  9 u8 0
//	10 u32 seq  14 u32 base seq (0 for a key frame)  18 u32 game ms  22 u32 payload length  26 payload
//
// The payload is raw DEFLATE (RFC 1951, what DecompressionStream("deflate-raw") and zlib.inflateRaw read) of:
// a bitmap of dirty tiles (row-major, bit 0 first), then for each dirty tile its RGB bytes row by row (alpha is dropped,
// d2d draws opaque). A key frame has every tile dirty and stores each tile row left-neighbour filtered (byte minus the
// byte 3 to the left, the first pixel raw); a delta stores the XOR with the same pixel of the base frame.
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
	HeaderSize = 26
	FlagKey    = 1
)

// ErrNeedKey: the message is a delta against a frame this decoder does not hold; reconnect or ask for a key frame.
var ErrNeedKey = errors.New("palfeed: delta against a frame we do not hold")

// Frame is one drawn frame: RGBA, W*H*4 bytes (alpha ignored on the wire).
type Frame struct {
	W, H   int
	Pix    []byte
	Seq    uint32
	GameMS uint32
}

// Encoder turns frames into messages. Not safe for concurrent use; one per client (it holds the deflate state).
type Encoder struct {
	Tile int // tile edge in pixels, default 16
	buf  bytes.Buffer
	raw  []byte
	fw   *flate.Writer
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
	bitmap := make([]byte, (tx*ty+7)/8)
	e.raw = e.raw[:0]
	e.raw = append(e.raw, bitmap...) // placeholder, patched below
	for j := 0; j < ty; j++ {
		for i := 0; i < tx; i++ {
			x0, y0 := i*t, j*t
			w, h := min(t, cur.W-x0), min(t, cur.H-y0)
			if !key && tileEqual(base.Pix, cur.Pix, cur.W, x0, y0, w, h) {
				continue
			}
			n := j*tx + i
			bitmap[n>>3] |= 1 << (n & 7)
			for y := y0; y < y0+h; y++ {
				row := cur.Pix[(y*cur.W+x0)*4 : (y*cur.W+x0+w)*4]
				if key {
					e.raw = appendSubRGB(e.raw, row)
				} else {
					e.raw = appendXorRGB(e.raw, base.Pix[(y*cur.W+x0)*4:], row)
				}
			}
		}
	}
	copy(e.raw, bitmap)
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
	return m, nil
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
// web/palfeed.test.js share fixtures).
type Decoder struct {
	Frame Frame
	have  bool
	raw   bytes.Buffer
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
	if w == 0 || h == 0 || t == 0 || len(m) != HeaderSize+n {
		return errors.New("palfeed: bad message")
	}
	if !key && (!d.have || d.Frame.W != w || d.Frame.H != h || d.Frame.Seq != base) {
		return ErrNeedKey
	}
	d.raw.Reset()
	if _, err := io.Copy(&d.raw, io.LimitReader(flate.NewReader(bytes.NewReader(m[HeaderSize:])), int64(w*h*3+w*h/8+1024))); err != nil {
		return fmt.Errorf("palfeed: inflate: %w", err)
	}
	raw := d.raw.Bytes()
	tx, ty := (w+t-1)/t, (h+t-1)/t
	nb := (tx*ty + 7) / 8
	if len(raw) < nb {
		return errors.New("palfeed: short payload")
	}
	pix := d.Frame.Pix
	if key || len(pix) != w*h*4 {
		pix = make([]byte, w*h*4)
	}
	pos := nb
	for j := 0; j < ty; j++ {
		for i := 0; i < tx; i++ {
			k := j*tx + i
			if raw[k>>3]&(1<<(k&7)) == 0 {
				continue
			}
			x0, y0 := i*t, j*t
			tw, th := min(t, w-x0), min(t, h-y0)
			if pos+tw*th*3 > len(raw) {
				return errors.New("palfeed: short payload")
			}
			for y := y0; y < y0+th; y++ {
				o := (y*w + x0) * 4
				var pr, pg, pb byte
				for x := 0; x < tw; x++ {
					a, b, c := raw[pos], raw[pos+1], raw[pos+2]
					pos += 3
					if key {
						pr, pg, pb = pr+a, pg+b, pb+c
						pix[o], pix[o+1], pix[o+2] = pr, pg, pb
					} else {
						pix[o], pix[o+1], pix[o+2] = pix[o]^a, pix[o+1]^b, pix[o+2]^c
					}
					pix[o+3] = 255
					o += 4
				}
			}
		}
	}
	d.Frame = Frame{W: w, H: h, Pix: pix, Seq: seq, GameMS: binary.LittleEndian.Uint32(m[18:])}
	d.have = true
	return nil
}
