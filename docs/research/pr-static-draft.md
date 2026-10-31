# PR body draft: game: Static Field hits as FUN_005c96a0 does

> status: draft, not posted (card decompstatic). Needs the verifier and the user's go.

## What

Static Field's damage on each unit follows game.exe's FUN_005c96a0: the target's life in whole points (stat 6 >> 8), nothing at or under DifficultyLevels StaticFieldMin % of max life in expansion games (0 / 33 / 50), calc1 % of it (MulDiv) leaving at least one point, at least calc2, a negative lightning resistance cancelled (MulDiv(dmg, 100, 100 - res)), then the lightning record in 256ths with the attacker's pierce as FUN_0057c6c0 applies it.

## Evidence

| Address | What it shows |
|---|---|
| FUN_005c9800 | params: calc2 (+0x13c), calc1 (+0x138), game +0x70 ? StaticFieldMin (+0x40) : 0 |
| FUN_005c96a0 | life >> 8, floor, MulDiv, life - 1, << 8, calc2, negative resistance step |
| FUN_0057c6c0 | record through the resistance with pierce, life in 256ths |

```sh
cd tools/emu && uv run python static_field.py 100000 1 > g.txt
../../build/tests/test_game --static-field 1 100000 | grep -E '^[BR][0-9]' | diff g.txt -   # empty
```

309,664 cases; before 115,103 differ, after 0. `tools/emu/static_field.py` needs CODEOWNERS review.

## Limits

The area walk is unchanged (no players, no line of sight test); pierce against immunity as before.

Agent: Claude Code
