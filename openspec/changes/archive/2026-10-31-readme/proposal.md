## Why

The user decided (2026-10-31) that the official styling of the project name is "d.idle" with a lowercase d, and the README predates the d2d backend, the drill suite and the retired Game.exe gates (ADR 0009, ADR 0013, docs/probes/postfix-soak1.md). A newcomer reading it today gets a stale picture.

## What Changes

- Display strings and prose that name the project say d.idle (README, GLOSSARY name lines, docs, container templates, setup wizard text).
- README.md says where the bot is now, how it runs and how we check it; CHANGELOG gets one entry.

no observable behaviour change: text only (the setup wizard strings change case of one word, no logic).
