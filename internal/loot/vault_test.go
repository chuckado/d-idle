package loot

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/lootrules"
	"d2lod-bot/internal/vault"
)

// These tests run on the synthetic item table (no game files): the vault path is the looter's own code.

type vClock struct{ t time.Time }

func (c *vClock) Now() time.Time        { return c.t }
func (c *vClock) Sleep(d time.Duration) { c.t = c.t.Add(d) }

type vMover struct{ n int }

func (m *vMover) GoTo(context.Context, input.Point) error { m.n++; return nil }

type vaultRig struct {
	l       *Looter
	snap    *d2.Snapshot
	in      *input.Fake
	mv      *vMover
	api     *d2dctl.VaultFake
	st      *vault.Store
	notices []string
	db      interface{}
}

func newVaultRig(t *testing.T, class d2.Class, capN int) *vaultRig {
	t.Helper()
	l, _ := synthLooter(t)
	v := &vaultRig{l: l, in: input.NewFake(input.Point{}), mv: &vMover{}, api: d2dctl.NewVaultFake(class.String())}
	l.mv = v.mv
	l.e.In, l.e.Clock = v.in, &vClock{time.Unix(1000, 0)}
	l.e.Tune.Loot.Vault = 1
	v.snap = &d2.Snapshot{State: d2.StateInGame, Inventory: &d2.Inventory{},
		Player: &d2.Player{Class: class, Level: 12, Gold: 5000, Strength: 25, Dexterity: 25, Area: 2, X: 100, Y: 100}}
	l.UseSource(func() (*d2.Snapshot, error) { return v.snap, nil }, nil)
	var err error
	if v.st, err = vault.Open(t.TempDir(), "hero", vault.Options{Capacity: func() int { return capN }, NoSync: true, AutoTrim: func() bool { return vaultTestAutoTrim }}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { v.st.Close() })
	l.UseVault(&VaultWiring{API: v.api, Store: func() (VaultRecorder, error) { return v.st, nil },
		Notice: func(s string) { v.notices = append(v.notices, s) }})
	l.SetVaultSeams(nil, func() bool { return false })
	return v
}

var nextUnit uint32 = 500

// vaultTestAutoTrim is loot.vaultAutoTrim as the rig's store sees it (default off, like the knob).
var vaultTestAutoTrim bool

// A full vault removes nothing and records nothing new (decision 2026-11-21); loot.vaultAutoTrim brings the trim back.
func TestVaultFullDoesNotTrimUnlessAutoTrim(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, auto := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/auto%v", class, auto), func(t *testing.T) {
				vaultTestAutoTrim = auto
				t.Cleanup(func() { vaultTestAutoTrim = false })
				v := newVaultRig(t, class, 2)
				v.drop(t, "rin", d2.QualityMagic, 110, 100, "a")
				v.drop(t, "rin", d2.QualityMagic, 111, 101, "b")
				v.pick(t)
				if v.st.Status().Count != 2 {
					t.Fatalf("seed: %d", v.st.Status().Count)
				}
				before := v.st.Dump()
				v.drop(t, "rin", d2.QualityRare, 112, 102, "worth")
				v.pick(t)
				arch, _ := v.st.Archived()
				if auto {
					if len(arch) != 1 {
						t.Fatalf("knob on: the old trim must run, archive %d", len(arch))
					}
					return
				}
				if len(arch) != 0 || v.st.Status().Count != 2 || v.clicks() != 0 {
					t.Fatalf("knob off: removed %d, count %d, clicks %d", len(arch), v.st.Status().Count, v.clicks())
				}
				for i, r := range v.st.Dump() {
					if r.ID != before[i].ID || r.State != vault.Recorded {
						t.Fatalf("record %d changed: %+v", i, r)
					}
				}
				if m := v.st.MissedDrops(); m.Count != 1 {
					t.Fatalf("missed %+v", m)
				}
				v.pick(t) // the same unit is not counted twice
				if m := v.st.MissedDrops(); m.Count != 1 {
					t.Fatalf("recounted %+v", m)
				}
			})
		}
	}
}

