# Verifier: decomp35, a player's leech and Crushing Blow (2026-11-16)

> status: current, 2026-11-16 (card verifydecomp35, TASK-564, Opus, singly; independent verifier before posting). Nothing posted. Subjects: private fork branches `fid/leech-r1` (head `99679ead`) and `fid/crush-r1` (head `de243e2b`, stacked), bodies `docs/upstream/bodies/decomp35-leech-body.md` and `decomp35-crush-body.md`, trace `docs/research/decomp35-combat.md` (TASK-556). Checklist from `verifier-p13-2026-11-16.md`, `verifier-seq-followup-2026-11-16.md`, `verifier-p18-2026-11-15.md`. The fixed `fid/leech-r2` and `fid/crush-r2` are on the private fork remote only.

Tags: **V run** (run by this verifier today, NAS builder), **V read** (disassembly or source read by this verifier), **I** (inferred), **U** (unverified).

## Verdicts

| PR | Verdict | Post |
|---|---|---|
| Leech (D1 to D3) | **PASS WITH CHANGES** | `fid/leech-r2`: oracle `26ae7020`, claim `55cb4e68`, on upstream main `c3dc51c3`; body after L1 to L6 |
| Crushing Blow (D4) | **r1 FAIL, r2 PASS WITH CHANGES** | `fid/crush-r2`: oracle `1a75fed3`, claim `e5dc17d3`, on `fid/leech-r2`; post only after the leech PR merges, rebased on upstream main; body after C1 to C6 |

The rules are game.exe's and both claims implement them (V read and V run below). Neither r1 can be posted:

1. **Crush r1 breaks the leech it stacks on (blocker, V run).** FUN_0057c6c0 raises the strike events (`0x57c7c7..0x57c7f3`; the crush handler runs on event 5) and only then reads stat 6 to cut the physical for the leech (`0x57c7f8..0x57c820`). So with a Crushing Blow the leech is cut to the life the crush LEFT. Crush r1 kept the cut at the life before the crush. A new `crush.py pair` mode runs FUN_005bffc0 natively, then FUN_0057c6c0 natively with event 5 setting stat 6 to the life the handler left: on the r1-equivalent tip the leech differs in **1,068 of 2,810** (seed 7) and **6,589 of 20,810** (seed 1), every one a blow whose physical is over the life the crush leaves; the crush itself differs in 0. r2 computes the leech after the crush with `life - crushing_fine`: **0 and 0**. r1's own `crush.py compare` cannot see this: its harness has no steal.
2. **Identity and trailer (blocker).** All four r1 commits are authored and committed as `Claude <noreply@anthropic.com>` and carry no trailer (V run, `git log --format`). r2: the user's noreply identity as author and committer, `Agent: Claude Code` the only trailer, after a blank line (`git interpret-trailers --parse`, all four).
3. **No ctest pins the new rules (high).** r1's only test change re-pins one Normal-difficulty leech value (640) and the two crush values; a mutant without the divisor, without the cut to the life, or in whole points passes ctest. r2 adds game.exe's values (from the oracle) to `test_combat`: Nightmare 320 and Hell 213, a target with 20 points left 256, half a point of physical 12 (leech claim); a crush and a leech on one blow, damage 90 + 25 and leech 1,920 (crush claim, on its own `Rng` so the asserts after it keep their stream). The mutants now abort `test_combat` (section Mutants).
4. **Usage errors and the compares (medium).** r1 (V read): `crush.py` with no argument and `gen` without N give an `IndexError`; both compares check the lengths with `assert`, exit 0 on an empty case file ("0 differ"), never check the ids and give a traceback on a malformed line. r2 (V run on hand-made files): `usage:` and exit 1 on no, unknown or short arguments; the compares exit 1 with a message on an empty file, unequal lengths, a malformed line or ids out of step. ruff B904 (`raise ... from None`) fixed; both harnesses include `<cstdint>` for `std::uint32_t`.
5. **Base (low).** r1 is on `b6e6940d`; upstream main is now `c3dc51c3` (19 commits, three in `fight/`). The stack rebases with no conflict; r2 is on `c3dc51c3` and every run below is on it.

## Game side, re-derived (V read by this verifier, then V run)

