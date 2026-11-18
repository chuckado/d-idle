# Sandbox image devp11: item colour payload (slice 1 fixtures)

> status: current (task-638, 2026-11-18). Built and checked on the builder, NOT loaded into the sandbox. Design: docs/design/item-colour-2026-11-18.md (slices). Tags: V = verified on the builder run described here, U = unverified.

## What it is
- Fork branch `didle/rebase-20261118` on the private remote, tip `0626bb6e` (one commit on the sandbox lineage tip `9df9b936`). Fork-only, never upstream without the user's go.
- Image `d2dbot:devp11`, id `b9c25586c407` (V: `docker build --cpu-shares 2 --build-arg DEV=1 --build-arg BUILD_JOBS=2` on the NAS builder from `deploy/d2d-container/context.sh` with the fork commit; exit 0). Not loaded or switched.
- Build note: `build-fork.sh` stages from the worker host's fork checkout and copies to the worker with a slash-free name only; this build used a local fork worktree, `git archive` into the builder container and a manual cmake (`-j2`). The binary used for the checks is the builder build of the same commit, not the image binary (U: the image binary was not started).

## Two new integers
Present only when d2d runs with `--item-verbs` (dump) or `--vault-verbs` (dump, `item full`, `ground full`); otherwise nothing changes.
- `inv_tint` (int): the base's `InvTrans` when it has a colormap set (1, 2, 5..8), else 0.
- `inv_colour` (int): Colors.txt index 0..20 from the `Colours::of(..., inv=true)` chain (magic/rare: first suffix transformcolor, else first prefix, else automod; unique/set: invtransform; else socketed gem or automod), -1 when none. It is reported even when `inv_tint` is 0.

## Shapes
`dump items` (JSON, one line; same fields on `dump merc` `items` rows): the two keys are appended after `name`.
```json
{"id":9,"code":"jew","loc":"inventory","slot":0,"x":9,"y":3,"w":1,"h":1,"quality":4,"ilvl":30,"sockets":0,"qty":-1,"dur":[0,0],"unique":-1,"set":-1,"flags":{"identified":false,"ethereal":false,"socketed":false,"runeword":false},"name":"Jewel","inv_tint":0,"inv_colour":5}
{"id":1,"code":"jav","loc":"body","slot":4,"x":0,"y":0,"w":1,"h":3,"quality":2,"ilvl":1,"sockets":0,"qty":60,"dur":[2,2],"unique":-1,"set":-1,"flags":{"identified":true,"ethereal":false,"socketed":false,"runeword":false},"name":"Javelin","inv_tint":8,"inv_colour":-1}
```
`item full [id]` and `ground full [id]` (text, key=value): ` inv_tint=T inv_colour=C` is appended after `children=N` (raw and norm hex are unchanged; `ground take` and `item vault` lines do not carry it):
```
#1 raw=<hex> norm=<hex> at=31.1,14.3 code=jew q=4 ilvl=30 uid=0 id'd=0 eth=0 sock=0 def=-1 dur=0/0 qty=-1 pre=0 suf=181 rare=0,0 aff=0.0.0.0.0.0. set=-1 uniq=-1 rw=-1 nprops=3 props=54:0:1;55:0:5;56:0:50; children=0 inv_tint=0 inv_colour=5
```
A parser must treat both fields as optional (missing = no tint, empty string in `status.Item.Tint`).

## Measured values (V: builder run, hero Sorceress, seed 5, `debug give`; fork test `tests/item_colour_d2d.py`, rc 0; `vault_verbs_d2d.py` still passes)
| item | quality | inv_tint | inv_colour |
|---|---|---|---|
| jewel, magic (suffix 181) | 4 | 0 | 5 |
| jewel, rare | 6 | 0 | 3 |
| jewel, unique | 7 | 0 | 13 |
| short sword `sbw`, normal | 2 | 8 | -1 |
| javelin `jav`, normal (worn) | 2 | 8 | -1 |
| ring, rare | 6 | 0 | -1 |
The carried copy (`item full`, `dump items`) equals the floor copy (`ground full`) for every row.

## Open points the bot slices must know
- The fork's table says the `jew` base has `InvTrans` 0 (V: printed above), so d2d's own `Scene::item_map` draws jewels WITHOUT a colormap even though the chain yields a colour. Whether game.exe tints jewels through another path (the design doc assumed the tint came with the colour) is U. Slices 3 and 4 must not require `inv_tint > 0` to show a colour for jewels until that is settled; fixtures should carry both `inv_tint=0, inv_colour>=0` and `inv_tint>0`.
- Ring colour -1 on the one rare ring rolled; which affix rows carry a colour is U (one sample).
- Not run: a drill against a live hero in the sandbox, comparison with the game window.
