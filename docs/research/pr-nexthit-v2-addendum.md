# NextHit v2 addendum: justhit ends one frame later (NEXTHITFIX, card nexthitfix, 2026-11-01)

> status: current. Supersedes the end-frame rule and the counts of [pr-nexthit-proof.md](pr-nexthit-proof.md) and the draft [pr-nexthit-draft.md](pr-nexthit-draft.md) (both keep their text; read them with this addendum). Follows the verifier [pr-nexthit-verifier.md](pr-nexthit-verifier.md), whose one real finding this fixes. Local only: nothing pushed to the public repo or posted upstream; the private fork branch was pushed to the private remote.

Tags: **V run** run by us on the worker host (how given); **V read** disassembly (the verifier's trace, restated); **S** source; **U** untraced.

## 1. The rule (V read, verifier section 1; V run in both oracles)

FUN_005414d0 runs one game frame's timers group by group: group 2 (missiles: FUN_005adcc0 -> FUN_005adbb0 -> the row's do function, so every missile's move and collision) before group 0 (players) and group 1 (monsters: FUN_005a7f80 -> the type-12 handler -> FUN_00627460 -> the stat-list callback 0x5adaf0, which clears state 0x56). A unit struck in frame f therefore keeps 0x56 through the missile pass of frame f + NextDelay and is struck again in frame f + NextDelay + 1 at the earliest. For the novas (NextDelay 4): blocked f+1..f+4, next hit f+5. The old patch (`now_ms < just_hit_until`) and the author's oracle (timer run at the start of its frame) allowed f+4; the author's "0 differences" on the end frame were built into the oracle, not measured.

## 2. The v2 candidate (V run, worker host)

- Branch `fix/nexthit-v2` in worktree `pr/nexthit2`, base upstream main 72132b99 (still upstream main on 2026-11-01), two commits, head **6b3964fa**: claim **b9e7a3d5** and tooling **6b3964fa**. The old `fix/nexthit` (claim 8abab0ed, tooling 7901fda2) is untouched.
- Claim commit (one claim, body 3 lines, numeric-ID noreply identity, `Agent: Claude Code`, no other trailer): `<` becomes `<=` in the justhit end test of `components/game/fight/missiles.cpp`; the `// ponytail:` note now states the traced rule and chain; `docs/research/re/skills.md:928` is corrected (verifier edit d).
- Tooling commit: `tools/emu/next_hit.py` expires justhit AFTER the frame's contacts (the timer list is walked at the end of `frame`), docstring corrected; the rest as 7901fda2.
- The builds below were made at 2f1f835a, the same two commits before one comment-only change (the chain names FUN_005a7f80, not FUN_005a7ef0, because pr_guard flags FUN_005a7ef0 as not a function entry in game-functions.tsv); code and oracle logic are identical to the head.

## 3. Differentials on the real build (V run)

Author's harness (`test_next_hit --next-hit 1 100000`, 100,198 cases, replayed with the corrected oracle; each build on its own contacts):

| build | oracle | cases whose per-monster hits differ | H lines differ | hits per touched monster d2d / game.exe |
|---|---|---|---|---|
| upstream main 72132b99 + mode (proof) | start-of-frame timer | 84,779 | 248,513 | 4.2 to 4.4 / 1.03 |
| fix/nexthit head 7901fda2 | start-of-frame timer (old oracle) | 0 (by construction) | 0 | 1.03 / 1.03 |
| fix/nexthit head 7901fda2 | corrected oracle | **8,473** | 8,725 | 1.04 / 1.00 |
| **fix/nexthit-v2** | corrected oracle | **0** | **0** | **1.00 / 1.00** |

Verifier's harness (`tools/emu/test_verify_nexthit.cpp` and `verify_nexthit.py` / `verify_nexthit_compare.py` copied unchanged from the merged verifier card into a separate build dir; seeds 0x5101, 0x6202, 0x7303, 0x8404, 25,000 each plus 624 edge cases = 100,624; moving monsters and second casts included; contacts replayed in monster-index order, expiry after the missiles, the traced order). The v1 row is the verifier's own run (its saved outputs, not re-run); the v2 row is this run:

| slice | v1 head: differ / cases | v1 d2d / game.exe per touched | **v2: differ / cases** | v2 d2d / game.exe |
|---|---|---|---|---|
| all | 25,252 / 100,624 | 1.30 / 1.21 | **0 / 100,624** | **1.21 / 1.21** |
| Nova | 12,069 / 50,077 | 1.29 / 1.20 | 0 / 50,077 | 1.20 / 1.20 |
| Frost Nova | 13,183 / 50,547 | 1.30 / 1.22 | 0 / 50,547 | 1.22 / 1.22 |
| standing | 6,065 / 50,302 | 1.16 / 1.11 | 0 / 50,302 | 1.11 / 1.11 |
| moving | 19,187 / 50,322 | 1.51 / 1.36 | 0 / 50,322 | 1.36 / 1.36 |
| one cast | 16,105 / 80,548 | 1.18 / 1.11 | 0 / 80,548 | 1.11 / 1.11 |
| two casts | 9,147 / 20,076 | 1.77 / 1.59 | 0 / 20,076 | 1.59 / 1.59 |

So the moving-monster and two-cast cases that differed under v1 do not differ any more: the whole residual was the end frame. Two controls: v2 against the pre-order oracle (expiry before the missiles) differs in 25,252 cases (the mirror image of v1 against the traced order), and v2 with contacts replayed in crossed-subtile order (P) instead of monster-index order (I) differs in **6** of 100,624 cases (4 standing, 2 moving; 1 Nova, 5 Frost Nova): the unit order inside one frame, the one limit the verifier named (unit order inside a subtile is U).

Fork build (section 6): the same author's harness gives 0 of 100,198 differ (T 0, H 0), and the d2d and game.exe lines equal the upstream v2 run's line for line (md5 f7a4d40cdc99 of the oracle output, 006311f9f474 of the C / F / H / T lines).

## 4. Det state hashes (V run, one hidden d2d slot)

Builds: base = the DETHASH2 lab tree (upstream 72132b99 + lab commits, 5e0a51bc; binary sha256 70c7b3694173, the one det-hashes-2026-11-01b.md used); v2 = that tree + cherry-pick of the claim b9e7a3d5 alone (binary sha256 13779361ed3d; the tooling commit adds no d2d code). One slot (`d2dslot.sh run -n 1` under `agentrun.sh nexthitfix`), capacity checked first (idle + nice 44 percent, above the 25 percent floor; `tools/capacity.sh` hangs, so read from /proc/stat over 10 s), the DETHASH2 driver unchanged: 14 scenarios, seed 3, lockstep, a fresh d2d per run, 28 runs.

The base hashes reproduce DETHASH2's and the verifier's base column in all 14 scenarios. Result: **9 of 14 identical** (passive, fight, cold, static, cold2, hit_minion, merc, pet, follow: no NextHit missile, or the Frost Nova reaches no minion), **5 differ** (hit_plain, hit_champion, hit_unique, hit_boss, hit_immune), **0 unexplained**. The 5 are exactly the v1 differences of the verifier's section 5, with the same hashes (hit_plain mon a5f5122cbaeb, hit_champion 3ac6324ac7e2, hit_unique 452a5ef2e450, hit_boss a2f25b89107c, hit_immune 6606f2c92ec3): v1 and v2 give the same det state on every scenario, because a single point-blank cast on one standing monster is hit once under either end frame.

| scenario | change base -> v2 | class |
|---|---|---|
| hit_plain | Frost Nova target #5 hp 106/133 -> 131/133, last chill 14000 -> 13960; then #6 spawns with another max life (119 -> 93) | EXPECTED: one hit not several; the RNG shift reaches the next spawn |
| hit_champion | #5 303/312 -> 310/312, chill unchanged; #6's damage roll shifted | EXPECTED, same cause |
| hit_unique | #5 494/520 -> 518/520, chill 14000 -> 13960; #6 is another unique | EXPECTED, same cause |
| hit_boss | #5 109/113 -> 112/113, chill 10160 -> 10120; #6's damage shifted | EXPECTED, same cause |
| hit_immune | only #10: the cold-immune #9 loses nothing either way; #10's spawn roll shifts (max life 4495 -> 5918) | EXPECTED (RNG), chill and stun deadlines unchanged |

DET-CHANGES-EXPLAINED. Coverage limit (unchanged): no scenario casts a Nova or a Frost Nova at a moving or multi-target pack, so the end-frame change itself is exercised only by the differentials of section 3.

## 5. CI-class checks on the v2 head (V run, worker host)

- pr_guard: 0 FAIL / 0 WARN on the range against upstream/main and on each commit (b9e7a3d5, 6b3964fa).
- presubmit --quick: rc 0, 0 FAIL, 4 WARN, answered for the PR body: (a) unit kinds: game.exe puts 0x56 on any unit, but d2d keeps `just_hit_until` on `Monster` only; players, mercs and pets as targets get no justhit, and monster novas (Mephisto's, Baal's, Andariel-class poison novas) still multi-hit the player; (b) hit paths: only missiles carry NextHit, no melee or skill path is touched; (c) doc-claims: re/skills.md:928 now states the traced rule (the claim is the rule itself, checked against FUN_005adb00 and FUN_005a8730); (d) build skipped by `--quick`, the build and tests are listed here. Ruff with upstream's rules plus B905: clean (ruff 0.17.3).
- Upstream tests (Release, MPQ dir set): test_rules, test_combat, test_protocol, test_game, test_monsters all rc 0.
- gcc `-Wshadow -Wunused -Wsign-compare`: 0 warnings on added lines (test_game, test_rules, test_combat, test_protocol, test_monsters, test_next_hit built).
- clang-tidy (repo config, tidy-pr-select: 136 TUs that include a touched file): 0 findings in the touched files; 29 elsewhere (identifier-length on structured bindings, a Qt moc include, a compiler flag clang does not know), in files this branch does not touch, as for the verifier.
- Replay on current upstream main: upstream main is still 72132b99 (`git fetch upstream`, 2026-11-01); the branch is cut from it.
- Not done: the verifier's optional ruff defaults fixes (I001, SIM115 in next_hit.py: not upstream's rules; edit e below stays open).

