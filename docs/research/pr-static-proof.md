# PR proof: Static Field hits as FUN_005c96a0 does

> status: local candidate, PROOF incomplete (independent verifier not run), nothing posted (card decompstatic, 2026-10-31). Trace: docs/research/static-field-chill-trace.md. Standard: docs/process/upstream-pr-standard.md.

## Claim

d2d's Static Field (`kStaticField` in `Fight::spot`, components/game/fight/strike.cpp, upstream 72132b99) takes `hit_points * pct / 100` whole points through `resisted`. game.exe's per-unit callback FUN_005c96a0 (from FUN_005c9800 via the area walk FUN_0056b7e0) takes the life's whole points (stat 6 >> 8), skips the unit at or under StaticFieldMin % of max life (expansion games), caps at life - 1, keeps at least calc2, cancels a negative lightning resistance, then applies the record in 256ths through FUN_0057c6c0 with the attacker's pierce. One behaviour (the per-unit damage), one commit.

## Address chain (all V)

| Link | Evidence |
|---|---|
| FUN_005c9800 builds params {calc2, calc1, expansion ? StaticFieldMin : 0, ELen, EType} | disasm 0x5c9800..0x5c98ed; loader descriptors calc1 +0x138 (0x616fc5), calc2 +0x13c (0x616feb), DifficultyLevels StaticFieldMin +0x40 (0x61805b) |
| FUN_005c96a0 rule | disasm 0x5c96a0..0x5c97f9 and the oracle (native) |
| FUN_0056c8e0 lightning -> +0x1c, res stat 0x29 | disasm (jump table 0x56ca0c) and the oracle (native) |
| FUN_0057c6c0 / FUN_0057c1e0 / FUN_0057bf80 / FUN_00483360 | native in the oracle (PR #135's life_apply.py hooks reused) |

## Differential

`tools/emu/static_field.py` (game.exe) vs `build/tests/test_game --static-field 1 100000` (d2d): 309,664 cases (209,664 boundary: floors 0/33/50 x 13 resistances x max life 1..6 x every life in 256ths; 100,000 random: max life 1..3000, res -200..150, pierce 0..60, floor 0..100, calc1 0..150, calc2 0..0x2000). Before (line-for-line copy of the old arithmetic in the oracle, cited above): 115,103 differ. After (the real d2d build): 0 (`grep -E '^[BR][0-9]'` of its output; game.txt md5 5583e13b).

Hooks (not the claim): stat getter / setter, max life, unit-kind predicates, difficulty row, events, logging, side effects of FUN_0057c6c0, and FUN_0057cee0 (after-hit).

## Scope and callers

`static_field_fine` is new; `resisted_fine`, `land`, `hurt` unchanged. The only caller is `kStaticField`. Upstream test_game (rc 0, 20 OK), test_rules (rc 0), test_combat (rc 0); pr_guard 0 FAIL 0 WARN; MSVC shadowing sweep: no new declaration hides an outer local or a Fight member. Not run: presubmit --quick, det state hashes.

## Limits

Not in the claim (ponytail in fight.hpp): the area walk (players as targets, line of sight 0x200, integer squared distance); pierce through immunity (the life_apply domain); co-op replication of life_frac. Verifier: not run.

## Independent verifier (card verifyprs, 2026-10-31)

Verdict: **HOLDS WITH LIMITS** (limits below; none is a counter-example inside the claim).

What was done, without reading the author's oracle first:

- Rule re-derived from disasm.py reads: FUN_005c9800 builds the block {+0 calc2 (skill +0x13c), +4 calc1 (+0x138), +8 game +0x70 ? DifficultyLevels +0x40 : 0, +0xc FUN_00644f20 length, +0x10 EType (+0x1dc)}; FUN_005c96a0: life = stat 6 sar 8, none under 1; with a floor, none at or under FUN_00483360(max life sar 8, floor, 100); dmg = MulDiv(life, calc1, 100), life - 1 if that leaves under 1; dmg shl 8, at least calc2; FUN_0056c8e0 (jump table 0x56ca0c: EType 2 -> record +0x1c, resist stat 0x29); a negative stat 0x29 gives MulDiv(dmg, 100, 100 - res); the record goes to FUN_0057c6c0 (flags 0x4d, hit flags 0x4001). In FUN_0057c1e0 / FUN_0057be00 the defender flag is "a monster that is not a hireling" (FUN_0063ee90 returns the hireling act, so the flag is 1 when it returns 0): pierce (stat 0x14e, table 0x732980 row 2) comes off only when res < 100, no difficulty penalty, clamp at -100; FUN_0057bf80 then takes MulDiv(v, 100 - min(res, 100), 100). Same rule as the author's (V, disassembly read).
- Own differential (`tools/emu/verify_static_field.py`, seed 4242): game side = game.exe's own FUN_005c96a0 then FUN_0057c1e0 natively under the emulator (hooks listed in the script docstring); d2d side = the REAL `Fight::spot` (a World, one monster, the player's pierce in psum) via `test_game --verify-static` (`verify_static_field_mode.inc`), not a copy of the call site. 232,152 cases: 132,152 boundary (every life in 256ths for max life 1..4 at 5 resistances x 3 difficulties x 2 game types; floor edges +-1 point and +0/1/255 256ths for max life 100..65,535; 13 resistances x 6 pierces; synthetic calc1 0..1000, calc2 0..0x10000, floors 0..100) and 100,000 random (max life to 2,000,000, res -300..300, pierce to 300, slvl 1..60). Before (upstream 72132b99 built): **100,624 differ** (94,243 with real data, 6,381 synthetic). After (fix/static-field built): **0**. game.tsv md5 prefix bd1714639a1b.
- Sweep: d2d has one Static Field path (`kStaticField` in strike.cpp); StaticFieldMin is read in game.exe only at 0x5c988b (the 23 FUN_00611d30 callers checked); FUN_005c96a0 is referenced only from 0x5c98d0, FUN_005c9800 only from the srvdo table.
- CI class: pr_guard 0 FAIL 0 WARN; test_game / test_rules / test_combat / test_protocol rc 0 (built with -Wno-error=maybe-uninitialized -Wno-error=array-bounds); presubmit --quick 0 FAIL, 3 WARN (siblings: players / mercs / pets as targets; docs-sync; build skipped); gcc -Wshadow -Wunused -Wsign-compare on strike.cpp, gamedata_load.cpp, test_game.cpp: 0 warnings on added lines; clang-tidy (LLVM 23, upstream .clang-tidy) on the same TUs: nothing on the touched files (only pre-existing outdoor.hpp structured-binding names, which CI's older clang-tidy does not flag). Commits: one claim (d9d64d53) plus tooling (6270d162); ponytails in fight.hpp and combat.hpp present and accurate.

Limits (what the claim does not cover):

1. The author's own after side (`test_game --static-field`) re-states the call site's arithmetic; the real `Fight::spot` is covered by this verifier's run only. The PR body should cite a reviewer command that drives `spot` (or say the mode is a restatement). Update 2026-11-01 (card spot155, pushed 8c56e96a to #155: the reviewer mode now drives Fight::spot): the reviewer mode (now `test_static_field --static-field`) drives Fight::spot (see Full path below); the limit is closed.
2. Real 1.14d data gives calc1 25 and calc2 0 at every level, so the life - 1 cap and the calc2 floor are reached only with synthetic Skills.txt values (checked, 0 mismatches).
3. Not modelled on either side of the claim: a monster's magic damage reduction (FUN_0057bf80 subtracts table row 2's flat reduction, MDR, for lightning) and lightning absorb (FUN_0057bf10); FUN_0057c060's damage percent (100 for a player hitting a monster; other attacker kinds not run); players, mercs and pets as targets (the aurafilter walk, line of sight).
4. Hooked in the game side: stat getters, max life, the difficulty row, events and the get-hit event; the life step of FUN_0057c6c0 (0x57c852..0x57c874) is applied in Python.
5. Det state hashes: no scripted scenario casts Static Field; not run.

