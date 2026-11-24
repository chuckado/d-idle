package main

import (
	"context"
	"fmt"
)

// tombAreas are Tal Rasha's seven tombs (levels.txt 66..72): one of them, picked per game, is the true tomb whose chamber has
// the orifice for the Horadric Staff (V devp28 doc: after `debug rooms up` in level 66 the objectlist shows the orifice, the true
// tomb depends on the seed). Each tomb is a door of the Canyon of the Magi (area 46).
var tombAreas = []int{66, 67, 68, 69, 70, 71, 72}

const canyonOfMagi = 46

// findTrueTomb walks the tombs from the Canyon, one after the other, and stops in the one that has the orifice. area reads
// the current level, exitTo takes an adjacent level, hasOrifice looks for the object in the current level's map. Standing in
// the true tomb or a tomb that has it already is the end at once; any other level is an error (the scan starts in the Canyon).
func findTrueTomb(ctx context.Context, area func() int, exitTo func(context.Context, int) error, hasOrifice func() bool) error {
	if a := area(); a >= tombAreas[0] && a <= tombAreas[len(tombAreas)-1] && hasOrifice() {
		return nil
	}
	for _, t := range tombAreas {
		if a := area(); a != canyonOfMagi {
			if a < tombAreas[0] || a > tombAreas[len(tombAreas)-1] {
				return fmt.Errorf("tomb scan from level %d: not the Canyon of the Magi", a)
			}
			if err := exitTo(ctx, canyonOfMagi); err != nil {
				return fmt.Errorf("back to the Canyon from tomb %d: %w", a, err)
			}
		}
		if err := exitTo(ctx, t); err != nil {
			return fmt.Errorf("tomb %d: %w", t, err)
		}
		if hasOrifice() {
			return nil
		}
	}
	return fmt.Errorf("none of the seven tombs has the orifice")
}
