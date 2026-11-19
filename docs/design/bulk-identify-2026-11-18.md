# Bulk identify: Cain-free, lazy reveal, filter (design, 2026-11-18)

> status: BUILT offline (B1-B5 and the town adapters, scroll path and Ether tap payer of slice B6, task bulkid2), 2026-11-19; nothing ran live. Slice plan 2026-11-18 (user go by popup, SME plan "Cain-free + lazy reveal + filter"). Built offline in the slices of section 8; nothing ran live. Tags: V = verified (code read or a named offline test), S = secondary (game knowledge, SME answer), U = unverified, needs the live drill of section 9. Related: `docs/vault.md` (Reveal), `docs/inventory-actions.md` (identify commands, stash actions), `docs/unlock-idreveal.md` (the Cards), `docs/lootrules.md`, `docs/design/item-verbs.md`, `docs/d2d-vault-verbs.md`.

## 1. What the user wants

Many unidentified magic and rare items collect (in the pack, the stash, the vault). Identify stays a STEP (decision 2026-11-16: a reveal costs a scroll, a tome charge or gold; it is free at Cain; the vault is off in measured runs). Make it cheap and automatic where it is free (Cain), cheap and deliberate where it is not (scrolls and gold), and visible in the viewer (cost preview, count, mass button).

SME numbers (S, d2-sme 2026-11-18): Cain (Act 1 camp, after Search for Cain) identifies for free, inventory items only; Scroll of Identify about 80 g, Tome of Identify about 280 g (20 scrolls). Scrolls only at 40 items per hour is about 3,200 g per hour (unaffordable at clvl 17 normal); with a filter (about 12 items per hour) about 1,000 g per hour. The ground full verb returns real affixes for unidentified items; the server hides them in the UI, and we keep that rule (V `docs/vault.md`: an unidentified record carries no property text in any API view).

## 2. Current flow (V: code read, 2026-11-18)

| Piece | Where | What it does today |
|---|---|---|
| Town identify chore | `internal/town/identify.go`, chores order `idtome`, `identify` (`chores.go`) | pack items the rules mark `NeedID` get a scroll or tome charge by right-click then left-click; scrolls are bought when none are held; Cain is NOT used. U live |
| Optional tome chore | `town.identifyTome` (default 0), `idtome` chore in `chores.go`, `needIDTome` (`playerid.go`) | buys a tome only from gold above `idBuyReserve + reserveGold`; logs a reason once per visit |
| Player identify | `/api/inventory/identify`, `town.PlayerIdentifyBatch` | one item, pack or stash in place (`identifyInStash`) |
| Player "Identify all (Cain)" | `/api/inventory/cain`, `invdesk.RequestCain`, `town.CainIdentify` (`playerid.go`) | town errand: walk to Cain, `openNPC`, pick the Identify line (`cainMenu` Talk, Identify, Cancel: U live), wait until `unidentified(s)` is empty. Locked until `CainOpen` (the Search for Cain done, or Cain seen: `cainOpen(ref)` in `cmd/d2bot/invwire.go`). Covers pack, cube and worn, NOT the stash (S, decomp finding) |
| Vault Reveal | `internal/vaultkeep/reveal.go`, `POST /api/vault/reveal` | per record: create into the pack, identify (Cain if `Identifier.Free()`, else scroll/tome), read back, supersede the record, remove the copy. Price view `RevealPrice` (`charges`, `free`, `perItem`). Refuses `noscroll` when requests exceed charges and Cain is absent; never buys scrolls |
| Cain presence | `heroActs.Free` (`cmd/d2bot/vaultkeep.go`) | true while an alive Cain-class monster (`town.IsCainClass`) is in the town snapshot |
| Cards | `internal/watch/unlockidreveal.go`, `/api/reveal`, `docs/unlock-idreveal.md` | face-down card per pickup of a quality with a card; the "Cards (N)" count on the tab; flipping is a UI reveal of the quality, not an identify |
| Ether | `internal/ether` | account wallet (vault capacity); ledger replay; not yet a payer of reveals |
| Item verbs | `internal/d2dctl/item_verbs.go`: `ItemMove id inv|stash|cube col row` (-1 -1 first free), `ItemIdentify id [tool]`; vault verbs: `ItemFull`, `ItemCreate`, `ItemVault` | V (fork test and `item_verbs_test.go`); stash moves need the hero at the chest (S docs/design/item-verbs.md) |

