---
id: TASK-717
title: 'bsm1: quest board SM scenarios (board.json) and the board/api_reply assert kinds'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:40:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 341000
card: "bsm1"
owns:
  - "tools/provingground/scenarios/board.json"
  - "internal/scenario"
  - "cmd/d2scenario"
exclusive: []
contracts: []
result: "18 red scenarios in board.json (pending, tier full); assert kinds board, board_heads_le, api_reply; runner samples GET /api/quests every 15 wall s and records api replies. Unit tested; not verified live (pending scenarios never run)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Card 1 of docs/design/fable-review-2026-11-21/quest-board.md section 5. Done when the scenario set and the runner additions are merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 board.json holds the 18 red scenarios of appendix B.3 (pending, tier full)
- [x] #2 assert kinds board, board_heads_le, api_reply and the runner side files, unit tested
- [x] #3 go test internal/scenario cmd/d2scenario green
<!-- AC:END -->
