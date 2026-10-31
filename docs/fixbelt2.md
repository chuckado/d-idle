# fixbelt2: why potions stayed in the pack (post-fix soak 1)

> status: code on branch agent/fixbelt2; live proof pending (drill variants below).

## Log findings (Sorceress, 119 wall min, 190 fills, offline log analysis; VERIFIED from the logs)
- 47 fills moved 0 potions. Failed moves by kind: 42 mana (`no belt column for ... (kind mp)`), 7 healing, 7 rejuvenation. Other errors: 0. All marks were on an 8-box belt (sig=8).
- Plan: Sorceress clvl 1-30 is hp, hp, mp, mp (resource.PlanFor), so (a) "no mana column" is ruled out. Plan capacity on 8 boxes: 4 hp, 4 mp.
- Belt at failure: e.g. 0,4,1,5,2 healing, 3,7 mana: three healing columns, mana column full, no empty column. Belts held 5 to 6 healing against a plan of 4. So (b) partly (legitimately no room), and the cause is how it got there.
- Cause: the fill caps healing at the plan, but ground potions did not: `needsOf.Want` took any potion the game would belt (same-kind column or empty column) with no count limit, and the game puts it in the first empty column, the plan's mana column. (c) Place is correct (matches the game rule); pickBeltBox does not refuse empty columns.
- Pack: packMP > 0 in 570 of 1253 minutes, packHP > 0 in 288; belt without any mana potion in 598 minutes, without healing in 28. So the belt-empty minutes left are mostly mana. `potion already in the pack` stays high because mana potions that cannot be belted hit the pack buffer. Combat drinks from the belt (bot resource input counts belt potions; pack mana is not counted: UNVERIFIED for d2d).

## Fixes (shared, both backends)
- `no belt column` is no room, not a full belt: no sticky mark, no failure count, others continue; candidates are placed on a belt copy so no futile fill opens the pack.
- `belt.PlaceFor`/`WillBeltFor`: a potion may not open an empty column if that leaves a planned kind (hp, mp, rv with plan capacity) without a column. Fill order puts kinds without a column first. Ground pickups use it (a healing potion that would take the last mana column stays on the ground).
- The bot's own choice is still `belt.Place` (the game rule), so the bot is not stricter than the game; only the pick-up and fill decisions reserve a column.

## Decision for the user (policy number, NOT changed)
Mana columns for a Sorceress: today 2 mana of 4 columns at every belt size below clvl 31. SME (UNVERIFIED): 4 boxes 1+1; 8 boxes 1 healing + 1 mana (2 healing only if life often below 50%); 12 and 16 boxes 2+2. The reserve above protects one mana column only. plan.json/resource.PlanFor left as is.

## Drill variants to add (live proof)
- P1 with a Sorceress sash: pack of 6 healing + 4 mana, belt empty: both kinds belted, a mana column exists.
- P7 with healing potions on the ground and healing already in 3 columns: left on the ground; a mana potion is picked.
- Fill with a mana potion and no box: no `marked full` line, healing still moves.
