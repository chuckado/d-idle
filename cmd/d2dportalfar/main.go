// Command d2dportalfar takes the standing town portal of the current level with d2dctl.EnterPortal and reports the level before and
// after (lab probe, card qdmatrix: live "portal click did not change the level" from 70 subtiles away). It needs a d2d control
// socket on a real-time (not lockstep) instance with the portal standing; tools/probes/portalfar.py sets that up.
//
//	d2dportalfar -sock PATH
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"d2lod-bot/internal/d2dctl"
)

func main() {
	sock := flag.String("sock", "", "d2d control socket")
	wall := flag.Duration("wall", 2*time.Minute, "wall limit")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), *wall)
	defer cancel()
	c, err := d2dctl.Dial(ctx, *sock, d2dctl.Options{})
	if err != nil {
		fmt.Println("dial:", err)
		os.Exit(2)
	}
	before, _ := c.Level(ctx)
	t0 := time.Now()
	err = c.EnterPortal(ctx)
	after, _ := c.Level(ctx)
	fmt.Printf("level %d -> %d at (%.1f,%.1f) in %v, err: %v\n", before.ID, after.ID, after.X, after.Y, time.Since(t0).Round(time.Second), err)
	if err != nil {
		os.Exit(1)
	}
}
