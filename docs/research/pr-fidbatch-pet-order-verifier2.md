# Second verifier report: fidbatch v2, "a monster tries a player's pets newest first, the merc by when it was hired or revived"

> status: verifier report 2026-11-24 (independent Opus verifier, the same one that returned NOT READY on v1: [pr-fidbatch-pet-order-verifier.md](pr-fidbatch-pet-order-verifier.md); docs only, nothing pushed or posted). Subject: [pr-fidbatch-pet-order-draft.md](pr-fidbatch-pet-order-draft.md) v2; fork branches `didle/fidbatch-pet-search-order-up-v2` @ `02368c52` (one commit on upstream main `ff6d47a66`) and `didle/fidbatch-20261124b` @ `42cfeaad` (on `0fcd512a5`). Tags: V = run or read by this verifier (how), S = read in a doc, U = unverified.

## Verdict

**READY AFTER EDITS.** Both v1 blockers are fixed and the claim holds on everything I could throw at it: the author's oracle at three revive shares and my own timeline oracle (hire, summons, merc death with game.exe's own removal FUN_005b1a90, revive, summon deaths, in any order) give 0 differences after the change; the parent fails 4 of the 8 in-game checks; ctest is green; the det hashes are unchanged. No blocker. The edits are wording and evidence: the body must not say the in-game test runs on the parent as committed (it does not compile there), must say the test sets the serials by hand, must carry the two U items below (two merc-maker callers not named; tidy findings outside the diff that CI's PR mode will reach), and the README row needs its number.

## 1. Re-derivation (V unless said)

- FUN_005b1900 inserts after the list head and FUN_005dd7f0 walks from the head taking only a strictly nearer one (`jae` at 0x5dd960): as in the v1 report, re-read, unchanged.
- FUN_005b1a90 (read in full this time, 0x5b1a90..0x5b1b6c): fastcall (game, unit); for a player list (index < 8) it walks from the head and unlinks the unit's node (`prev.next = next`, `next.prev = prev`), frees it to the pool, and sets the unit's list index (+0xd0) back to 0xb. So removal keeps the others' order and a re-inserted unit goes to the front. Callers (xref): FUN_005a6520 (death, 0x5a6586), FUN_005431f0 (a room's units stored), FUN_00555600, Confuse FUN_005c3d50, FUN_005c39c0, Conversion's undo FUN_005d01a0, FUN_005d7310.
- The merc enters the list through FUN_00573270 (its 5th argument nonzero skips the insert) from FUN_005770e0, FUN_005774f0 and FUN_00579180 (the free Rogue), and through FUN_00579aa0 (only caller: the revive handler FUN_00579c00). FUN_005774f0 passes its own last argument as that skip flag; its callers are FUN_00533c70 (twice; the join/load chain), FUN_005394a0 (flag 0) and FUN_0056aa50 (flag = bit 0x10000 of a stored record). **U:** FUN_005394a0 (called from 0x530237, 0x53a3eb, 0x53a4e6) and FUN_0056aa50 (from 0x56b313) are not named. If either one runs on a level or act change, game.exe would put the merc in front again where d2d does not re-stamp it. Nothing I read points that way. It is still the one open link in the chain.
- d2d's stamp sites (V read at `02368c52`): `pets.push_back` happens only in `Fight::summon_one` (summons.cpp:525), stamped on the line before it; `World::spawn_merc` (setup.cpp:496) stamps the merc and is called on a new game (setup.cpp:421), a join (356), a hire (commands.cpp:241), a resurrect (commands.cpp:258), the act 3 free merc (act3.cpp:142) and Kashya's reward (shared.cpp:381). It is not called on a level change. A dead summon leaves `actor->pets` (pets.cpp erase_if); a dead merc stays as a dead foe entry and can never win.
- `search_order` (monsters.cpp): with exactly one player entry, the entries after it are split stably into pets first and the rest after, then the pets are sorted newest first by serial. With any other number of player entries d2d's order is returned unchanged, so co-op does not change (the v1 medium finding is fixed). Converted monsters are `foe.monster`, so they are `other` and keep their place after the pets, as on main.

## 2. Runs (V: builder long lane, own trees `/data/src/v2f-*` and build dirs `/data/build/v2f-*`, no shared binary replaced)

