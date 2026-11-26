---
id: TASK-1010
title: 'auralive: Paladin aura state on d2d after pick, cast, other skill, area change'
status: Done
assignee:
  - live-agent
created_date: '2026-11-26T18:05:00Z'
updated_date: '2026-11-26T18:34:00Z'
labels:
  - d2d
  - d2d-measure
dependencies: []
priority: medium
ordinal: 483000
card: "auralive"
lane: next
owns:
  - "docs/probes/paladin-aura-live-2026-11-26.md"
  - "cmd/auraprobe"
exclusive: []
result: "Verified live on a d2d lab slot: a Paladin aura picked through the right skill popup is active at once with no cast (state aura field and a Holy Fire pulse that hurt a monster); it survives attack casts, aura switches and area changes, and a non-aura right skill turns it off. Not verified: character load. Doc docs/probes/paladin-aura-live-2026-11-26.md, probe cmd/auraprobe."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Aura state measured on d2d for steps A to F and recorded with exact commands
<!-- AC:END -->
