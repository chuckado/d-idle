---
id: TASK-578
title: 'vaultv0: Ethereal Vault V0 spike (ground full + item create, lab drill)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T02:42:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 300000
card: "vaultv0"
owns:
  - "docs/probes/vault-spike-2026-11-16.md"
exclusive:
  - "fork-branch:didle/vault-spike"
contracts: []
result: "GO for V1. Fork didle/vault-spike (b4d2f861, pushed): ground full / item full / item create behind --vault-verbs. Builder drill 5 rounds: 860 of 860 captured, 820 of 820 created bytes-equal, 40 of 40 reload-equal and cross-hero equal, 0 failures; U1 settled (unidentified items carry the full roll); finding: uid is 0 on d2d items so the seed dup guard is useless. Verified on a headless d2d on the builder, not on Game.exe; oracle not run."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: prove the vault capture and create path (design docs/design/loot-vault-2026-11-16.md section 1.7) before any bot change. Consumer: vault V1 (fork verbs and Go client). Done when: the drill numbers and a go/no-go are in the probe doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fork stub ground full / item full / item create behind --vault-verbs pushed to the private remote
- [x] #2 lab drill on the builder with numbers (items, bytes equal, reload equal, failures)
- [x] #3 report docs/probes/vault-spike-2026-11-16.md with verb shapes and go/no-go for V1
<!-- AC:END -->
