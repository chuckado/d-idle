# Backend contract specification

## ADDED Requirements

### Requirement: The d2d backend enters a town portal with the client's portal command
On d2d, interacting with a town portal object SHALL send the command the client sends for a click on that portal (Interact with id -2000 minus the portal's index), never a screen click computed from a transform, and the portal object's unit id SHALL carry the World's portal index. Source: docs/d2d-portal-entry.md (d2d resolves a world click through the hover pick of the last drawn frame; a lockstep d2d draws none, so the click walked instead, drills D2.t and D2.g).

#### Scenario: Portal interaction on a lockstep d2d
- **WHEN** the town layer interacts with the town portal object in the field
- **THEN** the control socket receives `cmd interact -2000` for portal 0 and no click

#### Scenario: Portal in town
- **WHEN** the portal in the current level is index 1 (the town twin)
- **THEN** the command names -2001
