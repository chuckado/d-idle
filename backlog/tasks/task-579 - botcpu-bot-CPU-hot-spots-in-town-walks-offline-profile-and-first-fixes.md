---
id: TASK-579
title: 'botcpu: bot CPU hot spots in town walks, offline profile and first fixes'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T02:49:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 300000
card: "botcpu"
owns:
  - "docs/probes/bot-cpu-2026-11-16.md"
  - "internal/senses"
  - "cmd/d2bot/sensesfeed.go"
  - "cmd/d2bot/sensesfeed_test.go"
  - "internal/nav/cpuprobe_test.go"
  - "internal/d2dsrc/cpuprobe_test.go"
  - "internal/d2dsrc/source_test.go"
  - "internal/d2dsrc/srv_test.go"
exclusive: []
contracts: []
result: "Offline profile only, not live. Town walk nav plus snapshot decode under 0.4 percent of a core; senses map rebuild 4.1x faster, roam coverage push throttled to 2 Hz. The sandbox burst cause is not found offline; live D2DCTL_PPROF checklist in docs/probes/bot-cpu-2026-11-16.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Offline CPU profile of the town walk path with a top table
- [x] #2 Top offenders fixed with equality tests
- [x] #3 Live profile of a repair trip (D2DCTL_PPROF) is a checklist
<!-- AC:END -->
