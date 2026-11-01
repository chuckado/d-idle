# Independent verifier: d2d candidate fix/blizzard-shard-offsets (card verifyblizzard)

> status: verdict READY WITH CHANGES (2026-11-01, offline). Candidate: upstream main 7d5b3a5c + 9a1a90eb (tooling) + 681d52a9 (claim), local scratch checkout on the worker host; nothing pushed or posted. Author evidence: [sk-ground2-trace.md](sk-ground2-trace.md), [pr-blizzard-draft.md](pr-blizzard-draft.md) (on agent/skground2), read only AFTER the results below.

Tags: **V run** executed here (game.exe code under the emulator, or a real d2d build); **V read** disassembled or read here; **S** source reading or opinion; **U** not verified.

## Claim under test

681d52a9 changes one expression in `Fight::missile_tick` (srv_do 10 / 25): shard offset `rng(2 radius - 1) - (radius - 1)` (-6..6) becomes `rng(2 reach) - reach`, `reach = radius - 1` (-6..5), to match game.exe's FUN_005a9820.

## 1. game.exe's rule, own harness (V read + V run)

Read from the disassembly only (upstream `tools/emu/disasm.py`, game.exe 1.14d): FUN_005a9820(eax every, esi missile; game, radius, submissile, mask), callee pops 0x10.

