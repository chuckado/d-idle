# Fork request card: verbs the step proving grounds are missing (card stepgrounds, 2026-11-23)

> status: draft for the coordinator. Private fork only (never an upstream PR). Not touched by this card: another agent owns the fork. Tags: V verified by the named probe on the devp23 labverbs binary (`/data/d2d-lab/qdbin/d2d-devp23-labverbs`, builder), U unverified. The catalogue column "MISSING F<n>" in `docs/probes/quest-step-catalogue.md` counts the steps each request unblocks.

## Present and used (V)

`debug goto <area>`, `debug warp <x> <y>`, `debug spawn <id> [n] [champion|unique]` (probe `tools/probes/spawnids.py`: accepted bloodraven, smith, andariel, radament, summoner, duriel, izual, hephasto, diablo, mephisto, baal, nihlathak, councilmember1, griswold; refused "no such monster": countess, fangskin, shenk, bishibosh, coldcrow, rakanishu, treehead, pitspawn, bonebreaker, corpsefire, blooddiablo), `debug kill` (kills every monster of the level and leaves the normal drops; `tools/probes/stepdrop.py`), `debug give <code> <q>` (an item on the ground at the hero's feet, listed by `ground`; `tools/probes/giveground.py`), `debug quest set`, `debug mon set <id> life|...`, `debug merc hire|level|heal`.

## F1: force the quest drop (or spawn a level-placed superunique with its own treasure class)

- Evidence (V, `tools/probes/stepdrop.py D2D DATA 11 28 smith` with and without `QSET=3:0x000c`, plain and `unique`): a Smith made by `debug spawn smith` and killed by `debug kill` leaves 49 ground rows and NO `hdm` (Horadric Malus); the same for the `unique` spawn (it becomes a random unique "Chaos Dancer the Cold" with 404 hp, not The Smith). The level-placed Barracks Smith is not reachable by the probe's warp lattice (U), so whether the real one drops the Malus in the lab d2d is open; the matrix run 20261123-171725 killed the real one and the Malus did not register (note 38), so the lab d2d may lack the quest drop altogether (S).
- Ask (either): `debug spawn super <name|index>` that places the superunique of superuniques.txt with its mods, treasure class and quest flag exactly as level generation does (also unblocks `countess`, `fangskin`, `shenk`, `bishibosh` ...); or `debug drop quest <code> [x y]` that runs the quest-item drop rule for the quest of that code, independent of who killed (user fact: the quest drop does not depend on the killer).
- Unblocks: Smith drop (the pending scenario `step.act1.smith.drop_pickup`), Countess, Fangskin, Radament's book, Khalim's parts, Hephasto's hammer, Shenk, Nihlathak.

## F2: place or drive a level object

- Evidence: no verb places a preset object (stones, tree, cage, chest, seal, anvil, barricade door, Anya); `debug objectlist` and `debug objects` only read them (V devp23 doc).
- Ask: `debug object add <class> [x y]` (a preset object of objects.txt class at the hero or the point, in its default state) and `debug object state <id> <mode>`.
- Unblocks: touch_stones, the Inifuss tree, Staff chest, seals, the Hellforge anvil, barricade doors, Anya.

## F3: observable quest words of records 16 to 47

- Evidence (matrix note 30, V runs 20261123-070440 and -071114): scenario assertions on records 17 to 40 read "unobserved". devp23 says `debug quest words` prints 48 live words per difficulty (S source, V the labverbs test); the scenario observer (`internal/scenario` state reader) does not call it yet.
- Ask: none for the fork if `debug quest words` already returns them; otherwise rows 16 to 47 in `debug quest list`. The bot-side half (the observer reading `debug quest words`) is a repo task, not a fork task.
- Unblocks: all Act 3 to 5 steps that end in a quest word.

## Not a fork matter (bot or format)

- `pre.kill` / `pre.warp` hooks in the scenario format (a scenario cannot say "kill the level after N seconds" or "warp to x,y"; the verbs exist).
- A talk-step assertion (an NPC dialog or its reward is not observable except through a quest word).
- An assertion kind for unspent skill points and stat points.
