# Ethereal Vault: a database of every drop, created in the game only on Keep (design and feasibility)

> status: DESIGN, 2026-11-16. Nothing here was built or run; the fork was read, not changed. "Ethereal Vault" is the WORKING NAME (user, 2026-11-16), used for the menu entry, the sheet title and the capacity meter label; it is provisional. User direction: memory `d2-quest-design-2026-11-16` (popups, same day): DB vault with CREATE ON KEEP (full item data, no re-roll, the bot never picks the item up), identify stays a step, unmarked items are kept forever up to a space cap, the inventory UI is liked and is the model; plus the later add: capacity is data and upgradable with a meta currency. Tags: V = read in code or measured (source named), S = secondary or inferred from a doc, U = unverified or from memory. Target classic 1.14d on the d2d backend (Game.exe has no item verbs: the vault is d2d only, section 3.7). Fork paths are in the private fork (the fork checkout on the worker host, branch `sandbox-mercverbs`, commit c2549842 at read time); a doc here may only cite pushed fork commits, so fork facts below name files, not hashes. Author: agent vaultdesign (Claude).

## 0. Summary and recommendation

- Feasible, and cheaper than it looks, because the fork already holds a bit-exact item codec. A ground item in d2d is a full `d2d::d2s::Item` (code, seed `uid`, ilvl, quality, prefix/suffix and rare affix ids, props, sockets, durability, identified flag) (V `components/game/loot.hpp` `GroundItem.item`, `components/d2s/d2s_items.hpp` `struct Item`), and the save writer `write_item` is the parser run backwards, tested against real saves per its header (S: comment says `test_d2s` round-trips every real save byte for byte; I did not run it).
- Capture source of truth: the d2d unit read of the GROUND item, serialised with `write_item` into the same bytes a save holds. The save file cannot be the source (the bot never picks the item up, so it is never in a save).
- Re-creation: a new fork verb `item create` that parses those bytes and puts the item in the hero's stash grid on the running server (plan on a copy, then the server's own `put_in_grid`, like the item verbs, `docs/design/item-verbs.md`). Fallback: edit the .d2s while the game is stopped through the fork's `write_save` (works, but needs the bot offline and a Go-side item codec we do not have, section 1.5). Recommendation: the verb; prove it with a lab spike (section 1.7) before any bot change.
- Also needed: a verb that reads AND removes the ground unit (`ground take`), so the item exists in exactly one place: the DB (section 1.2).
- Store: append-only JSONL log plus a versioned JSON index, from `internal/store` (stdlib only), per hero; about 20 to 45 recorded items per hour measured, tens of MB per month at worst (section 2.3).
- The existing account-wide ledger currency (`tally`, `internal/currency`, scope account) already is a meta wallet; the vault capacity upgrade can be its first sink with almost no new code (section 2.6).

## 1. Feasibility

### 1.1 What a dropped item is in d2d (V)

| Field | Where | Note |
|---|---|---|
| code, location, panel, column/row | `Item` | ground items use `location` 3 (`item_location::kGround`) |
| seed | `Item::uid` | "the item's seed (FUN_0062cbe0 seeds the unit with it)"; written as 32 bits in a save |
| ilvl, quality (1 low .. 8 crafted), qsub | `Item` | `ilvl` 7 bits in a save |
| identified, ethereal, socketed, runeword, personalized | `Item` flag bools and the raw `flags` word | writer: the named flags win over `flags` |
| magic prefix/suffix, rare/crafted affixes (6), set id, unique id, runeword id | `Item` | row ids into the MagicPrefix/Suffix, RarePrefix/Suffix, Sets, Uniques tables |
| props, set_props | `ItemProp{stat,param,value}` | widths from ItemStatCost via `ItemTables` |
| sockets, socketed_items | `Item` | children are written after the parent |
| defense, durability, max_durability, quantity | `Item` | |

The ground list is `Loot::ground` of `GroundItem{id, item, gold, x, y, label, ...}` (V `loot.hpp`). The existing devctl verb `ground` prints only `code, label, screen x, screen y, #id, x, y` (V `apps/d2d/devctl/verbs.hpp`), which is why the bot sees a label only and guesses quality (docs/d2d-loot.md "Quality guess (U)"). The data is there; no verb exposes it.

Open point U1: whether an UNIDENTIFIED ground item carries its real affixes in `props`/ids from the moment it drops (the display hides them). Evidence: `debug give lea pack q6 unid` gives "an unidentified rare" (docs/forkcheats.md, V live) and `Item.identified` is a separate bool, so the roll very likely exists. Section 1.7 step 3 settles it.

### 1.2 Capture: two new fork read/take verbs (proposal)