- **FUN_0057c6c0** (the apply function; 26 call sites by a scan for `call` rel32): attacker in `edx` (kept in `ebx`), defender `[ebp+8]`, second argument `[ebp+0xc]`, record `[ebp+0x10]`. The resistance pass FUN_0057c1e0 runs here only when the second argument is non-zero (`0x57c793`). The events (`0x57c7a5..0x57c7f3`) run only when the record's word `+4` has bit 1 and not 0x20. Second argument 0: attacker event 5, then defender event 1. Non-zero: only when the record's dword `+0` has 0x20, attacker event 6 (skipped when it also has 0x80) and defender event 2. **Then** `0x57c7f8` loads `+8`, reads stat 6 again (FUN_00625480) and stores the smaller back, `0x57c826` calls FUN_0057c420, and `LAB_0057c852` takes `+0x4c` off stat 6 (0 under 256). decomp35's dead end "the second argument 1 saw no event" is this `+0` 0x20 test, not a missing event.
- **FUN_0057c420** (leech): stops when both steals are 0; a defender of type 1 takes the MonStats Drain byte `row + 0xa0 + [game + 0x6d]` (0 or less: none), any other defender 100. Both steals `<< 6`; for an attacker of type 0 only, each is divided (`idiv`, toward zero) by DifficultyLevels `+0x28` / `+0x2c` when that is not 0 (FUN_00611d30). Physical `+8` 0 or less: none. Mana first: `MulDiv(phys, steal', 100)` (FUN_00483360), `MulDiv(.., Drain, 100)` unless 100, `/ 64` toward zero (`cdq; and edx, 0x3f; add; sar 6`), to FUN_0057aa00; then life the same way to FUN_0057a980. A monster attacker branches at `0x57c58a` (not read further; not claimed).
- **Skills that leech:** Hunger (FUN_005c7f10) adds its calc2 and calc3 to the record's `+0x38` / `+0x3c` (`0x5c803a`, `0x5c8043`) before the apply, so in game.exe the divisor and the cut apply to it; d2d adds them to `Fighter::life_steal` / `mana_steal` (`strike.cpp`), the same path (V read both). The charge-up steals (`charge_bonus`, prgdam 2) were not read on the game side (U). No Paladin skill leeches.
- **FUN_005bffc0** (Crushing Blow, an event handler; no direct caller): the chance is `rand % 100 < stat 136` on the attacker's seed. Divisor: 10 for a player defender (`0x5c003c`); for a monster 10 if FUN_0063ee90 (hireling), else 8 if FUN_0063e9f0 or FUN_005a0180 (boss, unique), else 4, plus `MulDiv(base, FUN_005738f0(players), 100)`; doubled on event 6. `crush = stat 6 / divisor` (`idiv`), less `MulDiv(crush, min(stat 0x24, 100), 100)` when stat 0x24 is above 0; stat 6 becomes `life - crush`, 0 when not above 0 (and the record's `+4 |= 2`).
- **Deadly and critical strike** double the record's physical in FUN_0057b7d0 before the apply; the leech and d2d's `fine` both take the doubled value (V read). Open Wounds was not traced (decomp35 D8, U).

**V run** (oracle scripts of `fid/crush-r2`, the game image of `docs/emu-image.md`): `leech.py selfcheck` 0 of 20,351 (seed 7) and 0 of 60,351 (seed 1); `crush.py order`: events 5 and 1 see the case's life and the end life is `life - damage` (0 under one point) in 2,360 of 2,360; DifficultyLevels read at run time: (1, 1), (2, 2), (3, 3).

## Differentials (V run)

Case files: `leech.py gen 20000 7` and `gen 60000 1`, `crush.py gen 2000 7`, `crush.py gen-pair 2000 7` and `gen-pair 20000 1`. Harnesses from `fid/crush-r2`, compiled against each tree. "differ" = cases where d2d's value is not game.exe's.

| Tree | leech seed 7 (20,351) | leech seed 1 (60,351) | crush (2,360) | pair seed 7 (2,810): crush / leech | pair seed 1 (20,810): crush / leech |
|---|---|---|---|---|---|
| upstream main + leech oracle (the parent) | **17,097**, exit 1 | **50,626**, exit 1 | n/a | n/a | n/a |
| leech claim (`fid/leech-r2`) + crush oracle | 0, exit 0 | 0, exit 0 | **2,223**, exit 1 | 2,745 / 1,068, exit 1 | 20,738 / 6,589, exit 1 |
| crush r1's claim on the r2 base (`vd35-c1x`) | | | 0, exit 0 | **0 / 1,068**, exit 1 | **0 / 6,589**, exit 1 |
| crush r2's claim (`fid/crush-r2`) | 0, exit 0 | 0, exit 0 | 0, exit 0 | 0 / 0, exit 0 | 0 / 0, exit 0 |

