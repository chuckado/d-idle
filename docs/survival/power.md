# Power estimate and margin, shadow (survival slice S-B)

> status: current (card survB, TASK-449; skill damage wired by card powerwire, TASK-508, section "Skill damage wiring"). Implements design-2026-11-14.md sections 4, 5, 6, 8 and 10 as far as slice S-B needs: archetypes, power terms with quality flags, `Margin`, the `zone shadow:` log line, `levelsim -shadow`. Log only, no behaviour change. Tags: V verified (how), S secondary (source read, not run), I inferred, U unverified. Nothing here was run against a live game (the user plays on the lab host; every live step is in the checklist below).

## What shipped

- `internal/power` (leaf package; data in `profiles.json`): five archetypes (`bruiser`, `ranged_physical`, `caster`, `summoner`, `hybrid_tank_caster`) with the design 6.2 fields; the resolver `Resolve(BuildSpec)` (rule of design 6.3, no class names; an unknown class, skill or a build with no attack is an error naming it); `SpecFromBuild` reads a plandata build file and the class combat profile (a build file may name `"archetype"`, which wins); the three dials (Cautious 2.0/1.6, Balanced 1.5/1.2 default, Bold 1.2/1.0: user answer 1, numbers U); `Hero` with a quality per term (`measured`, `derived`, `default`); `Margin(hero, archetype, pack)` returning best and low estimates with the terms; `Model.Zone` for typical, bad and elite packs of a `danger.Zone`; `ZoneMargin.Word()` (safe, risky, too hard: the only form the UI may show, user answer 2) and `ShadowLine`.
- `cmd/d2bot/powerwire.go` + one call in `internal/bot/goal.go`: after every `planner` line the log gets `zone shadow: area= name= diff= clvl= arch= dial= word= bad= low= typical= elite= need= stay= perKill= potions= weakest=`. Hook `bot.Deps.ZoneShadow`, wired in the d2d backend and the original-game path; no hook without game tables; errors warn once per reason. The build is the leveler's default pick (`leveler.SelectBuild`), so an operator's explicit `levelerBuild` is not seen (U).
- `tools/levelsim` `-shadow LOG -install DIR -arch ID [-diff normal] [-dial balanced]`: replays `zone entry`, `safety decision`, `character died`, merc-died and `combat potion` lines through the estimate and applies the design 9 pass rule (80 percent bad stays under the enter line, 80 percent quiet stays of 10 minutes or more above it; fewer than 3 cases of a kind prints n/a).
- `internal/d2dsrc`: `stats.defense` of the d2d dump now lands in stat 31 and the gear sum of stat 20 (block) in stat 20 (new, read-only; `d2.StatArmorClass`).

## What is readable today (the env card)

