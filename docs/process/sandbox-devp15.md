# Sandbox image devp15 (loaded, not switched)

> status: current (2026-11-20, card devp15). The sandbox container still runs the image it ran before; switching it to devp15 needs the user's own go.

## What devp15 is
- One private fork branch, `didle/devp15-20261120`, made by merging three branches that all start from the devp14 branch (`didle/input-verbs-20261119`): level layout (`debug levelmap`, a read verb), input items (item and gold verbs usable outside the lab) and ladder runewords (`--ladder-runewords`). V: the merge had no conflicts (`git merge` of the three tips, 12 files, tests and d2d sources only).
- Defaults are unchanged: `--ladder-runewords` is OFF (vanilla offline rules), so gates and fidelity runs behave as on devp14. V: the flag is a CLI11 bool flag with `false` as the initial value in `apps/d2d/main.cpp`.
- The image tag is `d2dbot:devp15`, built with the routine in `docs/process/fork-rebase.md` (image only: `tools/dev-update.sh --image d2dbot:devp15 --d2d <binary>`; load, never `up`).

## Tests run before loading
Built on the builder host (long lane); the python verb tests and the runeword C++ test ran against the freshly built binary with the game data on the builder: levelmap_d2d, input_items_diff_d2d, input_verbs_d2d, lab_verbs_d2d, vault_verbs_d2d, item_verbs_d2d, item_colour_d2d, item_pic_d2d, labels_d2d, merc_verbs_d2d, cursor_verb_d2d, overlay_d2d, smoke_d2d and test_runeword. Results: see the task result (docs/backlog task devp15). The routine does not run the fork's whole ctest.

## Recreate (when the user decides to switch)
1. On the worker host, in the sandbox compose directory, change the image tag of the d2d service from the current one to `d2dbot:devp15` (one line in the compose override that sets the d2d args; keep the volumes, ports and overlays as they are).
2. Recreate only that service (`up -d` for the one service); save data lives in the mounted data directory, not in the image.
3. Check: container healthy, `docker ps` shows the new image, the d2d control socket answers, one `debug levelmap` call works.
4. To use ladder runewords add `--ladder-runewords` to the d2d args; leave it off for gates.

## Rollback
Set the image tag back to the previous one (`d2dbot:devp13` on 2026-11-20; devp14 is also loaded) and recreate the same service. The old images stay loaded, nothing was deleted. Data is untouched by an image change.

## Not done
- No container was created, restarted or stopped, and no port, overlay or data directory was touched when loading (V: `docker ps` before and after, see the task result).
- The switch itself: user go required.
