# Frozen Armor striker (fix/armor-hit): addendum after the independent verifier (ARMORFINISH, 2026-11-01)

> status: current. Supersedes the limits, merge line and not-run list of [pr-armorhit-draft.md](pr-armorhit-draft.md) and the end + 1 / zero-length notes and "NOT READY" of [pr-armorhit-proof.md](pr-armorhit-proof.md) (both keep their text; read them with this addendum). Source: the verifier record `pr-armorhit-verifier.md` (card verifyarmorhit; on main once merged). Verdict there: READY after draft text edits, no code change. Final body: [pr-armorhit-body-final.md](pr-armorhit-body-final.md). Local only; nothing posted; the PR is NOT open (the user has not said go, six PRs are already open).

Tags: **V read** disassembly (verifier's dumps); **V run** run under upstream's emulator (verifier's oracle); **V** worker-host check by this card; **S** d2d source.

## 1. Edit 1: the AI turn after a freeze (replaces "event 2 at end + 1")

- The freeze-end callback FUN_0057b170 clears state 1 and refiles the monster's AI event (event 2) at end + MonStats aidel (aidel +0x4f, difficulty variants +0x50 / +0x51; 15 when 0 or less): 13 to 15 frames for most rows (V read). FUN_0057b230 files an event 2 at end + 1, but the callback removes it at the end, and FUN_005a7f80 drops type-2 events while state 1 is on (V read), so end + 1 never runs as an AI turn.
- d2d gates a stunned monster with `stun_until` (ai.cpp:1831, S), so it may act from the end frame: aidel frames earlier than game.exe. Numbers: the resume frame differs in 52,714 of the verifier's 134,400 cases on this branch (116,092 on upstream main); on this branch the whole residual is this, and 15 / 13 / 14 account for 22,612 / 10,508 / 10,433.
- It holds for EVERY d2d freeze (Glacial Spike's missile freeze, the freeze-rules candidate, this claim). It is not introduced by this claim and NOT part of this PR. Follow-up card FREEZERESUME (agent/freezeresume); the verifier's sketch: `next_act = max(next_act, stun_until + aidel)` for a freeze.

## 2. Edit 2: the zero-length freeze

length / MonsterFreezeDivisor = 0 still creates state 1 ending now and refiles the AI turn in game.exe (callback at the expiry). d2d does nothing, so the monster's AI is not interrupted there. 205 of the 134,400 cases (V run); these sit inside the resume-frame count above. Stated as a limit, not claimed.

## 3. Edit 3: proof table rows (add to the proof's differential)

| run | build | result |
|---|---|---|
| author's harness, 108,506 cases, 2 seeds | upstream main + test / this branch | 95,920 and 95,986 differ / 0 (unchanged) |
| verifier's own generator, oracle and harness: 134,400 cases, 4 seeds (0xA101 + 14,400 edge, 0xB202, 0xC303, 0xD404 at 30,000 each); chill end, freeze end and AI resume frame | upstream main 72132b99 | **109,573** differ in an end frame; 116,092 in the resume frame |
| same | this branch (head d5bd19c1) | end frames **0**; resume frame 52,714 (the limit of section 1); calc1 differs 0; RESFAIL 0 |
| isolation, md5 of the case lines per seed | base, tooling alone, claim alone, head | tooling = base, claim = head on all four seeds |
| det state hashes (DETHASH2 driver, 14 scenarios, 28 runs, one hidden slot) | base lab tree vs the same + claim 36cbe015 | **14 of 14 identical**, 0 unexplained. Coverage limit: no scenario wears Frozen Armor, so the hashes show only that the claim changes nothing outside its branch (no RNG draw added); the freeze itself is covered by the differentials, not by these hashes |

Verifier's CI-class checks on the head (pr_guard 0 / 0 per commit, presubmit --quick 0 FAIL / 3 WARN answered, test_monsters / test_rules / test_combat / test_protocol / test_game rc 0, gcc and clang-tidy 0 in touched files, ruff clean) agree with the proof's.

## 4. Edit 4: merges

Clean in either order (textual, verifier): #155, #156, #157, #158 (shares gamedata.hpp and gamedata_load.cpp with #155 and #158, re/skills.md with #158), and the local fix/freeze-rules-v2 (head 409903a6; shares gamedata.hpp, gamedata_load.cpp, freeze_rules.py). With freeze-rules-v2 the identical `freeze_divisor` member collapses into one but gamedata_load.cpp gets the `MonsterFreezeDivisor` load line twice (harmless, builds). Order decision: **this PR goes first** (it is READY; freeze-rules-v2 is not posted), and freeze-rules-v2 then drops its `freeze_divisor` member and its load line. If freeze-rules-v2 lands first instead, drop the two gamedata hunks here.

## 5. Edit 5: not run

Drop "det state hashes" (run, section 3). Keep "an MSVC build".

## 6. Edit 6: optional edits, not applied

- `merc_buff_events` (the Act 3 hireling's Frozen Armor): same block through a shared helper in this PR. Not applied: a code change and not in either differential; stated as Limits line 1.
- Comment correction for the `// ponytail:` note in states.cpp: "the AI turn at the freeze's end + aidel (FUN_0057b170 refiles event 2; the end + 1 event of FUN_0057b230 is removed by it) isn't kept". It is a code (comment) change to a commit that was verified as it stands; the trivial fix needs the author's tree, so it is left to the opener (list in the checklist, section 8). The body already states the correct rule.

## 7. Worker-host check (V, 2026-11-01)

The upstream candidate branch `fix/armor-hit` still exists in the PR checkout on the worker host: head **d5bd19c1** (tooling), parent **36cbe015** (claim), then upstream main 72132b99; working tree clean.

## 8. Ready checklist (for the OPENARMORHIT agent, on the user's go only)

Ask the user first (a seventh open PR). Then: confirm upstream main is still 72132b99 (`gh api repos/psi29a/d2decomped/commits/main`; if it moved, replay both commits and re-run everything); apply the ponytail comment correction of section 6 as an amend of the claim only if the user agrees, else leave the comment and keep the body's limit; re-run pr_guard on the range and per commit (0 FAIL / 0 WARN), the upstream tests (test_game, test_rules, test_combat, test_protocol, test_monsters) and presubmit --quick; re-run BOTH harnesses on the head: the author's `armor_hit.py` (gen / game / compare with test_armor_hit: expect 0 of 108,506) and the verifier's (`verify_armor_hit.py` gen / game, `test_verify_armor_hit`, compare: expect end frames 0 of 134,400 and the resume frame 52,714); check the #155 to #158 merges still apply and the freeze-rules-v2 order of section 4; keep the identity rules (numeric-ID noreply author and committer, body plus `Agent: Claude Code`, no Claude trailer; push a new branch with a plain push, no force; the PR footer `Agent: Claude Code`); post nothing else (no CodeRabbit trigger or Discord notice without the coordinator's say) and write the posted log as `pr-armorhit-posted.md` in the style of pr-nexthit-posted.md.

## Shared-file lines for the merger

- docs/BACKLOG.md: `- ARMORHIT: READY (draft edits done, docs/research/pr-armorhit-addendum.md, body docs/research/pr-armorhit-body-final.md); open only on the user's go (seventh PR); follow-up FREEZERESUME for the end + aidel resume`
- docs/INDEX.md, research: `- docs/research/pr-armorhit-addendum.md / pr-armorhit-body-final.md: fix/armor-hit after the verifier: verified limits, proof rows, final PR body, ready checklist (ARMORFINISH)`
- docs/research/pr-bundle-queue.md (owned by card armorhit, not edited): append `- ARMORFINISH 2026-11-01: fix/armor-hit READY, final body pr-armorhit-body-final.md, branch head d5bd19c1 checked; waiting for the user's go`
- CHANGELOG.md: none.
