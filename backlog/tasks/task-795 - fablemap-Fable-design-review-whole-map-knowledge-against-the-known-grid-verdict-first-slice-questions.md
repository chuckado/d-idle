---
id: TASK-795
title: 'fablemap: Fable design review, whole-map knowledge against the known grid (verdict, first slice, questions)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T01:50:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 366000
card: "fablemap"
owns:
  - "docs/design/fable-review-2026-11-22/map-knowledge.md"
exclusive: []
contracts: []
result: "Docs only, nothing run against a game (not verified live). Verdict: a whole-map verb is not the fix; on d2d the collision grid already arrives whole (V live for town and Blood Moor per docs, U for maze levels) and the stuck steps came from target knowledge, exit triggers, level links and guards. First slice of five cards (gridwhole, knownedge, warptrigger, exitsread, fogtrail), seven questions for the user. Build-specific: none."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: the user asked whether revealing the whole map would fix the Countess class of stuck steps. Done when: docs/design/fable-review-2026-11-22/map-knowledge.md is merged. Consumer: the user decision on its section 6 and the first-slice cards (gridwhole, knownedge).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 doc with verdict, costs, first slice and numbered questions, file:line evidence
- [x] #2 contracts, backlog and leak checks pass
<!-- AC:END -->
