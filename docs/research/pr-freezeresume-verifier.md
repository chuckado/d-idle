# Independent verifier: fix/freeze-resume (freeze end + aidel resume) (VERIFYFREEZERESUME, card verifyfreezeresume, 2026-11-01)

> status: verifier record for the local upstream candidate `fix/freeze-resume` (claim c6cc4f12, tooling efacc7f4, base upstream main 72132b99) in the worker host's PR checkout. Author's proof and draft: `docs/research/pr-freezeresume-proof.md` / `pr-freezeresume-draft.md` (branch agent/freezeresume, 7f2f848). Offline; no game, no d2d slot; nothing pushed or posted; the author's branch, worktrees and oracle untouched (my builds are separate detached worktrees). Verdict at the end.

Tags: **V read** disassembled by me with upstream's `tools/emu/disasm.py`; **V run** game.exe code run under upstream's emulator in MY oracle (`tools/emu/verify_freeze_resume.py`), or real d2d builds run with MY probe (`tools/emu/test_verify_freeze_resume.cpp`); **S** d2d source at the named commit; **D** 1.14d MPQ tables; **I** inferred; **U** not traced.

## 1. The event queue, run for real (the author's open risk)

The author's oracle hooked the queue (an event filed for X ran at X by construction). Mine hooks none of it:

| Code | What it does | Tag |
|---|---|---|
| FUN_005417d0 -> FUN_005416b0 | files an event: frame <= game +0xa8 becomes +0xa8 + 1; event from the pool (FUN_00540df0 / FUN_00540c30 / FUN_00540bb0, 600 a block), appended at the newest end of bucket `group x 64 + (frame & 63)` (group = 0x6e0b9c[unit type]: player 0, monster 1, missile 2, object 3, item 4), linked on the unit's own list (unit +0xdc) | V read, V run |
| FUN_005414d0 | one frame: buckets `frame & 63` of group 2 (missiles), 0 (players), 1 (monsters), 3, 4, each also its every-frame list | V read, V run |
| FUN_00541060 (monster group) | walks the bucket oldest first through +0x1c, keeping the next pointer at queue +0xa18 so removals during dispatch are safe; runs an event only when **event +4 == game +0xa8** (other frames of the same bucket, X + 64k, stay); callback +0x2c if set, else FUN_005a7f80 | V read, V run |
| FUN_00540e60 / FUN_00540cd0 | removes every event of the type (arg 0: any argument) from the unit's list; an event being dispatched is only marked | V read, V run |

So an event filed for frame X runs in the monster group of frame X, exactly; the end + aidel event of 0x57b170 runs at end + aidel. **No off-by-one in the queue** (V run: single freezes of 1, 30, 250 frames give first think at end + aidel for every aidel value, e.g. andariel Normal 30-frame freeze: 45; 1-frame: 16; 250-frame: 265; lengths > 64 cross the bucket wrap correctly).

Re-derived, agreeing with the author (V read): 0x57b170 (dying + FUN_0063a4a0: nothing; else state 1 off, FUN_00553570, then for a monster with a MonStats row: aidel = row +0x4f, or row +0x4f + game +0x6d when game +0x6a or +0x74 is non-zero; `jg` so 0 -> 15; FUN_00540e60(unit, 2); type 2 filed at game +0xa8 + aidel); FUN_0057b230 (end = now + length / MonsterFreezeDivisor for a monster; new list with callback 0x57b170 or the existing list's end raised only when later; timer 0xc at the list end; FUN_00540e60(unit, 2); type 2 at now + length + 1); FUN_005a7f80 (types 0, 1, 2, 6, 7, 9, 10, 11, 13, 14 gated on state 1 unless FUN_005541b0; 11, 13, 14 have no handler; 3, 4, 5, 8, 12 ungated; byte table 0x5a7fe8 = 00 00 00 01 01 00 01 01 01 00 and handler table 0x6e2490 read from the image, V run). The author's gated list omits 11 (no handler, harmless).

Frame order matters (V read FUN_005414d0): a missile's freeze in frame f lands before the monster group of f, so a think due at f is removed, and a refreeze on the end frame raises the list before the timer fires.

## 2. Own differential

- **Oracle** (`verify_freeze_resume.py game`): run natively: the queue above, FUN_0057b230, FUN_0057aae0 (stun), the state-list code and the expiry 0x5a7ef0 -> FUN_00627460 -> FUN_00626c00 -> FUN_006269f0 -> 0x57b170, the kind tests FUN_0063e9f0 / FUN_005a0180 / FUN_0063ee90, FUN_0057af30. Hooked (docstring): allocator, unit -> game, client update, stat getters / refresh, the AI driver 0x5b1740 (output: first frame it is entered), the stun's end callback 0x56e900 and the chill FUN_0057af80 (both fault in stat code under the emulator; V read: neither files nor removes a type-2 event, section 4). Freezes are applied before FUN_005414d0 of their frame (= the missile group). Data: MonStats aidel columns at +0x4f..+0x51, DifficultyLevels MonsterFreezeDivisor (D), game +0x6a = 3 (I: single player, upstream net-join doc).
- **d2d probe** (`test_verify_freeze_resume --vfr CASES n|g`): real `Fight::strike` with the freeze argument (the Glacial Spike path, missiles.cpp:662) at each case frame, the real `monster_update` every tick from two frames before the first event (the author started one tick after the freeze), a zombie1 body carrying the case row's loaded aidel; prior stun written as strike.cpp:448 writes it. Order **g** = game.exe's (strike, then update); order **n** = d2d's own (Fight::world runs monster_update before the player's missiles, fight/tick.cpp ~302 vs ~408, S).
- **Cases** (`gen`): own LCG (`x * 1664525 + 1013904223`, seed masked to 32 bits, top 24 bits); grid = 14 distinct MonStats aidel triples (D, first row of each) x 3 difficulties x plain / minion x 15-16 shapes (pending 0 / 5 / 31 / 100, lengths 1 / 30 / 250, prior freeze longer / shorter, refreeze inside / on the end frame / in the delay window / one before the think / two refreezes, prior stun ending after / inside, zero-length in NM / Hell) + specials + 30,000 random. Seeds 0xF3EE0001 and 0x1A5C30002 (masked to 0xA5C30002): 31,484 cases each (cases md5 69ed5a49ad25 / d81629a735cf; game md5 d46c4c57a0ab / 42981123f84f).
- Kinds: game.exe chilled 12,111 cases of seed 1 (all specials: champion, unique, superunique, boss row) and froze all plain / minion ones (V run).

