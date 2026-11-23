---
id: TASK-805
title: 'identifyall: Vault Identify all button, gold reveal scope all_unrevealed'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T15:38:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 369000
card: "identifyall"
owns:
  - "internal/vaultkeep"
  - "internal/bulkid"
  - "internal/watch"
  - "cmd/d2bot/vaultkeep.go"
  - "docs/vault.md"
  - "docs/ui/vault-identify-all"
  - "tools/provingground/scenarios/vaultreveal.json"
  - "tools/uishots"
exclusive: []
contracts: []
result: 'Vault toolbar button Identify all (N) (N = whole-vault unrevealed count, hidden at 0) with a confirm sheet (enough or short stash) posting scope all_unrevealed; the bot lists the records itself under the order lock (Keeper.EnqueueRevealGoldAll, bulkid.UnrevealedIDs), partial reveal, idempotent resend, records the game cannot decode are named. Go tests Paladin and Sorceress, -race -count=20 on vaultkeep and bulkid, DOM tests, pictures 390x844 and 1366x768. Scenario vault.reveal-all GREEN on a lab slot (builder long lane, d2d devp19, Paladin and Sorceress seed 11, UNMEASURED run, CHEATED setup, mechanics only); one seeded bare socketed magic armor had no decodable affix and stays unrevealed (U cause), so the assertions allow 29 or 30. Build-specific: none. Not deployed.'
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the player wants one button to identify everything unidentified in the vault, not only the filter. Done when: the button, the all_unrevealed order scope, tests, docs, pictures and the scenario are green.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Identify all (N) button, scope all_unrevealed, tests, docs, screenshots (scenario vault.reveal-all GREEN on a lab slot)
<!-- AC:END -->
