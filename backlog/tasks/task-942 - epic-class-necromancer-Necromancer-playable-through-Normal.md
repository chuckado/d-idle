---
id: TASK-942
title: 'epic-class-necromancer: Necromancer playable through Normal'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:03:00Z'
labels:
  - bot
  - feature
  - T4
dependencies: []
priority: medium
ordinal: 443000
card: "epic-class-necromancer"
owns:
  - "docs/epics/epic-class-necromancer.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Necromancer. Builds in internal/plandata: build-necromancer_summoner_normal30.json, build-necromancer_leveling_normal30.json, build-necromancer_summoner_guide.json, build-necromancer_poison_nova_sme.json; unlock table unlocks-necromancer.json. Docs: docs/classes/necromancer.md, docs/probes/necromancer-act1-drill-2026-11-22.md, docs/probes/necromancer-readiness-2026-11-22.md. What exists today: Act 1 stages run in the lab (Necromancer 11): necro.act1-andariel GREEN, Smith, caster belt and loop rows GREEN as Necromancer (matrix note 70); no Act 2 to 5 rows. Blockers: necro.act1-tools RED at Tools of the Trade word 0x204e; necro.tools-paladin-control BLOCKED by its seeding; Smith/plan_overlap rows RED unless run with -class Necromancer. Living status: docs/epics/epic-class-necromancer.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Necromancer: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Necromancer: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Necromancer: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Necromancer: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Necromancer: quest drills GREEN as Necromancer for every act (drill matrix rows with -class Necromancer; today: Act 1 stages run in the lab (Necromancer 11): necro.act1-andariel GREEN, Smith, caster belt and loop rows GREEN as Necromancer (matrix note 70); no Act 2 to 5 rows).
- [ ] #6 Necromancer: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Necromancer: known blockers listed and closed. Current: necro.act1-tools RED at Tools of the Trade word 0x204e; necro.tools-paladin-control BLOCKED by its seeding; Smith/plan_overlap rows RED unless run with -class Necromancer.
<!-- AC:END -->