Results, first think (swing) frame, game.exe vs d2d, differing / cases:

| class (plain + minion unless said) | base, order g (s1 / s2) | claim = head, order g (s1 / s2) | head, order n (s1) |
|---|---|---|---|
| single freeze | 7,384 / 7,416; 7,284 / 7,314 | **0 / 7,416; 0 / 7,314** | 0 |
| prior freeze | 4,027 / 4,034; 4,232 / 4,237 | **0; 0** | 19 |
| refreeze inside | 3,388 / 3,393; 3,304 / 3,305 | **0; 0** | 0 |
| refreeze on the end frame | 112 / 112; 106 / 106 | **0; 0** | 0 |
| refreeze in the delay window | 516 / 516; 492 / 492 | **0; 0** | 2 |
| refreeze on the think frame (end + aidel) | 38 / 38; 47 / 47 | **0; 0** | 38 / 38 |
| refreeze after the think | 1,412 / 1,418; 1,366 / 1,379 | **0; 0** | 0 |
| think due on the freeze frame | 253 / 253; 239 / 239 | **0; 0** | 253 / 253 |
| prior stun (state 21) running at the freeze | 1,878 / 1,881; 1,970 / 1,978 | 397 / 1,881; 422 / 1,978 (d2d later by 5..11 typ.) | 398 |
| zero-length freeze (NM / Hell, raw < divisor) | 308 / 312; 320 / 320 | 308 / 312; 320 / 320 | 308 |
| specials (champion, unique, superunique, boss row) | 10,659 / 12,111 | 11,912 / 12,111 | 11,814 |

Claimed frozen classes together: base 17,130 of 17,180 (s1), 17,070 of 17,119 (s2); claim / head **0 of 17,180 and 0 of 17,119**, every aidel triple and difficulty included. **No off-by-one** under game.exe's frame order: no delta of +-1 anywhere in the claimed classes.

Isolation (V run): claim alone and head give byte-identical probe output on both seeds and both orders (md5 of the case lines: s1-g bdb35fd39b2a, s1-n 11946f6770fb, s2-g 059e6933a30f, s2-n 06370481d788); base differs. The tooling commit adds no d2d behaviour.

