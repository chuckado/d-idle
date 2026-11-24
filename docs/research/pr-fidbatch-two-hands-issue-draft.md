# Issue draft (local, NOT posted): `test_two_hands_blow` aborts on upstream main

> status: draft 2026-11-24 (agent fidbatch). Found while running the whole ctest for the pet-order PR; needs the user's go before anything is filed. Not a tiny fix as far as I could see without a debugger, so it is reported, not fixed in that PR.

Title: `test_two_hands_blow aborts at test_two_hands_blow.cpp:58 on main (dadda9b98)`

Body:

> `ctest` on upstream main `dadda9b98` (GCC 14, Release, `D2D_WERROR=OFF`, the lab image's game data; the test is synthetic and needs none) fails one test of 506: `test_two_hands_blow` aborts with `Assertion 'fight.blow_fighter(&attack).phys_hi == 100' failed` at `tests/game/test_two_hands_blow.cpp:58`. The first assertion of that block (one weapon: the player's, 300) passes; the failing one comes right after `actor.two_hands = true` with `attack.weapsel = 0` and `blow_hand` left at -1, where `Fight::blow_fighter` (`components/game/fight/strike.cpp`) should return `hand_combat[0]` (100).
> Not checked: whether CI's build (RelWithDebInfo, `D2D_WERROR=ON`) passes it; whether a recent commit touching `Fight`/`Hero` (the "gamedata ... apart from GameData" regroup, the Double Throw event commit) is the cause. Re-check: `cmake --build build --target test_two_hands_blow && build/tests/test_two_hands_blow`.
>
> Agent: Claude Code
