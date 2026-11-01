# PR body (combined): Frozen Armor and the freeze resume (BUNDLEARMOR)

> status: READY, held for the user's go; NOT posted, NOT pushed, not opened (card bundlearmor, 2026-11-01). One upstream PR with two claims, the user's decision of 2026-11-01 (armor claim first, freeze-resume claim second; one PR, two commits). Det state hashes now run: 14 of 14 scenarios identical on the resume claim and on the combination, see [pr-bundle-armor-resume-det.md](pr-bundle-armor-resume-det.md). Everything below the rule is the text to post. Final branch `fix/frozen-armor-and-resume-2c` (TWO commits, f9e38b15 armor claim plus tooling, 9a4e726a freeze-resume claim plus tooling; BUNDLEDET squash, tree identical to the four-commit branch) in the worker host's PR checkout `bundlear`, built from upstream main 72132b99. The four-commit branch `fix/frozen-armor-and-resume` (6b5f0267, a2a16a87, 0ab7b5e7, a4ad2099) is kept intact; commit hashes below that name those four are its commits, and the per-commit statements apply to the squashed pair. Built from the two verified candidates: `fix/armor-hit` (claim 36cbe015, tooling d5bd19c1) and `fix/freeze-resume` (claim c6cc4f12, tooling efacc7f4). Both source branches and their worktrees are untouched. Sources: [pr-armorhit-body-final.md](pr-armorhit-body-final.md), [pr-armorhit-addendum.md](pr-armorhit-addendum.md), [pr-armorhit-verifier.md](pr-armorhit-verifier.md), [pr-freezeresume-proof.md](pr-freezeresume-proof.md), [pr-freezeresume-verifier.md](pr-freezeresume-verifier.md), [pr-freezeresume-draft.md](pr-freezeresume-draft.md) (revised version on branch agent/forkfreeze until merged). Process: [../process/upstream-pr-standard.md](../process/upstream-pr-standard.md).

**Title:** `Frozen Armor freezes by the freeze rule; a frozen monster thinks again at the freeze end + aidel`

---

## Summary

Two claims about what happens to a monster that game.exe freezes, two commits, each a claim plus its tooling (the oracle and the test mode):

| commit | what it claims |
|---|---|
| `f9e38b15` game: Frozen Armor freezes a striker by the freeze rule | `Fight::buff_events` resists the Frozen Armor length by cold resistance less pierce; specials are chilled for it; plain monsters are frozen only at coldeffect < 0, for length / MonsterFreezeDivisor, and not chilled. Adds `GameData::freeze_divisor`. Also adds `armor_hit.py` and `test_armor_hit --armor-hit` (tooling, no d2d behaviour). |
| `9a4e726a` game: a frozen monster thinks again at the freeze end + aidel | a monster whose freeze (state 1) ends thinks again at the end + MonStats aidel, not at the end. Adds `Monster::stunned_until` to tell a stun from a freeze. Also adds `freeze_resume.py` and `test_freeze_resume --freeze-resume` (tooling, no d2d behaviour). |

The claims are independent (each differential was proven on its own branch on upstream main) and they compose: Frozen Armor's new freeze sets `stun_until` without `stunned_until`, so the resume claim reads it as a freeze and applies the aidel delay. That composition is by reading (see Limits), not by a joint differential. Merging either alone is fine; the order here is the order the user chose.

Which skills this touches on d2d: Frost Nova and Shiver Armor chill; Glacial Spike and Frozen Armor freeze. The armor claim changes Frozen Armor (and no other freeze source); the resume claim changes the thaw of every freeze, which in practice means Glacial Spike and Frozen Armor.

## Limits (read first)

