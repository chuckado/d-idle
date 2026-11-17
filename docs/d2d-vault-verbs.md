# Vault verbs (d2d fork) and the d2dctl client, slice V1

> status: current, 2026-11-16 (task TASK-584). Design: `docs/design/loot-vault-2026-11-16.md`; V0 spike result: `docs/probes/vault-spike-2026-11-16.md`. Tags: V = measured (how stated), U = unverified. Fork additions are private: fully documented, never upstream without the user's go.

## What shipped

Fork branch `didle/vault-verbs` on the private remote (built on `didle/vault-spike`; the commits are in the task result). Flag `--vault-verbs`, default OFF, not implied by `--lab-verbs`; without it every verb answers `err vault verbs off: ...` and `caps` lists none of them (V, test step 1). Reference text is the fork's `docs/control_channel.md` section "Vault verbs".

| Verb | Effect |
|---|---|
| `ground full [id]` | `#id raw= norm= at=x,y <view>` per non-gold floor item |
| `ground take <id>` | the same line, then the unit leaves the floor; refuses gold, a missing id, a dead hero, a joined game |
| `item full [id]` | the same lines for carried items |
| `item vault <id>` | the line, then the item leaves the hero (grid or belt); refuses worn, in-hand, socketed child, dead hero |
| `item create <hex> [stash\|inv] [col row] [key=<t>]` | creates from the `norm` hex; `err parse\|version\|dup\|nospace\|place\|state` |
| `vault stats` | `vault creates=N takes=N vaults=N` (the ledger; each change also logs a line) |

The verbs run on the main loop's pump, so each is atomic within a tick (V0 finding 3 asked for a Command round trip; not needed: the lab `debug give pack` path mutates the same way). U: no chest-proximity rule (lab verb; the bot walks to the chest first).

Go client in `internal/d2dctl/vault_verbs.go`: `VaultAPI` (implemented by `*Client` and by `*VaultFake`, `vault_fake.go`), typed `VaultRecord` with `Hash()` and `DedupeKey(game)`, `VaultRefused` with sentinels `ErrVaultOff`, `ErrVaultState`, `ErrStashFull`, `ErrBadHex`, `ErrVaultDup`, `ErrVaultNoItem`, `ErrVaultVersion`, `ErrVaultUnsupported`, `VaultCaps` (zero value = keep the pickup path), `ValidHex` (bad hex never reaches the wire), `Dedupe` and `CreateOnce`. Persist with `SaveCharacter` (the `save` verb; read-back and `.bak` belong to the save writer).

## V0 open issues, resolved

1. `uid` is 0 on every d2d item (V0: 860 of 860). The record identity is content: `sha256(norm)` plus game and floor unit and position (`DedupeKey`). Two layers: `Dedupe` in the bot process and the fork's per-process `key=` set; both tested (`TestVaultKeepFlowTwoClasses`: retry, bot restart, identical twin from another floor unit stays a record). The fork's old guard (same non-zero seed and code carried) stays and cannot fire on d2d-made items. Identical content without a key is allowed on purpose (two normal items of one code are legitimate).
2. 40 `debug give` calls that put nothing on the floor: cause found (V: read in `Loot::land`, confirmed by the test). `land` returns without adding when the code has no ground animation (`item_info` missing or empty `flippy`): the drill gave `bow`, a TYPE code, not a base item (`sbw` is the Short Bow), so the verb answered the previous item's id. The fork's `debug give` now answers `err nothing landed` (`vault_verbs_d2d.py` checks `bow` refuses and `sbw` works). Not a capture bug.
3. Untested kinds (fork test `tests/vault_verbs_d2d.py`, V on the builder, headless d2d, throwaway heroes, 0 failures):
   - ethereal and socketed: found by rolling normal items (`debug give plt|gth|lsd|gsd|mbt|buc`, seed 5), created and compared byte for byte (V). The test prints a NOTE if a seed rolls none.
   - charms and jewels (cm1..cm3, jew): V. Unidentified magic ring: V (carries its roll, round-trips).
   - dead hero: `ground take`, `item vault` and `item create` answer `err state`; `item full` still reads (V).
   - full stash: `err nospace` (V). Stash tab or shared stash: the classic fork has one stash grid and no shared stash (V0, code read); nothing to test.
   - the game running while creating: every verb runs in the live game (the test creates, takes and saves while the game ticks; V). Save, quit, reload keeps created items, `.bak` kept (V).
   - NOT possible headless: runewords and personalized items (no verb makes one; same writer, covered by the fork's `test_d2s` on real saves, S). Needs a save fixture or the game host: a V2 check.
   - Regression: `item_verbs_d2d.py` and `lab_verbs_d2d.py` still pass on the same build (V).

## How to rerun

Fork test on the builder: lab scripts `build.sh <rev>` (stage `git archive`, `-j2` nice) then `runtest.sh`. Go: `go test -p 1 ./internal/d2dctl/...`.
