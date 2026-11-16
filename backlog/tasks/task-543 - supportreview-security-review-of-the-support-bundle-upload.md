---
id: TASK-543
title: 'supportreview: security review of the support bundle upload'
status: Proposed
assignee:
  - user
created_date: '2026-11-16T04:30:00Z'
labels:
  - docs
  - readonly
dependencies: []
priority: high
ordinal: 287000
card: "supportreview"
owns:
  - "docs/reviews/support-bundle-security-2026-11-15.md"
exclusive: []
contracts: []
result: "Security review of the support bundle (task 542) and re-verification of the fixes at 8981d1f8e: M1-M4 and minors verified fixed by re-running every failing input; client side ships; receiver needs one more fix before deploy (R1: global rate counted before auth lets one unauthenticated client lock out uploads; plus restrict the nginx location to the tunnel). Offline only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Independent Opus security review of agent/support (task 542) before merge. Done when the review doc is merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Review doc with ranked findings, failing inputs and verdict
<!-- AC:END -->
