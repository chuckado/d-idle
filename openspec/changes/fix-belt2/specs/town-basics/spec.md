## ADDED Requirements

### Requirement: A kind without room is not a full belt
A belt fill SHALL treat "no column can take this potion kind" as no room for that kind: the potion stays in the pack, the other kinds are still moved, and the belt MUST NOT be marked full. A potion MUST NOT open an empty belt column when that leaves a planned kind with no column, in the belt fill and for ground potions. Source: docs/fixbelt2.md.

#### Scenario: Mana potion, no free box in a mana column
- **WHEN** a sash holds healing in three columns and a full mana column, and the pack holds mana potions
- **THEN** no fill candidate exists, nothing is clicked and the belt is not marked full

#### Scenario: Healing would take the last mana column
- **WHEN** healing fills three columns of a Sorceress belt and a healing potion lies on the ground
- **THEN** it is not wanted, and a mana potion is
