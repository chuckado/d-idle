---
id: TASK-920
title: 'toolstrade43: readiness floors from game data (research, nothing applied)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T13:49:00Z'
updated_date: '2026-11-26T01:45:00Z'
labels:
  - docs
  - probe
dependencies: []
priority: medium
ordinal: 423000
card: "toolstrade43"
owns:
  - "docs/probes/readiness-floors-2026-11-24.md"
  - "internal/power/floors_report_test.go"
exclusive: []
contracts: []
result: "Our area levels match the third-party chart (Barracks 9, Catacombs 11, 11, 12, 12, Andariel 12). The Barracks hurts by crowd size (13.9 hits per minute against 4 in the fields), not per hit (6.3 life, as the table says). Worn armour made no difference in the log (74 against 77 life per min). Floors depend on the hero k (Barracks 13 to 22, Andariel 16 to 28 at k 1 to 3, as-played gear). Proposal only, nothing applied; read-only live; model numbers are S, the cause of k is U."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Floors derived from game data and the live logs in a probe doc
- [x] #2 Report whether the unworn gear explains the death rate
<!-- AC:END -->
