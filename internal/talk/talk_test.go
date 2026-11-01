package talk

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/quest"
	"d2lod-bot/internal/task"
	"d2lod-bot/internal/town"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time        { return c.t }
func (c *fakeClock) Sleep(d time.Duration) { c.t = c.t.Add(d) }

type fakeReader struct {
	d2.StateReader
	p *d2.Player
}

func (r *fakeReader) Snapshot() (*d2.Snapshot, error) { return &d2.Snapshot{Player: r.p}, nil }

type fakeIn struct {
	input.Controller
	closed int
}

func (f *fakeIn) CloseAll() error { f.closed++; return nil }

type pick struct{ idx, of int }

// fakeServices records the opens and line clicks; onPick can pay a skill point (the game's reward).
type fakeServices struct {
	opened  []int
	picks   []pick
	sockets int
	imbues  int
	openErr error
	onPick  func(n int)
}

func (f *fakeServices) OpenForTest(_ context.Context, class int) (d2.Unit, error) {
	f.opened = append(f.opened, class)
	return d2.Unit{}, f.openErr
}
func (f *fakeServices) PickLine(_ context.Context, _ d2.Unit, idx, n int) error {
	f.picks = append(f.picks, pick{idx, n})
	if f.onPick != nil {
		f.onPick(len(f.opened))
	}
	return nil
}
func (f *fakeServices) Socket(context.Context, town.SocketWant) (town.ServiceReport, error) {
	f.sockets++
	return town.ServiceReport{}, nil
}
func (f *fakeServices) Imbue(context.Context, town.ImbueWant) (town.ServiceReport, error) {
	f.imbues++
	return town.ServiceReport{}, nil
}

func rig(act int, svc *fakeServices) (Adapter, *d2.Player, *fakeIn) {
	p := &d2.Player{Act: act - 1}
	in := &fakeIn{}
	e := &env.Env{
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Read:  &fakeReader{p: p},
		In:    in,
		Clock: &fakeClock{t: time.Unix(1000, 0)},
	}
	return Adapter{T: svc, E: e, Status: testStatus}, p, in
}

// testStatus is the quest state of a character at the point each pick is meant for: the Den pending to be asked for,
// the travel rows (act1_done, act2_done, terrors_end) offered, everything else started.
func testStatus(q string) quest.Status {
	switch q {
	case "den_of_evil":
		return quest.Needed
	case "act1_done", "act2_done", "terrors_end":
		return quest.Done
	}
	return quest.Started
}

func TestClassIn(t *testing.T) {
	cases := []struct {
		npc  string
		act  int
		want int
		ok   bool
	}{
		{task.NPCAkara, 1, quest.Akara, true},
		{task.NPCAkara, 5, quest.Akara, true}, // any-act entry
		{task.NPCWarriv, 1, quest.Warriv, true},
		{task.NPCWarriv, 2, quest.WarrivA2, true},
		{task.NPCWarriv, 3, quest.Warriv, false}, // no act 3 entry and no any-act entry
		{task.NPCMeshif, 2, quest.Meshif, true},
		{task.NPCMeshif, 3, quest.MeshifA3, true},
		{task.NPCMeshif, 1, 0, false},
		{task.NPCCain, 3, quest.CainA3, true},
		{task.NPCTyrael, 4, quest.Tyrael, true},
		{task.NPCTyrael, 1, quest.TyraelTomb, true},
		{NPCAsheara, 3, quest.Asheara, true},
		{"Nobody", 2, 0, false},
	}
	for _, c := range cases {
		got, ok := ClassIn(c.npc, c.act)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("ClassIn(%s,%d) = %d,%v want %d,%v", c.npc, c.act, got, ok, c.want, c.ok)
		}
	}
}

