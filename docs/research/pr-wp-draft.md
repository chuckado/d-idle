# PR draft: waypoint refusal window (agents DECOMPWP, WPPR, 2026-10-31)

> status: DRAFT for a LOCAL branch; not pushed, not posted. Branch `fix/waypoint-hostile-window` on the worker host (74199f4f tooling, 4724e138 claim, on upstream/main d0ea14b7). Proof: [pr-wp-proof.md](pr-wp-proof.md). Evidence: [pr-wp-evidence-pack.md](pr-wp-evidence-pack.md).

**Title:** world: the waypoint refusal counts from hostility, not the level change

````
## What changes
One claim, one commit (plus one tooling commit), on the 0x49 command in `World::deal`:

A waypoint used right after arriving on a level now travels at once, as game.exe does. d2d refused it for 10 s after every arrival (`arrived_at`, stamped in `World::swap_npcs`). game.exe's handler FUN_0054c5d0 refuses while `GetTickCount() < player data +0x160 + 10000`, and the only code that writes player data +0x160 is FUN_0055b720, called only from the hostile declaration FUN_005a5e50: in single player it never fires. d2d has no hostile declaration, so the test and `arrived_at` go. The other refusals stay (not a waypoint, the current level, an inactive target, a joined game). act1-end.md and waypoint.md now say what the field is; tests/pad_use_d2d.py loses its 10 s sleep.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0054c5d0` | `FUN_0055b6c0` (edx 0), `add eax, 0x2710`, GetTickCount, `jbe`: else voice 0x13, `FUN_00554190`, return 1 |
| `FUN_0055b6c0` | with no second unit: player data +0x160 |
| `FUN_0055b720` | the only writer of player data +0x160 (0x55b76b, `= GetTickCount()`); every other `+0x160` store in .text is a client struct; player data (`FUN_00621f90`, 0x16c bytes) starts zeroed |
| `FUN_005a5e50` | `FUN_0055b720`'s one caller (0x5a5f51): the hostile declaration (C->S 0x5d / 0x5e, a second player, both clvl >= 9) |
| `FUN_00584870` | the town portal's 5 s on the same field (objects.md already calls it the hostility time) |

## How to re-check

```sh
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure      # test_game needs D2_MPQ_DIR and D2_PATCH_INSTALLER
# the oracle: game.exe's own FUN_0054d750 -> FUN_0054c5d0 -> FUN_0055b6c0 / FUN_00554190 (hooks in its docstring),
# 40,002 grid cases (0..20,000 ms since the last level change x target active) + 100,000 random single-player timelines,
# against d2d's World::apply(cmd::Waypoint) after real World::arrive level changes:
cd tools/emu && uv run python wp_refuse.py gen 100000 7 --solo > cases.txt && uv run python wp_refuse.py game cases.txt > g.txt
../../build/tests/test_game --waypoint-refuse cases.txt | grep -E '^[0-9]+ ' | diff g.txt -    # empty (66,969 lines differed before)
```

`tools/emu/wp_refuse.py` is new tooling (CODEOWNERS review). Without `--solo` the cases include hostile declarations: game.exe then refuses for 10 s and d2d, which has none, does not (the joined / PvP path is not modelled).

Limits: game.exe also refuses in the first 10 s of machine uptime (GetTickCount < 10,000 against the zeroed field); d2d does not, and the cases start at 70,000. Existing gaps unchanged: the target check's range and same-act tests and the dispatcher's busy / dead drop are not in d2d. An act change still arrives at once (game.exe's load not traced: ponytail).

Presubmit: siblings: only players use waypoints (mercs, pets, monsters do not); "never" in waypoint.md: single player has no hostile declaration (FUN_005a5e50 needs a second player); build and ctest run by hand (failures test_mpq, smoke_d2d, quests_act1, quests_act4 are the same on main).

## Screenshots
Not applicable.

## Checklist
- [x] A test in `tests/` (test_game: a waypoint right after a level change travels; not to its own level or an inactive one)
- [x] The finding in `docs/research/re/` (act1-end.md, waypoint.md)
- [x] `python3 tools/pr_guard.py` passes

---

Agent: Claude Code
````

## Ponytail (ours, not in the body)

- The bot's `WpCooldownSec` / `waitCooldown` (internal/nav/waypoint.go, internal/env/tuning_nav.go; finding W3 in waypoint-trace.md) copies d2d's post-arrival 10 s. Once this PR is merged and the lab binary carries it, the wait is unnecessary on d2d and wrong for game.exe single player: set the knob to 0 then (follow-up in docs/BACKLOG.md; the bot is not changed by this card).
- Bundle: 9 files; joins the next bundle after #149 / #152 merge (docs/research/pr-bundle-queue.md), or opens alone with the user's go.
