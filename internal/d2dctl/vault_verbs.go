package d2dctl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Vault verbs of the private fork (docs/design/loot-vault-2026-11-16.md, probe docs/probes/vault-spike-2026-11-16.md;
// fork branch didle/vault-verbs, flag --vault-verbs, default off). Wire:
//
//	ground full [id]                       #<id> raw=<hex> norm=<hex> at=<x>,<y> <k=v view...>   one line per non-gold floor item
//	ground take <id>                       the same line, then the unit leaves the floor
//	item full [id]                         the same lines for carried items (no at=)
//	item vault <id>                        the same line, then the item leaves the hero
//	item create <hex> [stash|inv] [col row] [key=<token>]    create id=<n> at=<col>,<row>
//	vault stats                            vault creates=N takes=N vaults=N
//
// V (fork test vault_verbs_d2d.py on the builder, 2026-11-16): every shape and refusal below; U: nothing here ran against a
// bot run yet. The record identity is the content: `uid` is 0 on every d2d item (V0 finding), so a record is keyed by
// sha256 of its norm bytes plus where it was found (VaultRecord.DedupeKey); two identical normal items stay legitimate.

// Capability names a --vault-verbs fork build advertises in "caps".
const (
	CapGroundFull  = "ground.full"
	CapGroundTake  = "ground.take"
	CapItemFull    = "item.full"
	CapItemCreate  = "item.create"
	CapItemVault   = "item.vault"
	CapVaultStats  = "vault.stats"
	CapItemDecode  = "item.decode" // read-only `item decode <hex>` (fork branch didle/devp18-20261121)
	vaultWhereStsh = "stash"
	vaultWhereInv  = "inv"
)

// VaultCaps is what a fork build can do with the vault verbs. The zero value (a stock build, an older fork, or a fork
// started without --vault-verbs: `caps` then lists none of them) means "nothing": callers keep the pickup path.
type VaultCaps struct {
	GroundFull, GroundTake, ItemFull, ItemCreate, ItemVault bool
	Decode                                                  bool // `item decode` (read-only); not part of Complete
	PayGold                                                 bool // `debug paygold` (fork flag --vault-pay, ADR 0029); not part of Complete
}

// Complete reports whether every verb of the keep flow is there (peek, take, create, and vault for the deposit side).
func (v VaultCaps) Complete() bool {
	return v.GroundFull && v.GroundTake && v.ItemFull && v.ItemCreate && v.ItemVault
}

// VaultCapsOf reads the vault verbs out of a caps list.
func VaultCapsOf(caps []string) VaultCaps {
	return VaultCaps{
		GroundFull: HasCap(caps, CapGroundFull), GroundTake: HasCap(caps, CapGroundTake), ItemFull: HasCap(caps, CapItemFull),
		ItemCreate: HasCap(caps, CapItemCreate), ItemVault: HasCap(caps, CapItemVault), PayGold: HasCap(caps, CapVaultPaygold), Decode: HasCap(caps, CapItemDecode),
	}
}

// Vault refusal codes (the first word after "err" in the fork's reply).
const (
	VaultErrOff     = "off"     // "vault verbs off": d2d runs without --vault-verbs (mapped from "vault verbs off: ...")
	VaultErrState   = "state"   // the hero is dead, or a joined game (the floor is the host's)
	VaultErrNoGame  = "nogame"  // "not in a game"
	VaultErrItem    = "item"    // no such item (floor or carried)
	VaultErrGold    = "gold"    // gold is not an item record
	VaultErrParse   = "parse"   // bad hex, truncated, trailing bytes
	VaultErrVersion = "version" // an item record of another codec version
	VaultErrDup     = "dup"     // the key (or a non-zero seed and code) is already in the game
	VaultErrNoSpace = "nospace" // no room in the stash or pack
	VaultErrPlace   = "place"   // outside the grid, overlap, worn, socketed child
	VaultErrCursor  = "cursor"  // the item is in the hand
	VaultErrUsage   = "usage"   // a malformed request
)

// Sentinels for errors.Is on a *VaultRefused.
var (
	ErrVaultOff     = errors.New("d2dctl: vault verbs off (d2d runs without --vault-verbs)")
	ErrVaultState   = errors.New("d2dctl: game state refuses the vault verb (dead hero, joined game or no game)")
	ErrStashFull    = errors.New("d2dctl: no room in the stash or pack")
	ErrBadHex       = errors.New("d2dctl: bad item hex")
	ErrVaultDup     = errors.New("d2dctl: record already created in this game")
	ErrVaultNoItem  = errors.New("d2dctl: no such item")
	ErrVaultVersion = errors.New("d2dctl: item record of another codec version")
	// ErrVaultUnsupported: stock d2d or an older fork (the verb is unknown): the caller keeps the pickup path.
	ErrVaultUnsupported = errors.New("d2dctl: the vault verbs are not in this d2d build")
)

// VaultRefused is a vault verb the fork understood and refused. Code is one of the VaultErr* names, Why the fork's text.
type VaultRefused struct{ Code, Why string }

func (e *VaultRefused) Error() string { return "vault verb refused (" + e.Code + "): " + e.Why }

// Is maps the refusal onto the sentinels.
func (e *VaultRefused) Is(target error) bool {
	switch target {
	case ErrVaultOff:
		return e.Code == VaultErrOff
	case ErrVaultState:
		return e.Code == VaultErrState || e.Code == VaultErrNoGame
	case ErrStashFull:
		return e.Code == VaultErrNoSpace
	case ErrBadHex:
		return e.Code == VaultErrParse
	case ErrVaultDup:
		return e.Code == VaultErrDup
	case ErrVaultNoItem:
		return e.Code == VaultErrItem
	case ErrVaultVersion:
		return e.Code == VaultErrVersion
	}
	return false
}

// RuleRefusal marks the error as final (repeating it through another path fails the same way).
func (e *VaultRefused) RuleRefusal() bool { return true }

// mapVaultErr turns a transport or command error into a typed one. Transport errors pass through.
func mapVaultErr(err error) error {
	var ce *CmdError
	if err == nil || !errors.As(err, &ce) || errors.Is(err, ErrBusy) || errors.Is(err, ErrLineTooLong) {
		return err
	}
	msg := ce.Msg
	switch {
	case strings.HasPrefix(msg, "vault verbs off"):
		return &VaultRefused{Code: VaultErrOff, Why: msg}
	case strings.HasPrefix(msg, "unknown"):
		return fmt.Errorf("%w: %s", ErrVaultUnsupported, msg)
	case strings.HasPrefix(msg, "not in a game"):
		return &VaultRefused{Code: VaultErrNoGame, Why: msg}
	case strings.HasPrefix(msg, "item decode <") || strings.HasPrefix(msg, "item create <") || strings.HasPrefix(msg, "ground take <") || strings.HasPrefix(msg, "vault stats"):
		return &VaultRefused{Code: VaultErrUsage, Why: msg}
	}
	code, why, _ := strings.Cut(msg, " ")
	return &VaultRefused{Code: code, Why: why}
}

// VaultRecord is one item as the fork serialises it: the bytes of a save's item record plus a decoded view.
type VaultRecord struct {
	UnitID int
	Raw    string // hex as the item lies (file version 0x60); carries the cell
	Norm   string // hex with the place cleared: the comparison key and what `item create` takes
	X, Y   float64
	HasPos bool // a floor item (ground full / ground take)

	Code                 string
	Quality, ILvl, UID   int
	Identified, Ethereal bool
	Sockets              int
	Defense              int
	Dur, MaxDur          int
	Qty                  int
	Prefix, Suffix       int
	Rare1, Rare2         int
	Affixes              []int
	SetID, UniqueID      int
	RunewordID           int
	NProps, Children     int
	Props                string // "stat:param:value;..." as sent
	Pic                  int    // inv_pic + 1: the item's own picture variant (jewel, ring, amulet, charm), 0 = the base picture
	Tint                 string // "t<set>c<colour>" from inv_tint and inv_colour, "" = none
}

// Hash is sha256 of the norm bytes, hex: the content identity of the record (the seed is 0 on every d2d item).
func (r VaultRecord) Hash() string {
	b, err := hex.DecodeString(r.Norm)
	if err != nil {
		b = []byte(r.Norm)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// DedupeKey is the token that makes a create idempotent: content hash (16 hex) plus the game and the floor unit and
// spot it was found at, so two identical drops of one game stay two records and a retry of one stays one.
func (r VaultRecord) DedupeKey(game string) string {
	k := r.Hash()[:16] + "-" + game + "-" + strconv.Itoa(r.UnitID)
	if r.HasPos {
		k += fmt.Sprintf("-%.1f_%.1f", r.X, r.Y)
	}
	return k
}

// ParseVaultLine parses one `#id raw=.. norm=.. [at=x,y] code=.. ...` line.
func ParseVaultLine(line string) (VaultRecord, error) {
	var r VaultRecord
	if !isUnitTok(strings.TrimSpace(line)) {
		return r, fmt.Errorf("d2dctl: not a vault record line: %q", line)
	}
	fs := strings.Fields(line)
	id, err := unitID(fs[0])
	if err != nil {
		return r, err
	}
	r.UnitID = id
	kv := ParseKV(strings.Join(fs[1:], " "))
	r.Raw, _ = kv.Get("raw")
	r.Norm, _ = kv.Get("norm")
	if r.Norm == "" || r.Raw == "" {
		return r, fmt.Errorf("d2dctl: vault line without raw/norm: %q", line)
	}
	if p, ok := kv.Get("at"); ok {
		a, b, _ := strings.Cut(p, ",")
		x, e1 := atof(a)
		y, e2 := atof(b)
		if e1 != nil || e2 != nil {
			return r, fmt.Errorf("d2dctl: bad at= in %q", line)
		}
		r.X, r.Y, r.HasPos = x, y, true
	}
	r.Code, _ = kv.Get("code")
	r.Quality, _ = kv.Int("q")
	r.ILvl, _ = kv.Int("ilvl")
	r.UID, _ = kv.Int("uid")
	v, _ := kv.Int("id'd")
	r.Identified = v == 1
	v, _ = kv.Int("eth")
	r.Ethereal = v == 1
	r.Sockets, _ = kv.Int("sock")
	r.Defense, _ = kv.Int("def")
	if p, ok := kv.Pair("dur"); ok {
		r.Dur, r.MaxDur = p.Cur, p.Max
	}
	r.Qty, _ = kv.Int("qty")
	r.Prefix, _ = kv.Int("pre")
	r.Suffix, _ = kv.Int("suf")
	if rare, ok := kv.Get("rare"); ok {
		a, b, _ := strings.Cut(rare, ",")
		r.Rare1, _ = strconv.Atoi(a)
		r.Rare2, _ = strconv.Atoi(b)
	}
	if aff, ok := kv.Get("aff"); ok {
		for _, s := range strings.Split(aff, ".") {
			if n, err := strconv.Atoi(s); err == nil {
				r.Affixes = append(r.Affixes, n)
			}
		}
	}
	r.SetID, _ = kv.Int("set")
	r.UniqueID, _ = kv.Int("uniq")
	r.RunewordID, _ = kv.Int("rw")
	r.NProps, _ = kv.Int("nprops")
	r.Props, _ = kv.Get("props")
	if v, ok := kv.Int("inv_pic"); ok && v >= 0 {
		r.Pic = v + 1
	}
	if t, ok := kv.Int("inv_tint"); ok && t > 0 {
		if c, ok := kv.Int("inv_colour"); ok && c >= 0 {
			r.Tint = fmt.Sprintf("t%dc%d", t, c)
		}
	}
	r.Children, _ = kv.Int("children")
	return r, nil
}

func parseVaultLines(r *Reply) ([]VaultRecord, error) {
	var out []VaultRecord
	for _, l := range r.Lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		rec, err := ParseVaultLine(l)
		if err != nil {
			return out, err
		}
		out = append(out, rec)
	}
	return out, nil
}

// VaultAPI is the vault surface the bot codes against: *Client talks to a fork, *VaultFake is the in-memory stand-in.
type VaultAPI interface {
	GroundFull(ctx context.Context, id int) ([]VaultRecord, error)
	GroundTake(ctx context.Context, id int) (VaultRecord, error)
	ItemFull(ctx context.Context, id int) ([]VaultRecord, error)
	ItemVault(ctx context.Context, id int) (VaultRecord, error)
	ItemCreate(ctx context.Context, hex string, o CreateOpts) (CreateResult, error)
	VaultStats(ctx context.Context) (VaultStats, error)
	VaultCaps(ctx context.Context) (VaultCaps, error)
}

// CreateOpts places a created item. The zero value is the stash, first free spot, no key.
type CreateOpts struct {
	Pack     bool   // the backpack instead of the stash
	Col, Row int    // a fixed cell; used only when Fixed
	Fixed    bool   // Col and Row are set
	Key      string // idempotency token (VaultRecord.DedupeKey): the fork refuses a second create with it
}

// CreateResult is a created item: its new unit id and cell.
type CreateResult struct{ ID, Col, Row int }

// VaultStats is the fork's ledger for this d2d process.
type VaultStats struct{ Creates, Takes, Vaults int }

// ValidHex checks the hex text an `item create` takes (even length, hex digits, at least 4 bytes, one line).
func ValidHex(s string) error {
	if len(s) < 8 || len(s)%2 != 0 || len(s) >= MaxLine-64 {
		return &VaultRefused{Code: VaultErrParse, Why: "odd, short or oversize hex"}
	}
	if _, err := hex.DecodeString(s); err != nil {
		return &VaultRefused{Code: VaultErrParse, Why: "not hex"}
	}
	return nil
}

func (c *Client) vaultRecords(ctx context.Context, line string) ([]VaultRecord, error) {
	r, err := c.Do(ctx, line)
	if err != nil {
		return nil, mapVaultErr(err)
	}
	return parseVaultLines(r)
}

func (c *Client) vaultOne(ctx context.Context, line string, id int) (VaultRecord, error) {
	recs, err := c.vaultRecords(ctx, line)
	if err != nil {
		return VaultRecord{}, err
	}
	for _, rec := range recs {
		if rec.UnitID == id {
			return rec, nil
		}
	}
	return VaultRecord{}, &VaultRefused{Code: VaultErrItem, Why: fmt.Sprintf("reply holds no record for unit %d", id)}
}

// GroundFull captures floor item id (id < 0: every non-gold floor item).
func (c *Client) GroundFull(ctx context.Context, id int) ([]VaultRecord, error) {
	if id < 0 {
		return c.vaultRecords(ctx, "ground full")
	}
	return c.vaultRecords(ctx, fmt.Sprintf("ground full %d", id))
}

// GroundTake captures floor item id and removes the unit from the floor in the same tick. Record it (durably) BEFORE
// taking, or take only what a prior GroundFull returned: after the take the floor no longer has it.
func (c *Client) GroundTake(ctx context.Context, id int) (VaultRecord, error) {
	return c.vaultOne(ctx, fmt.Sprintf("ground take %d", id), id)
}

// ItemFull captures carried item id (id < 0: every carried item).
func (c *Client) ItemFull(ctx context.Context, id int) ([]VaultRecord, error) {
	if id < 0 {
		return c.vaultRecords(ctx, "item full")
	}
	return c.vaultRecords(ctx, fmt.Sprintf("item full %d", id))
}

// ItemVault captures a carried grid or belt item and removes it from the hero (worn items refuse with place).
func (c *Client) ItemVault(ctx context.Context, id int) (VaultRecord, error) {
	return c.vaultOne(ctx, fmt.Sprintf("item vault %d", id), id)
}

// ItemCreate puts a record's norm hex into the stash or pack. A bad hex is refused locally (ErrBadHex) without a
// round trip. The change is in memory: call SaveCharacter to persist (the save writer reads it back and keeps a .bak).
func (c *Client) ItemCreate(ctx context.Context, hexNorm string, o CreateOpts) (CreateResult, error) {
	if err := ValidHex(hexNorm); err != nil {
		return CreateResult{}, err
	}
	if strings.ContainsAny(o.Key, " \t\r\n") {
		return CreateResult{}, &VaultRefused{Code: VaultErrUsage, Why: "key holds whitespace"}
	}
	line := "item create " + hexNorm + " " + map[bool]string{false: vaultWhereStsh, true: vaultWhereInv}[o.Pack]
	if o.Fixed {
		line += fmt.Sprintf(" %d %d", o.Col, o.Row)
	}
	if o.Key != "" {
		line += " key=" + o.Key
	}
	r, err := c.Do(ctx, line)
	if err != nil {
		return CreateResult{}, mapVaultErr(err)
	}
	if len(r.Lines) == 0 {
		return CreateResult{}, fmt.Errorf("d2dctl: item create answered no line")
	}
	kv := ParseKV(r.Lines[0])
	res := CreateResult{Col: -1, Row: -1}
	var ok bool
	if res.ID, ok = kv.Int("id"); !ok {
		return CreateResult{}, fmt.Errorf("d2dctl: item create reply %q has no id", r.Lines[0])
	}
	if p, ok := kv.Get("at"); ok {
		a, b, _ := strings.Cut(p, ",")
		res.Col, _ = strconv.Atoi(a)
		res.Row, _ = strconv.Atoi(b)
	}
	return res, nil
}

// VaultStats reads the fork's ledger.
func (c *Client) VaultStats(ctx context.Context) (VaultStats, error) {
	r, err := c.Do(ctx, "vault stats")
	if err != nil {
		return VaultStats{}, mapVaultErr(err)
	}
	var st VaultStats
	for _, l := range r.Lines {
		kv := ParseKV(l)
		if _, ok := kv.Get("creates"); ok {
			st.Creates, _ = kv.Int("creates")
			st.Takes, _ = kv.Int("takes")
			st.Vaults, _ = kv.Int("vaults")
			return st, nil
		}
	}
	return st, fmt.Errorf("d2dctl: vault stats reply without counters")
}

// VaultCaps asks `caps`. Without --vault-verbs (or on a stock d2d) it returns the zero value and no error.
func (c *Client) VaultCaps(ctx context.Context) (VaultCaps, error) {
	caps, err := c.Caps(ctx)
	return VaultCapsOf(caps), err
}

// SaveCharacter writes the hero to its .d2s now (the `save` verb: read-back check and .bak are the writer's).
func (c *Client) SaveCharacter(ctx context.Context) error {
	_, err := c.Do(ctx, "save")
	return err
}

// Dedupe is the keep flow's guard in the bot process: a key is claimed once. It backs up the fork's own per-process
// key set (a restarted d2d forgets its keys; this one lives with the vault store's run).
type Dedupe struct {
	mu   sync.Mutex
	seen map[string]bool
}

// Claim reports whether key was free (and takes it).
func (d *Dedupe) Claim(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.seen == nil {
		d.seen = map[string]bool{}
	}
	if d.seen[key] {
		return false
	}
	d.seen[key] = true
	return true
}

// Release gives a key back (a create that failed for a retryable reason).
func (d *Dedupe) Release(key string) {
	d.mu.Lock()
	delete(d.seen, key)
	d.mu.Unlock()
}

// CreateOnce creates rec in the hero at most once per game: it claims rec.DedupeKey(game) in d, passes it to the fork as
// the idempotency token, and gives the claim back when the create failed without creating (nospace, game state, bad
// hex). A second call for the same record answers ErrVaultDup without a round trip.
func CreateOnce(ctx context.Context, api VaultAPI, d *Dedupe, rec VaultRecord, game string, o CreateOpts) (CreateResult, error) {
	key := rec.DedupeKey(game)
	if !d.Claim(key) {
		return CreateResult{}, &VaultRefused{Code: VaultErrDup, Why: "key already created: " + key}
	}
	o.Key = key
	res, err := api.ItemCreate(ctx, rec.Norm, o)
	if err != nil && !errors.Is(err, ErrVaultDup) {
		d.Release(key)
	}
	return res, err
}

// FileIndex is the unique or set id that picks the unique or set icon row (0 for other qualities; d2d prints -1 when absent).
func (r VaultRecord) FileIndex() int {
	switch r.Quality {
	case 7:
		return max(r.UniqueID, 0)
	case 5:
		return max(r.SetID, 0)
	}
	return 0
}

// Decoder is the read-only decode verb: the bot codes against it where it needs the affixes of an item it only holds the bytes of.
// *Client and *VaultFake implement it; a stock d2d or an older fork answers ErrVaultUnsupported.
type Decoder interface {
	ItemDecode(ctx context.Context, hex string) (VaultRecord, error)
}

// ItemDecode asks the fork what the item bytes (a record's norm or raw hex) show once identified: `item decode <hex>` answers
// `decode was=<0|1> code=.. q=.. id'd=1 .. props=stat:param:value;...`. Nothing is created and no game state changes (V, fork test
// tests/itemdecode_d2d.py: the state digest is equal before and after). The returned record has the identified flag forced on
// (Identified true) and carries no unit, raw or norm of its own: Raw and Norm echo the argument.
func (c *Client) ItemDecode(ctx context.Context, hexNorm string) (VaultRecord, error) {
	if err := ValidHex(hexNorm); err != nil {
		return VaultRecord{}, err
	}
	r, err := c.Do(ctx, "item decode "+hexNorm)
	if err != nil {
		return VaultRecord{}, mapVaultErr(err)
	}
	for _, l := range r.Lines {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), "decode "); ok {
			return ParseDecodeLine(rest, hexNorm)
		}
	}
	return VaultRecord{}, fmt.Errorf("d2dctl: item decode answered no decode line")
}

// ParseDecodeLine parses what follows "decode " in the verb's answer: `was=<0|1> <the view fields of a vault line>`.
func ParseDecodeLine(rest, hexNorm string) (VaultRecord, error) {
	_, view, _ := strings.Cut(strings.TrimSpace(rest), " ")
	rec, err := ParseVaultLine("#0 raw=" + hexNorm + " norm=" + hexNorm + " " + view)
	if err != nil {
		return VaultRecord{}, err
	}
	rec.UnitID = 0
	return rec, nil
}
