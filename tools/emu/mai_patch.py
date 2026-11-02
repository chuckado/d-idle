"""monsterai: turn the replay block of d2d's tests/test_monsters.cpp (Act 2's monster thinks, rules::mon_think on game.exe's recorded runs)
into a counting program over tests/mai_cases.inc / mai_names.inc (scratch tree only; the upstream test is not edited in the repo).
usage: python3 mai_patch.py <scratch d2d tree>   then build test_monsters and run it: prints one `mai AI cases differ` line per AI."""
import sys

root = sys.argv[1]
p = root + "/tests/test_monsters.cpp"
t = open(p).read()
a = t.index('#include "mon_ais_cases.inc"')
b = t.index('#include "mon_ais_act1_cases.inc"') + len('#include "mon_ais_act1_cases.inc"')
t = t[:a] + '#include "mai_cases.inc"' + t[b:]
t = t.replace('#include "mon_ais_names.inc"', '#include "mai_names.inc"')
t = t.replace("#include <array>", "#include <array>\n#include <map>", 1)
t = t.replace("                int checked = 0;\n                for (const std::string_view case_line : kCases) {",
              "                int checked = 0;\n                std::map<std::string, std::array<int, 2>> tally;\n                for (const std::string_view case_line : kCases) {", 1)
for old, new in (
    ("auto away = [&](int reach, bool) {\n                        assert(asked < asks.size() && asks[asked].first == reach);",
     "bool away_bad = false;\n                    auto away = [&](int reach, bool) {\n                        if (!(asked < asks.size() && asks[asked].first == reach)) { away_bad = true; return false; }"),
    ("for (const auto& effect : effects)                     // the toggle back",
     "if (rally) effects.push_back(d2d::rules::ThinkEffect{ d2d::rules::ThinkEffect::rally, 0, 0 });   // the leader's rally, as game.exe's effect\n                    for (const auto& effect : effects)                     // the toggle back"),
    ("assert(asked == asks.size());", "bool bad = asked != asks.size() || away_bad;"),
    ("assert(effects.size() == want_effects.size());", "if (effects.size() != want_effects.size()) bad = true;"),
    ("for (std::size_t k = 0; k < effects.size(); ++k)\n                        if (int(effects[k].kind) != want_effects[k][0]", "for (std::size_t k = 0; k < std::min(effects.size(), want_effects.size()); ++k)\n                        if (int(effects[k].kind) != want_effects[k][0]"),
    ("for (std::size_t k = 0; k < effects.size(); ++k)\n                        assert(int(effects[k].kind) == want_effects[k][0] && effects[k].id == want_effects[k][1] && effects[k].value == want_effects[k][2]);",
     "for (std::size_t k = 0; k < std::min(effects.size(), want_effects.size()); ++k)\n                        if (!(int(effects[k].kind) == want_effects[k][0] && effects[k].id == want_effects[k][1] && effects[k].value == want_effects[k][2])) bad = true;"),
    ("assert(same);\n                    ++checked;", "if (!same) bad = true;\n                    ++tally[std::string(ai_name)][bad ? 1 : 0]; ++checked;"),
    ("assert(checked > 300);\n            }\n            // The AI toggle",
     "(void)checked; int total = 0, differ = 0;\n                for (const auto& [name, count] : tally) { std::printf(\"mai %s %d %d\\n\", name.c_str(), count[0] + count[1], count[1]); total += count[0] + count[1]; differ += count[1]; }\n                std::printf(\"mai TOTAL %d %d\\n\", total, differ);\n                return differ != 0;\n            }\n            // The AI toggle"),
):
    assert old in t, old[:60]
    t = t.replace(old, new, 1)
open(p, "w").write(t)
