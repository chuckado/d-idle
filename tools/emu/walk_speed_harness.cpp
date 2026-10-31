// d2d's walk / run step a tick, before and after the patch (d2lod-bot, card
// decompwalk), line for line from upstream main d0ea14b7, on the cases
// tools/emu/walk_speed.py writes. Prints `id before after`, the step a tick in
// game.exe's path-velocity units (1/4096 subtile), so it compares with the
// velocity FUN_00623f50 stores at path +0x7c.
//
//   g++ -O2 -std=c++20 -o walk_speed_harness walk_speed_harness.cpp
//   ./walk_speed_harness < cases.tsv > d2d.tsv
//
// before:
// - monster (components/game/ai.cpp:2724..2727, 643): chill = (100 +
//   coldeffect) / 100 while chilled; walk = cells_per_sec(Velocity) x elapsed
//   x chill x max(100 + speed_pct + boss_speed + pace, 10) / 100; run = the
//   same with MonStats Run; RN moves by run, WL by walk.
// - player (components/game/world/tick.cpp:102..109, 356..361;
//   fight/strike.cpp:185..186): frw = stat 96 + velocitypercent deltas;
//   walk % = max(base + effective_speed(frw, 150) + chill, 25) with base 100
//   or Run x 100 / Walk, chill -50, effective_speed(v) = v > 0 ? 150 v / (150
//   + v) : v (rules/combat.hpp:170); vel = Walk x % / 100.
// - cells_per_sec(v) = v / 16 x 25 / 5 (gamedata.hpp:810), elapsed 0.04 s
//   (a 40 ms tick); a cell is 5 subtiles.
// after (the candidate): velocity = Velocity (any moving mode) x max(100 +
// deltas + chill x coldeffect, 25) / 100 for a monster; for a player
// max(base + E(frw) + vp + chill, 25) with E(v) = v != 0 ? 150 v / (150 + v)
// (C division) : 0, velocitypercent deltas not diminished.
#include <algorithm>
#include <cstdio>

static float cells_per_sec(float v) { return v / 16.f * 25.f / 5.f; }
static int effective_speed(int speed, int cap) { return speed > 0 ? cap * speed / (cap + speed) : speed; }
static int effective_any(int speed, int cap) { return speed != 0 ? cap * speed / (cap + speed) : 0; }
static double units(float cells_per_tick) { return double(cells_per_tick) * 5.0 * 4096.0; }

int main() {
    long id;
    int kind, cls, mode, vel, run, d, frw, chill, cold, walkv, runv;
    char line[512];
    const float elapsed = 0.04f;
    while (std::fgets(line, sizeof line, stdin)) {
        if (line[0] == '#' || std::sscanf(line, "%ld %d %d %d %d %d %d %d %d %d %d %d", &id, &kind, &cls, &mode, &vel, &run, &d, &frw, &chill, &cold, &walkv, &runv) != 12) continue;
        double before, after;
        if (kind == 1) {
            const float c = chill ? float(100 + cold) / 100.f : 1.f;
            const float walk = cells_per_sec(float(vel)) * elapsed * c * float(std::max(100 + d, 10)) / 100;
            const float runs = cells_per_sec(float(run)) * elapsed * c * float(std::max(100 + d, 10)) / 100;
            before = units(mode == 15 ? runs : walk);
            after = units(cells_per_sec(float(vel)) * elapsed * float(std::max(100 + d + (chill ? cold : 0), 25)) / 100);
        } else {
            const int w = std::max(walkv, 1);
            const int base = mode == 3 ? runv * 100 / w : 100;
            const int ch = chill ? -50 : 0;
            const int pct_before = std::max(base + effective_speed(frw + d, 150) + ch, 25);
            const int pct_after = std::max(base + effective_any(frw, 150) + d + ch, 25);
            before = units(cells_per_sec(float(w * pct_before) / 100.f) * elapsed);
            after = units(cells_per_sec(float(w * pct_after) / 100.f) * elapsed);
        }
        std::printf("%ld %.3f %.3f\n", id, before, after);
    }
}
