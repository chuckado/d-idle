---
id: TASK-623
title: 'sandbox3: sandbox image devp10 (vault and labels verbs on current upstream)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T15:37:00Z'
updated_date: '2026-11-17T16:57:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 310000
card: "sandbox3"
owns:
  - "docs/process/sandbox-devp10.md"
exclusive: []
result: "Fork branch didle/rebase-20261117 tip 9df9b936 built (NAS builder -j2), headless verb tests pass (item, lab, vault, labels, cursor, merc, overlay), image d2dbot:devp10 built and loaded on the sandbox host, not deployed; switch and rollback in docs/process/sandbox-devp10.md. Not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork branch didle/rebase-20261117 pushed, patch list in the doc
- [x] #2 d2d built with build-fork.sh --tests; verb tests run
- [x] #3 Image d2dbot:devp10 built and loaded on the sandbox host (not deployed)
- [x] #4 Switch and rollback recipe documented
<!-- AC:END -->
