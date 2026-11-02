# PR draft: a monster missile's launch record: physical brackets, HitShift and the row's own element (MISSILEDIFF, D1 / D2 / D3)

> status: NOT READY for the user's go until the independent verifier has RE-VERIFIED the fixed head (nothing pushed, opened or posted upstream). Local candidates on the private fork remote only: `local/missilediff-record` (head `c301eaef`, on the unsubmitted roll-top pair) and `local/missilediff-record-up` (head `ec09c73f`, three commits on upstream main 6935cf64, clean rebase). First verdict: NOT READY (`docs/research/pr-missilerecord-verifier.md`, branch agent/verifymissile): the claim held, the row poison had the wrong units; fixed here (card missilefix). Evidence: [missilediff.md](missilediff.md).

Title (fix commit subject): `A monster missile row's physical brackets, HitShift and own element reach the missile (FUN_0064b860)`

## What changes

game.exe makes a monster's MissA / MissS1 / MissC missile and fills its record with FUN_0064b860 (row path): physical min / max = `(owner damage x 256 x SrcDamage / 128, truncated) + ((MinDamage + MinLevDam bracket at the missile level) << HitShift)`, and the row's own element by FUN_0064b100 / 0064b1d0 / 0064b2a0 (poison per frame in 256ths for ELen frames, fire and the others as min / max).

d2d built the record inline in `Ai::tick` as `(owner x SrcDamage / 128) + MinDamage` in whole points: no level bracket (D1: Nightmare / Hell Quill Rat spikes, Pantherpot, regurgitator), HitShift ignored (D2: Sand Maggot spit 2-6 256ths became 1-3 points), no row element (D3: the poison of the Sand Maggot, Unravelers, Radament and the Night Slinger pots, the Slinger fire pot).

The change moves the expressions into `monster_shot_phys` / `monster_shot_stats` (`components/game/ai.hpp`; `Ai::tick` calls them), reads MinLevDam1..5 / MaxLevDam1..5 in the loader, and computes the record as game.exe does. The row's poison is carried in `MonStats::poison_min / poison_max / poison_len` (per frame in 256ths, frames), which `monster_blow` already applies on every hit; fire, cold, lightning and magic go in a free `elements` slot.

## Proof

`tests/game/test_monster_missile.cpp --monster-missile CASES` calls the real functions; `tools/emu/mm_diff.py` runs game.exe's FUN_0059fa30 -> FUN_0064b860 under the emulator on a real spawned shooter and compares line by line.

| build | cases | physical differ | elements (and poison) compared / differ |
|---|---|---|---|
| base (test mode + extraction only) | 120,000 | 32,690 | 55,769 / 28,300 |
| head `c301eaef` | 120,000 | 0 | 55,769 / 0 |
| rebased head `ec09c73f` | 120,000 | 0 | 55,769 / 0 |

