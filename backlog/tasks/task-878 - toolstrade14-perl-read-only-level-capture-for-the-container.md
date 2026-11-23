---
id: TASK-878
title: 'toolstrade14: perl read-only level capture for the container'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T08:52:00Z'
updated_date: '2026-11-23T09:11:00Z'
labels:
  - ops
  - probe
dependencies: []
priority: medium
ordinal: 395000
card: "toolstrade14"
owns:
  - "tools/probes/capture.pl"
  - "tools/probes/capture_pl_test.py"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "capture.pl: same read verbs as capture.py, JSON lines on stdout, nothing written in the container; tested against a fake socket; not yet run on the live container."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. The sandbox container has perl, no python. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 capture.pl captures one level through read verbs only, to stdout; allow-list and refusal tested against a fake socket
<!-- AC:END -->
