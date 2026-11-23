---
id: TASK-843
title: 'fitfix2: phone More sheet scrolls again and the Vault Reveal button always has a name'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T03:12:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 383000
card: "fitfix2"
owns:
  - "tools/uishots"
  - "tools/uishots.sh"
  - "internal/watch/web/utilities.css"
  - "internal/watch/web/vault.js"
  - "internal/watch/testdata/vault_reveal.mjs"
exclusive: []
contracts: []
result: "Phone More sheet: data-scroll no longer beats .d-tilesheet overflow (was 21 px taller, no scroll), must-fail moredefect-phone; Vault reveal-bar-go had an empty label when nothing is payable, now Reveal, node test audits unnamed buttons. Fit run on the worker host; phone cockpit still red for dead space and page scroll (not fixed). Not live game."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 More sheet on a phone scrolls inside its tile (must-fail fit row moredefect-phone)
- [x] #2 Vault Reveal button has a label when nothing is payable; node test audits every visible button name
<!-- AC:END -->
