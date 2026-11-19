# Sandbox d2d build devp13 (2026-11-19)

> status: current. Task: forkrebase (first round of `docs/process/fork-rebase.md`). Tags: V = verified in this run (how said), S = read from a tool or log, U = unverified. Image built and LOADED on the sandbox host only; the running container was not touched (it runs `d2dbot:devp12`, S: `docker ps` on the sandbox host after the load).

## What devp13 is

Our carried patches (item / vault / lab verbs, `inv_tint` / `inv_colour` / `inv_pic`, labels, merc verbs, dump stats, debug presets / objects / portals / warps, lockstep, framefeed, cursor, overlay: 50 commits) rebased onto upstream main `da98b7f89` (2026-11-19, PR #629), plus one new fork change (`item create ... identified=0|1`, below). devp12 sat on upstream `f15337707` (PR #597, 2026-11-17).

- Fork branch `didle/rebase-20261119`, tip `f32837d8` (the rebase itself ended at `89ea1d19`; `f32837d8` adds the identified flag), pushed to the private remote the same day. V (`git push` output; the worker's fork checkout fetched it as the ref `devp13` and `rev-parse` gave `89ea1d19`). The previous branch `didle/rebase-20261118` is untouched.
- Image `d2dbot:devp13`, id `97e985ca5eb5` (an earlier build of the tag, `4039ae497734` from `89ea1d19`, was replaced by it), loaded next to `devp12` (`a064b241ba61`) and `devp10`. V (`docker images` on the sandbox host). The d2d binary in it has sha256 prefix `ae7f490584213a56`, the same as the builder's output, and `d2d --help` lists `--item-verbs`, `--vault-verbs`, `--lab-verbs`. V.
- Build: on the builder (first build `build-fork.sh` 532 s; the final build on the long lane, d2d only, -j3, about 10 minutes including its tests), rc 0 (no compile error: the one semantic conflict is in the notes below). V.

## New fork change: `item create ... identified=0|1`

Requested for the vault reveal (user decision: Ether replaces the scroll, so a revealed vault item is created identified on Keep). Grammar: `item create <hex> [stash|inv] [col row] [key=<token>] [identified=0|1]`. `identified=1` sets the item's identified flag after the parse (the affixes are in the record either way, so `item full` and `dump items` show them), `identified=0` clears it; absent, the record's own flag stands (default unchanged); any other `identified=` value is `err parse identified=0|1`. Code: `apps/d2d/devctl/vault_verbs.hpp` (behind `--vault-verbs`; fork only, never upstream without the user's go). Docs: `docs/control_channel.md` of the fork. Test: `tests/item_create_identified_d2d.py` (new, wired in `tests/CMakeLists.txt`): an unidentified magic and a rare long sword are captured with `ground full`; default create stays unidentified (`id'd=0`), `identified=1` gives `id'd=1` with the affixes visible and the item listed by `dump items`, `identified=0` and a bad value behave as above. Ran rc 0 on the final build. V. Bot side: pass `identified=1` on Keep; nothing else changes.

## Conflicts met (3 text, 1 semantic)

1. `apps/d2d/devctl/verbs.hpp`, the `cmd` usage string (repeated for 9 of our commits): upstream added `place <object> [item]`, we extended the line with `belt put`, `tobelt`, `body`, `drop`, `idwith`, `stashgold`. Merged both lists. V (the final string lists all; `item_verbs_d2d` drives them).
2. `apps/d2d/devctl/debug_fight.cpp`: our verb patch moved `states` into `debug_skills.cpp` (adds `states set|clear`) while upstream edited the old copy. Dropped the old copy.
3. Semantic, found by the compiler only: upstream removed `Monster::cry` (states list instead); our moved `states` listing used it. Ported upstream's change into `debug_skills.cpp` (`monster.states.empty() ? -1 : monster.states.front().skill` in the `cry=` column). The `cry=` column keeps its name; the value is the first state's skill id. U: a drill that reads `cry=` on a monster with two states sees only the first.

## What the rebase carries from upstream (diff vs devp12)

236 commits (174 non-merge) in 35 merged PRs, 255 files, +28,225 / -855 lines (S: `git diff --shortstat`). Wire verbs: nothing removed. What changed on the devctl layer (S: `git log` over `apps/d2d/devctl`, `docs/control_channel.md`, `main.cpp`):
- `debug kill [n] [spare <id>] [within <cells>]`: new `within`; kill takes only this level's monsters (4ca629761).
- `cmd place <object> [item]` documented (3129d9f28).
- Other `main.cpp` changes (74 lines) are UI and menu ports; no bot-visible flag.

Gameplay and rules that can move the bot's numbers (S: subject lines; not measured here):
- Cairn Stones portal and the Scroll of Inifuss read (#626): the stones' portal in the ring (offsets are subtiles), the scroll read through `Town::send`. The quest-1 Cairn step the bot drives can now meet the real portal.
- Summons (#627): summon level, defense, to-hit and Summon Resist follow game.exe; matters for a Necromancer build's skeleton and golem numbers.
- Curses (#603, #625): Attract refuses every curse state; a curse shares the monster's one curse slot (Taunt and Battle Cry too); monsters keep other states in their own nodes.
- Mercenary (#604, #615): a hurt merc regenerates its level-up life; a merc resurrect answers with S->C 0x2a.
- Combat and paths: melee reach (#605), gear absorb (#606), open wounds (#600), mana burn (#599), crush (#609), leech (#587), Whirlwind path type 7 round walls, Blessed Hammer spiral path, Nova targets (#583), aura pulse radius (#624), click-enemy pick (#598), ring targets (#621).
- Drops: a drop with no floor is not made or counted; free-spot rings stop under radius 50; no corpse past 15 with gear to the ground.
- Harrogath markers hand their DS1 paths to their NPCs.
Fork-side effect of our own PRs that merged upstream: none dropped by patch-id this round (the fork stack stays at 50 commits, S: `git rev-list --count upstream/main..HEAD`).

## Tests (on the builder, game data there; `tests/` of the branch)

Each ran as `D2_MPQ_DIR=<data> python3 tests/<t>.py <builder d2d>` (V, from `tools/fork-rebase-builder.sh test`).
Result (final build `f32837d8`, long lane): 10 of 11 pass (rc 0): `item_create_identified_d2d`, `lab_verbs_d2d`, `item_verbs_d2d`, `vault_verbs_d2d`, `item_colour_d2d`, `item_pic_d2d`, `labels_d2d`, `merc_verbs_d2d`, `cursor_verb_d2d`, `overlay_d2d`. V.
- `smoke_d2d` FAILS (line 444: after Save and Exit the double-click on the first roster row plays `Beta`, the test expects `Alpha`). It fails the same way on pure upstream main `da98b7f89` built without any fork patch (builder long lane, same game data, rc 1, same assertion; it fails the same way on the final build), so it is an upstream test or roster-order issue (char select orders saves by file time since 68fbb47c, the test's fake file times are in 2033), not a rebase regression. V. To report upstream if the user wants it (draft only, nothing posted).
- Harness note (first build `89ea1d19`): with `smoke_d2d` run first, `vault_verbs_d2d` failed twice with "a worn item is not vaulted" in the same container; without the failing smoke before it, vault passed 6 of 6 solo runs and in sequence. Cause not found (U); the script now runs `smoke_d2d` last.
- Upstream full `ctest` (all unit tests) was not run in this round: the main lane stayed busy with gates, the long lane carried the baseline build. U. Next round: run it in the long lane (docs/process/fork-rebase.md).

Bot side (V): `go test -p 1 ./internal/d2dctl/... ./internal/d2dsrc/...` ok (recorded replies, offline). `TestRealD2D` of `internal/d2dctl` and `internal/d2dsrc` (static test binaries) against a headless d2d with `--no-save --lab-verbs --item-verbs --vault-verbs`: both PASS, 0 unparsed lines. A d2dprobe session on the same d2d (ping, info, state, debug level / warps / presets / objects / portals, `debug mana full`, `debug give`, `debug equip`, `ground`, `dump items`, `dump stats`, `dump player`, `dump merc`, `dump vendor`, `dump ui`, `debug players`, `cmd move 40 40`) answered every step ok, 0 bad lines. Not run: a full `d2drill` or soak leg (needs a worker slot); `debug spawn` takes arguments I did not pass (usage error, not a regression).

## Payload examples (V: `ground full` after `debug give jew|rin|cm1|sbw` on a fresh `--no-save` hero (`--start-class 3`), seed 5, `--lab-verbs --item-verbs --vault-verbs`; raw and norm hex shortened)

```
#1 raw=.. norm=.. at=31.1,14.3 code=jew q=2 ilvl=30 ... children=0 inv_tint=0 inv_colour=-1 inv_pic=2
#2 raw=.. norm=.. at=30.9,14.3 code=rin q=2 ilvl=30 ... children=0 inv_tint=0 inv_colour=-1 inv_pic=4
#3 raw=.. norm=.. at=31.3,14.3 code=cm1 q=2 ilvl=30 ... children=0 inv_tint=0 inv_colour=-1 inv_pic=2
#4 raw=.. norm=.. at=31.1,14.1 code=sbw q=2 ilvl=30 ... children=0 inv_tint=8 inv_colour=-1 inv_pic=-1
```

`item full` on a carried item ends the same way (a worn buckler: `... children=0 inv_tint=8 inv_colour=-1 inv_pic=-1`). `caps` lists `state.caps,ui.labels,item.equip,item.unequip,item.move,item.identify,item.drop,gold.stash,item.merc-equip,item.merc-unequip,ground.full,ground.take,item.full,item.create,item.vault,vault.stats`; `vault stats` prints `vault creates=0 takes=0 vaults=0`. A worn item is refused: `item vault <id>` answers `err place worn: take it off first`. The three `inv_*` keys are unchanged from devp12 (same names, same order, same ranges).

## Switch (needs the user's go; not done)

The sandbox host runs `d2play-20261108` on `d2dbot:devp12`. Same recipe as devp12's, image only (compose files and `iv/iv2.yml` as in use):

1. `docker stop -t 60 d2play-20261108`; back up the volume: `docker run --rm --user 0 -v d2play-data:/data:ro --entrypoint tar d2dbot:devp13 czf - -C /data . > bak-devp12-before-devp13.tgz`.
2. `D2_IMAGE=d2dbot:devp13 docker compose -p play-build-20261108 -f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv2.yml up -d --no-deps d2dbot` with the exports of `up.sh`.
3. `tools/dev-update.sh --serve start`, then one `--commit origin/main --to d2play-20261108`.

What the switch brings: the upstream fidelity batch above (summon stats, curse slot, merc regen, melee reach, absorb, Cairn portal and scroll) in the sandbox hero's fights and quests, nothing new for the bot's verbs. Risk: U, a d2s written by devp13 is read back by devp12 on rollback; the save format is unchanged by the verbs (S), but upstream's own save-path changes (Convert to Expansion, `save_bytes`) are in this range, so back up first.

## Rollback

Same compose command with `D2_IMAGE=d2dbot:devp12`; `devp12` and `devp10` stay loaded. Restore the volume from the tarball only if the hero was saved by devp13 and something is wrong.
