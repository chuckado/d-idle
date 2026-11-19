# freebuild (prototype, never merged)

Question: if the player spends every skill point freely, can the bot derive how to play from the character's ACTUAL
skills, and say so honestly when it cannot? Offline, no game ran. Run: `go test -p 1 -v ./internal/proto/freebuild/`.

## Files
- `capabilities.json` (schema `freebuild.cap/1`): per skill `role`, `delivery`, `tier`, `element`, `synergies`, `note`.
- `caps.go` loader, `Lookup` (missing skill -> row derived from power `skillKinds`, `experimental`; in neither -> not
  found, treated as unsupported or an error for unknown names). `derive.go` `Derive`. `validate.go` `Validate`,
  `CanAdd`, `Preview`, `Diff`. Fixture prerequisites: `fixture_test.go`.
- Reads: `internal/power` (`SkillData.Estimate`, `Profiles.Resolve`), `internal/combat` (`LookupProfile`,
  `AttackNames`: the bot's current choice), `internal/leveler.ParseBuild`, `internal/plandata`, `data.Skill` (struct only).

## What it proves (V = ran in the test)
- V: Derive picks a main attack by argmax of the emulator-derived dps from hard points; every choice has a `Why`;
  warnings `no_attack tourist single_element unsupported_points weapon_unread mana_starved prereq_only` (plus
  `synergy_unknown`) fire on the off-plan allocations of `TestOffPlan` (all-Fire-Ball sorceress, 12 x 1 point paladin,
  Whirlwind-maxed barbarian falls back to Bash/Double Swing and names Whirlwind, necro summoner is archetype summoner with
  main = Raise Skeleton and no `no_attack`, zero-attack sorceress gets `no_attack`).
- V: unknown class or skill is an error naming it; Validate/CanAdd/Preview behave (reqlevel, prereq, 20 cap, budget).
- V (read from profiles.json, not a test): `power.Profiles.SkillKinds` has no Fire Bolt, Ice Bolt, Fire Ball, so `Resolve`
  alone fails for the commonest Sorceress attacks. Derive merges kinds from the capability table first. Production should
  put those kinds into power's profiles.json.

## What it does not prove
- U: no game ran; tiers are read from comments in `internal/combat` and `classdata`, not re-tested.
- U: skilldamage.json limits carry over: masteries are 0 (`mastery_none`), multi-strike skills count one strike
  (Zeal, Frenzy under-estimated; Sacrifice looks 2.5x Zeal), synergies are uniform points (mean over the named
  synergy skills), melee needs a weapon (test env uses a nominal 20-average weapon, U).
- U: `synergies` lists exist only for Charged Bolt, Lightning family, Nova, Blizzard, Zeal, Blessed Hammer (from
  memory, U). Everything else is `null` = not known: warning `synergy_unknown`, damage is a lower bound.
- U: the mana budget (`capabilities.json` `mana`: pool from base + Ene + clvl, 30 s fight plus 1.1 percent regen per
  second) is a guess, not the game's regen. `mana_starved` fires too often (Bash at clvl 6): do not trust the number.
- U: prerequisites in the fixture (34 rows) were written from memory, not read from skills.txt.
- Not modelled: life cost of Sacrifice (profile `lifePrice`), Blessed Hammer standing still, immunities beyond the
  element split, merc, gear, FCR/IAS (default 0, flagged `fcr_default`).

## Capability tiers (table rows; `TestCapabilityTable` logs them, V)
| class | supported | experimental | unsupported |
|---|---|---|---|
| sorceress (30 of 30 skills) | 9 | 12 | 9 |
| paladin (30 of 30) | 3 | 23 | 4 |
| barbarian (used skills only) | 4 | 8 | 4 |
| necromancer (used skills only) | 4 | 7 | 2 |

Evidence: supported = Fire Bolt (`classdata.go` header: V live), Zeal and Holy Fire/Might aura (same header,
`docs/classes/smoke-paladin.md`, `aura.go`), Raise Skeleton, Clay Golem, Bone Armor (`necro.go`: V live 2026-10-29).
Static Field, Frozen Armor, Charged Bolt, Ice Bolt are `supported` on the brief's say-so, not re-verified here.
Unsupported = no `skills.json` row or no profile band, or a profile avoid list: Whirlwind (hold row, no band), Charge,
Vengeance, Fist of the Heavens, Hydra (pet recognised in `friendly.go`, no cast loop), Fire Wall, Blaze, Inferno,
Frozen Orb, Telekinesis, Lightning and Chain Lightning (both on every Sorceress `avoid` list, broken on the d2d build).

## Agreement table (bot's first known band attack vs Derive, 23 of 30 agree, V)
Weapon = nominal avg 20. Each table's `combat` key picks the profile (cold, lightning, class default).
| table | clvl 6 | 12 | 18 | 24 | 30 |
|---|---|---|---|---|---|
| sorceress_normal30 | agree Charged Bolt | agree Charged Bolt | agree Ice Blast | agree Ice Blast | agree Ice Blast |
| fallback-lightning-v2 | agree Charged Bolt | agree | agree | agree | agree (after Chain Lightning marked unsupported; before it differed) |
| paladin-zealot-lite | agree Sacrifice | agree Sacrifice | DIFFER bot Zeal, derive Sacrifice | DIFFER bot Hammer, derive Sacrifice | agree Blessed Hammer |
| paladin_hammerdin_endgame90 | agree Attack | agree Attack | agree Hammer | agree | agree |
| barbarian_melee_normal30 | agree Bash | agree Bash | agree Bash | agree Frenzy | agree Frenzy |
| necromancer_summoner_normal30 | DIFFER bot Teeth, derive Attack | same | same | same | same |

Why they differ (the finding): (1) Paladin: Sacrifice 93 dps vs Zeal 37 and Hammer 47. Zeal is counted as one strike,
Hammer has no synergy points, and the life price of Sacrifice is ignored, so the argmax is not trustworthy for
multi-strike or life-cost skills. The bot's band order encodes knowledge the estimator lacks. (2) Necromancer: the plain
swing with a nominal weapon (26 dps) beats Teeth (5 dps); the bot prefers Teeth because a summoner should not melee.
Derive still reports archetype `summoner`; its Main is the pets, `Attack` is only the filler. Differences are
estimator gaps, not proof the bot is right.

## Other findings
- Tourist fires for the stock plan tables at clvl 12 to 30 (barbarian, paladin hammerdin early): the tables put points in
  passives and prerequisites. The threshold (35 percent) needs tuning before it gates anything.
- `prereq_only` fires on most tables for the same reason; it is advice, not an error.
- `sorceress_normal30` buys Telekinesis: 1 point Unplayed.

## What a production version needs
- Inputs the bot must read: hard points per skill (not +skills), weapon class/speed/damage, FCR, IAS, energy and
  level, quest skill points; the real skills.txt rows (replace the fixture; `data.DB`).
- Plug in: `classdata` band selection (replace first-known-in-band by Derive's ranked candidates, keep band lists as
  tie-break and as the unsupported filter), `power.SpecFromBuild` (build the spec from the live allocation, not a plan
  table), the leveler (`CanAdd`/`Preview` to refuse or warn on off-plan points).
- Data work: kinds for all attack skills in power `profiles.json`; synergy names for every attack skill; multi-strike
  and mastery damage in skilldamage.json; a verified mana model; per-skill life cost; tiers re-checked live per skill.
