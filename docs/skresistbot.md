# SKRESISTBOT: the bot's monster-resistance model after the trace

Source of the rules: `docs/research/sk-resist-trace.md` (Q1 to Q10, V read / V run). Findings B1 to B4 of its section 6.

## What changed

| finding | change | before / after (from `internal/combat/skresist_test.go`) |
|---|---|---|
| B1 | `expFactor` floors the resistance at -100 (`minResist`, Q8), all six types | res -250: factor 3.5 before, 2.0 after; -101: 2.01 before, 2.0 after; -100 and above unchanged |
| B2 | `Target.Pierce` (per type, from the player's stat list, `PierceFromStats`) and `Target.EffRes`: pierce subtracted only while the monster's total is under 100 (Q4), then the -100 floor; `DamageFactor` and `ImmuneWorkaround` use it | cold 75 with pierce 20: factor 0.25 before, 0.45 after; at 100 or more (pierce 20 or 60) still 0 and `Immune` stays true; a 75-vs-40 element choice keeps the pierced element (0.45 >= 0.6 * 0.60) where it used to switch (0.25 < 0.36) |
| B3 | not changed | the trace tags it S (monster-mods-trace.md says game.exe skips every add once two immunities exist, counting physical and magic; the bot skips only an add that creates a third among magic and the four elements). Fallback only (no stat list). A test pins the current behaviour |
| B4 | `resistsOf` comment corrected (a monster has no difficulty penalty, Q10) | wording only |

Pierce ids: cold 335 is V (Q4); fire 333, lightning 334 and poison 336 follow the ItemStatCost order (S). The ids are stat-table constants like `resistStat`, not skill ids; the skill that grants them is never named.

## Not done

- The comment "curses, difficulty" above `Fighter.refineResists` and the 1/5 curse rule: `internal/combat/fighter.go` belongs to the open card armorcut (Overlap, declared), so that one comment line is left for the merger to fix after armorcut. No code in the repo predicts the effect of Lower Resist or an aura on an immune target (`curseFor` only compares current resists), so the 1/5 rule is not modelled (nothing to correct, no feature invented).
- B5 (d2d backend reads d2d's undivided curse resistances) is a d2d matter, not touched.
