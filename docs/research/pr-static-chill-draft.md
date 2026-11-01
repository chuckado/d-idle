# PR #155 title and body delta: Static Field and cold-missile chill

> status: draft, not posted (card chillpatch, 2026-10-31). Applies only if the user says go for appending fix/static-field-chill to #155 before its first CodeRabbit review. Proof: [pr-static-chill-proof.md](pr-static-chill-proof.md); the Static Field part of the body stays as in [pr-static-draft.md](pr-static-draft.md).

## New title

game: Static Field and cold-missile chill as game.exe does (differential proof vs game.exe)

## Body: section to add after the Static Field section

### Chill on a cold missile

A cold skill missile chills the monster: `missile_blow` no longer needs a whole resisted point of cold before it sets `chill_ticks`. game.exe's FUN_0057c6c0 passes the record's resisted cold length (+0x30) to FUN_0057af80 right after the life step, with no test on the cold damage. The commit changes whether a cold skill missile chills, not the chill length. The "whatever damage it deals" reading holds only while the element's max roll is above 0: when the max roll is <= 0, `missile_blow` returns before the cold, while game.exe chills on the length alone (only the Missiles.txt row rogue3 has a cold length with no damage). The length, resistance, pierce and `Fight::land`'s chill rules were already right for lengths under 0x100000 (above that the integer math differs, int overflow vs MulDiv, unreachable in play) and are unchanged.

| Address | What it shows |
|---|---|
| FUN_0057c6c0 (0x57c8a0) | `push [rec+0x30]` and the call to FUN_0057af80; nothing tests +0x24 first |
| FUN_0057c1e0, table 0x732980 row 5 | the length goes through cold resistance less cold pierce (FUN_0057be00 / FUN_0057bf80), no flat reduction, no absorb |
| FUN_0057af80 | length <= 0 or coldeffect 0: nothing; negative coldeffect: / MonsterColdDivisor; at least 1; only lengthened |

```sh
cd tools/emu && uv run python chill_length.py 100000 1 > g.txt
../../build/tests/test_game --cold-length 1 100000 | grep -E '^[BR][0-9]' | diff g.txt -   # empty
```

121,312 cases (21,312 boundary, 100,000 random); before 14,239 differ, after 0. `tools/emu/chill_length.py` needs CODEOWNERS review.

Verifier limits (independent run, 201,757 cases on real d2d builds; the claim: 34,635 differing before, 0 after). Residual differences that are NOT claimed and stay unchanged: C3 4,282; C5 2,836; max roll <= 0 17,912; length >= 0x100000 29.
- Vengeance is among the gated `player_blow` cold paths (skill cold incl. Vengeance, gear cold, converted cold) that still need a whole point; none is claimed.
- Freeze is not claimed. C3: game.exe turns a freeze into a chill on MonStats bosses, monsters with data flag 8 and hirelings (the earlier word 'champions' is unverified: flag 8 is unique in the common layout, and in the verifier's run flags 4 and 0x10 were frozen). C5: d2d's freeze ignores cold resistance and the freeze divisor, adds a chill, and freezes even at coldeffect >= 0.
- Uncovered chill paths: the three `player_blow` cold sites, the freeze path in missiles.cpp:657-662, Frozen Armor, `monster_blow` on the player, deaths.cpp:164 (the suicide-explosion chill length looks wrong).

Other limits: only skill missiles (`player_blow`'s weapon cold keeps the whole-point gate, ponytail); stats 0x99 / 0x76 are not read for a monster, since no 1.14d MonProp, MonUMod or state row gives them to one (players already get both through `chill_length`); record flags that skip positive resistance are not exercised; game.exe's 20 % state 0x6b roll is still not modelled.

Presubmit WARNs: players are chilled through `chill_length`; mercs' and pets' cold missiles use the same `missile_blow` and get the change; no doc describes the rule.

## Body: first line to replace

From "Static Field's damage on each unit follows ..." to: "Two claims, one commit each, each with its own tooling commit: Static Field's damage on each unit (FUN_005c96a0), and a cold skill missile's chill (FUN_0057c6c0 -> FUN_0057af80)."

Agent: Claude Code
