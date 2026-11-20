# Sandbox reset to a fresh hero, 2026-11-15

> status: record. Task sandboxfresh. Tags: V = verified in this run, U = not verified.

User request: a fresh hero in the play sandbox (container d2play-20261108, image devp8, bot 0.0.261115193109, nothing swapped).

## Archive (on the worker host, under the sandbox archive directory in the user's home, mode 700)

- `<stamp 20261115-210202>/`: `state.tgz` (345734 bytes, sha256 8f121fac17...31ef0, in `state.tgz.sha256`), `files.txt` (72 entries), `meta.txt`, written by `tools/dev-reset.sh` and read back before anything was removed (V). It holds instances/ (hero save, goals, bot state, tuning, ledger), config/ and wizard.env of the clvl 17 Paladin. No key file was read or printed.
- Extra copies in the same directory: `docker-logs-12h.gz` (3651808 bytes, `gzip -t` ok) and `SHA256`; the same files also sit in a sibling `hero17-20261115-logs` directory.

## Reset (V)

`D2_LABRUN_HOST=<worker host> tools/dev-reset.sh --to d2play-20261108 --yes` from a checkout at current main. `tools/dev-sandbox.sh` ignores `D2_SANDBOX_HOST` (only the dev-update scripts honor it), so the labrun host is given through `D2_LABRUN_HOST`; without it the command went to the builder and failed with "no container". The script waited 240 s for the new save and printed "HERO not seen", yet the log shows the new hero playing: the wait is too short for the first character creation (U why).

## After (V, container logs)

Same name (Doofer) and class (Paladin, class index 2; Zeal in the combat log), level 1 start, clvl 2-3 within about 6 minutes, 61 kills, 0 deaths, xp rising, container healthy, bot version unchanged, 0 error/panic/fatal and 0 watchdog lines. No `zone entry` line appeared in the first 15 minutes (the hero stays in its first area), so the stats fields are unchecked (U). Zero setup writes: a normal fresh character.

## Rollback

`D2_LABRUN_HOST=<worker host> tools/dev-sandbox.sh restore --to d2play-20261108 --from 20261115-210202 --yes` (archives the fresh state first as `<stamp>-pre-restore`, so it is undoable). Then hard-refresh the watch page.
