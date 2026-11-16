# Verifier: decomp35, a player's leech and Crushing Blow (2026-11-16)

> status: current, 2026-11-16 (card verifydecomp35, TASK-564, Opus, singly; independent verifier before posting). Nothing posted. Subjects: private fork branches `fid/leech-r1` (head `99679ead`) and `fid/crush-r1` (head `de243e2b`, stacked), bodies `docs/upstream/bodies/decomp35-leech-body.md` and `decomp35-crush-body.md`, trace `docs/research/decomp35-combat.md` (TASK-556). Checklist from `verifier-p13-2026-11-16.md`, `verifier-seq-followup-2026-11-16.md`, `verifier-p18-2026-11-15.md`. The fixed `fid/leech-r2` and `fid/crush-r2` are on the private fork remote only.

Tags: **V run** (run by this verifier today, NAS builder), **V read** (disassembly or source read by this verifier), **I** (inferred), **U** (unverified).

## Verdicts

| PR | Verdict | Post |
|---|---|---|
| Leech (D1 to D3) | **PASS WITH CHANGES** | `fid/leech-r2`: oracle `26ae7020`, claim `55cb4e68`, on upstream main `c3dc51c3`; body after L1 to L6 |
| Crushing Blow (D4) | **r1 FAIL, r2 PASS WITH CHANGES** | `fid/crush-r2`: oracle `1a75fed3`, claim `e5dc17d3`, on `fid/leech-r2`; post only after the leech PR merges, rebased on upstream main; body after C1 to C6 |

(Build and ctest rows: section "Build, ctest, tidy".)

WIP: the rest of this file is being written.
