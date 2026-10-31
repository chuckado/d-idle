package main

import (
	"context"
	"d2lod-bot/internal/d2map"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dworld"
)

// Mon is one monster on the map (tile coordinates in the act frame, like Snap.X and Y).
type Mon struct {
	X, Y   float64
	Name   string
	HP, Mx int
	Boss   bool
	Dead   bool
	Friend bool `json:",omitempty"`
}

// Snap is what the page draws; it is rebuilt at every poll.
type Snap struct {
	Seq        int64           `json:"seq"`
	OK         bool            `json:"ok"`
	Err        string          `json:"err,omitempty"`
	Level      int             `json:"level"`
	Area       string          `json:"area"`
	MapVer     uint64          `json:"mapVer"`
	MapRect    [4]int          `json:"mapRect"` // x0, y0, w, h in act-absolute subtiles of the PNG
	Stitch     int             `json:"stitch"`  // layer group of the level: the map and the trail persist while it is unchanged
	X          float64         `json:"x"`       // player tile in the act frame (level origin added: continuous across a stitched border)
	Y          float64         `json:"y"`
	Name       string          `json:"name"`
	Class      string          `json:"class"`
	CLvl       int             `json:"clvl"`
	XP         int             `json:"xp"`
	Gold       int             `json:"gold"`
	Life       [2]int          `json:"life"`
	Mana       [2]int          `json:"mana"`
	Mons       []Mon           `json:"mons"`
	Dead       int             `json:"deadSeen"`
	GameMS     int             `json:"gameMs"` // virtual clock of the lockstep d2d (resets when d2d restarts)
	Tick       int             `json:"tick"`
	Speed      float64         `json:"speed"` // virtual s per wall s, smoothed over about 2 s
	Lockstep   bool            `json:"lockstep"`
	Exits      [][3]int        `json:"exits"` // dest, x, y (tiles, act frame)
	Bot        json.RawMessage `json:"bot,omitempty"`
	Mode       Mode            `json:"mode"`
	FrameTicks int64           `json:"frameTicks"` // ticks between the last two sampled frames
	FrameMS    int             `json:"frameMs"`    // wall ms the last screenshot took
	Frames     int64           `json:"frames"`
	Tap        TapInfo         `json:"tap"`
}

// TapInfo shows what the viewer costs, so the page can say it.
type TapInfo struct {
	Steps, Ticks, ViewOps, BotLines int64
}

// Feed polls the d2d through the viewer socket and keeps the latest Snap and map PNG. It polls only while the page is
// being watched (Touch), so an unwatched run pays nothing.
type Feed struct {
	c    *d2dctl.Client
	w    *d2dworld.World
	mux  *Mux
	fr   *Frames
	hz   float64
	bot  string // bot API base URL ("" off)
	mu   sync.RWMutex
	snap Snap
	png  []byte
	mapV uint64 // counts composites: bumps when the stitched map changes
	mapK string // stitchKey of the composite drawn
	mapR [4]int
	seen map[int]bool
	// speed estimate
	lastMS   int
	lastWall time.Time
	ema      float64
	seq      int64
	lastSeen int64
	botJSON  json.RawMessage
	botAt    time.Time
}

func NewFeed(viewSock string, mux *Mux, hz float64, bot string) *Feed {
	c := d2dctl.New(viewSock, d2dctl.Options{Timeout: 3 * time.Second, NoReconnect: false})
	f := &Feed{c: c, mux: mux, hz: hz, bot: bot, seen: map[int]bool{}}
	f.w = d2dworld.New(c, d2dworld.Options{Timeout: 5 * time.Second, MinRefetch: 3 * time.Second})
	return f
}

func (f *Feed) Touch() { f.lastSeenStore(time.Now().UnixNano()) }

func (f *Feed) lastSeenStore(v int64) { f.mu.Lock(); f.lastSeen = v; f.mu.Unlock() }

func (f *Feed) watching() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return time.Since(time.Unix(0, f.lastSeen)) < 5*time.Second
}

// Run polls until ctx ends.
func (f *Feed) Run(ctx context.Context) {
	t := time.NewTicker(time.Duration(float64(time.Second) / f.hz))
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		if !f.watching() {
			continue
		}
		f.poll(ctx)
	}
}

