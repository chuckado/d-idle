# fixtalk: the bot's town talk code in a package the drills can import

> status: offline refactor done and tested; D8.q and D8.w are built on the bot's path, awaiting live proof. Card: openspec/changes/fixtalk.

## Why

D8.q (Akara reward, P20) and D8.w (Warriv act hop, P21) were PARTIAL in DRILLS5 to DRILLS9 (docs/probes/drills/README.md): the talk nodes were package main (cmd/d2bot), so the drills could only prove the game rule (`debug quest set` plus a bare menu pick), not the bot's own code path.

## What moved where

Wiring (unchanged): the task nodes (internal/task, for example act1 "warriv" `StepTalk(NPCWarriv, "go east")`) call `TaskDeps.Dialog.Talk(npc, pick)`; `cmd/d2bot/task.go newTaskDeps` supplied the dialog adapter. The act hop is a task node plus one Talk row, not a separate decision function (the Hop tables in internal/quest are data).

| was (cmd/d2bot) | now (internal/talk) |
|---|---|
| `talkPaths`, `actTalkPaths` (acts.go, task.go) | `Paths` (tables.go) |
| `npcByAct`, `npcClassIn`, `npcFara` and the other extra NPC names | `ByAct`, `ClassIn`, `NPCFara` ... |
| `townDialog` (quest.Dialog over the Townsman) | `townDialog` (talk.go) |
| `dialogAdapter`, `talkOnce`, `akaraRespecSeen` | `Adapter`, `talkOnce`, `RespecSeen`/`SetRespecSeen` |

The game I/O sits behind `talk.Services` (`OpenForTest`, `PickLine`, `Socket`, `Imbue`), implemented by `*town.Town`, and `env.Env` (snapshot, clock, `In.CloseAll`). cmd/d2bot keeps `newTaskDeps`, which builds `talk.Adapter{T: tw, E: e, Status: rd.status}`.

## Identical behaviour

The function bodies were moved by script, with only renames (`talkPaths` to `Paths`, `d.t`/`d.e` to the exported fields). No table entry, row label, count, timeout (700 ms after a click, 2.5 s or 8 s speech wait, 3 reward tries, the flip on the second try) or log line was edited. cmd/d2bot builds and its existing tests (which now reach the tables through `talk.*`, including the check that every Act 2 to 5 node resolves in the adapter) pass unchanged in meaning. Nothing in `go.mod` changed.

## Tests (internal/talk/talk_test.go)

Fakes: `Services`, `env.Env` with a fake reader (skill points), fake input (close count) and a fake clock that advances on Sleep. Cases: `ClassIn` per NPC and act (including the missing-entry cases), every `Paths` entry opens the right class and clicks one line per path label and closes once, per-NPC rows (Akara quest/bark, Kashya, Charsi, Warriv, Meshif), the row index of Warriv's Go East against `quest.RowIndex`, unknown NPC and pick errors, a failed open closes the dialog, the Larzuk socket and Charsi imbue pickers, the Akara reward retry (paid on try 1, 2, 3, never) and the respec line count flipping on the retry.

## The drills (harness edits: cmd/d2drill/field_lab.go and internal/drill/dgame.go only)

`lab.AkaraReward` and `lab.Warriv` still set the record with `debug quest set` (logged cheat), then call `talk.Adapter.Talk` (via `lab.talker`) with the bot's game-time Pause installed, instead of `OpenNPC` / `OpenNPCMenu` + `PickMenuLine`. The quest status given to the adapter is fixed per drill (`drillQuestStatus`: Den, Andariel and Act 1 done, nothing else started), matching the cheat. The Pass reasons say built, awaiting live proof. The QuestLab interface did not change.

## Live recipe for a later DRILLS agent (one hidden slot, newest cheats build, a fresh throwaway in Rogue Encampment)

1. `tools/live.sh <agent> d2drill -group D8` (or `-drill D8.q`, `-drill D8.w`; D8.c first, it needs a character that never hired a merc).
2. D8.q: the drill runs `debug goto 1`, `debug quest set 1 0x0002`, then the bot's Talk(Akara, "reward"). Expect skill points +1, `debug quest list` record 1 at 0x0001, dialog closed. A retry line `akara reward not paid, retrying` is a finding, not a pass.
3. D8.w: `debug goto 1`, `debug quest set 6 0x0001`, then Talk(Warriv, "go east"). Expect area 40 after 200 ticks. Needs the Warriv row row index (Go East sits above Cancel); a row error shows as `quest: row not present` and is the UNVERIFIED label text.
4. Run each twice (P35 gate around each), leave the character in town with no panels open. Update primitives.md rows P20/P21 and the README B2 row from PARTIAL to PASS only with the DRILL lines.