- `ground full [id]`: one line per ground unit `#id <hex of write_item bytes> <json decoded view>`; the JSON view carries what `dump items` carries today plus `ilvl`, `uid` (seed), `identified`, ethereal, sockets, and, for identified items only, the property text lines. Without the JSON the Go side would need a bit decoder (the Go `internal/cheat/d2s` explicitly does NOT understand the item section, V `internal/cheat/d2s/d2s.go` header) and stays stdlib-only and small.
- `ground take <id>`: reads like `full`, then removes the unit from `Loot::ground` and the client view in the same tick (what a pickup does minus the inventory). Returns the same line. Without this the item would still lie on the floor: a second copy after Keep, and the one-per-game unique flag (`Loot::found_uniques`, V) is set by the drop itself, so no new duplicate rule is needed.
- Order of operations (crash safety, section 5): peek with `ground full`, write the record to the log as `state: pending` and fsync, then `ground take`, then mark `recorded`. A crash between leaves either a pending record plus an item still on the floor (the next sweep sees the id gone or present; reconcile by seed) or a fully recorded item. Never take before the record is durable.
- The tooling gate: a new d2d flag `--vault-verbs` (own gate, like `--item-verbs` since TASK-383; `--lab-verbs` implies it). The friends release starts d2d with neither, so the vault is off there by construction (legal checklist item of docs/design/item-verbs.md section 3).

### 1.3 Re-creation option A (recommended): `item create` verb

