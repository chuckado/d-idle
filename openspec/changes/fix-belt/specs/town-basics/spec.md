## ADDED Requirements

### Requirement: Belt fill retries a vanished item id
A belt move that fails because the item id is no longer listed SHALL NOT mark the belt full; the fill SHALL re-read the pack and try the remaining potions again, and the sticky belt-full mark SHALL reset when a new game starts or the belt or the pack's potion set changes. Source: docs/primitives-review.md F1, F9.

#### Scenario: Put fails once after a re-entry
- **WHEN** the first belt put fails with item-not-found and the pack holds healing potions
- **THEN** the same fill retries and moves them, and a later fill is not blocked

### Requirement: Belt filled before leaving town
A town visit SHALL end with the pack's healing potions in the belt up to the plan, including potions bought in that visit, because d2d purchases land in the pack. Source: docs/primitives-review.md F3.

#### Scenario: Purchase lands in the pack
- **WHEN** the shop buys healing potions and they land in the pack
- **THEN** a belt step after the shop moves them before the chores end

### Requirement: Belt column holds one kind
A potion SHALL be put only into a belt column that holds its kind or is empty, never above another kind (SME, UNVERIFIED in this repo), and combat SHALL see every belt box 0..15 with column = box mod 4 on both backends. Source: docs/primitives-review.md F7, F8.

#### Scenario: Healing next to mana columns
- **WHEN** the belt row holds healing, mana, mana, mana and two more healing potions are put
- **THEN** both go into the healing column and combat counts them

### Requirement: Kit counts pack potions
The kit gate SHALL treat a healing or rejuvenation potion in the pack as present, because the belt fill places it. Source: docs/primitives-review.md F6.

#### Scenario: Empty belt, potion in the pack
- **WHEN** the belt holds no healing potion and the pack holds one, with a portal scroll
- **THEN** the kit is not missing
