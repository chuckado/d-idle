## ADDED Requirements

### Requirement: Ground potion goes where the game puts it
With the inventory panel closed, a clicked ground potion SHALL be treated as landing on the belt when a belt column whose bottom potion is of the same kind has a free box, or a column is empty; otherwise in the pack. The pickit MUST want a potion that will land on the belt even when the pack already holds one, MUST apply the pack buffer only to potions that would land in the pack, and MUST count a potion that arrives on the belt or in the pack (by the held count of its class, not the unit id) as picked up. The belt fill SHALL choose its box by the same rule. Source: docs/research/belt-pickup-trace.md, docs/fixpickup.md.

#### Scenario: Matching column has room, one in the pack
- **WHEN** a healing potion lies on the ground, the belt has a healing column with a free box and the pack holds one healing potion
- **THEN** the potion is wanted

#### Scenario: Belt full for the kind
- **WHEN** no column can take the kind and the pack already holds the buffer
- **THEN** the potion is skipped by the buffer rule

#### Scenario: Other kind full, empty column free
- **WHEN** the healing column is full and another column is empty
- **THEN** the potion is wanted (the empty column takes it)

#### Scenario: Potion lands on the belt
- **WHEN** the ground unit disappears and the belt holds one more potion of its class
- **THEN** the pickup counts as arrived
