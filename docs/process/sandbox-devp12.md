# Sandbox d2d build devp12 (2026-11-18)

> status: current. Task: TASK-645. Tags: V = verified in this run (how said), S = read from a tool or log, U = unverified. Image built and LOADED on the sandbox host only; the running container was not touched. Follows devp10 (`docs/process/sandbox-devp10.md`, on its own branch when this was written); devp11 (id b9c25586c407) had `inv_tint` / `inv_colour` only.

## What devp12 adds

`inv_pic`, the item's inventory picture variant (d2s `Item::picture`, the ItemTypes `VarInvGfx` index), next to `inv_tint` and `inv_colour` in `dump items`, `item full` and `ground full`. It only appears behind `--item-verbs` or `--vault-verbs` (fork only; nothing for upstream).

## Fork and image

- Private fork branch `didle/rebase-20261118`, build tip `caea2368` (on `0626bb6e`, the inv_tint / inv_colour slice). V (`git ls-remote` on the private remote showed `caea2368`). A test-only commit `4a083dfe` (`tests/item_pic_d2d.py`, wired in `tests/CMakeLists.txt`) was pushed on top afterwards; the binary is built from `caea2368`, the commit changes no source. V (`git push` output).
- Built on the NAS builder at -j2: `tools/ops.sh run build-fork.sh devp12 --jobs 2 --timeout 3000` through `tools/agentrun.sh`. The branch was fetched into the worker host's fork checkout as the slash-free ref `devp12` (a ref with a slash makes the script's copy to `fork/bin` fail). d2d about 8.5 minutes, rc 0, 17,862,648 bytes, sha256 prefix `a05f1c3ea0032a37`. V.
- Image `d2dbot:devp12`, id `a064b241ba61`: `tools/dev-update.sh --image d2dbot:devp12 --d2d <that binary>` with the host values loaded (23 s, the same step as devp10). V. The binary inside the image has the same sha256 prefix as the builder's, and `d2d --help` in a throwaway `docker run --rm --entrypoint` lists `--item-verbs`, `--vault-verbs`, `--lab-verbs`. V.
- Loaded on the sandbox host: `docker images` there lists `d2dbot:devp12 a064b241ba61` next to `devp10 51e4773cf727`. The container `d2play-20261108` still runs `devp10` (healthy) and no compose file was changed. V.

## Tests (on the builder, not the sandbox host)

`D2_MPQ_DIR=/data/gamedata nice python3 tests/<t>.py <builder d2d>`; all three printed ok, rc 0. V.
- `item_pic_d2d.py` (new): `dump items` carries `inv_pic` with `--item-verbs`; 36 rolled jewels give pics `[0,1,2,3,4,5]`, 12 rings `[0,1,2,3,4]`, 12 charms `[0,1,2]`; a short sword (non-variant) has `-1`; the carried copy and `dump items` agree with the floor copy.
- `item_colour_d2d.py` (existing, now also checks `inv_pic` in `dump items`): ok.
- `vault_verbs_d2d.py` (existing): ok, 0 failures.

## JSON and line shapes

Three integers, all present together or all absent (flags off). Index meaning: `inv_tint` is the tint index of the base picture (S: `Scene::inv_tint_colour`), `inv_colour` is the colour class of the affix (-1 = none, magic jewel seen 5, rare 3, unique 13 in the colour test), `inv_pic` is the picture variant: jewels 0..5, rings 0..4, amulets and charms 0..2, `-1` when the item type has no variants (base picture). Value ranges V (the test above).

`dump items` (one JSON object per item inside `items`, last three keys):

```
{"id": 9, "code": "jew", "loc": "inventory", "slot": 0, "x": 9, "y": 3, "w": 1, "h": 1, "quality": 4, "ilvl": 30, "sockets": 0, "qty": -1, "dur": [0, 0], "unique": -1, "set": -1, "flags": {...}, "name": "Jewel", "inv_tint": 0, "inv_colour": 5, "inv_pic": 2}
```

`ground full <id>` and `item full <id>` (text line; `raw` and `norm` hex shortened here), the tail is ` inv_tint=T inv_colour=C inv_pic=P`:

```
#1 raw=.. norm=.. at=31.1,14.3 code=jew q=4 ... children=0 inv_tint=0 inv_colour=5 inv_pic=2
#2 raw=.. norm=.. at=30.9,14.3 code=rin q=6 ... children=0 inv_tint=0 inv_colour=-1 inv_pic=4
#3 raw=.. norm=.. at=31.3,14.3 code=cm1 q=4 ... children=0 inv_tint=0 inv_colour=-1 inv_pic=2
#4 raw=.. norm=.. at=31.1,14.1 code=sbw q=2 ... children=0 inv_tint=8 inv_colour=-1 inv_pic=-1
```

(jewel, rare ring, charm, short sword; lines from a `--lab-verbs --vault-verbs` session on a fresh test hero, rolled with `debug give`; they are mechanics examples, not game-behaviour evidence.) The bot should read the three as optional keys: a missing `inv_pic` means an older d2d or flags off.

## Switch (needs the user's go)

The compose set already in use for devp10 on the sandbox host is `-f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv2.yml` in `~/play-build-20261108` (S: `iv/iv2.yml` sets `D2BOT_ARGS: '--d2d-args "--item-verbs --vault-verbs"'`). Switch is the image only:

1. `docker stop -t 60 d2play-20261108`; back up the volume: `docker run --rm --user 0 -v d2play-data:/data:ro --entrypoint tar d2dbot:devp12 czf - -C /data . > bak-devp10-before-devp12.tgz`.
2. `D2_IMAGE=d2dbot:devp12 docker compose -p play-build-20261108 -f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv2.yml up -d --no-deps d2dbot`, with the exports of `up.sh` (`D2_GAME_DIR`, `D2_WATCH_ORIGIN`, `D2_PUBLISH`, `D2_DEV_URL`, `D2_DEV_PUBKEY`).
3. `tools/dev-update.sh --serve start`, then one `--commit origin/main --to d2play-20261108`.

Not run in this job. U until done: that the bot code in main reads `inv_pic`.

## Rollback

Same compose command with `D2_IMAGE=d2dbot:devp10` (same `iv/iv2.yml`); `devp10` stays loaded. Restore the volume from the backup tarball only if the hero was saved by devp12 and something is wrong (the d2s format is unchanged, `inv_pic` is output only; S).
