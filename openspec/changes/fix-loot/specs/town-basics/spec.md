# Town basics specification

## Purpose
Delta of the fix-loot change: the recovery pickup rules must be reachable from the bot loop. Source: docs/primitives-review.md F2.

## ADDED Requirements

### Requirement: Recovery pickup is wired
While the recovery phase is on, the Looter handed to the bot loop SHALL be in recovery mode and SHALL report recovery totals, so that every gold pile and every sellable item up to the kit line is picked up and the `recovery: loot` line counts them. Wrapping the Looter in an adapter MUST NOT hide the recovery methods.

#### Scenario: Recovery pass through the adapter
- **WHEN** the recovery phase switches recovery mode on through the bot-facing Looter and a small gold pile and a sellable sword lie on the ground at a gold amount between the poor line and the kit line
- **THEN** both are picked up and the recovery totals are non-zero

#### Scenario: Not recovering
- **WHEN** recovery mode is off at the same gold amount
- **THEN** neither is picked up
