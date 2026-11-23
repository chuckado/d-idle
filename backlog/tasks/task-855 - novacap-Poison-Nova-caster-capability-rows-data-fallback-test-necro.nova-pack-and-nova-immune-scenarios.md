---
id: TASK-855
title: 'novacap: Poison Nova caster capability rows (data), fallback test, necro.nova-pack and nova-immune scenarios'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:09:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 387000
card: "novacap"
owns:
  - "data/capabilities/capabilities.json"
  - "internal/combat/classdata/skills.json"
  - "internal/delivery/testdata/skills.golden.json"
  - "internal/combat/necro_data_test.go"
  - "tools/provingground/scenarios/necro.json"
exclusive: []
contracts: []
result: "Capability rows only (driver attack for Poison Nova, Poison Dagger, Poison Explosion; delivery of Poison Explosion as Corpse Explosion); the fighter already read the nova band and the immune fallback from profiles.json, no Go skill literal added. Test TestPoisonNovaAttackAndImmuneFallback; scenarios GREEN seed 11 (job 20261123-000500-2301): nova kills a Fallen pack, no nova kill on poison-immune skeletons (the golem kills them, so Bone Spear is unit-tested only). Tier stays experimental: proven.json is not promoted (CHEATED evidence)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Poison Nova, Poison Dagger, Poison Explosion have capability rows with a driver (experimental, not unsupported)
- [x] #2 Test: Poison Nova leads at clvl 40, poison-immune falls back to Bone Spear, Paladin gets neither
- [x] #3 Scenarios necro.nova-pack and necro.nova-immune GREEN on the builder
<!-- AC:END -->
