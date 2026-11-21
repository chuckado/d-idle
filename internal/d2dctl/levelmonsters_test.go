package d2dctl

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl/fake"
)

const levelMonstersJSON = `{"level":25,"unit":"cell","rooms_up":[0,3],"presets":[{"index":2,"kind":"superunique","id":740,"super":6,"name":"The Countess","x":61.3,"y":44.1,"room":5,"up":false},{"index":4,"kind":"monster","id":12,"super":-1,"name":"zombie1","x":10.5,"y":11.5,"room":0,"up":true}],"monsters":[{"id":9,"type":3,"name":"zombie1","display":"Zombie","x":10.5,"y":11.5,"boss":"none","super":-1,"mode":"NU","life":50,"alive":true}]}`

func TestLevelMonstersOverTheWire(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	fix := fake.Sample()
	fix.LevelMonsters = levelMonstersJSON
	srv, err := fake.New(path, fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := New(path, Options{Timeout: 2 * time.Second})
	defer c.Close()
	lm, err := c.LevelMonsters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if lm.Level != 25 || len(lm.RoomsUp) != 2 || len(lm.Presets) != 2 || len(lm.Monsters) != 1 {
		t.Fatalf("%+v", lm)
	}
	if p := lm.Presets[0]; p.Kind != "superunique" || p.Super != 6 || p.Name != "The Countess" || p.X != 61.3 || p.Up {
		t.Fatalf("preset %+v", p)
	}
	if m := lm.Monsters[0]; m.ID != 9 || m.Display != "Zombie" || !m.Alive || m.Boss != "none" {
		t.Fatalf("monster %+v", m)
	}
}

// A fork without the verb (or stock d2d) answers err: the caller gets an error, never an empty list.
func TestLevelMonstersVerbMissing(t *testing.T) {
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
	if lm, err := c.LevelMonsters(context.Background()); err == nil || lm != nil {
		t.Fatalf("want an error, got %+v %v", lm, err)
	}
}
