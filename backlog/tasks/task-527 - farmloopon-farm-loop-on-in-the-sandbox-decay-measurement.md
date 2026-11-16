---
id: TASK-527
title: 'farmloopon: farm loop on in the sandbox, decay measurement'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:26:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 283000
card: "farmloopon"
result: "bot.farmReset=1 set in the sandbox tuning file (backup kept) and the bot restarted; healthy, hero alive, verified from logs; no farm reset line yet. Measurement script and recipe added with a baseline run."
owns:
  - "docs/probes/farmloop-on-20261115.md"
  - "docs/probes/farmloop-measure.md"
  - "tools/farmloop/measure.py"
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
- [x] #1 bot.farmReset=1 set in the sandbox tuning file with backup, effective after restart
- [x] #2 Measurement script and recipe with a baseline run
<!-- AC:END -->
