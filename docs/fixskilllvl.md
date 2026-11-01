# fixskilllvl: +skills in the cast level

> status: offline only, not verified live. Source of the rules: SKPASS trace Q8/B2 (docs/research/sk-passives-trace.md on branch agent/skpass, V read of game.exe FUN_00644180).

## What changed

- `internal/d2/skilllevel.go`: `SkillBonus(stats, states, class, id, ref)` sums the bonus; `Player.Base(id)` is base points; `Player.Effective(id, ref)` is base + bonus. `Skill.Effective()` (base + stat 107) is unchanged and stays the "no data" reading (used by `classOfSkills`, where a class-aware bonus would be circular).
- `internal/combat/skills.go`: `KnownSkills` uses `Player.Effective` for the known gate and the mana cost; `SkillRefOf(player, skill)` resolves own class (charclass), tab (`data.Skill.Page`, skilldesc SkillPage) and element byte (EType fire 1, ltng 2, cold 4).

## Rules

| Source | Rule | Tag |
|---|---|---|
| 127 all skills (sub 0) | every skill | V read |
| shrine (state 134) | +2 every skill | V read (state id 134 U live) |
| 83 class skills (sub class id) | own-class skills only | V read |
| 188 skill tab (sub class*8 + tab - 1) | own-class skill with a tab | V read |
| 97 oskill (sub skill id) | own class: min(v, 3); other class: v | V read |
| 126 elemental (sub element byte) | skills with that element | V read; byte mapping beyond Fire Bolt 1, Charged Bolt 2, Ice Bolt 4 U (poison, magic get 0; mastery passives not special-cased) |
| 107 single skill (sub skill id) | that skill | V read, V live (starting staff, Fire Bolt) |
| synergies | base points only | V read |

The sub index convention (key = sub<<16 | id in `d2.Stats`) is V live for 107 only; for the others it is U live.

## Backend exposure

- Native memory reader: `UnitStats` returns the full effective stat vector, so every id above is readable in principle (V live that 107 is present). Whether 127/83/188/97/126 appear with these sub indexes: U until a live read.
- d2d control socket: `dump stats` gives base attributes, panel bonuses and resists only (internal/d2dsrc/player.go: `player.stats.effective` unknown, `player.skills.bonus` unknown); skill bonus stays 0 there and there is no state list. None of the stats is exposed. Needed: an effective stat vector (or a per-skill level verb) and the state list from the fork; not added here (no fork verbs in this card).

## Callers audited

- `KnownSkills` (changed). `classOfSkills` keeps `Skill.Effective()` on purpose. `cmd/d2bot/display.go:180` shows base + 107 (display only, unchanged).
- Synergies: no bot code reads another skill's points for damage; `Synergies()` is only a key list. The leveler counts base (`Skill.Level`): correct per the trace (B3). `necro.go:501` `AuraFrames(sp.Level)` passes the total level into a calc whose synergy term should read base: unchanged, not provably wrong for the current profiles; follow-up if aura synergies matter.

## Live recipe (not done)

1. Wear or give gear with +all skills / +class / +tab; read `dump stats` raw / the native stat vector and list ids 127, 83, 188, 97, 126 and their sub indexes.
2. SKPASS drills P1-P6; P4 needs an item-properties verb in the fork (not part of this card). Without it: shrine drill (state 134) and the starting staff for 107.
3. Compare the mana cost the game charges for Charged Bolt at base 5 with +3 Lightning tab (expected 1,664 256ths = level 8) against `KnownSkills` cost.

## Shared-file lines for the merger

- BACKLOG: fixskilllvl live proof (stat ids on a real vector; d2d exposure of bonus stats).
- CHANGELOG: +skills (all, class, tab, oskill, elemental, shrine) now raise the cast level the bot uses for mana cost and the known-skill gate; synergies keep base points.
