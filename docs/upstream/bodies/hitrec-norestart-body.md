## What changes

A player in a get-hit or a block is not put into a new get-hit by a big hit. `FUN_0057edd0(unit, new mode)` is the first call of `FUN_005809d0`, the server's player mode switch, and its current-mode table (`0x57eea8`) refuses mode 4 (and 9) from DT, GH, BL and DD unconditionally; when it answers 0 nothing is set and no end event moves. d2d tested BL only, so every big hit inside a get-hit called `set_pmode(GH)` again and restarted the hold: two big hits 5 ticks apart held 15 ticks in d2d, 10 in game.exe.

One condition in `Fight::hurt_side`. d2d draws game.exe's knockback (mode `0x13`) as GH with `knock` set, and game.exe ALLOWS mode 4 from `0x13` (`cmp eax, 0x12; ja 0x57ee68`, allowed, at `0x57ee0b`), so only a real get-hit (GH without a `knock`) refuses: a big hit during a knockback slide is a get-hit, and it ends the slide. One hit is a knockback OR a get-hit in game.exe (the flag chain at `LAB_0057d364`), so a call that just started a knock starts no get-hit either.

Scope: this PR changes what a plain big hit does inside a get-hit. It does not change a knocking hit inside a get-hit (see the limits).

Three commits: the d2d probe (`test_hit_recovery --restart` and `--hit-modes`), the game.exe oracle `tools/emu/mode_switch.py` (it builds and runs alone against the probe), then the claim (`tick.cpp`, two stale comments, `test_hit_recovery`, a paragraph in `docs/research/re/combat.md`).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0057edd0` `0x57eea8` (current-mode table) | for new mode 4 and 9: DT, GH, BL, DD refused with no other read; A1, A2, SC, TH allowed; KK needs a queued end event; S1 refused for class 0 (Amazon), S3 for class 5 (Druid); SQ (18) needs an end event |
| `0x57ee0b` | any current mode above 0x12, including the knockback `0x13`, is allowed |
| `0x580adc` (`FUN_00580a70`, the knockback's switch) | calls `FUN_0057edd0` with new mode 0x13: refused from GH, DT, BL, DD (not changed here, see the limits) |
| `0x57d364..0x57d3c3`, `FUN_005809d0` (`0x5809fd`) | flag 8 goes to the knockback and returns, flag 4 to the get-hit: a hit is one or the other; a refusal returns at once with no state, no flag, no event |

## Counted differentials (V run: game.exe's `FUN_0057edd0` under the emulator, d2d built and run, same day)

`mode_switch.py game` (game.exe from `D2_MPQ_DIR`/bin, or `GAMEEXE`) calls `FUN_0057edd0` for modes 0..19 (`DT NU WL RN GH TN TW A1 A2 BL SC TH KK S1 S2 S3 S4 DD SQ KB`), new modes 4 and 9, with and without a queued end event, classes 0..6 (80 `M` lines, 8 `C` lines for the class-dependent cells). `test_hit_recovery --hit-modes` gives d2d's side: one big hit on a Paladin in each mode 0..18 with a hold running, and in a knockback slide as mode 19. `compare` judges 13 cells and lists 7 excluded by name (DT and DD, a dead unit; A1 KK S1 S3 S4, swings, not interrupted in d2d: separate claim).

| Build | cells differ | `--restart` |
|---|---|---|
| probe `eacb46e2` (parent) | 1 of 13: GH from GH (restarted 1, game refuses) | `R 10 15` |
| oracle `f40e6769` (builds and runs before the claim) | 1 of 13 (GH) | `R 10 15` |
| claim `a2e61a4c` | 0 of 13 | `R 10 10` |
| the claim's tests with upstream's `tick.cpp` (mutant, not pushed) | 1 of 13 (GH); `test_hit_recovery` aborts at its restart check | `R 10 15` |

The oracle's `compare` exits 1 on any difference, on zero compared cells, and on a missing, repeated or malformed line on either side (`mode_switch.py selftest` runs 9 checks of that, no emulator needed). `test_hit_recovery` (the knock test: a big hit in a slide enters GH, ends the slide, moves the hold's end), `test_game`, `test_combat`, `test_rules`, `test_swing_start`: rc 0.

## How to re-check

```sh
set -o pipefail
cmake --build build --target test_hit_recovery test_game test_combat
cd tools/emu && uv sync
export D2_MPQ_DIR=<the game data folder, with bin/game.exe>   # or GAMEEXE=<path to game.exe>; EMU=<dir of emu.py> only when run from elsewhere
../../build/tests/test_hit_recovery --hit-modes | grep '^M ' > d2d.tsv
uv run python mode_switch.py game > game.tsv
uv run python mode_switch.py compare game.tsv d2d.tsv     # claim: 13 cells, 0 differ; probe commit alone: 1 differ (GH from GH)
uv run python mode_switch.py selftest
../../build/tests/test_hit_recovery --restart            # claim R 10 10; probe R 10 15
cd ../.. && python3 tools/pr_guard.py && python3 tools/presubmit.py --quick
```

## Checklist

- [x] A test in `tests/` (`test_hit_recovery`: two hits 5 ticks apart, and a hit in a knockback slide) and the oracle in `tools/emu/`, same pull request (probe, oracle, claim; the oracle commit builds and runs before the claim)
- [x] The finding in `docs/research/re/combat.md`, under the block-length paragraph
- [x] `python3 tools/pr_guard.py`: 0 FAIL, 0 WARN
- [x] `python3 tools/presubmit.py --quick`: 0 FAIL, 2 WARN (build and gcc, skipped by `--quick`; built and tested by hand)
- [x] clang-tidy-19 on `fight/tick.cpp` and `test_hit_recovery.cpp`: 0 findings

## Honest limits

- NOT changed by this PR: a knocking hit inside a get-hit. `FUN_00580a70`, the knockback's own switch (flag 8, `0x57d332`), calls `FUN_0057edd0` with new mode `0x13` (`0x580adc`), which is above 0x11, so the same current-mode table decides and refuses it from GH (V read; also run in the emulator by the verifier: refused from GH with and without a queued end event). d2d still starts the knockback there (`tick.cpp` knock block), and that restarts the hold. The 13 cells here judge new modes 4 and 9 only, so they do not see it (U against a differential; follow-up candidate).
- DT and DD (a dead player) are excluded from the 13 cells and not traced: d2d's probe answers "entered a get-hit" there, game.exe refuses; harmless only if d2d never hurts a dead player (U).
- A refused big hit while in a block falls to d2d's struck-sound branch; game.exe sets nothing on a refusal. Existing behaviour, untouched.
- Not modelled, refuses more only: `FUN_0057edd0`'s `unit + 0x60` branch (a player holding a cursor item is refused mode 4 from modes 0..20 in the emulator; d2d has no such rule). It cannot make a get-hit from a get-hit allowed.
- A hit whose knockback has no room (game.exe: flag 8 still goes to the knockback, the player stays NU) still gets a get-hit in d2d: not traced (U), unchanged here.
- Where game.exe ends a slide when a get-hit starts is U; d2d keeps the position and ends the slide.
- A call that both knocks back and rolls big now plays the knock only (the flag chain, V read); the test pins the knock-slide hit, not this same-call case (`World` in the test has no level to block the knock).
- A refused big hit plays no struck cue in game.exe (V read: only the too-small path sets `0x8000`); the client's own presentation of the damage is not traced.
- BL from GH and from attack modes is d2d's existing `pmode < 0` rule, untouched; swing modes stay uninterrupted; monster and merc get-hits use other functions and are unchanged.
- The boundary tick (a second hit exactly at the hold's end) is V read only: d2d clears the mode before `hurt_side` in the same tick, as game.exe runs events before units.

---

Agent: Claude Code
