# PR draft: a monster missile's launch record: physical brackets, HitShift and the row's own element (MISSILEDIFF, D1 / D2 / D3)

> status: DRAFT held for the user's go (nothing pushed, opened or posted upstream). Local candidate on the private fork remote only: `local/missilediff-record-up` (head: see the SrcDamage-byte commit at the end of this line's branch, `1824870b` plus one; four commits on upstream main 6935cf64: `554067e0`, `4224a2be`, `ec09c73f` and the owner-poison commit `1824870b`); the older `local/missilediff-record` (`c301eaef`, on the unsubmitted roll-top pair) does not have the last commit. Verdicts: first verifier NOT READY (`pr-missilerecord-verifier.md`, branch agent/verifymissile; poison units, fixed); independent re-verifier on `ec09c73f` READY WITH CHANGES (`docs/research/pr-missilerecord-reverify.md`, branch agent/reverifymissile; the owner-poison add and the PR text), both done in card missilefix2. The narrow third check is done: `docs/research/pr-missilerecord-reverify.md`, section "Third check" (READY WITH CHANGES; the changes are done in card missilefix3). Sibling PR: [pr-missilepoison-draft.md](pr-missilepoison-draft.md). Evidence: [missilediff.md](missilediff.md).

Title (fix commit subject): `A monster missile row's physical brackets, HitShift and own element reach the missile (FUN_0064b860)`

## What changes

game.exe makes a monster's MissA / MissS1 / MissC missile and fills its record with FUN_0064b860 (row path): physical min / max = `(owner damage x 256 x SrcDamage / 128, truncated) + ((MinDamage + MinLevDam bracket at the missile level) << HitShift)`, and the row's own element by FUN_0064b100 / 0064b1d0 / 0064b2a0 (poison per frame in 256ths for ELen frames, fire and the others as min / max).

d2d built the record inline in `Ai::tick` as `(owner x SrcDamage / 128) + MinDamage` in whole points: no level bracket (D1: Nightmare / Hell Quill Rat spikes, Pantherpot, regurgitator), HitShift ignored (D2: Sand Maggot spit 2-6 256ths became 1-3 points), no row element (D3: the poison of the Sand Maggot, Unravelers, Radament and the Night Slinger pots, the Slinger fire pot).

The change moves the expressions into `monster_shot_phys` / `monster_shot_stats` (`components/game/ai.hpp`; `Ai::tick` calls them), reads MinLevDam1..5 / MaxLevDam1..5 in the loader, and computes the record as game.exe does. The row's poison is carried in `MonStats::poison_min / poison_max / poison_len` (per frame in 256ths, frames), which `monster_blow` already applies on every hit; fire, cold, lightning and magic go in a free `elements` slot.

