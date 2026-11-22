# The Countess three ways: nav.fullGrid and nav.warpRead on a lab slot (card mapverify, 2026-11-22)

> status: current. Scenario `act1.forgotten.countess` (Paladin) and `act1.forgotten.countess.sorc` (Sorceress), `tools/provingground/knobs3way.sh`, lab d2d of the fork branch `didle/rebase-20261122-fix1` (611fc53b, image devp19; the sandbox is NOT on it and was not touched), builder lab slots, one run at a time. Setups are CHEATED (waypoint, given item, clvl 24 hero): mechanics evidence about nav, never about play balance. Tags: V = measured in the named run, S = reading, U = unverified.

## Method

`D2SCEN_KNOBS="nav.fullGrid=1,nav.warpRead=1"` (new, `cmd/d2scenario/run.go`, test `TestEnvKnobs`) layers knob overrides over the scenario's own `tuning.json`, so one scenario runs three ways with the same bot build: `off` (both 0), `grid` (`nav.fullGrid=1`), `gridwarp` (`nav.fullGrid=1`, `nav.warpRead=1`). `knobs3way.sh <scenario> <class> "<seeds>"` runs them serially (`-parallel 1`), `tools/probes/knobsummary.py provingground-out/k3-<stamp>` prints the table below from `results.json` and the bot log. Columns: goal game s = bot clock at the Countess kill (virtual, lockstep), clicks = `nav.trace click` lines up to the kill (the step count), mouth gate = `nav mouth gate` lines (click held, rewritten, sent inside the bound) up to the kill, stuck = `nav stuck` up to the kill.

The Paladin runs were built from the commit before the Monastery fixes of this card, the Sorceress runs after them (they only act after the Countess, at the Monastery Gate): compare conditions inside a class, not across.

## Paladin, seeds 11, 12, 13 (all 9 runs GREEN, no death before the kill)

| condition | seed | goal game s | clicks to goal | mouth gate | stuck | no usable gate (whole run) |
|---|---|---|---|---|---|---|
| off | 11 | 1208 | 1721 | 23 | 52 | 5 |
| off | 12 | 683 | 1032 | 0 | 29 | 20 |
| off | 13 | 1242 | 1720 | 32 | 50 | 3 |
| grid | 11 | 1208 | 1721 | 23 | 52 | 5 |
| grid | 12 | 683 | 1032 | 0 | 29 | 19 |
| grid | 13 | 1242 | 1720 | 32 | 50 | 3 |
| gridwarp | 11 | 713 | 1040 | 0 | 26 | 20 |
| gridwarp | 12 | 683 | 1032 | 0 | 29 | 19 |
| gridwarp | 13 | 858 | 1169 | 0 | 32 | 6 |

Mean over the seeds: off and grid 1044 s and 1491 clicks, gridwarp 751 s and 1080 clicks (28 percent fewer of both), mouth gate lines 55 against 0.

- V: `nav.fullGrid=1` changed nothing. Every run of `grid` has the same click count, goal time and gate lines as `off` on the same seed (the lockstep run is deterministic: 1721, 1032, 1720). Expected from the code (S): on d2d the grid is whole on arrival (gridwhole, 106 runs) so the only difference is that a goal outside the rectangle is refused, and the Countess walk never asks for one.
- V: `nav.warpRead=1` removed every mouth-gate refusal before the kill (55 lines in `off`, 0 in `gridwarp`) and shortened the walk on the two seeds that had them (11: 1721 to 1040 clicks, 13: 1720 to 1169); seed 12 had none and is identical. All GREEN.
- Deaths: none before the kill in any run. `gridwarp` seed 11 died once AFTER the quest at the Monastery Gate (Returned Mage, 12 hostiles near, clvl 24 against monster level 8, the post-quest level plan); no other run died. U whether the knob had a hand in it (the run arrives there at a different game time than `off`, so the fight is a different one); one run is no evidence either way.
- "No usable gate" in the last column is the Monastery Gate to Outer Cloister problem of `strip-monastery-2026-11-22.md`: it happens after the quest, in every condition, and is unrelated to the knobs.

## Sorceress, seeds 3 and 4: no comparison possible (BLOCKED)

V: 5 runs started (off, grid, gridwarp on seed 3; off, grid on seed 4; the job was cancelled after that, 150 minutes of lane for 9 runs did not fit), all BLOCKED `unobserved: quest record 5 not observed`: the lab Sorceress runs out of mana (`combat out of mana: backing off` 1178 lines in `off` seed 3, 300 and more `nav stuck`) and never reaches the Countess, in every condition alike. This is the lab-pace block already recorded for the board scenarios (`boardsm-progress-2026-11-21.md`), not a knob effect. Consequence: the Sorceress half of the three-way table is EMPTY; nothing here says how the knobs act for a caster. Next step is a Sorceress setup that can fight (spent skill points, mana potions) or a Paladin-only decision.

## Monastery Gate to Outer Cloister, the side finding

The Countess runs do not answer it, the strip probe does: `strip-monastery-2026-11-22.md` (placement of linked levels the layout lacks, then the two doors). `act1.andariel.slaughter` Paladin 11 went RED to GREEN on those fixes.

## Memo: default ON?

For the coordinator and the user; nothing was flipped in this branch.

| knob | verdict | evidence | missing |
|---|---|---|---|
| `nav.fullGrid` | safe to default ON for d2d | V: 3 Paladin seeds, `grid` identical to `off` to the click (1721, 1032, 1720); grid whole on every level probed (gridwhole, 106 runs); the code only refuses a goal outside a whole grid | no run where a goal outside the rectangle is actually asked for (so the refusal itself is untested live); a caster (Sorceress runs BLOCKED); non-d2d backends keep the old behaviour by construction (S) |
| `nav.warpRead` | safe to default ON once the sandbox has the fork verb (image devp19 or later); without the verb it reads nothing (S, `internal/d2dworld/triggers_test.go`) | V: 3 Paladin seeds GREEN, mouth-gate refusals 55 to 0, clicks and goal time down 28 percent, no death before the kill | one Sorceress or other-class run; more seeds (3 is thin; the only post-quest death was in a `gridwarp` run, cause U); a run on the sandbox image itself, which was not touched |

Both knobs: every run is a CHEATED setup, so this is evidence about nav mechanics only. Not run: seeds beyond 13, the other class, a real game.exe.

## Quest board default (card 10 `boardsm-default`, knob `quests.boardSM`; `quests.board` itself already defaults to 1 since retirej)

Not flipped, and the evidence is still short (S, from `backlog/tasks` and `boardsm-progress-2026-11-21.md`, `board.json`): the card needs "the scenario set green on a determinism pair". Today 5 of the board scenarios named `board.sm.*`, `hold.*`, `park.*`, `handin.*` stay `pending`: `regress-proxy` (a setup that leaves the stones done cannot be made), `park.persist-restart` (no deterministic park), `hold.too-weak` (no guard refused at clvl 10), `hold.timeout` (no hero that stays broke) and `park-retry-after-release` (never run). The Sorceress half of the set is BLOCKED by the same lab pace as above, so no class pair is green; the full set has not been run on a determinism pair with the knob at 1; and the panel text of bsm8 was never looked at on a live page. The old paths and the dual-write stay until a release after the flip.
