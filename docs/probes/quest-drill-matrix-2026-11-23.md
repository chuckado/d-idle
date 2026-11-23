# Quest drill matrix, 2026-11-23 (card qdmatrix)

> status: current, kept up to date through the night of 2026-11-22. Tags: V verified (named run), S secondary (source reading), U unverified. Every setup is CHEATED (level, gold, waypoints, quest words, items, merc by SETUP verbs): lab mechanics only, never a statement about the live hero. Scenario file: `tools/provingground/scenarios/questdrill.json`. Runs: Tower builder long lane, lab d2d, `tools/provingground/run.sh -id <id> -class <C> -seed <n>`.

Knobs of the sandbox as scenario tuning: `nav.warpRead=1 nav.fullGrid=1 quests.boardSM=1 town.plan=1` (a baseline with the knobs off is a second run).

| chunk | scenario | class / seed | verdict | first failing step | fix card / PR | merged sha |
|---|---|---|---|---|---|---|
| sisters_to_the_slaughter/andariel | drill.andariel.clvl19 | Paladin 11 | RED (readiness holds it) | board: too hard (see notes 1) | | |
| sisters_to_the_slaughter/andariel | drill.andariel.kit19 | Paladin 11 | running | | | |
| act1_done/sail_east | drill.act1.sail | Paladin 11 | not run | | | |

## Notes

1. andariel clvl 19 (V, job 20261123-000339-c312, cancelled after 6 min): the tapper logs `not startable ... readiness "too_hard"`: `You would likely lose to Andariel and have an unread attack`, later `a dead merc` and `short of potions`. The zone shadow line shows `weakest=offence_weapon_class_default dps=4.8` for a short-sword hero: the margin is the hero's real weak weapon (2-7), not a logic fault (S). The lab merc (`debug merc hire 0`) is a level 1 Rogue (`merc: seen ... level=1`) and dies at once in the Catacombs, so every run says `dead merc` after the first trip (lab limit, not a bot defect).
