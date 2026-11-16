---
id: TASK-537
title: 'friendfb: friend feedback 0.1.3, Items tab, Help card, skills size'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T03:31:00Z'
updated_date: '2026-11-16T04:50:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 284000
card: "friendfb"
owns:
  - "internal/watch/dropsfeed.go"
  - "internal/watch/dropsfeed_test.go"
  - "internal/watch/logtab.go"
  - "internal/watch/cardtext.go"
  - "internal/watch/web/skills.css"
  - "internal/watch/web/panels/log"
  - "internal/watch/web/panels/onboarding"
  - "data/onboarding/en.json"
  - "docs/ui/friend-feedback-2026-11-15.md"
  - "tools/uishots/fixtures_skills.go"
  - "tools/uishots/serve.go"
  - "tools/uishots.sh"
  - "internal/watch/views.go"
  - "internal/watch/funwire.go"
  - "internal/watch/web/shell.js"
  - "internal/watch/testdata/logpanel"
  - "internal/watch/testdata/onboarding_dom.txt"
  - "internal/decide/firsts_act1.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: friends (release 0.1.3) found the firsts popup unclear, the Skills page oversized, no drop list, no explanation of what the bot does. Done when: see docs/ui/friend-feedback-2026-11-15.md (a to f), the checklist is ticked and the fit pages pass on the builder.

result: Firsts card reads as information with the first's own question, answers and a nothing-is-required footer; Skills board capped (pictures viewed at 1914x930 and 390x844); Items tab (GET /api/drops, default tab of the log drawer); Help card "What does the bot do for me?" (the bot never equips: gear-policy; no model, offline); Help opens the onboarding drawer in the cockpit. Verified offline (Go and node tests, tower gate, fit pass incl. the new demo-items and demo-help pages at all sizes, pictures of Items and Help viewed); not verified live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 First card says what happened and that nothing is required
- [x] #2 Skills board capped, pictures viewed
- [x] #3 Items tab (/api/drops) and Help card shipped
- [x] #4 fit pages demo-drawer-items and demo-help pass on the builder
<!-- AC:END -->
