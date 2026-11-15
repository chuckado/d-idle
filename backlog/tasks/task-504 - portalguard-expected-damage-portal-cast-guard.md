---
id: TASK-504
title: 'portalguard: expected-damage portal cast guard'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T15:18:00Z'
updated_date: '2026-11-15T16:22:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 277000
card: "portalguard"
owns:
  - "internal/resource/"
  - "internal/combat/"
  - "internal/env/tuning_bot.go"
  - "docs/fixportalmelee.md"
  - "docs/knobs.md"
  - "tools/knobdump/main.go"
exclusive: []
contracts: []
result: "bot.portalGuard (default 0 until drill R3; mode 1 behind the knob): resource.CastGuard + combat.castGuard; offline tests on Sorceress and Paladin pass; NOT verified live, drill R3 must measure cast+walk seconds and damage (knobs are generous U bounds 3000+3000 ms, hit 700 ms, elite x3, 50 percent of life)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: a hit does not interrupt a Town Portal cast (V run, decomp28), so the three-adjacent rule guarded the wrong thing; the risk is damage during cast and walk. Done when: bot.portalGuard 1 (default) judges by expected damage and falls back to the count rule when a term is missing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 bot.portalGuard 1 refuses a portal cast on expected damage, falls back to the count rule on missing terms
- [x] #2 Tests on Sorceress and Paladin: 3 weak allowed, 1 strong boss blocked, low life, unknown terms
<!-- AC:END -->
