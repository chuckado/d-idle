package charsel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"
)

// Header offsets (V: the same constants internal/cheat/d2s verifies against real saves). Only these fields are read;
// the item section, stats and quests are never touched here.
const (
	offSignature  = 0x00
	offVersion    = 0x04
	offName       = 0x14
	offStatus     = 0x24
	offClass      = 0x28
	offLevel      = 0x2B
	offLastPlayed = 0x30
	headerLen     = 0x34

	signature = 0xAA55AA55
	version96 = 0x60

	flagHardcore  = 0x04
	flagDied      = 0x08
	flagExpansion = 0x20
)

// Header is what the list shows of one save.
type Header struct {
	Name       string
	Class      Class
	Level      int
	Hardcore   bool
	Died       bool
	Expansion  bool
	LastPlayed time.Time // zero when the field is 0
}

// ErrNotACharacter is the header check failing: wrong signature or version, or a file too short.
var ErrNotACharacter = errors.New("charsel: not a version 96 character file")

// ParseHeader decodes the header fields of a .d2s image. It does not verify the checksum (the list must still show a
// damaged save so the player can archive it); Verify at switch time does.
func ParseHeader(b []byte) (Header, error) {
	if len(b) < headerLen || binary.LittleEndian.Uint32(b[offSignature:]) != signature || binary.LittleEndian.Uint32(b[offVersion:]) != version96 {
		return Header{}, ErrNotACharacter
	}
	name := b[offName : offName+16]
	if i := indexNUL(name); i >= 0 {
		name = name[:i]
	}
	cl, err := ClassByD2S(int(b[offClass]))
	if err != nil {
		return Header{}, err
	}
	h := Header{Name: string(name), Class: cl, Level: int(b[offLevel]),
		Hardcore: b[offStatus]&flagHardcore != 0, Died: b[offStatus]&flagDied != 0, Expansion: b[offStatus]&flagExpansion != 0}
	if ts := binary.LittleEndian.Uint32(b[offLastPlayed:]); ts != 0 {
		h.LastPlayed = time.Unix(int64(ts), 0).UTC()
	}
	return h, nil
}

func indexNUL(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return -1
}

// ReadHeader reads only the first bytes of a file.
func ReadHeader(path string) (Header, error) {
	f, err := os.Open(path)
	if err != nil {
		return Header{}, err
	}
	defer f.Close()
	b := make([]byte, headerLen)
	if n, err := f.Read(b); n < headerLen {
		if err == nil {
			err = fmt.Errorf("short read")
		}
		return Header{}, fmt.Errorf("%w: %v", ErrNotACharacter, err)
	}
	return ParseHeader(b)
}