| Term | Original game (memory) | d2d backend (dump) | Used as |
|---|---|---|---|
| Max life, max mana | V `d2.Player` (24.8 fixed, read live many runs) | V (`player.life`, `dump_head_late.json`) | measured |
| Resistances (fire, cold, light, poison) | stat ids U (`internal/d2/stats.go`) | V panel values; before or after the difficulty penalty is U (TASK-216) | derived |
| Defence | stat 31 in the player stat vector: S (stat list is read for all stats; nothing read id 31 before) | `stats.defense` exists in `dump_head_late.json` (S: fixture, value semantics, Dexterity and buff share U); mapped to stat 31 now | measured when present, else default (`defence_default` flag) |
| Block chance | stat 20 gear sum only: S | sandbox dump had no `gear` (docs/survival/herostats-2026-11-15.md); fork branch `sandbox-herostats` adds `gear` and `stats.hero.block` (shield base + class BlockFactor + gear + Dexterity, no skills; V as test arithmetic, U against the panel) | derived |
| Weapon damage, to-hit, damage percent | stats 21/22/25/19: S (`internal/combat/melee.go`, drill R1 pending) | V live for an empty kit; with gear U (`docs/forkdbg.md`); mapped in `gearStats` since lab2-stats | derived; weapon skills need it (no weapon read: no estimate, `offence_weapon_unread`) |
| Skill damage, mana and cast / attack period of the main attack | not read by the game: data `internal/power/skilldamage.json` (V run: game.exe's own functions in the emulator, decomp29), skill by NAME from the build, level from the build rows by clvl | same | derived (`OffenceDPS`, card powerwire); synergy points not known: 0 (`offence_synergy_unknown`); masteries and gear spell damage 0 (`offence_mastery_none`) |
| Attack speed (IAS), FCR | stats 93 / 105 when the player's stat vector has them (S: ItemStatCost ids, not read live yet) | not in the dump | read: term known; else 0 flagged (`offence_ias_default`, `offence_fcr_default`); period from the rate table, not a constant |
| Block frames | not read | not in the dump | not used by the margin |
| Healing potions | belt and pack read (V, `Kit`) | V | derived (heal per tier is the constant `potionHeal`, U) |
| Merc alive, hired, level | V (`MercSheet`) | V (mercdump) | derived; merc dps and hold share are constants (`mercDpsPerLevel`, `mercHold`, U): the hireling table is not read |
| Regeneration, auras (sustain) | not read | not read | 0 |
| Measured kill speed (self-measurement, design 4) | needs the fight log aggregation: not wired | same | `Hero.MeasuredDPS`, 0 today |

A term that is not read keeps quality `default`, its field 0 or a named constant, and the low estimate discounts it (x0.6 default, x0.85 derived: design 4, U). The log line names the weakest term (`weakest=defence_default`). Nothing is invented: every constant is in `profiles.json` with a tag.

The reads sit behind `power.Reader` (`Hero() (Hero, error)`) with `power.Fake`; `HeroFromPlayer` maps a `d2.Player`, `MercSheet` and potion count, and is what the bot hook uses.

## Live verification checklist (a later lab window; do not run while the user plays)

1. One hidden slot, a fresh Paladin with a shield and body armour equipped through the normal town flow: `d2drill -probe 'dump stats'` and read `stats.defense`, `stats.gear` (stat 20 and 21/22). Compare `defense` with the Defence number of the character panel (`tools/shot.sh OUT.png iN`). Result: V or the offset to the panel (then fix `playerStats` and the table above).
2. Same slot, bot with the planner on for one step: `grep 'zone shadow:' bot.log` shows one line per `planner` line; `weakest=` is `none` or names a term; the `word` for Cold Plains at clvl 5 is safe.
3. Original game (memory) path: read stat 31 and stat 20 from the player list in the stat dump (`d2qmem` or the existing stat dump) with and without a shield; record in `docs/mem.md`.
4. Replay: `go run ./tools/levelsim/cmd/levelsim -shadow swap7.log -install <game dir> -arch bruiser` (and h4, hero14) on the host that holds the logs.

## Pass rule of design section 9 on the recorded runs

The raw logs of swap7, h4 and hero14 are not in the repository (`docs/levelsim.md`: the raw logs are not committed; `git ls-files` has no such file), so the rule was NOT run on them here. What is verified: the replay and the rule on a synthetic log for two archetypes (bruiser, caster), including a FAIL case (`tools/levelsim/shadow_test.go`). The real run is step 4 above. The log's `zone entry` line carries no defence or damage, so a replay keeps `defence_default` and `offence_default` on every row; if the 80 percent rule fails because of that, the fix is to add defence and the weapon to that line (a one-line bot change, S-C) before S-C starts.

## Tests

`go test ./internal/power ./tools/levelsim ./internal/d2dsrc ./cmd/d2bot -run 'Power|Shadow|Margin|Archetype|Synthetic|Build|Hero|Pack|Words|Unread|Profiles|Dead|Snapshot|Gear'`: monotonicity per archetype (life, potions, mana, defence, block, resist, sustain, damage, merc up never lowers; pack, damage, life, level up never raises) for the Paladin (bruiser) and the Sorceress (caster); quadratic pack growth; a golden archetype per shipped build at clvl 12 and 30; a synthetic class gets the nearest archetype; unknown skill, missing attack and missing class are errors; every class's leveler build resolves in the bot wiring.

## What blocks S-C (ledger and controller)

1. Defence and block with gear are unverified live (checklist 1 and 3); until then every margin carries `defence_default` or `block derived` and the low estimate is the one that decides.
2. Spell and skill damage is not readable (no Skills.txt damage evaluation): casters run on `defaultDpsPerClvl` unless `MeasuredDPS` exists. S-C needs the self-measurement (kill seconds per monster type from the `combat fight` lines) wired into `Hero.MeasuredDPS`; that also gives `k_hero`.
3. The pass rule has not run on real logs (not in the repo).
4. The merc's damage and hold share are constants (hireling table not read).
5. The shadow hook sees the leveler's default build, not an explicit `levelerBuild`.

## Skill damage wiring (card powerwire, TASK-508; shadow only, log and API)

What changed: `internal/power/skilldamage.go` loads `skilldamage.json` (schema `power.skilldamage/1`, additive: unknown fields are ignored, another major version is refused; an unknown skill, class or weapon class is an error naming it). `SkillData.Estimate(class, Attack)` returns an `OffenceDPS` (damage per use, seconds per use = the larger of the animation ticks and the skill delay, dps, mana per second, flags). The main attack comes from the build by NAME (`BuildSpec.Attack`, level = points the build rows spent by that clvl, `Model.Attach`); the class code is the first three letters of the hero's class (the frame table key), no class check. `Margin` uses it in this order: `MeasuredDPS`, the skill estimate, the weapon branch (swing from the rate data for the class, constant `swingSeconds` only when the class is unknown), the constant `defaultDpsPerClvl`. The mana burst limit takes the skill's own mana per second when known (`manaPerSecond` stays the fallback). The shadow line gains `dps=` and `src=` (skill, weapon, default, measured). Terms stay tagged: the data is V run (decomp29), the level and gear speeds are derived, a missing FCR/IAS/synergy/mastery is a flag, never a number (`offence_fcr_default`, `offence_ias_default`, `offence_weapon_class_default`, `offence_synergy_unknown`, `offence_mastery_none`, `offence_one_strike`, `offence_sequence_rate_u`, `offence_weapon_unread`, `offence_mana_default`, `offence_swing_default`).

Model terms, before and after (the constants stay in `profiles.json` as fallbacks, tagged U):

| Term | Before | After |
|---|---|---|
| Swing of a one-hand Paladin (weapon read) | 1.28 s constant | 0.60 s (15 ticks, class and weapon class from the rate data; IAS read lowers it) |
| Caster kill speed (no damage readable) | 2.5 x clvl dps | Fire Bolt level 10 alone: 38.5 dps at 13 ticks (0.52 s), from `elem` min/max; Blizzard level 20 is delay-bound at 1.8 s |
| Mana per second | 4.0 | the skill's mana / period (Fire Bolt 4.8) |
| Skill delay | not modelled | `delay_frames` (Blizzard 45 ticks, Meteor 30, ...) |
| Weapon skill without a read weapon | constant dps, silently | no estimate, `offence_weapon_unread`, the constant stays and is flagged |

Not modelled (flagged or listed): Zeal's strikes per use and Charged Bolt's bolts count as one (a lower bound, U3 of the research doc), the skill's to-hit bonus is ignored, no monster resist, no area or second target, synergy points of the build are not read (0, a lower bound), `attack_sq` rates are unverified (U1), IAS/FCR come from stats 93/105 when the vector has them (S, not read live).

### Replay before and after (V: `sh tools/levelsim/replay_compare.sh` on the worker host, balanced dial, the sandbox log of replay-2026-11-14.md)

The log has no skill levels and no weapon, so `levelsim` got `-build paladin-normal30` (skill and level by clvl) and, as a sensitivity input, an ASSUMED weapon (`-weapon`, `-ar`: U, not a reading). The Paladin's weapon skills need the weapon, so without it nothing changes: that is the honest result of the log.

| Run | pass rule (bad stays under the line) | k_hero at the end | planner picks the shadow agrees on | stay minutes safe / risky / too hard |
|---|---|---|---|---|
| before (this branch without `-build`, equals main) | FAIL, 6 of 14 (good 8 of 8) | 2.00 (the unread cap) | 5 of 45 | 330 / 36 / 2 |
| `-build paladin-normal30`, no weapon | same | 2.00 | 5 of 45 | 330 / 36 / 2 |
| plus assumed weapon 6-12, AR 150 | FAIL, 1 of 14 | 3.00 | 5 of 45 | 343 / 23 / 2 |
| plus assumed weapon 10-20, AR 200 | FAIL, 0 of 14 | 3.00 | 10 of 45 | 365 / 1 / 2 |

Reading: with a modelled weapon the estimate makes the hero kill faster, the bad packs look safer and the pass rule moves further from passing; k_hero 3.00 (the observed kill time is about three times the model's) says the real fights are slower than the skill period predicts (walking, hits that miss, monsters out of reach, the merc taking the kill; one-strike Zeal). n = 1 hero, one build, an assumed weapon: nothing was tuned to pass, no default was changed. The Sorceress is covered by unit tests only (no Sorceress log exists).

Next: put the weapon, AR, IAS and FCR into the `zone entry` line (a one-line bot change) so a replay needs no assumption; then the 3x kill-time gap is a measurement, not a sensitivity.

## Block and unblockable ranged attackers (card dangerfit2)

`hitOn` multiplies the hit chance by `1 - block` only for the blockable share of an attacker: `Mon.Unblockable` (0 to 1, from `danger.Monster.Unblockable`) removes that share from the block factor, since game.exe cannot block a missile with no physical damage (decomp26 Q10 to Q13, V). `TestUnblockableRanged` runs it for a block-60 Paladin (margin drops) and a Sorceress (unchanged). Zone margins for caster-heavy areas (Black Marsh, Dark Wood, Catacombs) fall for shield heroes; no live check.