// Wait blocks until a snapshot newer than seq exists (or ctx ends).
func (f *Feed) Wait(ctx context.Context, seq int64) Snap {
	for {
		s := f.Snap()
		if s.Seq > seq {
			return s
		}
		select {
		case <-ctx.Done():
			return s
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func (f *Feed) Snap() Snap {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.snap
}

func (f *Feed) PNG() ([]byte, uint64) {
	b, v, _ := f.Map()
	return b, v
}

// Map is the stitched map PNG with its version and its box (x0, y0, w, h in act-absolute subtiles), read together.
func (f *Feed) Map() ([]byte, uint64, [4]int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.png, f.mapV, f.mapR
}

func (f *Feed) poll(ctx context.Context) {
	cctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	// One write for all reads (docs/d2d-pace.md Prefetch); each typed call below finds its reply.
	_ = f.c.Prefetch(cctx, "state", "monsters", "debug level", "clock")
	defer f.c.DropPrefetch()
	s := Snap{Seq: f.seq + 1, Mode: f.mux.Mode()}
	f.seq++
	f.mu.RLock()
	s.MapVer, s.MapRect, s.Stitch = f.snap.MapVer, f.snap.MapRect, f.snap.Stitch // kept over a loading screen: the pane keeps its map
	f.mu.RUnlock()
	s.Tap = TapInfo{f.mux.steps.Load(), f.mux.ticks.Load(), f.mux.viewOps.Load(), f.mux.botLines.Load()}
	st, _, err := f.c.State(cctx)
	if err != nil {
		s.Err = err.Error()
		f.put(s, nil)
		return
	}
	s.Name, s.Class, s.CLvl, s.XP, s.Gold = st.Name, st.Class, st.Level, st.Exp, st.Gold
	s.Life, s.Mana = [2]int{st.Life.Cur, st.Life.Max}, [2]int{st.Mana.Cur, st.Mana.Max}
	ds, werr := f.w.Update()
	if werr == nil {
		s.OK = true
		s.Level, s.Area = ds.LevelNo, areaName(ds.LevelNo)
		s.X, s.Y = float64(ds.Pos.X)/5, float64(ds.Pos.Y)/5
	} else {
		s.Err = werr.Error()
	}
	if ms, e := f.c.Monsters(cctx); e == nil {
		for _, m := range ms.Items {
			mo := Mon{X: m.CellX, Y: m.CellY, Name: m.Name, HP: m.HP.Cur, Mx: m.HP.Max, Boss: m.IsBoss(), Dead: m.Dead()}
			if m.Dead() && !f.seen[m.Unit] {
				s.Dead++
			}
			f.seen[m.Unit] = true
			if mo.Dead && (len(s.Mons) > 150 || abs(mo.X-s.X) > 30 || abs(mo.Y-s.Y) > 30) {
				continue // corpses far away are noise on the map
			}
			s.Mons = append(s.Mons, mo)
		}
	}
	if len(f.seen) > 5000 {
		f.seen = map[int]bool{}
	}
	if ci, e := f.c.Clock(cctx); e == nil {
		s.GameMS, s.Tick, s.Lockstep = ci.GameMS, ci.Tick, ci.Lockstep
		s.Speed = f.speed(ci.GameMS)
	}
	f.mu.RLock()
	prev := f.snap.Dead
	f.mu.RUnlock()
	s.Dead += prev
	if werr == nil {
		f.stitch(&s, f.w.Current(), f.w.Level)
	}
	s.Bot = f.botStatus()
	if f.fr != nil {
		fr := f.fr.Current()
		s.FrameTicks, s.FrameMS, s.Frames = fr.Ticks, fr.WallMS, fr.Seq
	}
	f.put(s, nil)
}

// stitch moves the snapshot into the act frame and refreshes the stitched map. One frame for the pane: the act's tiles
// (level origin added), so a crossing into a level the game draws on the same layer moves nothing on screen and the
// map keeps every explored level of the layer; a level of its own layer (a cave) is a map of its own.
func (f *Feed) stitch(s *Snap, lm *d2map.LevelMap, lookup func(int) *d2map.LevelMap) {
	ox, oy := float64(lm.Origin.X)/5, float64(lm.Origin.Y)/5
	s.X, s.Y = s.X+ox, s.Y+oy
	for i := range s.Mons {
		s.Mons[i].X, s.Mons[i].Y = s.Mons[i].X+ox, s.Mons[i].Y+oy
	}
	for _, e := range lm.Exits {
		s.Exits = append(s.Exits, [3]int{e.Dest, (e.Pos.X + lm.Origin.X) / 5, (e.Pos.Y + lm.Origin.Y) / 5})
	}
	s.Stitch = stitchID(lm.No)
	parts := stitchParts(lm.No, lookup)
	if k := stitchKey(parts); k != f.mapK {
		b, r := renderStitch(parts)
		f.mu.Lock()
		f.png, f.mapK, f.mapV, f.mapR = b, k, f.mapV+1, [4]int{r.X, r.Y, r.W, r.H}
		f.mu.Unlock()
	}
	f.mu.RLock()
	s.MapVer, s.MapRect = f.mapV, f.mapR
	f.mu.RUnlock()
}

func (f *Feed) put(s Snap, p []byte) {
	f.mu.Lock()
	f.snap = s
	if p != nil {
		f.png = p
	}
	f.mu.Unlock()
}

// speed is virtual seconds per wall second from game_ms, smoothed; a game_ms that went backwards (d2d restarted) resets it.
func (f *Feed) speed(ms int) float64 {
	now := time.Now()
	if !f.lastWall.IsZero() && ms >= f.lastMS {
		dt := now.Sub(f.lastWall).Seconds()
		if dt > 0 {
			inst := float64(ms-f.lastMS) / 1000 / dt
			if f.ema == 0 {
				f.ema = inst
			} else {
				f.ema += (inst - f.ema) * (1 - 0.85)
			}
		}
	} else {
		f.ema = 0
	}
	f.lastMS, f.lastWall = ms, now
	return f.ema
}

// botStatus reads the bot API status (read-only GET, at most once a second, no token is sent or kept).
func (f *Feed) botStatus() json.RawMessage {
	if f.bot == "" {
		return nil
	}
	if time.Since(f.botAt) < time.Second {
		return f.botJSON
	}
	f.botAt = time.Now()
	cl := http.Client{Timeout: 700 * time.Millisecond}
	resp, err := cl.Get(f.bot + "/api/v1/status")
	if err != nil {
		f.botJSON = nil
		return nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if json.Valid(b) {
		f.botJSON = b
	}
	return f.botJSON
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
