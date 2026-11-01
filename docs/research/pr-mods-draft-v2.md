# Draft PR body: boss and minion spawn stats

> status: draft v2 (card modsprep, 2026-11-01), not posted; replaces [pr-mods-draft.md](pr-mods-draft.md) (the verifier's four edits applied). Proof: [pr-mods-proof.md](pr-mods-proof.md), addendum with the verifier verdict and the det state hashes: [pr-mods-evidence-addendum.md](pr-mods-evidence-addendum.md). Opens only with the user's go, in the next bundle slot.
>
> status: MERGED upstream as #157 on 2026-11-01 (maintainer); #152 and #153 are merged too (card upstreamsync). The text below is the pre-merge state.

Title: `Boss and minion spawn stats: base-level life, minion leveladd, ghostly and berserker`

---

## What this changes

Four rules in how a champion, unique, superunique or a unique's minion gets its stats when it spawns, one commit each, read in game.exe 1.14d and checked against it with a differential (game.exe's own code under `tools/emu` vs d2d's real `make_boss`).

**This is a gameplay-visible change.** In Act 1 Normal, d2d's champions and uniques today have 1.1-2.1x the life, 1.2-4x the defence and give 1.2-2x the experience of game.exe's (Bishibosh: about 1.9x life; at MonStats level 10 a champion has 1.14x the life); a unique's minions give one fifth of game.exe's experience and are 3 levels low (1.4-3.3x less AR, 1.3-2x less damage); ghostly champions walk a third slower; berserker champions are 3 levels low and give one fifth of the experience. After this PR they match game.exe. Champion and unique AR and damage do not change (d2d already reads them at the raised level, as game.exe's hit does).

### 1. A boss's life, defence and experience come from its MonStats level

`FUN_00573cb0` reads life, defence and experience once (`FUN_006538a0` mask 7, `0x573f78`) at the MonStats level; leveladd (`FUN_005a0e40`) and the champion's -1 (`FUN_005a0e80`) only change stat 12 afterwards. To-hit and damage are read at the hit (`FUN_005a4f50`, `FUN_00625480(unit, 12)`), so at the raised level. `rules::monster_stats` now takes life, defence and experience from the MonStats level's row and keeps the rest at the raised level.

### 2. A unique's minions get its leveladd

`FUN_005a2120` runs each fixed mod on the boss's party too (`FUN_0058f380` -> `FUN_005a2100`, flag 0); `FUN_005a0e40` tests no flag. A minion is 3 levels up and gives 5x experience (`boss_stats`, `Boss::minion`).

### 3. A ghostly champion is not slowed

`FUN_005a0e80` returns for mod 36 at `0x5a0ff8`; the -33 % store at `0x5a1062` is unreachable. `re/monsters.md` corrected.

### 4. A berserker champion keeps the fixed leveladd

`FUN_005a1280` doesn't call `FUN_005a0e80`, so the fixed mods' +3 and x5 stand.

## Evidence

`tools/emu/boss_level.py` runs, on a synthetic unit with every table loaded by game.exe, the real `FUN_00573cb0` (life rolled on the unit seed by the real `FUN_0045c3e0`), `FUN_005a2120` (or a minion's fixed-mod inits) and `FUN_005a4f50` (A1); `tests/game/test_boss_level --boss-level` runs d2d's real `make_boss` on the same cases with the same seed. 136,624 cases (every MonStats row x difficulty x kind, plus 100,000 random), field mismatches by commit:

| | before | after |
|---|---|---|
| 1 boss life / defence / experience | 286,774 | 0 (+ 803 stone-skin defence, see below) |
| 2 minion level / experience / AR / damage | 109,003 | 0 |
| 3 ghostly speed | 5,939 | 0 |
| 4 berserker level / experience / AR / damage | 26,868 | 0 (+ 176 BaseId 0x76 damage, see below) |

```sh
cd tools/emu
uv run python boss_level.py gen 100000 20261101 > cases.tsv
uv run python boss_level.py game cases.tsv > game.tsv
build/tests/test_boss_level --boss-level cases.tsv 2>/dev/null | grep ' | ' > d2d.tsv
uv run python boss_level.py compare cases.tsv game.tsv d2d.tsv
```

`tools/emu/boss_level.py` is under `tools/`: it needs the CODEOWNERS review.

An independent verifier re-derived the two levels from the disassembly and ran its own differential on the same builds (120,000 cases, four seeds, each commit removes only its own slice): D1 (life, defence, experience) 143,984 -> 0, D2 36,203 -> 0, D3 3,427 -> 0, B 7,300 -> 0.

## Limits

- Base level: the "MonStats level" in section 1 is game.exe's rule in Normal or classic games. In an expansion Nightmare or Hell game game.exe takes the area level instead (`FUN_0061a1b0` -> `FUN_0061dca0`, `0x573e48`, unless MonStats flags `0x6ce270` / `0x6ce280`). d2d models no area-level rule, before or after this PR, so NM / Hell expansion stats stay approximate.
- Follow-on callers, traced: `FUN_005a2620` (Fire Enchanted death blast), `FUN_005a6ff0` (death-timer missile blast), `FUN_005a3ef0` (suicide-minion blast), `FUN_005c4df0` (Corpse Explosion) and `FUN_005c56c0` (Revive) read MonLvl life at the raised level (stat 12) for their own damage or life. None writes the unit's own life, defence or experience, so none changes section 1. But d2d feeds the unit's rolled life into its equivalents (`Fight::fire_blast`, the monster Corpse Explosion, Resurrect, the player's Corpse Explosion), so this PR changes their input for boss units: closer to game.exe, still not equal. A follow-on candidate, not part of this PR.
- Not changed and left as they are: a unique's own mods on its minions (game.exe runs them with flag 0; d2d gives minions no mods), stone skin's doubling with two immunities (327 cases in the verifier's run, NM / Hell; the same before and after), the BaseId 0x76 damage halving (`// ponytail:` already in `boss_stats`), life rounding (game.exe 256ths vs whole points: 18,591 cases within 2 points, unchanged by the patch), rows without life (4,089, never a boss; unchanged), rows 0x137 / 0x138 (313, their extra roll is not in the harness). None of these depends on the four claims.
- The -33 % store at `0x5a1062` is dead code in game.exe (section 3), so no behaviour hides behind the removed slowdown.
- Hooked on the game side: the stat store, player count 1, skills, MonProp, events, the element part of `FUN_005a4f50`; the to-hit / damage % are added with d2d's formula.
- presubmit `siblings` WARN ("pet"): the word only (`summons.cpp:143` passes `level_add 0`, unchanged); players, mercs and pets have no MonUMod spawn mods. This PR shares no file with #152, #153, #155 or #156.
- Gameplay-visible: bosses die faster, unique packs pay about 2x, champion packs less; re-measure any soak baseline built on the old stats.

## Checklist

- [x] one claim per commit; tooling in its own commit
- [x] pr_guard 0 FAIL / 0 WARN (range and per commit)
- [x] test_game, test_rules, test_combat, test_protocol, test_monsters green
- [ ] det scenario state hashes: see the addendum (DETRESULT)
- [x] independent verifier: claims hold (D1 narrow form and D2 with limits above, D3, B)

Agent: Claude Code
