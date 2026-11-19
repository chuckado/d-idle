---
id: TASK-677
title: 'agencyp1a: Agency P1a capability table, shadow skill planner, build files'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T22:05:00Z'
updated_date: '2026-11-19T23:04:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 331000
card: "agencyp1a"
owns:
  - "internal/agency"
  - "internal/buildfile"
  - "data/capabilities"
  - "docs/agency.md"
  - "internal/watch/planapi.go"
  - "internal/watch/planapi_test.go"
  - "cmd/d2bot/agencywire.go"
  - "cmd/d2bot/agencywire_test.go"
  - "internal/env/tuning_agency.go"
  - "tools/provingground/scenarios/agency.json"
  - "internal/watch/funwire.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/web/points.js"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/live.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/behaviors.go"
  - "internal/env/tuning.go"
  - "docs/knobs.md"
  - "tools/provingground/scenarios/behaviours.json"
exclusive: []
contracts: []
result: "Shadow only, knob agency.planner default 0, no change to what the hero does; NOT verified live (offline tests only). Capability table for 30 Paladin and 30 Sorceress skills (data/capabilities), planner agrees with the combat profile at 25 of 30 Paladin/Sorceress checkpoints (30 of 40 with paper classes), GET /api/hero/plan, Points page line, internal/buildfile with plan table import/export. Scenario agency.plan-matches-profile is pending (needs a log assertion). docs/agency.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Capability table as data for all Paladin and Sorceress skills (kind, delivery, tier with reason, requirements, mana source, synergies)
- [x] #2 Shadow planner: rotation, aura, warnings, viability line; knob agency.planner 0 off 1 shadow; planner shadow log line a minute; GET /api/hero/plan; Points page line
- [x] #3 internal/buildfile with validator and plan table import/export; shipped Paladin and Sorceress tables validate against the capability table
- [x] #4 Tests: agreement fixture vs combat profiles, unsupported never chosen, paper Barbarian and Necromancer, -race -count=20
- [x] #5 docs/agency.md; scenario agency.plan-matches-profile (pending: needs a log assertion kind)
<!-- AC:END -->
