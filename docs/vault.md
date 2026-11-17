# Ethereal Vault store (internal/vault) and its file formats

> status: V2 store BUILT offline, V3 bot wiring BUILT offline (section 'Bot wiring (V3)'), V4 read API and page BUILT offline (section 'UI and API (slice V4)'), V5 Keep queue BUILT offline (section 'Keep (slice V5)', live drill NOT run) (slice V2 of docs/design/loot-vault-2026-11-16.md). Additive and unused until V3 (rules action) and V5 (keep queue) call it. Not run against a live game: every claim below is about the offline package (V: `go test ./internal/vault/`). Item hex comes from the fork codec later (V1); here it is opaque bytes.

## Layout (per hero)

`<root>/<hero>/vault/` (`vault.HeroDir`), hero key validated (no `/ \ :`, no `..`, 64 chars max).

| File | What | Durability |
|---|---|---|
| `log.jsonl` | the truth: one event per line, append-only, `O_APPEND`, fsync per event | a torn last line (no newline) is cut on open; a complete bad line is skipped and counted (`Status.Anomalies`) |
| `index.json` | snapshot of the replayed state plus `offset` and the sha256 of the last covered line | a CACHE: missing, damaged, other hero or not matching the log = full replay; temp file + rename |
| `archive.jsonl` | every record that left the vault (trim, discard) with time and reason | written and synced BEFORE the log line that removes the record |

Capacity and the account wallet are not stored here: the vault reads capacity through `Options.Capacity` (a func read at every insert) and the `Account` / `Wallet` interfaces in `capacity.go` (V7 implements them over the ledger).

## Log events

`{"seq":n,"ev":"...","id":"v-...","t":"RFC3339","item":{...},"trim":"v-...","reason":"...","n":0}`

| ev | from state | to | note |
|---|---|---|---|
| `add` | (new) | recorded | carries the whole item; `trim` names the displaced record (one line, so insert-with-trim is atomic) |
| `pin` / `unpin` | recorded | recorded | flag only; pins capped at `pinPercent` of capacity |
| `keepq` | recorded | keepq | Keep clicked |
| `kbegin` | keepq | keepq, `inflight` | logged BEFORE the game create is sent; a second `kbegin` is refused |
| `kept` | keepq | kept (final) | create confirmed |
| `kretry` | keepq | keepq, not in flight | the create is known not to have happened |
| `keepfail` | keepq | recorded | reason on the card (nospace, version, dead) |
| `keepcancel` | keepq (not in flight) | recorded | |
| `reveal` | recorded | recorded, identified | carries the identified item bytes; the id stays, the content hash changes |
| `sellq`, `sellfail`, `sold` | recorded / sellq | sellq, recorded, sold (final) | V6 |
| `discard` | recorded | discarded (final) | archive line first |

States: recorded, keepq, kept, sellq, sold, discarded, trimmed. "Pinned" and "revealed" from the design are flags shown by `Rec.Status()`, not states (a pinned item can also be revealed). Live states (recorded, keepq, sellq) count against capacity; final ones do not.

## Identity and dedupe

d2d items have `uid` 0 (V0 finding), so identity is `sha256(raw bytes)[:8]`. ID is `v-<hash16>`; the same bytes recorded again after the first record finished get `v-<hash16>~<generation>`. While a record with the same bytes is live, Add returns `ErrDuplicate`. Two genuinely different drops with identical bytes (two Minor Healing Potions) therefore collapse into one live record: the vault stores distinct items; stackables are a V3 decision (record as one item with a count or skip). U: whether real different drops can ever share bytes outside stackables.

## Capacity, trim, shrink (data/vault/capacity.json, value.json; numbers are placeholders)

- Capacity = `base` plus owned steps, capped at `max`; `Tiers.Capacity(owned, override)`; unknown step ids are ignored.
- Insert at exact full (count == capacity) trims ONE record first: lowest `value`, then lowest ilvl, then oldest, among unpinned `recorded` records. Protected items (qualities set, rare, unique, crafted; runes, jewels, charms by code prefix, from value.json) go only when nothing else is left AND `Options.TrimProtected` is set, otherwise `ErrFull`. The victim is archived first.
- Over capacity (count > capacity, after a shrink): Add returns `ErrOverCapacity`, nothing is trimmed, nothing is deleted; recording resumes once discards or keeps bring the count back to capacity.
- Value (`ValueTable.Of`) is computed at Add and stored, so replay never changes a trim order. It does not reuse the loot rules (those decide pickup; none of their helpers rank records).

