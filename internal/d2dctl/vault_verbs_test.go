package d2dctl

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// A line in the shape of the fork's `ground full` (V: fork test vault_verbs_d2d.py, vault-spike drill).
const vaultLine = "#12 raw=00aabbccdd11 norm=aabbccdd11 at=45.5,30.5 code=lsd q=6 ilvl=30 uid=0 id'd=0 eth=1 sock=2 def=0 dur=24/24 qty=0 pre=0 suf=0 rare=11,22 aff=5.7. set=0 uniq=0 rw=-1 nprops=3 props=7:0:20;9:0:5; children=0"

func TestParseVaultLine(t *testing.T) {
	r, err := ParseVaultLine(vaultLine)
	if err != nil {
		t.Fatal(err)
	}
	if r.UnitID != 12 || r.Norm != "aabbccdd11" || r.Raw != "00aabbccdd11" || !r.HasPos || r.X != 45.5 || r.Y != 30.5 {
		t.Fatalf("head %+v", r)
	}
	if r.Code != "lsd" || r.Quality != 6 || r.ILvl != 30 || r.Identified || !r.Ethereal || r.Sockets != 2 || r.Dur != 24 || r.MaxDur != 24 ||
		r.Rare1 != 11 || r.Rare2 != 22 || len(r.Affixes) != 2 || r.Affixes[1] != 7 || r.RunewordID != -1 || r.NProps != 3 || r.Props != "7:0:20;9:0:5;" {
		t.Fatalf("view %+v", r)
	}
	for _, bad := range []string{"", "no unit", "#1 raw=aa", "#x raw=aa norm=bb", "#1 raw=aa norm=bb at=1"} {
		if _, err := ParseVaultLine(bad); err == nil {
			t.Fatalf("%q parsed", bad)
		}
	}
	if FormatVaultLine(r) != vaultLine {
		t.Fatalf("format round trip:\n%s\n%s", FormatVaultLine(r), vaultLine)
	}
}

func TestVaultHashAndKey(t *testing.T) {
	a, _ := ParseVaultLine(vaultLine)
	b := a
	b.UnitID = 13
	if a.Hash() != b.Hash() || len(a.Hash()) != 64 {
		t.Fatalf("hash %s %s", a.Hash(), b.Hash())
	}
	if a.DedupeKey("g1") == b.DedupeKey("g1") {
		t.Fatal("two floor units of identical content share a key")
	}
	if a.DedupeKey("g1") == a.DedupeKey("g2") || strings.ContainsAny(a.DedupeKey("g1"), " \t") {
		t.Fatalf("key %q", a.DedupeKey("g1"))
	}
	c := a
	c.Norm = "aabbccdd22"
	if a.Hash() == c.Hash() {
		t.Fatal("different bytes, same hash")
	}
}

