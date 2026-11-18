package d2dctl

import (
	"context"
	"errors"
	"testing"
)

// The keep flow on two hero classes (the client is class-agnostic; the fixtures differ in class, grid size and what lies on the floor).
var vaultClasses = []struct {
	class                string
	stashCols, stashRows int
	code                 string
	w, h                 int
}{
	{"sorceress", 6, 8, "lsd", 1, 3},
	{"paladin", 6, 4, "plt", 2, 3},
}

func fixtureFake(class string, cols, rows int, code string, w, h int) *VaultFake {
	f := NewVaultFake(class)
	f.StashCols, f.StashRows = cols, rows
	f.Sizes = map[string][2]int{code: {w, h}}
	for i := 0; i < 3; i++ {
		r := FakeRecord(10+i, code, 6, 30, "roll-"+string(rune('a'+i)))
		r.X, r.Y = float64(40+i), 30
		f.Floor = append(f.Floor, r)
	}
	// a second identical drop (same content, other unit): a legitimate duplicate
	twin := f.Floor[0]
	twin.UnitID = 20
	f.Floor = append(f.Floor, twin)
	f.Gold = map[int]bool{30: true}
	f.Floor = append(f.Floor, VaultRecord{UnitID: 30, Code: "gld", Norm: "aabbccdd", Raw: "aabbccdd"})
	return f
}

func TestVaultKeepFlowTwoClasses(t *testing.T) {
	ctx := context.Background()
	for _, k := range vaultClasses {
		t.Run(k.class, func(t *testing.T) {
			f := fixtureFake(k.class, k.stashCols, k.stashRows, k.code, k.w, k.h)
			var api VaultAPI = f
			d := &Dedupe{}
			caps, _ := api.VaultCaps(ctx)
			if !caps.Complete() {
				t.Fatal("caps")
			}
			floor, err := api.GroundFull(ctx, -1)
			if err != nil || len(floor) != 4 { // gold is not listed
				t.Fatalf("floor %d %v", len(floor), err)
			}
			if _, err := api.GroundTake(ctx, 30); !isCode(err, VaultErrGold) {
				t.Fatalf("gold: %v", err)
			}
			// peek, create, then take: the order of the crash-safe flow (design 1.2); the record is durable first
			rec := floor[0]
			res, err := CreateOnce(ctx, api, d, rec, "g1", CreateOpts{})
			if err != nil || res.ID == 0 {
				t.Fatalf("create %+v %v", res, err)
			}
			took, err := api.GroundTake(ctx, rec.UnitID)
			if err != nil || took.Norm != rec.Norm {
				t.Fatalf("take %+v %v", took, err)
			}
			// the same record again (a retried keep): refused, nothing created
			if _, err := CreateOnce(ctx, api, d, rec, "g1", CreateOpts{}); !errors.Is(err, ErrVaultDup) {
				t.Fatalf("retry: %v", err)
			}
			// a fresh Dedupe (the bot restarted) still hits the fork's key set
			if _, err := CreateOnce(ctx, api, &Dedupe{}, rec, "g1", CreateOpts{}); !errors.Is(err, ErrVaultDup) {
				t.Fatalf("restart: %v", err)
			}
			// identical content from another floor unit is its own record
			twin := floor[3]
			if twin.Norm != rec.Norm {
				t.Fatal("fixture twin differs")
			}
			if _, err := CreateOnce(ctx, api, d, twin, "g1", CreateOpts{}); err != nil {
				t.Fatalf("twin: %v", err)
			}
			if st, _ := api.VaultStats(ctx); st.Creates != 2 || st.Takes != 1 {
				t.Fatalf("stats %+v", st)
			}
			// byte equality: what ItemFull returns is what was captured
			have, _ := api.ItemFull(ctx, res.ID)
			if len(have) != 1 || have[0].Norm != rec.Norm {
				t.Fatalf("created copy %+v", have)
			}
			// deposit side: vault the carried item, create it back
			v, err := api.ItemVault(ctx, res.ID)
			if err != nil || v.Norm != rec.Norm {
				t.Fatalf("vault %+v %v", v, err)
			}
			if _, err := api.ItemVault(ctx, res.ID); !errors.Is(err, ErrVaultNoItem) {
				t.Fatalf("vault twice: %v", err)
			}
			// a full stash: nospace, the claim is given back so a later retry can work
			n := 0
			for ; n < 100; n++ {
				r := FakeRecord(200+n, k.code, 4, 20, "fill-"+string(rune('A'+n%26))+string(rune('a'+n/26)))
				if _, err := CreateOnce(ctx, api, d, r, "g1", CreateOpts{}); err != nil {
					if !errors.Is(err, ErrStashFull) {
						t.Fatalf("fill %d: %v", n, err)
					}
					if !d.Claim(r.DedupeKey("g1")) {
						t.Fatal("a failed create kept its claim")
					}
					break
				}
			}
			if n == 0 || n >= 100 {
				t.Fatalf("stash never filled (%d)", n)
			}
			// the pack is a separate grid
			if _, err := api.ItemCreate(ctx, rec.Norm, CreateOpts{Pack: true}); err != nil {
				t.Fatalf("pack: %v", err)
			}
			if _, err := api.ItemCreate(ctx, rec.Norm, CreateOpts{Fixed: true, Col: 99, Row: 0}); !isCode(err, VaultErrPlace) {
				t.Fatalf("outside the grid: %v", err)
			}
		})
	}
}

