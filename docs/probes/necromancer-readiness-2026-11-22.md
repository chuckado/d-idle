# Necromancer readiness audit (card necro2, 2026-11-22)

> status: current; section 6 (update after the merges of the same day) supersedes sections 1 to 3 where they differ. Offline audit of the tree at origin/main; no live game was used (the run was outside the night window 23:00-07:00 EST and no builder slot was claimed). Tags: V verified (how said), S read in code or docs, U unverified.

Verdict: NOT ready to play. Most pieces exist as code but are `experimental`; the three gaps that block "ready" are (1) curse/golem/state rules are Go literals, not data (breaks AGENTS "Build flexibility"), (2) no `necro.*` scenario exists, so nothing is green, (3) no earned Act 1 run was ever done on a Necromancer. The earlier design note `docs/process/necromancer-support.md` (card necro1) is still accurate; this page grades it.

## 1. S1-S11 (from docs/process/necromancer-support.md section 4)

| id | item | state | evidence |
|---|---|---|---|
| S1 | live summon loop through a whole Normal act, writes 0 | MISSING | live seen only for single casts (docs/summon.md "Live results", V 2026-10-29); no act-length record; scenario absent (section 3) |
| S2 | minion caps from skills.txt | MISSING | `wantMinions` guesses level/12/4/8 (internal/combat/necro.go:170-180, comment says U) |
| S3 | curses beyond Amplify/Weaken/Lower Resist; state read-back | MISSING | `curseFor` knows three curses (necro.go:353-372); `curseStates` only Amplify and Weaken (necro.go:39). Iron Maiden, Life Tap, Decrepify, Terror, Dim Vision, Confuse, Attract: no cast rule. Upstream d2d tests #645 (Life Tap/Iron Maiden) show the engine side exists |
| S4 | layout keys beyond F1-F8 | MISSING | summoner layout F1-F8 only (internal/combat/classdata/profiles.json, necromancer/summoner); Weaken, Lower Resist, Fire Golem unbound so `usable` skips them (necro.go:187-193) |
| S5 | golem ladder naming and choice | PARTIAL | `golemLadder = {"Fire Golem","Clay Golem"}` (necro.go:47) vs catalog name `FireGolem`: Fire Golem is never matched (U live); a Go literal, should be profile data |
| S6 | corpse rules: Revive vs Raise vs Explosion | PARTIAL | `Corpses` (necro.go:102), `bestRevivable` and `tried` set (necro_corpse.go); no reservation, boss-corpse rule U |
| S7 | mana budget by plan | PARTIAL | only `armyMin = 3` skips the rest-before-pack (necro.go:155-168); no recast-cheapest-first rule; V live 71 to 14 mana at clvl 20 (docs/summon.md) |
| S8 | positioning, do not outrun the pack | MISSING | no code; d2d upstream #659 (minions walk back to the leader) is the engine half; bot side not written |
| S9 | pet damage in the power model | MISSING | planner viability is a lower bound (necro-support section 3); internal/agency/necrosummoner_test.go logs it |
| S10 | merc auras reach pets | N/A | upstream d2d deviation A3, not ours |
| S11 | survival rule when the army dies (boss) | MISSING | no code |

Also MISSING: the combat code counts minions from the snapshot (`CountMinions`, necro.go:72); the fork read verbs `debug minions` / `debug skillstate` / `debug combatlog` have a client (internal/d2dctl/combat_reads.go:198-216, V: unit test combat_reads_test.go) but no caller in internal/combat or internal/bot (S: grep). They would give exact census, TTL and curse state and remove the 30 s cast memory guess (`curseLifeSec`, necro.go:23).

## 2. Surfaces

