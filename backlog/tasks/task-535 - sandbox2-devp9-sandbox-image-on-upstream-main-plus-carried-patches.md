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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Bring the sandbox d2d build up to date with merged upstream PRs. Done when the sandbox runs devp9 and the doc records patches and PRs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fork branch didle/rebase-20261115 pushed to the private remote
- [ ] #2 Image d2dbot:devp9 built from a builder-made d2d
- [ ] #3 Sandbox deployed, verified read-only, rollback kept
<!-- AC:END -->
