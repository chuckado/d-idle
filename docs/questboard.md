# Quest board model (slice S1)

> status: built and tested offline (card quests1, TASK-577); not wired: no bot code imports it until slice S3. Design: [docs/design/quest-board-2026-11-16.md](design/quest-board-2026-11-16.md). Tags: V verified by the named test or file, S secondary, U unverified live, I design choice.

## What is here

| piece | where | notes |
|---|---|---|
| definitions | `data/questboard/act1.json`, `act2.json`, `families.json`; embedded by `data/questboard/embed.go` | schema `quests.def/1`; strict decode (an unknown field is an error) |
| loader, validation | `internal/questboard/def.go`, `load.go` | unknown class, act, quest, chunk, difficulty, kind, risk: an error naming it (V `TestParseErrors`, `TestLookupsNameTheUnknown`); no silent fallback |
| states, derive | `internal/questboard/state.go` | nine states, `Derive`, `CanStart`, the `Transitions` table with the evidence tier each needs |
| plan records | `internal/questboard/plan.go` | per hero `<dir>/<hero>.quests.json` through `internal/store.File` (tmp, fsync, `.bak`, quarantine, future version refused) |

Quest ids are the save ids (`internal/saves` questTable; V `TestQuestIDsAreSaveIDs` reads that file). `taskId` carries the task-layer id where it differs (`sisters_burial`, `sisters_to_slaughter`).

## Rules the code enforces

- Derive is a pure function of its input: derive twice, same result (V `view` helper in every test). Handler progress is never stored; the only store is the player's decisions (accepted, skipped, auto-start opt-ins, seen toasts).
- Evidence: a quest prerequisite unlocks only on T1 (save) or T2 (memory); T3 (proxy) never unlocks (V `TestTransitionTableEvidence`, `TestChunkOrderAnyThenAssemble`). T3 may stop the hero at the NPC and show "done, confirming"; it pays nothing until T1/T2 confirms.
- Evidence conflict (I): for a single-chunk quest a fresh T2 "not done" beats a T3 "done"; a stale T1 "not done" does not beat a proxy (shown as confirming). A part of a multi-chunk quest has no bit of its own, so its proxy stands.
- Chunk order (I): a chunk's `after` is satisfied by any evidence of the earlier chunk (an offer skips no work, the handler re-derives); a chunk's or quest's `requires` need T1/T2.
- Start tap: `CanStart` refuses anything that is not `ready` (locked, out of order, available, done); `Store.Accept` refuses a second accepted chunk in the same difficulty.
- Readiness: S2 is not built. Without a readiness word a chunk stays `available` ("readiness not known"); the tests stand in for S2 with the legacy floor per class family (clvl by melee, caster, summoner) on Paladin and Sorceress.
- Meta currency (design section 7): `Award.Key` is `quest/<quest>/<chunk or ->/<difficulty>`, never the hero (once per account per quest per difficulty). A quest the board never started (`byHand`) pays nothing: no backfill (user decision 2026-11-16). T3-only completion pays nothing yet. The amounts are data placeholders (U) in the `meta` columns, about 5 for a first Normal story quest and 2 for an optional chore.
- `autoStart: never` rows (hazard risk) refuse the opt-in (V `TestStoreSkipAndAutoStart`); the loader requires `autoStart never` on every `hazard` quest.
- Build flexibility: no class or skill literal in Go; the class-to-family table is `families.json`, the floors are the `legacyFloor` column per quest.

## The data and where it came from

Act 1 merges `data/quests/act1-hints.json` (giver, clvl floors, EXP/UNSURE tags), `act1-givers.json` (accept and hand-in picks) and `docs/goals/execution.md` section 3 (steps, boss, rewards, Req); the per-field `tags` map names the source of each field. Act 2 has no hint or giver table, so it comes from execution.md section 3 and docs/progression/quests.md alone and every row is U until a live drill. Chunks: Cain's chain is four (scroll, read, free Cain, report), the Horadric Staff is three (cube, shaft in any order, assemble after both and the Tainted Sun amulet). The old hint and giver files stay until S9 retires them.

