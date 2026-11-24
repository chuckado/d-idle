# Fork request card: verbs the step proving grounds are missing (card stepgrounds, 2026-11-23)

> status: F1 and F2 DONE in devp25 stepverbs (2026-11-23, V proofs in the devp25 doc and docs/probes/stepgrounds.md round 3); F3 observer half done in the bot. Original draft follows. Private fork only (never an upstream PR). Not touched by this card: another agent owns the fork. Tags: V verified by the named probe on the devp23 labverbs binary (`/data/d2d-lab/qdbin/d2d-devp23-labverbs`, builder), U unverified. The catalogue column "MISSING F<n>" in `docs/probes/quest-step-catalogue.md` counts the steps each request unblocks.

## Present and used (V)

`debug goto <area>`, `debug warp <x> <y>`, `debug spawn <id> [n] [champion|unique]` (probe `tools/probes/spawnids.py`: accepted bloodraven, smith, andariel, radament, summoner, duriel, izual, hephasto, diablo, mephisto, baal, nihlathak, councilmember1, griswold; refused "no such monster": countess, fangskin, shenk, bishibosh, coldcrow, rakanishu, treehead, pitspawn, bonebreaker, corpsefire, blooddiablo), `debug kill` (kills every monster of the level and leaves the normal drops; `tools/probes/stepdrop.py`), `debug give <code> <q>` (an item on the ground at the hero's feet, listed by `ground`; `tools/probes/giveground.py`), `debug quest set`, `debug mon set <id> life|...`, `debug merc hire|level|heal`.

## F1: spawn a level-placed superunique with its own treasure class (superuniques by name)

- Evidence (V, `tools/probes/spawnids.py`): `debug spawn countess|fangskin|shenk|bishibosh|coldcrow|rakanishu|treehead|pitspawn|bonebreaker|corpsefire|blooddiablo` answers "no such monster"; `debug spawn <x> unique` makes a RANDOM unique, not the named one.
- CORRECTION (coordinator, devp24 doc): the Smith's Malus is NOT a fork gap. The d2d drops it only when a player operates the Malus stand (object 108), never from the corpse; the Smith step scenario is `step.act1.smith.stand_click_pickup` (waits for the bot's stand click, card toolstrade a7b2ec8e). My probe `stepdrop.py` agrees (no hdm after a spawned Smith dies).
- Ask: `debug spawn super <name|index>` placing the superunique of superuniques.txt with its mods, treasure class and quest flag. Unblocks Countess, Fangskin, Shenk.

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
