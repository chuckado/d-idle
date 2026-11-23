# Fork sync cadence

> status: current (card forksync, 2026-11-23). Keeps the private d2d fork close to upstream so a rebase stays small. Patches never go into this repo (GPL, docs/decisions.md); this page lists them by branch and file area only.

## Rules

1. Rebase the fork onto upstream main DAILY when it is more than 40 commits behind, and after any of our upstream PRs merges (the merge brings our own commit back through upstream; carry it once, not twice). The user's cap of about three days is the longest gap, not the target.
2. A read-only drift check (`tools/fork-drift.sh` in the ops repo) runs daily at 05:30 New York time from a user timer on the head box: fetch plus `git merge-tree`, no build. It writes `fork-drift.json` in the status folder and, when behind > 40 commits, a conflict is forecast, or one of our merged upstream PRs is missing from the fork base, adds one line to the needs-you list and sends one phone push. The line closes itself on the next clean run.
3. The integration branch is `didle/rebase-YYYYMMDD`: upstream main merged into the newest integration branch (which already contains every patch branch below). Never rewrite or force-push an existing branch; a new day is a new branch name. Push to the private remote the same day.
4. Conflicts: keep upstream's version wherever upstream already has the change; our lab hunks stay (`--lab-verbs`, `--vault-verbs`, `--item-verbs`, `--ladder-runewords` are gated, default off or vanilla).
5. Differential tests that fail after a merge because upstream behaviour changed are results: report them, do not hide them.
6. Image loading is automatic by an agent only in the night window and only loads (`d2dbot:devpN`); creating or restarting a container, and the SWITCH of the sandbox, always need the user's go.
7. Keep each feature in small, separate commits (verbs, reads, vault pay, lab), so a conflict stays local to one area and a patch can be dropped when upstream takes its own version.

## Mechanical rebase

```
git worktree add <dir> -b didle/rebase-<date> <newest integration branch>
git merge upstream/main          # resolve per rule 4
git push <private remote> didle/rebase-<date>
FORK_DEVP=<N> fork-rebase-builder.sh build|test|image didle/rebase-<date>   # builder, one heavy job at a time
```

Run the regression set on the builder: levelmap, levelmonsters, levelobjects, combat_reads, input_items_diff, input_verbs, input_verbs2_diff, lab_verbs, vault_verbs, item_verbs, paygold, creditgold, itemprice, itemdecode (flaky), warptriggers, smoke_d2d.

## Patch inventory (S: read from the branch history 2026-11-23)

Every branch below descends from the previous integration branch and is contained in the newest one; the "area" is where its hunks live.

| Branch (private remote) | What | Files |
|---|---|---|
| lab and item verbs (rebase-20261113/15) | `--lab-verbs` debug verbs, item equip/move/identify/drop, gold, merc equip, dump stats/merc, labels, overlay dir, frame feed, cursor | apps/d2d (devctl, itemverbs, lab verbs), components/devctl, components/game, tests/*_d2d.py |
| vault-verbs, vault-spike | `--vault-verbs`: ground take, item vault, create (identified flag), item full | apps/d2d, tests/vault_verbs_d2d.py |
| input-verbs, input-verbs2 | cmd select/repair/menu, buy, sell, waypoint, portal, pickup as client messages | apps/d2d, tests/input_verbs*_d2d.py |
| input-items | item/identify/gold verbs promoted to ordinary input | apps/d2d, tests/input_items_diff_d2d.py |
| level-layout, level-monsters, level-objects | read-only `debug levelmap`, `levelmonsters`, objpresets | components/devctl, apps/d2d, tests/level*_d2d.py |
| combat-reads | `debug minions`, `skillstate`, `combatlog` and the combat journal | components/game (fight), components/devctl, tests/combat_reads_d2d.py |
| ladder-runewords | `--ladder-runewords` switch (default off) | components/rules |
| paygold, vaultsell | `debug paygold`, `debug creditgold`, `item price` (`--vault-pay`) | apps/d2d, tests/paygold_d2d.py, creditgold, itemprice |
| warp-triggers | read-only `debug warptriggers` | apps/d2d, tests/warptriggers_d2d.py |
| recallsorc | item create takes the next free unit id | components/game, tests/vault_verbs_d2d.py |
| fork-only | window icon and Steam art, README, SPDX headers | assets/icon, README-DIDLE.md |

Conflict hotspots seen so far: `components/game/fight/tick.cpp` (combat journal and `debug_blow` next to upstream's comment rewrites) and `components/rules/rules.hpp` (the runeword row helpers).
