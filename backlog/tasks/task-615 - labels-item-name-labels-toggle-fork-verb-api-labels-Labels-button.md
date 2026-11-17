---
id: TASK-615
title: 'labels: item name labels toggle (fork verb, /api/labels, Labels button)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T14:36:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 310000
card: "labels"
owns:
  - "internal/d2dctl/labels.go"
  - "internal/d2dctl/labels_test.go"
  - "internal/watch/labelsapi.go"
  - "internal/watch/labelsapi_test.go"
  - "cmd/d2bot/labelswire.go"
  - "cmd/d2bot/labelswire_test.go"
  - "internal/watch/web/panels/video"
  - "internal/watch/testdata/video_labels_panel.js"
  - "tools/uishots/fixtures_commands.go"
  - "docs/video-feed.md"
  - "internal/watch/funwire.go"
  - "internal/watch/web/shell.js"
  - "internal/watch/watchlayout_test.go"
  - "tools/uishots/serve.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/d2dbackend.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user wants to switch the D2 ground item name labels (Alt held) on and off from the watch page. Done when: the fork verb, the API, the per-hero desk and the Labels button exist, offline tested; live check on the sandbox is a checklist in docs/video-feed.md section 14.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fork verb labels on|off|toggle|status with caps ui.labels, headless test, pushed to didle/labels-verb
- [ ] #2 d2dctl.Labels, GET/POST /api/labels, per-hero desk re-applied at game entry
- [ ] #3 Labels button on the video card and a menu entry, node tests, fit pictures viewed
- [ ] #4 docs/video-feed.md section 14
<!-- AC:END -->
