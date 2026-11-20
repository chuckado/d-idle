---
id: TASK-516
title: 'devp8: sandbox image with hero stats and merc item verbs'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T19:16:00Z'
updated_date: '2026-11-15T19:32:00Z'
labels:
  - container
  - chore
dependencies: []
priority: medium
ordinal: 280000
card: "devp8"
owns:
  - "docs/probes/devp8-2026-11-15.md"
exclusive:
  - "fork-branch:sandbox-devp8"
contracts: []
result: "d2dbot:devp8 (fork sandbox-devp8 dbaf61fa: herostats + mercverbs on the devp7 lineage) built on the NAS, loaded and swapped into the sandbox 2026-11-15; healthy, zone entry shows def block dmg dmgpct ar; fork tests and both differentials pass; hero dump and merc verbs not run on the live hero; verified live: log fields only"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the survival power model needs the hero stats block and the merc desk needs the merc item verbs in the sandbox d2d. Done when: image devp8 runs in the sandbox, with devp7 as rollback.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork branch sandbox-devp8 (devp7 lineage + sandbox-herostats + sandbox-mercverbs) pushed to the private fork
- [x] #2 Fork tests and python differentials pass on the NAS-built d2d
- [x] #3 Image d2dbot:devp8 built on the NAS and loaded on the sandbox host
- [x] #4 Sandbox swapped to devp8 and checked read-only (or rolled back)
<!-- AC:END -->
