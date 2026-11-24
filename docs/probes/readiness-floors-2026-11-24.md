# Readiness floors from game data, not from the SME ballpark (2026-11-24)

Author: toolstrade43 (internal agent). Read-only live (container logs of the sandbox, no token, no save read). Nothing here is applied: the numbers are proposals for the user's go.

Tags: V verified (how), S secondary, U unverified.

## 1. Area and monster levels from our own tables (V)
`docs/research/areamon-table.csv` (12 lab seeds per area, Normal) and `danger.Table` (built from the install, run `D2_MPQ_DIR=<install> go test ./internal/power -run TestFloorsReport -v`): Barracks mlvl 9 (units 72-143), Outer Cloister 9, Jail 1-3 10, Inner Cloister 10, Cathedral 11, Catacombs 1-4 11 / 11 / 12 / 12, Tower Cellar 1-5 7 (8 mean), Dark Wood 5, Black Marsh 6. Andariel: level 12, life 1024, defence 60, damage 45.6, to-hit 169 (`danger.Boss`). The Smith: level 13, life 404, `extra_strong`. The third-party chart the user quoted (Barracks 9 ... Catacombs 11/11/12/12, Dark Wood 5, Black Marsh 6, Cellar 7) matches our table exactly, so the monsters are NOT alvl 20: a clvl 20 hero fights level 9 monsters, and the SME floors (clvl 24, life 400, Andariel 17/20) are not "area level" numbers.

## 2. What the live hero wears (V, `worn at zone entry` / `worn set changed` lines, 30 h)
Slots as the log prints them (`slot:value/class`): 1 head, 3 torso, 4 right hand, 5 left hand, 8 belt, 9 boots, 10 gloves.
- 2026-11-23 05:57Z: head, torso, weapon, shield, belt, boots, gloves (starter-class items, the values 1 to 13 read like near-broken durability, U).
- Then, for most of the day: nothing (09:58Z), a weapon only (12:53Z, 22:01Z), head + weapon + shield (22:03Z), head + weapon + shield + gloves (01:14Z). No body armour at all (`gear risk: medium (no body armor equipped)`), no rings, amulet, boots, belt.
- 2026-11-24 13:32Z the Wear set landed: torso added, head/torso/weapon/shield/gloves worn. The rare Small Shield, Heavy Gloves, Boots, Skull Cap and the rare Scepter stayed in the vault ("unidentified (back in the vault)"), so the weapon is still the starting one. The minute line has `armor:"?"` (not read): no AC or resist number is logged, so AC is an estimate (U): about 25 to 45 for the starter pieces, 100+ for the plan's gear class.

## 3. What actually hurts in the Barracks (V, minute and fight lines, 11-23 06:00Z to 11-24 02:44Z, clvl 19 to 20, max life 279)
| area | minutes | life lost/min | hits taken/min | life per hit |
|---|---|---|---|---|
| Dark Wood (5) | 358 | 14 | 4.2 | about 3.4 |
| Black Marsh (6) | 183 | 13 | 3.7 | about 3.6 |
| Outer Cloister (27) | 63 | 58 | 9.5 | about 6 |
| Barracks (28) | 66 | 91 | 13.9 | about 6.5 |
- Per hit the damage is what the table says: Barracks monsters hit for 5.4 to 8.1 (`corruptrogue` 5.5, `goatman` 8.1, `sk_archer` 5.4 ranged with to-hit 197). Level 9 monsters are weak; the observed 6.3 life per hit equals that (V). It is the NUMBER of hits: 3 times the field (13.9 against 4.2 per minute), because the Barracks holds 100 units and a fight there has 13.5 hostiles near on average (max 58; 226 of 375 fights ended in `error`, 99 in `timeout`, 6 in `chicken`; `combat fight area=28` lines). Killed by: Black Rogue (the archers) with 18 to 24 hostiles near, twice at "3 healing potions < 4" in the belt.
- The hero's own kill speed is the multiplier: the survival ledger holds `k_hero` at its clamp 3.00 (every unvisited zone showed k 3.00 on the live `zone choice` lines): the hero kills about 3 times slower than the model assumes, so each fight lasts three times longer and takes three times the hits (S, from the ledger definition: kill-time ratio).

