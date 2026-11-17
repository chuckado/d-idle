# Ethereal Vault store (internal/vault) and its file formats

> status: BUILT offline (slice V2 of docs/design/loot-vault-2026-11-16.md). Additive and unused until V3 (rules action) and V5 (keep queue) call it. Not run against a live game: every claim below is about the offline package (V: `go test ./internal/vault/`). Item hex comes from the fork codec later (V1); here it is opaque bytes.

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