`item create <hex> [stash|inv] [col row]` (devctl). Plan on a copy exactly as `plan_move` does: parse the bytes with the fork parser (`parse_item`-style, the same code the save loader uses, `components/d2s`), check `rules::free_spot` / `put_in_grid` against `stash_layout` (the stash is 6 wide by 8 tall in classic and 10 by 10 in LoD in the original, U; the fork's `inv_layout`/`stash_layout` is the authority), then insert into `character.items` with `location` stored, `panel` stash, a NEW runtime unit id (`id` is "the game's unit id while it's in play (not saved)", V `d2s_items.hpp`), and keep `uid` (the seed). Reply `create id=<new> at=c,r` or `err <code>` with the item-verb vocabulary (`nospace`, `place`, `far` if the stash proximity rule applies, plus new `parse`, `version`, `dup`).
- Why the server path: the hero is saved by d2d's own `character_store.save`, which writes with `write_save`, parses the written bytes back (name, level, item count) before replacing, keeps a `.bak`, and fixes the checksum (V `components/game/character/character_store.hpp`; `write_save` computes the size field and the rotating checksum, V `d2s_write.hpp`). A created item takes the same road as a picked-up one: no hand-made checksum.
- Item id collisions: unit ids are runtime values reassigned at load, not stored in the .d2s (V comment above), so creating a unit takes a fresh id; the seed `uid` can repeat only if the same record is created twice, which the DB state machine forbids (`kept` is final) and the verb can refuse when a stash/inventory item already has the same `uid`+`code` (`dup`, an extra guard, not a game rule).
- Version: the Item `version` field (default 101) is written back as read; a record made on one fork build and created on a newer one must still parse: store the fork `caps` string and the item-codec version with every record (section 2.1) and refuse with `version` rather than guess.
- Shared stash: classic LoD 1.14d has no shared stash; each hero's stash is in its own .d2s (S: PlugY-style shared stash is a mod; U that nothing in the fork emulates it). So Keep targets the HERO's stash. The Ethereal Vault itself is the cross-hero store if the user wants one (open question 1).

### 1.4 Re-creation option B: write the save while the game is stopped

`write_save(original, header, stats, items, tables, corpse, merc)` already exists in the fork, but we would call it outside d2d (a CLI) and need Go bytes for the items. The repo's Go side cannot parse or write items (V `internal/cheat/d2s/d2s.go`: "does NOT understand the item section ... tail untouched"; `internal/saves.Verify` "item section is not parsed", V docs/saves.md). So B means either a fork CLI (`d2s-add-item <save> <hex>`) or a Go codec for 100+ stat widths; both cost more than A and need the bot stopped (every Keep waits for a game exit). Risks: checksum, a hero file edited while a container still holds the old one (the sandbox container owns the volume), and a stale `.bak`. Use B only as the offline recovery tool (section 5).

### 1.5 Comparison

| | A: `item create` verb | B: save edit, game stopped |
|---|---|---|
| Bot online while keeping | yes, at the next town visit | no |
| Item codec | the fork's, in the server | fork CLI or a new Go codec |
| Checksum, backup, read-back | the fork's `character_store.save` (V) | must be repeated by the tool |
| Collision risk | none (fresh unit id) | none |
| Provenance ledger | verb call is ledgered like item verbs | file edit ledgered like `saveedit` (docs/cheats.md) |
| Cost | 1 fork verb, 1 Go client, tests | CLI plus codec |

### 1.6 Risks of A, ranked

1. A parse/semantic difference between a record created long ago and a newer fork (mitigation: version tag, `version` refusal, a corpus test that re-creates every stored record on CI, section 7).
2. Items whose data depends on the hero (personalized name; class-specific affix `class_affix`, V): keep them as they are; they are identical to the drop by definition.
3. Item level or affix tables changed by a fork rebase: ids point at table rows; the record stores the ids AND the display name at record time so a mismatch is visible, not silent.
4. Stackables (arrows, runes, keys): quantity is part of the bytes; each ground stack is one record.
5. The Game.exe fidelity gap: item data from a created item must equal a real pickup. The spike (1.7) diffs the two.

### 1.7 Spike plan (lab only, throwaway hero, the builder)

Not run here. Lab-only on the builder (the NAS worker that builds the fork, `docs/builder.md`; nothing on the game host), headless d2d, `--no-save` first, then one saved throwaway hero on its own slot (declared in `docs/probes/test-characters.md`, ledger-clean: the verbs are not memory writes, so `writes after start: 0`; the creates are listed as `vault creates: N`).
1. Add `ground full` to the fork; `debug give` ground items of kinds: normal, superior, magic unid and id, rare unid and id, unique, set, crafted, ethereal, socketed with gems, a stack of arrows, a tome with charges, a charm, a jewel. Record `full` for each.
2. `ground take`, restart nothing, then `item create` the same bytes into the stash of a SECOND fresh hero (and the same hero). Compare `dump items` and tooltips: every field equal except unit id and cell.
3. Unidentified: does the created item identify (with a scroll, `item identify`) into the SAME affixes as the original would? Compare with the original identified via `debug give` ... same seed? (U1: settle whether the roll exists before identify.)
4. Save and exit the hero (d2d writes the .d2s), reload, diff the hex of every item with the recorded hex: seed, defense and durability must match. Optional oracle: load the saved hero in real Game.exe on a hidden instance and read the item back by memory (the oracle of docs/d2-1.14d-memory.md): checksum accepted, same tooltip. This is the check that the fork's writer matches the original.
5. Negative cases: stash full, duplicate seed, truncated hex, version mismatch, hero dead, create while a panel is open.
6. Time and tick cost per create (expected one tick, like `item move`).
Stop criteria: any field differing on step 2 or 4 stops the design (fall back to B or drop the feature).

## 2. Data model

### 2.1 Record

One JSON object per drop, per hero:

```
{ "id": "v-0001a2", "t": "2026-11-16T14:03:11Z", "run": "<run id>", "area": "Burial Grounds", "diff": 0,
  "src": "monster name or chest or unknown", "fork": "<d2d caps/version>", "codec": 101,
  "raw": "<hex of write_item bytes>", "seed": 123456, "code": "wsp", "name": "Rare Short Sword",
  "q": 6, "ilvl": 14, "identified": false, "eth": false, "sockets": 0,
  "view": { ...what the hero would see: lines, only when identified... },
  "value": 340, "state": "recorded|keepq|kept|sellq|sold|pinned|discarded|trimmed",
  "pin": false, "kept": null }
```

- `raw` is the only truth; `view`, `name`, `value` are derived and rebuilt from it. An unidentified item stores NO property text in `view` (spoilers: the UI shows only what the game shows, quality colour, base, size), though `raw` holds the real roll; Reveal changes the identified bit through the real identify path (section 3.4), never by editing hex in the bot.
- `value` is the sell value lower bound for the item (`loot.SellValue`, V docs/lootrules.md), used for trim order and for the sell errand quote.
- Events (append-only log): `add`, `pin`, `unpin`, `keepq`, `kept`, `keepfail`, `reveal`, `sellq`, `sold`, `discard`, `trim`, `cap` (capacity change). State is a pure replay; an index file (store.File) caches counts and the id-to-offset map, rebuilt from the log if missing.

### 2.2 Store choice

The repo is stdlib only: no SQLite. Use `internal/store`: `store.Log` (append-only JSONL, fsynced) for events and `store.File` (versioned JSON, tmp+fsync+rename, `.bak`, `.corrupt-<unix>` on damage, `ErrFutureVersion`, V `internal/store/file.go` header) for the index and the per-hero settings (filters saved, pins). Per hero directory next to `state/tuning.json` (the sandbox keeps per-hero state there, S docs/design/item-verbs.md 14a). The account wallet is separate (2.6). Query load is tiny (thousands of rows, one watch page): the whole log is replayed into memory at start (a 30 MB log replays in well under a second, U measured) and the page filters in memory; no index needed beyond the file offsets for compaction. A snapshot compaction rewrites the log (archive-then-replace, section 5) when it grows 3x past live data.

### 2.3 Sizing from the sandbox log (V measured, read-only `docker logs --tail 30000`, 2026-11-08 container, window 00:29 to 02:20Z, 1.85 h, an early-game hero in Act 1)

| Measure | Count | Per hour |
|---|---|---|
| `loot:` lines (one per pickup round) | 131 | 71 |
| items seen on the ground | 381 | 206 |
| picked | 76 | 41 |
| skipped: small gold piles | 96 | 52 |
| skipped: filtered "drop [C] plain or inferior item" | 156 | 84 |
| skipped: filtered "keep [C]" (potions, scrolls, keys already carried) | 45 | 24 |
| `Found X` lines (distinct pickups logged) | 45 gold, 6 potions, 3 chipped diamond ... | n/a |

Caveats: 30,000 lines is short and mostly nav trace, one hero, early levels (more magic and rare per hour at higher areas and in Act 3 to 5, U); it is a sizing figure, not a drop-rate claim. Read with the Starter rules: roughly 41/h are worth picking up today; the "worth recording" set (everything except white/low junk, gold and supplies already carried) is about 25 to 45 items per hour now, up to about 130 per hour if everything non-gold is recorded.
Raw bytes: a magic item is about 40 to 70 bytes (S: the save format), a rare 90 to 140; the hex doubles it; with the JSON view a decoded record is about 0.6 to 1 KB. At 45/h x 24 h x 30 d = 32,400 records, about 25 to 35 MB per month per hero; at 130/h about 90 MB. Capacity is slots, not bytes, so these numbers only size disk (cheap) and replay time.

### 2.4 Capacity (first class, data driven; user add)

- Unit: a SLOT per record (one item or one stack), not cell footprint: the Vault grid is a scrolling list, footprint would punish big armor in a store that is not a backpack. (Option: a footprint budget; not recommended.)
- Data table `data/vault/capacity.json` (embedded, like `data/unlocks/ladder.json`): `{ "base": 400, "steps": [ {"id":"v1","add":200,"price":20}, {"id":"v2","add":300,"price":40}, ... ], "max": 2400 }`. All numbers are PROPOSALS (U); prices are in the meta currency (2.6). Config override: knob `vault.capacityOverride` (0 = use tiers; a positive number replaces the computed value for tests and support), read once per start and on change.
- Effective capacity = base + sum of owned steps (account-level purchase, applies to every hero's vault) unless the override is set.
- Pins: a pinned item is protected from trim; total pins are capped at 25% of capacity so trim always has material.

Trim and shrink rules (the user's rule: never lose items on a decrease):
1. Trim happens ONLY at insert time when the vault is exactly full (count >= capacity) and a new worth-recording item arrives. Candidates: not pinned, not `keepq`/`sellq`, state `recorded`. Order: lowest `value` class first (junk: low, normal, superior), then magic by lowest ilvl, then oldest `t`. Unique, set, rare, crafted, runes, charms, jewels, runeword bases are trimmed only when nothing else is left, and then only after a Needs-you confirmation (so by default never).
2. The trimmed record is moved to a `trashed` list (kept 14 days, recoverable from "Recently trimmed", then purged): archive-then-delete (section 5). A visible warning row appears on the Vault ("Vault full: trimmed 1 item, oldest white Short Sword").
3. SHRINK (config lowers capacity, or a downgrade): capacity drops below the current count: the vault becomes OVER CAPACITY. Nothing is deleted. New items are NOT recorded (the ground stage falls back to today's pickit decision for that item, or ignore) and the sheet shows a red banner with how many to discard. Trim does not run in the over state (trim only acts at exact-full on insert). Once the player discards or keeps enough to get under capacity, recording resumes.
4. Capacity can only be raised by buying a step (permanent). A downgrade is only possible through the config knob; no player-side downgrade (open question 4).

### 2.5 Identify state

A record stores `identified` from the ground unit. Unidentified: the card shows the base and quality colour only ("Unidentified Rare Short Sword"), no property lines. Keep on an unidentified record creates it unidentified (identical to the drop; bytes unchanged); the player then identifies it in the game the usual way (Cain, scroll, or the existing identify request). Reveal inside the vault (section 3.4) is optional. The `view` of an identified record is stored for the grid and search; searching by property works on identified records only.

### 2.6 Meta wallet (user add 2026-11-16)

Finding (V docs/tally-strip.md, docs/currency.md, docs/unlock-ladder.md): a currency with exactly the meta role exists: `tally`, a LEDGER currency of SCOPE `account` (`internal/currency/registry.json`; scope account vs per_character validated), earned from events (kills, first-of-kind monsters and areas, boss kills, quest done, waypoint, expeditions; "firsts and quests are account-wide: a second hero does not earn them again"), spent through `POST /api/unlocks/buy` as a ledger `spend` entry (idempotent key), with 402 `insufficient`. It survives hero switches because it is not on a hero. So the "general meta wallet mechanism" is: the account ledger plus the registry (currency data) plus a price table (data). Proposed shape for the vault:
- Data model: `registry.json` row for the vault currency (account, ledger), `tallyrows.json`-style earn rows (event, amount, key), a sink table `data/vault/capacity.json` (2.4) and later more sinks as data rows (`kind`: `vault_slots`, `auto_pickup_range`, `filter_slots`), each with `price`, a stable `id`, `requires`. One purchase = one idempotent ledger `spend` key `vault/<step id>`.
- Reuse versus a new currency: V1 can spend `tally` as is (zero currency code; the price table is the only new data). A separate currency would let the vault price in its own units but forks the earn rows and the strip; recommendation: reuse tally under a vault-flavoured NAME in the UI if the user wants (section 8, open question 5).
- Survives hero switches and Character Select: the ledger is account-scoped and read-only for a hero; backup/export: the ledger file is in the data dir the same snapshot/backup treatment as the cheat ledger (S, `docs/cheats.md` ledger) and an `export` button on the Hero sheet writes a JSON copy; import is NOT offered (a hand-edited wallet is a cheat).
- Anti-footgun: no real-money path of any kind (no purchase, no premium wording, no timers, no expiry, no decay); the wallet cannot go negative (402 before any change); a spend is idempotent (double click books once); nothing is lost on a lower tier because tiers are only raised by the player; a refund exists only for the LAST step and only while the content fits the lower tier (full price back as a ledger `refund` entry), otherwise `409 in-use`; the wallet and its history are visible (ledger rows) so a bug is auditable.
- Earn sources (open, section 8): quest hand-ins (the Quest board design, +10 per quest today under `tally_quest`), first-time achievements (the existing "a first for this hero" rows `tally_firsts`), boss kills (`tally_boss`, once per name), milestones (clvl 10/20/30, a difficulty cleared), plus today's kills and trips. Candidate NEW source: a drop worth recording and kept to the stash (small, capped per day) so using the vault pays back. Spend candidates beyond capacity: auto-pickup (vault) range, extra rule/filter slots, Reveal discounts. Names: section 8.

## 3. Bot behaviour

### 3.1 The `vault` action in the rules engine

`internal/lootrules` actions today: `pickup, keep, keepmerc, stash, sell, drop, identify, ask, ignore` (V docs/lootrules.md 2.3; `Action` field in `internal/lootrules/types.go`). Add `vault`, scope GROUND (V1) and later PACK ("deposit": a carried item recorded and removed through `ground take`'s sibling `item vault <id>`). Semantics: the engine answers `vault` instead of `pickup`; the looter does not pick the item up. Non-destructive, so it needs no `confirm`/`ack` floor (the floors guard sell/drop). Invariants kept: quest items, start kit, tomes/scrolls/keys needed by the kit stay `pickup`/`keep`; player marks still win (docs 2.4). Wave 2 sits it between `pickup` and `ignore`: ground default stays `ignore`.
Starter is NOT changed (it is today's behaviour, the shadow baseline). A new preset "Vault" (data) turns it on, selectable with knob `loot.vault` (default 0): rules in order: quest, kit and supplies still `pickup` (potions needed, town scrolls, keys, ammo for the worn launcher), gold `pickup` by the goldprio rules (it is not an item: gold must stay physical: it is the purse), then runes, gems from regular up, charms, jewels, jewelry, circlets, uniques, sets, rares, crafted, magic gear above ilvl or value floor X, runeword bases (3+ sockets, white/superior), socketed items, ethereal bases: `vault`; low/normal/superior without sockets, chipped/flawed gems: `ignore` (today's drop-junk behaviour; poor-mode junk-for-sale stays `pickup` while gold is under `loot.poorGoldBelow` so the early gold line is unchanged, section 4).

### 3.2 Ground stage

`Looter.pickOne` walks to the item and calls `In.Interact(item)` which maps to `cmd pickup <unit>` (V docs/d2d-loot.md). For a `vault` decision the same walk happens (recommended: keeps pace and fidelity: no free teleport pickup), then the looter calls the vault adapter instead of `Interact`: `ground full` (peek), log-record, `ground take`. No pack room check, no `arrived` proof from the pack list (the proof is the unit leaving the ground and the record becoming `recorded`). Consequences: "pack full" disappears for vaulted items (they never enter the pack; the `LeftBehind` valuables flag, V docs/d2d-loot.md, is not raised for them), and the trip reason "loot: pack full" can only come from pickup-class items.
Two options for the walk (open question 3): WALK (recommended, same time cost as today) or REMOTE within a radius (`vault.radius`, faster, shifts XP/h and gold/h, a fidelity change that the measured-run rule would have to exclude anyway).
Stackables and gold: gold stays a pickup (the purse is not an item). Potions, scrolls, keys, ammo the hero needs stay pickups (needed in the pack). Rune and gem stacks and arrows beyond need: vault as one record per stack.

### 3.3 Town chores interplay

- Sell chore: sells only pack items; vault items are not in the pack, so unmarked vault items are never sold (user rule). The "bot counts N, town finds 0" disagreement of docs/sellskip.md cannot recur for vault items (they never reach the count).
- Identify chore: unchanged for pack items; a vault Keep of an unidentified item arrives unidentified and the normal marks/identify flow applies (an `identify` request protects the item from the sell, V docs/inventory-actions.md 153).
- Stash chore: a new step `vaultKeep` at the stash: for each `keepq` record up to `vault.keepPerVisit` (default 20): `item create` into the stash (a full stash refuses with `nospace`, shown on the card). Order: after sell and before stash tidy so the tidy pass cannot move a fresh item (the tidy policy would stash it anyway: runes, uniques).
- Gear policy: unchanged. The bot never equips (V docs/gear-policy.md); Keep puts the item in the stash, not on the hero; "Compare with equipped" is a read.

### 3.4 Keep, Reveal, Sell, Discard (state machine)

`recorded --Keep--> keepq --(town, stash reachable, create ok)--> kept` (final; the record stays as history, greyed "in your stash" for 30 days, then compacted away), `keepfail(nospace|version|dup|dead)` back to `recorded` with the reason on the card.
- Reveal (identify): the cost must be real. Mechanism (reusing the real rules, no new gold or tool bookkeeping): at a town visit, `item create` the unidentified item into the PACK, `item identify <id> <tool>` (consumes a scroll or one tome charge through the server, V docs/design/item-verbs.md "Identify"), read the item bytes back, then `item vault <id>` it back into the DB (the same take path). Equivalent to the user's rule (scroll or tome charge); the Cain alternative (Cain identifies the carried items) costs a gold-or-nothing step in the original (S: free in 1.14d; U for the exact dialog, V docs/inventory-actions.md 41: the dialog rows are UNVERIFIED) and is offered when the hero stands at Cain. The UI shows the cost before the click ("uses 1 Identify scroll or 1 tome charge; you have 3").
- Sell from vault (the gold flow, section 4): `sellq`: at the next town visit create into the pack, the existing sell step sells only marked items (`sellMarksOnly`, V docs/inventory-actions.md 178), record becomes `sold` with the gold amount from the town budget line.
- Discard: final; asks to confirm; goes to the 14-day trashed list; frees a slot.
- Pin: protects from trim; capped at 25%.

### 3.5 When Keep arrives

Items are created only when the hero is in a game, alive, in town and standing at the stash chest (the stash proximity of the verbs, S docs/design/item-verbs.md "Move": the fork's 2-cell test is a `ponytail`, U versus the original) and no panel conflicts. Process: Keep clicks queue `keepq` (the card shows "queued: arrives at your next town visit"); a "Keep now" button reuses the errand mechanism of "sell now" (V docs/inventory-actions.md 173: a town errand that names no item, the bot walks home at the next safe moment) with kind `vault`. Items keep the identify-requests' queue semantics: visible states queued, doing, done, failed with a reason, cancelable while queued.

### 3.6 Provenance and crutch rule (decision text for the user, not written to docs/decisions.md)

Creating an item is a gameplay-affecting change of hero state, but it is NOT a memory write (ADR 0014 is about memory writes): it is a server verb like the item verbs, whose ADR 0026 draft (docs/design/item-verbs.md section 10) says verbs "add no powers". The vault verb adds exactly one conservation guarantee: it can only re-create what was taken from the same game earlier. Rules proposed:
1. Legitimate use: the player's product choice (Keep), for items recorded by `ground take` from this hero's own drops. The verb accepts only bytes whose seed+code the DB marks `recorded` for this hero and not yet `kept` (the Go side checks; the fork cannot, so a bare `item create` stays a lab verb behind `--vault-verbs`).
2. MEASURED runs (gate, soak, baseline, fidelity): the vault is OFF (`loot.vault` forced 0 like `gearDecisions` is laid by `withGearPolicy` for `D2BOT_RUN`, V docs/gear-policy.md), pickups are classic, no create verb is called. A run with any `vault create` is not a gate run. Reason beyond the crutch rule: remote or take-based looting changes pace and the pack-full dynamics the gates measure.
3. Ledger: every `create`/`take` is appended to the vault log with run id, hero, record id and seed; the run record carries `vault creates: N` next to `writes after start: N`. N > 0 is allowed outside measured runs.
4. Test heroes: Keep on a CHEATED hero stays a mechanics test only (docs provenance rule); items made with `debug give` that go through the vault are tagged with the hero's tag.
5. Never upstream: the verbs are fork additions (documented, vetted) and stay out of upstream PRs without the user's go (AGENTS.md; docs/forkcheats.md policy).

### 3.7 Backends

d2d with `--vault-verbs` only. Game.exe: no verbs, `loot.vault` forced 0 (classic pickit). The click path remains required and tested, as for item verbs.

## 4. Gold without selling drops (docs/sellskip.md, docs/town.md budget)

What actually changes (V measured, section 2.3): of 206 ground items per hour, 84 are junk the pickit already drops, 52 are small gold piles, and picked items are 41/h. Today the sell chore sells only poor-mode junk and pickit sell rows (V docs/sellskip.md 1: it ran in 1 of 158 trips); the vault does not remove a large income, it removes the CHANCE of selling magic/rare finds. Proposals, in order of preference:
1. No change below `loot.poorGoldBelow` (300): the poor-mode rule keeps `pickup` for gear worth at least `poorMinSell` (20), so the early gold line, repair and potions are unchanged (V `g-poor-gear` in docs/lootrules.md).
2. "Sell from vault" errand (3.4): the player marks records, the next visit sells them. Price shown from `loot.SellValue` (a lower bound for identified magic+, V).
3. Opt-in `vault.sellWhenBroke` (default 0): when the town budget needs R gold (repair + potion floor + merc revive reserve `town.mercReserve`, V docs/sellskip.md 3) and gold is short, a Needs-you row asks "Short by N for repair: sell these 4 vault items (about N gold)?" listing records of the lowest `value` class not pinned and not rare+; Yes queues them as `sellq`. No auto-sell without the answer (user rule).
4. A value floor BEFORE recording keeps junk out of the DB and unsold: low, normal and superior non-socketed items are `ignore`d (today's behaviour), so nothing sellable is lost that the bot sold before except what the poor-mode line (1) still sells.
Effect on gold/h is U until a soak with the vault on measures it (not a measured gate run, section 3.6 rule 2).

## 5. Risks and rollbacks

| Risk | Mitigation / rollback |
|---|---|
| create-from-data corrupts a hero | the d2d save path reads back and keeps `.bak` (V); the sandbox hero is snapshotted by `saves.Snapshot` at the menu before a Keep batch is enabled (V docs/saves.md: copy of .d2s/.map with retention, fail closed in game); first week: Keep only on throwaway heroes (section 7); `vault.keepEnabled` knob off by default |
| item lost between `ground take` and the DB write | pending-first protocol (1.2); the pending record contains the bytes, so recovery re-creates from `raw` |
| DB damaged | `store.File` `.bak` and `.corrupt-<unix>`; the log is the truth and replays; nightly copy of the vault dir to cold storage (`docs/cheats.md` backups pattern) |
| trim or discard regret | `trashed` list 14 days (archive-then-delete), "Recently trimmed" tab, restore button |
| a bug creates items in a loop | idempotent state machine (`kept` final), `vault.keepPerVisit` cap, duplicate-seed refusal in the verb, verb rate limit |
| fork rebase changes the codec | `codec` and `fork` tags per record; a CI corpus test creates every stored sample on the new build; a `version` refusal never guesses |
| the vault changes pace and economy | measured runs force it off; a soak with it on compares gold/h and XP/h against the classic baseline before it becomes the default |
| data leak of unidentified rolls via the UI | the API view for unidentified records is built server side without props (V design rule 2.1) and the raw hex is never sent to the browser |
| wallet bug | ledger spend/refund are idempotent entries with keys; 402 before any change; audit by ledger rows |
| offline recovery | option B CLI (1.4) as the escape hatch, run with the game stopped, same backup-first rule as `d2cheat saveedit` |

## 6. UI (working name: Ethereal Vault)

Reuse the inventory canvas (V files `internal/watch/web/inventory-canvas.js`, `inventory-tip.js`, `inventory-drag.js`, `inventory-actions.js`, `inventory.css`): the liked grid, tooltips and the action menu, with a different data source. Route `/watch/vault` (desktop and phone, same page), menu entry "Ethereal Vault" (provisional) next to Inventory.

```
Desktop (>= 900 px)
+----------------------------------------------------------------------------------+
| Ethereal Vault (working name)    [Capacity 412 / 600  ====----]  [Upgrade: 20 X]   Wallet: 35 X |
| Search [______]  Quality [All v] Type [All v] ilvl [__-__] [x] Unique [x] Set [x] Runes [x] Sockets |
| Sort [Newest v]   [ ] Show identified only        Selected: 3  [Keep] [Pin] [Sell] [Discard] |
+--------------------------------------------+-------------------------------------+
| grid (items as in Inventory, cells sized    | Item card (tooltip)                 |
| by the item; scroll; quality border;        | Rare Short Sword (Unidentified)     |
| badge: ? unidentified, pin, queued)         | ilvl 14, from Burial Grounds 14:03  |
|                                             | [Reveal: 1 scroll or tome charge]   |
|                                             | [Keep to stash]  [Compare: worn]    |
|                                             | state: recorded                     |
+--------------------------------------------+-------------------------------------+
| Queue: keep 2 queued (arrive at next town visit) [Keep now]   Recently trimmed (3) |
+----------------------------------------------------------------------------------+

Phone (< 600 px): one column; header = title + capacity meter + wallet chip; filters in a bottom sheet;
grid 4 across (tap = card as a bottom sheet with the same buttons); multi-select by long press;
the action bar sticks to the bottom.
```

- Filters: quality, type tree (`data.DB.IsA`), ilvl range, unique, set, runes, gems, charms, sockets, ethereal, identified, state (recorded, queued, kept, sold), pinned; sort: newest, quality, ilvl, value, name; search by name and, for identified items, by property text.
- Compare with equipped: the card shows the worn item of that slot beside it (a read of the inventory snapshot; no equip button, V gear policy).
- Multi-select: Keep (queues), Pin, Sell (queues), Discard (confirm dialog: "Final. Items stay recoverable for 14 days", the count and a typed number for more than 10).
- Reveal: button with cost text and your charges; disabled with the reason when none (the bot does not buy scrolls for it).
- Keep to stash: progress states queued, arriving, done, failed with the reason (stash full: "free N cells in your stash").
- Capacity meter: a bar with count/capacity, colour at 80% and 100%, "OVER CAPACITY" red when shrunk (2.4) with a "Discard to resume" hint; Upgrade button with the cost in the meta currency and the next tier's gain; disabled with "need 12 more" when short. The wallet chip also appears on the Hero sheet.
- Drops feed link: each Drops feed row of a vaulted item gets an "In vault" chip linking `/watch/vault#<id>` (U: the feed's row shape is the watch owner's; one field added); a kept item's row turns into "Kept to stash".
- API (answer-listener pattern of docs/inventory-actions.md: origin list, token or passcode cookie, JSON, small body, unknown fields refused): `GET /api/vault` (page, filters, counts, capacity, queue), `POST /api/vault/{keep,reveal,pin,unpin,sell,discard,restore,cancel}`, `POST /api/vault/upgrade {id}` (spends through the unlock buy path), schema `watch.vault/1`.

## 7. Slices and live verification

| Slice | Content | Size | Verified how |
|---|---|---|---|
| V0 spike | fork `ground full` + `item create` stub; the lab drill of 1.7 on a throwaway hero on the builder; result in `docs/probes/` | M | live on the builder: bytes equal, save reload equal, optional Game.exe oracle |
| V1 | fork verbs `ground full`, `ground take`, `item create`, `item vault`, flag `--vault-verbs`, headless tests (like `item_verbs_d2d.py`), `caps`; Go client in `internal/d2dctl`, fakes | L | fork ctest, `go test ./internal/d2dctl` with a stub |
| V2 | `internal/vault`: record, log, replay, index, capacity and trim and shrink rules, state machine, 2-build tests (Sorceress and Paladin) | M | offline tests incl. crash-between-steps and over-capacity. **DONE offline (agent vaultv2, docs/vault.md)**: unused until V3/V5 |
| V3 | rules engine action `vault`, Vault preset, looter adapter (walk, peek, record, take), knob `loot.vault`, forced off in measured runs | M | offline; shadow compare against Starter |
| V4 | `/api/vault` read side and the page (grid, filters, capacity meter, Drops feed chip) | M | watch UI golden tests; phone width by screenshot of the game-free page |
| V5 | Keep queue: stash chore step, errand button, states, failures | M | offline with fakes; then a lab drill: 20 drops, Keep 5, reload, items equal |
| V6 | Reveal and Sell-from-vault, `sellWhenBroke` Needs-you row | M | offline; live on a throwaway hero (scroll consumed once) |
| V7 | meta wallet: capacity table, upgrade API, wallet chips, ledger spend/refund | S | offline with the ledger fake; no live needed |
| V8 | sandbox rollout, flags on for the sandbox hero only (runbook like item-verbs 14a), watch counts and keep failures | S | live on the sandbox, 24 h |
| V9 | PACK-scope `vault` (deposit a carried item) and offline Option B recovery CLI | S | offline |

Live-verification plan: V0 (fork bytes), V5 (drill with `vault creates: N` declared, `writes after start: 0`), V8 (24 h sandbox, compare Keep failures and gold/h). Cards for the fork are bot-side `exclusive: fork-branch:<name>`, `d2d-lab-binary`.

## 8. Open questions for the user (max 8, with recommendations)

1. Per hero or per account? One vault per hero is simplest and matches the stash; an account vault lets a new hero start from old drops (and needs class filters). Recommend per hero; capacity upgrades account-wide.
2. What counts as "worth recording"? Recommend the Vault preset of 3.1 (about 25 to 45 items per hour) with a "record everything non-gold" option (about 130 per hour) as a toggle on the preset.
3. Walk to the item (same pace as today) or take remotely? Recommend walk in V3; a radius later.
4. Capacity numbers and downgrade: base 400 slots, steps +200, +300, ... max 2400, prices 20, 40, 80 ... (all U); shrink never deletes, over-capacity stops new records; refund only of the last step while content fits. OK?
5. Meta currency NAME (not decided; none is final). Candidates fitting the Ethereal theme, not clashing with runes, gems, Token of Absolution or gold: (a) "Ether" (short, ethereal, one word); (b) "Echoes" (what a vanished drop leaves); (c) "Remnants". Or reuse the existing "tally" under a new display name. Avoid "Wisps" (a fork stat blessing, V loot.hpp) and "Shards" (D2R shard items).
6. Earn and spend sources. Recommend: earn = quest hand-ins, "a first for this hero" rows, boss kills, milestones (clvl and difficulty cleared) and a small capped daily trickle from kills (today's `tally` rows); spend = vault slot tiers first, later auto-pickup range, extra filter slots. Do you want any earn rule tied to Keep (no recommendation: skip it, it rewards hoarding)?
7. Gold flow: Sell-from-vault errand plus the opt-in "short on gold: sell these?" prompt (default off) plus the poor-mode pickup below 300 gold. Acceptable, or do you want a value floor auto-sell before recording (no, per your rule)?
8. Reveal cost: scroll or tome charge (server-checked) and Cain when at Cain (free in the original, S/U). OK, and may a measured-run rule force the vault off (3.6 rule 2) without a separate decision? Recommend yes; it is an ADR line for your grill session, not written by this doc.

Draft decision lines (yours to write): "Loot vault (user, 2026-11-16): the Ethereal Vault (working name) records drops as full item data and creates them in the hero's stash on Keep through fork verbs behind `--vault-verbs`; identify is a step; unmarked items are kept up to a capacity that the meta currency raises; the vault is off in measured runs."

## 9. Dead ends and notes

- Reading the item from a save: impossible for items never picked up. Reading from game memory on Game.exe: no unit read of the exact bytes; d2d only.
- Footprint-based capacity: rejected (punishes armor, confuses a list).
- Creating the item remotely without a stash visit: possible technically, rejected: the stash proximity rule and the user's picture ("arrives at the next safe moment").
- Selling vault items for gold without creating them (a gold credit): rejected, it is a gold write with no item sold; the errand path uses the real vendor.
- `ground full` JSON view duplicates `dump items` formatting; reuse its writer to keep one format.
