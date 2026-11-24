---
id: TASK-921
title: 'toolstrade44: owner try-anyway start (one shot readiness override)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T14:00:00Z'
updated_date: '2026-11-24T16:01:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 424000
card: "toolstrade44"
owns:
  - "internal/questboard/exec.go"
  - "internal/questboard/exec_sm.go"
  - "internal/questboard/desk.go"
  - "internal/questboard/desk_action.go"
  - "internal/questboard/try.go"
  - "internal/questboard/try_test.go"
  - "internal/ops/ops.go"
  - "internal/ops/ops_test.go"
  - "internal/watch/questsapi.go"
  - "internal/watch/questsapi_test.go"
  - "internal/watch/web/panels/quest-board/panel.js"
  - "internal/task/try.go"
  - "internal/task/route.go"
  - "internal/task/gate.go"
  - "internal/task/act1.go"
  - "internal/task/act2.go"
  - "internal/task/route_test.go"
  - "internal/boss/run.go"
  - "internal/bot/questboard.go"
  - "internal/bot/healcap_test.go"
  - "cmd/d2bot/opscmd.go"
  - "cmd/d2bot/opscmd_test.go"
  - "cmd/d2bot/boss.go"
  - "docs/operator.md"
  - "docs/questboard.md"
  - "docs/adr/0030-operator-quest-control.md"
  - "docs/decisions.md"
  - "internal/questboard/desk_action_test.go"
  - "internal/watch/testdata/quests/"
  - "internal/watch/vaultperf_test.go"
  - "internal/watch/race_on_test.go"
  - "internal/watch/race_off_test.go"
exclusive: []
contracts: []
result: "Built offline, not verified live: operator `d2bot ops quest start <quest> --try-anyway --why ...` and the viewer button Try anyway (risk) (POST /api/quests/<id>/try with confirm) start one Available chunk past its readiness refusal; Exec.AcceptTry keeps the override in memory (ends on park, death, finish, stop, restart); task.SetTry skips only the readiness words and floors (boss gates, Barracks, road, tower, Andariel hold, boss script floor); kit gates, guards and combat reactions untouched; audit path operator/try-anyway with the skipped refusal; same rate limits; refused in a measured run. ADR 0030 amended, decisions.md line."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: L
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Operator start --try-anyway and the viewer Try anyway button start one not-ready chunk past its readiness refusal, audited, one shot, rate limited, never in a measured run
- [x] #2 Only readiness words are skipped: kit gates, guards and combat reactions stay (tests on two classes)
- [x] #3 ADR 0030 amended, decisions.md line, operator and questboard docs
<!-- AC:END -->
