---
id: TASK-530
title: 'charui: Characters page (UI for Character Select)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:34:00Z'
labels:
  - watch
  - feature
dependencies:
  - TASK-521
priority: medium
ordinal: 283000
card: "charui"
owns:
  - "internal/watch/web/characters.html"
  - "internal/watch/web/characters.js"
  - "internal/watch/web/characters.css"
  - "internal/watch/web/characters.json"
  - "internal/watch/charactersui_test.go"
  - "internal/charsel/rules.go"
  - "tools/uishots/fixtures_chars.go"
  - "docs/charselect-ui.md"
exclusive: []
contracts: []
result: "Characters page built; fit green; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Page /watch/characters: cards, switch, create, delete (typed name), restore, switching strip
- [x] #2 Entry points: the ... menu, Orders More, ?view=characters
- [x] #3 Fit fixtures at all sizes plus must-fail defects, node DOM test, goldens
- [x] #4 Gate and uishots fit green on the worker host
<!-- AC:END -->
