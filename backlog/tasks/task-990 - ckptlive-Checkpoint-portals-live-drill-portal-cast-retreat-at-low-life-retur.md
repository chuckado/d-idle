---
id: TASK-990
title: 'ckptlive: Checkpoint portals live drill: portal cast, retreat at low life, retur'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:59:00Z'
labels:
  - bot
  - drill
dependencies: []
priority: medium
ordinal: 477000
card: "ckptlive"
owns:
  - "docs/probes/checkpoint-live-drill.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

bot.checkpointEnable is built (checkpoint1, 2026-11-25) and UNVERIFIED live; ledger row says a live drill is the enable criterion. This card runs it (hidden instance only, a throwaway character, ADR 0014 SETUP writes logged) and records: portal cast healthy, retreat at low life, return after death. Cause classes: the Dark Wood death series (class G / safety). Depends on nothing; the 24 h invariant-monitor watch with auto-revert follows the flip. Risk: low. Enable criteria (data/darklaunch.json): drill passes; 24 h monitor watch, auto-revert on a hard breach.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Drill steps run on a hidden instance with the writes ledger line
- [ ] #2 Result says verified live or not
<!-- AC:END -->
