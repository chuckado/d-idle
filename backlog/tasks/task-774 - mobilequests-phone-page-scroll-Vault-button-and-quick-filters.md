---
id: TASK-774
title: 'mobilequests: phone page scroll, Vault button and quick filters'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T14:15:00Z'
updated_date: '2026-11-21T15:40:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 359000
card: "mobilequests"
owns:
  - "internal/watch/web"
  - "internal/watch/vaultlink_test.go"
  - "tools/mobilescroll"
  - "docs/ui/mobile"
exclusive: []
contracts: []
result: "Phone (flow breakpoint cmode=phone) is a page: the cockpit frame grows to its content, tab lists lose their inner scroll boxes (utilities.css), the Vault page scrolls as one document under 700 px with 2-line notices; band Vault link; Vault quick filter chips (type chain, counts from facets, persisted). Verified in headless Chrome 412x915 touch on the worker host with tools/mobilescroll/run.sh (old CSS correctly caught); NOT verified on a real phone. Omitted: an Other chip (the server type filter cannot exclude); the Points page looked 981 px wide in the fixture viewer (a PRE element, not investigated)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: on a phone the Quests tab list and the Vault page were scroll traps (fixed 100dvh frame, inner scroll boxes). Done when: the probes pass.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a 412 px phone every cockpit tab and the Vault page scroll with the page (no nested scroll box)
- [x] #2 Band Vault link and Vault quick filter chips
- [x] #3 Must-fail probe tools/mobilescroll/run.sh (old CSS caught)
<!-- AC:END -->
