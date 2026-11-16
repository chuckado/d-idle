## What changes

A Teleport cast that is refused (the target is outside the caster's room and its near list, no free spot within 50 rings, the level's Teleport column 0, or a Teleport 2 line that is not clear) costs no mana in game.exe. d2d took the mana on the action frame and kept it.

FUN_0056f7f0 (the do wrapper) calls the do through the table at `0x7322b0` and keeps its return at `[ebp-0xc]` (`0x56f968..0x56f96a`). For a skill with no missile (`srvmissile` -1: `0x56f971..0x56f974` jumps to `0x56fa3d`) a zero return leaves at `0x56fa41` before the payment FUN_0056bfe0 at `0x56fa70`. Do 27 (FUN_005ca360) returns 0 on every refusal. So the payment is made only when the unit was put somewhere.

`Fight::strike` pays a skill with no start function on the action frame (`settle_cast`) and then runs its do; this change gives the mana back when the do was Teleport and put the player nowhere (`last_paid`, the amount `pay()` took). At level 1 a refused hop was 24 points.

Two parts: the oracle (`tools/emu/tele_mana.py` game side, `tools/emu/tele_mana_diff.py`, `tests/game/test_tele_mana.cpp --tele-mana`; on the parent it prints what the cast costs), then the claim (the refund, in `strike.cpp`, `tick.cpp`, `fight.hpp`).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0056f7f0` `0x56f953..0x56f96a` | the do is `[0x7322b0 + srvdo * 4]`; its return is kept at `[ebp-0xc]` |
| `0x56f971..0x56f984`, `0x56fa3d..0x56fa41` | no missile: a zero return goes to `0x56f8b3` and returns 0 |
| `0x56fa51..0x56fa7a` | the payment (FUN_0056bfe0) for a skill with no start function, reached only past the check above |
| `FUN_005ca360` | the Teleport do: 0 on a refusal, 1 when FUN_00554ea0 put the unit somewhere |

## Counted differentials (V run)

Oracle: `tele_mana.py game N SEED` runs FUN_0056f7f0 natively (the mana check and the stat reads hooked as `sk_mana.py` does; FUN_0056bfe0 recorded) with the REAL do 27 on `sk_tele_spot.py`'s synthetic rooms, for the cases `sk_tele_spot.py gen N SEED --no-units` makes. d2d side: `test_tele_mana --tele-mana CASES` casts Teleport through `Fight::cast_missile` and World ticks and prints the mana taken against a control run. Two seeds of 2,000 cases:

| | cases | refused hops that cost mana (d2d) | differ after |
|---|---|---|---|
| seed 1 parent | 2,000 (927 refused) | **927** | 0 |
| seed 2 parent | 2,000 (970 refused) | **970** | 0 |

`tele_mana_diff.py` exits 1 on a difference, on an empty file and on unequal lengths (all three run). The self-check of `test_tele_mana` (a landing hop pays 24, a refused one pays 0) fails on the parent at its second assert. `test_cast_mana` and `test_tele_spot` pass on both.

## Limits

- Teleport only. Other skills whose do can return 0 with no missile (Leap, Telekinesis without a target, curses) pay on the same rule in game.exe; they were not run here.
- Two of 4,000 cases are hops onto the caster's own subtile: game.exe lands, the test sees no position change (it reads the landing from the position because the World tick consumes `reassigned`). They agree on the mana.
- Regeneration leaves under 400 (256ths) of noise in the d2d column; it is read with a tolerance of 1,024 against a cost of 6,144.
- The mana check (FUN_0056c160) is hooked to pass: what the cast needs is `sk_mana.py`'s claim.

## Checklist

- [x] Test in `tests/` (`tests/game/test_tele_mana.cpp`)
- [ ] Finding in `docs/research/re/` (to add by the verifier pass: one paragraph in `combat.md`)
- [ ] `tools/pr_guard.py upstream/main` (not run in this card)

---

Agent: Claude Code