decomp35's numbers reproduce exactly on the rebased tree (17,097, 50,626, 2,223). Every pair leech difference is in the class "physical over the life the crush leaves" (1,102 and 6,964 cases); none where the physical fits.

## Mutants (V run)

| Mutant (on `fid/crush-r2`) | leech seed 7 / seed 1 | crush | pair seed 7 / 1 (all) | `test_combat` |
|---|---|---|---|---|
| no difficulty divisor | 11,620 / 34,295 | 0 | 1,824 / 13,166 | aborts (`leech_fine(..) == 320 && .. == 213`) |
| no cut to the life | 3,133 / 9,216 | 0 | 1,074 / 6,619 | aborts (`nearly_dead` 256) |
| whole points | 17,068 / 50,579 | | | aborts (`blow.life == 640`) |
| no Drain term | 8,376 / 25,023 | | | aborts (`blow.life == 640`) |
| crush from the life after the hit (r1's parent rule) | | 2,241 | 2,809 / 20,796 | aborts (`damage == 50 + 0 + 50`) |
| leech cut before the crush (crush r1) | | 0 | 1,068 / 6,589 | TC_C1X |

## Build, ctest, tidy

Builder container, Release, GCC, tests and tools on, `-j2` under the builder's lane lock (V run).

| Tree | Build | ctest | test_combat, test_game |
|---|---|---|---|
| `fid/crush-r2` tree (claim tip; `vd35-c2c`, equal except two `from None` lines in the oracle scripts) | full, rc 0 | 2 of 342 fail: `test_outdoor`, `smoke_d2d` | ok, rc 0 |
| upstream main `c3dc51c3` | full, rc 0 | the same 2 of 342 fail | not run alone |
| `fid/leech-r2` tree (`vd35-l2c`) | test_combat, test_game | not run (the stack tip covers it) | ok, rc 0 |

The stack adds no ctest entry (its tests are asserts in `test_combat`); the per-test difference to upstream is none. The oracle commits change only `tools/emu/`, which is not in the CMake build, so each commit builds as its parent does.

TIDY_PENDING

## RNG luck

BUILD_RESEED_PENDING

## Stacking: two PRs, in order

The crush claim needs `Target::life_fine` from the leech claim, and its leech fix edits the leech lines, so it cannot be posted on upstream main alone. Post the leech PR first; post the crush PR only after the leech PR merges, rebased on upstream main (its diff is then the two crush commits). Posting both at once would show the leech commits twice and ask the maintainer to review a stack. If the maintainer asks for one PR, `fid/crush-r2` as a whole (four commits) is the fallback, with the two bodies merged.

## Required changes to the leech body (`decomp35-leech-body.md`)

- **L1 (high).** Tree: "upstream `c3dc51c3` plus the branch". Counts unchanged (table above), plus seed 1's mutants (34,295 / 9,216 / 50,579 / 25,023). Add: "`test_combat` pins game.exe's values for Nightmare (320), Hell (213), a target with 20 points left (256) and half a point of physical (12)."
- **L2 (high).** Replace "The full ctest result is in the card result of the branch's task" by the ctest row of section "Build, ctest, tidy".
- **L3 (medium, claims limited to proven).** Limits: "With a Crushing Blow on the same hit game.exe cuts the physical to the life the crush left (`0x57c7f8` reads stat 6 after the strike events); this PR keeps the cut at the life before the hit; the Crushing Blow PR that follows changes both." And: "The oracle runs the melee path (second argument 0); the missile path reaches the same `0x57c7f8` and FUN_0057c420 (read), not run."
- **L4 (medium).** "Smite does not leech (its record is marked built), as before" states a game behaviour the PR did not check: write "Smite's early return is unchanged".
- **L5 (low).** One line: "Hunger's calc2 and calc3 go into the same record fields (FUN_005c7f10 `0x5c803a`, `0x5c8043`), so the divisor applies to it too; d2d adds them to `life_steal` / `mana_steal`, the same path."
- **L6 (low).** Re-check block: `leech.py game` reads DifficultyLevels through `mpq-cat`, so it needs a build with tools and `D2_MPQ_DIR` (or `d2d.cfg`); say so as the neighbouring README rows do. Add "compare exits 1 on an empty file, unequal lengths, a malformed line or ids out of step".

## Required changes to the crush body (`decomp35-crush-body.md`)

- **C1 (blocker, fixed in r2).** "What changes": add "The leech of the same blow now cuts the physical to the life the crush left: FUN_0057c6c0 reads stat 6 for that cut at `0x57c7f8`, after the strike events." Counted differentials: add the pair rows (on the leech PR's tree 1,068 of 2,810 and 6,589 of 20,810 leeches differ, 0 after), the mutant "leech cut before the crush" and the `test_combat` row (damage 90 + 25, leech 1,920). Re-check block: add the `gen-pair` / `pair` / `--pair` / `compare-pair` lines (README row).
- **C2 (high, claims limited to proven).** Answer to "Hit paths": the missile path is read, not run: FUN_0057c6c0 raises events 6 and 2 only when its second argument is non-zero and the record's dword `+0` has 0x20 (event 6 not with 0x80), `0x57c7b1..0x57c7d6`; the claim moves the missile crush to the life before the hit on that reading.
- **C3 (medium).** Answer to "Unit kinds": FUN_005bffc0 is the same for every attacker, but which of FUN_0057c6c0's 26 callers a merc's or a pet's swing takes was not traced. Write "the merc's and the pets' blows take the same new crush in d2d; in game.exe the handler is the same, the apply path of their swings was not traced".
- **C4 (medium).** First line: "Follows #<leech PR>; posted after it merged."
- **C5 (low).** The divisor for a player defender is 10 (`0x5c003c`): one clause in the MonProp limit line (D5).
- **C6 (low).** Tree, ctest and mutant rows as in this file.

## Hygiene (V run, r2)

- Identity: all four commits have the user's noreply identity as author and committer (as `fid/tele-refund-r2`); `Agent: Claude Code` is the only parsed trailer.
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN on both heads; `pr_guard.py fid/leech-r2 HEAD` on the crush head: 0 FAIL, 0 WARN.
- `presubmit.py --quick --base upstream/main`: leech 0 FAIL, 3 WARN (siblings unit kinds, build, gcc); crush 0 FAIL, 4 WARN (adds siblings hit paths); the crush head against `fid/leech-r2`: 0 FAIL, 3 WARN. The bodies answer them (L3, C2, C3). r1's leech oracle had a ruff FAIL (B904), fixed.
- Diff: leech oracle 3 files (+299), claim 6 files (+41 -7); crush oracle 3 files (+383), claim 2 files (+26 -12). No private-only code, no host names or paths (read of the diff).

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill (decomp35's R1 to R3 recipes not run); nothing ran on the worker host.
- The missile path (event 6, the record flags), a monster attacker's leech (`0x57c58a`), the charge-up steals on the game side, Open Wounds, a player count above one (FUN_005738f0) and the MonProp crush on the player (D5).
- The pair oracle joins two native calls: the life FUN_005bffc0 leaves is written into stat 6 at the dispatcher's event 5. That FUN_005bffc0 runs there rests on `crush.py order` and the handler's own event test, not on a run of the item property's event table.

## Re-check commands (as run)

Builder container, the emulator image of `docs/emu-image.md`; lab scripts `verifyd35/{bld,chain,runs,stage,emu,tidy}.sh` (lab only).

```sh
python3 leech.py gen 20000 7 > lc.tsv; python3 leech.py game lc.tsv > lg.tsv; python3 leech.py selfcheck lc.tsv lg.tsv
g++ -O1 -std=c++23 -I<tree>/components/rules -I<tree>/components -I<tree>/components/d2s -I<tree>/components/txt -o lh leech_harness.cpp
./lh < lc.tsv > ld.tsv; python3 leech.py compare lc.tsv lg.tsv ld.tsv; echo $?                 # parent 17,097, 1; claim 0, 0
python3 crush.py gen-pair 2000 7 > pc.tsv; python3 crush.py pair pc.tsv > pg.tsv
./crush_harness --pair < pc.tsv > pd.tsv; python3 crush.py compare-pair pc.tsv pg.tsv pd.tsv; echo $?   # r1: leech 1,068, 1; r2: 0, 0
```
