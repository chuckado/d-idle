---
id: TASK-674
title: 'pgexpand: Proving Ground expansion, Cain chain and behaviour scenarios runnable'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-19T18:48:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 329000
card: "pgexpand"
owns:
  - "internal/scenario/"
  - "cmd/d2scenario/"
  - "tools/provingground/"
  - "docs/process/proving-ground.md"
  - "docs/probes/provingground-run-2026-11-19.md"
exclusive:
  - "d2d-lab-binary"
contracts: []
result: "Proving Ground expanded: board-taps driver, area-change counter, portal and ground observers, 19 scenarios (14 runnable, 1 BLOCKED by design, 3 PENDING). Fast tier run live on the builder (lab slots, CHEATED/SETUP, mechanics only): GREEN 4 (belt.refill-after-fight, cain.read, town.resupply, town.stash-gold-withdraw), RED 3 (cain.scroll, flail.areachanges 26 area changes, heal.town scenario artifact). Determinism pair submitted (job 20261119-141800-a155), result not in this commit. Coverage 22 of 91 behaviours (was 10). Record: docs/probes/provingground-run-2026-11-19.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Expand the Proving Ground (card provingground). Done when the listed scenarios run and the run table is committed.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Cain stones, cage, read, report and the whole chain are runnable scenarios with a recorded observable
- [x] #2 Behaviour scenarios (flail, stash withdraw, resupply, rare ring, chicken, heal, belt) are runnable; the area-change counter is in the observer
- [ ] #3 Fast tier run on the builder with a G/R/P/B table and a determinism pair committed
- [x] #4 Coverage report updated
<!-- AC:END -->
