---
id: TASK-517
title: 'cut013: friends release 0.1.3 built, drilled and staged (not published)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T22:54:00Z'
labels:
  - container
  - chore
dependencies: []
priority: medium
ordinal: 281000
card: "cut013"
result: "Friends beta 0.1.3 built from main dfdc3d29e (supersedes the unpublished 0.1.2) and staged, NOT published. For friends: the new watch cockpit is the default (way back: /watch?layout=classic or Menu, Layout, Classic), phone layout and character screen, a Reset button, fewer stuck steps (park rule, object finder), portal guard and farm loop off by default, installer fixes (game folder with a space, Diablo II subfolder found) and key-free per-archive game mounts. d2d unchanged (fork acadbe47f989). Pack sha256 1babaef86b84e38764193437a2df8076cf06b2a38148f6daf328a0a4ca046b74; checklist 7/7; verified live on the worker host with the real game files: install, 21 minute health run, first live Reset (21 s, hero rejoined), update 0.1.1 to 0.1.3 with a fighting hero, crash-release rollback. Screenshots not taken. Details docs/releases/0.1.3.md and 0.1.3-drills.md."
owns:
  - "docs/releases/0.1.3.md"
  - "docs/releases/0.1.3-drills.md"
  - "docs/friend-checklist-0.1.3.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Image, pack and signed set 0.1.3 built from main dfdc3d29e and staged; 0.1.2 superseded
- [x] #2 Checklist 7/7; install, 21 minute health run, update 0.1.1 to 0.1.3, rollback and Reset drills on the worker host
- [x] #3 docs/releases/0.1.3.md and drills record written
<!-- AC:END -->
