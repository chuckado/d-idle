---
id: TASK-779
title: 'heavyguard: builder memory guard for browser harnesses'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T15:27:00Z'
labels:
  - ops
  - chore
dependencies: []
priority: medium
ordinal: 246000
card: "heavyguard"
owns:
  - "tools/lib/heavy-guard.sh"
  - "tools/lib/heavy-guard-test.sh"
  - "tools/uishots.sh"
  - "docs/process/worker-capacity.md"
exclusive: []
contracts: []
result: "tools/lib/heavy-guard.sh (exit 3 unless long lane on the builder, 3000 MB free, no cc1plus/ninja/ctest) with offline test, used by uishots.sh; rules documented. Offline only, not verified live. mobilescroll lives on agent/mobilequests and must source it."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 tools/lib/heavy-guard.sh with offline test
- [x] #2 uishots.sh uses the guard
- [x] #3 Builder memory rules documented
<!-- AC:END -->
