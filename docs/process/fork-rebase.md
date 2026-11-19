# Fork rebase routine

> status: current (task forkrebase, user decision 2026-11-19: rebase the private d2d fork onto upstream main regularly, every few days and after each batch of our PRs merging upstream). Tags: V = verified in the first round (2026-11-19, say how), S = read from a tool or log, U = unverified. Private details (hosts, paths, builder wrapper names) are in the private ops repo, `docs/private/fork-rebase.md`.

## Why

The private d2d fork carries patches upstream does not have (item/vault/lab verbs behind `--item-verbs`, `--vault-verbs`, `--lab-verbs`, the `inv_tint` / `inv_colour` / `inv_pic` fields, the debug presets, objects and portals, the lockstep clock, the framefeed and cursor verbs, the overlay). The bot's d2d client (`internal/d2dctl`, `internal/d2dsrc`) parses those verbs and fields. Every day the fork stays on an old base, the bot plays a d2d that upstream has already changed (quests, loot, combat, paths). A short routine, run often, keeps each rebase small (a few hundred upstream commits, two or three conflicts) instead of a big one.

## Rules

1. One new branch per round: `didle/rebase-YYYYMMDD`, created from the previous round's branch. Never rewrite or force-push a previous branch; the sandbox image is built from a named branch tip and the doc cites it.
2. Strategy that kept the history clean (S: branches 20261115, 20261117, 20261118 and 20261119 are all linear stacks of our own commits on one upstream commit, no merge commits): `git rebase --onto upstream/main <old merge-base>` on a copy of the previous branch. Our commits stay one patch each, in the original order. Patches upstream took (same patch-id) drop out by themselves.
3. Conflict policy: keep both sides' intent. Upstream behaviour wins where a function was rewritten; our verb or field is then re-attached to the new code. Typical shapes seen so far: a usage string both sides extended (merge the two lists); a verb we moved into a larger one while upstream edited the old copy (delete the old copy, port the upstream change into ours). A conflict that needs a decision about meaning (not text) stops the round: write it in the round's doc and ask.
4. Push the branch to the private remote the same day; a doc may only cite a commit that is there. Nothing goes to GitHub from this routine.
5. The build runs on the builder, one job at a time, only when `labrun builder-status` shows the main lane free (gates have priority). More than 45 minutes: the long lane. Never on the worker host for the build; the image is assembled and loaded there.
6. The running sandbox container is never switched by the routine. A switch is a proposal in `docs/process/sandbox-devpNN.md`; the coordinator asks the user first.

## Steps

`tools/fork-rebase.sh` does 1 to 4 and prints the commands for the rest; it stops with a clear message on a conflict (resolve, `git add`, `git rebase --continue`, then run the same sub-command again).

1. Measure (`fork-rebase.sh measure`): fetch upstream; commits behind and ahead, files touched on both sides (predicted conflicts), the upstream commits that touch the devctl layer (`apps/d2d/devctl`, `docs/control_channel.md`, `apps/d2d/main.cpp`).
2. Rebase (`fork-rebase.sh rebase`): new worktree and branch from the previous round, `rebase --onto`, stop on conflict. After a clean run: compile check (step 4) before push.
3. Push the new branch to the private remote.
4. Build d2d on the builder from a slash-free ref (the wrapper's copy to the worker fails when the branch name has a slash, S; the script makes a ref `devpNN`).
5. Tests on the builder (below). All must pass before the image is built.
6. Image `d2dbot:devpNN`: the dev-update step with the new binary; load on the sandbox host only; check `d2d --help` lists the three gates and the binary hash equals the builder's.
7. Doc `docs/process/sandbox-devpNN.md`: payload examples from the new binary, the diff against the previous image (what upstream changes it carries), the switch (needs the user's go) and rollback.

## Tests (all on the builder, with the game data there)

- Fork verb tests (python, one d2d process each): `lab_verbs_d2d`, `item_verbs_d2d`, `vault_verbs_d2d`, `item_colour_d2d`, `item_pic_d2d`, `labels_d2d`, `merc_verbs_d2d`, `cursor_verb_d2d`, `overlay_d2d`, `smoke_d2d`.
- Upstream ctest: full test build and `ctest` in the long lane when the round carries a lot of upstream change (more than about 100 commits); the d2d-driven tests above are part of it. Tests without game data report Skipped, that is not a failure.
- Bot side (regression checklist below): the bot's own parser tests against recorded replies, then the read-only real-d2d tests against a d2d started on the builder.
- Lab smoke: a headless d2d with a throwaway character on the builder, a scripted devctl session (walk, fight a spawned monster, ground items, `dump items` with the three item gates). The full bot drill (`cmd/d2drill`) runs on a worker slot after a sandbox switch, not in the round.

## Regression checklist: what the bot's client assumes

The client code is `internal/d2dctl` (`verbs.go`, `item_verbs.go`, `combat_verbs.go`, `town_verbs.go`, `debug_skills.go`, `parse.go`) and `internal/d2dsrc` (`items.go`, `player.go`, `quests.go`, `missiles.go`, `gearstats.go`, `cursor.go`). A round must keep these true on the new binary:

| Assumption | Check |
|---|---|
| `ping`, `info`, `state` lines parse with no `Bad` lines | `TestRealD2D` in `internal/d2dctl` and `internal/d2dsrc` (needs `D2D_SOCK`), 0 unparsed lines |
| `cmd move`, `skill`, `interact`, `pickup`, `resurrect`, `belt`, `hand`, `grid`, `use`, `trade`, `buy`, `sell`, `close`, `said`, `chat`, `imbue`, `caravan`, `waypoint` exist (the `err cmd ...` usage string lists them) | send an unknown `cmd` and read the usage line; the fork's own `item_verbs_d2d` drives the same cmds |
| `item equip/unequip/move/identify/drop`, `gold deposit/withdraw`, `item merc-equip/merc-unequip` with `--item-verbs` | `item_verbs_d2d`, `merc_verbs_d2d` |
| `ground full`, `item full`, `item vault`, `ground take`, `item purge` with `--vault-verbs` | `vault_verbs_d2d` |
| `dump items` JSON keys incl. optional `inv_tint`, `inv_colour`, `inv_pic` (all three or none) | `item_colour_d2d`, `item_pic_d2d` |
| `dump merc`, `dump stats`, `dump vendor`, `dump ui`, `dump player` | `lab_verbs_d2d`, the bot parser tests against `internal/d2dctl/testdata` |
| `debug` verbs the drills use: `level`, `warps`, `enter`, `presets`, `objects`, `portals`, `missiles`, `states`, `mana full`, `mon set`, `give`, `kill [n] [spare id] [within cells]`, `quest`, `layout`, `grid`, `blocked`, `corpses`, `players` | `lab_verbs_d2d` plus the session in the smoke step |
| `lockstep`, `frame`, `framefeed`, `cursor on/off`, `labels` | `cursor_verb_d2d`, `labels_d2d`, the smoke session |
| Caps and refusals are named and stable (`err ... refused`, ledger lines) | `vault_verbs_d2d`, `item_verbs_d2d` |

A check that fails is a fork or bot fix before the image, not after. A reply change on the upstream side (a usage string, a field order) is read from the round's `measure` output (devctl commits) and from the parser tests; list it in the sandbox doc.

## Cadence

Every few days, and after each batch of our PRs merges upstream (those commits then drop out of the fork by patch-id; the fork stack shrinks). The coordinator starts a round; one agent does measure to image; the switch proposal waits for the user.
