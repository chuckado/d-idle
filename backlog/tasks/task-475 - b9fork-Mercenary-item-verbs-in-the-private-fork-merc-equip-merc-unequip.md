---
id: TASK-475
title: 'b9fork: Mercenary item verbs in the private fork (merc-equip, merc-unequip)'
status: In Progress
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
result: "Fork half built and tested on the NAS builder, NOT verified live: private fork branch sandbox-mercverbs head c2549842 (on the private remote): item merc-equip / merc-unequip, caps item.merc-equip and item.merc-unequip, test_merc_verbs (Rogue 271, Guard 338, Wolf 359) and merc_verbs_d2d.py (differential vs the manual command sequence) pass; dump merc now reads the live list. Go client names unchanged. Not deployed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the Go half (uib9) refuses merc item requests with code merc until a fork build offers the verbs. Done when: the fork has them behind --item-verbs, tested, pushed, and the research doc names verbs, capability and errors. Fork checkout used: a clone on the head box, built and tested on the NAS builder.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork branch sandbox-mercverbs: verbs, caps item.merc-equip and item.merc-unequip, unit test (three hireling types) and headless differential written; commit on the private remote and named in the card
- [x] #2 docs/research/merc-gear.md section 8 and docs/inventory-actions.md say what the fork provides (verbs, capability names, error codes) and the live drill
- [x] #3 Go client contract unchanged (names match item_verbs_merc.go); tools/ci/check.sh clean
<!-- AC:END -->
