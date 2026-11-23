# Step-level proving grounds (card stepgrounds, 2026-11-23)

> status: current. Tags: V verified by a named run or probe, U unverified. Every setup is CHEATED (lab verbs, declared in each scenario's `provenance`): mechanics only, never evidence about the live hero or game balance.

## What it is

Every step of the quest board (`data/questboard/act1..5.json`, 134 steps) has an explicit status in `docs/probes/quest-step-catalogue.md` (generated): a step scenario, only a chunk-level scenario, or `MISSING F<n>|bot|data: why`. `go run ./cmd/stepcat check` (also `go test ./cmd/stepcat`) fails when a step has neither.

## Adding a step (data only)

1. Add or edit its row in `data/stepgrounds/steps.json`: `gen` (start area by name, quest words, ground or pack items, spawns, expected `enter` / `kill` / `kills` / `picked`), or `scenario` (a hand-written id), or `chunk`, or `missing`.
2. `go run ./cmd/stepcat gen` (writes `tools/provingground/scenarios/stepgrounds_gen.json`), `go run ./cmd/stepcat catalogue`.
3. Run it on the builder (long lane, one job at a time, a private `PG_DIR`): `labrun BUILDER --lane long --detach --pull provingground-out -- env PG_DIR=/data/d2d-lab/pg-stepgrounds PG_D2D_BIN=/data/d2d-lab/qdbin/d2d-devp23-labverbs tools/provingground/run.sh -id <ids> -class Paladin -seed 11 -parallel 1`.
4. Write the result into `data/stepgrounds/verdicts.json` (GREEN needs a run id and tag V; the check enforces it) and regenerate the catalogue.

Hand-written scenarios of this card live in `tools/provingground/scenarios/stepgrounds.json` (the Smith scenarios). Generated ones are never edited by hand.

## The Smith step (the example of the brief)

- `step.act1.smith.malus_pickup` (and `.sorc`): the Malus lies on the Barracks floor (`debug give hdm 2`), two Fallen beside the hero, the bot farms the Barracks. GREEN on Paladin 11 (run 20261123-175946) and Sorceress 3 (run 20261123-180122): `loot: clicked item=Horadric Malus`, `Found Horadric Malus`, `hdm` in the pack. Must-fail control `step.act1.smith.stand_click_pickup.control(.sorc)` (no Malus, only Fallen) is GREEN: it asserts the Malus is absent and never clicked, so the pickup assertions discriminate.
- Finding (V runs 20261123-175729 and -175828): with the board chunk running (`Kill The Smith`) the bot did not pick the Malus lying on the floor; it only picked it in a plain farm goal. Cause not isolated (the boss script clears and the leveling plan walks the hero away first); a bot finding for the toolstrade owner, not fixed here.
- `step.act1.smith.stand_click_pickup` (Smith dies, room clear, the bot clicks the Malus stand, hdm in the pack): PENDING on the bot card toolstrade a7b2ec8e. Correction by the fork agent: the d2d drops the Malus only from the stand (object 108), never from the corpse; my probe `stepdrop.py` agrees (no hdm after a spawned Smith dies, quest 3 started).

## Probes added

`tools/probes/stepdrop.py` (spawn then kill then read the ground), `stepdrop_real.py` (find the level-placed superunique by a warp lattice; did NOT find the Barracks Smith, U), `giveground.py`, `spawnids.py`. Fork verbs missing: `docs/probes/stepgrounds-fork-request.md`.

## Round 2 (2026-11-23, same card)

- The 4 RED generated scenarios of round 1: the shaft pair was a prerequisite lock (the board parks the chunk "locked until Radament's Lair"); with the Act 2 words in the kit (`kits.act2` of steps.json, records 4, 6, 7, 9, 11) both are GREEN (run 20261123-183100). Andariel (Catacombs 4) and Palace Cellar stay RED: a pack-full town chore (cause U) and a death.
- MISSING (bot): the talk steps need a data row NPC to quest record and word (U) plus the observer; the observer half is done: `cmd/d2scenario/questwords.go` merges `debug quest words` (all 48 words of the active difficulty, devp23) into the state for records `debug quest list` did not print, so assertions on records 16 to 47 are observed on a lab d2d that has the verb (must-fail tests in questwords_test.go). U live: not yet run against a quest assertion on records 16 to 47.
- Eight travel steps of Acts 1 to 5 got generated scenarios with the named kits `stand30` and `stand60`; all RED on the builder (deaths in Acts 3 to 5, and the stand60 words read Mephisto, Izual and Terror's End done). Verdict rows say so; the scenarios are a starting point, not a result.
- Ledger noise: runs on a labrun-synced tree (no .git) printed "ledger line refused, missing bot_commit"; tools/provingground/run.sh now stamps `tree-<hash of the Go sources>` when git gives nothing (V: the runs of round 2 print no ledger line).
- Open for the toolstrade owner: with the board chunk running, a Malus lying on the Barracks floor is not picked up (runs 20261123-175729 and -175828, RED) but is under a plain farm goal (run 20261123-175946 GREEN). The message to the owner could not be delivered (agent not reachable); the coordinator passes it on.