## Keep crash recovery

Order: `Keep` (keepq) -> `BeginKeep` -> game create -> `CompleteKeep`. After a restart `Recover(present)` looks at records in flight and asks `present(rec)` (is that item in the stash): yes completes the keep, no clears the mark so it is queued again, an error leaves it in flight. The create is therefore sent at most once per confirmed state; the one unavoidable window (create done, `kept` not logged) is closed by the stash check, never by sending again.

## Export / import (`Store.Export`, `Store.Import`)

`{"version":1,"hero":"...","items":[Item...]}`: live records, oldest first. Import: at most 64 MiB and 5000 items, version must match, every item validated before any is added; items go through Add (dedupe, capacity; import never trims, it counts `Blocked`). Not the wallet: a hand-edited wallet is a cheat (design 2.6).

## Measured (V, `go test -bench`, a small 4-core box, NoSync, indicative only)

Insert about 8 us (10k and 50k); insert with trim at 10k about 0.6 ms; Open by full replay 59 ms at 10k and 296 ms at 50k; Open with snapshot about 270 ms at 50k (JSON decode dominates, so the snapshot does not beat replay yet: U whether a binary snapshot is worth it); query with code+quality+ilvl+sockets about 0.6 ms at 10k and a name substring scan about 10 ms at 50k. fsync cost per event is not in these numbers.

## Not done here

Log compaction (design 5: archive-then-replace at 3x live data); the trashed-list purge schedule (`PurgeArchive(before)` exists, nobody calls it); HTTP API and UI (V4); wallet (V7); the item codec that fills `Raw`, `Name`, `Chain` (V1).

## Bot wiring (V3, `internal/loot/vault.go`, `cmd/d2bot/vaultwire.go`)

Offline only (U live). With `loot.vault` 1 the ground stage uses the Vault preset (docs/lootrules.md section 14); for a `vault` decision `Looter.vaultOne` walks like a pickup, calls `GroundFull` for the unit, builds a `vault.Item` from the fork's `norm` bytes (the create key) plus code, quality, ilvl, sockets, identified and ethereal as the fork reports them (an unidentified item keeps its roll and bit), and calls `Store.Add`. Outcomes: recorded; `ErrDuplicate` (counted, never revisited); full (blocked 2 min, one Needs-you line per episode, items skipped as `vault full`); anything else (peek refused, unit gone, store error) marks that unit as fallen back and today's pickup decision applies. `ErrVaultUnsupported` / `ErrVaultOff` also make the feature inert for the process. Records of one game are told apart by unit id plus spot, so a recorded drop is not walked to twice; the vault dedupes by content, so two identical stackable-like drops collapse into one record (see Identity above).
`Looter.VaultInfo()` (active, inert reason, blocked, recorded, dups, fallback) is the visible state. `wireVault` opens `<data>/vaults/<hero>/vault` on the first record. Measured runs: `D2BOT_RUN` forces the knob to 0 (the same detection as gear policy and the item-verb fast lane).
Not done: `ground take` after recording (Keep slice V5 creates from the record), the Needs-you row itself (V3 emits a log line through `VaultWiring.Notice`), the `/api/vault` page (V4), a Vault entry in the preset list of the store.

Log compaction (design 5: archive-then-replace at 3x live data); the trashed-list purge schedule (`PurgeArchive(before)` exists, nobody calls it); wallet (V7); the item codec that fills `Raw`, `Name`, `Chain` (V1).

## UI and API (slice V4: `/watch/vault`, `/api/vault`)

