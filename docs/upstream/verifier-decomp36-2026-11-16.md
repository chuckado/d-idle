# Verifier: decomp36, the gamble price and MonUMod Multiple Shots (2026-11-16)

> status: current, 2026-11-16 (card verifydecomp36, TASK-569, Opus, singly; independent verifier before posting). Nothing posted. Subjects: private fork branches `fid/gamble-price-r1` (head `eba73ede`) and `fid/umod-can-take-r1` (head `6d1fc304`), both on upstream `b6e6940d`; bodies `docs/upstream/bodies/gamble-price-body.md` and `umod-multishot-body.md`; trace `docs/research/dc36-lane-b.md` (TASK-558). The third decomp36 branch (`fid/death-corpse-exp-r1`) is held, as decided, and was not verified. Checklist from `verifier-decomp35-2026-11-16.md`, `verifier-p13-2026-11-16.md`, `verifier-seq-followup-2026-11-16.md`. The fixed r2 branches are on the private fork remote only.

Tags: **V run** (run by this verifier today, NAS builder), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdicts

| PR | Verdict | Post |
|---|---|---|
| Gamble price | **r1 FAIL (hygiene), r2 PASS WITH CHANGES** | `fid/gamble-price-r2`: oracle `f9a127e1`, claim `48d8f6ee`, on upstream main `def80df7`; body after G1 to G6 |
| MonUMod Multiple Shots | **r1 FAIL (hygiene), r2 PASS WITH CHANGES** | `fid/umod-can-take-r2`: oracle `810b7508`, claim `a4988666`, on upstream main `def80df7`; body after U1 to U6 |

The two PRs are independent (they share only neighbouring `tools/emu/README.md` rows); post in either order, gamble first by value. Both rules are game.exe's and both claims implement them (V read and V run below). Neither r1 can be posted:

