package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/ether"
	"d2lod-bot/internal/merc"
	"d2lod-bot/internal/scenario"
	"d2lod-bot/internal/vault"
)

// seedOut is what the seeding leaves for the run: the ids of the seeded records in seed order (the {{vault.N.id}} of
// trigger.api) and the highest sequence number of the vault log the seeding wrote (the journal starts after it).
type seedOut struct {
	VaultIDs []string
	Seq      int64
}

// seedGame is the part of the d2d control client the seeding uses (a test double stands in for it).
type seedGame interface {
	Do(ctx context.Context, line string) (*d2dctl.Reply, error)
	Step(ctx context.Context, n int) (d2dctl.ClockInfo, error)
	Items(ctx context.Context) ([]d2dctl.Item, *d2dctl.Reply, error)
	ItemFull(ctx context.Context, id int) ([]d2dctl.VaultRecord, error)
}

// chainFunc gives the item type chain and display name of a base code (the vault record fields the bot's recorder fills).
type chainFunc func(code string) (chain []string, name string)

// gameChain opens the item tables of the game install, cached under the bot's data dir like the bot does.
func gameChain(dataDir, gameData string) func() (chainFunc, error) {
	return func() (chainFunc, error) {
		db, err := data.Open(gameData, filepath.Join(dataDir, "cache", "tables"))
		if err != nil {
			return nil, fmt.Errorf("item tables for the type chain: %w", err)
		}
		return func(code string) ([]string, string) {
			if b := db.ItemByCode(code); b != nil {
				return merc.TypeChain(db, b.Type), b.Display
			}
			return nil, ""
		}, nil
	}
}

// carriedIDs lists the unit ids of the carried items (`item full`, vault verbs on).
func carriedIDs(ctx context.Context, c seedGame) (map[int]bool, []d2dctl.VaultRecord, error) {
	recs, err := c.ItemFull(ctx, -1)
	if err != nil {
		return nil, nil, err
	}
	m := map[int]bool{}
	for _, r := range recs {
		m[r.UnitID] = true
	}
	return m, recs, nil
}

// packIDs lists the unit ids `items` reports (works without the vault verbs).
func packIDs(ctx context.Context, c seedGame) (map[int]d2dctl.Item, error) {
	items, _, err := c.Items(ctx)
	if err != nil {
		return nil, err
	}
	m := map[int]d2dctl.Item{}
	for _, it := range items {
		if id, err := strconv.Atoi(it.Attrs["id"]); err == nil {
			m[id] = it
		}
	}
	return m, nil
}

// applySeeds does the ordered seeding of su.Seeds on the prep d2d: every give and equip is a ledgered SETUP verb; a vault
// record is read with `item full`, taken off the hero with `item vault` and added to the hero's vault store under
// data/vaults, exactly as the bot's own recorder does. It returns the record ids in seed order.
func applySeeds(ctx context.Context, c seedGame, su scenario.Setup, ledger, dataDir string, chainOf func() (chainFunc, error), hero string) (seedOut, error) {
	var out seedOut
	if len(su.Seeds) == 0 {
		return out, nil
	}
	lf, err := os.OpenFile(ledger, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return out, err
	}
	defer lf.Close()
	enc := json.NewEncoder(lf)
	do := func(line, reason string) error {
		l := ledgerLine{Time: time.Now().Format(time.RFC3339), Verb: line, Reason: reason}
		rep, err := c.Do(ctx, line)
		if rep != nil {
			l.Reply = strings.Join(append([]string{rep.Rest}, rep.Lines...), " | ")
		}
		_ = enc.Encode(l)
		if err != nil {
			return fmt.Errorf("%s: %w", line, err)
		}
		_, err = c.Step(ctx, 2) // the game applies a verb on its next tick
		return err
	}
	var st *vault.Store
	var chain chainFunc
	for _, sd := range su.Seeds {
		switch sd.Kind {
		case scenario.SeedWorn:
			before, err := packIDs(ctx, c)
			if err != nil {
				return out, err
			}
			if err := do(sd.Give, sd.Reason); err != nil {
				return out, err
			}
			after, err := packIDs(ctx, c)
			if err != nil {
				return out, err
			}
			id, ok := newOf(before, after, sd.Code)
			if !ok {
				return out, fmt.Errorf("worn %s: the give left no new %s in the pack", sd.Code, sd.Code)
			}
			if err := do(fmt.Sprintf("item equip %d %d", id, sd.Slot), sd.Reason+" (equip)"); err != nil {
				return out, err
			}
			worn, err := packIDs(ctx, c)
			if err != nil {
				return out, err
			}
			if it, ok := worn[id]; !ok || it.Loc() != 1 {
				return out, fmt.Errorf("worn %s in %s: the item is not worn after the equip verb", sd.Code, sd.SlotName)
			}
		case scenario.SeedVault:
			if st == nil {
				if st, err = vault.OpenHero(filepath.Join(dataDir, "vaults"), hero, vault.Options{}); err != nil {
					return out, err
				}
				defer st.Close() // a second Close after the explicit one is harmless
				if chain, err = chainOf(); err != nil {
					return out, err
				}
			}
			id, err := seedRecord(ctx, c, st, do, sd, chain)
			if err != nil {
				return out, err
			}
			out.VaultIDs = append(out.VaultIDs, id)
		default:
			return out, fmt.Errorf("seed kind %q", sd.Kind)
		}
	}
	if st != nil {
		for i, sd := range su.Seeds {
			if sd.Kind == scenario.SeedVault && sd.Pin {
				if err := st.Pin(out.VaultIDs[vaultIndex(su.Seeds, i)]); err != nil {
					return out, fmt.Errorf("pin: %w", err)
				}
			}
		}
		if err := st.Close(); err != nil {
			return out, err
		}
		out.Seq = lastSeq(filepath.Join(vault.HeroDir(filepath.Join(dataDir, "vaults"), hero), vault.LogFile))
	}
	return out, nil
}

