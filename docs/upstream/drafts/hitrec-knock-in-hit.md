# Draft (local, NOT posted): "A knocking hit does not start a knockback in a get-hit" (P03c)

> status: 2026-11-15 (card decomp27, supersedes the decomp26 candidate note). Fork branch `fid/knock-in-gethit-r1` (private fork remote), two commits on upstream main `77990456`: oracle `64047f23`, claim `4e911f21`. Ready for the independent verifier (brief: [briefs/verify-knock-gethit.md](../briefs/verify-knock-gethit.md)); not posted. Body: [bodies/hitrec-knock-in-hit-body.md](../bodies/hitrec-knock-in-hit-body.md). Follows #548 (a big hit does not restart a get-hit).

PR title (when unblocked): `A knocking hit does not start a knockback in a get-hit (FUN_0057edd0)`

## What changes

A hit with the knockback flag (result flag 8: Smite, a monster's knock) goes through `FUN_0057cee0` to `FUN_00580a70` with new mode 0x13 (`0x57d350`), which asks `FUN_0057edd0` first (`0x580adc`); game.exe refuses the mode from GH, BL, DT and DD and the hit stops there (the get-hit is not tried). d2d refused BL and the swing modes already, but a player standing in a get-hit was slid and her hold moved. `Fight::hurt_side` (`tick.cpp:674`) now skips the knock in a get-hit that is not itself a slide (`pmode == GH` with `knock` unset).

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe refuses new mode 0x13 from DT, GH, BL, DD (hold queued) and allows every other mode, KB included | V run: `FUN_0057edd0` under the emulator, class 3, modes 0..19 | `mode_switch.py game` |
| d2d before | V run (oracle commit): 13 cells compared, 1 differ (GH), 7 excluded (DT, DD dead; A1, KK, S1, S3, S4 swings, P04) | `mode_switch.py compare-knock` (exit 1) |
| d2d after | V run (claim): 13 cells, 0 differ; #548's `compare` 13 / 0; `selftest` 12 of 12; `test_hit_recovery` (a new assert: a knock in GH moves neither hold nor position), `test_game`, `test_combat`, `test_rules`, `test_swing_start` rc 0 | same |

## Limits

- A knocking hit in a swing: game.exe allows 0x13 from A1, KK, S1, S3, S4 (hold queued); d2d never interrupts a swing for a knock. Not changed; it belongs with P04 (what a hit does inside a swing).
- Only the mode switch is claimed, not the knock's distance or timing.
- Class-dependent cells: none for 0x13 beyond the two known class cells the 4 and 9 columns already have.
