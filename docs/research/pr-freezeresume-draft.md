# Draft PR body: a frozen monster thinks again at the freeze end + aidel (FREEZERESUME)

> status: DRAFT for the local candidate `fix/freeze-resume` (claim c6cc4f12, tooling efacc7f4, base upstream main 72132b99). NOT READY: needs an independent verifier and the det state hashes; nothing is posted without the user's go. Evidence: [pr-freezeresume-proof.md](pr-freezeresume-proof.md).

Title: `A frozen monster thinks again at the freeze end + aidel`

---

## Limits (read first)

- Only the think after a **freeze** (state 1) moves. Champions, uniques, superuniques and MonStats boss rows are never frozen in game.exe (FUN_0057b230 chills them); d2d still freezes them until the freeze-rule fix lands, and with this change they also wait aidel. Players are chilled, not frozen, by the same function; mercs (hireling classes) are chilled too.
- Not modelled (`// ponytail:`): the other events game.exe drops while a unit is frozen (types 0, 1, 6, 7, 9, 10: never refiled); a freeze that ends while a longer stun (state 21) is on (d2d keeps one stand end; game.exe refiles at the freeze end).
- A zero-length freeze (length / MonsterFreezeDivisor = 0) delays the next think to frame + 1 + aidel in game.exe; d2d never makes one, so nothing changes there.
- A get-hit that ends during a freeze still thinks once in d2d (game.exe drops it); the end + aidel think is right either way, since the change replaces `next_act`.
- Not run: det state hashes, an MSVC build. The game side hooks the event queue (an event filed for frame X runs at X; FUN_005416b0's move of a past frame to now + 1 is read, the per-frame walk is not run).

## What

When a frozen monster's freeze ends, game.exe's state-1 callback (0x57b170, registered only by FUN_0057b230) removes the monster's pending AI events and files the next think at the end + MonStats aidel (+0x4f, plus the difficulty in a single-player game; 0 reads as 15). While frozen, FUN_005a7f80 drops every think that comes due (types 0, 1, 2, 6, 7, 9 and 10 are gated on state 1). The end + 1 event FUN_0057b230 files never runs as a think: the callback removes it, or it comes due while still frozen. d2d let the monster think at the freeze end.

This adds `Monster::stunned_until` (the stun's end, set at the one stun site) and, at the first tick past a stand that was a freeze, sets `next_act` to its end + aidel. A stun's end (state 21, callback 0x56e900) refiles nothing, and neither does a chill's (state 11, callback 0x57ad80).

## Evidence

`tools/emu/freeze_resume.py` runs game.exe's FUN_0057b230, FUN_005a7f80, the expiry walk (0x5a7ef0, FUN_00627460) and the callback natively; `test_freeze_resume --freeze-resume` freezes a Zombie (every think is a swing with its foe in melee) through the real `Fight::strike` and runs the real `monster_update` a tick at a time. 37,776 cases on each of two seeds (every MonStats aidel value, all difficulties, priors, second freezes inside the freeze, at its end, in the delay window and after):

| | plain + minion (frozen in game.exe) differ |
|---|---|
| upstream main | 12,503 of 12,548; 12,459 of 12,516 |
| this branch | 0; 0 |

Specials differ before and after (game.exe does not freeze them; the freeze-rule fix's job). Merged with that fix (local), the specials' first think matches too (0 of 6,544 Normal cases).

## Checks

pr_guard 0 FAIL / 0 WARN per commit; presubmit --quick 0 FAIL; test_monsters, test_rules, test_combat, test_protocol, test_game pass; gcc -Wshadow -Wunused -Wsign-compare and clang-tidy clean on the touched lines. Merges cleanly with #155, #156, #157 and #158 in either order.

Answers to presubmit: siblings (unit kinds): only monsters are frozen by this path (players and hirelings are chilled by FUN_0057b230's other branches); data-gating: the test mode skips without the 1.14d patch.

Agent: Claude Code
