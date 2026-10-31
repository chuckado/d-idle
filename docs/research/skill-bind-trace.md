# Skill bind and cast: game.exe 1.14d vs d2d vs our bot (agent DECOMPSKILLBIND, 2026-10-31)

> status: research, current (card decompskillbind, decomp-queue item 4, primitive P26). Offline only: no game instance, no d2d slot; nothing pushed, posted or filed. d2d read at upstream main 72132b99 (a read worktree in the PR checkout dir on the worker host). Tags: V = read in game.exe 1.14d (the read-only copy, md5 2b3600c1...) with upstream's `tools/emu/disasm.py`, or run in its own code under unicorn (`tools/emu/skill_bind.py`); V-d2d = read in d2d's source; D = patched Skills.txt via mpq-cat; S = upstream docs (`docs/research/re/*.md`) or ours; I = inferred from V pieces; U = unverified.

## Answer in five lines

1. **Bind (a).** A skill reaches a hand only through C->S **0x3c** (u32 id, bit 31 = LEFT; u32 owner guid, -1 = the character's own). The server (`FUN_0054be70`) stores it only when the id is a Skills.txt row, the unit has an entry with that id AND owner, and the entry's level (base + item bonus, at least 0, capped) is above 0; otherwise it returns 3 and the hand keeps its old skill (V, oracle). The server never checks `leftskill`: only the client's picker hides non-left skills from the left list (V). A hotkey is a client table entry (skill id, hand, owner); pressing it sends the same 0x3c (S/I).
2. **Cast (a, b).** The cast packets carry **no skill id**: 05/08 (left at x,y), 06/07/09/0a (left on a unit), 0c/0f (right at x,y), 0d/0e/10/11 (right on a unit); flags: 0x8 held, 0x20 no approach (`FUN_00461700`, V). The server casts the hand's stored skill. The client gate `FUN_004610c0` -> `FUN_004d9fc0` -> `FUN_00647960` sends nothing on status 1 (mana), 2 (weapon / ammo / charges), 3 (not InGame), 4 (restrict), 6 (aura), 7 (level 0), 8 (skill delay or the client's cast clock); AttackNoMana skills fall back to Attack (V). A right-hand skill without InTown is refused in town by the client with a voice line (V, `0x4621ed`). The unit is the one under the cursor at the click (the hover kept at 0x7a6a94, S cursor-units.md); no unit -> the x,y packet.
3. **d2d (c).** World `cmd::SelectSkill` stores ANY id on either hand (commands.cpp:352-357, V-d2d) -> deviation **D1**, proven by the oracle (127,032 own-select cases, 81,756 differ, 0 under the proposed rule). `cmd::UseSkill` carries the skill id (game.exe's packets do not) and ignores the hands (V-d2d). d2d's client picker filters like game.exe (not passive, `leftskill` for the left, level > 0; V-d2d). Lockstep blind spot: d2d's hover (`hovered_monster()`) is a draw-time result, so under lockstep a CLICK picks no unit; the bot is not affected because its casts carry the unit id.
4. **Our bot (d).** On game.exe everything goes on the RIGHT hand: key (hotkey bound by the popup) then a right click on the monster's feet pixel (internal/input/controller.go:300). On d2d `BindSkills` binds nothing in the game: it maps "sk<id>" to the id and `cmd skill <id> x y [unit]` casts it directly, no 0x3c (internal/d2dcombat/rig.go:121, d2dinput/controller.go:255). So the P26 gate "BindSkills binds 2" on d2d tests the bot's map, not the game's binding.
5. **Deviations (e).** d2d: D1 (select validation) proven; D2 (cast packets carry the id) and D3 (shift / no-approach on a unit) are design / client gaps, not oracle-proven. Bot: B1-B5 below. PR candidate for D1 not built (low bot impact; d2d's SelectSkill has no item field).

## Q1 bind: the server's 0x3c (V unless tagged)

| Address | Rule |
|---|---|
| `0054be70` | 0x3c handler: size must be 9 (else 3); dword +1: bit 31 -> hand, low 31 bits -> id; id < DataTables +0xba0 (357 rows, D) else 3; `FUN_006439b0(unit, id, guid)` finds the entry else 3; `FUN_006442a0(unit, entry, 1)` level 0 -> 3; then `FUN_005701b0(unit, hand, id, guid)` and the client cache refresh `FUN_00622f10` (hand 1) / `FUN_00622ea0` (hand 0); returns 0 |
| `006439b0` | the unit's skill list: unit +0xa8 -> +4 first entry; entry +0 = Skills.txt record (word +0 = id), +4 next, +0x34 owner guid; match on id AND guid |
| `006442a0` | level: entry +0x28 (base); with flag 1 and an own entry (+0x34 == -1) plus `FUN_00644180` (the item bonus); then max(v, 0) and min(v, `FUN_00611830(0)`) (the cap's table value: U) |
| `005701b0` | apply: id 5 (Left Hand Swing, D) -> nothing; hand 1 -> `FUN_00643bc0` writes the list's +0x08; hand 0 -> clears state 0x55 if set, `FUN_00643c50` writes +0x0c, stops the old right skill's aura state (record +0x80) when the old skill had the aura flag, starts the new one's aura (`FUN_0056ff10`) when it has it. So **bit 31 = LEFT** (+0x08 is the left skill: getter `00620190` -> `00643930` reads +8, right `006201d0` -> `00643960` +0xc) and only the right hand runs an aura. Read, not run |
| Skills.txt flags (record +4, bit n at byte n/8) | from the loader descriptors at 0x6166xx..0x616bxx: passive 4 (I: index in a register), aura 5, InTown 8, InGame 10, AttackNoMana 28, leftskill 30, scroll 36; masks at 0x6ce268.. (1 << k) |

Client side (V): the picker list `FUN_004aa3a0` (one call per hand) skips an entry whose record is passive (+4 & 0x10), whose entry flag (`FUN_006446a0` & 8) is set, and on the LEFT list one without `leftskill` (+7 & 0x40); rows 0..3 are SkillDesc ListRow pages, the top row holds item entries (charged / scroll counts). The picker click `FUN_004a9bd0` hit-tests 60 icons (48 x 48), refuses a charged entry with 0 charges, and sends 0x3c with `id | (0x7c07fc ? 0x80000000 : 0)` and the entry's owner; the hand changes when the server's answer comes, not at the click (I: no local write in `004a9bd0`). Hotkeys: 16 slots at 0x7c06c8 (id; a cleared slot is -1), 0x7c07b8 / 0x7c0768 beside it; `004a9d70` clears every slot bound to a removed entry and sends C->S 0x51 for it (V for the clear; 0x51's meaning S: the hotkey save). What a hotkey press sends is S (upstream controls.md, skillbar.hpp: "a hotkey selects into its hand" = 0x3c), I for 0x3c.

Early Sorceress (D, patched Skills.txt): **left-capable**: Attack, Throw, Fire Bolt, Charged Bolt, Ice Bolt, Inferno, Glacial Spike. **Right only** (leftskill empty): Frozen Armor, Static Field, Telekinesis, Frost Nova, Nova, Teleport, Blizzard, the tome skills. InTown: Frozen Armor, Telekinesis, Warmth, Book of Identify; NOT Teleport, Static Field, Frost Nova, Fire Bolt, Book of Townportal (its refusal in town is the tome's own path, town-portal-trace.md). Mana: Fire Bolt 5 << 7 = 2.5, Static Field 9, Frost Nova 9 (+1 / level), Teleport 24 (-1 / level), Frozen Armor 7.

## Q2 cast: packets and client gates (V unless tagged)

| Address | Rule |
|---|---|
| `00461700` | packet chooser, the flag word at [esi]: bit 0 left, bit 1 right, 0x8 held, 0x20 no approach; unit form (arg 1): left 06 / 07 (0x20) / 09 (0x8) / 0a (both), right 0d / 0e / 10 / 11, args (type, id); x,y form: left 05 / 08 (held), right 0c / 0f. Both hand bits set send both. Nothing when the cursor's y is past the play area (`FUN_00454970`) |
| `004610c0` | gate before a send: non-player unit -> go; status = `FUN_004d9fc0`; with status 1, 2 or 4 and AttackNoMana (+7 & 0x10) the entry becomes Attack's (`FUN_006439f0(unit, 0)`) and is retested; status 0 or 5 -> go; else voice `0x711ddc[status]` (1 -> 0x15 needmana, 2..4 -> 0x13 impossible, others silent) and no packet |
| `00647960` | status (shared with the server's checks): no record or not InGame -> 3; level 0 -> 7; aura -> 6; passive -> 5; `FUN_00647640` (scroll quantity, Throw / Left Hand weapons) and `FUN_00643f80` (itype / etype weapon match) -> 2; `FUN_00647540` mana cost -> 1; `FUN_00644060` restrict (0: not shapeshifted, 2: one of State1..3) -> 4; `FUN_006440f0` minmana -> 1; `FUN_00647840` an item entry's charges (stat 0xcc) -> 2; `FUN_006478f0` the skill's delay (state 0x79 and `FUN_00646ca0` > 0) -> 8 |
| `004d9fc0` | status 0 for the local player but the client clock `FUN_0044da90` below 0x7a04fc -> 8 (a client cast clock: I) |
| `004621ed..00462221` | right hand, the entry not InTown (+5 & 1) and the player in a town level (`FUN_0061ab00`) -> class voice 0x18 and no packet |
| `0054be70`-style server handlers `549ad0` / `549ba0` (S net-packets.md) | x,y refused past 0x32 subtiles on an axis; unit type 0, 1 or 3 unless the skill allows others; approach: a melee-range skill walks up first (`FUN_00548a50`) |

Target resolution (b): the unit is the client's hover unit (`FUN_00467a10`, 0x7a6a94, picked every frame by `FUN_00467ac0` with the hit test `FUN_00466870`, S cursor-units.md). Over a monster a click sends the unit form; else the x,y form at the cursor. Left click on the ground with no unit walks (the left skill's x,y form only with Stand Still / shift held: flag 0x7a0664, S controls.md); a right click always casts the right skill. Shift on a unit sets 0x20 (07 / 0e): no approach, so a melee skill swings from where the player stands (I from the flag's packet split; the shift key read itself not traced). A targeted skill with no unit casts at the cursor point; Teleport and the bolts are x,y skills anyway (V packets, I for the per-skill server reaction).

## Q3 d2d (upstream 72132b99), read (V-d2d)

| Part | d2d | vs game.exe |
|---|---|---|
| select | `cmd::SelectSkill{skill, left}` (protocol.hpp:57); World (commands.cpp:352-357) stores it in the header's left_skill / right_skill for ANY id, sets the right-hand aura from Skills.txt | **D1**: no id range, entry, owner or level check; no item field (game.exe's guid) |
| cast | `cmd::UseSkill{skill, x, y, unit, left, item}`: the id rides the command; World (commands.cpp:584-668) gates: InTown (`not_in_town` voice), `skill_refused` (Throw), self cast, Whirlwind, missile / spot at x,y when no live unit, summons, else a live unit: walk up, attack (`engage` casts a missile / spot skill from where it stands) | **D2**: game.exe's packets have no id; d2d's own client always sends the hand's id, so equal in play; a devctl or net client can cast an unselected skill |
| gates | `Fight::cast` / `cast_missile`: level <= 0 or mana short -> refuse (`need_mana` voice); `pmode >= 0` busy | status 3, 4, 6, 8 and the minmana / weapon rules not modelled (U which matter early: none of the early Sorceress skills has restrict, delay or itype) |
| client | `Town::input` (town/input.cpp:55-96): left click on a live hovered monster -> UseSkill left with the unit (approach); Stand Still + ground -> left skill at x,y; right click -> right skill (unit if hovered) | **D3**: Stand Still + a unit still approaches (game.exe 07 / 0e: no approach) |
| picker | skillbar.hpp:77-98 `usable`: not passive, `leftskill` on the left, level > 0 or a charged entry | equal to `004aa3a0` (D: same columns) |

Lockstep blind spot: `hovered_monster()` comes from the last drawn frame; lockstep draws none, so a scripted CLICK (`click` / `rclick` verbs) over a monster finds no unit and walks (left) or casts at x,y (right). The bot's d2d casts use `cmd skill <id> x y <unit>` with the unit from `NearestAlive` within one cell of the target (internal/d2dcombat/input.go:63), so they never depend on the hover. Only the fallback `CastAt` path (no id known: key + `rclick`, d2dinput/controller.go:255) would hit the blind spot.

## Q4 our bot (V by read unless tagged)

- game.exe backend: `EnsureHotkeys` binds each layout skill to its F-key through the RIGHT-skill popup (combat/bind.go:191; VERIFIED LIVE 2026-10-28 per its comment); `CastAt` = `Key(hotkey)` then `Click(Right, feet pixel)` (input/controller.go:300). So Frozen Armor, Static Field, Frost Nova, Teleport (right-only) and Fire Bolt all go through the right hand: consistent with leftskill. The left hand stays Attack.
- d2d backend: `BindSkills` (d2dcombat/rig.go:121) gives every catalog skill in the snapshot a pseudo key "sk<id>" and calls `Fighter.SetHotkeys`; no 0x3c is sent. `CastAt` -> `cmd skill <id> x y [unit]`, right hand (left flag 0), unit = the living monster within 1 cell of the target.

| id | finding | file:line | tag |
|---|---|---|---|
| B1 | P26's "BindSkills binds 2" on d2d checks the bot's own map, not a game binding: the game hands are never set by the bot on d2d, and the d2d cast takes the id from the command. A d2d drill passing P26 says nothing about the game.exe bind path | d2dcombat/rig.go:117-141 | V |
| B2 | `BindSkills` keys every skill in `snap.Player.Skills` the catalog knows, with no level, passive or charged check; on game.exe the server refuses a select of a level-0 entry (D1 rule), so the same list could not be bound there | d2dcombat/rig.go:131 | V (code), effect I |
| B3 | game.exe `CastAt` presses the key and clicks with no wait for the server's answer; the client gate (`004610c0`) tests the hand the client knows, which is the OLD skill until the 0x3c answer arrives, while the server casts the new one (packets in order). Mana / in-town refusals can be judged on the wrong skill for one cast | input/controller.go:300 | I |
| B4 | game.exe: a right click whose pixel misses the monster's hit box sends 0x0c (x,y), which for a bolt still flies at that point; the bot clicks the feet pixel, not the sprite (interact uses `interactProbe` offsets, casts do not) | input/controller.go:300-309 | V (code), effect I |
| B5 | Nothing in the bot models status 8 (skill delay) or the right-hand in-town refusal; the early Sorceress has no delay skill before Blizzard (delay 45, D) and the bot casts in town only Frozen Armor (InTown 1, D) | combat (grep) | V (grep) |

## Q5 differential (tools/emu/skill_bind.py)

game.exe's own `FUN_0054be70`, `FUN_006439b0`, `FUN_006442a0` under unicorn; hooks: `FUN_00644180` (bonus), `FUN_00611830` (cap), `FUN_005701b0` and `FUN_00622f10` / `FUN_00622ea0` (recorded). Cases: exhaustive over id 0..360 x hand x {no entry, own entry base -1..3 x bonus 0..2, item entry (right guid, wrong guid, own select)} + size and cap edges + 250,000 random (seed 7; 268,058 lines). d2d "before" = commands.cpp:352-357 line for line (every select stored); "after" = the proposed rule (stored when the id is a row and the level for that id and owner is above 0). Output `i set hand id guid`.

| Domain | cases | before | after |
|---|---|---|---|
| own select, size 9 (the d2d claim) | 127,032 | 81,756 | 0 |
| all (incl. item guids and wire sizes, not d2d-comparable) | 268,058 | 216,932 | 0 |

Refusal causes in the first 118,058-line run: id past the rows 23,082, no entry for id + owner 27,961, level 0 7,404, size 32,981. Commands: `uv run python skill_bind.py gen 250000 7 > cases.txt; ... game cases.txt > g.txt; ... d2d before|after cases.txt | diff g.txt -`.

## Deviations

| id | where | what | proof | status |
|---|---|---|---|---|
| D1 | d2d World select | any id stored on either hand; game.exe refuses a non-row id, a missing entry (id + owner) and a level-0 entry (returns 3, the hand unchanged) | oracle 81,756 / 127,032 -> 0 against the rule | PROOF incomplete: the after side is the rule in Python, not patched d2d code; no harness, no verifier. Bot impact low (d2d's client and the bot never send such a select; World casts ignore the hand) |
| D2 | d2d protocol | UseSkill carries the id; game.exe casts the hand's skill | design read (V-d2d, V packets) | not a PR (architecture); note for any net-parity work |
| D3 | d2d client / World | Stand Still on a monster approaches; game.exe's 07 / 0e do not | read (I for the shift key) | client gap, U on the server side of 07 / 0e |
| D4 | d2d gates | status 6 / 8 / restrict / minmana / InGame not modelled | read | none affects the early Sorceress (D) |

## Answer the drills should expect (P26)

- game.exe: after binding, the right-skill button shows the skill; a select of a skill with 0 points does nothing (the button keeps the old skill). A right click on a monster with Fire Bolt: mana -2.5 (5 << 7 / 256), one missile; Static Field / Frost Nova -9; Teleport -24 at level 1; a right-only skill never appears in the left picker. Casting a non-InTown right skill in town: a voice line, no mana spent.
- d2d: `dump`'s `lskill` / `rskill` follow `cmd select` for ANY id (D1); a `cmd skill <id>` cast works whatever the hands hold (D2); mana deltas as above (level > 0), none at level 0.

### DRILL RECIPE (live check of P26, for the drills lane; not run)

On a hidden slot, throwaway Sorceress, lockstep, every cheat ledgered: (1) fresh clvl 1: `cmd select 54 0` (Teleport, 0 points) then `dump` -> expect rskill=54 on d2d (D1; game.exe would keep the old one); (2) `debug points` to learn Fire Bolt and Static Field (clvl 6 via the level cheat if needed), `BindSkills` -> 2; (3) the lab picker verb `debug skill right <id>` (d2dctl.SetSkill) for a 0-point skill -> expect `err not usable` (the lab verb follows the picker rule, unlike `cmd select`); (4) `debug spawn zombie1 1` in Blood Moor (`debug goto 2`), `cmd skill 36 x y <unit>` -> mana -2.5 (read the 8.8 value), a missile, the zombie's life drops; `cmd skill 42 x y` (Static Field) -> -9 mana; (5) in town `cmd skill 42 ...` -> no mana change and the not-in-town cue; (6) `step` past the cast and leave the character in town, no panels. Pass: (4) and (5) as stated; (1) records D1.

## Verified vs not

V: every game.exe address in Q1 / Q2 (disasm.py reads), the 0x3c rule run in the oracle. V-d2d: Q3 file:line. D: Skills.txt columns. I: what a hotkey press sends (0x3c), the client cast clock, shift = flag 0x20, B3 / B4 effects. U: the cap value of `FUN_00611830(0)`, state 0x55, the server reaction to 07 / 0e for each skill kind. Not run live (no game).
