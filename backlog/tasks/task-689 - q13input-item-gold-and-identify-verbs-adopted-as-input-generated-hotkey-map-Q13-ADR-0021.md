---
id: TASK-689
title: 'q13input: item, gold and identify verbs adopted as input; generated hotkey map (Q13, ADR 0021)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T16:19:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 336000
card: "q13input"
owns:
  - "internal/invdesk/mover.go"
  - "internal/invdesk/verbs.go"
  - "cmd/d2bot/invverbs.go"
  - "cmd/d2bot/d2dwire.go"
  - "data/input/caps_test.go"
  - "docs/knobs.md"
  - "internal/hotkeymap/"
  - "internal/invdesk/inputcaps.go"
  - "internal/invdesk/inputcaps_test.go"
  - "cmd/d2bot/hotkeymapwire.go"
  - "cmd/d2bot/hotkeymapwire_test.go"
  - "data/input/caps/d2d-fork-input-items.json"
  - "internal/env/tuning_input.go"
  - "docs/process/sandbox-input-items.md"
exclusive: []
contracts: []
result: "Fork branch didle/input-items-20261120 (tip 81195ff04, pushed to the private fork): item equip/unequip/move/identify/drop and gold deposit/withdraw are ordinary input after tests/input_items_diff_d2d.py, 19 steps, 0 differences (verb vs message sequence and pointer path, builder, V); merc verbs, labels and cheats stay lab. Bot: internal/hotkeymap generator and read-back (fake binder, Paladin and Sorceress, must-fail cases), knob input.hotkeyMap default 0; input.item-* caps adopted only behind input.planner. Not verified live: no sandbox image, the hotkey route on a real game, the town portal row name."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Q13 follow-up (ADR 0021). Done when the fork differential, the generator and the knob are merged; live check on a sandbox image is a separate go.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork: item equip/unequip/move/identify/drop and gold deposit/withdraw promoted to input after a 0-difference differential
- [x] #2 Bot: hotkey map generator from planner slots, written and read back (fake), knob input.hotkeyMap default 0
- [x] #3 Bot: input.item-* caps adopted behind input.planner only
<!-- AC:END -->