- every == 0, or age % every != 0 (signed idiv; age = missile data +0xe minus +0x10, both int16, FUN_0064a3b0): nothing.
- owner (FUN_00552fd0) or room (FUN_00620bb0) missing: nothing.
- seed set to {cx + age, 666} (FUN_00650e40; cx = the centre's subtile x, path +2); r = radius - 1; dx = roll(2r) - r, then dy = roll(2r) - r (FUN_0045c3e0: n < 1 gives 0 without a step; else one LCG step, then `lo & (n-1)` for a power of two, else `lo % n`).
- FUN_0064cb30(room, cx + dx, cy + dy, mask) non-zero: no shard; else FUN_0059fa30 makes it. Callers: 0x5aeb3e (srvdo 10, Blizzard) passes mask 5, 0x5af95c (srvdo 25, Eruption) mask 0x45. Two more callers (srvdo 7 at 0x5ae92a, srvdo 17 at 0x5af2b5) take radius / every from Missiles.txt params; their rows (guidedarrow, bonespirit, monbonespirit: no SubMissile; cairnstones) are not touched by the candidate.
- Data (D, read from the 1.14d MPQs here): Blizzard calc1 = par1 = 7, calc2 = par3 = 4; Eruption calc1 = 7, calc2 = 6. So r = 6 for both skills at every level.

Harness `tools/emu/verify_blizzard.py` (own generator, own model, 32-bit masked seeds): game.exe's FUN_005a9820 runs natively with FUN_0045c3e0, FUN_00650e40, FUN_0064a3b0, FUN_0045adf0 / FUN_0045ae20 / FUN_006488c0 / FUN_00648900, FUN_0064a280 / FUN_0064a210 and memset unhooked; hooked: owner, room, the collision test (a per-case rule, every query logged with its mask) and the missile maker (logs x, y, id).

| run | cases | model vs game.exe differ | outcomes |
|---|---|---|---|
| gen 120000 seed 20261101 | 120,000 | **0** | none 92,759, shard 23,609, blocked 3,632 |

Covered: every in {0, 1, 4, 6, 7, -4}; radius -3..100 incl. 1 (2r = 0), 3 / 5 / 9 / 17 / 33 (2r a power of two: the mask branch); cx 0, 65535, 32767 / 32768; ages incl. negative and int16 wrap (frame / start pairs); owner or room missing; block rules none / all / parity / two hashes (46,050 cases with a rule), masks 5 and 0x45. Radius 7: 22,805 rolls, all on -6..5 on both axes, all 144 (dx, dy) cells reached (129..194 each). A blocked spot makes no missile and the query carries the caller's mask (0 ODD / MASK / MAKE flags). Radius <= 0 gives +(1 - radius) offsets in game.exe (r negative, roll 0): unreachable with the shipped data.

## 2. d2d base vs head on real builds (V run)

Own probe `tools/emu/verify_blizzard_mode.inc`, pasted into test_game.cpp of two throwaway clones (base 7d5b3a5c, head 681d52a9), Release, `-ffp-contract=off`, built on the worker host (capacity2 said BUSY (cpu) on phantom iowait, load 0.06 per core; builds ran at nice 15). Per case: world rng set to the case's {low, high}, the REAL `Fight::fire` (Caster at a random target offset, level 1..30), then per frame the REAL `Fight::missile_tick`, pending joins, spent go (as `Fight::world` orders them); each new shard's offset from the centre in subtiles. 5,000 casts (`gen-d2d 5000 777`, edge rng states {0,0}, {0,666}, {max,max}, {1,0}, {2^31, 2^31-1}, {12,13} first), different from the author's (xorshift seeds 1 / 7, `world()` tick, open Blood Moor).

| build | skill | casts | shards | off game.exe's -6..5 grid | offsets seen |
|---|---|---|---|---|---|
| base 7d5b3a5c | Blizzard | 3,997 | 99,925 | **14,804** (14.81 %) | -6..6 (13 values) |
| base 7d5b3a5c | Eruption | 1,003 | 13,039 | **1,872** (14.36 %) | -6..6 |
| head 681d52a9 | Blizzard | 3,997 | 99,925 | **0** | -6..5 (12 values, none missing) |
| head 681d52a9 | Eruption | 1,003 | 13,039 | **0** | -6..5 |

Unchanged between base and head (must stay equal): shards per cast (Blizzard 25 on frames 4..100, Eruption 13 on 6..78), the frame of every shard, and the world rng state after every case: **5,000 / 5,000 equal**. rng(12) and rng(13) both take one step (neither is a power of two), so the fix shifts no other roll. Expected base share 1 - (12/13)^2 = 14.79 %.

## 3. Diff review and merge cleanliness (V run / V read)

- 681d52a9 touches only `missiles.cpp` (the expression + a comment), the `fight.hpp` comment and one line of `docs/research/re/skills.md`. The only other d2d offset spread found (ai.hpp hit_spawn, hit 58, Baal Taunt) is a different game function (FUN_005acdf0) and untouched.
- 9a1a90eb adds `tests/game/test_sk_ground.cpp` (new; SKIP rc 77 without `--blizz`, builds clean on head) and `tools/emu/sk_ground_blizz.py`; no behaviour change.
- pr_guard: 0 FAIL / 0 WARN on the range and on each commit. **presubmit --quick: 1 FAIL**: `short-names: tests/game/test_sk_ground.cpp:91: 'cx' is shorter than 3 letters (readability-identifier-length)` (and `cy` on the same line). ruff: not available on the worker host (WARN, Python unchecked).
- Upstream main is still 7d5b3a5c after a fetch: the candidate fast-forwards. `git merge-tree` vs PR #160 (802dee68): clean (auto-merges fight.hpp, missiles.cpp, skills.md). vs PR #159 (c4e9db97): conflict in `components/game/ai.hpp` only, and #159 has the same conflict against main without the candidate: the candidate adds no conflict.

## 4. Side claims D2 and D3

- **D2 (seed)**: game.exe reseeds per shard from {cx + age, 666}; d2d draws from the shared world rng. Determinism only: the claim is the SET of offsets, which does not depend on the seed (V run: game.exe reaches all 144 cells; head reaches all 144). The fix does NOT need the seed change. Taking D2 would need the centre's integer subtile and would take two world-rng steps out per shard, shifting every later roll: a separate, larger PR (S).
- **D3 (blocked spot)**: game.exe makes no shard where FUN_0064cb30(mask 5 / 0x45) says blocked (V run). d2d launches it anyway, BUT d2d's `missiles_update` (ai.hpp:695) removes a standing missile on a `level->blocked(x, y, 0x04)` cell on its first update, before any hit test, with a visual-only explosion if the row has one (S, read). So the author's impact line "they can still hit a monster next to the wall" does not hold for mask-0x04 walls; what remains is the mask difference (5 / 0x45 vs 0x04), the one-frame visual, and Eruption's 0x40 bit (S). Separate card; not part of this one-line claim.
- One-line or more: the claim is faithful as one expression. D2 and D3 live in the same function but are independent deviations with their own evidence needs; keeping them out keeps the PR minimal (S).

## 5. Comparison with the author (read after the results)

- Agree: rule, r = calc1 - 1, x first, -6..5 at radius 7, blocked spot gets no shard, 25 Blizzard shards per cast, before > 0 / after 0 (author 14,926 / 14,957 of 100,000; here 14,804 of 99,925; all three within noise of 14.79 %).
- Author did not test: Eruption on the d2d side (1,872 -> 0 here); rng-stream equality base vs head; negative / int16-wrapped ages, owner / room missing, power-of-two 2r, radius <= 0 on the game side; presubmit (the short-names FAIL).
- Disagree: D3 impact (see 4). The draft calls the collision mask "mask 5"; Eruption's caller passes 0x45 (V read 0x5af957).
- U: game.exe's shard count per Eruption cast (13 or 14 depends on whether the centre's do runs at age 0 and at its last frame; d2d gives 13). Not this claim.

