# Proof: curse and aura lowering on an immune monster (SKRESISTFIX)

> status: evidence for the local candidate `fix/curse-immune` (claim 6a04cb3e, tooling 557d14e3, base upstream main 72132b99) on the worker host. Offline; nothing pushed or posted. Not independently verified. PR body draft: [pr-curseimmune-draft.md](pr-curseimmune-draft.md). Trace: `docs/research/sk-resist-trace.md` (D1, section 7).

Tags: **V run** executed by this agent on the worker host; **S** source reading; **I** inference.

## Candidate

Worktree `skresistfix` of the PR checkout `skresist` (detached at upstream main 72132b99), branch `fix/curse-immune`. Author and committer: the numeric-ID noreply identity of the other candidates; commit bodies end with `Agent: Claude Code`, no Claude trailer.

| commit | what | files |
|---|---|---|
| 6a04cb3e | claim: `target_of` keeps the base resistances, `apply_target_stats` takes a fifth of a negative change when the base is 100 or more (FUN_005c3540) | `components/game/fight/missiles.cpp`, `components/game/fight.hpp`, `docs/research/re/skills.md` (one line) |
| 557d14e3 | tooling: `tests/game/test_sk_resist.cpp` (`--resist CASES`, real `Fight::target_of` and `rules::missile_blow`), `tools/emu/sk_resist.py` (oracle, generator, port, compare) | 2 new files |

5 files, 2 commits. D2 (no -100 floor on magic and physical in d2d) is NOT in the claim: it lives in `rules::missile_blow` / `resisted`, a different code path and not the same fix; with 1.14d data it is unreachable (trace section 4).

PR standard (docs/process/upstream-pr-standard.md): the tests ship with the claim as a tooling commit; claim and tooling are separate, so the claim stands alone.

## The claim

`Fight::target_of` copies `target.res` right after `Monster::target()` (base column plus boss adds, before any curse) and passes it to `apply_target_stats`. A lambda `lowered(type, value)` returns `value / 5` (C++ truncation toward 0) when `value < 0 && base[type] >= 100`, else `value`; the six resistance cases (stats 36 / 37 / 39 / 41 / 43 / 45) use it. The base is the pre-curse value, so a monster with a curse AND the player's aura counts both against the same base (game.exe reads the BASE stat, FUN_006253b0, V read in the trace).

## Differential (V run, the worker host)

Cases from `sk_resist.py gen 100000 SEED` (47,628 grid + 100,000 random = 147,628). Game side: `sk_resist.py game` (FUN_005c3540, FUN_0057be00, FUN_0057c6c0 natively); d2d side: `test_sk_resist --resist` built Release (D2D_WERROR=OFF) on upstream main + the test (before) and on the branch (after).

| run | cases | adj differ | effres differ | dealt differ |
|---|---|---|---|---|
| seed 1, before (upstream main + test) | 147,628 | 88,101 | 92,035 | 14,355 |
| seed 1, after (branch) | 147,628 | **0** | 3,986 | 1,958 |
| seed 7, before | 147,628 | 88,224 | 92,149 | 14,289 |
| seed 7, after | 147,628 | **0** | 3,992 | 1,977 |
| seed 1 and 7, line-for-line port of the fixed rule vs the real after build | 147,628 each | 0 | 0 | 0 |

The before numbers equal the trace's (88,101 / 92,035 / 14,355 and seed 7 88,224 / 92,149 / 14,289). Residual after, by category in the compare output: only "magic total under -100" (dealt 1,958; effres 1,965) and "physical total under -100" (effres 2,021) for seed 1: D2, outside the claim. Nothing else differs; every "base 100 or more and a negative curse" case is gone. md5 (seed 1): cases a1ea534a871f, game fd1425e80544. Isolation: the before tree is a detached worktree at 72132b99 with only the new test file (= the tooling commit's effect on d2d), so claim alone = head.

Commands (the worker host; `<upstream>` = the checkout):

```sh
cd <upstream>/tools/emu
python sk_resist.py gen 100000 1 > c1.tsv
python sk_resist.py game c1.tsv > g1.tsv            # about 80 s; D2_MPQ_DIR as the other oracles
<build>/tests/test_sk_resist --resist c1.tsv > d1.tsv
python sk_resist.py cmp c1.tsv g1.tsv d1.tsv
python sk_resist.py port c1.tsv > p1.tsv && python sk_resist.py cmp c1.tsv p1.tsv d1.tsv
```