Independent counts (the verifier's own harness, written from the disassembly; Act 1 / 2 rows and all 78 monster Miss rows of every act, two seeds): base 11,804 of 34,067 cases differ at the game's own levels 1 / 4 / 8, head 0; re-run on the fixed head with the poison compared per frame: 0 differences in physical, elements and poison (about 22,000 poison cases per set), on both the original and the rebased head. Rows that moved on base: goospit1-3, spike1-4, pantherpotorange / green, mummy1-4; every row with no bracket, HitShift 8 and no own element is byte-identical base against head.

The consumer (what the hit applies, the verifier's blocker): `monster_blow` on goospit1 gives 8 a frame for 200 frames at level 1 and 50 at level 8 (6.25 and 39 points), mummy4 88 for 400 at level 8 (137.5), pantherpotgreen 102-108 for 400 (159-169): the numbers of game.exe's record. `test_monster_missile` (no arguments, no game data) asserts them; it fails on the first version of the fix (poison 60 a frame for 400 frames at goospit1 level 1, 94 points).

Checks (V, run on the worker host): `pr_guard` 0 FAIL / 0 WARN; `presubmit --quick` 0 FAIL; ruff 0.8.6 clean; clang-tidy clean on the touched files except pre-existing `outdoor.hpp` findings; `test_game`, `test_combat`, `test_game_rules`, `test_rules`, `test_monster_ai`, `test_monsters` pass; the 14 deterministic scenarios hash identically base against head (no affected row fires in them); a drill (Hell Night Slinger pot, 500 frames, player poison resistance -100): old first version 5,922 life, now 359.

## Not in this PR

- D5, the whole-point `>> 8` of `a2_min` / `a2_max` in `Ai::tick` and `monster_blow`'s whole-point roll (game.exe rolls 256ths): the record is exact, the roll is not; a physical part under one point (Sand Maggot) stays 0 on the roll. Separate change with a bigger blast radius.
- Skill-sourced rows (`andariel_missile` / `missile_blow`): the poison there has the same wrong-units bug (Andariel's poison bolt x 62 of the record, on upstream main too). Separate PR: `local/missilepoison-skillrows` (`70fa479d` on upstream main, one commit and a no-data test). Named here as the sibling.
- The imp double fire (Act 5: SrcDamage 0 should drop the owner's El rows; d2d carries them and the row's fire is added) and a SrcDamage `-1` (game.exe: 255, treated as 0 here; no monster Miss row has one): known gaps, low.
- MissS1 / MissC / MissSQ rows are loaded only when also MissA1 / MissA2 / skill rows (`gamedata_load.cpp`): the Council's lightning, the wisps' bolts and Blood Raven's raven2 never fire (pre-existing).

## Limits

- U: a shooter that has its own poison hit (a champion or unique) gets the row's poison added per frame; the length is the larger of the two (game.exe's FUN_0064a930 +0x3c sum not traced). No Act 1 / 2 MissA row has an owner poison.
- The row element takes a free `elements` slot of three: no shooter in monstats has three El rows of the shot's mode (D, all acts).
- Not run: an MSVC build, a live game, the author's drill M2 on a built lab binary after the fix (the verifier's `spit_*` det scenarios do the same job on the lab patch set).
- Tests that fail on the box and not through this change: `test_mpq` (integrity assertion on this MPQ folder), `smoke_d2d` (no d2speech), `quests_act1` / `quests_act4` (an Andariel / Diablo quake timeout; the same assertion on a lab binary without the candidate).

## Base (the sequencing problem)

The first candidate sits on the roll-top pair 9ba8d369 / 23e2baec (SKMISS: a missile element roll never reaches its max), which is neither upstream nor an open PR (read-only `gh pr list`, 2026-11-01). Options:

1. Open the roll-top PR first, then this one on top of it (the pair is two commits; this PR then merges on a main that has it). Slow, keeps one history.
2. Open this PR from `local/missilediff-record-up`: the three candidate commits (`554067e0` test mode and extraction, `4224a2be` the record, `ec09c73f` the poison units) cherry-picked onto upstream main 6935cf64 with no conflict (one auto-merge in `ai.hpp`), tested the same way. It does not need the roll-top pair: the claim is the record, the pair is the roll. Recommended: the record PR stands alone.

## How to review

1. `components/game/ai.hpp`: `monster_shot_phys`, `monster_shot_stats` (the whole change; `ai.cpp` calls them).
2. Disassemble FUN_0064b860 (row path), FUN_0064af20 / 0064aff0, FUN_0064b100 / 0064b1d0 / 0064b2a0 with `tools/emu/disasm.py`.
3. Run `tools/emu/mm_diff.py game TABLES N SEED OUT`, then `test_monster_missile --monster-missile OUT/cases.txt | grep -E '^[0-9]+ ' > d2d.txt` and `mm_diff.py cmp OUT d2d.txt`. The unit test: `test_monster_missile` with no arguments.

## Files

`components/game/ai.hpp`, `components/game/ai.cpp` (the call), `components/game/gamedata/gamedata.hpp`, `components/game/gamedata/gamedata_load.cpp`; tooling `tests/game/test_monster_missile.cpp`, `tools/emu/mm_diff.py`, `tools/emu/mm_launch.py`.

## What the re-verifier must check

1. The record claim on the fixed head (`c301eaef`) and on `ec09c73f`, with the poison read as per frame: his harness with the poison compare of `vm_cmp_p.py` (copy in the author's worker directory), 0 differences.
2. The consumer: `monster_blow` on the produced `src` equals the record's poison total (goospit1, mummy4, pantherpotgreen at 1 / 4 / 8), including the player's poison length reduction (`plr`) and resistance paths.
3. U items: the owner-poison add (length `max`), and that no other consumer of `MonStats::elements` type 3 assumed the row's total (merc, pet, co-op targets).
4. The split: that `local/missilepoison-skillrows` is a sound sibling and the candidate does not need it.
5. The base-equals-head scenarios and `test_game` / `test_combat` on the rebased branch.

## Shared-file lines for the merger

In `openspec/changes/missilefix/merge-lines.txt`.