## 4. Does the missing gear explain it? (V for the comparison, U for the cause)
Split the Barracks/Outer Cloister minutes by the worn set at the time: 0 to 1 slots worn: 74 life/min, 11.7 hits/min (84 minutes); 5 or more slots worn: 77 life/min, 11.7 hits/min (45 minutes). Same hits, same damage: AC did not change the rate. The "5+ slots" set was the starter set with near-zero durability and the SAME weapon throughout, so this does not prove AC is useless, but it says the gear that was missing (armour) is not what multiplies the damage. The model's own AC sensitivity agrees: the Barracks bad pack margin at clvl 20 is 3.8 with the as-played defence (30) and 9.0 with a plan-class set (defence 140, block 30, resist 25): AC is worth about a factor 2, k_hero is a factor 3. The weapon (offence) is the open lever: a Short Sword kills in 3 times the time. Answer to (5): worn versus unworn armour does not explain the death rate in the logs; the weak offence (and the 14 to 24 hostiles of the Barracks crowd) does, U until a run with the vault weapon.

## 5. Derived floors (S: the existing power model, `internal/power/floors_report_test.go`, balanced dial enter 1.5)
The lowest clvl at which the bad-pack low margin (zones) or the boss margin (Andariel) divided by k reaches `enter` 1.5. The hero inputs are a Paladin model (life 279 at clvl 20, +11 per level; defence 30 as played or 140 for the plan's class; damage from the clvl; merc alive); the caster rows reuse the same hero numbers with the caster archetype, so they only show the archetype effect (U).

| hero | gear | k | Barracks | Outer Cloister | Jail 1 | Andariel |
|---|---|---|---|---|---|---|
| bruiser | as played | 1.0 | 13 | 8 | 11 | 16 |
| bruiser | as played | 2.0 | 18 | 10 | 16 | 23 |
| bruiser | as played | 3.0 | 22 | 12 | 19 | 28 |
| bruiser | plan class | 1.0 | 8 | 8 | 8 | 11 |
| bruiser | plan class | 2.0 | 12 | 8 | 11 | 17 |
| bruiser | plan class | 3.0 | 15 | 8 | 13 | 21 |
| caster | as played | 3.0 | 21 | 12 | 14 | 26 |
(the full grid is the test's output.)
Reading: the live floors (Barracks 24 melee / 22 caster, Andariel 17 with 20 recommended) are what the model gives for a hero at k = 3: they are not 2x too high for THIS hero, they are right for a hero that kills three times slower than modelled. For a hero with k 1 to 1.5 and decent gear the same model says Barracks 8 to 16, Andariel 11 to 20. So the floor should depend on the measured k_hero (and a gear class), not be one number per class:
- proposal: keep the class floors as a CEILING and make the Barracks, Jail and Andariel gate use `floor(k)` read from this table (data, one row per k and gear class), recomputed by the test from game data. The ledger's k must be trusted only when it rests on enough kills (`shrinkKills`); with k below 1.5 and plan-class gear the Barracks opens from clvl 10 to 15.
- Test fixture to add on approval: the bruiser rows above as a table test in `internal/power` (needs the install; skipped without it), plus a unit test that the gate floor is monotone in k.

## 6. Findings that need a follow-up (U)
- `danger` rows for Catacombs 1 to 4 are dominated by monster `bighead` (ranged, damage 99 to 100, dps 43 to 46) and make every Catacombs zone "too hard" at any level up to 30. A ranged damage of 99 at monster level 11 is not plausible for Normal; it looks like a table artefact (a missile or an area spell counted as one hit). Barracks `skmage_fire2` is the opposite: damage 0.0 (a spell not counted). Both bias the zone words; check the monster rows against monstats before the floors are applied.
- The hero k comes from kill time, but the Paladin's observed damage per kill is half of the hero k suggests (Dark Wood k 1.5 after the zone term), see `docs/probes/zone-pooledk.md`.