Gaps this design closes: (1) nothing identifies in bulk without a player click; (2) the town chore ignores Cain even when he is free; (3) stash items cannot reach Cain; (4) no policy, filter or spending cap exist; (5) the vault reveal has a per-item price but no mass action or gold/Ether price; (6) the viewer does not say when Cain is not rescued.

## 3. Policy and knobs

Knob group `identify.*` (new, `internal/env/tuning_identify.go`, listed in `docs/knobs.md` by the generator; numeric knobs, the policy is an enum code because knobs are numeric). Defaults are the NEW-hero defaults; no existing knob changes value.

| Knob | Default | Meaning |
|---|---|---|
| `identify.policy` | 1 | 0 manual (nothing automatic; the player's commands only), 1 cain-free (only Cain, free; never spends gold or scrolls), 2 scrolls-above-value (Cain when open, else scrolls/tome for items whose estimated value passes `identify.minValue`), 3 never (nothing at all, also no ride-along; player commands still work) |
| `identify.minValue` | 0 | skip items whose `loot.SellValue` lower bound is below this; 0 = identify everything the filter table selects. Plain white items are never unidentified, so there is nothing to skip: the knob matters for scrolls (policy 2) |
| `identify.maxScrolls` | 10 | scroll or tome charges one town visit may spend |
| `identify.goldFloor` | 600 | gold that must stay in the purse after any identify purchase |
| `identify.tripCap` | 40 | items one Cain visit identifies at most (pack moves are chunked: section 5) |
| `identify.chunk` | 10 | items moved per stash-to-pack batch |
| `identify.autoSellRest` | 0 | 1: the sell chore may sell identified items the rules rate `sell`; 0 (default): identified items are kept for the player to see first |
| `town.identifyTome` | 0 (unchanged) | stays 0 unless the user opts in; when 1 the tome chore additionally honours `identify.goldFloor` (the existing reserve test stays) |

`identify.policy=cain-free` is inactive until Cain is rescued; while he is not, the policy behaves as manual and the viewer says so (section 7). All of it is forced to manual in measured runs (`D2BOT_RUN` set), like the vault (V: the same helper the vault gate uses). Build-flexibility note: no class, skill or build literal appears; the item filter below is a data table, the planner takes the hero's gold, charges and snapshot only, and the tests run a Sorceress and a Paladin hero.

### 3.1 Filter table (data, not code)

`data/bulkid/filter.json` (embedded, schema `bulkid.filter/1`, validated at load, unknown fields refused). Rules are an ordered list; the first match decides `identify` (true/false) and carries a priority for the cap:

```
{ "schema": "bulkid.filter/1",
  "rules": [
    { "id": "jewelry",  "types": ["ring","amul"],          "quality": ["magic","rare"], "priority": 80 },
    { "id": "charms",   "types": ["char"],                 "quality": ["magic"],        "priority": 70 },
    { "id": "jewels",   "types": ["jewl"],                 "quality": ["magic","rare"], "priority": 70 },
    { "id": "rare_gear","types": ["*"],  "quality": ["rare"], "minIlvl": 20,              "priority": 60 },
    { "id": "set_unique","types": ["*"], "quality": ["set","unique"],                    "priority": 90 } ],
  "default": false }
```

Types are item base type codes resolved through the data tables (`data.DB.IsA`), never literals in Go. U: the exact ilvl default (20 here) is the SME's "rare+ gear above an ilvl"; it is a tunable table value. Priorities order the batch when the cap bites (highest first, ties by item id: deterministic).

## 4. The Cain trip procedure

Trigger: a town visit that is ALREADY planned (the chores run). The identify slice never starts a trip of its own. Policy 1 or 2, `CainOpen(snapshot)` true (quest word "search_for_cain" reads reward-pending or done: the free-Cain flag, quest record 4 bit 1, V `docs/probes/search-for-cain-l2-2026-11-18.md` via `task.QCain`; or an alive Cain-class monster in the town), at least one selected unidentified item exists anywhere on the hero (pack, stash), and no fight is running (`Runner.safeMoment`).

```
plan      pick candidates (filter + policy), split into pack candidates and stash candidates
stage     if stash candidates: open stash; for each batch of identify.chunk: move to pack (ItemMove id inv -1 -1), verify, log
talk      close panels, walk to Cain (town.openNPC with the Cain classes of the act), pick the Identify line
verify    read the snapshot until no carried item is unidentified, or time out
return    move every item that came from the stash back (ItemMove id stash -1 -1), verify, log
report    chore report: identified N, moved N, left N, reason
```

Rules, each with a must-fail test (section 8):

1. **Chunked moves with a verified move per batch.** Take at most `identify.chunk` items out, bounded by the pack's free cells (a 2x4 piece needs a fitting spot, `Occupancy.FindSpot`); after each move the item must be read in the pack; a missing one stops the batch and the trip goes on with what arrived. Nothing is moved for which there is no room; a full pack is never overfilled (the loot floor `loot.freeCellsKeep` stays free). Each batch logs one line (`bulkid: batch 2/4 moved 10 verified 10`) BEFORE the next batch starts. V design; tests on fakes.
2. **Never lose an item.** Every item that left the stash is tracked in a journal (`bulkid.Journal`, persisted next to the hero like the vault stages). The return step moves back exactly the journal's items that are still in the pack and unlisted elsewhere; a restart at any step (the same cut-point style as `TestRevealRestartAtEveryStep`) resumes from the hero's contents, so an item is either in the stash, in the pack, or identified in the pack. If the stash is full on return the item stays in the pack and a Needs-you row names it. Pack items are never moved.
3. **No spending of all gold.** Cain costs nothing. For policy 2 (scrolls) the spend is capped by `identify.maxScrolls` per visit and the purse must keep `identify.goldFloor` after any purchase (the reserve of `idBuyReserve` for repair, kit and merc still applies first). A purchase that would break the floor is skipped with one logged reason per visit.
4. **Unidentified items never expose affixes.** The planner and every API view it feeds work from the same stripped item view the vault uses (no property text for an unidentified item); the filter reads only type, quality, ilvl and the base value (all visible in the game's own tooltip). Test: a record with real hidden affixes yields JSON without them.
5. **Policy never does nothing else.** `identify.policy=3` makes the planner return an empty plan and the chore a no-op; 0 only the player's commands. Test per policy.
6. **Cain is not here.** Not in town, not rescued, panel busy: the step is skipped with a reason (`cain absent`), the items stay where they are, no moves happen (the staging happens only after Cain is seen in the town snapshot: nothing moves for an absent Cain).

Verbs: item and stash moves use `ItemMove` (verb path) or `hand.Transfer` (click path), whichever the town already uses for its stash chore (V: `invstash`); the Cain talk uses `openNPC` and `PickLine` (the existing `CainIdentify`); the identified flag is read from the snapshot (`Item.Identified()`), V. U: whether Cain's Identify covers the stash (S says no: stash items move first), whether the verb path `ItemMove` between stash and pack is accepted away from the chest (fork 2-cell rule), whether `CainIdentify` waits long enough for 40 items.

## 5. Lazy reveal and mass reveal (vault and viewer)

The vault Reveal (V6) stays the engine; this adds the price and the mass action.

- **Cost preview per item and per set** (`RevealInfo` grows additive fields): `charges` and `free` (today), plus `mode` (`cain`, `scroll`, `tome`, `ether`, `none`), `gold` (scroll price: 80 g from the vendor table, S; `data` base cost of `isc`), and `ether` (the Ether price, if the wallet is wired: a table value in `data/bulkid/prices.json`, `identify.etherPerReveal`, default 0 = Ether is not used; U until the user fixes a price). The page text: "Reveal 3 items: free at Cain" / "3 charges (you have 2)" / "about 240 gold" / "N Ether".
- **Lazy reveal on tap** in the grid and tooltip: the card/tooltip button shows the preview and queues the existing Reveal (the queue `revq` of V6, no new queue). Nothing identifies without a tap or the mass button, except the Cain trip of section 4 for items on the hero.
- **Mass-reveal button** "Reveal filtered (N)": reveals the vault records that pass the filter table and the viewer's current filter, up to the cost the hero can pay now (free at Cain: all; else the charge count; never gold the floor forbids). The preview names the count it will queue and the count it skips and why. Endpoint: `POST /api/vault/revealfilter {confirm:true}` (same origin and token rules as the other vault POSTs, unknown fields refused), a thin loop over `EnqueueReveal` with the filter table's selection (V design).
- **Cards**: the Cards tab keeps its quiet "Cards (N)" count; N counts unflipped cards, unchanged. A reveal done by the bulk path does not create extra cards. U: whether identified items should also flip their card automatically; left as is.

## 6. Failure modes

Cain missing from the town snapshot (quest not done, Act mismatch, he is busy): skip, no moves. Panel or cursor conflict: postpone (the existing `ErrPostpone`). Pack full: the batch shrinks to the free cells, zero free: nothing moves and a note appears. Stash move refused (away from the chest, fork rule): the trip aborts before Cain, nothing moved. Identify line not reached: report the items left, return the stash items (rule 2). Process dies mid trip: the journal resumes. A hero death or a portal out: the items stay in the pack; the next visit's return step puts them back (journal).

## 7. Viewer notes

- Settings page rows for the `identify.*` knobs (the generic knob list; no custom UI).
- When `identify.policy` is 1 or 2 and Cain is not rescued: the inventory and vault pages show "Cain is not rescued yet: identify is manual (scrolls or the item menu)" (the existing `CainTeaser` locked state of the Cain tile gets this wording in the vault Reveal line; the sentence comes from the server, not hard-coded in the page).
- The Cain tile says "free, pack only; stash items are brought over in batches of N" once the planner is on.
- Cards: no new popup (the 2026-11-18 rule stands).

## 8. Slices (each offline, tests on two heroes, Sorceress and Paladin)

| Slice | Content | Tests |
|---|---|---|
| B1 policy + filter | `internal/env/tuning_identify.go` knobs, `data/bulkid/filter.json`, `internal/bulkid` loader and pure planner (`Plan(items, hero, policy, caps) -> picks, skips with reasons`) | policy 0..3 each; `never` returns nothing; minValue; maxScrolls cap; gold floor; deterministic order; unknown type in a rule is an error; unidentified view strips affixes |
| B2 chunked moves | `bulkid.Stager` over a small `Mover` interface (pack/stash snapshot, move, journal), fake d2d mover | full pack never overfilled and no item lost; chunk size; restart at every cut point; stash-full return; a mover that lies (move returns ok but the item is not there) |
| B3 Cain trip step | town chore `bulkid` after `identify`, behind policy 1/2, `CainOpen` from `QuestView`, ride-along only on a planned visit, "Cain not rescued" state | Cain absent: no moves; present: stage, talk, verify, return; per-trip cap; policy 3 no-op; measured run forces manual |
| B4 reveal preview + mass | `RevealInfo` additive fields, `POST /api/vault/revealfilter`, page text in the vault JS | cost text for cain, scroll, none; mass reveal respects charges and the floor; the API view of an unidentified record has no affixes |
| B5 viewer note | server sentence for "Cain not rescued: manual" on the vault and inventory views | golden for both heroes |

Merge order: this doc first (`agent/bulkid-design`), then `agent/bulkid` with B1 and B2 first and B3 to B5 as far as they need no live data.

## 9. Live drill (lab slot, later, a d2-live agent, not run here)

1. Throwaway hero with Cain rescued, 12 unidentified magic/rare items: 6 in the pack, 6 in the stash. Policy 1. Visit town: batches of the chunk size leave the stash, Cain identifies, items return. Read: all identified, stash count unchanged, gold unchanged, scrolls unchanged, `writes after start: 0`.
2. Ask: does Cain's Identify cover the stash (S: no)? Put 2 stash items, talk to Cain WITHOUT moving, read the flags. Answer decides whether staging is needed.
3. Does `ItemMove` stash to pack work two cells from the chest only? Walk to the chest first (S rule).
4. Cain's Identify line index and the wait for 40 items (time it); prices of a scroll and a tome at the vendor.
5. Policy 2 without Cain (hero before the rescue): purchases stop at the gold floor, at most `identify.maxScrolls` per visit.
6. Kill the bot between stage and return: after restart the items are back in the stash (journal), none lost.
7. Hero before the rescue with policy 1: the viewer shows the manual note and no item moves.

## 10. Open points for the user (not blocking the offline slices)

- Ether as a payer of reveals needs a price decision (`identify.etherPerReveal`, default 0 = off).
- Default ilvl of the rare gear rule (20).
- Whether Cain trips should also be allowed to start a trip when many items wait (today: ride-along only, as decided).

## 11. Slice B6: town adapters, scroll path, Ether payer (built offline 2026-11-19, nothing ran live)

| Piece | Where | Notes |
|---|---|---|
| Town side of a trip | `internal/town/bulkid.go` (`BulkID`, `SetBulkID`, chore step `bulkid` after `identify`) | `bulkSide` implements Mover (`moveOne`: the item verb when the mover has it, else the click path; the stash is opened on demand, the pack keeps `loot.freeCellsKeep` cells free), Lister (pack and stash, unidentified by the `Identified` flag) and CainTalk (`Present` = a Cain-class monster alive in the town snapshot, found by class never by position; `IdentifyAll` = `CainIdentify`). Cheap need check `needBulk`: journal pending, or policy 1 with Cain rescued and standing here, or policy 2; an idle latch stops a second walk to the chest for the same pack |
| Journal | `bulkid.FileJournal` (`<data>/bulkid-<hero>.json`), recovery runs first in every `Trip.Run`, under any policy | a damaged file leaves the chore off |
| Rescued | `bulkid.CainRescued(word)` (record 4 bit 1 or 0, V `search-for-cain-l2`), `cainQuest(ref, task.StageCleared)` in cmd/d2bot | Cain absent from the snapshot while policy 1: nothing is listed, opened or moved |
| Scroll path (policy 2) | `bulkid.RunScrolls`, `town.bulkSide` as `ScrollIO` | held charges only; buys only when `town.identifyTome` is 1, within `identify.goldFloor` and `identify.maxScrolls`; `ScrollPrice` is the base cost of `isc` (S: a lower bound of the vendor price) |
| Ether | `bulkid.EtherPayer`, `vaultkeep.Keeper.EnqueueRevealEther`, `POST /api/vault/reveal {"ether":true}`, `RevealInfo.etherPer` | only an explicit tap; refused (nothing queued, nothing spent) when `identify.etherPerReveal` is 0 or the wallet is short; charged ONCE per record AFTER the verified reveal, ledger key `reveal:<record id>`. U: nothing makes a reveal possible without a scroll, tome charge or Cain, so today an Ether tap pays in addition to the identify tool; what Ether should replace is the user's price decision (section 10) |

Live drill additions (a d2-live agent, on top of section 9):
8. Cain's identify scope: put 2 stash items and 2 pack items unidentified, talk to Cain without staging, read which flags changed (stash coverage: S says no).
9. Item-move verbs away from the chest: `item move <id> inv -1 -1` from the stash panel at 2, 5, 10 cells from the chest (the fork's "far" refusal; does `moveOne` fall back to the click path?).
10. Cain's Identify line index and the dialog for a full pack: time 40 items (`identify.tripCap`), read the chore log line `bulkid trip`.
11. Prices: a scroll and a tome at the vendor (`isc` base cost vs the shop price), to replace the `ScrollPrice` lower bound.
12. Journal recovery: kill the bot between the stage and the return of a 12 item trip; after restart the journal file lists the ids and the first town visit returns them (policy manual too).
13. Cain's position: no coordinates are assumed; check `Present` finds the class in the Act 1 camp after he left the cage (V l2: he shows in town later).
14. Ether: set `identify.etherPerReveal` to 1, tap a reveal with Ether on a hero with a scroll, read the ether ledger line `reveal:<id>`; with a zero balance the tap is refused.