func isCode(err error, code string) bool {
	var r *VaultRefused
	return errors.As(err, &r) && r.Code == code
}

func TestVaultFakeStates(t *testing.T) {
	ctx := context.Background()
	mk := func() *VaultFake {
		f := NewVaultFake("amazon")
		f.Floor = []VaultRecord{FakeRecord(1, "rin", 4, 10, "x")}
		f.Carried = []FakeCarried{
			{Rec: FakeRecord(50, "cap", 2, 5, ""), Where: FakeWorn},
			{Rec: FakeRecord(51, "rin", 4, 5, "h"), Where: FakeHand},
			{Rec: FakeRecord(52, "rin", 4, 5, "b"), Where: FakeBelt},
		}
		return f
	}
	f := mk()
	if _, err := f.ItemVault(ctx, 50); !isCode(err, VaultErrPlace) {
		t.Fatalf("worn: %v", err)
	}
	if _, err := f.ItemVault(ctx, 51); !isCode(err, VaultErrCursor) {
		t.Fatalf("hand: %v", err)
	}
	if _, err := f.ItemVault(ctx, 52); err != nil {
		t.Fatalf("belt: %v", err)
	}
	f = mk()
	f.Dead = true
	if _, err := f.GroundTake(ctx, 1); !errors.Is(err, ErrVaultState) {
		t.Fatalf("dead take: %v", err)
	}
	if _, err := f.ItemCreate(ctx, FakeRecord(0, "rin", 4, 1, "").Norm, CreateOpts{}); !errors.Is(err, ErrVaultState) {
		t.Fatalf("dead create: %v", err)
	}
	if _, err := f.ItemFull(ctx, -1); err != nil {
		t.Fatalf("a dead hero can still be read: %v", err)
	}
	f = mk()
	f.Joined = true
	if _, err := f.GroundTake(ctx, 1); !errors.Is(err, ErrVaultState) {
		t.Fatalf("joined: %v", err)
	}
	f = mk()
	f.FlagOff = true
	if c, _ := f.VaultCaps(ctx); c != (VaultCaps{}) {
		t.Fatal("flag off lists caps")
	}
	if _, err := f.GroundFull(ctx, -1); !errors.Is(err, ErrVaultOff) {
		t.Fatalf("flag off: %v", err)
	}
	f = mk()
	f.NoGame = true
	if _, err := f.ItemCreate(ctx, FakeRecord(0, "rin", 4, 1, "").Norm, CreateOpts{}); !errors.Is(err, ErrVaultState) {
		t.Fatalf("no game: %v", err)
	}
	f = mk()
	if _, err := f.ItemCreate(ctx, "aabbccdd", CreateOpts{}); !errors.Is(err, ErrBadHex) {
		t.Fatalf("garbage record: %v", err)
	}
	if _, err := f.ItemCreate(ctx, "zz", CreateOpts{}); !errors.Is(err, ErrBadHex) {
		t.Fatalf("bad hex: %v", err)
	}
}
