## Why

The user directed on 2026-11-01 that the bot must not hard-code the one default Sorceress build: it should end up flexible, with several builds (other Sorceress builds, later other classes) added as data plus small profile pieces. This card audits the code for what is fitted to the one build or the one class and designs the seams.

## What Changes

- New doc `docs/research/build-flexibility-audit.md`: inventory by layer with file:line, severity and seam per finding, what already works as data, proposed seams, a phased roadmap, and guardrail text for agent briefs.
- Docs only. No product code, no game, no edits to AGENTS.md, GLOSSARY.md or docs/decisions.md (proposed text goes in the doc for the coordinator).
