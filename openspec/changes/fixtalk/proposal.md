## Why

DRILLS5 to DRILLS9 left D8.q (Akara reward, P20) and D8.w (Warriv act hop, P21) at PARTIAL: the bot's own talk nodes (which menu row for which NPC and state, the Akara reward retry, the class of an NPC per act) were package main in cmd/d2bot, so the drills in internal/drill could not import them and only proved the game rule through `debug quest set` plus a bare menu pick.

## What Changes

- New package internal/talk: the tables (`Paths`, `ByAct`, `ClassIn`, the `NPC*` names) and `Adapter` (the former `dialogAdapter` and `townDialog`, the `akaraRespecSeen` flag) moved out of cmd/d2bot unchanged; game I/O stays behind the `Services` interface (`*town.Town`) and `env.Env`.
- cmd/d2bot/task.go and acts.go: the moved code is deleted, `newTaskDeps` builds `talk.Adapter`; the tests use `talk.*`. Nothing else changes.
- internal/talk/talk_test.go: table tests per NPC and menu state with fakes (class per act, every table entry, row index of Go East, Akara reward retry and line count, item pickers, error paths).
- cmd/d2drill/field_lab.go and internal/drill/dgame.go: D8.q and D8.w call `talk.Adapter.Talk` (the bot's code path) after `debug quest set`; marked built, awaiting live proof.
- No policy, numeric or behaviour change (no delta spec). Docs: docs/fixtalk.md.

Overlaps (declared): cmd/d2drill/** and internal/drill/** belong to the active card m1-liveproof (docs/OWNERS); this card edits only cmd/d2drill/field_lab.go and internal/drill/dgame.go there (the D8.q and D8.w wiring). No other open card owns cmd/d2bot/task.go, cmd/d2bot/acts.go or internal/town files.