| surface | state | evidence |
|---|---|---|
| class traits table | PARTIAL | the caster triple is still written out five times (docs/research/build-flexibility-audit.md rows 27, 28); necro is a caster in gear (internal/bot/geargate.go:50), merc policy (internal/merc/policy.go:98), resource (internal/resource/gear.go:37) |
| OOG character creation | DONE | `"necromancer": {301, 241}` (internal/launch/oog_table.go:169); scenario class index (internal/scenario/scenario.go:206) V live per its comment |
| fresh save | DONE | internal/saves/fresh.go:10 |
| viewer/setup class picker | PARTIAL | setup wizard lists necromancer (web/setup.js:49) and posts charClass (web/setup.js:208); the in-viewer "new hero" and build picker with auto-allocate and build lock (docs/design/autoalloc-and-currencies-2026-11-21.md, knobs `agency.autoAlloc`, `respec.*`, default 0/off) is not Necromancer-specific work and its UI state is U here |
| build plan tables | PARTIAL | `necromancer_leveling_normal30`, `necromancer_summoner_normal30`, `_endgame90`, `_guide` exist (internal/plandata/build-necromancer_*.json); NO Bone/Poison-Nova caster table; unlocks-necromancer.json present. Summoner late points land late (Weaken clvl 50, Fire Golem 74, Revive 75: necro-support section 3, V from expansion) |
| combat profile | PARTIAL | profiles `necromancer` (Teeth, Bone Spear from clvl 18) and `necromancer/summoner` (Teeth, Attack only) exist; fields for upkeep list, recast trigger, curse rules per target class, corpse rules: none (they are Go: necro.go) |
| town/economy | PARTIAL | caster mana-potion and merc-aura rules exist via the class triple; ManaNeed 2 is a class switch (resource/types.go:146-153); no Necromancer-specific live check (docs/live-checks.md classpacing row open) |
| items/loot rules | PARTIAL | class fallback test exists (internal/lootrules/store2b_test.go:98); necro-specific gear (wands, skull/head, Bone Armor items) U |
| quest path | PARTIAL | all seven classes pass the campaign test (internal/campaign/campaign_test.go:47, S, offline); no earned run |
| scenarios | MISSING | only `points.necro-summoner-lands-at-30` (tools/provingground/scenarios/points.json:105), PENDING |
| lab test character | MISSING | `Necki` is named in the private rules as a throwaway but docs/probes/test-characters.md has no Necromancer entry (S: grep); a new one must be created and logged on a hidden instance |

## 3. Plan (dependency order, one branch each)

a. necro3 class traits + build tables as data (caster/manaNeed traits; Bone/Poison-Nova caster table; check against skills.txt). b. necro4 profile fields (upkeep, curse rules by target class, corpse rules, golem ladder) in classdata, Go literals removed, unknown name errors naming it. c. necro5 engine cards behind knobs default 0: upkeep recast driven by `debug minions`, curse application via `debug skillstate`, do-not-outrun the pack. d. necro6 `necro.*` scenarios (upkeep through a pack fight, curse on a unique, revive after a minion dies, 10-minute mana budget; Paladin regression pair) green on a lab slot. e. necro7 viewer new-hero class and build picker per the autoalloc doc. f. necro8 earned Act 1 drill, Rogue Encampment to Andariel, verdict in docs/probes/.

## 4. Questions for the user (build data)

1. Summoner: skill order for Skeleton vs Mage vs Revive split before clvl 30 (maxroll route in `necromancer_summoner_normal30`: keep, or Clay Golem first)? 2. Bone/Poison-Nova caster: which guide is the source, and is Poison Nova the main attack from clvl 30 (Nightmare) or Bone Spear until Hell? 3. Curse rotation: Amplify on packs, Decrepify only on bosses, Lower Resist on immune-to-physical casters, Iron Maiden vs Life Tap never? 4. Corpse rule: reserve the best corpse for Revive, or always Raise first? 5. Stat points: Necromancer vitality-heavy or energy for Bone Armor/Nova?

## 5. Live checklist (not run)

On a hidden instance (i1..i5) inside the night window, or a builder slot: create a fresh Necromancer through the OOG tooling, log it in docs/probes/test-characters.md, run `levelerBuild necromancer_summoner_guide`, confirm the start line shows combat profile `necromancer/summoner`, play to clvl 6, check Raise Skeleton and Clay Golem casts, read `debug minions` once, `writes after start: 0`.

