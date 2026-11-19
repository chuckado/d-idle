# Sandbox d2d build devp14 (input verbs candidate, 2026-11-19)

> status: candidate, LOADED only on the sandbox host, NOT switched (the switch needs the owner's go). Task: inputlayer (R17 stages I1 and I2 of docs/design/fable-review-2026-11-19/07-input-layer.md; question Q13 of that review is still open). Tags: V = verified in this run (how said), S = read from source or a tool, U = unverified.

## What devp14 is

devp13 (`didle/rebase-20261119`, tip `f32837d8`) plus the fork branch `didle/input-verbs-20261119` (private remote only, never upstream without the owner's go). Build ref `devp14`. STATUS_PLACEHOLDER

## Verbs

| verb | message | state | bot side |
|---|---|---|---|
| `cmd select <skill> <left>` | 0x3c | fixed: sets the skillbar first (S: walk.cpp reverted a bare 0x3c one frame later, deviation R08) | `internal/input` planner, `combat.setPlanned`, knob `input.planner` |
| `cmd repair <item\|-1>` | 0x35 | new | not wired yet |
| `menu choose <text>` | client UI branch | new | not wired yet |
| `cmd use / buy / sell / trade / waypoint / belt / stat / skillpt` | various | exist already | not wired (I3 to I5) |
| `cmd move` without the priming click | 0x01 | open (U), the test prints an observation | `Kick` stays |

`caps` lists `input.select,input.repair,input.menu-choose`; the planner uses a verb only when `caps` names it (devp13 does not, so with the knob on it still falls back to the verified popup click).

## Bot side (all behind `input.planner`, default 0)

Skill and aura selection: verb when advertised, else key, else popup click, each verified by reading the hand back; the path is logged (`input path`) and counted in the minute line as `input=verb/key/click` (only when the knob is on). Proving Ground scenario `input.skill-select-verb` (pending until the switch).

## Tests

RESULTS_PLACEHOLDER

## Switch (needs the owner's go; not done)

Same recipe as devp13: `docker stop -t 60 d2play-20261108`; back up the volume; `D2_IMAGE=d2dbot:devp14 docker compose -p play-build-20261108 -f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv2.yml up -d --no-deps d2dbot`; then the bot with `input.planner=1`. What the user sees: aura and skill switches without the popup flashing open, `input=N/0/0` in the minute line.

## Rollback

Same compose command with `D2_IMAGE=d2dbot:devp13`; devp13, devp12 and devp10 stay loaded. Set `input.planner` back to 0.
