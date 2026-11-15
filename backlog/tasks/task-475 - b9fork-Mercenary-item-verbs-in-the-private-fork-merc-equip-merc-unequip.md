---
id: TASK-475
title: 'b9fork: Mercenary item verbs in the private fork (merc-equip, merc-unequip)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T05:45:00Z'
labels:
  - d2d
  - feature
dependencies: []
priority: medium
ordinal: 266000
card: "b9fork"
owns:
  - "docs/research/merc-gear.md"
  - "docs/inventory-actions.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fork branch sandbox-mercverbs: verbs, caps item.merc-equip and item.merc-unequip, unit test (three hireling types) and headless differential written; commit on the private remote and named in the card
- [ ] #2 docs/research/merc-gear.md section 8 and docs/inventory-actions.md say what the fork provides (verbs, capability names, error codes) and the live drill
- [ ] #3 Go client contract unchanged (names match item_verbs_merc.go); tools/ci/check.sh clean
<!-- AC:END -->
