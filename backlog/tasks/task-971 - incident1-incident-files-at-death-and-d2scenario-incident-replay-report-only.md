---
id: TASK-971
title: >-
  incident1: incident files at death and d2scenario incident replay
  (report-only)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T03:53:00Z'
updated_date: '2026-11-26T07:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 455000
card: "incident1"
owns:
  - "internal/bot/incident.go"
  - "internal/bot/incident_test.go"
  - "internal/bot/observe.go"
  - "internal/scenario/incident.go"
  - "internal/scenario/incident_test.go"
  - "cmd/d2scenario/incident.go"
  - "cmd/d2scenario/incident_test.go"
  - "cmd/d2scenario/main.go"
  - "tools/provingground/incidents/"
  - "docs/probes/incident-scenarios.md"
exclusive: []
contracts: []
result: "incident files at death, d2scenario incident, three Questy rows (report-only tier); offline only, no lockstep run, rows predicted RED before their fixes"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 bot writes an incident file at each death
- [x] #2 d2scenario incident builds replay rows
- [x] #3 three Questy rows validate and plan
- [x] #4 docs/probes/incident-scenarios.md
<!-- AC:END -->
