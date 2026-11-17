# Verifier: gear absorb r3, the heal-back (2026-11-17)

> status: current, 2026-11-17 (card verifygearabsorb3, TASK-626, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/gear-absorb-r3` (oracle `982a65c1`, claim `270b9799`, on upstream `f1533770`), body `docs/upstream/bodies/gear-absorb-body.md`, trace `docs/research/gear-absorb-r3.md` (TASK-621). Checklist from `verifier-decomp35..39` (the r2 FAIL is `verifier-decomp37-2026-11-17.md`, A1 to A6). The fixed `fid/gear-absorb-r4` is on the private fork remote only.

Tags: **V run** (run by this verifier today, the builder container), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r3 FAIL (stale base, oracle recipe, gaps), r4 PASS WITH CHANGES

The rule is game.exe's and r3 implements it: 0 differ on 60,000 oracle cases (three seeds, one fresh) and on six hand cases at the edges (V run). r3 cannot be posted as it is:

1. **Stale base (high).** Upstream main moved to `895f8a67` (Mana Burn r3 merged: `tick.cpp`, `ai.hpp`). r3 rebases without conflict, but the posted commits must be the rebased ones, built and tested there.
2. **The README recipe fails (medium).** `test_gear_absorb --life` logs `the player died` to stdout (7,310 lines on seed 7, V run), so the README row's `test_gear_absorb --life c.tsv > d.tsv && ... compare` stops with `malformed or out-of-step line`. The body's re-check block has the `grep`; the README row and the script's docstring do not.
3. **Untested branches of the claim (medium).** Energy Shield's recompute carries the absorb with no test (r3 says so); a life above max (FUN_0057a980 brings it DOWN to max, V read and V run) and a blow the flat absorb takes whole (heal, nothing subtracted) are in no pin and in no oracle case (`gen` clamps life to max). Mutants m13, m14 and m17 below pass r3's tests.
4. **The merc reason is wrong (medium, claims).** r3 says d2d's merc "keeps its life in whole points with a fractional carry and no heal-back path". The merc has 256ths (`Monster::life_fine` / `set_life_fine`, `life_frac`) and is healed elsewhere (`missiles.cpp:1087`, `setup.cpp:288`, V read). Leaving the merc out is a scope choice, not a d2d limitation.

**Post `fid/gear-absorb-r4`**: oracle `ac5c3dd1`, claim `411bb48e`, on upstream main `895f8a67`; author and committer `shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>`, oracle first, `Agent: Claude Code` the only trailer after a blank line (V run, `git interpret-trailers --parse`). Against r3: rebased; the README row and the docstring pipe through `grep -E '^[0-9]+ -?[0-9]+$'`; pins 10 to 12 (game.exe's lives, V run); an Energy Shield case in `test_gear_absorb_rules`; the merc comment and `combat.md` say "not modelled yet", not "cannot". The body then needs C1 to C8 below.

## Game side, re-derived (V read by this verifier)

- **`FUN_0057bf10`** (`0x57bf10..0x57bf75`): the row's percent stat (`[row+0x10]`, `-1` skips) of the struck unit (`[ctx+0xc]`, `FUN_00625480`): at least 40 is 40, 0 or less skips, else `MulDiv(field, p, 100)` (`FUN_00483360`) is added to `[record+0x48]` and taken off; then the flat stat `<< 8`: 0 or less skips, else `min(flat, field)` (signed `jl`, so a negative field is cleared and gives a negative share) added and taken. Same as `rules::absorbed_fine` and `elemental_took`.
- **`FUN_0057bf80`** (`0x57bf80..0x57c057`): a field at or under 0 is stored 0 and stops (no absorb); the flat magic reduction, the resistance (at most 100) on what is above 0, then `FUN_0057bf10` only when the ignore-resistances argument is 0 (`0x57c01e..0x57c02b`).
- **`FUN_0057c6c0`**: `FUN_0057c1e0` (`0x57c7a0`), the on-hit events, the attacker's leech (`FUN_0057c420`: record `+0x38` / `+0x3c` to the ATTACKER, so not the defender's heal), then `[record+0x48]` to `FUN_0057a980` (`0x57c844..0x57c84a`), then `+0x4c` above 0 off stat 6, under `0x100` is 0 (`0x57c852..0x57c874`); Mana Burn and the rest after. A life at 0 sets result flag 2 (`0x57c93d`).
- **`FUN_0057a980`** (`0x57a980..0x57a9f0`): amount 0 or less, `FUN_005541b0` (dead) or state `0x5c`: nothing. Else `life + amount`, and when that is at least `FUN_00625d10` (max life) the max: stat 6 is SET to it. So a life above max with any positive amount is brought down to max (V run: case 10, 120 of 100 with a 10-point heal ends at 100 - 40 = 60).
- **`FUN_005541b0`** (`0x5541b0..0x5541f6`): dead is a mode test (player mode 0 or 0x11, monster 0 or 0xc, or flag `+0xc6` bit 0), not life 0. The oracle hooks it as "life at or under 0" and stops a case at the first death; the real switch to the death mode is outside `FUN_0057c6c0`. Whether a second record in the same server frame can reach a player at life 0 before that switch (and heal it) was not read (U, low: d2d's `absorb_heal` stops at the death, as the oracle does).

## Scope answers (the brief's a to e)

- **(a) The merc.** game.exe reads the absorb and heals through the struck unit's stat list whatever its type (V read: no type test in `FUN_0057bf10`, `FUN_0057bf80`, `FUN_0057a980`; `FUN_0057c6c0` tests the type only for a flagged monster). A hireling's item stats on that list is how its gear resistances work too, which d2d already models from the merc's gear (`merc_fighter`), so a merc's absorb gear almost surely acts in game.exe (I, strong). The parent has no merc absorb, so r3 / r4 change nothing for it: no regression, a known deviation left. Recommendation: keep this PR players only (r2's lesson: absorb without heal is half a rule), say it as scope, and make the merc a follow-up card (merc heal through `life_fine` / `set_life_fine` and `merc_max_life()`, an oracle case on a type-1 unit). Do not let the merc absorb without its heal.
- **(b) Energy Shield and other pools.** game.exe's Energy Shield takes from the raw field in event 0xb before `FUN_0057c1e0` (S, upstream's `test_energy_shield`), so the absorb works on what it leaves, as the claim recomputes in `energy_shield_blows`. r4 pins one case with the rules (shield 50 % paid 1:1, 20 % cold absorb, 100-point cold hit, 90 of 100 life: 60 left); m14 and m17 now abort. The value is game.exe's two rules composed by hand (S plus V), not an emulator run: the ES handler needs its skill row's calcs, which `gear_absorb.py`'s rig does not set up. Bone Armor takes physical only (no absorb row). Cyclone Armor is in the oracle (7,923 / 8,038 / 8,035 cases).
- **(c) Heal order.** Heal before the subtract: V read, pinned (case 3) and in the oracle. The attacker's leech in the same record goes to the attacker. d2d batches a tick's blows and `absorb_heal` replays them one record at a time (cap per blow; pin 4); Crushing Blow comes after, as in game.exe. Regen, potions and the player's own leech in the same tick are applied elsewhere in d2d's tick; the order only matters at the cap (U, low).
- **(d) Life above max.** game.exe brings it down to max on any positive heal (above). d2d's `absorb_heal` does the same (`gain` negative); the hurt path does not clamp otherwise (case 11: 110 minus 10 = 100 on both). Now pinned (10, 11), m13 aborts.
- **(e) Checks.** Below.

## Differentials (V run)

Game side `gear_absorb.py game` on the builder; seeds 7 and 5001 reproduce the author's case and game files byte for byte (`cmp`); seed 4242 is fresh. d2d side `test_gear_absorb --life | grep`, `compare`.

| Tree | seed 7 | seed 5001 | seed 4242 (fresh) | hand cases (6) |
|---|---|---|---|---|
| r4 oracle `ac5c3dd1` (parent d2d) | 9,639, exit 1 | 9,519, exit 1 | 9,484, exit 1 | 5 of 6 differ |
| r3 claim `270b9799` | 0, exit 0 | 0, exit 0 | 0, exit 0 | equal |
| r4 claim `411bb48e` | R4_7 | R4_5001 | R4_4242 | R4_HAND |

Hand cases (game.exe, V run): 120 of 100 life, 20 % cold absorb, 50 cold: 60; 120, flat 5, 10 cold: 95; 110, no absorb, 10 cold: 100; 100, 20 %, 100 cold: 20; 1 point, flat 20, 10 cold: 11; 1.17 points, flat 20, blows 30 and 10 cold: 21.17. `compare` exits 1 on an empty file, unequal lengths, a malformed line and an unknown verb (V run).

## Mutants (V run, local, never pushed)

MUTANTS

## Build, ctest, tidy (V run)

BUILD

## Hygiene (V run)

- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN on r4 (r3 against its base: the same). `presubmit.py --quick --base upstream/main`: 0 FAIL, 3 WARN (siblings: unit kinds; build; gcc), answered by C3 and C6.
- Identity as above; the oracle commit's test has no failing assert (no-argument run exits 77, skipped, on the parent tree).
- Diff: oracle 4 files, claim 10 files; no private code, host names or paths (read).

## Required changes to the body (`gear-absorb-body.md`)

- **C1 (high).** Commits: "upstream `895f8a67` plus `ac5c3dd1` (oracle) and `411bb48e` (claim)"; the Claims table rows and the counts (seed 7 9,639; 5001 9,519; add fresh seed 4242 9,484; claim 0).
- **C2 (high, template).** Add `## Checklist`, ticked: tests (`tests/test_combat.cpp`, `tests/game/test_gear_absorb.cpp`, `tests/game/test_gear_absorb_rules.cpp`), finding (`docs/research/re/combat.md`), `pr_guard` 0 FAIL 0 WARN. Keep the `Agent: Claude Code` footer.
- **C3 (high, claims).** "Players only" and the `siblings` answer: replace the merc reason by "game.exe reads the absorb and heals through the struck unit's stats whatever its type, the same list that carries a hireling's gear resistances, so a merc's absorb gear acts there too. This PR leaves the merc out on purpose: its absorb without the heal-back would be half the rule; the merc's heal (its life in 256ths through `life_fine`) is a follow-up. No change for the merc against the parent."
- **C4 (medium).** Limits: replace the Energy Shield line by "Energy Shield: the absorb works on what the shield left (`energy_shield_blows`); one case pinned in `test_gear_absorb_rules` from the two rules (the shield's from `test_energy_shield`), not from an emulator run." Add "A second record reaching a player at life 0 in the same frame before its death mode (game.exe's dead test is the mode, `FUN_005541b0`) was not read; d2d and the oracle stop at the death."
- **C5 (medium).** Evidence row `FUN_0057a980`: "`life + total`, at least max life gives max: a life above max is set DOWN to max (pinned: case 10)". Pins: "thirteen pinned lives" and add cases 10 to 12 to the list.
- **C6 (low).** WARN answers: `build` / `gcc` with this file's ctest numbers on r4.
- **C7 (low).** Re-check block: base `895f8a67`, and add seed 4242.
- **C8 (low).** Mutants: add m13 to m17 of this file.

## For the coordinator (not body changes)

- Trace `gear-absorb-r3.md`: Q6's merc reason (above), the "heads are final" line (r4), A4's open Energy Shield point (closed by r4's pin).
- Follow-up card: the merc's absorb and heal (D3 of the trace), an oracle case on a type-1 unit.
- Bot (S, opinion): the trace's B1 (`internal/power/hero.go:118..123`) undervalues absorb below full life (F - 2A, not F - A); worth a bot card once this lands.

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill; nothing on the worker host.
- The ignore-resistances record, death_delay (state `0x5c`), the burn and blast paths, a missile's own element roll (the body's Limits; U).
- Energy Shield on game.exe under the emulator (see b).
