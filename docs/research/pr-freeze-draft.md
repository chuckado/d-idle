# Draft PR body: freeze rules

> status: DRAFT for `fix/freeze-rules-v2` (head 409903a6, cards decompfreeze and freezepolish, 2026-11-01); not posted. Open only after #155 merges and the branch is rebased (`git rebase --onto upstream/main 8c56e96a fix/freeze-rules-v2`, pr-freeze-proof.md "How to open"), the gates are re-run, and the det run exists (the det line below says it was not run until then). Needs the user's go.

---

**Title:** Glacial Spike's freeze follows FUN_0057b230

### What

Glacial Spike's freeze on a monster now follows game.exe's FUN_0057b230:

- The freeze length is cut by cold resistance less cold pierce, the same way the chill length is (row 6 of the element table at 0x732980, next to row 5).
- Glacial Spike's freeze on a MonStats boss, champion, unique, superunique or act boss (MonStats boss column, or monster-data flag 8) is a chill for that length. It is never frozen.
- Any other monster is frozen only when its coldeffect is negative. The freeze lasts length / MonsterFreezeDivisor, and the monster gets no chill from it.
- Blood Raven's and the Countess's freezes are no longer halved. Their `0x76` call is `FUN_00639db0`, the state setter (state 118, corpse_noselect). It is not stat 0x76.

Before this, d2d stood every monster for the raw length at any coldeffect and also chilled it.

### Proof

`tools/emu/freeze_rules.py` runs game.exe's FUN_0057c6c0 natively, including FUN_0057c1e0, FUN_0057af80, FUN_0057b230 and the boss, flag-8 and hireling tests. `test_freeze_rules --freeze-rules` runs the real `Fight::strike` on the same cases. Over 225,748 cases (seed 20261101):

| build | mismatches |
|---|---|
| base (#155 head) | 185,746 |
| + resisted / divisor / coldeffect | 85,040 |
| + special monsters chilled | 32,379 |
| + no halving (head) | 0 |

```sh
cd tools/emu
uv run python freeze_rules.py gen 100000 20261101 > cases.tsv
uv run python freeze_rules.py game cases.tsv > game.tsv
../../build/tests/test_freeze_rules --freeze-rules cases.tsv > d2d.tsv
uv run python freeze_rules.py compare cases.tsv game.tsv d2d.tsv
```

`freeze_rules.py flags` sweeps every monster-data flag byte against the boss column and a hireling class. Only flag 8, the boss column or a hireling turns the freeze into a chill. `freeze_rules.py state` runs the Blood Raven / Countess call natively: it sets state bit 118, writes no stat, and stat 0x76 then reads 0.

Det state hashes: not run for this candidate: needs a d2d slot.

### Limits

- Only Glacial Spike's freeze. Frozen Armor goes through the same FUN_0057b230 rule in game.exe (FUN_005cac50 sends it through FUN_0057c6c0 at 0x5cacfd) but stays unpatched in d2d (states.cpp, player and merc): it still freezes and chills any striker for the raw length. Freezing Arrow, Ice Blast and item freeze are not modelled in d2d; Frozen Orb and Blizzard only chill (no freeze). All are left for separate PRs.
- Which monsters chill (verified on game.exe's own spawn path, 12,830 monsters): flag word 0x0d champions, 0x2d possessed, 0x4d ghostly, 0x09 uniques, 0x0b superuniques and 0x08 act bosses are chilled; minions (0x10) and plain monsters (0x00) are frozen. Only bit 8 matters, and hirelings are always chilled.
- Not modelled: state 0x36 (no freeze while it is set), the freeze's events, and a length / divisor of 0 (game.exe sets state 1 for no frame).
- Residual: a freeze length of exactly 0x100001 with a nonzero resistance differs (game.exe's MulDiv divides first above 0x100000, d2d multiplies); 6 of 146,591 verifier cases, unreachable in play.
- A freeze is still d2d's stand (`stun_until`).
- `half_freeze` keeps its name and now only marks the state-118 call.
- The hireling branch has no d2d caller, because skill missiles don't strike mercs.
- Siblings: a player's freeze is already a chill in d2d (FUN_0057b230's player branch). Pets and a co-op partner as freeze targets are not traced.

Agent: Claude Code
