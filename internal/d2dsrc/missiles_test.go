package d2dsrc

import (
	"context"
	"strings"
	"testing"

	"d2lod-bot/internal/d2dctl"
)

// A d2d that lists missiles fills Snapshot.Missiles with the fields Game.exe's reader fills.
func TestMissilesListed(t *testing.T) {
	r := lateReplies(t)
	r["debug missiles"] = "5\t41\t40.30\t22.10\t1.0\t0.0\t8.0\t1\t752\t30\n6\t7\t41.0\t22.0\t-1\t0\t12\t0\t1\t20\nok\n"
	_, c := startScript(t, r)
	src := New(c, Options{Tables: fakeTables{}, Logf: quiet})
	res, err := src.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := src.FillMissiles(context.Background(), res); err != nil {
		t.Fatal(err)
	}
	ms := res.Snap.Missiles
	if len(ms) != 2 {
		t.Fatalf("missiles %+v", ms)
	}
	m := ms[0]
	if m.ID != 5 || m.Class != 41 || m.X != 201 || m.Y != 110 || m.OwnerType != 1 || m.OwnerID != 752 {
		t.Fatalf("missile %+v", m)
	}
	if ms[1].OwnerType != 0 || ms[1].OwnerID != res.Snap.Player.ID {
		t.Fatalf("own missile %+v (player id %d)", ms[1], res.Snap.Player.ID)
	}
}

// No listing (a d2d without the verb) or an empty one: no missiles, and the verb is not asked again after an err.
func TestMissilesAbsent(t *testing.T) {
	s, c := startScript(t, lateReplies(t))
	src := New(c, Options{Tables: fakeTables{}, Logf: quiet})
	for i := 0; i < 3; i++ {
		res, err := src.Read(context.Background())
		if err == nil {
			err = src.FillMissiles(context.Background(), res)
		}
		if err != nil || len(res.Snap.Missiles) != 0 {
			t.Fatalf("read %d: %v %+v", i, err, res.Snap.Missiles)
		}
	}
	n := 0
	for _, l := range s.seen {
		if l == "debug missiles" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("debug missiles sent %d times, want 1 (latched off after the err)", n)
	}

	r := lateReplies(t)
	r["debug missiles"] = "ok\n"
	_, c = startScript(t, r)
	src = New(c, Options{Tables: fakeTables{}, Logf: quiet})
	res, err := src.Read(context.Background())
	if err == nil {
		err = src.FillMissiles(context.Background(), res)
	}
	if err != nil || len(res.Snap.Missiles) != 0 {
		t.Fatalf("empty listing: %v %+v", err, res.Snap.Missiles)
	}
}

func TestParseMissile(t *testing.T) {
	m, err := d2dctl.ParseMissile("5\t41\t40.3\t22.1\t1\t0\t8\t1\t752\t30")
	if err != nil || m.Class != 41 || m.OwnerID != 752 || m.Speed != 8 || m.TTL != 30 {
		t.Fatalf("%+v %v", m, err)
	}
	if _, err := d2dctl.ParseMissile(strings.Repeat("1\t", 5)); err == nil {
		t.Fatal("short row accepted")
	}
}
