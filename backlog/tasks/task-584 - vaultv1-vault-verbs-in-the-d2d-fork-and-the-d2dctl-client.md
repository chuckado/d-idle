---
id: TASK-584
title: 'vaultv1: vault verbs in the d2d fork and the d2dctl client'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T03:26:00Z'
labels:
  - d2d
  - feature
  - d2d-loot
dependencies: []
priority: medium
ordinal: 301000
card: "vaultv1"
owns:
  - "internal/d2dctl"
  - "docs/d2d-vault-verbs.md"
exclusive:
  - "fork-branch:didle/vault-verbs"
contracts: []
result: "Fork didle/vault-verbs f2ed52c8 (private remote): ground full/take, item full/create/vault, vault stats, caps, key dedupe, dead-hero refusal; tests/vault_verbs_d2d.py 0 failures on the builder headless (not on a live bot). d2dctl typed client, VaultFake, two-class tests. Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Slice V1 of the Ethereal Vault (docs/design/loot-vault-2026-11-16.md, probe docs/probes/vault-spike-2026-11-16.md). Done when the criteria hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork branch didle/vault-verbs on the private remote: ground full/take, item full/create/vault, caps, flag off by default, tests/vault_verbs_d2d.py passes on the builder
- [x] #2 internal/d2dctl typed client, VaultFake, dedupe and error mapping, tests on two classes, go test -p 1 and labrun race clean
- [x] #3 V0 open issues resolved or documented in docs/d2d-vault-verbs.md
<!-- AC:END -->
