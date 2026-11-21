---
id: TASK-783
title: 'mgmtfast: immediate management actions (knob, Recall now, audit)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T19:05:00Z'
updated_date: '2026-11-21T19:53:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 361000
card: "mgmtfast"
owns:
  - "internal/bot/recall.go"
  - "internal/bot/recall_test.go"
  - "internal/bot/runner.go"
  - "internal/bot/goal.go"
  - "internal/bot/playeraction.go"
  - "internal/watch/recallapi.go"
  - "internal/watch/recallapi_test.go"
  - "internal/watch/funwire.go"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "cmd/d2bot/recallwire.go"
  - "cmd/d2bot/invverbs.go"
  - "cmd/d2bot/invverbs_test.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/live.go"
  - "internal/env/tuning.go"
  - "internal/env/tuning_ui.go"
  - "docs/knobs.md"
  - "docs/process/management-fast-path.md"
  - "tools/provingground/scenarios/manage.json"
  - "tools/provingground/scenarios/behaviours.json"
exclusive: []
result: "Built and tested offline (not verified live): audit doc, knob ui.managementFast, Recall now route and Do it now button, 3 scenarios declared pending. Live RED and green not run."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: management actions must not wait behind the goal loop (user principle 2026-11-21). Done when: the audit, the umbrella knob and Recall now exist and are tested offline; the live runs are a checklist in the doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Audit of every management action in docs/process/management-fast-path.md
- [x] #2 ui.managementFast knob opens the item-verb lane (default 0)
- [x] #3 Recall now route, state and Do it now button
- [x] #4 Scenarios declared (live RED and green not run)
<!-- AC:END -->
