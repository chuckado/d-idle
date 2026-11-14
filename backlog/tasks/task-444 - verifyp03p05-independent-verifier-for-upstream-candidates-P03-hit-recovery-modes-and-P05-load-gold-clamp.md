---
id: TASK-444
title: 'verifyp03p05: independent verifier for upstream candidates P03 (hit-recovery modes) and P05 (load gold clamp)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T21:28:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 259000
card: "verifyp03p05"
owns:
  - "docs/upstream/verifier-p03-p05-2026-11-14.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Opus verifier pass before posting; docs only; nothing posted. Consumer: decomp21 (TASK-440) and the S6 post step of docs/upstream/pipeline-2026-11-14.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Verdict per item (PASS / PASS WITH CHANGES / FAIL) with required changes
- [ ] #2 Builds and d2d tests rerun on the NAS builder for oracle, claim and head commits
- [ ] #3 Proven vs unproven split, game.exe side vs our docs
- [ ] #4 Required worker-host emulator runs listed with exact commands
- [ ] #5 Emulator runs done on the worker host (live/emulator; blocked today, Orion off limits)
<!-- AC:END -->