func TestVaultWireVerbs(t *testing.T) {
	ctx := context.Background()
	s, c := startStub(t, map[string]string{
		"caps":                                   "caps=state.caps,item.equip,ground.full,ground.take,item.full,item.create,item.vault,vault.stats\nok\n",
		"ground full":                            vaultLine + "\n" + strings.Replace(vaultLine, "#12", "#13", 1) + "\nok\n",
		"ground full 12":                         vaultLine + "\nok\n",
		"ground take 12":                         vaultLine + "\nok\n",
		"ground take 99":                         "err item no such ground item\n",
		"item full":                              strings.Replace(vaultLine, " at=45.5,30.5", "", 1) + "\nok\n",
		"item vault 12":                          strings.Replace(vaultLine, " at=45.5,30.5", "", 1) + "\nok\n",
		"item vault 7":                           "err place worn: take it off first\n",
		"item create aabbccdd11 stash":           "create id=101 at=0,0\nok\n",
		"item create aabbccdd11 inv 3 1 key=k-1": "create id=102 at=3,1\nok\n",
		"item create aabbccdd22 stash":           "err nospace no room there\n",
		"item create aabbccdd33 stash":           "err version 154\n",
		"item create aabbccdd44 stash":           "err dup key already created in this game\n",
		"item create aabbccdd55 stash":           "err state hero is dead\n",
		"item create aabbccdd66 stash":           "err vault verbs off: item create needs d2d --vault-verbs\n",
		"item create aabbccdd77 stash":           "err not in a game\n",
		"item create aabbccdd88 stash":           "err unknown verb 'item'\n",
		"item create aabbccdd99 stash":           "err parse d2s items: read past end\n",
		"vault stats":                            "vault creates=3 takes=2 vaults=1\nok\n",
	}, nil)
	caps, err := c.VaultCaps(ctx)
	if err != nil || !caps.Complete() {
		t.Fatalf("caps %+v %v", caps, err)
	}
	recs, err := c.GroundFull(ctx, -1)
	if err != nil || len(recs) != 2 || recs[1].UnitID != 13 {
		t.Fatalf("ground full %v %v", recs, err)
	}
	if r, err := c.GroundTake(ctx, 12); err != nil || r.UnitID != 12 || !r.HasPos {
		t.Fatalf("take %+v %v", r, err)
	}
	if _, err := c.GroundTake(ctx, 99); !errors.Is(err, ErrVaultNoItem) {
		t.Fatalf("take 99: %v", err)
	}
	if recs, err = c.ItemFull(ctx, -1); err != nil || len(recs) != 1 || recs[0].HasPos {
		t.Fatalf("item full %v %v", recs, err)
	}
	if r, err := c.ItemVault(ctx, 12); err != nil || r.Norm != "aabbccdd11" {
		t.Fatalf("vault %+v %v", r, err)
	}
	var ref *VaultRefused
	if _, err := c.ItemVault(ctx, 7); !errors.As(err, &ref) || ref.Code != VaultErrPlace || !ref.RuleRefusal() {
		t.Fatalf("vault worn: %v", err)
	}
	if r, err := c.ItemCreate(ctx, "aabbccdd11", CreateOpts{}); err != nil || r.ID != 101 || r.Col != 0 || r.Row != 0 {
		t.Fatalf("create %+v %v", r, err)
	}
	if r, err := c.ItemCreate(ctx, "aabbccdd11", CreateOpts{Pack: true, Fixed: true, Col: 3, Row: 1, Key: "k-1"}); err != nil || r.ID != 102 || r.Col != 3 || r.Row != 1 {
		t.Fatalf("create at %+v %v", r, err)
	}
	// error mapping
	for hexv, want := range map[string]error{
		"aabbccdd22": ErrStashFull, "aabbccdd33": ErrVaultVersion, "aabbccdd44": ErrVaultDup, "aabbccdd55": ErrVaultState,
		"aabbccdd66": ErrVaultOff, "aabbccdd77": ErrVaultState, "aabbccdd88": ErrVaultUnsupported, "aabbccdd99": ErrBadHex,
	} {
		if _, err := c.ItemCreate(ctx, hexv, CreateOpts{}); !errors.Is(err, want) {
			t.Fatalf("%s: got %v want %v", hexv, err, want)
		}
	}
	// a request the client refuses without a round trip
	n := len(s.seen)
	for _, bad := range []string{"", "zz", "abc", "aabbccdd1", "aabb"} {
		if _, err := c.ItemCreate(ctx, bad, CreateOpts{}); !errors.Is(err, ErrBadHex) {
			t.Fatalf("%q: %v", bad, err)
		}
	}
	if _, err := c.ItemCreate(ctx, "aabbccdd11", CreateOpts{Key: "a b"}); err == nil {
		t.Fatal("key with a space went out")
	}
	if len(s.seen) != n {
		t.Fatalf("bad requests reached the wire: %v", s.seen[n:])
	}
	if st, err := c.VaultStats(ctx); err != nil || st != (VaultStats{Creates: 3, Takes: 2, Vaults: 1}) {
		t.Fatalf("stats %+v %v", st, err)
	}
}

func TestVaultCapsStockAndFlagOff(t *testing.T) {
	ctx := context.Background()
	_, c := startStub(t, map[string]string{"caps": "err unknown verb caps\n"}, nil)
	if caps, err := c.VaultCaps(ctx); err != nil || caps.Complete() || caps != (VaultCaps{}) {
		t.Fatalf("stock %+v %v", caps, err)
	}
	// a fork started without --vault-verbs lists its item verbs but none of the vault ones
	_, c = startStub(t, map[string]string{"caps": "caps=state.caps,item.equip,item.unequip\nok\n"}, nil)
	if caps, err := c.VaultCaps(ctx); err != nil || caps != (VaultCaps{}) {
		t.Fatalf("flag off %+v %v", caps, err)
	}
	// partial: create without take is not a complete keep flow
	_, c = startStub(t, map[string]string{"caps": "caps=ground.full,item.create\nok\n"}, nil)
	if caps, _ := c.VaultCaps(ctx); caps.Complete() || !caps.ItemCreate {
		t.Fatalf("partial %+v", caps)
	}
}

func TestParseVaultLinePicAndTint(t *testing.T) {
	r, err := ParseVaultLine("#1 raw=aa norm=aa at=1.1,2.2 code=jew q=7 ilvl=30 uid=0 id'd=1 eth=0 sock=0 def=-1 dur=0/0 qty=-1 pre=0 suf=0 rare=0,0 aff=0. set=-1 uniq=4 rw=-1 nprops=0 props= children=0 inv_pic=2 inv_tint=1 inv_colour=5")
	if err != nil {
		t.Fatal(err)
	}
	if r.Pic != 3 || r.Tint != "t1c5" || r.FileIndex() != 4 {
		t.Fatalf("%+v", r)
	}
}
