# Hero stats in the d2d dump (card herostats, TASK-493)

> status: current. Fork branch `sandbox-herostats` (private fork, tip `02ceea54`, pushed to the private forge, not upstream: a lab/dump feature). Tags: V verified (how), S read in source, U unverified. Not deployed and not run against a live game.

## Finding (S: fork source read)

The sandbox lineage `didle/rebase-20261114b` has `stats.defense` but NOT `stats.gear` / `stats.weapons`: those came with branch `lab2-stats` (docs/forkdbg.md), which sat on the old 52362d52 base and was never carried into the rebase (docs/fork-rebase-2026-11-13.md inventory). So on the sandbox `zone entry` logs `def` only, as diagnosed in docs/survival/wiring-2026-11-15.md.

## What the branch adds (fork `sandbox-herostats`, from `didle/rebase-20261114b`)

1. `lab2-stats` re-applied (cherry-pick, clean): `stats.gear` (sums of stats 16..25, 83, 107, 127, 188) and `stats.weapons` (base damage per worn weapon). `worn_props` now delegates to upstream `gear_props` (panel's and fight's own filter, so broken items and short charms are out).
2. New `stats.hero` (`components/game/character/hero_stats.hpp`): the fight code's own model, `rules::make_fighter`, fed with the same gear sum as `Fight::player_fighter` but with NO passive skills, auras or self states:
   - `defense` (the panel's value as of the last panel update), `block` (percent, 0 without a shield: (shield block + class BlockFactor + gear stat 20) x (dex - 15) / (clvl x 2), cap 75), `shield` (bool),
   - `ar` / `ar_base` / `ar_pct` (attack rating: (dex - 7) x 5 + class ToHitFactor + gear stat 19, then x (1 + gear to-hit percent)), `dmgpct` (stat 25 summed over the gear),
   - `hands`: per worn weapon `{slot (4 right, 5 left), code, min, max, pct}`: min and max are the base range with the weapon's own enhanced damage and flat damage, before the percentage; `pct` is the total percentage (stat 25 plus the weapon's strength and dexterity bonus). `bare`: fists. `null` when the class is unknown (never a default class).

## Go side (this repo, additive)

`internal/d2dsrc/gearstats.go` keeps reading `gear` and `weapons` (older forks unchanged) and now overlays `hero`: stat 20 becomes the whole block chance, stats 21 / 22 the right-hand `min` / `max` (enhanced damage included). The melee model still applies stat 25 and the strength bonus itself, so `pct` is not read. `power.HeroFromPlayer` and the `zone entry` keys (`def block dmg dmgpct ar`) need no change.

## Verified and not

| Claim | Tag |
|---|---|
| The sandbox branch lacks `gear` / `weapons` (file list of the branch tip) | V: `git ls-tree` on `didle/rebase-20261114b`, no gear_stats.hpp |
| `test_hero_stats`: Paladin-like worked example (block 61, ar 275, sword 4-19, pct 80) and no-shield, nothing-worn, unknown-class cases | V as arithmetic of the fork's `make_fighter` formulas (built and run on the NAS builder, `OK`, rc 0; fork commit 02ceea54); NOT V against game.exe or the character panel |
| `apps/d2d/main.cpp` (the dump verb) compiles with the change | V (ninja step on the builder); the full `d2d` link was not finished: two unrelated translation units were killed for memory while another job ran, so no binary was produced or deployed |
| `test_gear_stats` (sums per stat and param) still passes after switching to `gear_props` | V (same run) |
| `defense` equals the panel's Defence number | U (docs/survival/power.md checklist step 1) |
| Item bonuses to strength and dexterity feed block, attack rating and the strength damage bonus | U: `make_fighter` reads the character's base stats only, as the fork's fight code does; gear with +dexterity would make `block` and `ar` low |
| Passive skills (masteries, Holy Shield, Defiance auras, Iron Skin) | not included by design |

## Live check for a later window (sandbox, one hidden slot)

1. Equip a weapon, shield and body armour through the normal flow; `d2drill -probe 'dump stats'` and read `stats.hero`.
2. Compare `defense`, `block` (the panel's block chance), `ar` and the damage line with the character panel (`tools/shot.sh OUT.png iN`). Record which differ (expected: gear +dex, skills).
3. Bot: `grep 'zone entry' bot.log` shows `def block dmg dmgpct ar`.
