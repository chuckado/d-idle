## What changes

A Teleport cast that is refused costs no mana in game.exe. d2d took the mana on the action frame and kept it. The differential covers two refusal causes: the target is outside the caster's room and its near list, or the level's Teleport column is 0. Other causes are under Limits.

FUN_0056f7f0 (the do wrapper) calls the do through the table at `0x7322b0` and keeps its return at `[ebp-0xc]` (`0x56f953..0x56f96a`). For a skill with no missile (`srvmissile` -1: jump to `0x56fa3d`) a zero return goes to `0x56f8b3` (return 0) before the payment FUN_0056bfe0 at `0x56fa7a`. Do 27 (FUN_005ca360) returns FUN_00554ea0's result, which is 1 only when the unit was put somewhere. So the payment is made only when the unit landed.

`Fight::player_modes` (the SC branch) pays a skill with no start function on the action frame (`settle_cast`) and then runs its do; this change gives the mana back when the do was Teleport and did not set `reassigned` (`last_paid`, the amount `pay()` took). At level 1 a refused hop was 24 points.

Two parts. The oracle commit: `tools/emu/tele_mana.py` (game side), `tools/emu/tele_mana_diff.py`, `tests/game/test_tele_mana.cpp --tele-mana`, one `tools/emu/README.md` row; on the parent it prints what the cast costs. The claim commit: the refund in `strike.cpp`, `tick.cpp`, `fight.hpp`, a finding in `docs/research/re/skills.md`, and `settle_cast()` zeroes `last_paid`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0056f7f0` `0x56f953..0x56f96a` | the do is `[0x7322b0 + srvdofunc * 4]`; its return is kept at `[ebp-0xc]` |
| `0x56fa3d` | no missile: a zero return goes to `0x56f8b3` and returns 0, before the payment |
| `0x56fa7a` | the payment (FUN_0056bfe0) for a skill with no start function, reached only past the check above |
| `FUN_005ca360` | the Teleport do: 0 on a refusal (no room, level Teleport column 0 at `0x5ca39a`, Teleport 2 line test at `0x5ca3b5`), else FUN_00554ea0's result |
| `FUN_00554ea0` | 0 when no room, no free spot or the position set fails; 1 after the move |

## Counted differentials (V run)

Oracle: `tele_mana.py game N SEED` runs FUN_0056f7f0 natively (the mana check and the stat reads hooked as `sk_mana.py` does; FUN_0056bfe0 recorded) with the real do 27 on `sk_tele_spot.py`'s synthetic rooms, for the cases `sk_tele_spot.py gen N SEED --no-units` makes. d2d side: `test_tele_mana --tele-mana CASES` casts Teleport through `Fight::cast_missile` and World ticks, reads the landing from `ev::Reassigned`, and prints the mana taken against a control run of the same case. Game side: payments equal landings in every case.

| Seed, cases | game: refused / landed | oracle commit (parent d2d) | claim | mutant (refund always) |
|---|---|---|---|---|
| 1, 2,000 | 927 / 1,073 | 927 DIFF, exit 1 | all same, exit 0 | 1,073 DIFF, exit 1 |
| 2, 2,000 | 970 / 1,030 | 970 DIFF, exit 1 | all same, exit 0 | 1,030 DIFF, exit 1 |
| 3, 20,000 | 9,113 / 10,887 | 9,113 DIFF, exit 1 | all same, exit 0 | 10,887 DIFF, exit 1 |
| 4, 100,000 | 45,991 / 54,009 | 45,991 DIFF | all same | not run |

Both sides are deterministic per case (no roll in do 27, FUN_00554ea0 or the payment). `tele_mana_diff.py` exits 1 on an empty file, unequal lengths, a malformed line and ids out of step. The self-check of `test_tele_mana` (a landing hop pays 24, a refused one pays 0, the level's Teleport 0 pays 0) aborts on the parent at `refused.paid == 0`; the mutant aborts it at the landing assert. `test_cast_mana` and `test_tele_spot` pass on the oracle commit, the claim and the mutant. ctest on the claim build: `test_outdoor` and `smoke_d2d` fail as on upstream main; `coop_lobby_d2d` failed once under `-j3` and passed 3 of 3 alone (a flake, no Teleport in it).

## Limits

- Only two refusal causes are in the differential: the target outside the caster's room and its near list, and the level's Teleport column 0. Refusals for "no free spot" (FUN_0064e7b0 in FUN_00554ea0) and a Teleport 2 line that is not clear are not: game.exe returns 0 for them too (`0x554f4c`, `0x5ca3b5`, read), and d2d's `spot()` refuses them before `reassigned`, so they are refunded, but no case contains either (Teleport 2's line test is hooked clear in `sk_tele_spot.py`, and no case lacked a free spot). The near list is `sk_tele_spot.py`'s 3 x 3 hook.
- Teleport only. The wrapper's rule (`0x56fa3d`) applies to any skill with no srvmissile and no start function; no other skill was run, and Leap has a start function.
- A refused Teleport cast from charges still uses the charge in d2d (`pay()` sets `last_paid = 0` for a charged cast); whether FUN_0056bfe0 is where game.exe uses the charge is not claimed, and not compared.
- game.exe also skips the delay on a refusal; d2d still calls `start_delay`. Teleport has no delay, so nothing differs.
- Regeneration is read through the mana difference against a control run with a tolerance; the mana check (FUN_0056c160) is hooked to pass: what the cast needs is `sk_mana.py`'s claim.
- No Windows, macOS or Clang build was run.

## Checklist

- [x] Test in `tests/` (`tests/game/test_tele_mana.cpp`)
- [x] Finding in `docs/research/re/` (`skills.md`)
- [x] `tools/pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `tools/presubmit.py --quick --base upstream/main`: 0 FAIL; WARN `siblings`: melee and missile skills pay elsewhere (a start function or a srvmissile); only the no-missile, no-start path is changed, and only for Teleport.

---

Agent: Claude Code