> status: BUILT offline, not seen live. V: `go test ./internal/watch/ -run Vault` (API against a 10k-item store), `go test ./cmd/d2bot/ -run Vault`, node tests of the page logic (`TestVaultLogicJS`), the fit pages of `tools/uishots.sh --fit` (pages `vault-*`, must-fail `vaultdefect-*`) and the pictures of `tools/uishots.sh --vt`, all on the worker host. The recorder (V3) writes a vault only with `loot.vault` 1 and a fork with the vault verbs, so on a bot with the default knob the page shows its empty state or the read-only banner.

What it is: a read side plus Pin, Unpin and Discard of the Ethereal Vault (Keep since V5: section 'Keep (slice V5)'; Reveal and Sell since V6: section 'Reveal and Sell-from-vault'). No Upgrade (V7): the Upgrade button is drawn, disabled, with the reason in text (`aria-disabled` and a description, never a dead press: pressing says why).

Routes (code: `internal/watch/vaultapi.go`; bound in `cmd/d2bot/vaultpage.go`, mounted in `internal/watch/funwire.go` as `Fun.Vault`):

| Route | Listener | What |
|---|---|---|
| `GET /api/vault` | viewer | schema `watch.vault/1`: `mode` (`on`, or `off` while knob `loot.vault` is 0), `exists`, `hero`, `total`, `offset`, `limit`, `items[]`, `capacity` (count, capacity, pins, pinLimit, percent, `level` ok/warn/full/over, `over`), `upgrade` (always `available:false` until V7, the next step and its price from `data/vault/capacity.json`, currency `Ether`), with `facets=1` also `kinds` and `types[]` (item types in the vault with counts) |
| `GET /api/vault/item/<id>` | viewer | the row plus `where` (found time, area), `lines`, and `actions` (pin, unpin, discard, keep, reveal each `{ok, reason}`; keep and reveal are never ok here) |
| `POST /api/vault/pin`, `/unpin`, `/discard` | answer | `{"ids":[...], "confirm":true}`; 1 to 150 ids (the answer body is capped at 4 KiB, the page sends longer selections in chunks); discard needs `confirm`; the store archives each record before it leaves (docs above). Reply `{ok, done[], failed[{id, code, error}]}`; 200 when something was done, 409 when nothing was; `409 off` while the vault is read only |
| `GET /watch/vault` (+ `.js`, `-logic.js`, `.css`, `.json`) | viewer | the page and its files |

