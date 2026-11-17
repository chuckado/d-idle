# Ethereal Vault store (internal/vault) and its file formats

> status: V2 store BUILT offline, V3 bot wiring BUILT offline (section 'Bot wiring (V3)'), V4 read API and page BUILT offline (section 'UI and API (slice V4)') (slice V2 of docs/design/loot-vault-2026-11-16.md). Additive and unused until V3 (rules action) and V5 (keep queue) call it. Not run against a live game: every claim below is about the offline package (V: `go test ./internal/vault/`). Item hex comes from the fork codec later (V1); here it is opaque bytes.

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

What it is: a read side plus Pin, Unpin and Discard of the Ethereal Vault. No Keep, Reveal, Sell or Upgrade (V5, V6, V7): the Keep and Upgrade buttons are drawn, disabled, with the reason in text (`aria-disabled` and a description, never a dead press: pressing says why).

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
- Selection: Select button, shift-click range, ctrl/meta-click, long press (500 ms), "Select all shown", Escape clears. Pin, Unpin, Discard act on the selection. Discard opens a modal (focus loop, Escape, default focus on Cancel); above 10 items the count must be typed. The Keep button is disabled with "Keep arrives with the Keep queue (V5)".
- A wide window (900 px and up) shows the card of the open item beside the list (state, lines, where found, Pin or Unpin, Discard, disabled Keep). On a phone a tap shows the tooltip as a bottom sheet and actions go through Select.
- States: loading, error (retry), empty vault (explains the Ethereal Vault), off (read only banner), nothing matches (clear filters), over capacity, full. The list polls every 6 s and reloads only when the total, count or first row changed; pages of 60, "Show more" and an end-of-list sentinel up to 600 rows (search narrows the rest).
- Accessibility: every control has a name (the fit check fails one without), targets are at least 44 px, cells are buttons with a full label and `aria-pressed` for selection, arrows move between cells, no motion.

Measured (V, `TestVaultPagingPerformance`, 10,000 records, 4-core box, GET handler in process, mean of 10): first page 2 ms, sort by value 5 ms, sort by name from offset 5000 16 ms, kind and ilvl sort under 1 ms, name search with sort 2 ms, facets 1 ms. The ceiling the test enforces is 150 ms.

Fit and pictures: `tools/uishots.sh --fit` has the pages `vault-*` (states pop, warn, full, over, empty, none, off, error, big; list view; filters open; select; the confirm sheet with and without the typed number; tooltip hover; the menu and More entries) at the usual sizes plus 412x780 and 360x640, and the must-fail `vaultdefect-*` (clipped text, wide grid, list under the header, small target, covered control, cut sheet, unnamed control, enabled placeholder). `tools/uishots.sh --vt` writes the pictures (`vt-NAME-WxH.png`). Fixtures: `tools/uishots/fixtures_vault.go` (real stores, frozen drops).

Dead ends and not done: no loading-state picture (headless Chrome's virtual time waits for the pending request, so the harness never sees it); the Keep, Reveal, Sell, Upgrade verbs; "Recently trimmed" list; property-text search (needs the codec's view); compare with the worn item (needs the inventory view on this page).

## Wallet (V7)

The account wallet is ETHER, implemented in `internal/ether` and described in [ether.md](ether.md): it implements `vault.Account` and `vault.Wallet`, derives the owned capacity steps and the balance by ledger replay, and serves `GET /api/ether`, `POST /api/vault/upgrade` and `POST /api/vault/downgrade` (`internal/watch/etherapi.go`). The store's `Options.Capacity` is `wallet.Capacity(override)`; shrinking (refund of the last step) is allowed only while the active hero's count fits the lower tier. Offline only, not yet wired into the bot.