## Verdict: READY WITH CHANGES

1. (blocker for upstream CI) Rename `cx` / `cy` in `tests/game/test_sk_ground.cpp:91` (e.g. `centre_x` / `centre_y`) and rerun presubmit --quick to 0 FAIL; run ruff on `tools/emu/sk_ground_blizz.py` where ruff exists.
2. (low) Say "Blizzard and Eruption (do 10 / 25)" in 681d52a9's subject or body: the line changes both, and both are right per FUN_005a9820.
3. (low) Re-wrap the edited comment line in `fight.hpp` and the `skills.md` line to the surrounding width.
4. Keep D2 and D3 out of this PR; correct D3's impact line before it becomes a card.

Reproduce (worker host, upstream `tools/emu` as cwd, `D2_MPQ_DIR` with bin/game.exe): `verify_blizzard.py gen 120000 20261101`, `game`, `model`, `compare`, `grid`; d2d: `patch` into a throwaway clone's test_game.cpp, build `test_game`, `test_game --vb-blizzard d2dcases.txt` (from `gen-d2d 5000 777`), `verify_blizzard.py d2d`.

Independence note: the harness and probe were written from the disassembly and the d2d source before reading the author's docs; the author's `test_sk_ground.cpp` was read only for the diff review (section 3), after the probe was written. No author script was run.

## Shared-file lines for the merger

docs/BACKLOG.md:
- VERIFYBLIZZARD: fix/blizzard-shard-offsets READY WITH CHANGES (docs/research/pr-blizzard-verifier.md): game.exe FUN_005a9820 vs own model 0 / 120,000; d2d real builds Blizzard 14,804 -> 0 of 99,925, Eruption 1,872 -> 0 of 13,039, rng stream equal 5,000 / 5,000; to do before the PR: presubmit short-names FAIL (test_sk_ground.cpp:91 cx / cy), name Eruption in the commit, ruff; D2 / D3 separate (D3 impact line to correct)

docs/INDEX.md:
- docs/research/pr-blizzard-verifier.md - independent verifier of the Blizzard shard-offset candidate (READY WITH CHANGES)

CHANGELOG.md:
- docs: independent verifier for the d2d Blizzard shard-offset candidate (tools/emu/verify_blizzard.py, verify_blizzard_mode.inc)
