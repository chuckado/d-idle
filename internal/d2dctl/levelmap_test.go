package d2dctl

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl/fake"
)

const levelMapJSON = `{"level":5,"unit":"cell","width":120,"height":90,"rooms":[{"x":44,"y":44,"w":8,"h":8,"preset":true,"def":161},{"x":2,"y":2,"w":8,"h":8,"preset":false,"def":-1}],"objects":[{"index":3,"class":30,"name":"Tree of Inifuss","x":48.5,"y":48.5,"kind":"tree","state":"NU","hidden":false}],"warps":[{"index":0,"x":1.0,"y":2.0,"to":2}],"portals":[]}`

func TestLevelMapOverTheWire(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	fix := fake.Sample()
	fix.LevelMap = levelMapJSON
	srv, err := fake.New(path, fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := New(path, Options{Timeout: 2 * time.Second})
	defer c.Close()
	lm, err := c.LevelMap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if lm.Level != 5 || lm.Unit != "cell" || len(lm.Rooms) != 2 || lm.Rooms[0].Def != 161 || !lm.Rooms[0].Preset {
		t.Fatalf("rooms %+v", lm)
	}
	if o := lm.Objects[0]; o.Class != 30 || o.X != 48.5 || o.Kind != "tree" || o.Name != "Tree of Inifuss" {
		t.Fatalf("object %+v", o)
	}
	if len(lm.Warps) != 1 || lm.Warps[0].To != 2 {
		t.Fatalf("warps %+v", lm.Warps)
	}
}

// A fork without the verb (or stock d2d) answers err: the caller gets an error, never an empty layout.
func TestLevelMapVerbMissing(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	srv, err := fake.New(path, fake.Sample(), fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := New(path, Options{Timeout: 2 * time.Second})
	defer c.Close()
	if lm, err := c.LevelMap(context.Background()); err == nil || lm != nil {
		t.Fatalf("missing verb: %+v %v", lm, err)
	}
	if _, err := ParseLevelMap(&Reply{Lines: []string{"not json"}}); err == nil {
		t.Fatal("garbage parsed")
	}
}
