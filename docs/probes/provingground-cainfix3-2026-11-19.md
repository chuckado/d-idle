# Proving Ground: Search for Cain fixes (card cainfix3, 2026-11-19)

> status: current. Tags: V verified (run named), S secondary, U unverified. Builder, lab d2d slots, bot of branch agent/cainfix3, one scenario per job, `tools/provingground/run.sh -id <id> -class <C> -seed <n> -parallel 1`. Every setup is CHEATED (see each scenario's `provenance`): mechanics evidence only, never balance or play evidence.

## Verdicts (before from docs/probes/provingground-run-2026-11-19.md, after from this card)

| scenario | before | after | what changed it |
|---|---|---|---|
| cain.stones Paladin 11 | RED | GREEN (V) | fight grace and no re-gate of a running chunk (F4), no-gold chunk skips the purse recovery (F3) |
| cain.stones Sorceress 3 | RED | GREEN (V) | same |
| cain.cage Paladin 11 | RED | GREEN (V) | scenario setup (potions, key, chunk accepted at setup, bit 13 assert), grace 20 s, idle tick spends time |
| cain.cage Sorceress 3 | RED | GREEN (V) | same |
| cain.scroll Sorceress 3 | RED | GREEN (V) | fight grace (F4) |
| cain.chain Paladin 11 | RED | GREEN (V, 80 virtual min, 815 s wall) | all of the above plus the quest crowd line held between slices (F2) |

## Findings and fixes

- F1 (area ban): the ban itself is a death-spiral breaker of the farm plan. A chunk the player started is no longer stopped by it unless the readiness word of the chunk is `too_hard` (`Runner.banWaive`); the refusal that remains names the deaths and the end time ("area banned after N deaths (until HH:MM)", `task.World.BanNote`). Nav-fault and border-loop skips are untouched. U: no run in this card hit a death ban, the behaviour is unit-tested only (`TestBannedAreaNoteAndWaiver`, `TestFieldReadyBanNote`).
- F2 (crowd): `combat.crowdAbandonQuestN` (40, 0 = off) replaces `crowdAbandonN` (15) while a started chunk runs and its readiness is not too hard. The first chain run in this card (crowd 23 to 24 in Tristram, "abandoning the walk" lines in a potion trip) showed the line must stay up between board slices, not only inside one (`Fighter.SetQuestCrowd`, lowered by `boardStep`/`boardPreemptsWhy` when no chunk runs). The `crowd-retreat` (stepping away at life 70 or less when surrounded) is unchanged. U: the cage archers were not a crowd in the cage runs (the chunk freed Cain within seconds), so the override is proven by the chain run and unit tests, not by a dedicated scenario.
- F3 (purse recovery): `Chunk.NoGold` (data/questboard, true for scroll, stones, free_cain) lets a chunk start during the recovery, on the preempt path and the campaign path.
- F4 (start priority): after `boardQueueGrace` (20 s virtual) an accepted chunk starts although hostiles are near, and a chunk under way is never gated by a fight again (`boardRT.startedKey`). U: the 20 s is a judgment value.
- F5 (scenarios): purse 5000 (stones, cage), `resource.pointsDecisions` 1 was already set, the crowd-abandon default is no longer switched off in cain.cage and cain.chain, cain.cage has potions, the key and the chunk accepted at setup, and asserts bit 13 of record 4 (the report clears bit 1).

## Found on the way

- A tick of a chunk with no step to run (free_cain started outside Tristram) did nothing, so under the lockstep clock no time passed and the failure counter never advanced: 17 minutes wall in a field, ending in a death. The tick now goes to town (`rest`) and three counted failures park the chunk with its reason (V cain.cage run 192200: parked after 2 minutes).
- free_cain started outside Tristram cannot do anything: a `to-tristram` node does not help (V: `exit to Tristram: nav: timeout: roaming for 5m0s`; the portal is only taken by the stones step). Left as is; the chunk parks with the reason above. Candidate for a card: let free_cain take the Tristram portal.
- Stones from Underground Passage: the first stones run showed the level goal's `travel` taking the hero out of Stony Field again between two board steps (the fight gate re-queued the chunk each tick); fixed by `startedKey`, not a nav defect.
