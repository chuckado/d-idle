---
id: TASK-605
title: 'quests10: Ether hook in the quest board hand-in (S10)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T10:01:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 307000
card: "quests10"
owns:
  - "internal/questboard"
  - "internal/ether/earner.go"
  - "cmd/d2bot/questwire.go"
  - "internal/bot/questboard.go"
  - "internal/bot/runner.go"
  - "internal/watch/web/panels/quest-board"
  - "internal/watch/questsapi_test.go"
  - "internal/watch/testdata/quests"
  - "internal/watch/testdata/questui"
  - "tools/uishots/fixtures_quests.go"
  - "tools/uishots.sh"
  - "docs/questboard.md"
  - "docs/ether.md"
  - "docs/design/quest-board-2026-11-16.md"
exclusive: []
contracts: []
result: "Ether hook built offline (Exec.HandIn/Settle, plan-file baseline, pays/awarded/already, UI, goldens); tests on Paladin and Sorceress; NOT verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: quests pay Ether once per account per quest per difficulty (design section 7). Done when: the hand-in books the award through the wallet, idempotent, no backfill; API and UI show it. Result: built and tested offline on two builds; NOT verified live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Awards booked in the hand-in path through the wallet, idempotent, no backfill
- [x] #2 pays/awarded/already in the API; UI toast, chip, row; tests on two builds
- [x] #3 Docs and design doc updated
<!-- AC:END -->