## Full path (card spot155, 2026-11-01)

The reviewer mode `test_static_field --static-field SEED N` casts Static Field through the real `Fight::spot` in a World: the caster (level 1..20, the case's lightning pierce in its stat sums, an expansion game), the skill row's calc1 / calc2 and DifficultyLevels StaticFieldMin set from the case, two monsters in range (the case's, and the previous case's max life, life and resistance: the per-unit loop). Columns after the case: the first monster's life after the cast, the second's, then the inner-function result (static_field_fine + resisted_fine, or `skip`); `tools/emu/static_field.py` prints the same three from game.exe's FUN_005c96a0 / FUN_0057c6c0 under the emulator. Not covered, as before: the area walk (line of sight, players as targets, the range test itself: monsters sit well inside aurarange), magic damage reduction, absorb.

Result on the worker host (real d2d builds, same cases): seeds 1 and -7, 309,664 cases each (209,664 boundary plus 100,000 random), 619,328 in all. After (PR head with the fixup, 8c56e96a): 0 differ on all three columns (no RESFAIL: the monster's resistance read back equals the case's). Before (upstream main 72132b99 built with the same mode, spot and second columns only, no floor): 261,335 lines differ (first monster 115,103 and 115,353, the same counts as the line-for-line copy of the old arithmetic). Also re-run on the final tree: 212,664 cases (seed 1, 3000 random) 0 differ; the cold-length mode 24,312 cases 0 differ; test_game, test_rules, test_combat, test_protocol rc 0; pr_guard 0 FAIL / 0 WARN; presubmit --quick 0 FAIL; clang-tidy (upstream config) clean on the new file; gcc -Wshadow -Wunused -Wsign-compare clean; ruff clean.

posted: 2026-11-01 fixup 8c56e96a pushed to #155 (old head ddc8281d, fast-forward) and one PR comment (see docs/research/pr-static-evidence-pack.md); CI all pass.