The owner's own poison hit (a poisonhit champion or unique) follows game.exe's FUN_0064b860 -> FUN_0064a930 / FUN_0064aa60 (V read, and V run against the re-verifier's oracle): it is added only when the row's SrcDamage is not 0, and then the rate pair (owner + row) is scaled by SrcDamage / 128 and the lengths are summed. All 32 poison rows of Missiles.txt have SrcDamage 0, so for every poison row the owner's poison is dropped and the shot carries the row's alone (the last commit; the head before it added the owner's unscaled and kept the larger length).

## Proof

`tests/game/test_monster_missile.cpp --monster-missile CASES` calls the real functions; `tools/emu/mm_diff.py` runs game.exe's FUN_0059fa30 -> FUN_0064b860 under the emulator on a real spawned shooter and compares line by line.

| build | cases | physical differ | elements (and poison) compared / differ |
|---|---|---|---|
| base (test mode + extraction only) | 120,000 | 32,690 | 55,769 / 28,300 |
| head `c301eaef` | 120,000 | 0 | 55,769 / 0 |
| rebased head `ec09c73f` | 120,000 | 0 | 55,769 / 0 |

Independent counts, the re-verifier's own harness and own runs (`pr-missilerecord-reverify.md`, sections 3 to 5; Act 1 / 2 rows and all 78 monster Miss rows of every act, two new seeds): on `ec09c73f` 0 differences in physical (108,140 and 112,760 cases), other elements (16,903 and 7,916) and poison per frame (23,884 and 23,926); through the real `monster_blow` 0 of 18,092 (seed C, four defender settings) and 0 of 22,743 (all acts) poison hits differ from game.exe's record; game.exe's resist stage (FUN_0057c1e0, run natively) equals `monster_blow` in all 5,926 in-domain cases. On the base the first verifier counted 11,804 of 34,067 cases differing at levels 1 / 4 / 8. Rows that moved on base: goospit1-3, spike1-4, pantherpotorange / green, mummy1-4; every row with no bracket, HitShift 8 and no own element is byte-identical base against head.

Re-run by the author on the new head `1824870b` with the re-verifier's harness unchanged (his case files and his game.exe record files; the author's runs, not a second verification): seed C and all acts, poison, other elements and physical 0 differences (108,140 and 112,760 cases, 23,884 and 23,926 poison compared), the poison hits 0 differences at four defender settings (res / plr 0 / 0, -100 / 0, 40 / 25, 75 / 50), and the owner-poison set (2,176 cases: 17 rows x levels x 4 owner poisons x SrcDamage {the row's own, 128, 64, 255}) 0 differences in poison (2,112 compared; the head before it differed in 1,776 of them: the third check's count with his oracle, counting the patched SrcDamage cases). The dot set (resist stage) is unchanged code: 5,926 in-domain cases, 0 differences.

The consumer (what the hit applies, the verifier's blocker): `monster_blow` on goospit1 gives 8 a frame for 200 frames at level 1 and 50 at level 8 (6.25 and 39 points), mummy4 88 for 400 at level 8 (137.5), pantherpotgreen 102-108 for 400 (159-169): the numbers of game.exe's record. A player's Hell penalty moves both the poison resistance and the length reduction (Nightmare -40, Hell -100 on each; V run in the re-verifier's oracle), so a Hell pot hit on a player is x 4 in total (rate and length each x 2: a full pot with no resistance is 204-216 x 800 = 637-675 points), not x 2. `test_monster_missile` (no arguments, no game data) asserts them; it fails on the first version of the fix (poison 60 a frame for 400 frames at goospit1 level 1, 94 points). It also asserts the owner-poison rule: goospit1 with an owner poison of 40 / 40 / 300 gives 8 / 8 / 200, and with SrcDamage patched to 128 / 64 / 255 gives 48 / 48 / 500, 24 / 24 / 500 and 95 / 95 / 500 (the re-verifier's oracle numbers).

Checks (V, run on the worker host): `pr_guard` 0 FAIL / 0 WARN against 6935cf64; `presubmit --quick` 0 FAIL, 4 WARN (answered under Limits), ruff clean, short names ok; ctest `test_game`, `test_combat`, `test_game_rules`, `test_rules`, `test_monster_ai`, `test_monsters` and `test_monster_missile` 7 of 7 pass on `1824870b`; the first-round checks (clang-tidy on the touched files, the 14 deterministic scenarios identical base against head) were on `ec09c73f` and are not re-run for the last commit (it touches only `monster_shot_stats` and its test). The author's earlier drill (Hell Night Slinger pot, 500 frames, player poison resistance -100: 5,922 life on the first version, 359 now) is S: the re-verifier did not re-run it and notes that its "x 2 for the resistance" explanation is incomplete (the length doubles too and the 500-frame window cuts it); it is not a claim of this PR.

## Not in this PR

- D5, the whole-point `>> 8` of `a2_min` / `a2_max` in `Ai::tick` and `monster_blow`'s whole-point roll (game.exe rolls 256ths): the record is exact, the roll is not; a physical part under one point (Sand Maggot) stays 0 on the roll. Separate change with a bigger blast radius.
- Skill-sourced rows (`andariel_missile` / `missile_blow`) and the trap missiles (`world/objects.cpp`, x 20 x ELen / 256: the trap d2d fires is primepoisoncloud, x 23 / 28 / 34 at trap levels 1 / 4 / 8): the poison there has the same wrong-units bug (Andariel's poison bolt x 62 of the record, on upstream main too; V run by the re-verifier). Separate PR: `local/missilepoison-skillrows` (`71f98145` on upstream main: the Andariel split `70fa479d` and the trap fix, with no-data tests), draft [pr-missilepoison-draft.md](pr-missilepoison-draft.md). Named here as the sibling.
- Death clouds (`fight/deaths.cpp:135`, the mummies' `corpsepoisoncloud` on death): the rate is right to within 9 / 256, the length doubles (2 x ELen frames): x 2 in all. In neither PR; named as a follow-up.
- The top of the poison roll: game.exe's FUN_0057a880 rolls `min + rand(max - min)` and never reaches the max, `monster_blow` rolls inclusive (pre-existing, low; the same top-of-range issue as the roll-top pair).
- The imp double fire (Act 5: SrcDamage 0 should drop the owner's El rows; d2d carries them and the row's fire is added) known gap, low. SrcDamage `-1`: game.exe's plain shot path reads the byte as 255 (V read: `movzx` at 0x64b8ed). The owner-poison half now does the same (`src_damage & 0xff`); the physical half (`monster_shot_phys`) still treats `<= 0` as 0 and so differs from game.exe on a -1 row. Not reachable: only player skill missiles (poisonjavcloud, plaguejavcloud, furylightning) have -1, no monster Miss row.
- MissS1 / MissC / MissSQ rows are loaded only when also MissA1 / MissA2 / skill rows (`gamedata_load.cpp`): the Council's lightning, the wisps' bolts and Blood Raven's raven2 never fire (pre-existing).

## Limits

- Owner poison: traced and matched (V read of FUN_0064b860 / FUN_0064a930 / FUN_0064aa60, V run on the re-verifier's oracle, 0 differences on the owner-poison set). U: whether a cast (MissC) shot takes the same path. The rule is applied to every row, as game.exe's record does (a SrcDamage-0 row drops the owner's poison, a SrcDamage 128 row such as bighead1 / spike1 keeps it). The shooters that carry owner poison in d2d are Baal Subject 2 (unraveler5, poisonhit, MissA mummy4) and the Maggot Queen class when spawned as a non-boss; Coldworm the Burrower: U.
- Presubmit WARNs, answered: *siblings, unit kinds*: a player, merc, pet and co-op partner are all hit through the same `monster_blow` with their own Fighter, so the units hold for all of them (the merc's and pet's plr / res penalty: S, not re-read); *siblings, hit paths*: the skill-sourced path and the trap path are the sibling PR, the death clouds are named above, melee has no row poison; *docs-sync*: no doc in the repo describes the monster missile record or its poison (the only hit for `missile_blow` is `docs/research/re/skills.md` line 671, the player's `rules::missile_blow`, unchanged), so no `.md` changes; *build*: answered by the ctest runs above.
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

## Third check

Done by the independent re-verifier: `docs/research/pr-missilerecord-reverify.md`, section "Third check" (heads `1824870b` and `71f98145`; owner-poison rule 0 of 2,112 differ, merge clean, 8 of 8 suites). Its change list (1,776 count, the trap row, SrcDamage -1) is applied in this draft and in the code.

## Shared-file lines for the merger

In `openspec/changes/missilefix/merge-lines.txt`.
