# PR draft: NextHit marks the struck unit `justhit` (Nova / Frost Nova)

> status: DRAFT, NOT READY (no independent verifier, no det state hashes); not posted. Proof: [pr-nexthit-proof.md](pr-nexthit-proof.md). Branch `fix/nexthit` on the worker host (claim 8abab0ed, tooling 7901fda2, base upstream main 72132b99).

**Title:** game: a NextHit hit marks the unit justhit, one unit a frame

## What

A missile whose Missiles.txt row has NextHit puts state 0x56 (`justhit`) on the unit it strikes for NextDelay frames (FUN_005adb00); every NextHit missile, of any cast, passes a unit in that state by (the CollideType filter at 0x5a8850 -> FUN_005a8730, and FUN_005adf10); a LastCollide row also passes the unit it struck last (FUN_0064a470); FUN_005ae1f0 hands a missile one unit a frame. d2d kept a per-missile `hit_at` list, so each of a nova's 64 missiles struck a monster it touched, and a missile went on to the next monster in the same frame.

## Change

- `Monster::just_hit_until`, set by a NextHit hit for NextDelay frames and tested by every NextHit missile.
- `Missile::last_hit` (LastCollide, now loaded from Missiles.txt) replaces `hit_at`; Blade Sentinel's turnaround clears it as it cleared `hit_at`.
- The frame loop ends a NextHit missile's frame after its hit.

## Evidence

Differential against game.exe's own code (tools/emu/next_hit.py, tests/game/test_next_hit.cpp): one Nova or Frost Nova through the real `Fight::fire` / `Fight::world` on 1..6 standing monsters; the contacts d2d's flight produced replayed through game.exe's filter, FUN_005adf10 and FUN_005adb00 under the emulator. 100,198 cases (198 boundary, 100,000 random, seed 1):

| | cases whose per-monster hits differ | hits per touched monster per cast (d2d / game.exe) |
|---|---|---|
| before (72132b99) | 84,779 | Nova 4.43 / 1.03, Frost Nova 4.16 / 1.04 (point blank: 64 on a size-3 monster vs 1) |
| after | 0 | 1.03 / 1.03, 1.04 / 1.04 |

```sh
build/tests/test_next_hit --next-hit 1 100000 > d2d.txt
cd tools/emu && uv run python next_hit.py ../../d2d.txt > g.txt
grep -E '^[HT]' ../../d2d.txt | diff g.txt -
```

`tools/emu/next_hit.py` is under `tools/`: needs the CODEOWNERS review.

## Limits

Novas only in the differential (Chain Lightning, Twister / Tornado, Shock Field, the war cries and Blade Sentinel's row get the same rule, not compared); one cast at a time, standing monsters, monsters as targets; contacts come from d2d's flight (game.exe's flight and its unit order within a frame not compared); the frame `justhit` ends (start of frame + NextDelay) is not traced in game.exe; FUN_005ae1f0's one-unit loop restated in the harness. Chain Lightning's hop search still picks a `justhit` unit (FUN_0056b3e0 flag 0x80000): separate claim later. docs/research/re/skills.md:928 ("strike a unit again NextDelay frames on") should be corrected with it.

## Checklist

pr_guard 0 FAIL / 0 WARN per commit; presubmit --quick 0 FAIL (WARNs answered above); test_rules, test_combat, test_protocol, test_game green; gcc shadow sweep and clang-tidy names / includes clean on added lines; no game data.

Agent: Claude Code