Residual classes (none is the claim's error):

1. **Prior stun** (author's ponytail, confirmed): game.exe refiles at the freeze end + aidel even under a longer stun (the driver is entered while stunned; whether it acts there is U), d2d keeps one stand end equal to `stunned_until` and refiles nothing. Also d2d's stun site assigns, so a stun landing during a freeze can shorten the freeze (S strike.cpp:448), and the claim then treats it as a stun.
2. **Zero-length freeze**: game.exe freezes for 0 frames and thinks at now + 1 + aidel (V run); d2d's strike with freeze 0 does not freeze. Unreachable on upstream (no divisor, so no zero length) but latent: had d2d a zero freeze, the claim would refile at now + aidel (stun_until = now), one frame early. Say so for fix/freeze-rules-v2, which the author says skips them.
3. **d2d's own tick order** (order n): monsters update before the player's missiles strike, the reverse of FUN_005414d0. A think due on the frame a freeze lands still runs in d2d (253 / 253), and a refreeze on the resume frame comes after the think (38 / 38). Pre-existing on base too, independent of the claim; the author's harness (strike before update) cannot see it. Not a reason to hold this PR; worth a ponytail line or a separate card.
4. Specials: freeze-rules C3 (game.exe chills them, V run here), as the author says; the claim adds the aidel delay to their (wrong) d2d freeze.

Dead end: a second observable (first tick where `next_act` moves other than by the refile) proved useless: the Zombie's swing does not move `next_act` at the swing frame (it lands 16 frames later, the attack end), so only the swing column is used.

## 3. Side claims

- **Only plain monsters and minions are frozen in game.exe**: HOLDS (V read FUN_0057b230 + V run). Before the freeze path a monster is chilled when FUN_0063e9f0 (MonStats +0xc & [0x6ce280] = 0x40, the boss flag), FUN_005a0180 (monster data +0x16 & 8) or FUN_0063ee90 (class 271 / 338 / 359 / 560 / 561: the hirelings) is non-zero; players always chilled. Flag 8: FUN_005a0320 sets it (V read) and FUN_005a48c0 then ORs 5 for champions (0xd, V read), uniques carry 8; superuniques |= 2 plus 8 (S, upstream superuniques.md / monsters.md); minions only 0x10, so frozen. Coldeffect >= 0 (FUN_0057af30) means neither.
- **Stun and chill end callbacks refile nothing**: HOLDS (V read). 0x56e900: FUN_005c0b50 (frees, one fog free), state off, FUN_00623f50 (stat refresh), FUN_00646f20, FUN_00575900 for players only; no FUN_005417d0 / FUN_00540e60 in it or its direct callees (FUN_00623f50's deep callees U). 0x57ad80: states 0xb and 0x6b off, FUN_00623f50. The stun's application FUN_0057aae0 and the chill FUN_0057af80 file only type-0xc timers.
- **On d2d Frost Nova and Shiver Armor chill, Glacial Spike and Frozen Armor freeze**: HOLDS (S, 72132b99). The only `freeze > 0` caller of `Fight::strike` / `area` is pSrvHitFunc 13 (missiles.cpp:687, Glacial Spike's hit, auralen); Frost Nova's missile strikes through missiles.cpp:595 without freeze; Shiver Armor lands a cold missile blow (chill, states.cpp:163); Frozen Armor sets `stun_until` (states.cpp:160, the A3 merc's states.cpp:194).

## 4. Code review notes (S, c6cc4f12)

1. The refile sits before the GH / BL and stand checks: right, the end + aidel think replaces `next_act` whatever the mode.
2. `stunned_until` vs `stun_until` equality as the freeze / stun discriminator works for every case above; it misreads a freeze whose end equals a stun's end to the frame (rare; game.exe would refile).
3. Zero-length freezes: see residual 2; one line in the PR body.
4. LOW: the comment cites FUN_0056e900 for the stun callback; it is a callback address (0x56e900) like 0x57b170, the same pr_guard point the author fixed for the others.

## 5. Verdict

**READY WITH CHANGES** (body text only, no code change): the claim holds under the real queue with no off-by-one (0 of 34,299 claimed frozen cases on two seeds, base 34,200; claim alone = head). Changes: (a) draft Limits: say the verifier ran the real queue (FUN_005416b0 / FUN_005414d0 / FUN_00541060) and drop "the per-frame walk is not run"; (b) add the zero-length note (game.exe now + 1 + aidel; the claim would give now + aidel if d2d ever makes one) and the d2d tick-order residual (monsters update before missiles; pre-existing, out of scope); (c) gated types: add 11 (no handler); (d) 0x56e900 as an address. Still owed before the user's go: det state hashes (needs a d2d slot; not run here).

## Commands (worker host)

```sh
# detached worktrees of 72132b99 (base), c6cc4f12 (claim), efacc7f4 (head) under the PR checkout, each with
# tests/game/test_verify_freeze_resume.cpp; Release build of target test_verify_freeze_resume
python verify_freeze_resume.py gen 30000 0xF3EE0001 > c1.tsv            # PYTHONPATH = upstream tools/emu, D2_MPQ_DIR set
python verify_freeze_resume.py game c1.tsv > g1.tsv
build/tests/test_verify_freeze_resume --vfr c1.tsv g > d-g.tsv            # n: d2d's own tick order
python verify_freeze_resume.py compare c1.tsv g1.tsv d-g.tsv 1
```

## Shared-file lines for the merger

- docs/research/pr-bundle-queue.md: `- VERIFYFREEZERESUME (2026-11-01, docs/research/pr-freezeresume-verifier.md): fix/freeze-resume READY WITH CHANGES (body only): real event queue run under the emulator, no off-by-one; claimed frozen classes 0 of 34,299 on two seeds (base 34,200), claim alone = head; residuals: prior stun (ponytail), zero-length (latent), d2d tick order (pre-existing); det hashes still owed`
- docs/BACKLOG.md: `- FREEZERESUME: apply the verifier's draft edits (docs/research/pr-freezeresume-verifier.md section 5), run det state hashes; separate card for d2d's tick order (monster_update before the player's missiles, reverse of FUN_005414d0)`
- docs/INDEX.md, research: `- docs/research/pr-freezeresume-verifier.md: independent verifier of fix/freeze-resume (real event queue, own differential, side claims)`
- CHANGELOG.md: none.
- GLOSSARY.md: none missing.
