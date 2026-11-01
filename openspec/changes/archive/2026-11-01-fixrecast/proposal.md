## Why

The armor recast timer (`combat.buffRecastSec`, 150 s) is longer than Frozen Armor lasts at levels 1 and 2 (120 s and 132 s, Skills.txt auralencalc), so the armor is down about a fifth of every cycle and its defence and chill-on-hit are lost for that time (finding F2 of docs/research/sorc-skill-matrix.md). The buff state is readable (`Player.States`), so the bot can use the real duration and the real state instead.

## What Changes

- internal/data: `Skill.AuraLenCalc` and `AuraFrames(level)` (the leading `lnAB` term of auralencalc).
- internal/combat `buffDue`: Frozen Armor state 10 is read like Bone Armor; an armor with a known duration is recast `combat.buffMarginSec` (new, default 6 s) before it ends; `buffRecastSec` stays the upper bound and the fallback.
- docs/knobs.md regenerated (one new knob), docs/fixrecast.md with the live recipe.

## Impact

Owns the files in tasks.md. Offline only. No spec delta: no spec names the recast interval and no numeric gate changes.