// vaultIndex is the position of seed i among the vault seeds.
func vaultIndex(seeds []scenario.Seed, i int) int {
	n := 0
	for _, s := range seeds[:i] {
		if s.Kind == scenario.SeedVault {
			n++
		}
	}
	return n
}

// newOf finds the unit that appeared between two listings, preferring the code asked for.
func newOf(before, after map[int]d2dctl.Item, code string) (int, bool) {
	best, found := 0, false
	for id, it := range after {
		if _, had := before[id]; had {
			continue
		}
		if it.Code == code && (!found || id > best) {
			best, found = id, true
		}
	}
	return best, found
}

// fillerRetries is how often a filler that duplicates a record is replaced by another magic item.
const fillerRetries = 4

// seedRecord gives one item, reads it, takes it off the hero and adds it to the vault. A filler that duplicates a record
// already there is replaced (a magic item rolls other affixes); a named record never is: that is an error.
func seedRecord(ctx context.Context, c seedGame, st *vault.Store, do func(line, reason string) error, sd scenario.Seed, chain chainFunc) (string, error) {
	give := sd.Give
	for try := 0; ; try++ {
		before, _, err := carriedIDs(ctx, c)
		if err != nil {
			return "", fmt.Errorf("vault seed %s: item full: %w", sd.Code, err)
		}
		if err := do(give, sd.Reason); err != nil {
			return "", err
		}
		_, recs, err := carriedIDs(ctx, c)
		if err != nil {
			return "", err
		}
		var rec *d2dctl.VaultRecord
		for i := range recs {
			if !before[recs[i].UnitID] && recs[i].Code == sd.Code {
				rec = &recs[i]
			}
		}
		if rec == nil {
			return "", fmt.Errorf("vault seed %s: the give left no new %s in the hero", sd.Code, sd.Code)
		}
		if err := do(fmt.Sprintf("item vault %d", rec.UnitID), sd.Reason+" (take off the hero)"); err != nil {
			return "", err
		}
		ch, name := chain(sd.Code)
		it := vault.Item{Raw: rec.Norm, Code: rec.Code, Name: name, Quality: rec.Quality, Ilvl: rec.ILvl, Identified: rec.Identified,
			Chain: ch, Sockets: rec.Sockets, Eth: rec.Ethereal, Found: vault.Found{Time: time.Now().UTC(), Zone: "scenario seed"}}
		if rec.Identified {
			it.Props = rec.Props
		} else if it.Sealed = rec.Props; it.Sealed == "" {
			it.Sealed = sd.Sealed // the fork sent no affixes for the unidentified item: the scenario holds its own
		}
		it.Pic, it.Tint, it.FileIndex = rec.Pic, rec.Tint, rec.FileIndex()
		r, err := st.Add(it)
		switch {
		case err == nil:
			return r.ID, nil
		case errors.Is(err, vault.ErrDuplicate) && sd.Filler && try < fillerRetries:
			give = scenario.GiveLine(sd.Code, 4, false)
		default:
			return "", fmt.Errorf("vault seed %s: %w", sd.Code, err)
		}
	}
}

// seedEther credits the account wallet with n Ether before the start: one ledger line in <state>/currency/ether.jsonl
// (S: currencyDirFor(stateDir) is that directory for the bot's state dir). The wallet replays it like any credit.
func seedEther(state string, n int) error {
	if n <= 0 {
		return nil
	}
	dir := filepath.Join(state, "currency")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, _ := json.Marshal(ether.Entry{Seq: 1, TS: time.Now().Format(time.RFC3339), Kind: ether.Credit, Amount: n, Source: "scenario", Key: "scenario-seed", Note: "SETUP: pre.ether"})
	return os.WriteFile(filepath.Join(dir, ether.FileName), append(b, '\n'), 0o644)
}

// lastSeq is the highest "seq" of a vault log (0 when there is none).
func lastSeq(path string) int64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var n int64
	for _, l := range strings.Split(string(b), "\n") {
		var e struct {
			Seq int64 `json:"seq"`
		}
		if json.Unmarshal([]byte(l), &e) == nil && e.Seq > n {
			n = e.Seq
		}
	}
	return n
}
