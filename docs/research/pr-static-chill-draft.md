# PR #155 title and body delta: Static Field and chill length

> status: draft, not posted (card chillpatch, 2026-10-31). Applies only if the user says go for appending fix/static-field-chill to #155 before its first CodeRabbit review. Proof: [pr-static-chill-proof.md](pr-static-chill-proof.md); the Static Field part of the body stays as in [pr-static-draft.md](pr-static-draft.md).

## New title

game: Static Field and chill length as game.exe does (differential proof vs game.exe)

## Body: section to add after the Static Field section

### Chill on a cold missile

A cold skill missile chills the monster whatever damage it deals: `missile_blow` no longer needs a whole resisted point of cold before it sets `chill_ticks`. game.exe's FUN_0057c6c0 passes the record's resisted cold length (+0x30) to FUN_0057af80 right after the life step, with no test on the cold damage. The length, resistance, pierce and `Fight::land`'s chill rules were already right and are unchanged.

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

Limits: only skill missiles (`player_blow`'s weapon cold keeps the whole-point gate, ponytail); stats 0x99 / 0x76 are not read for a monster, since no 1.14d MonProp, MonUMod or state row gives them to one (players already get both through `chill_length`); record flags that skip positive resistance are not exercised; game.exe's 20 % state 0x6b roll is still not modelled.

Presubmit WARNs: players are chilled through `chill_length`; mercs' and pets' cold missiles use the same `missile_blow` and get the change; no doc describes the rule.

## Body: first line to replace

From "Static Field's damage on each unit follows ..." to: "Two claims, one commit each, each with its own tooling commit: Static Field's damage on each unit (FUN_005c96a0), and a cold missile's chill whatever its damage (FUN_0057c6c0 -> FUN_0057af80)."

Agent: Claude Code
