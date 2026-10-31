# Evidence pack: trace/belt (for the user's go / no-go)

> status: current (agent DECOMPBELT, 2026-10-31). One page. Detail: docs/research/pr-belt-proof.md; body: docs/research/pr-belt-draft.md; trace: docs/research/belt-pickup-trace.md; standard: docs/process/upstream-pr-standard.md. posted: nothing.

**Claim.** On a ground pickup, an `isc` / `tsc` with no tome room goes into a belt column whose bottom is the same scroll (game.exe FUN_00628ba0's isc/tsc test via FUN_0063c560 never refuses; FUN_0063c600 picks the box). d2d sent them to the inventory.

**Proof summary.** game.exe's own FUN_0062bad0 / FUN_00628ba0 / FUN_0063c600 run in upstream's emulator on the loaded 1.14d tables with NO hooks (`tools/emu/belt_pickup.py`, in the branch) against d2d's real `rules::pick_up` (`test_game --belt-pickup`), same cases, joined by `diff`. Every link V (read and run), the data D.

**Numbers.**

| Scenario | Cases | Mismatches before | after |
|---|---|---|---|
| Grid: 8 belts x bottom row {-, hp1, mp1, rvs, isc}^4 x {bottom only, full} x 29 picks | 290,000 | in the total | 0 |
| Random stacked belts (any kinds, some rows above the belt), seed 1 | 100,000 | in the total | 0 |
| Total | 390,000 | 3,111 (isc 2,859, tsc 252; other codes 0) | 0 (md5 equal) |

Game outcomes, for the record (belted / pack of each pick's ~13,450 cases): hp1-5 and mp1-5 about 10,000 / 3,400; rvs, rvl about 9,750 / 3,700; isc 2,859 / 10,639; tsc, 0sc, vps, yps, wms, elx, mpo, rps, bpl about 250 / 13,200 (same-code column only); key, rin, jew, amu, aqv, gcv 0 / all.

Their checks: `ninja test_game test_rules test_combat` (Release), `test_rules` ok (new isc case), `test_combat` rc 0, `test_game` rc 0 (27 OK), `pr_guard.py upstream/main` 0 FAIL 0 WARN, `presubmit.py --quick` 0 FAIL 2 WARN (ruff absent; build by hand). Two commits authored by the user, bodies one line `Agent: Claude Code`, no trailer.

**Commands (reviewer re-runs on their copy).**

```sh
python3 tools/pr_guard.py && cmake --build build && ctest --test-dir build --output-on-failure
cd tools/emu && uv run python belt_pickup.py gen 100000 1 > cases.tsv && uv run python belt_pickup.py game cases.tsv > g.txt
../../build/tests/test_game --belt-pickup cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -     # empty
```

**What the bot gains.** Nothing measurable: the bot never puts scrolls in the belt. The value is fidelity (the belt step is now proven equal to game.exe for every Beltable code) and the oracle, which pins the potion rules our pickit must copy (docs/research/belt-pickup-trace.md, findings B1-B5).

**Residual risks.**
- No tome / inventory items in the cases (those steps unchanged).
- The client gaps (inventory-open pickup to the cursor, shift-click to the belt, ctrl-click sell) are not in this PR.
- Branch on upstream/main f8aab0e8; fetch again right before a push.

**Verifier.** Independent Opus agent (read-only): **VERDICT: holds, with limits.** Fresh seed 320,000 cases 0 diffs; its adversarial set of 60,000 (15 belt codes, gaps, rows above the belt, 55 % scrolls in boxes, all 27 Beltable codes) 0 diffs; 63c560's four exits and the caller tests re-read; every function on the path reads only units' +0/+4. Limits: 63afd0 placement, stacking and auto-equip not run; the before count not reproduced by it. Full report in the proof doc.

**PROOF: complete.** Recommendation: go, with the user's go per step (push `trace/belt` renamed `fix/belt-scroll`, open the PR with the draft body). Small claim: may also be bundled with another PR as its own commit, per the maintainer's preference.