## 6. The private fork (V run)

- Branch `lab-nexthit` of the private fork got one new commit on top (a fast-forward, no rewrite; the old tip 989517079642 is its parent): **2c6c4a1a651ce45a3b70ececc7c03208b6492ba7**, pushed to the private remote on 2026-11-01 (verified with `git ls-remote`). It applies the v1 -> v2 difference of the upstream candidate (the `<=`, the ponytail note, the corrected oracle, the skills.md line).
- New binary `bin/d2d.lab-nexthit2` on the worker host (Release, `D2D_WERROR=OFF`, sha256 prefix 4e27d6c915e3; built in `build-nexthit2`, tree `wt-nexthit`). The old `bin/d2d.lab-nexthit` (sha256 prefix 7a1bea454bc9) stays in place so running soaks keep their binary.
- **Rule: soaks and drills use `bin/d2d.lab-nexthit2` from now on**, and say so in their report (binary, branch, hash). Frost Nova and Nova numbers measured on `d2d.lab-nexthit` are slightly high where a ring touches a monster that moves or is under a second cast (the verifier's v1 residual: 1.30 vs 1.21 hits per touched monster on its mixed set); on one standing monster the two builds are equal, see the readback.
- Differential on the fork build: 0 of 100,198 differ (section 3).
- Live readback (one hidden slot, `--no-save` throwaway Sorceress, `d2drill -probe` through the cheat ledger, the same probe as docs/forknext.md; Blood Moor, one Fallen at 100000 life with no damage and all resistances 0, adjacent for casts 3 to 6): life points lost per cast.

| Cast | `d2d.lab-nexthit` (docs/forknext.md) | `d2d.lab-nexthit2` |
|---|---|---|
| Frost Nova 1 (about 2.7 cells away) | 66 | 66 |
| Frost Nova 2 | 61 | 61 |
| Frost Nova 3 (adjacent) | 67 | **67** |
| Nova 1 (adjacent) | 174 | **174** |
| Nova 2 | 151 | 151 |
| Nova 3 | 136 | 136 |

Equal in all six: one standing, adjacent monster is struck once by the whole ring in both builds, so the end frame cannot show up here (the readback confirms nothing broke; the end-frame difference is in the differentials). Nothing of the readback is left running (slot 3 released, `agentreap.sh --list` empty).

## 7. The five draft edits the verifier asked for (text for the merger to fold into the PR body)

(a) Change, end-frame rule: "A NextHit hit puts state 0x56 on the struck unit; every NextHit missile passes it by until the missile pass of frame + NextDelay has run (FUN_005414d0 runs the missile timers, group 2, before the monsters', group 1; expiry FUN_005a7f80 -> FUN_00627460 -> 0x5adaf0): struck at f, struck again at f + NextDelay + 1 at the earliest (novas: blocked f+1..f+4, next hit f+5). `Monster::just_hit_until` is compared with `<=`."

(b) Limits: drop "the frame justhit ends ... is not traced". Add: moving monsters and second casts are covered (the verifier's 100,624 cases); only Nova and Frost Nova are compared, the other 46 NextHit rows (Chain Lightning, Multiple Shot / Strafe, the war cries, Fissure, Twister / Tornado, Volcano, Shock Field, Blade Sentinel, Dragon Breath, Claws of Thunder) go through the same `skill_missile_hits` and get the same rule, not compared (Multiple Shot and Strafe become one arrow of a volley per monster per 5 frames, the game.exe rule); the order of units inside one frame (d2d monster index, game.exe subtiles crossed) is not compared and changes 6 of 100,624 cases; players, mercs and pets as targets keep no justhit and monster novas still multi-hit the player; Chain Lightning's hop search and piercing Lightning are not claimed.

(c) Evidence table: replace the counts by section 3: author's harness 84,779 -> 0 of 100,198 (before) and 0 after with the corrected oracle (the old head differs in 8,473 under it); verifier's harness 0 of 100,624 (the old head 25,252); both oracles now run the expiry after the frame's contacts; the verifier's oracle runs the stat-list expiry chain for real.

(d) Correct `docs/research/re/skills.md:928` ("strike a unit again NextDelay frames on", wrong under both readings): done in the claim commit b9e7a3d5.

(e) Optional: `ruff check --fix` for I001 / SIM115 in `next_hit.py` (ruff 0.17 defaults, not upstream's rules): not applied.

## 8. Verdict

**READY** (for the user's go to open it upstream): both harnesses show 0 differences on the v2 build (author's 0 of 100,198, verifier's 0 of 100,624, moving monsters and second casts included), the det hashes are explained (9 of 14 identical, 5 differ only by the Frost Nova hit count and the RNG shift after it, 0 unexplained), the verifier's edits (a) to (d) are applied (the text above and the code commit; (e) optional, open), CI-class checks clean, upstream main unchanged. Nothing was pushed upstream or posted. Remaining before opening: the user's go; the merger folds sections 7 into `pr-nexthit-draft.md`'s body and cites the head 6b3964fa (not the old 7901fda2); the open PRs #155 and #156 share files (the verifier's textual merges were clean for the old head; v2 adds only a comparison character, a comment, one doc line and the oracle, so re-checking the merge is cheap but was not repeated).

## Shared-file lines for the merger

- docs/BACKLOG.md: `- NEXTHIT v2 (NEXTHITFIX, docs/research/pr-nexthit-v2-addendum.md): fix/nexthit-v2 (head 6b3964fa) READY: justhit end frame f+NextDelay+1, author's harness 0 of 100,198, verifier's 0 of 100,624, det hashes explained; fork lab-nexthit 2c6c4a1a651c, soaks use bin/d2d.lab-nexthit2; open: the user's go, edit (e)`
- docs/INDEX.md, research: `- docs/research/pr-nexthit-v2-addendum.md: NextHit v2: end-frame fix (f+NextDelay+1), re-run of both harnesses, det hashes, fork lab-nexthit2, the verifier's five edits`
- docs/forknext.md (not edited here, it is on main): change "Binary" to `bin/d2d.lab-nexthit2`, pushed tip `2c6c4a1a651ce45a3b70ececc7c03208b6492ba7`; add: "soaks and drills use `bin/d2d.lab-nexthit2`; `bin/d2d.lab-nexthit` is retired once running soaks finish; the differential line reads 0 of 100,198 with the corrected oracle".
- CHANGELOG.md: none.
- GLOSSARY.md (guard rule: not edited): `justhit` (state 0x56) still missing; proposed text: "justhit: game.exe unit state 0x56, put on a unit struck by a NextHit missile until the missile pass of frame + NextDelay has run; while on, every NextHit missile passes the unit by."