| Check | Result |
|---|---|
| `test_pet_search_order` on `02368c52` | `OK`, rc 0 |
| The same test on the parent `ff6d47a66` | **does not compile**: it sets `list_serial`, `merc_listed` and `Pet::listed`, which the parent lacks (8 errors) |
| The same test on the parent with its 4 serial-stamping lines removed (the parent's order ignores them) | 4 of 8 checks FAIL (took 2 / 2 / 1 / 1, game.exe's order 3 / 4 / 2 / 3), rc 1. The revive tie and the nearer summon after a revive pass on the parent too: with the merc revived last, the old "merc first" order agrees with game.exe |
| Author's `pets_merc.py` + harness, seeds 21..24 x 25,000, REVIVE_SHARE 0.0 / 0.3 / 1.0 | before 1,212 / 1,006 / 527, after 0 / 0 / 0. Case and game md5 prefixes `250eb5c1 d24291ff` / `36a3277d b9566f9d` / `95f579f4 782e4567`. Matches the author's numbers |
| Verifier's timeline oracle (a copy of `pets_merc.py` with one change: per case, a random order of summon, hire, merc death plus revive, merc death with no revive, and a summon's death; game.exe side `FUN_005b1900` per insert and the native `FUN_005b1a90` per death, its pool free `0x40b480` hooked to a no-op; d2d side d2d's own rule: one counter bumped per summon and per hire or revive, a dead merc as a dead entry, a dead summon dropped), seeds 41..44 x 25,000 | before 664, **after 0**. Every event pattern is 0 after: h, s, hs, hr, hrs, hks, hkrs, dh, dhs, dhks, ks (29,712 revives, 8,531 merc deaths, 18,048 summon deaths, 29,284 hires after a summon). The first run reported 1,633 after: a bug in my script wrote a merc that died and stayed dead as alive. I fixed it, all 1,633 were in that pattern, and the rerun is the line above |
| `-Wshadow -Wconversion -Werror -fsyntax-only` (each TU's compile command) | ai.cpp, monsters.cpp, pets.cpp, summons.cpp, tick.cpp, setup.cpp, test_pet_search_order.cpp: rc 0 |
| clang-tidy-19 as CI's PR mode (`tidy-pr-select.py` over the 10 changed C++ files: 286 TUs because `fight.hpp` and `ai/units.hpp` changed) | No finding in a changed file. Findings in 5 TUs outside the diff: `std::erase_if` include-cleaner in `apps/d2d/devctl/debug_hero.cpp:30,37`, `tests/game/test_stand_attack.cpp:132` and `tests/game/test_two_hands_blow.cpp:149`, **the same on upstream main** (rerun on the base tree). The other two (`apps/d2d/main.cpp:56`, `video.hpp:22` libavcodec) come from this builder having no ffmpeg build |
| Whole tree at `02368c52` (GCC Release, `D2D_WERROR=OFF`, `-j4`), `ctest -j4` | 506 tests, 0 failed (13 skipped for data this builder lacks, the same 13 as in v1) |
| Det state hashes, 14 scenarios, lab d2d `0fcd512a5` (sha256 `8c4fb56e554e`) against `42cfeaad` (`e22ec08d6c57`), `--lab-verbs --no-save` | All 14 identical (`mon` and `state`). The merc, pet and follow reruns on the base are identical too. As in v1 no scenario holds a tie, so this shows nothing else moved; it does not exercise the claim |
| `pr_guard.py upstream/main HEAD` | 0 FAIL, 0 WARN |
| `presubmit.py --quick` | 2 FAIL (the `test_structs.py` row that already fails on main, V in v1), 5 WARN: siblings (co-op partner; **new: net roles, client not host**), doc-claims README:294, build skipped, no compile_commands |

## 3. Findings (no blocker)

| # | Rank | Finding | Recommendation |
|---|---|---|---|
| 1 | high | The draft and the body say the in-game test "fails on the parent". As committed, the test cannot compile on the parent. The parent run needs the 4 serial lines removed (section 2: 4 of 8 fail) | Say so in the body with the exact edit, or guard the serial lines so the same file builds on both |
| 2 | high | The in-game test sets `pet.listed` and `merc_listed` by hand. The stamp sites (`summon_one`, `spawn_merc` via hire / resurrect) are proven only by reading | Add one check that makes the summons through `summon_one` (or a cast) and revives through `cmd::ResurrectMerc`, or state the gap in Limits |
| 3 | medium | CI's PR tidy job will reach the 4 `std::erase_if` findings (pre-existing on main, outside the diff) because the PR touches `fight.hpp` and `ai/units.hpp`. If CI's clang-tidy reports them as clang-tidy-19 does, the PR's tidy check fails for reasons that are not its own | Say so in the body, or move `list_serial`/`listed` to a header that fewer TUs include. Run CI's exact tidy command once before posting |
| 4 | medium | U: the merc-maker callers FUN_005394a0 and FUN_0056aa50 are not named (section 1). If one runs on an act or level change, game.exe re-inserts the merc there | Name them (a short read) or list them under Limits as "merc re-insert on other paths: not looked for" |
| 5 | low | presubmit WARN "net roles: client, not host": answer it in the body (the list order is the server's pick; one player only) | One line |
| 6 | low | README row: "`./pm_h before` differs in about a thousand S cases" | Give the numbers per REVIVE_SHARE (1,212 / 1,006 / 527 at 0 / 0.3 / 1.0) |
| 7 | low | Lab fork only: the lab verbs `merc level` / `merc heal` / `merc hire` call `spawn_merc` and so make the merc the newest. game.exe does not do that for a level or a heal | Note for lab drills; not upstream |
| 8 | low | The draft's limit "the removal on death is read, not run" is now run: my timeline oracle calls FUN_005b1a90 natively | Replace that limit with the timeline result |

## 4. Claim wording

The title and claim are now accurate (V): `fight: a monster tries a player's pets newest first, the merc by when it was hired or revived (FUN_005dd7f0)`. The title is 109 characters; the last 30 upstream subjects are at most 84. Suggested shorter form: `fight: a monster tries a player's newest pet first, the merc by its last hire or revive`. Limits to keep in the body: one player (several players keep d2d's order); a converted monster keeps its place after the pets (game.exe: the caster's newest pet); findings 2 and 4 above; the hooked functions; `test_structs.py` already fails on main.

## Addendum 2026-11-24: final confirmation of v5 (short, against the edit list above)

Subject: `didle/fidbatch-pet-search-order-up-v5` @ `87d9d7428` (one commit on upstream main `dadda9b98`), body `docs/research/pr-fidbatch-pet-order-body.md` @ `0b9a29736`. Builder long lane, one job, own trees `/data/src/v5f-*`. V unless said.

**Verdict: READY AFTER EDITS.** The code, test and oracle are ready as committed. Only the PR body needs edits, and they are short (below).

| Edit-list item | v5 |
|---|---|
| Test must build on the parent and fail there for the right reason | It builds on `dadda9b98` + the test file only. 3 of 9 checks fail (took 2 / 2 / 1, game.exe's order 3 / 4 / 2: two summon-vs-summon ties and the merc-vs-summon tie). The 4 revive / hire-later checks pass on the parent too, because the old "merc first" order is right when the merc is the newest. On `87d9d7428`: OK |
| No hand-set serials | Done: summons come through `Fight::summon_one`, the merc through `World::spawn_merc` (a revive = `merc.reset()` then `spawn_merc`, which is what `cmd::ResurrectMerc` calls) |
| Oracle on the final tree | `pets_merc.py` at REVIVE_SHARE 0 / 0.3 / 1.0: 1,212 / 1,006 / 527 differ before, 0 / 0 / 0 after (md5 as before). My timeline oracle (native FUN_005b1a90): 664 before, 0 after |
| ctest on the final tree | 506 tests: 505 passed, 1 failed, 13 skipped (the same 13, game data missing). The failure is `test_two_hands_blow` (Subprocess aborted: `test_two_hands_blow.cpp:58` `phys_hi == 100`), and it fails the same way on upstream main `dadda9b98` (built and run on the parent tree) |
| `-Werror -Wshadow -Wconversion` on the 7 changed .cpp | rc 0 on all 7 |
| clang-tidy-19 on the 7 changed .cpp | rc 0, no finding |
| `pr_guard.py upstream/main HEAD` | 0 FAIL, 0 WARN |
| `presubmit.py --quick` against `dadda9b98` | **0 FAIL**, 4 WARN (siblings: co-op partner; net roles; build skipped; no compile_commands). The `test_structs.py` failure is gone on this upstream main |
| Det hashes, 14 scenes: lab `0fcd512a5` (sha `8c4fb56e554e`) against `didle/fidbatch-20261124c` @ `8195a62f` (`fcb597451a80`, the same components change as v5, checked line by line) | all 14 identical. No scene holds a tie |
| Wording and limits | Title (86 characters) and claim match the evidence. FUN_005394a0 / FUN_0056aa50 are named, and the act-change path is tagged U. One player only, converted monster, lab-only restamp: all stated |
| Author and disclosure | Author and committer `shatteredeagle <26069301+...@users.noreply.github.com>` (numeric noreply), no Claude identity or trailer in the commit; commit body is 2 lines. `Agent: Claude Code` is in the body, as upstream AGENTS.md section 6 asks |

Body edits before posting:
1. Remove the stale `test_structs.py` text: the checklist line "fails on main too: `tools/emu/test_structs.py`" and the result block's "presubmit: 2 FAIL". On `dadda9b98` presubmit has 0 FAIL.
2. Result block: replace the pre-amend ctest caveat with this run (final tree, 505 / 1 / 13). Drop "Not done: independent verification of this exact commit; det hashes for v5": both are done above.
3. Upstream AGENTS.md section 4 step 5: a test that also fails on main is "named there, and fixed in its own commit or reported in an issue". `test_two_hands_blow` is named, but the body must also say it will be reported in an issue (filing it needs the user's go) or fixed in its own commit.
4. The template says to delete the sections that don't apply: remove "Claims (a batch only)" for this one-claim PR.

Labels `trace` and `port`: the body correctly leaves them for the user's go.
