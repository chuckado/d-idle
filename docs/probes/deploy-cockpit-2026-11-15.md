# Deploy record: cockpit UI to the sandbox, 2026-11-15

> status: record. Task TASK-490. Tags: V = verified in this run (command named), U = not verified.

## What was deployed

- Bot and watch build from origin/main 7dde87e96 (cockpit default, S7, all slices, bot cards B1-B9 Go halves) with `tools/dev-update.sh --commit origin/main --to d2play-20261108` (the sandbox host is selected with `D2_SANDBOX_HOST`; the builder host is not the sandbox host since the build-hosts change). V: `UPDATED d2play-20261108 0.0.261114180251 -> 0.0.261115105052 in 17s`.
- Image unchanged (devp7): the new private fork merc item verbs are optional, the Go client degrades without the capability. No image build or swap was done.
- The sandbox host was asleep and was woken with `labrun wake orion` (the container restarted with the host).

## Rollback state

Before: bot 0.0.261114180251, image devp7, container healthy. Rollback: `tools/dev-update.sh --rollback --to d2play-20261108` (returns to the previous release; V in docs/dev-update.md). Not needed. The d2play-data volume and the game mount were not touched.

## Sanity (V, from container logs and `docker inspect`)

- Container healthy; the `d2bot start` line shows 0.0.261115105052, equal to the deployed version.
- No ERROR/FATAL/panic lines after the update (checked at 1, 2 and 5 minutes). The only two ERROR lines in the window are the previous release's watchdog exit at the host wake (runner loop idle while the host slept), before the update.
- Hero alive: clvl 15, deaths 0, xp rising 201596 -> 201811 over about 4 minutes, `minute` lines continuous.
- `zone entry` now carries `def`; dmg/block/dmgpct/ar did not appear yet (U: absent-when-unread or not wired at this point). No `zone shadow:` or `zone choice:` line appeared in 8 minutes (U: they need a planner line that did not occur).

## Not done

- Watch UI HTTP checks (/watch layouts, X-Layout-Source, /api/*) and the screenshots at 1914x930, 1280x620 and 390x844: the watch screen needs its access token, even on loopback, and using that token was refused by the permission system. U. Ready-to-run list for a session with the token: GET /watch (cockpit/1), ?layout=classic|standard|simple|overlay (200 plus X-Layout-Source), /api/glance, commands, pause, survival, merc, hero, inventory, danger; then headless chromium screenshots on the builder at the three sizes. GET only.
