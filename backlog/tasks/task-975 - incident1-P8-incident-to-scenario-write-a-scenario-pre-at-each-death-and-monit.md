---
id: TASK-975
title: 'incident1: P8 incident to scenario: write a scenario pre at each death and monit'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:55:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-967
priority: medium
ordinal: 463000
card: "incident1"
owns:
  - "docs/design/incident-p8.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Cause class killed: the lab green, live red gap (review 5.2). Roadmap rows P8 and T2, T7 (docs/research/fable-logic-review-2026-11-25.md 5.3, 7.1). The bot writes the scenario `pre` of that moment (level, life, gold, stash, belt and pack counts, kit, quest bits, area, accepted chunk) at each death, park, flail hold or monitor hit; a nightly job replays the newest N on the lockstep engine; first three scenarios are the Questy deaths of 2026-11-26; report-only scenario tier in the merge queue (the ops half, pipeline audit row 6). Incident pre is a recorded state, tagged SETUP (test-character provenance rule). Provides the outcome measure for arbmig. Risk: low. Enable criteria: report-only tier runs for 7 clean days, then the owner decides to block (ADR 0027 item 2).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 pre writer at death and monitor hit with a test
- [ ] #2 First three Questy-death scenarios red-then-green; the ops report-only tier named in the result
<!-- AC:END -->