// Every (npc, pick) of the table opens the NPC of the character's act and clicks one line per path label.
func TestEveryPathOpensAndClicks(t *testing.T) {
	var keys [][2]string
	for k := range Paths {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i][0]+keys[i][1] < keys[j][0]+keys[j][1] })
	for _, k := range keys {
		npc, pk := k[0], k[1]
		if (npc == task.NPCLarzuk && pk == "sockets") || (npc == task.NPCCharsi && pk == "imbue:best") || (npc == task.NPCAkara && pk == "reward") {
			continue // covered by their own tests
		}
		act := 0
		for a := 1; a <= 5 && act == 0; a++ {
			if _, ok := ClassIn(npc, a); ok {
				act = a
			}
		}
		if npc == task.NPCTyrael && pk != "reward" {
			act = 4 // the Act 4 Tyrael (the reward pick is the tomb one)
		}
		svc := &fakeServices{}
		ad, _, in := rig(act, svc)
		if err := ad.Talk(context.Background(), npc, pk); err != nil {
			t.Errorf("%v: %v", k, err)
			continue
		}
		want, _ := ClassIn(npc, act)
		if len(svc.opened) != 1 || svc.opened[0] != want {
			t.Errorf("%v: opened %v, want class %d", k, svc.opened, want)
		}
		if len(svc.picks) != len(Paths[k]) {
			t.Errorf("%v: %d clicks for path %v", k, len(svc.picks), Paths[k])
		}
		if in.closed != 1 {
			t.Errorf("%v: dialog closed %d times", k, in.closed)
		}
	}
}

func TestPerNPCRows(t *testing.T) {
	cases := []struct {
		name  string
		act   int
		npc   string
		pk    string
		class int
		rows  int // clicks
	}{
		{"Akara quest opens the topic list", 1, task.NPCAkara, "quest", quest.Akara, 2},
		{"Akara bark is the plain Talk", 1, task.NPCAkara, "bark", quest.Akara, 1},
		{"Kashya reward", 1, task.NPCKashya, "reward", quest.Kashya, 1},
		{"Charsi quest", 1, task.NPCCharsi, "quest", quest.Charsi, 1},
		{"Warriv act 1 go east", 1, task.NPCWarriv, "go east", quest.Warriv, 1},
		{"Meshif sail east", 2, task.NPCMeshif, "sail east", quest.Meshif, 1},
		{"empty pick is the plain Talk", 1, task.NPCCharsi, "", quest.Charsi, 1},
	}
	for _, c := range cases {
		svc := &fakeServices{}
		ad, _, _ := rig(c.act, svc)
		if err := ad.Talk(context.Background(), c.npc, c.pk); err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if len(svc.opened) != 1 || svc.opened[0] != c.class || len(svc.picks) != c.rows {
			t.Errorf("%s: opened %v picks %v, want class %d and %d clicks", c.name, svc.opened, svc.picks, c.class, c.rows)
		}
	}
}

// The row index of Warriv's "Go East" is the one quest.Menus gives for the visible rows.
func TestWarrivGoEastRow(t *testing.T) {
	svc := &fakeServices{}
	ad, _, _ := rig(1, svc)
	if err := ad.Talk(context.Background(), task.NPCWarriv, "go east"); err != nil {
		t.Fatal(err)
	}
	idx, n, _, err := quest.RowIndex(quest.Warriv, quest.LevelMain, "Go East", testStatus)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(svc.picks, []pick{{idx, n}}) {
		t.Errorf("clicked %v, want %d of %d", svc.picks, idx, n)
	}
}

func TestErrors(t *testing.T) {
	svc := &fakeServices{}
	ad, _, in := rig(1, svc)
	if err := ad.Talk(context.Background(), "Nobody", "quest"); err == nil || !strings.Contains(err.Error(), "no NPC class") {
		t.Errorf("unknown NPC: %v", err)
	}
	if err := ad.Talk(context.Background(), task.NPCAkara, "nonsense"); err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("unknown pick: %v", err)
	}
	if len(svc.opened) != 0 {
		t.Errorf("opened %v before the lookup failed", svc.opened)
	}
	svc.openErr = errors.New("too far")
	err := ad.Talk(context.Background(), task.NPCCharsi, "quest")
	if err == nil || !strings.Contains(err.Error(), "open Charsi") {
		t.Errorf("open failure: %v", err)
	}
	if in.closed != 1 {
		t.Errorf("a failed open closes the dialog once, got %d", in.closed)
	}
}