Query of `GET /api/vault`: `q` (name substring, at most 64), `quality` (ids or words, comma list, any of), `kind` (one of `web/vault.json` `kinds`: unique, set, rune, jewel, charm; a quality set and/or item code prefixes, data not code), `type` (a member of the item's base type chain), `ilvlMin`, `ilvlMax`, `sockets` (exact), `eth`, `identified`, `pinned` (`1`/`0`), `state` (comma list; default the live states), `sort` (`new` default, `old`, `quality`, `ilvl`, `value`, `name`), `offset`, `limit` (default 60, at most 200). A bad value is a 400 `{code:"query"}` naming it. `vault.Filter` gained `CodePrefix`, the sorts `ilvl`, `quality`, `name`, and `Store.Facets()`.

Read-only: with no desk the API answers 503 `off`; with no vault on disk (`exists:false`) or the knob at 0 the page opens, GET works (a vault that exists stays readable at knob 0) and every POST answers 409 `off`. The page never creates a vault. `cmd/d2bot` reads `<data>/vaults/<hero>/vault` (the recorder's root) only when it is open or its log exists. `vaultHome` (`cmd/d2bot/vaultpage.go`) is the one shared Store per hero: the recorder (`wireVault`, V3) takes it through `vaultHome.Open`, the page through `Existing`, because two Stores over one log would each hold their own index. The knob is V3's `loot.vault`.

Unidentified items: the row carries the as-displayed name the recorder stored (the base for an unidentified item), quality word, item level; no property text exists in the row or the detail, and the tooltip adds "Unidentified" itself. U: the stored `Name` of an unidentified record must be the base name; V3's codec adapter has to guarantee it (the API cannot tell).

The page (`internal/watch/web/vault.html`, `vault.js`, `vault-logic.js`, `vault.css`, texts `vault.json`):
- Reuse: the cell look of `components.css` section 14 (quality border, tags), `inventory-tip.js` and `.css` unchanged (the grid is `id="pack"`, the rows are dispatched as an inventory view on `d2:inventory`, each row numbered), the Characters page frame (header, no page scroll, one scroll region, modal sheet), the "..." menu entry, Orders More entry, and a link in the Drops feed (Items tab of the Log drawer). Cells are slots, not footprints (design 2.4); art is not drawn (the cells show the name; U: item art needs the codec's `art` field).
- Header: capacity meter (`role=meter`, count/capacity, colour at 80 and 100 percent, red when over capacity with the "discard N to resume, nothing is deleted" banner), pins used, the disabled Upgrade button with the next tier's gain and price and the text "arrives with the Ether wallet (V7)".
- Filters: search, quality chips, kind, type, item level range, sockets, Ethereal, Unidentified only, Pinned only, state, sort, grid/list toggle. Below 1100 px wide or 700 px high the panel folds behind a Filters button (with the active count). The filters are in the address (`?quality=7&kind=rune`), so a link or reload keeps them; `#<id>` opens an item.
- Selection: Select button, shift-click range, ctrl/meta-click, long press (500 ms), "Select all shown", Escape clears. Pin, Unpin, Discard act on the selection. Discard opens a modal (focus loop, Escape, default focus on Cancel); above 10 items the count must be typed.
- A wide window (900 px and up) shows the card of the open item beside the list (state, lines, where found, Pin or Unpin, Discard, Keep). On a phone a tap shows the tooltip as a bottom sheet and actions go through Select.
- States: loading, error (retry), empty vault (explains the Ethereal Vault), off (read only banner), nothing matches (clear filters), over capacity, full. The list polls every 6 s and reloads only when the total, count or first row changed; pages of 60, "Show more" and an end-of-list sentinel up to 600 rows (search narrows the rest).
- Accessibility: every control has a name (the fit check fails one without), targets are at least 44 px, cells are buttons with a full label and `aria-pressed` for selection, arrows move between cells, no motion.

Measured (V, `TestVaultPagingPerformance`, 10,000 records, 4-core box, GET handler in process, mean of 10): first page 2 ms, sort by value 5 ms, sort by name from offset 5000 16 ms, kind and ilvl sort under 1 ms, name search with sort 2 ms, facets 1 ms. The ceiling the test enforces is 150 ms.

Fit and pictures: `tools/uishots.sh --fit` has the pages `vault-*` (states pop, warn, full, over, empty, none, off, error, big; list view; filters open; select; the confirm sheet with and without the typed number; tooltip hover; the menu and More entries) at the usual sizes plus 412x780 and 360x640, and the must-fail `vaultdefect-*` (clipped text, wide grid, list under the header, small target, covered control, cut sheet, unnamed control, enabled placeholder). `tools/uishots.sh --vt` writes the pictures (`vt-NAME-WxH.png`). Fixtures: `tools/uishots/fixtures_vault.go` (real stores, frozen drops).

Dead ends and not done: no loading-state picture (headless Chrome's virtual time waits for the pending request, so the harness never sees it); the Keep, Reveal, Sell, Upgrade verbs; "Recently trimmed" list; property-text search (needs the codec's view); compare with the worn item (needs the inventory view on this page).

## Keep (slice V5: `internal/vaultkeep`, `POST /api/vault/keep`)

> status: BUILT offline, not run against a live game. V: `go test ./internal/vaultkeep/ ./internal/vault/ ./internal/watch/ ./internal/bot/ ./cmd/d2bot/` (queue, executor and recovery against `d2dctl.VaultFake` on a Sorceress and a Paladin hero), the fit pages `vault-keep*` and the pictures `tools/uishots.sh --vt --only 'keep*'`. U: everything that needs the fork (the `item create` byte round trip with real items, `item full` listing stash items, the stash proximity of the verb, `save` after a create); the live drill below checks them.

Knobs (docs/knobs.md): `loot.vault` 1 (the vault itself) and `loot.vaultKeep` (default 1) both on, otherwise the Keep gate answers `off`; `loot.vaultKeepPerVisit` (default 20). Forced off in measured runs (`D2BOT_RUN`, as gear policy and the recorder). `loot.vaultKeep` 0 is the kill switch: nothing is created at once, queued items stay queued.

Flow. The player presses Keep (card, or "Keep selected (n)" in the selection bar): `POST /api/vault/keep {ids, target?, now?}`. `target` is `stash` (default) or `inv`. The route only QUEUES (`Store.KeepTo`: recorded to keepq); it never creates anything and there is no instant lane. The reply is `{ok, done[], already[], failed[{id, code, error}]}`; the page shows the real sentence and then reads the state back from `GET /api/vault` (`keep` object: enabled, why, queued, creating, done, full, line; each row: `keep` = queued, creating, kept or failed, `reason`, `where`). Refusals before anything is queued (409): `off` (knob off or measured run), `dead` (hero dead), `verbs` (the d2d build lacks `item.create`), `stopped` (a read-back mismatch tripped the switch for this run); per id: `missing`, `state`, `target`.

Idempotent: a record that is queued or kept answers ok with `already`; the state machine refuses a second `kbegin`; the create carries `key=vk-<record id>`, so the fork refuses a second create in one game too (`dup`).

Executor (`Keeper.Run`, called by `bot.vaultKeepStep` after the player's own item actions): only in town, alive, no blocking panel, cursor free (`keepReady`, from the last snapshot), no fight, no trip (the bot's `safeMoment`). Per call: at most `loot.vaultKeepPerVisit` creates and 8 s of wall time (`Keeper.Budget`); the bot retries a queue that did not move after 30 s, a changed queue length at once, so a long queue never stalls the chores. Order per item, each step a line of the vault log:

1. take the save backup before the FIRST create of the process (`Backup`: the fork's `save` verb, which writes through the character store, reads the bytes back and keeps its `.bak`; `saves.Snapshot` is not usable, it refuses in a game and the d2d path has no save directory). A failing backup stops the batch (fail closed).
2. `BeginKeep` (logged BEFORE the verb), `item create <raw hex> stash|inv key=vk-<id>`; the verb returns the unit id and the cell.
3. read-back: `item full <unit>` must equal the recorded bytes (`Raw` = the fork's norm bytes). A mismatch removes the new copy again (`item vault`), fails that record with the reason, and TRIPS the keeper: nothing more is created until the bot restarts, and a Needs-you row says so. The record itself never leaves the vault on a Keep.
4. after the batch, ONE `save` (read-back and `.bak` are the fork's); only then `kept` is logged for each item with where it landed (`stash 3,5`). A failed save leaves the items in flight; the next run looks at the hero, sees them, saves again and confirms: it never creates a second one.

Unidentified items are created unidentified (the bytes are the drop's); identify stays the normal step.

Crash and failure recovery (V: tests `TestRestart*`, `TestSaveFailure*`, `TestUnanswered*`): a record still in flight is resolved first by `item full` (is there more of this item in the hero than kept records of the same bytes the vault already counts?): present completes it, absent queues it again (`kretry`); a transport error or timeout on the create itself leaves it in flight for the same look. Never by creating again. U: a player who moved or sold an identical earlier kept copy makes the count too low and one more copy may be created; the key of the fork covers the same game.

Failure reasons (stored `keep: <sentence>`, shown without the prefix): the stash is full: free space or upgrade (also the Needs-you row "Stash full: free space or upgrade. N kept item(s) wait."; the items STAY queued and the next visit retries; the backpack variant says so), the hero is dead / not in a game (stay queued), this d2d build has no vault verbs (stay queued), the game is busy (stays queued), the game does not accept this item's data (`parse`: failed, back to recorded), another game version (`version`: failed), the created item differs (failed, tripped), the key was already used but the item is not in the hero (`dup`: failed), the save backup / save failed (stay queued). Partial success is normal: items before a failing one are kept, the failing one is marked, the batch goes on after per-item failures and stops after hero-wide ones (stash full, dead, verbs, busy).

Page. The Keep button of the bar and the card are real. The bar has "Keep selected (n)"; a status row under it shows "Keeping 5 items: arrives at your next town visit" (or why it waits) with "Keep now": it only enqueues (any queued id) and asks the bot for the errand trip home at the next safe moment (`Keeper.TripWanted`, `bot.vaultKeepStep`, same backoff as the other errands); creating still happens in town at a safe moment. Rows show Q (queued), ... (creating), K (kept), ! (failed with the reason in the label, the list and the card); a "Kept" chip in the filters shows `state=kept`.

Live drill checklist (not run; needs the d2d fork with `--vault-verbs`, a throwaway hero, `loot.vault` 1, NO `D2BOT_RUN` for the Keep part and `vault creates: N` declared in the run record, `writes after start: 0` from `d2cheat ledger --since-run`; a Keep run is not a measured run):
1. Start the sandbox hero fresh (zero setup writes), `loot.vault` 1, let it play until at least 20 vault records exist (`GET /api/vault` total).
2. Select 5 on the page, Keep selected: the progress line says "Keeping 5 items", `vault stats` creates stays 0.
3. Let the hero reach town (or press Keep now): expect one `save`, five `item create` lines in the log, rows turn K with `stash c,r`, `vault stats` creates=5, no panel was opened.
4. Reload the game (exit and re-enter): the five items are in the stash, `item full` equals the recorded bytes, unidentified ones are still unidentified.
5. Restart the bot with one item queued: it is created once (vault log shows one `kept`); kill the bot between the create and the save on a second try: no second copy after restart.
6. Fill the stash, Keep one more: Needs-you row "Stash full", item stays queued, free a cell, next visit creates it.
7. Dead hero: press Keep: the toast says the hero is dead; nothing queued.
Record the `.bak` of the hero before step 3 and restore it if anything differs.
## Reveal and Sell-from-vault (slice V6: `internal/vaultkeep`, `POST /api/vault/{reveal,sell,cancel}`)

> status: BUILT offline, not run against a live game. V: `go test ./internal/vaultkeep/ ./internal/vault/ ./internal/watch/ ./internal/loot/ ./cmd/d2bot/` (executors against `d2dctl.VaultFake` on a Sorceress and a Paladin hero, fake clock, a "process dies after N game calls, with and without the last call taking effect, then restarts" test per flow), node tests of the page logic, a stub-DOM mount of the page (price line, buttons, sheet). NOT run: the fit pages `vault-reveal*`, `vault-sell*`, `vault-revealing*` and the pictures (`tools/uishots.sh --fit --only 'vault-*'`, `--vt --only 'reveal*'`): they need a browser, which this slice was not allowed to use (the worker host has one). U: everything that needs the fork (item create and identify on a pack item, the town identify and sell chores acting on a created item, the purse read-back).

Default off: nothing here runs unless `loot.vault` 1 and `loot.vaultKeep` 1 (the Keep switch is the kill switch of all three errands); forced off in measured runs (`D2BOT_RUN`). New knobs (docs/knobs.md): `loot.vaultSellPerVisit` (10), `loot.vaultSellPrompt` (0), `loot.vaultPoorPickup` (1).

### Identify stays a step (user decision 2026-11-16)

An unidentified record shows as unidentified. Reveal costs one identify scroll or one tome charge (the fork's identify verb or the town's click path checks it; the vault never keeps its own count) and is FREE when Deckard Cain stands in the town the hero is in (his Identify line; `Identifier.Free`, the same Cain test as the inventory page). The bot never buys scrolls for Reveal: with none the press is refused with `no scroll; buy on request` (code `noscroll`) and the page offers the existing manual Buy (`POST /api/inventory/buyid`). `town.identifyTome` stays 0.

### Reveal flow

`POST /api/vault/reveal {ids, now?}` only QUEUES (`Store.RevealWant`, flag `RevQ` on the record). Idempotent (queued: `already`), refusals: `state` (identified, sold, discarded), `noscroll` (more requests than charges, none at Cain), gate codes `off`, `dead`, `verbs`, `stopped`. The page shows the cost BEFORE the click from `GET /api/vault` (`act.reveal`: `charges`, `free`, `perItem`): "Reveal 3 items: 3 charges (free at Cain)" or "...(you have 1): no scroll, buy on request".

Executor (`Keeper.runReveal`, same safe-moment rules, per-visit cap `loot.vaultKeepPerVisit` and 8 s budget as Keep). Per record, each stage logged BEFORE its game step (`Store.SetStage`):

1. `rb`: `item create` of the recorded bytes into the PACK (`key=vr-<id>-<attempt>`), save backup before the first create of the process (as Keep), read-back must equal the record, a mismatch removes the copy and trips the keeper (nothing is created until restart).
2. `rc`: identify the unit (scroll or tome charge, or Cain), read the item back: it must be identified with the same code, quality and item level.
3. The vault record is superseded IN PLACE (`reveal` event: same id, new raw bytes and hash, `identified`, name from the town, value recomputed; no second record). Stage `rt`.
4. `item vault <unit>` removes the temporary copy; stage cleared. No save is needed: the net change in the hero is zero.

Keep and Reveal on one record are ONE flow: the Keep create is made into the backpack (identify works on pack items), identified there, the record is superseded and kept; one create, nothing removed (`TestRevealKeepIsOneFlow`). A Keep+Reveal item therefore lands in the backpack and the stash chore files it (docs/vault.md Keep says stash: this is the one exception).

Failures (stored `reveal: <sentence>`, shown on the row and card): `no scroll; buy on request` (also if the charge disappeared while the request waited: nothing is created), `the game did not identify it: ...`, `the item is still unidentified`, `the game identified a different item`, an identical identified item already in the vault, pack full or hero dead or busy (stays queued), version or bad data. A failed identify removes the temporary copy and leaves the record exactly as recorded.

Restart recovery (`TestRevealRestartAtEveryStep`, 2 classes x keep or not x last call lost or not x 14 cut points): the stage tells what may exist; `resumeReveal` looks at the hero. Copy of the old bytes present: identify it. An identified copy of the same code, quality, level, sockets and ethereal bit that no vault record owns (`adoptable`; ambiguous means none): adopt its bytes. Stage `rt`: remove the copy of the new bytes. Nothing found: the stage is cleared and the request stays queued. After an identify whose reply was lost the hero decides (read back), so a charge is never spent twice. U: two identical unidentified items carried by the hero make "present" ambiguous; the create key and the kept-record count are the only guards.

### Sell-from-vault flow

`POST /api/vault/sell {ids, confirm:true, now?}` (the page asks first in a sheet with the count and the lower-bound price) queues (`sellq`). Refusals: `pinned` (never sold: also enforced in the store, `ErrPinned`, and again by the executor), `state`, gate codes. `POST /api/vault/cancel {ids}` takes a queued Keep, Reveal or Sell back while no step of it ran.

Executor (`Keeper.runSell`, cap `loot.vaultSellPerVisit`, 8 s budget), per record, one at a time:

1. purse read, stage `sb` (with the purse) logged, `item create` into the pack (`key=vs-<id>-<attempt>`), read-back equal to the record (else removed and tripped), stage `sc`.
2. `Seller.Sell`: the created item is marked sell on the inventory desk and the town's own sell chore runs (the same marked-items sale as "Sell now": real vendor, real price, nothing new decides what to sell; the chore also sells the player's own marks, which they asked for).
3. Read-back, in the game: the copy must be gone AND the purse must have grown. Both: `sold` with the gold read from the purse. Copy still there: it is removed again; a refusal by the vendor ends the sale with the reason (`sell: <why>`, record back in the vault), a postponement (fight, busy panel: `ErrPostpone`) leaves it queued. Copy gone but no gold: the record is closed `sold` with 0 gold and a note, the keeper TRIPS (no more sales until restart) and a Needs-you row says so.

Pack full: `the backpack is full`, stays queued, one Needs-you notice per episode. Restart: stage `sb` with no copy in the hero queues the sale again; a copy present continues at the sale; stage `sc` with no copy means it was sold, settled `sold` with the purse change as the gold (`TestSellRestartAtEveryStep`).

### Gold flow and the prompt

1. Poor-mode pickup (`loot.vaultPoorPickup`, default 1): with `loot.vault` 1 an item the Vault preset would only record is picked up instead when the hero has less than `loot.poorGoldBelow` gold and the item's `loot.SellValue` is at least `loot.poorMinSell` (today's poor-mode thresholds, no new numbers); an item the pickit keeps anyway keeps that verdict. `Looter.vaultPoorPick`, `TestVaultPoorModePicksSellableItems` (2 classes).
2. Sell errand above: the player queues, the bot sells at the next visit.
3. Opt-in prompt (`loot.vaultSellPrompt`, default 0): when `town.RepairEstimate + town.PlannedSpend - gold` is above 0 (the merc revive price is not in that sum: U), a Needs-you row offers "Short by N gold ...: sell K vault items (about G gold)?" with Sell them (`POST /api/vault/sell`) and Not now (`POST /api/vault/sellprompt {dismiss:true}`, muted 30 minutes). The list is the unpinned records of the lowest quality class present, up to quality magic (rare, set, unique and crafted are never offered), highest value first until the shortage is covered, at most 8. Nothing is sold without the answer. The same offer shows on the vault page.

### Page

Selection bar: "Reveal selected (n)" and "Sell selected (n)" next to Keep, the cost line (`act.reveal`, `revealPlan` in vault-logic.js), "Buy identify scrolls" when short, queue sentences ("Revealing 2 items: at your next town visit", "Selling 3 items: ..."), the prompt row. Rows: tags `R` (reveal queued), `...` (working), `$` (sale queued or sold), `!` (failed with the reason on the row, list and card); the card has Reveal, Sell and Cancel queued with the server's reason on a disabled button. Fit pages: `vault-revealing*`, `vault-reveal-{press,none,cain,dead}`, `vault-sell-sheet*` (`tools/uishots.sh`, fixtures `fixtureActor` in `tools/uishots/fixtures_vault.go`); the fit check counts the new rows above the list.

### Live drill checklist (not run; d2d fork with `--vault-verbs`, a throwaway hero, `loot.vault` 1, NO `D2BOT_RUN`, `vault creates: N` declared, `writes after start: 0`)

1. Record 10 drops with at least 3 unidentified magic items. Hero in town with 2 identify scrolls, no Cain in reach (Act 1 before the rescue). Select the 3: the line says "Reveal 3 items: 3 charges (you have 2): no scroll, buy on request"; press: refused `noscroll` for one, two queued.
2. At the next town visit: two creates then two `item vault`, scrolls 2 -> 0, records identified with real names, `vault stats` creates 2 vaults 2, hero pack unchanged, unidentified record count 1.
3. Rescue Cain (or use a hero past it), stand in town: the third reveals free, scrolls unchanged.
4. Kill the bot between create and identify (and once between identify and the reveal event): after restart exactly one scroll is spent and the pack holds no stray copy.
5. Keep + Reveal one unidentified item: one create, the item sits identified in the backpack, vault record `kept`.
6. Sell: select 3 unpinned magic items, Sell: sheet shows the lower bound; at the visit three create/sell cycles, gold rises by the vendor price each time, rows turn sold with the gold; a pinned item cannot be queued. Kill between create and sale: after restart one sale, no stray item.
7. Set the vendor out of reach (kill the sale mid-way): record returns to the vault with the reason, nothing stays in the pack.
8. Poor mode: gold under 300, drop a sellable rare: it is picked up (pickup click) not recorded; the gold line recovers after the next sell visit. Prompt: `loot.vaultSellPrompt` 1 with repair due and too little gold: the Needs-you row appears, Not now hides it for 30 minutes.
Record the `.bak` of the hero before step 2 and restore it if anything differs.

## Wallet (V7)

The account wallet is ETHER, implemented in `internal/ether` and described in [ether.md](ether.md): it implements `vault.Account` and `vault.Wallet`, derives the owned capacity steps and the balance by ledger replay, and serves `GET /api/ether`, `POST /api/vault/upgrade` and `POST /api/vault/downgrade` (`internal/watch/etherapi.go`). The store's `Options.Capacity` is `wallet.Capacity(override)`; shrinking (refund of the last step) is allowed only while the active hero's count fits the lower tier. Offline only, not yet wired into the bot.