// drop puts an item on the floor of the snapshot and its twin on the fork's floor, so the peek finds it.
func (v *vaultRig) drop(t *testing.T, code string, q int, x, y int, extra string) d2.Unit {
	t.Helper()
	nextUnit++
	db := v.l.e.Data
	u := synthUnit(t, db, nextUnit, code, q, d2.FlagIdentified, d2.ContGround)
	u.Mode, u.X, u.Y = d2.ItemModeGround, x, y
	v.snap.Ground = append(v.snap.Ground, u)
	v.api.Floor = append(v.api.Floor, d2dctl.FakeRecord(int(u.ID), code, q, 20, extra))
	return u
}

func (v *vaultRig) clicks() int {
	n := 0
	for _, c := range v.in.Calls {
		if strings.HasPrefix(c, "interact") {
			n++
		}
	}
	return n
}

func (v *vaultRig) pick(t *testing.T) PickReport {
	t.Helper()
	rep, err := v.l.PickUp(context.Background(), PickOpts{})
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

var vaultClasses = []d2.Class{d2.Sorceress, d2.Amazon}

func TestVaultRecordsWithoutPickupClick(t *testing.T) {
	for _, cls := range vaultClasses {
		v := newVaultRig(t, cls, 10)
		u := v.drop(t, "r01", d2.QualityNormal, 110, 100, "a")
		rep := v.pick(t)
		if len(rep.Vaulted) != 1 || len(rep.Picked) != 0 || rep.Vaulted[0].ID != u.ID {
			t.Fatalf("%v: vaulted %+v picked %+v skipped %+v failed %+v", cls, rep.Vaulted, rep.Picked, rep.Skipped, rep.Failed)
		}
		if n := v.clicks(); n != 0 {
			t.Errorf("%v: %d pickup clicks for a vaulted item", cls, n)
		}
		if v.mv.n != 1 {
			t.Errorf("%v: walked %d times, want the one walk of a pickup", cls, v.mv.n)
		}
		if st := v.st.Status(); st.Count != 1 {
			t.Errorf("%v: store %+v", cls, st)
		}
		if s, _ := v.api.VaultStats(context.Background()); s.Takes != 0 || s.Creates != 0 {
			t.Errorf("%v: V3 only peeks: %+v", cls, s)
		}
		// a second visit does not walk or peek again
		if rep = v.pick(t); len(rep.Vaulted) != 0 || v.mv.n != 1 || v.st.Status().Count != 1 {
			t.Errorf("%v: second call vaulted %d walks %d", cls, len(rep.Vaulted), v.mv.n)
		}
		if info := v.l.VaultInfo(); !info.Active || info.Recorded != 1 {
			t.Errorf("%v: info %+v", cls, info)
		}
	}
}

func TestVaultKeepsKitRules(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	pot := v.drop(t, "hp1", d2.QualityNormal, 102, 100, "p")
	v.drop(t, "r33", d2.QualityNormal, 103, 100, "b")
	rep := v.pick(t)
	if len(rep.Vaulted) != 1 || v.clicks() == 0 || v.st.Status().Count != 1 {
		t.Fatalf("picked %+v vaulted %+v clicks %d", rep.Picked, rep.Vaulted, v.clicks())
	}
	for _, r := range rep.Vaulted {
		if r.ID == pot.ID {
			t.Error("the potion was vaulted: potions keep their own rule")
		}
	}
}

func TestVaultDedupe(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	v.drop(t, "r01", d2.QualityNormal, 110, 100, "same")
	v.drop(t, "r01", d2.QualityNormal, 111, 101, "same") // identical bytes: one record
	v.drop(t, "r33", d2.QualityNormal, 112, 102, "other")
	rep := v.pick(t)
	info := v.l.VaultInfo()
	if len(rep.Vaulted) != 2 || info.Dups != 1 || v.st.Status().Count != 2 || v.clicks() != 0 {
		t.Fatalf("vaulted %d info %+v live %d clicks %d", len(rep.Vaulted), info, v.st.Status().Count, v.clicks())
	}
}

func TestVaultOverCapacityStopsNewRecords(t *testing.T) {
	v := newVaultRig(t, d2.Amazon, 1)
	v.drop(t, "r01", d2.QualityNormal, 110, 100, "a")
	v.drop(t, "r33", d2.QualityNormal, 111, 101, "b")
	rep := v.pick(t)
	if len(rep.Vaulted) != 1 || v.st.Status().Count != 1 {
		t.Fatalf("vaulted %d live %d", len(rep.Vaulted), v.st.Status().Count)
	}
	info := v.l.VaultInfo()
	if !info.Blocked || len(v.notices) != 1 || !strings.Contains(v.notices[0], "full") {
		t.Fatalf("info %+v notices %v", info, v.notices)
	}
	// still blocked: the item is skipped with a visible reason; nothing is picked up, walked to or peeked
	walks := v.mv.n
	rep = v.pick(t)
	why := ""
	for _, sk := range rep.Skipped {
		why += sk.Why
	}
	if len(rep.Picked) != 0 || !strings.Contains(why, "vault full") || v.clicks() != 0 || len(v.notices) != 1 || v.mv.n != walks {
		t.Errorf("picked %d skipped %q clicks %d notices %v walks %d->%d", len(rep.Picked), why, v.clicks(), v.notices, walks, v.mv.n)
	}
	// time passes: one more insert is tried (and refused again, with no second notice)
	v.l.SetVaultSeams(func() time.Time { return time.Now().Add(time.Hour) }, func() bool { return false })
	if rep = v.pick(t); len(rep.Vaulted) != 0 || len(v.notices) != 1 {
		t.Errorf("retry: vaulted %d notices %v", len(rep.Vaulted), v.notices)
	}
}

func TestVaultPeekFailureFallsBackToPickup(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	v.api.NoGame = true // the fork refuses the peek: "not in a game"
	v.drop(t, "r01", d2.QualityNormal, 102, 100, "a")
	rep := v.pick(t)
	// today's pickup decision runs: a click is sent (the sim floor never lets go of the unit, so it ends as a failure)
	if len(rep.Vaulted) != 0 || v.clicks() == 0 || v.st.Status().Count != 0 {
		t.Fatalf("picked %+v vaulted %+v clicks %d", rep.Picked, rep.Vaulted, v.clicks())
	}
	if info := v.l.VaultInfo(); info.Fallback != 1 || !info.Active {
		t.Errorf("info %+v", info)
	}
}

func TestVaultVerbsAbsentIsInert(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	v.api.FlagOff = true // d2d without --vault-verbs: caps list nothing
	v.drop(t, "r01", d2.QualityNormal, 102, 100, "a")
	for i := 0; i < 2; i++ {
		if rep := v.pick(t); len(rep.Vaulted) != 0 {
			t.Fatalf("vaulted %+v", rep.Vaulted)
		}
	}
	info := v.l.VaultInfo()
	if info.Active || info.Inert != "verbs absent" || len(v.notices) != 1 || v.clicks() == 0 {
		t.Errorf("info %+v notices %v clicks %d", info, v.notices, v.clicks())
	}
}

func TestVaultForcedOffInMeasuredRun(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	v.l.SetVaultSeams(nil, func() bool { return true }) // D2BOT_RUN set
	v.drop(t, "r01", d2.QualityNormal, 102, 100, "a")
	rep := v.pick(t)
	if len(rep.Vaulted) != 0 || v.clicks() == 0 || v.st.Status().Count != 0 {
		t.Fatalf("picked %+v vaulted %+v", rep.Picked, rep.Vaulted)
	}
	if info := v.l.VaultInfo(); info.Active || info.Inert != "measured run" {
		t.Errorf("info %+v", info)
	}
}

func TestVaultKnobZeroIsToday(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	v.l.e.Tune.Loot.Vault = 0
	v.drop(t, "r01", d2.QualityNormal, 102, 100, "a")
	rep := v.pick(t)
	if len(rep.Vaulted) != 0 || v.clicks() == 0 || v.l.VaultInfo().Active {
		t.Fatalf("picked %+v vaulted %+v", rep.Picked, rep.Vaulted)
	}
	v.l.e.Tune.Bot.LootShadow = 0
	if p := v.l.ShadowPass(v.snap, 0, 0); p != nil {
		t.Error("knob 0, shadow 0, stage 0 must give no pass")
	}
}

func TestVaultRecordAllToggle(t *testing.T) {
	for _, all := range []int{0, 1} {
		v := newVaultRig(t, d2.Amazon, 10)
		v.l.e.Tune.Loot.VaultAll = all
		v.drop(t, "rin", d2.QualityNormal, 102, 100, "plain") // plain ring: left on the floor unless everything is recorded
		rep := v.pick(t)
		if got := len(rep.Vaulted); got != all || len(rep.Picked) != 0 || v.clicks() != 0 {
			t.Errorf("vaultAll %d: vaulted %d picked %d clicks %d", all, got, len(rep.Picked), v.clicks())
		}
	}
}

// The shadow expects the vault action only with loot.vault 1; with 0 a vault rule folds to ignore.
func TestVaultShadowFold(t *testing.T) {
	v := newVaultRig(t, d2.Sorceress, 10)
	d := lootrules.Decision{Action: lootrules.Vault, Prio: 90, Tier: "S"}
	p := v.l.ShadowPass(v.snap, 0, 0)
	if p == nil || !p.vault || p.groundFold(d) != lootrules.Vault {
		t.Fatalf("vault on: pass %+v", p)
	}
	v.l.e.Tune.Loot.Vault = 0
	v.l.e.Tune.Bot.LootShadow = 1
	if p = v.l.ShadowPass(v.snap, 0, 0); p == nil || p.vault || p.groundFold(d) != lootrules.Ignore {
		t.Fatalf("vault off: pass %+v", p)
	}
}

// Poor mode (docs/vault.md, gold flow): below loot.poorGoldBelow gold an item worth loot.poorMinSell is picked up to be sold
// instead of only recorded; a hero with enough gold, or the knob at 0, still records it.
func TestVaultPoorModePicksSellableItems(t *testing.T) {
	for _, cls := range vaultClasses {
		for _, tc := range []struct {
			name         string
			gold, knob   int
			wantVaulted  bool
			wantSellPick bool
		}{
			{"broke", 0, 1, false, true},
			{"rich", 5000, 1, true, false},
			{"broke but knob off", 0, 0, true, false},
		} {
			t.Run(cls.String()+"/"+tc.name, func(t *testing.T) {
				v := newVaultRig(t, cls, 10)
				v.snap.Player.Gold = tc.gold
				v.l.e.Tune.Loot.VaultPoorPickup, v.l.e.Tune.Loot.VaultAll = tc.knob, 1
				u := v.drop(t, "rin", d2.QualityRare, 110, 100, "r")
				rep := v.pick(t)
				vaulted := len(rep.Vaulted) == 1 && rep.Vaulted[0].ID == u.ID
				// the rig has no pickup physics: a pickup attempt is a click and a "still on the ground" failure
				picked := v.clicks() > 0 && len(rep.Failed) == 1 && rep.Failed[0].ID == u.ID
				if vaulted != tc.wantVaulted || picked != tc.wantSellPick {
					t.Fatalf("vaulted %v picked %v; report %+v skipped %+v", vaulted, picked, rep, rep.Skipped)
				}
				if tc.wantSellPick && v.st.Status().Count != 0 {
					t.Error("a picked item must not also be recorded")
				}
			})
		}
	}
}

// A drop whose peek carries no affixes is recorded with the ones its own bytes carry (item decode, a read): no record is left
// that only the in-game identify can reveal. A fork without the verb records it as before.
func TestVaultRecordFillsSealedFromDecode(t *testing.T) {
	for _, cls := range vaultClasses {
		v := newVaultRig(t, cls, 10)
		v.drop(t, "r01", d2.QualityMagic, 110, 100, "a")
		norm := v.api.Floor[0].Norm
		v.api.Bytes = map[string]d2dctl.VaultRecord{norm: {Code: "r01", Props: "7:0:20;93:0:10;"}}
		rep := v.pick(t)
		if len(rep.Vaulted) != 1 {
			t.Fatalf("%v: %+v", cls, rep)
		}
		recs := v.st.Dump()
		if len(recs) != 1 || recs[0].Sealed != "7:0:20;93:0:10;" || recs[0].Identified || recs[0].Props != "" {
			t.Fatalf("%v: held, not shown: %+v", cls, recs)
		}
		if v.api.Decodes != 1 {
			t.Errorf("%v: decodes %d", cls, v.api.Decodes)
		}
	}
}