`cmp` skips the d2d build's timestamped log lines (it keeps `id adj effres dealt` rows).

## CI-class checks (head 557d14e3)

- pr_guard: 0 FAIL / 0 WARN on the range 72132b99..HEAD and on the claim commit alone.
- presubmit --quick: 0 FAIL; WARN: siblings (hit paths; answer: the change is inside `target_of`, which every monster-resist reader calls), build skipped under --quick (full build below); docs-sync answered by the `skills.md` line.
- Tooling unittests (test_tooling, test_gaps, test_pr_guard, test_presubmit, test_tidy_pr_select): pass (inside presubmit).
- gcc `-Wshadow -Wunused -Wsign-compare` Release build of test_game, test_rules, test_combat, test_protocol, test_monsters, test_sk_resist: no warning in the touched files (4 elsewhere, untouched); the five tests rc 0.
- clang-tidy (names and includes) on `missiles.cpp` and `test_sk_resist.cpp`: 0 findings in the touched files after dropping an unused include and adding `<skills.hpp>`.
- ruff 0.8.6 with presubmit's rules plus B905: clean on `sk_resist.py`.

## Merge check (textual, `git merge-tree --write-tree`, both orders)

Merged against: #152 (8a45cfc9), #153 (5315e66a), #155 (8c56e96a), #156 (3f418788), #157 (cf85634a), #158 (6b3964fa), local `fix/armor-hit` (d5bd19c1), `fix/freeze-resume` (efacc7f4), `fix/frozen-armor-and-resume-2c` (f9e38b15): 0 conflicts in either order for all nine. `apply_target_stats` has no other caller on any of them (the signature change cannot collide semantically). Heads are the refs on the worker host's PR checkouts as of this run; no rebuild of the merged trees.

## Where this and the oracle could share a mistake (for the verifier)

1. The test mode carries each case's value through Lower Resist's Skills row (aurastat1 and its calc replaced for the case, not restored: the process is one-shot), not through the real skill calcs; the value calcs (dm56, ln34) are not covered.
2. `effres` is computed in the test from the real `target_of` result with `missile_blow`'s pierce / floor lines copied (the real function does not return it); `dealt` is the real `missile_blow` for magic, fire, lightning, cold. Physical and poison print `-` for `dealt`.
3. The oracle hooks the base stat getter with the case's base; a monster whose base includes a boss add is just a larger base (both sides).
4. First-aurastat-zero (D1b), hirelings as defenders and the curse's duration / replacement rules are not covered.

## Shared-file lines for the merger

docs/research/pr-bundle-queue.md (owned by another card; add to the candidates table or list):

```
| curse / aura lowering on immune monsters (SKRESIST D1, docs/research/pr-curseimmune-*.md) | **BUILT, NOT READY** (card skresistfix, 2026-11-01): local `fix/curse-immune` on the worker host (claim 6a04cb3e, tooling 557d14e3, on upstream main 72132b99); differential adj 88,101 -> 0 of 147,628 (seed 7: 88,224 -> 0), dealt 14,355 -> 1,958 (seed 7: 14,289 -> 1,977; the rest is D2, magic under -100, not claimed); real `Fight::target_of` and `missile_blow` vs game.exe FUN_005c3540 natively; pr_guard 0/0, presubmit --quick 0 FAIL, five tests rc 0; merges clean both orders with #152 #153 #155 #156 #157 #158, fix/armor-hit, fix/freeze-resume, fix/frozen-armor-and-resume-2c. Needs an independent verifier and the det state hashes before any go (a later card carries it on the private fork after the verifier) | 2 commits, 5 files | nothing pushed |
```

docs/BACKLOG.md (fix queue):

```
- SKRESISTFIX: independent verifier + det state hashes for fix/curse-immune (curse / aura lowering a fifth on a base >= 100: adj 88,101 -> 0 of 147,628); then the private-fork carry card and the user's go for an upstream PR; D2 (-100 floor magic / physical in missile_blow) stays a separate, unreachable-in-data claim (docs/research/pr-curseimmune-proof.md)
```
