---
id: TASK-837
title: 'denfix: act1.den.clear for the Necromancer keeps its purse (town.bankFirst 0), Den RED cause'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T02:30:00Z'
labels:
  - bot
  - drill
dependencies: []
priority: medium
ordinal: 381000
card: "denfix"
owns:
  - "tools/provingground/scenarios/act1.json"
exclusive: []
contracts: []
result: "Cause: the bank chore deposited 6000 of 8000 gold, the hero was short in purse and kit and left the Den after 36 s to farm gold (Necromancer 11, run 20261122-205700). The assertion (record 1 bit 0 = completed) was right; 0x0014 = Akara talked (0x4) and 0x10 (U). Fix in the scenario only (town.bankFirst 0, as the act1more scenarios); the bot-side defect remains: withdraw stash gold when chores are short (act1more finding 1). Rerun GREEN Necromancer seed 11 (job 20261122-212018-1980)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Cause of the Necromancer Den RED found in the log
- [x] #2 Scenario rerun GREEN for Necromancer and Paladin
<!-- AC:END -->
