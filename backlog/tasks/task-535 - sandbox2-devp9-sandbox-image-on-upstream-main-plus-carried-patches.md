---
id: TASK-535
title: 'sandbox2: devp9 sandbox image on upstream main plus carried patches'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-16T03:19:00Z'
labels:
  - d2d
  - chore
dependencies: []
priority: medium
ordinal: 246000
card: "sandbox2"
owns:
  - "docs/process/sandbox-devp9.md"
exclusive:
  - "fork-branch:didle/rebase-20261115"
contracts: []
result: "Sandbox d2play-20261108 runs d2dbot:devp9 (fork didle/rebase-20261115 fabae935 on upstream 485fe535, d2d built on the NAS builder) plus a fresh dev bot release. Verified live read-only for 10 min: healthy, hero clvl 7 to 8, 180 kills in 10 min (166 before), 0 errors, 1 chicken Save and Exit (1 before). Builder tests: 4 of 8 ran and passed, 4 skipped (no data there); data oracles not run. Rollback not exercised. Details: docs/process/sandbox-devp9.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Bring the sandbox d2d build up to date with merged upstream PRs. Done when the sandbox runs devp9 and the doc records patches and PRs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork branch didle/rebase-20261115 pushed to the private remote
- [x] #2 Image d2dbot:devp9 built from a builder-made d2d
- [x] #3 Sandbox deployed, verified read-only, rollback kept
<!-- AC:END -->