Claim 1, Frozen Armor:
- Only the player's Frozen Armor (`Fight::buff_events`). The Act 3 hireling's Frozen Armor (`merc_buff_events`) still stuns and chills every striker; game.exe runs the same event function for it, so the rule is the same by reading, but neither differential covers it.
- Monster strikers only; game.exe chills a player striker, and d2d has no player-on-player path here.
- Not modelled: state 0x36 (no freeze), the chill's 20 percent state 0x6b roll, a zero-length freeze (length / divisor = 0: game.exe still creates state 1 for 0 frames, d2d does nothing; see the residuals below).
- The trigger stays d2d's (`damaged_by`: any damage; game.exe tests the physical field after reduction); the armors' stacking (game.exe's state group 1) is unchanged.
- The game side hooks the calc (Frozen Armor calc1 restated from Skills.txt); d2d's own calc1 agrees on every case.

Claim 2, freeze resume (residuals, none introduced by the claim):
- **Prior stun** (state 21) running at the freeze: game.exe refiles at the freeze end + aidel even under a longer stun; d2d keeps one stand end equal to `stunned_until` and refiles nothing (`// ponytail:`). A stun landing during a freeze also assigns the stand end, so it can shorten the freeze. Independent differential: d2d acts 5 to 11 frames later in 397 of 1,881 (seed 1) and 422 of 1,978 (seed 2) prior-stun cases.
- **Zero-length freeze**: game.exe freezes for 0 frames and thinks at now + 1 + aidel. d2d never makes one (upstream has no divisor; the armor claim above refuses a zero length: `length / divisor > 0`), so nothing changes today. If d2d ever made one, the resume claim would refile at now + aidel, one frame early.
- **d2d's own tick order**: `Fight::world` runs `monster_update` before the player's missiles strike, the reverse of game.exe's per-frame order (missiles, players, monsters). A think due on the frame a freeze lands still runs in d2d, and a refreeze on the resume frame comes after the think. This is already on upstream main, independent of this PR, and out of scope (the verifier's order `n` run shows it: 253 of 253 think-on-freeze-frame cases and 38 of 38 refreeze-on-think-frame cases).
- **Specials** (champions, uniques, superuniques, boss rows) are never frozen in game.exe (they are chilled); d2d still freezes them through other paths until a freeze-rules change lands (a separate local candidate, not in this PR). With this PR the Frozen Armor path no longer freezes them, but Glacial Spike still does, and with the resume claim they also wait aidel. They are the "differ" rows in the tables below and are not claimed.
- Not modelled (`// ponytail:`): the other events game.exe drops while a unit is frozen (types 0, 1, 6, 7, 9, 10, never refiled).
- A get-hit that ends during a freeze still thinks once in d2d (game.exe drops it); the end + aidel think is right either way, since the change replaces `next_act`.

Both, together:
- **No joint differential.** Each harness ran on the combined head (below) and gives the same answer as on its own branch, but no harness drives Frozen Armor and then steps the monster's AI. The composition claim (an armor freeze thaws at end + aidel) rests on the code: the armor path writes `stun_until` only, the resume branch fires when `stun_until != stunned_until`.
- **The armor verifier's resume column** (it measures the stun end, not the AI turn) still reports 52,714 of 134,400 on this head, exactly as on the armor branch alone. That column predates the resume claim and is answered by the resume harness, not by that one.
- **Det state hashes** (BUNDLEDET, [pr-bundle-armor-resume-det.md](pr-bundle-armor-resume-det.md)): the 14 DETHASH2 scenarios, a fresh d2d per run, on the lab base, the lab base plus the resume claim, and the lab base plus both claims: **14 of 14 identical** (monster hash and state hash) in both; the base reruns agree with themselves. The armor claim alone ran 14 of 14 identical earlier. Coverage limit: no scenario wears Frozen Armor, and the hit scenes (Glacial Spike included) show a stun deadline but no changed hash, so the hashes show only that nothing outside the claims moved (no RNG draw added); the freeze and the thaw are covered by the differentials, not by these hashes.
- Not run: an MSVC build.

## What

game.exe's Frozen Armor event (FUN_005cac50, damagedinmelee) gives a striker whose blow did damage a hit carrying only calc1 as its freeze length through FUN_0057c6c0. FUN_0057c1e0 resists that length by the striker's cold resistance less the wearer's cold pierce (no pierce through immunity), and FUN_0057b230 decides: a MonStats boss, champion, unique or superunique is chilled for it (FUN_0057af80: nothing at coldeffect 0, divided by MonsterColdDivisor when negative); anything else is frozen only at a negative coldeffect, for length / MonsterFreezeDivisor, and is not chilled. d2d stood and chilled every damaging striker for the raw calc1 frames: Andariel and every champion stopped for 1.2 s a hit, cold immunes too, and Nightmare / Hell froze 2x / 4x too long.

