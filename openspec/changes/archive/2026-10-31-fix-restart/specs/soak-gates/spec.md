# Soak gates specification

## ADDED Requirements

### Requirement: A game-lost recovery counts as a restart
When the bot restarts the game after the game process was lost (a crash, a kill, a hang), the minute line's `restarts` SHALL rise by exactly one at the successful restart, on every backend. A restart already counted through the menu observer or as a bot-started new game SHALL NOT be counted a second time.

#### Scenario: d2d process killed
- **WHEN** the slot's d2d is killed under the managed bot and the bot restarts it and re-enters a game
- **THEN** `restarts` on the next minute lines is one higher than before the kill

#### Scenario: Game.exe recovery
- **WHEN** the same recovery runs on Game.exe and the menu observer sees the new game
- **THEN** `restarts` rises by one, not two

### Requirement: xpRisk is XP since the last save on d2d
The minute line's `xpRisk` and the xp-at-risk warnings SHALL be the experience in memory minus the experience stored in the character's save file, on d2d as on Game.exe; when the save file cannot be read the fallback is the experience at the last game entry or restart. The save file SHALL be re-read when it changes, so a Save and Exit brings `xpRisk` to 0.

#### Scenario: Unsaved gain
- **WHEN** the save file holds 1000 XP and the character has 1500
- **THEN** `xpRisk` is 500, and 0 after a save rewrote the file to 1500

#### Scenario: Kill and re-entry
- **WHEN** the d2d process is killed with 300 unsaved XP and the character reloads from the save
- **THEN** `xpRisk` is 0 again, and the lost 300 is not counted as at risk
