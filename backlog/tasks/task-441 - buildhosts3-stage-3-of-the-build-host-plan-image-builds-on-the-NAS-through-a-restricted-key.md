---
id: TASK-441
title: 'buildhosts3: stage 3 of the build host plan, image builds on the NAS through a restricted key'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T21:02:00Z'
labels:
  - ops
  - chore
dependencies: []
priority: medium
ordinal: 246000
card: "buildhosts3"
result: "Stage 3 applied: restricted forced-command key on the NAS (build, save, rmi scratch-, list), tools/imgbuild.sh client, dev-update --image and friendpack --image-from-nas use it, ops gate and status scripts default to the builder; small image 12 s and real didle image about 9 min built on the NAS into scratch tags and removed; verified live (no game, no user PC)."
owns:
  - "docs/process/build-hosts-stage3-2026-11-14.md"
  - "deploy/nas/imgbuild.sh"
  - "tools/imgbuild.sh"
  - "tools/dev-update.sh"
  - "tools/dev-update-host.sh"
  - "tools/friendpack/build.sh"
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
- [x] #1 Restricted forced-command key installed on the NAS and its refusals drilled
- [x] #2 A real small image and a real didle image built on the NAS into scratch tags
- [x] #3 dev-update, friendpack and the ops scripts no longer default to the worker host
<!-- AC:END -->
