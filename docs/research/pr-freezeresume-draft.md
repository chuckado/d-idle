# Draft PR body: a frozen monster thinks again at the freeze end + aidel (FREEZERESUME)

> status: READY, held for the user's go (card forkfreeze, 2026-11-01). Local candidate `fix/freeze-resume` (claim c6cc4f12, tooling efacc7f4, base upstream main 72132b99). The independent verifier's verdict was READY WITH CHANGES (body text only); the changes are applied below. Still owed before posting: det state hashes (needs a d2d slot; not run). Nothing is posted without the user's go. Evidence: [pr-freezeresume-proof.md](pr-freezeresume-proof.md), verifier: [pr-freezeresume-verifier.md](pr-freezeresume-verifier.md). The same claim runs in the private fork as `lab-freezeresume` ([forkfreeze.md](../forkfreeze.md)).

Title: `A frozen monster thinks again at the freeze end + aidel`

---

## Limits (read first)

- Only the think after a **freeze** (state 1) moves. Champions, uniques, superuniques and MonStats boss rows are never frozen in game.exe (FUN_0057b230 chills them; freeze-rules C3); d2d still freezes them until the freeze-rule fix lands, and with this change they also wait aidel. Players are chilled, not frozen, by the same function; mercs (hireling classes) are chilled too. On d2d, Frost Nova and Shiver Armor chill; Glacial Spike and Frozen Armor freeze, so these are the skills the change touches.
- Residual: a **prior stun** (state 21) running at the freeze. game.exe refiles at the freeze end + aidel even under a longer stun; d2d keeps one stand end equal to `stunned_until` and refiles nothing (`// ponytail:`). A stun landing during a freeze also assigns the stand end, so it can shorten the freeze. The independent differential: d2d acts 5 to 11 frames later than game.exe in 397 of 1,881 cases (seed 1) and 422 of 1,978 (seed 2).
- Residual: a **zero-length freeze** (length / MonsterFreezeDivisor = 0). game.exe freezes for 0 frames and thinks at now + 1 + aidel. d2d never makes one (upstream has no divisor, so no zero length), so nothing changes there; but if d2d ever made one, this change would refile at now + aidel, one frame early.
- Residual: **d2d's own tick order**. `Fight::world` runs `monster_update` before the player's missiles strike, the reverse of game.exe's FUN_005414d0 (missiles, players, monsters). A think due on the frame a freeze lands still runs in d2d, and a refreeze on the resume frame comes after the think. This predates the change and is independent of it; the author's harness (strike, then update) cannot see it. Out of scope here.
- Not modelled (`// ponytail:`): the other events game.exe drops while a unit is frozen (types 0, 1, 6, 7, 9, 10: never refiled).
- A get-hit that ends during a freeze still thinks once in d2d (game.exe drops it); the end + aidel think is right either way, since the change replaces `next_act`.
- Not run: det state hashes, an MSVC build. The independent verifier ran game.exe's real event queue (FUN_005416b0 files, FUN_005414d0 steps a frame, FUN_00541060 walks the monster bucket) under the emulator and found no off-by-one; the author's oracle hooked the queue.

## What

When a frozen monster's freeze ends, game.exe's state-1 callback (0x57b170, registered only by FUN_0057b230) removes the monster's pending AI events and files the next think at the end + MonStats aidel (+0x4f, plus the difficulty in a single-player game; 0 reads as 15). While frozen, FUN_005a7f80 drops every think that comes due (types 0, 1, 2, 6, 7, 9, 10, 11, 13 and 14 are gated on state 1; 11, 13 and 14 have no handler). The end + 1 event FUN_0057b230 files never runs as a think: the callback removes it, or it comes due while still frozen. d2d let the monster think at the freeze end.

This adds `Monster::stunned_until` (the stun's end, set at the one stun site) and, at the first tick past a stand that was a freeze, sets `next_act` to its end + aidel. A stun's end (state 21, callback 0x56e900) refiles nothing, and neither does a chill's (state 11, callback 0x57ad80).

## Evidence

`tools/emu/freeze_resume.py` runs game.exe's FUN_0057b230, FUN_005a7f80, the expiry walk (0x5a7ef0, FUN_00627460) and the callback natively; `test_freeze_resume --freeze-resume` freezes a Zombie (every think is a swing with its foe in melee) through the real `Fight::strike` and runs the real `monster_update` a tick at a time. 37,776 cases on each of two seeds (every MonStats aidel value, all difficulties, priors, second freezes inside the freeze, at its end, in the delay window and after):

| | plain + minion (frozen in game.exe) differ |
|---|---|
| upstream main | 12,503 of 12,548; 12,459 of 12,516 |
| this branch | 0; 0 |

The independent verifier (own oracle on the real event queue, own d2d probe, 31,484 cases per seed) reports the claimed frozen classes (single, prior freeze, refreezes inside / on the end frame / in the delay window / on the think frame / after, think due on the freeze frame): upstream main differs in 34,200 of 34,299 plain and minion cases on two seeds, this branch in 0.

Specials differ before and after (game.exe does not freeze them; the freeze-rule fix's job). Merged with that fix (local), the specials' first think matches too (0 of 6,544 Normal cases).

## Checks

pr_guard 0 FAIL / 0 WARN per commit; presubmit --quick 0 FAIL; test_monsters, test_rules, test_combat, test_protocol, test_game pass; gcc -Wshadow -Wunused -Wsign-compare and clang-tidy clean on the touched lines. Merges cleanly with #155, #156, #157 and #158 in either order.

Answers to presubmit: siblings (unit kinds): only monsters are frozen by this path (players and hirelings are chilled by FUN_0057b230's other branches); data-gating: the test mode skips without the 1.14d patch.

Before posting (verifier note 4, LOW): the claim's code comment cites `FUN_0056e900` for the stun callback; it is a callback address, so the comment should read `callback 0x56e900` (same pr_guard point as the other callbacks).

Agent: Claude Code