1. **Identity and trailer (blocker, both).** All four r1 commits are authored and committed as the scratch `Claude <noreply@anthropic.com>` and carry no trailer (V run, `git log --format`). r2: the user's noreply identity as author and committer (as the merged `fid/tele-refund-r2`), `Agent: Claude Code` the only parsed trailer, after a blank line (`git interpret-trailers --parse`, all four).
2. **The compare is `diff` (high, both).** r1's README rows and bodies compare with `diff game.tsv d2d.tsv`: two empty files pass silently (exit 0), and nothing reports a count. r2 adds a `compare` verb to each oracle script: `N cases, M differ`, exit 1 on any difference, an empty or unequal file, a malformed line or ids out of step (V run on hand-made files, six cases each, both empty included).
3. **Usage errors (medium, both).** r1 `gamble_price.py` with an unknown verb boots the emulator and exits 0, `game` without a file is an `IndexError`, and its docstring holds `\.` in a non-raw string (a `SyntaxWarning` on Python 3.12+, an error under `-W error`, V run). r1 `umod_can_take.py` runs the whole grid for any argument but `probe`, and fails presubmit's ruff (F841, unused `rows`): **presubmit 1 FAIL on r1** (V run). r2: `usage:` and exit 1 on no, unknown or short arguments; ruff clean.
4. **The tests do not pin game.exe's values (high, both).** r1's gamble test asserts only `9ha == hax` (an exceptional base): a mutant that maps only exceptional bases to their normal code (elites still wrong) passes it. r1's umod test asserts only that vampire1 refuses the fPick 2 rows: a mutant that refuses every row for a nomultishot monster passes it (section Mutants). r2 pins game.exe's numbers from the oracle: `hax`, `9ha`, `7ha` at clvl 1, 30, 45 and 45 with 10 percent reduced; vampire1 and slinger5 take 40 of 43 rows, sk_archer1 41 and the Multiple Shots row.
5. **No `docs/research/re/` finding (medium, both; the template's checklist item).** r2's claims edit `store.md` (Gambling: the normal code, the version branch) and `monsters.md` (FUN_005a03e0: isMelee or nomultishot, the 11 rows).
6. **Base (low).** r1 is on `b6e6940d`; upstream main is now `def80df7` (33 commits on). Both rebase with no conflict.

## Game side, re-derived (V read by this verifier, then V run)

### Gamble price

- **Callers** (a scan of `.text` for `call`/`jmp` rel32 and of every section for absolute pointers, V run): FUN_00629370 has **one** caller, `0x62f07a` in FUN_0062efb0 mode 2. FUN_006287d0 has three: `0x62937a` (FUN_00629370), `0x62f09e` (mode 2's version-0 branch) and `0x630282` (a bit writer: the normcode as 32 bits after two 16-bit fields, an item save or packet, not a price). So the normcode rule reaches the gamble price only; buy, sell and repair (the other modes of FUN_0062efb0) never call FUN_00629370.
- **FUN_0062efb0 mode 2** (`0x62f040..0x62f0bb`): reduced = stat 0x57 capped at 99 (`cmp esi, 0x63`); FUN_0062a670 reads the item data word `+0x30`; at least 1: `FUN_00629370(item, clvl)` less `FUN_00483360(price, reduced, 100)` when reduced is not 0; 0: the normal record's `+0xd4` (`FUN_006287d0` -> FUN_00633680 -> FUN_006335f0, `0x62f09d..0x62f0af`). **The word is the item's version** (S, `docs/research/re/drops.md`: `game+0x78`, 101 in an expansion game, 0 in a classic one), so decomp36's "not understood" is answerable: in a **classic** game every gamble base costs its normal record's `+0xd4` (V run, `probe`: `lsd` 42,900 with the word 0, 8,496 with 1). d2d models only the expansion branch; a Limit, not a claim (G3).
- **FUN_00629370** (`0x629370..0x6293fc`): `FUN_006287d0(item)` = the record's `+0x84` (normcode) when non-zero, else `+0x80` (`0x62880e..0x62881b`), then the record of that code; rings and amulets (`rin `, `amu ` against the normal record's `+0x80`) cost `+0xd4`; everything else reads the **normal record's** `+0x88` / `+0x8c` (uber and ultra codes), `+0xe4` / `+0xe8` (stack) and cost. The claim does the same: after the lookup every field comes from `base`, the normal record (V read, `rules.hpp`).
- **d2d callers** of `rules::gamble_price` (grep of the whole tree, V read): `store_gamble` (the payment) and `apps/d2d/store.cpp:146` (the hover price); both take the fix. `tests/test_rules.cpp:746..749` uses synthetic tables with no normcode, so its values do not move. No other price function changed.
- **V run**: `gamble_price.py gen` = 73,062 cases, `game` = 73,062 lines, byte-identical (md5) to decomp36's `gam_game.tsv`. `probe`: no dependence on quality (1, 2, 4, 6, 7), identified, ethereal, NPC (Gheed, Akara, Charsi) or difficulty (0, 2). Extra grid (not in decomp36): the 369 codes x clvl 1, 7, 30, 45, 63, 99 x reduced 1, 3, 7, 13, 25, 33, 50, 77, 99 = 19,926 cases, because only 0 and 10 were compared and game.exe subtracts a `MulDiv` while d2d truncates.

### MonUMod eligibility

- **FUN_005a03e0** (`0x5a03e0..0x5a04ce`, V read): a disabled row (`+6` 0) refused; in a classic game (second argument 0) a row whose word `+4` is 100 or more refused; the two exclusions (`+0xa`, `+0xc`, FUN_005a0070); fPick (`+9`) 1 and 3 through FUN_0046c140; fPick 2 refused when the MonStats row's byte `+0xc` or byte `+0xe` has `[0x6ce26c]` set (`0x5a046e..0x5a04c7`). `[0x6ce26c]` = 2 (V run, read from the image).
- **The column semantic, read rather than inferred** (V run, `mon.py`): for all 734 MonStats rows, bit 1 of the dword at `+0xc` equals MonStats.txt's `isMelee` and bit 17 (`+0xe & 2`) equals `nomultishot`; 0 mismatches. The 11 rows with bit 17 are vampire1..5 (131..135), slinger5, slinger6 (394, 395), vampire6..8 (695..697) and vampire9 (731), as decomp36 said.
- **Callers**: FUN_005a03e0 is called only from FUN_005a0500 (`0x5a057a`, champion pick) and FUN_005a0600 (`0x5a067c`, unique pick); FUN_005a0600 from FUN_005a0760 (`0x5a08f3`, a random unique) and FUN_005a49b0 (`0x5a4b3f`, a superunique's extra mods) (V run, the scan). d2d: `can_take` only through `unique_detail::pick`, used by `roll_boss` (champion and unique) and `superunique_mods` (V read). Mods that game.exe gives without FUN_005a03e0 (an act boss's, a family's, Skills sumumod, a minion's copy; `monsters.md` row 24) do not use `can_take` in d2d either.
- **Effect** (V run, MonUMod.txt, SuperUniques.txt and Levels.txt through the emulator): row 29 (multishot) has `champion` empty and upick 6 in every difficulty, so only **random uniques** are affected; no superunique is of the 11 classes. They can be uniques (Levels `umon`) in **Act 1** Catacombs level 3 (vampire5); **Act 2** Dry Hills and Lost City (slinger5, slinger6), the seven Tal Rasha's Tombs and the Arcane Sanctuary (vampire1); Act 3 Travincal, the Temples and the Durance (vampire2..4); and in Act 5 they spawn in Nightmare and Hell only (`nmon`: Worldstone Keep level 3, Throne of Destruction, Baal's Temple 1, vampire6..8). **decomp36's "Acts 3 to 5 uniques only, none in Acts 1 and 2" (dc36-lane-b.md line 10) is wrong**: Acts 1 and 2 are in the bot's horizon. Such a unique also draws its other mods from a smaller total weight, so the draw itself changes (I, RNG-visible).
- **V run**: `umod_can_take.py game` = 31,562 lines, byte-identical (md5) to decomp36's `umod_game.tsv`. The classic branch is not run (the oracle passes 1).

