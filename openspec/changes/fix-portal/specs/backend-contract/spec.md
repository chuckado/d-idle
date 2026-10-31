# Backend contract specification

## ADDED Requirements

### Requirement: The portal verb picks a source with charges
The town portal verb SHALL click a tome only when its quantity (stat 70, absent counts as 0) is above 0, else a scroll, and with neither SHALL return the no-portal error without a click. Combat's picker and the d2d town portal SHALL use the same pick. Source: docs/primitives-review.md F4; docs/fixportal.md (how charges are read per backend).

#### Scenario: Empty tome and two scrolls
- **WHEN** the pack holds an empty tome and two scrolls
- **THEN** the portal verb clicks a scroll

#### Scenario: Empty tome and no scroll
- **WHEN** the pack holds only an empty tome
- **THEN** no click is made and the verb returns the no-portal error

### Requirement: HasTP counts charges
The kit input `HasTP` SHALL be true only when the pack holds a scroll or a tome with charges above 0, so an empty tome makes the bot buy scrolls through the existing refill rule. Source: docs/primitives-review.md F5.

#### Scenario: Empty tome
- **WHEN** the only portal item in the pack is a tome with no charges
- **THEN** `HasTP` is false
