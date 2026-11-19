---
id: TASK-453
title: 'uis4: Orders, read-only, as a slot panel'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:36:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uis4"
owns:
  - "internal/watch/commands.go"
  - "internal/watch/commands_test.go"
  - "internal/watch/testdata/commands/"
  - "internal/watch/web/panels/orders/"
  - "internal/watch/testdata/orders_dom.js"
  - "internal/watch/testdata/orders_dom.txt"
  - "tools/uishots/fixtures_commands.go"
exclusive: []
result: "GET /api/commands (watch.commands/1, commands.go) joins the watch orders, the order queue, the loop, the repair desk, the item queue, the gold farm and the plan into rows with one state vocabulary, the bot's ladder order, one next mark, a cancel route per row; and the slot panel orders (panels/orders, slot true, scroll true, optIn, fit) with verb row (inert, S5 writes), safety line, 28 px rows, blocked rows, plan row, a list that scrolls inside (data-scroll), cancel cross to the existing route. Tests: handler goldens 5 states x 2 classes plus cancel-all, state table, cancel routes exist as POST routes, ladder/next, no duplicate rows, race test, node DOM golden (all fixtures, 8 rows, no connection, route off) and the cancel node test (order, queue, loop, repair, item, set group, gold farm). Verified on the builder (not live, no game): whole gate TOWERGATE PASS (gofmt, check.sh, go test -short, race on ./internal/watch/..., leakcheck 0 hits) and tools/uishots.sh --fit: 168 ok, 0 FAIL at all 8 sizes (the real cockpit, the real Orders with 3 rows, 8 rows, empty, safety, blocked, all-sources, and the demo rail with the real panel; crafted must-fail pages still fail). Looked at renders (builder, headless Chromium) 1366x768, 1440x900, 1366 with 8 rows, 1280x620, 390 and 1920. Critique: (1) at 456 px (1440) the reason column is still cut on long rows; fixed by smaller type, a 6rem floor for the reason and shorter reason texts, still tight for titles over 25 characters; (2) with 8 rows the list scrolls inside and cuts a row in half at the edge, acceptable, no fade yet; (3) the verb buttons are inert and look live (title says so): S5 must wire them; the Go to button is the one gold accent. Not verified: a live bot (the join over real desks is tested with fakes and the sim hub; real inventory/campaign JSON shapes are read from the Go structs, S), pause (no pause desk is wired: Fun.Cmd.Paused nil, so no Pause verb on a real bot), journey go-now rows (the campaign view has no open go-now request, U), blocked rows come from the loop and Fun.Cmd.Blocked only. Edits outside the owned paths (one-liners): funwire.go (Fun.Cmd field, route mounted in mountView), layout_test.go/panels_test.go/layout_s1.js (the first built opt-in panel: allow-list, route list, fallback table skips built panels, classic mounts every non-opt-in panel), tools/uishots.sh, fixtures.go, serve.go (orders fit pages, cockpit-orders fixture layout, fixture /api/commands). Based on agent/uis1b; not queued until S1b is on main."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the Orders panel of the cockpit rail (J section 2, C-orders.md): one list of the player's commands in the bot's own order with one state vocabulary, instead of five queues in five places. Read-only: the verb row is drawn but inert (S5 adds the write side); the cancel cross posts to the route the row's own source already has. Done when: the criteria below are ticked and the result says what was verified. Based on agent/uis1b (S1b is not on main yet); one-line edits outside the owned paths are listed in the result.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Handler goldens testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json
- [x] #2 A table test: every source state maps to exactly one row state
- [x] #3 DOM goldens for the five fixtures; with eight rows the list scrolls inside and the fit check still passes at all 8 sizes
- [x] #4 Cancel of an order, a queue entry, a repair and an item action each reaches its existing route (node test)
- [x] #5 A race test over concurrent desk updates and reads passes under -race
<!-- AC:END -->
