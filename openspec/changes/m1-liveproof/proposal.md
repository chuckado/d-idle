## Why

Policy work (trips, kit, gold floor, early game) sits on basic actions that were only tested against fakes: the healing belt is empty in about a third of the minutes of the long soaks, ground potions are skipped as "already in the pack" while the belt is empty, town portals fail thousands of times, and several specs are tagged UNVERIFIED (skill binding, meter totals, the portal route of town interact, waypoint row choice, belt rows by belt type). docs/primitives.md lists 36 primitives with a definition of correct each. docs/basics-audit.md lists 10 live checks and 2 unknowns (starting gold, unidentified sell prices) that no offline test can close. This card becomes the live drill suite that proves them on a hidden d2d with throwaway characters, before policy work resumes.

## What Changes

- A drill harness (`cmd/d2drill`, scripted scenario runner) and one drill per primitive group, run on one hidden d2d slot; output is a table P-id, drill, setup, pass condition, measured, verdict (PASS, FAIL, BLOCKED), cheat ledger ids.
- Setup only through the logged internal/cheat tool or cmd/d2cheat, each with a reason, on throwaway characters, hidden instances, offline only. Gate characters are never used.
- End states are read through the StateReader (state, items, belt, UI flags, level info, unit table), never from the bot's own log lines.
- docs/primitives.md gets a verdict column filled per drill; docs/basics-audit.md checks 1 to 10 and the two unknowns are run as drills D2 to D4 and recorded in docs/basics-live.md.
- Acceptance: every primitive with criticality S or A in docs/primitives.md has a live drill PASS, or a FAIL with a fix card, before policy cards (goldfloor follow-ups, m2-*, m3-*) resume.
- No production code in this card except the harness; the harness is a separate implementation task (see tasks).

## Dependencies

After m1-audittests, m1-stashcap and m1-trips are archived (it proves their fixes) and after xpgap releases soak-slot:3. Needs the fork verbs of docs/d2d-town.md "Fork verb requests" (click modifiers, `debug wp`) for D1 and D2; a drill blocked by a missing verb reports BLOCKED, not PASS. Cheats needed: gold, level, stats, heal, quest, waypoints (all exist in cmd/d2cheat); durability and "kill merc" are not and are named in tasks.
