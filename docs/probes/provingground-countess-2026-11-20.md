# Forgotten Tower: finding the Countess on d2d (card countess, 2026-11-20)

> status: current. Tags: V verified (named run), S secondary, U unverified. The scenario setup is CHEATED (level, gold, waypoints, potions by SETUP verbs): mechanics evidence only.

## Findings

1. V (builder long lane, K1 lab d2d, jobs 20261120-144848-dd8d, 20261120-160224-c5c8, 20261120-162206-c36f): Paladin/11 reaches Tower Cellar 5 (area 25) with `quests.layoutRead` 1 and `debug levelmap` answered (264 calls). The level lists 32 object presets (classes 269, 37, 240, 7, 11, 105, 106), no monster presets and NO object 580 (the Kolbot "Countess chest"). So the layout read does not show her room through the chest; the chest-preset anchor tried first was wrong and was dropped. Rooms/monsters of levelmap were not logged (U whether a preset room row names her room).
2. V (same logs): on d2d `Env.World` is nil, so `bossGame.Presets` returns nothing (no "boss presets" line before the diagnostic edit); the Kolbot hint {12578,11043} is level cell (122,71), next to the arrival (125,75), and nav refuses the click ("would cross the trigger area of an unwanted exit"); every d2d level map is whole, so `Explore` has nothing to reveal and the run ends in 6 to 10 s with "boss not found in the arena".
3. S: the old preset row 836@12534,11040 and her stand (12542,11036) came from a live game (comment in `internal/boss/act1.go`); the frame is the same absolute one the hint uses (origin 12456,10972, derived from the hint's cell).

## Fix (data in the script, not a knob)

`Anchor.Spawns` (new, `internal/boss/script.go`): after the preset and the hint fail, `locate` walks to each known spawn before exploring. The Countess lists (12534,11040) and (12542,11036). Independent of `quests.layoutRead` and of the K1 d2d: knob-off behaviour only gains this extra walk in a path that ended in "not found" before. Test: `TestCountessWalksToKnownSpawnWhenPresetAndHintFail` (Paladin and Sorceress arrival), fails 2 of 2 without the fix.

## Live result: NOT green (U)

Job 20261120-180344-84ae (stock lab d2d, 3 runs): all BLOCKED "quest record 5 not observed"; the Paladin log shows the quest chunk held by the potion guard ("tower guard (3 healing potions < 4)") after the Black Marsh walk, the Sorceress never started the chunk (mana guard). The same seed gave different outcomes between runs (record 0x0014 reached once, guard held another time), so the run is not deterministic yet and the fix has NOT been seen to kill her. Next: a setup that survives the Black Marsh (more potions that the guard counts, or start at the tower), then the determinism pair.
