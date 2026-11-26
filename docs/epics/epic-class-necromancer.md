# epic-class-necromancer: Necromancer playable through Normal

> status: OPEN epic, written 2026-11-25. Tags V/S/U as in AGENTS.md. Lab setups are CHEATED (stand-in level, gear): mechanics only.

## What exists (V, files)

- Builds (internal/plandata/): `build-necromancer_summoner_normal30.json`, `build-necromancer_leveling_normal30.json`, `build-necromancer_summoner_guide.json`, `build-necromancer_poison_nova_sme.json`
- Unlock table: `internal/plandata/unlocks-necromancer.json`
- Research and drill docs: docs/classes/necromancer.md, docs/probes/necromancer-act1-drill-2026-11-22.md, docs/probes/necromancer-readiness-2026-11-22.md
- Drill rows: Act 1 stages run in the lab (Necromancer 11): necro.act1-andariel GREEN, Smith, caster belt and loop rows GREEN as Necromancer (matrix note 70); no Act 2 to 5 rows

## Gaps against the acceptance criteria

- Criteria are all unticked: data existing is not the same as a drilled, uncheated, live-proven class. Tick a box only with a named run or test.
- Real current level of the live hero and of any live run of this class: not recorded in this repo (U); the coordinator fills it from the live notes.
- Blockers: necro.act1-tools RED at Tools of the Trade word 0x204e; necro.tools-paladin-control BLOCKED by its seeding; Smith/plan_overlap rows RED unless run with -class Necromancer

## Links

- docs/classes/ranking.md (class order), ADR 0020 (class order and agency), docs/probes/quest-drill-matrix-2026-11-23.md

## Priority (user decision 2026-11-26)

- 2026-11-26 priority: 1, first in the class order (user decision). Source: user class order 2026-11-26. No criteria ticked by this line.

## Audit of 2026-11-26 (card necroaudit; details docs/probes/necro-readiness-2026-11-26.md)

- Still no criterion ticked: no earned (uncheated, walked) Act 1 run exists. Lab stage rows are GREEN as Necromancer except the Tools of the Trade reward (backend limit).
- Fixed: the default table `necromancer_summoner_normal30` had no `combat` key, so it fought with the class default layout that has no Corpse Explosion key (S, code reading; not yet seen in a live log: U).
- Blockers: Tools of the Trade reward word 0x204e (fork), no earned Act 1 run, no `summon-missing` invariant, minion caps and curse durations are guesses, Weaken/Decrepify/Lower Resist/Fire Golem have no key (F8 is the last).
- Cards: necrosummoninv, necrocorpse, necrofloors, necromercrogue, necrocaps, necrokeys, hotkeydrift.
