---
id: TASK-448
title: 'prfix2: apply the verifier''s required changes to P03 and P05'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:05:00Z'
labels:
  - docs
  - upstream-pr
dependencies: []
priority: medium
ordinal: 259000
card: "prfix2"
owns:
  - "docs/upstream/drafts/hitrec-modes.md"
  - "docs/upstream/drafts/hitrec-norestart.md"
  - "docs/upstream/drafts/load-gold-clamp.md"
  - "docs/upstream/bodies/hitrec-modes-body.md"
  - "docs/upstream/bodies/load-gold-clamp-body.md"
  - "docs/upstream/briefs/hitrec-norestart-oracle.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive: []
contracts: []
result: "P05 fid/load-gold-clamp-r3 d229705a and P03 fid/hitrec-modes-r2 3264d61d (block length only) verified-ready, not posted; no-restart split to fid/hitrec-norestart-r1 9169146a, BLOCKED on an Orion oracle run. All three pushed to the private fork. d2d tests and differentials V run on the NAS builder; the literal emulator diff against the new heads was not run. Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 P05 fork branch fid/load-gold-clamp-r3 has the template body, docs row, exit-1 on an unopened case file with a test, the fresh 539/539 game run cited
- [x] #2 P03 fork branch fid/hitrec-modes-r2 is block length only: clang-tidy clean, README and docs rows, body sections, fresh game-side proof
- [x] #3 the no-restart change is on fid/hitrec-norestart-r1, BLOCKED, with the Orion run written down
- [x] #4 branches pushed to the private fork remote; d2d tests rebuilt and rerun on the builder
<!-- AC:END -->