Known gaps (U): the hand-in picks of Akara, Atma, Cain, Charsi (`imbue`), Warriv (`Go East`) and Meshif (`Sail East`) are not in `internal/talk/tables.go` as read here; `minutes` are my estimates (I); the Andariel `softAfter` the Burial quest is the hint table's UNSURE availability, kept soft on purpose so it cannot lock a player out; whether Radament gates the Staff (execution.md says V, kept as a hard `requires`).

## For S2 to S4

S3 builds `Facts` from `quest.Reader` (`Info.Status`, `Pending`, `Evidence` map to `Fact` and `Tier` by value: `EvProxy`=1, `EvSave`=2, `EvMemory`=3) and `ChunkFact` from the handler checks; S4 calls `CanStart` then `Store.Accept`. No knobs were added (the design's `quests.*` knobs belong to S3).

## Readiness (slice S2)

> Built and tested offline (card quests2, TASK-586). Pure query: no bot code imports it, `bot.zoneModel` stays 0 and nothing here changes zone choice. Tags as above.

| piece | where |
|---|---|
| thresholds (data) | `data/questboard/readiness.json` (`quests.readiness/1`): dial lines `enter`/`stay` per cautious/balanced/bold, kit minimums (`minPotions`, `minResist` per difficulty, U), default dial |
| core | `internal/questboard/readiness.go`: `Thresholds.Assess(AssessInput) Readiness`, reason codes, the sentence |
| power adapter | `internal/questboard/readiness_power.go`: `HeroFrom(power.Hero)`, `EncounterFrom(name, kind, power.Result, k)` (k = survival ledger `KZoneFor * KHeroFor`, the same division the controller does) |

- Word rule (V, same as `survival.word` and `power.ZoneMargin.Word`; a test compares them): `safe` when the Low margin is at or over the dial's `enter`, `risky` when Best is at or over `stay`, else `too hard`. The word of a chunk is the worst of its encounters (the boss, and the zone on the way when the caller passes it). The dial lines mirror `internal/power/profiles.json` (V `TestDialsMirrorThePowerProfiles`).
- `Readiness{Word, Reasons, LegacyFloor, Unknown, FromLegacy, Sentence, Debug}`; the margin is only in `Debug` (tools), never in the sentence. The S1 string type `Readiness` is now `Word` (`ReadySafe`, `ReadyRisky`, `ReadyTooHard`, `ReadyUnknown`); `Input.Ready` maps chunk keys to `Word`.
- Reason codes (closed): `level_under`, `no_merc`, `kit_potions`, `kit_resist`, `margin_low`, `margin_wall`, `offence_unread`, `legacy_only`, `no_data`. Order: fixable first (merc, kit), then level, then the margin, then offence. The level number appears only in `level_under` and is the legacy floor (user asked for it beside the margin).
- Legacy floor fallback (I): where no encounter has a margin (boss missing from the tables, hero unread), the old clvl floor of the quest's family decides: at or over the floor `safe`, under it `risky`, never `too hard` (warn, never wall). `FromLegacy` is true so a later auto-start can refuse to trust that `safe`.
- Unknown: no margin and no floor gives `Unknown`, an empty word, reason `no_data`; S1 keeps the chunk `available`, never offered (V `TestUnknownReadinessKeepsQuestAvailable`).
- Tests (V, `internal/questboard/readiness_test.go`): Paladin and Sorceress on made-up zones sized to the logged outcomes (Cold Plains clvl 4 and 5 safe, Black Marsh clvl 4 too hard, clvl 12 with a merc safe, without a merc risky; docs/survival/README.md). The Sorceress rows are regression rows, not logged outcomes. A synthetic `survival.Ledger` with two merc deaths and a death in Black Marsh turns a safe word into a harder one. U: real zone and boss tables are not used (the install is on the worker host); no live run.
- Not done (for S3): choosing the encounters per chunk. Quest rows carry only a `boss` id, not an area; S3 resolves `boss` through `danger.Table.FindBoss`/`power.BossPack` and passes the zone of the route, plus the dial from the hero's setting.
