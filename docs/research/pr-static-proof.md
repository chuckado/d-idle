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