When a frozen monster's freeze ends, game.exe's state-1 callback (0x57b170, registered only by FUN_0057b230) removes the monster's pending AI events and files the next think at the end + MonStats aidel (+0x4f, plus the difficulty in a single-player game; 0 reads as 15). While frozen, FUN_005a7f80 drops every think that comes due (types 0, 1, 2, 6, 7, 9, 10, 11, 13 and 14 are gated on state 1; 11, 13 and 14 have no handler). The end + 1 event FUN_0057b230 files never runs as a think: the callback removes it, or it comes due while still frozen. d2d let the monster think at the freeze end. A stun's end (state 21, callback 0x56e900) and a chill's (state 11, callback 0x57ad80) refile nothing.

## Change

- `Fight::buff_events`: the rule above (resisted length, specials chilled with `land`'s chill rule, plain monsters frozen at coldeffect < 0, not chilled). `GameData::freeze_divisor` from DifficultyLevels (MonsterFreezeDivisor). `docs/research/re/skills.md`: the event 2 line completed.
- `monster_update`: at the first tick past a stand that was a freeze, `next_act` = its end + aidel (0: 15). `Monster::stunned_until`, set at the one stun site in `Fight::land` (`strike.cpp`); `docs/research/re/monster-ai.md`: the freeze rule.
- Tooling (needs the CODEOWNERS review for `tools/`): `tests/game/test_armor_hit.cpp`, `tools/emu/armor_hit.py`, `tools/emu/freeze_rules.py` (shared helper); `tests/game/test_freeze_resume.cpp`, `tools/emu/freeze_resume.py`. Both test modes skip without the 1.14d patch.
- Callers of the changed functions: `Fight::land` (one assignment more, same stun), `Fight::buff_events` (the claimed rule), `monster_update` (one branch before the stand check). Players, mercs and pets are untouched (only monsters are frozen by these paths).

## Evidence

Differentials against game.exe's own code run natively under upstream's emulator, d2d side through the real functions (`Fight::buff_events`; `Fight::strike` plus `monster_update`). Counts are on the COMBINED head (a4ad2099 on the four-commit branch; 9a4e726a on the two-commit branch has the identical tree), Release build with `-Wshadow -Wunused -Wsign-compare`; "base" is the count each claim's own proof recorded on upstream main 72132b99 plus the tooling files.

**Claim 1, Frozen Armor** (differing cases, chill and freeze end frames):

| harness | cases | upstream main | combined head |
|---|---|---|---|
| author's `armor_hit.py`, seed 1 (0x41524d48, 50,000 + edge) | 108,506 | 95,920 | **0** |
| author's, seed 2 | 108,506 | 95,986 | **0** |
| independent verifier's own generator, oracle and harness (`verify_armor_hit.py`, 4 seeds, length and end frames; calc1 and RESFAIL also 0) | 134,400 | 109,573 | **0** |

Every kind (plain, champion, unique, superunique, minion, boss column) and difficulty agrees after the change. Isolation (verifier): tooling = base, claim = head on all four seeds.

**Claim 2, freeze resume** (first think frame; plain and minion monsters, the ones game.exe freezes):

| harness | cases | upstream main | combined head |
|---|---|---|---|
| author's `freeze_resume.py`, seed 1 (0xF2E50001, 30,000 + grid; every MonStats aidel value, 3 difficulties) | 12,548 frozen of 37,776 | 12,503 | **0** |
| author's, seed 2 | 12,516 frozen of 37,776 | 12,459 | **0** |
| independent verifier's own oracle on game.exe's real event queue (FUN_005416b0 files, FUN_005414d0 steps a frame, FUN_00541060 walks the monster bucket; no hooks on the queue), seed 1 | 17,180 claimed frozen | 17,130 | **0** |
| same, seed 2 | 17,119 claimed frozen | 17,070 | **0** |

The verifier's claimed classes are single freeze, prior freeze, refreezes inside / on the end frame / in the delay window / on the think frame / after, and think due on the freeze frame, in game.exe's frame order (its order `g`; d2d's own tick order is the residual above). No off-by-one: no delta of one frame anywhere in the claimed classes. On the combined head the residual classes read exactly as on the resume branch alone: prior stun 397 and 422, zero-length 308 and 320, specials 11,912 of 12,111 (seed 1; they differ before as well, 10,659).

Reproduce (worker host, build the test targets, `D2_MPQ_DIR` the 1.14d data, `D2_PATCH_INSTALLER` the 1.14d patch installer):

```sh
cd tools/emu
uv run python armor_hit.py gen 50000 0x41524d48 > cases.tsv
D2_MPQ_DIR=<dir with bin/game.exe> uv run python armor_hit.py game cases.tsv > game.tsv
build/tests/test_armor_hit --armor-hit cases.tsv > d2d.tsv
uv run python armor_hit.py compare cases.tsv game.tsv d2d.tsv          # mismatches 0
uv run python freeze_resume.py gen 30000 0xF2E50001 > c1.tsv
D2_MPQ_DIR=<dir with bin/game.exe> uv run python freeze_resume.py game c1.tsv > g1.tsv
build/tests/test_freeze_resume --freeze-resume c1.tsv > d1.tsv
uv run python freeze_resume.py compare c1.tsv g1.tsv d1.tsv            # "frozen in game.exe ... 0 differ"
```

## How to review

Commit by commit; each commit stands alone on upstream main (the tooling inside each adds no d2d behaviour).

1. `f9e38b15`: the Frozen Armor rule in `states.cpp` (about 30 lines), the divisor member and its load line, one doc line; the oracle (`armor_hit.py`, `test_armor_hit`) is in the same commit: it runs FUN_005cac50 and everything under it natively and compares against `Fight::buff_events`.
2. `9a4e726a`: one member, one assignment in `Fight::land`, one branch in `monster_update`, one doc paragraph; the oracle (`freeze_resume.py`, `test_freeze_resume`) is in the same commit: it runs FUN_0057b230, FUN_005a7f80, the expiry walk and the callback natively and drives a Zombie (every think a swing) through the real `Fight::strike` and `monster_update`.
3. Check the two residual lists above against the code comments (`// ponytail:` in `states.cpp` and `ai.cpp`).

## Checks

On the combined head a4ad2099, whose tree is identical to 9a4e726a (worker host): `tools/pr_guard.py` 0 FAIL / 0 WARN on the range and on each of the four commits, and again 0 / 0 on the range and on both commits of the two-commit branch; each of the two commits builds on its own (armor test 0 mismatches of 108,506 on commit 1 and on commit 2, resume test 0 differ of 12,548 frozen on commit 2); `tools/presubmit.py --quick` 0 FAIL / 3 WARN; test_monsters, test_rules, test_combat, test_protocol and test_game pass; gcc `-Wshadow -Wunused -Wsign-compare` Release build 0 warnings on added lines; clang-tidy (names and includes, the PR-selected translation units) 0 findings in the touched files; ruff clean with presubmit's rules, plus `--select B905` clean. Commits: author and committer the numeric-ID noreply identity, body at most three lines, `Agent: Claude Code`, no other trailer.

Answers to the presubmit WARNs:
- siblings (unit kinds): only monsters are frozen by these paths. Players are chilled, and hireling classes are chilled, by FUN_0057b230's other branches; the Frozen Armor wearer's hireling path (`merc_buff_events`) is not changed (see Limits).
- siblings (hit paths): the claim is the melee-strike reaction of Frozen Armor and the thaw of any freeze. Missile and skill hits that freeze (Glacial Spike) reach the resume claim through the same state; their freeze length is not changed here.
- build: skipped by `--quick`; the build, the four test binaries and the clang-tidy sweep were run separately (above).

Merges: the two candidates merge cleanly in either order (the authors' report, textually). Each also merges cleanly with #155 to #158 (the armor claim shares `gamedata.hpp` and `gamedata_load.cpp` with #155 and #158). A freeze-rules change that is local only reads MonsterFreezeDivisor too: whichever lands second drops its copy of the member and the load line.

## What this does not fix

- Frozen specials that still come from other paths (Glacial Spike freezing a champion, unique, superunique or boss), and the freeze's own length rules there.
- The Act 3 hireling's Frozen Armor.
- d2d's tick order against game.exe's (monsters before missiles), the prior-stun refile, zero-length freezes, the chill's 20 percent roll and state 0x36.
- Player-targeted chill from monsters.

Agent: Claude Code
