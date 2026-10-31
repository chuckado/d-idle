# Tasks

```yaml
change: decompskillbind
status: archived
agent: decompskillbind
branch: agent/decompskillbind
worktree: ~/code/d2lod-wt/decompskillbind
skip_specs: true
merge: 0815238
owns:
  - docs/research/skill-bind-trace.md
  - tools/emu/skill_bind*
  - openspec/changes/decompskillbind/**
  - docs/decomp-queue.md
  - docs/research/README.md
  - docs/INDEX.md
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, V-d2d, D, S, I or U with an address or file:line"
  - "tools/emu/skill_bind.py runs game.exe's code on >= 100,000 cases: mismatches before > 0, after the proposed rule 0"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "Bind: only C->S 0x3c sets a hand (bit 31 = left); game.exe's server (0x54be70) refuses a non-row id, a missing entry (id + owner) or a level-0 entry, never checks leftskill (the client picker does). Cast packets carry no skill id (0x461700: 05/08, 06/07/09/0a, 0c/0f, 0d/0e/10/11; 0x8 held, 0x20 no approach); client gates via 0x647960 statuses 1-8 and a right-hand in-town refusal. d2d D1: World SelectSkill stores any id (oracle 127,032 own-select cases, 81,756 differ, 0 under the proposed rule; PROOF incomplete: rule model, no patched d2d, no verifier; PR candidate not built). D2 UseSkill carries the id, D3 Stand Still on a unit approaches, D4 unmodelled gates (none early). Bot: on d2d BindSkills binds nothing in the game (B1), casts carry the id and the unit (no hover dependence); on game.exe all skills go to the right hand by hotkey + right click (consistent with leftskill). Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read AGENTS.md, decomp-queue rules, the PR standard, earlier traces, P26
- [x] 1.2 Trace 0x3c (server and client picker / hotkeys) in game.exe
- [x] 1.3 Trace the cast packets, the client gates and the in-town refusal
- [x] 1.4 Read d2d upstream main: select, cast, client input, picker, the lockstep hover
- [x] 1.5 Read the bot: BindSkills, CastAt on both backends, EnsureHotkeys
- [x] 1.6 Oracle tools/emu/skill_bind.py: 268,058 cases, before 216,932 (own 81,756 / 127,032), after 0
- [x] 1.7 Trace doc, drill recipe, decomp-queue row and result line, index lines
- [x] 1.8 PR candidate for D1 (not built: low impact, needs an item field on cmd::SelectSkill; left for a later card)
