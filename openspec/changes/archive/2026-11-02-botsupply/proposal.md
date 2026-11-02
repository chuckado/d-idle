## Why

The treasure and monster-stats traces (docs/research/treasure-trace.md B1-B5, docs/research/monster-stats-trace.md B1-B4) found that the bot's supply assumptions are wrong: 6 gold per kill (real Blood Moor 0.39), a flat 10-gold pile bar that skips every early pile, an XP-per-kill table 0.45x to 1.15x of the real means, wrong fallback monster levels, and a simulator whose monsters have 2.5 to 10 times the real life. Poverty mode and the kit gate (docs/research/levelspawn-trace.md, docs/soaks/gate1r2-20261101.md) depend on these.

## What Changes

- internal/data/supply.go (new): the per-kill supply table (plain 0.375 drops / 0.131 piles / 0.057 potions, champion 3, unique 5, superunique 6) and the level-aware gold model (pile = ilvl + rand(5 ilvl), mean 3.5 ilvl - 0.5); pile-share helpers.
- Gold pile bar: new knob `loot.goldPileShare` (default 50): `Looter.goldBar` lowers `goldMinStack` to the pile size half of the area's piles reach. ON by default: the trace proves the flat 10 skips all of Blood Moor. Knobs doc regenerated.
- internal/bot/recoveryxp.go: traced area XP table (areaXP), fallback `normalMLvl` Den 1 / Cave 2 2.
- internal/data/db.go, monsters.go: `ScaledHP` fixed (NoRatio rows, level clamp), `MeanHP` added; `Monster.NoRatio` read from MonStats.
- internal/resource/sim.go: gold from the model (`MonLevel`), monster life `BloodMoorLife` / `LifeCycle` from real rows; stock Sorceress replay re-pinned (45 kills / 15 swings / 45 casts, was 22/28/66 on the old life).
- internal/data/loot.go: wrong `Picks` comment fixed. docs/botsupply.md: the supply note.

All class independent (no class names; the numbers describe monsters).

## Impact

Behaviour changes only where a trace proved the old value wrong: pile bar in low-level areas, recovery area ranking (Cave 1 > Cave 2 > Cold Plains), sim assumptions. Not verified live.

Overlap (declared): internal/resource/sim.go, sim_test.go, palprobe_test.go (archived skmeleewire), internal/env/tuning_loot.go, internal/loot/*, internal/data/db.go, monsters.go, loot.go, data_test.go, docs/knobs.md and internal/bot/recoveryxp.go are owned by archived cards or by no open card; this card makes small edits there and lists only the new files in owns.
