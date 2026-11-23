---
id: TASK-844
title: 'fitfix3: Vault fit flakes gone, a list refresh no longer ends Select mode'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T03:32:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 384000
card: "fitfix3"
owns:
  - "tools/uishots"
  - "tools/uishots.sh"
  - "internal/watch/web/vault-logic.js"
  - "internal/watch/testdata/vault_logic.mjs"
exclusive: []
contracts: []
result: "The 12 flaky fit cases were a real race: a refresh or lazy page (absorb -> prune) ended Select mode while nothing was picked, so the next tap opened the card. prune now ends the mode only when the last selected item left; harness settles (fetches done, DOM quiet) and retries hits; reveal rows no longer expect select-done. vault* fit twice on the worker host: 550 ok 1 fail both runs (vault-more-entry-390x844 dead space of the phone cockpit, not vault). Not live game."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Root cause found: selection prune ended Select mode on an empty selection when a refresh arrived; fixed with a node test
- [x] #2 Fit harness settles after each press; vault* fit is deterministic (two runs identical)
<!-- AC:END -->
