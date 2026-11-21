---
id: TASK-784
title: 'skillsbadge: Skills page kind badges and filter'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T19:05:00Z'
updated_date: '2026-11-21T19:32:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 361000
card: "skillsbadge"
owns:
  - "internal/skilltree"
  - "internal/watch/web/skills.js"
  - "internal/watch/web/skills.css"
  - "internal/watch/skillsui.go"
  - "internal/watch/skillsjs_test.go"
  - "internal/watch/testdata"
  - "tools/uishots/fixtures_skills.go"
  - "docs/skilltrees.md"
exclusive: []
contracts: []
result: "Skills page: Node.kind (aura flag, else passive flag, else active, from skills.txt; V on the real table) with a badge strip (glyph+word) per tile, All/Active/Aura/Passive chips with counts stored in localStorage, legend line. Real counts: Paladin 10 active/20 aura/0 passive, Sorceress 26/0/4. Go and node stub-DOM tests; headless screenshots at 412x915 and 1366x768 on the worker host (after only). Not verified on the live hero."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Kind from skills.txt flags, documented
- [x] #2 Badges and filter chips on the Skills page, phone safe
- [x] #3 Go and node tests
<!-- AC:END -->
