// Command objclass prints objects.txt rows (name, IsDoor, Selectable) for object classes: objclass GAMEDATA CLASS...
// (card mapverify: is the object at the Monastery Gate border a door?). Read-only.
package main

import (
	"fmt"
	"os"
	"strconv"

	"d2lod-bot/internal/data"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: objclass GAMEDATA CLASS...")
		os.Exit(2)
	}
	db, err := data.LoadInstall(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, a := range os.Args[2:] {
		c, _ := strconv.Atoi(a)
		if o := db.ObjectByID(c); o != nil {
			fmt.Printf("%d name=%q isdoor=%v selectable=%v blocksvis=%v\n", c, o.Name, o.IsDoor, o.Selectable, o.BlocksVis)
		} else {
			fmt.Printf("%d none\n", c)
		}
	}
}