## 6. Update after the merges (2026-11-22, cards necro3, necro4, necro5, classtraits, necroscen)

Verdict: still NOT ready to play. Data and engine gaps of section 1 are closed or switchable; no earned Act 1 run exists (section 6.3).

### 6.1 S1 to S11 now
| id | state | what changed (S = read in code; V = verified, how said) |
|---|---|---|
| S1 | MISSING | no act-length run; builder scenarios (V, run 20261122-170600, CHEATED setup, 6 to 10 virtual minutes) are green but are not an act |
| S2 | MISSING | `wantMinions` still a guess; skills.txt not read for caps |
| S3 | PARTIAL | curse list is data now (profile `necro.curses`, state numbers per rule, Decrepify before Weaken on elites); curse application on a unique V in `necro.curse-unique` (Amplify Damage); read-back through `debug skillstate` not wired |
| S4 | PARTIAL | layout is data (profile `layout`); summoner stops at F8 so Decrepify and Weaken stay unbound (a bound key is needed, U how) |
| S5 | DONE (offline) | golem ladder is profile data (`necro.golems`), names resolve through the catalog |
| S6 | PARTIAL | Revive cast V in `necro.revive-after-death` at clvl 36; no corpse reservation rule |
| S7 | PARTIAL | `necro.mana-budget-10min` green (U threshold); no cheapest-first recast |
| S8, S9, S11 | MISSING | unchanged |
| S10 | N/A | unchanged |
| census | DONE behind a knob | `combat.minionRead` (default 0) counts from `debug minions`; needs the fork flag `--read-combat`; V that the read works in the builder runs, not V that it beats the snapshot census |
| class traits | DONE | `internal/d2/classtraits.json` (caster, manaNeed); five copies removed |
| build tables | DONE | `necromancer_summoner_sme` and `necromancer_poison_nova_sme` (opt-in by name), profile `necromancer/poisonnova` |
| scenarios | DONE | five in `tools/provingground/scenarios/necro.json` (four Necromancer, one Paladin pair), GREEN 5 on the lab d2d devp20 |

### 6.2 Known limits
- Goal format: a stat goal cannot start at a given level, so the SME's Strength 25 by clvl 11 to 24 is spread over five levels (no scenario needs it yet).
- Decrepify and Weaken are unbound at F8 (above).
- Lower Resist lands near clvl 50 for the caster because the tool adds the Weaken, Terror, Decrepify chain first.
- Teeth is not in the SME table: the army kills, the profile's attack skills are never cast by the summoner build.
- Revive needs clvl 36 in the setup (level 30 grants 29 skill points and stops before the clvl 30 Revive row).
- Still U: curse durations, Revive count and duration, Poison Nova numbers (SME tagged them UNVERIFIED; read skills.txt), `debug skillstate` skill names as `hero_casts_at_least` keys, whether the web builds picker lists the two builds (section 6.4).

### 6.3 Earned Act 1 drill
Run, BLOCKED (docs/probes/necromancer-act1-drill-2026-11-22.md): a fresh hero with zero setup writes reached clvl 14 in 400 virtual minutes with 0 deaths, completed the Den only, and never started Blood Raven. Top blockers: the Rogue Encampment to Cold Plains exit click refused as an unwanted trigger (nav), no driver that taps a whole campaign, slow leveling under the board's training loop. NOT ready to play.

### 6.4 Viewer class and build picker
The Go side of the picker was already generic: `allocWire.builds()` lists every leveler table by class, so a Necromancer hero is offered `necromancer_summoner_sme`, `necromancer_poison_nova_sme` and the older tables, a Paladin or Sorceress never sees them (V offline, test `TestAllocBuildListTwoClasses` with three classes). New: each entry carries the build catalogue's name and identity as `Note` (`internal/plandata/build-catalog.json` owns the tables). `agency.autoAlloc` and the respec knobs stay default off; the web cards (E6) are the UI agent's. Not verified in a browser (U).
