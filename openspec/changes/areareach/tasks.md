# Tasks

```yaml
change: areareach
status: in-progress
agent: areareach
branch: agent/areareach
worktree: ~/code/d2lod-wt/areareach
skip_specs: true
owns:
  - internal/bot/task.go
  - internal/bot/task_test.go
  - internal/bot/plan.go
  - internal/bot/plan_test.go
  - internal/bot/goal.go
  - internal/bot/recoveryxp.go
  - docs/research/areareach.md
  - openspec/changes/areareach/**
exclusive: []
gates:
  - "TestDarkWoodReachableByWalkingTheChain, TestLongestWalk and TestPickFallbackByXP fail when the fix is reverted"
credit:
  component: AREAREACH
  author: owner
  agents: [Claude Code]
merge:
result: "offline only (fake mover and plan tests, go test ./internal/bot); not verified live. Gate 1 re-run must show no 'no activated waypoint' lines for Dark Wood/Black Marsh from a field area."

```

## 1. Work

- [x] 1.1 Root cause from the 0628 bot log (docs/research/areareach.md)
- [x] 1.2 unreachable(): walk the level chain between waypoints (longestWalk, TestDarkWoodReachableByWalkingTheChain, TestLongestWalk)
- [x] 1.3 Pick fallback ranked by XP per kill when the preferred area is excluded (TestPickFallbackByXP)
