# Ethereal Vault V0 spike: capture and create on the d2d fork (result)

> status: RESULT, 2026-11-16 (task TASK-578). Design: `docs/design/loot-vault-2026-11-16.md` section 1.7. Tags: V = measured in this spike (how stated), U = unverified. Author: agent vaultv0 (Claude).

## Verdict: GO for V1 (with the four changes under "Findings that change V1")

On a headless d2d built on the NAS builder, `ground full` captured 860 of 860 ground items, `item create` put 820 distinct captures into a hero (stash) with byte-identical re-serialisation (820 of 820), 40 of 40 kept ones survived save, quit and reload byte-identical, and 40 of 40 created into a second fresh hero were byte-identical too. No failure of the codec path was seen (V, `drill.py`, below). Not done: the optional Game.exe oracle (needs the game host; nothing ran there).

## What was built (private fork, pushed)

Fork branch `didle/vault-spike` on the private remote, commits `5484ac94` (verbs), `cb44e33e` (purge), `b4d2f861` (docs in the fork's `docs/control_channel.md`); base `c2549842` (`sandbox-mercverbs`). New file `apps/d2d/devctl/vault_verbs.hpp`, small hooks in `verbs.hpp`, `debug.cpp/hpp`, `main.cpp`. Flag `--vault-verbs` (default off, NOT implied by `--lab-verbs`); without it each verb answers `err vault verbs off: <verb> needs d2d --vault-verbs` (V, `flagoff.py` run: all four refuse, plain `ground` still works).

Verb shapes that worked (V):
- `ground full [id]` -> `#<id> raw=<hex> norm=<hex> <view>` per non-gold ground item. `raw` is `write_item` output (file version 0x60) as the item lies; `norm` is the same item with its place cleared (stored, stash, col 0, row 0) so copies compare whatever the cell. The view lists code, q, ilvl, uid, id'd, eth, sockets, def, dur, qty, prefix, suffix, rare ids, affix ids, set, unique, runeword id, props `stat:param:value`, children count.
- `item full [id]` -> the same lines for carried items.
- `item create <hex> [stash|inv] [col row]` -> `create id=<new unit id> at=<col>,<row>`; parses with the save reader (`d2s::detail::item`), then `rules::free_spot` and `put_in_grid`. Refusals: `err parse ...` (odd or non-hex, truncated: "read past end", trailing bytes), `err version <n>` (a byte-flipped record read as version 154), `err dup ...`, `err nospace ...`, `err place ...`. Applied in the same call (the spike mutates the item list directly like `debug give pack`; V1 should use a command like the item verbs).
- `item purge <id>`: drill housekeeping only.

## Drill (V)

`drill.py` inside the builder container, `d2d --headless --lab-verbs --vault-verbs --seed N` on a throwaway hero in a temp HOME (declared in `docs/probes/test-characters.md`), 5 rounds (seeds 21..25, classes alternating Amazon and Paladin). Per round: 172 ground items by `debug give` (20 base codes x qualities 1..8 plus 12 unidentified magic and rare of 6 codes), `ground full` each, `item create` each, `item full` of the copy compared with the capture (norm hex), 8 kept, the rest purged; `save`, quit, reload from the `.d2s` via the character select, compare the 8 kept; then the first 8 into a second fresh hero. Build: `didle/vault-spike` at `cb44e33e`, d2d target at `-j2`, nice; drill under `nice -n 10`, one process at a time. Raw rows: the drill output of run `o2` (results.jsonl, kept on the builder).

| Measure | Result |
|---|---|
| ground items given and captured | 860 of 860 (100%) |
| distinct captures that reached `item create` | 820 (40 `debug give` calls put nothing on the ground: bow q1..7 and aqv q8 each round, the verb answered the previous item's id; U cause, a `give` quirk, not the capture path) |
| created OK | 820 of 820 |
| bytes equal (norm hex of the created copy vs capture) | 820 of 820 (100%) |
| reload equal (norm hex found after save, quit, reload) | 40 of 40 (100%) |
| created into a second fresh hero, bytes equal | 40 of 40 (100%) |
| unidentified items captured, with a real roll in the unit data | 60 of 60 (U1 settled below) |
| failures | 0 |
| `save` replies | ok x5; `.d2s.bak` present each time (character_store keeps it, V `bak: true`); saved sizes 1.2 to 1.9 KB |

Qualities covered: 1 to 8 (low, normal, superior, magic, set, rare, unique, crafted), armor, weapons, jewelry, charms, jewel, stackable (arrows), tome, rune, gem, misc. Not covered, because `debug give` cannot make them: ethereal, socketed with gems, runewords, personalized (U: the codec path handles them, same writer, but this spike did not exercise them; the V1 corpus test must).

## Open unknowns settled

- U1, do UNIDENTIFIED ground items carry their real affixes? YES (V): 60 of 60 unidentified magic and rare items had prefix/suffix or rare affix ids and props from the moment they exist, `identified` is a separate flag. Identify check: a magic ring created unidentified into the pack, `item identify <id>` with a scroll: the view before and after are equal except `id'd` 0 to 1 (same prefix, suffix, props), so identify reveals and does not re-roll (V, one item per round 0 run; one sample, U beyond it).
- Item id collisions (V): unit ids are fresh per create (`max carried id + 1`); the created copies never clashed; the id is not stored in the save.
- Shared stash / version / checksum (V): classic fork has no shared stash (code read). The item `version` 101 round-trips; an altered version refuses with `err version`. The save writer's read-back and `.bak` ran on every save (`save` returned ok each time); I did not recompute the checksum separately, the fork's own `write_save` does.
- Failure modes (V): stash full (60 more rings: `err nospace no room there`); truncated hex: `err parse d2s items: read past end`; non-hex: `err parse not hex`; corrupt byte: `err version 154`; flag off: `err vault verbs off`. "Game running": the verbs only run while the game runs; there is no offline path in the spike (Option B of the design stays a V9 tool). Hero dead and panel-open were not tested (U).

## Findings that change V1

1. `uid` (the seed) is 0 on every d2d-generated item (V: `uid=0` in all 860 captures; `rules::generate_item` rolls with an internal seed and never stores it). The record does not need the seed: the roll is explicit in the affix ids and props, and re-creation is byte-exact without it. But the design's `dup` guard on `uid`+`code` does not work (it never fires, V: a dup of an existing record was created). V1 must dedupe by the DB state machine (a record created once, `kept` final), or use a hash of the whole `norm` hex, not the seed. Also drop "seed" from the record's identity.
2. `norm` hex (place cleared) is the right comparison key; `raw` carries the cell and is not stable across copies.
3. The create must go through an item command like `plan_move` (applied at the next tick, chest proximity, fresh unit id, ledger line `vault creates: N`); the spike mutates the list in the verb thread.
4. `ground take` and the crash-safe order of design 1.2 were not part of V0; still open, small (remove a ground unit by id).

## Not done

Game.exe oracle (load the saved hero in the real game and read the item back): not cheap here, needs the game host. Ethereal, socketed, runeword and personalized items; dead hero; panel open. Whole-repo ctest of the fork (the builder's ctest is a V1 task).

## Commands (reproduce, on the builder)

Build: `build.sh <rev>` (lab script) (stage `git archive`, cmake Ninja, `ninja -j2 d2d`, about 25 min cold, ccache cold because the source path is new). Drill: `rundrill.sh o2 5 21` (lab script) (runs `python3 drill.py <d2d> <out> <rounds> <seed>` with `D2_MPQ_DIR` on the builder's game data). Both are lab scripts in the lab directory, not in this repo.
