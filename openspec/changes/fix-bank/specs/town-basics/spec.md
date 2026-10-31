# Town basics specification

## ADDED Requirements

### Requirement: Gold deposit works on d2d
The town stash chore SHALL deposit gold, a partial amount or the whole purse, through the controller's gold-move capability when it has one, because the d2d amount prompt starts empty and the d2d town controller cannot type. A controller without the capability SHALL keep the typed path. The deposit SHALL still be clamped to the room left under the stash gold cap and confirmed by the stash gold rising in memory.

#### Scenario: Partial deposit on a controller that cannot type
- **WHEN** the purse holds 2000 gold above a keep line of 550 and the controller only offers the gold-move capability
- **THEN** 1450 gold reaches the stash with one move and the chore reports no error

#### Scenario: Whole purse
- **WHEN** the whole purse of 1500 gold is deposited on such a controller
- **THEN** the purse is 0 and the stash holds 1500

#### Scenario: Stash nearly full
- **WHEN** the stash has room for 100 gold
- **THEN** the move is for 100

### Requirement: Repair is not attempted without gold
The repair chore SHALL NOT be needed when the character holds less gold than a lower-bound estimate of the repair price, so a character that cannot pay never opens the smith.

#### Scenario: Gold 0 with worn gear
- **WHEN** the character has 0 gold and badly damaged gear
- **THEN** no repair step runs and no error is reported, and with 500 gold the next visit repairs