func TestItemPickers(t *testing.T) {
	svc := &fakeServices{}
	ad, _, _ := rig(5, svc)
	if err := ad.Talk(context.Background(), task.NPCLarzuk, "sockets"); err != nil || svc.sockets != 1 || len(svc.opened) != 0 {
		t.Errorf("Larzuk: err %v sockets %d opened %v", err, svc.sockets, svc.opened)
	}
	ad, _, _ = rig(1, svc)
	if err := ad.Talk(context.Background(), task.NPCCharsi, "imbue:best"); err != nil || svc.imbues != 1 || len(svc.opened) != 1 {
		t.Errorf("Charsi: err %v imbues %d opened %v", err, svc.imbues, svc.opened)
	}
}

// Akara's reward: paid on the first Talk; a miss retries (3 tries) with the other main-menu line count.
func TestAkaraRewardRetry(t *testing.T) {
	SetRespecSeen(false)
	defer SetRespecSeen(false)
	for _, c := range []struct {
		name    string
		payOn   int // the open (1-based) whose click pays; 0 never
		wantErr bool
		opens   int
	}{
		{"paid at once", 1, false, 1},
		{"paid on the retry", 2, false, 2},
		{"paid on the third try", 3, false, 3},
		{"never paid", 0, true, 3},
	} {
		SetRespecSeen(false)
		svc := &fakeServices{}
		ad, p, _ := rig(1, svc)
		svc.onPick = func(n int) {
			if c.payOn != 0 && n == c.payOn {
				p.SkillPoints++
			}
		}
		err := ad.Talk(context.Background(), task.NPCAkara, "reward")
		if (err != nil) != c.wantErr || len(svc.opened) != c.opens {
			t.Errorf("%s: err %v opens %d, want err=%v opens %d", c.name, err, len(svc.opened), c.wantErr, c.opens)
		}
		if RespecSeen() == c.wantErr {
			t.Errorf("%s: respec flag %v after err=%v", c.name, RespecSeen(), c.wantErr)
		}
		for _, pk := range svc.picks {
			if pk.idx != 0 {
				t.Errorf("%s: the reward claim clicks the Talk line (row 0), got %v", c.name, pk)
			}
		}
	}
}

// The respec line changes Akara's main-menu line count; the retry flips the assumption (the click is bottom-anchored).
func TestAkaraRespecLineCount(t *testing.T) {
	defer SetRespecSeen(false)
	counts := map[bool][2]int{}
	for _, seen := range []bool{false, true} {
		SetRespecSeen(seen)
		svc := &fakeServices{}
		ad, p, _ := rig(1, svc)
		svc.onPick = func(int) { p.SkillPoints++ }
		if err := ad.Talk(context.Background(), task.NPCAkara, "reward"); err != nil {
			t.Fatal(err)
		}
		counts[seen] = [2]int{svc.picks[0].idx, svc.picks[0].of}
	}
	if counts[true][1] != counts[false][1]+1 {
		t.Errorf("line counts without/with the respec row: %v", counts)
	}
	// the first retry (flip) of a fresh process uses the count with the respec line
	SetRespecSeen(false)
	svc := &fakeServices{}
	ad, p, _ := rig(1, svc)
	svc.onPick = func(n int) {
		if n == 2 {
			p.SkillPoints++
		}
	}
	if err := ad.Talk(context.Background(), task.NPCAkara, "reward"); err != nil {
		t.Fatal(err)
	}
	if svc.picks[0].of != counts[false][1] || svc.picks[1].of != counts[true][1] {
		t.Errorf("tries clicked %v, want of=%d then of=%d", svc.picks, counts[false][1], counts[true][1])
	}
}