## Differentials (V run)

@@DIFF@@

## Mutants (V run)

@@MUT@@

## Build, ctest, tidy

@@BUILD@@

## RNG luck

Both grids are exhaustive and deterministic: every code a gamble tab can hold x clvl 1..99 x reduced 0 and 10, and every MonStats row x every MonUMod row; no draw is involved on either side, and neither new test has an `Rng`, so the decomp35-style 200-seed reseed does not apply. The extra reduced-prices grid covers what the exhaustive grid fixes at 0 and 10. The umod change moves the unique-mod draw of the 11 rows; the existing boss-roll tests (`test_thief_mod` and the others) ran in the full ctest below.

## Required changes to the gamble body (`gamble-price-body.md`)

- **G1 (high).** Commits and tree: "upstream `def80df7` plus `f9a127e1` (oracle) and `48d8f6ee` (claim)". "Two commits" paragraph: the claim also adds the finding in `docs/research/re/store.md` and the asserts in `test_gamble_price` (game.exe's 771, 6,896, 12,933 and 11,640 for `hax`, `9ha`, `7ha`); the oracle commit's test has no failing assert.
- **G2 (high).** Re-check block: replace `diff game.tsv d2d.tsv` by `uv run python gamble_price.py compare game.tsv d2d.tsv` (`73062 cases, 0 differ`, exit 0; on the oracle commit `48310 differ`, exit 1); say the d2d side needs `D2_MPQ_DIR` (as `test_game`) and `gamble_price.py game` the emulator's game image and a built `mpq-cat`, as the neighbouring rows.
- **G3 (medium, claims limited to proven).** Limits: replace "The data word `+0x30` ... is not understood" by "Mode 2 takes FUN_00629370 when the item's version (item data `+0x30`, FUN_0062a670; 101 in an expansion game, 0 in a classic one) is at least 1; with 0 the price is the normal record's `+0xd4` (`0x62f09d..0x62f0af`). d2d models the expansion branch only; the classic one is not compared." Keep the probe sentence.
- **G4 (medium).** Add the extra grid row (reduced 1..99, 19,926 cases; section Differentials) and "FUN_00629370 has one caller (`0x62f07a`, mode 2), so buy, sell and repair prices are untouched; d2d's two callers are the payment (`store_gamble`) and the hover price (`apps/d2d/store.cpp`)".
- **G5 (medium, template).** Claims table with the template's four columns (`Commit | Claim or change | Evidence | Re-check`); the template's three checklist items, ticked (test `tests/game/test_gamble_price.cpp`; finding `docs/research/re/store.md`; `pr_guard` 0 FAIL 0 WARN), with presubmit's WARNs answered: `data-gating` "the test reads the 1.14d tables through `D2_MPQ_DIR` and skips without them, as `test_load_gold`"; `build` and `gcc` "built and ctest run (Release, GCC)". End with `---` and `Agent: Claude Code`. No Screenshots section.
- **G6 (low).** The effect for a player in one line: "at clvl 45 an elite Hand Axe (`7ha`, a Tomahawk) cost 51,186 in d2d and costs 12,933 in game.exe; an exceptional one (`9ha`, a Hatchet) cost less than in game.exe below clvl 45 (clvl 30: 5,493 against 6,896)". r1's test comment called `9ha` a Tomahawk.

## Required changes to the MonUMod body (`umod-multishot-body.md`)

- **U1 (high).** Commits and tree: "upstream `def80df7` plus `810b7508` (oracle) and `a4988666` (claim)"; the claim adds the `monsters.md` finding and pins game.exe's counts in `test_umod_can_take` (vampire1 and slinger5 take 40 of 43 rows, sk_archer1 41; the Multiple Shots row refused for the first two, taken by the third).
- **U2 (high, claims limited to proven).** Limits: drop "inferred from the equal count": the bit is read (bit 17 of MonStats `+0xc` equals the `nomultishot` column in all 734 rows, bit 1 equals `isMelee`; mask `[0x6ce26c]` = 2). State the effect as run: only random uniques (row 29 is no champion row and no superunique is of these classes), in Act 1 (Catacombs 3), Act 2 (Dry Hills, Lost City, the Tombs, the Arcane Sanctuary), Act 3, and Act 5 in Nightmare and Hell; the other mods of such a unique are drawn from a smaller total weight.
- **U3 (high).** Re-check block: `uv run python umod_can_take.py compare game.tsv d2d.tsv` instead of `diff` (`31562 pairs, 0 differ`; oracle commit 11 differ, exit 1); `D2_MPQ_DIR` for the d2d side.
- **U4 (medium).** Answer presubmit's WARNs: `siblings` "no skill or hit path is touched: the change is the mod pick's eligibility, used by the champion, unique and superunique picks (FUN_005a0500; FUN_005a0600 from FUN_005a0760 and FUN_005a49b0; `pick` in d2d)"; `data-gating` as G5; `doc-claims` "every MonStats row x every MonUMod row: the grid is exhaustive (734 x 43)".
- **U5 (medium, template).** As G5: four-column Claims table, the template's three checklist items ticked (test, `monsters.md`, `pr_guard`), `---` and `Agent: Claude Code`, no Screenshots section.
- **U6 (low).** Limits: "The classic-game refusal (row id 100 or more when the second argument is 0) is not run; the oracle passes an expansion game."

## For the coordinator (not body changes)

- `docs/research/dc36-lane-b.md` line 10 and section 4 ("Effect: Acts 3 to 5 uniques only (U), none in Acts 1 and 2"; "U kept: the Acts where the 11 nomultishot rows spawn"): replace by the list above (V run). The pipeline row P21 should rank by that: vampire uniques in the Tombs and the Arcane Sanctuary and slinger uniques in Act 2 are on the bot's route.
- `dc36-lane-b.md` section 3: "U what `+0x30` is" is the item version (drops.md).

## Hygiene (V run, r2)

- Identity: all four r2 commits have the user's noreply identity as author and committer (`git log --format`), `Agent: Claude Code` the only parsed trailer.
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN on both heads. `presubmit.py --quick --base upstream/main`: gamble 0 FAIL, 3 WARN (`data-gating`, `build`, `gcc`); MonUMod 0 FAIL, 6 WARN (`siblings`, `data-gating`, two `doc-claims`, `build`, `gcc`); answered by G5 and U4. r1 against `b6e6940d`: gamble 0 FAIL, 3 WARN; MonUMod **1 FAIL** (ruff F841).
- Diff: gamble oracle 3 files (+166), claim 3 files (`rules.hpp` +6 -2, `store.md`, the test asserts); MonUMod oracle 3 files, claim 5 files (`gamedata_load.cpp`, `montypes.hpp`, `uniques.hpp`, `monsters.md`, the test asserts). Each oracle commit leaves `components/` untouched and its test has no failing assert (the asserts come with the claim), so each commit is bisect-safe. No private-only code, no host names or paths (read of the diffs).

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill (decomp36's recipes 2 and 3 not run); nothing ran on the worker host.
- The classic branches (gamble version 0; MonUMod rows of 100 and up in a classic game); the gamble stock roll (`FUN_00578790`, F1); how often a vampire or slinger unique rolled Multiple Shots in d2d before the fix.
- `fid/death-corpse-exp-r1` (held).

## Re-check commands (as run)

Builder container, the emulator image of `docs/emu-image.md` (`D2_BUILD` pointing at a build with `tools/mpq-cat`); lab scripts `verifyd36/{bld,chain,stage,emu,emuruns,runs}.sh`, `py/{calls,mon,su}.py` (lab only).

```sh
python3 gamble_price.py gen > gc.tsv; python3 gamble_price.py game gc.tsv > gg.tsv
test_gamble_price --gamble-price gc.tsv | grep -E '^[a-z0-9]+\.[0-9]+\.[0-9]+ -?[0-9]+$' > gd.tsv
python3 gamble_price.py compare gg.tsv gd.tsv; echo $?        # oracle commit: 48310 differ, 1; claim: 0, 0
python3 umod_can_take.py game > ug.tsv
test_umod_can_take --umod-can-take | grep -E '^[0-9]+ [0-9]+ [01]$' > ud.tsv
python3 umod_can_take.py compare ug.tsv ud.tsv; echo $?      # oracle commit: 11 differ, 1; claim: 0, 0
```
