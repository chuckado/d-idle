package art

import (
	"fmt"
)

// tintSets names the colormap files for InvTrans 1..8 (itemcolour design section 1; the game ships them as
// data\global\items\palette\<name>.dat, 21 colours x 256 bytes, read from the user's install at runtime).
// The fork never sends sets 3 and 4 (tints_with is false for them), so they are listed only for completeness.
var tintSets = [9]string{1: "grey", 2: "grey2", 3: "gold", 4: "brown", 5: "greybrown", 6: "invgrey", 7: "invgrey2", 8: "invgreybrown"}

const (
	tintColours = 21
	paletteDir  = `data\global\items\palette\`
)

// tintSetFile is the install path of the colormap for set, "" when the set has none.
func tintSetFile(set int) string {
	if set < 1 || set >= len(tintSets) || tintSets[set] == "" {
		return ""
	}
	return paletteDir + tintSets[set] + ".dat"
}

// colormap returns the 256-entry remap for colour inside set. The rendered PNG is cached per key by the store, so one remap runs per (base, set, colour).
func (s *Store) colormap(set, colour int) ([]byte, error) {
	name := tintSetFile(set)
	if name == "" || colour < 0 || colour >= tintColours {
		return nil, fmt.Errorf("%w: tint set %d colour %d", ErrNotFound, set, colour)
	}
	raw, err := s.read(name)
	if err != nil {
		return nil, err
	}
	return ColorMap(raw, colour)
}

// ColorMap cuts colour's 256 bytes out of a colormap file (21 x 256).
func ColorMap(raw []byte, colour int) ([]byte, error) {
	if len(raw) < tintColours*256 || colour < 0 || colour >= tintColours {
		return nil, fmt.Errorf("art: colormap has %d bytes or colour %d out of range", len(raw), colour)
	}
	return raw[colour*256 : colour*256+256], nil
}

// Remap returns a copy of f with every palette index sent through cm (index 0 stays transparent).
func (f Frame) Remap(cm []byte) Frame {
	out := f
	out.Pix = make([]byte, len(f.Pix))
	for i, c := range f.Pix {
		if c != 0 && len(cm) == 256 {
			c = cm[c]
		}
		out.Pix[i] = c
	}
	return out
}
