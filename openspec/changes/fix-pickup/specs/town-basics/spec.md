## ADDED Requirements

### Requirement: Ground potion goes where the game puts it
A clicked ground potion (inventory panel closed) SHALL be treated as landing on the belt when a same-kind column has a free box or a column is empty, else in the pack, and the pickit, the pickup arrival check and the belt fill MUST all follow that rule. Source: docs/fixpickup.md.

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
