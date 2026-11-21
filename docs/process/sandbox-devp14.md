# Sandbox d2d build devp14 (input verbs candidate, 2026-11-19)

> status: candidate, LOADED only on the sandbox host, NOT switched (the switch needs the owner's go). Task: inputlayer (R17 stages I1 and I2 of docs/design/fable-review-2026-11-19/07-input-layer.md; question Q13 of that review is still open). Tags: V = verified in this run (how said), S = read from source or a tool, U = unverified.

## What devp14 is

devp13 (`didle/rebase-20261119`, tip `f32837d8`) plus the fork branch `didle/input-verbs-20261119` (private remote only, never upstream without the owner's go). Build ref `devp14`. Image `d2dbot:devp14`, id `f5bc525d159c`, loaded on the sandbox host next to `devp13` (`97e985ca5eb5`); the running container `d2play-20261108` is still on `devp13` (V: `docker ps` after the load). Fork tip at build: `cef4abfd` (the two later commits change only `tests/input_verbs_d2d.py`; the binary, sha256 prefix `d3a11812638bf219`, equals the builder's). Built on the builder's main lane in 773 s (long lane was held by another job), d2d only, rc 0.

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

On the builder, game data there, `D2_MPQ_DIR=/data/gamedata`:
- Pass (rc 0): `input_verbs_d2d` (new), `lab_verbs_d2d`, `item_verbs_d2d`, `vault_verbs_d2d`, `item_colour_d2d`, `item_create_identified_d2d`, `item_pic_d2d`, `labels_d2d`, `merc_verbs_d2d`, `cursor_verb_d2d`, `overlay_d2d`. V.
- `smoke_d2d` rc 1: the known upstream roster-order failure (docs/process/sandbox-devp13.md), not compared again here. U.
- `input_verbs_d2d` V: `caps` lists the three input names; `cmd select 7 0` (right, Fire Arrow on the class the headless start gives with `--start-class 3`, the list `debug points` accepts is 6..35) sets `state` rskill and it is still set 1.5 s later; a second right select and a left select change their own hand only; an unlearned skill is refused. This is the R08 fix (before: only the message went and the hand reverted a frame later; S: source of walk.cpp). On devp13 the same select left `rskill` unchanged (V live, docs/palready.md).
- NOT verified: `menu choose` and `cmd repair` (the headless hero never reached the vendor in 4 runs: the walk stalled about 4 cells short, so the part prints SKIPPED); the select differential against the click path; and the priming-click quirk: `cmd move` right after `debug goto` moved 0.2 cells in 4 s, but a plain `cmd move` also moved only 1.8 of 6 cells in 8 s in the same slow headless run, so the observation is inconclusive (U).
- A Paladin aura on the real bot path is the Proving Ground scenario `input.skill-select-verb` (pending, needs the switch).

## Switch (needs the owner's go; not done)

Same recipe as devp13: `docker stop -t 60 d2play-20261108`; back up the volume; `D2_IMAGE=d2dbot:devp14 docker compose -p play-build-20261108 -f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv2.yml up -d --no-deps d2dbot`; then the bot with `input.planner=1`. What the user sees: aura and skill switches without the popup flashing open, `input=N/0/0` in the minute line.

## Rollback

Same compose command with `D2_IMAGE=d2dbot:devp13`; devp13, devp12 and devp10 stay loaded. Set `input.planner` back to 0.
