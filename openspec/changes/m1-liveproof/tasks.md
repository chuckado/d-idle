# Tasks

```yaml
change: m1-liveproof
status: in-progress
agent: drills
branch: agent/drills
worktree: ~/code/d2lod-wt/drills
owns:
  - docs/basics-audit.md
  - docs/basics-live.md
  - docs/primitives.md
  - docs/probes/basics/**
  - cmd/d2drill/**
  - internal/drill/**
  - docs/probes/drills/**
exclusive:
  - soak-slot:0
gates:
  - "docs/primitives.md: every primitive with criticality S or A has a drill verdict PASS (or FAIL with a named fix card) in its evidence column, checked by the coordinator by eye"
  - "docs/basics-audit.md coverage table has 0 G rows and every S1 or S2 row is C with a test name in the offline-test column and a live proof line (`grep -c '| C ' docs/basics-audit.md` is at least 15)"
  - "starting gold and unidentified sell prices recorded V in docs/basics-live.md with cheat ledger ids"
  - "tools/agentreap.sh m1-liveproof leaves nothing; changecheck -cli 0 warnings; leakcheck 0"
merge:
result: "D0+D1 live 2026-10-31 (docs/probes/drills): P7a PASS (engine auto-belts a ground potion; same kind goes above in its column, new kind lowest free column), P7 PASS on end state (belt gets both, pickit KEEPs, but the looter reports picked 0: bookkeeping on d2d), P1 FAIL (F8 reproduced: lowest-free-box fill puts hp in mp columns), P2 PASS (5 belt types, reader boxes = engine boxes: 8/8/12/12/16), P3 PASS; BLOCKED P1s (click modifier verb) and P3m (merc cheats); each drill twice, identical, also across two processes; P35 gate live-tested (D0). Not verified live: nothing of D2 to D9."
```

Drill design (one scripted scenario runner on a hidden d2d slot): a drill is `setup` (cheat steps, each `cheat.Do` with a reason and the drill id as actor), `act` (the bot primitive under test, called directly, not through the goal loop), `read` (StateReader snapshot before and after), `assert` (numeric pass condition from docs/primitives.md). The runner runs `ui-clean` (P35) before and after every drill and aborts the drill group on a dirty UI. Each drill runs 3 times (10 for P10, P11, P22) and passes only if all repeats pass (P10, P11, P22: at least 9 of 10). Output: `docs/probes/basics/<drill>.jsonl` plus the table in docs/basics-live.md. Each run under 30 virtual minutes, started through tools/d2dslot.sh under tools/agentrun.sh; hidden instance only; throwaway characters only; never a gate character.

## 1. Harness (1 agent-day)

- [x] 1.1 `cmd/d2drill` and `internal/drill`: scenario type, cheat-step wrapper, StateReader before/after, result table, ui-clean pre/post check (P35); offline tests with mem.Fake for the runner itself
- [x] 1.2 Name the missing cheats and verbs in the report (durability, merc death, `debug wp`, click modifiers); drills that need them report BLOCKED

## 2. Drill groups (each at most 1 agent-day; order = priority)

- [x] 2.1 D1 belt: P1 P2 P3 P7 (fill from pack, belt rows per belt type, drink per column and merc, ground potion with empty belt, shift-click and ctrl-click behaviour recorded)
- [ ] 2.2 D2 portals and town entry: P10 P11 P12 P14 P23 (scroll and tome, 10 casts, return, tome fill, every Act 1 NPC window, waypoint row choice on 3 activated waypoints)
- [ ] 2.3 D3 vendor and gold: P15 P16 P17 P13 P5 P8, plus the two unknowns (starting gold on 2 classes, unidentified sell prices)
- [ ] 2.4 D4 pack: P4 P9 P6 (placement with fragmentation, pack-full pickups once per item, cube recipe)
- [ ] 2.5 D5 combat: P26 P27 P28 P29 (BindSkills count and cast effect, walled target, threshold drinks with meter totals, retreat then TP)
- [ ] 2.6 D6 movement: P22 P24 P25 (10 walks, 5 exits, pocket escape)
- [ ] 2.7 D7 build: P30 P31 P32 (level cheat, points, equip swap with empty cursor)
- [ ] 2.8 D8 NPC and travel: P18 P20 P21 P19 (merc claim, hire and revive, Akara reward, Warriv, no gamble when poor)
- [ ] 2.9 D9 lifecycle and policy re-gate: P33 P34 P36 (10 leave and enter cycles, death and corpse, paired 8 seed trip and empty-belt check as in goldfloor)

## 3. Close

- [ ] 3.1 Fill the verdict per P-id in docs/primitives.md and update docs/basics-audit.md (rows to C with test and live proof line); open a fix card per FAIL
- [ ] 3.2 changecheck, leakcheck, reap
