---
id: TASK-455
title: 'uis2: The glance endpoint and the band'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T22:39:00Z'
updated_date: '2026-11-14T22:39:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uis2"
owns:
  - "internal/watch/glance.go"
  - "internal/watch/glance_test.go"
  - "internal/watch/testdata/glance/"
  - "internal/watch/web/panels/statusline/"
  - "internal/watch/web/panels/needs-count/"
  - "internal/watch/web/panels/hero-meters/"
  - "internal/watch/testdata/glance_dom.js"
  - "internal/watch/testdata/glance_dom_*.txt"
exclusive: []
result: "GET /api/glance (watch.glance/1) with sentence, needs, health, progress, optional journey and survival (words only); panels statusline and needs-count (cockpit-only, optIn, slot); hero-meters v2 slim mode; cockpitOnly manifest key keeps free-grid layouts unchanged. Goldens for both classes, DOM goldens, gate PASS and fit check at all 8 sizes on every page on the builder. Looked at renders 1440, 1280x620, 390, light. Critique: desktop band matches the mockup; phone band is tight (sentence nearly touches the nav); meters show no bar fill states beyond life tone. NOT verified live; needs/journey/merc/supplies/gear/voice/paused/survival have no bot wiring yet (setters in GlanceSources)."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Handler goldens testdata/glance/view-{daily,needs,blocked,paused,dead,unknown}-{sorceress,paladin}.json
- [x] #2 For every code in whychange.json the sentence has a non-empty reason (table test)
- [x] #3 DOM goldens of statusline (ok, warn, danger, info, loading, error; long reason cut) and needs-count (0, 1, 3)
- [x] #4 The band is 44 px at the three desktop sizes and two lines on the phone; fit check passes
- [x] #5 Classic and the other free-grid layouts are unchanged
<!-- AC:END -->
