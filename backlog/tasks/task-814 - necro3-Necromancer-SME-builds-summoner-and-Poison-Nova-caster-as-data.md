---
id: TASK-814
title: 'necro3: Necromancer SME builds (summoner and Poison Nova caster) as data'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T19:35:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 373000
card: "necro3"
owns:
  - "internal/buildfile/testdata/necromancer-summoner-sme.json"
  - "internal/buildfile/testdata/necromancer-poison-nova-sme.json"
  - "internal/buildfile/necrosme_test.go"
  - "internal/plandata/build-necromancer_summoner_sme.json"
  - "internal/plandata/build-necromancer_poison_nova_sme.json"
  - "docs/plan/data/build-necromancer_summoner_sme.json"
  - "docs/plan/data/build-necromancer_poison_nova_sme.json"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/necro_data_test.go"
  - "internal/leveler/necrosummoner_test.go"
exclusive: []
contracts: []
result: "Tables necromancer_summoner_sme and necromancer_poison_nova_sme (opt-in) from didle.build files; poisonnova profile; Decrepify before Weaken on elites in the summoner; offline only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: SME answers merged (#1110). Done when: both builds exist as data. NOTE: read skills.txt for curse durations, Revive count and Poison Nova numbers; the SME tagged them UNVERIFIED. Not done here: the class traits table (caster triple), stat goals that start at a level, a layout beyond F8 (Decrepify and Weaken are unbound).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Two opt-in tables from build files per the SME answers, validated under skills.txt rules; defaults unchanged
- [x] #2 Poisonnova combat profile with its own curse data; tests on both builds
- [x] #3 Card note: read skills.txt for curse, Revive and Poison Nova numbers (SME UNVERIFIED items) before relying on the point split
<!-- AC:END -->
