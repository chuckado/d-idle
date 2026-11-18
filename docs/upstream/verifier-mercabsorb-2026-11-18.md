# Verifier: merc gear absorb r1 (2026-11-18)

> status: in progress, 2026-11-18 (card verifymerc, Opus, singly). Nothing posted.

## Working notes (replaced by the verdict)

- Upstream main moved from `9b64c11a` to `563f9257`; r1 cherry-picks without conflict (V run). r2 = rebased oracle `e724e726`, claim `a8168234` (r1's claim plus two pins with part of the max life from the merc's armor).
- Not dead code: `World` commands equip the merc (`world/commands.cpp:529`), `merc_gear` sums the worn items' props, `make_fighter` reads 142..149 (V read).
- pr_guard 0/0, presubmit --quick 0 FAIL 5 WARN on r2 (V run). Build, ctest, mutants, tidy running on the builder.
