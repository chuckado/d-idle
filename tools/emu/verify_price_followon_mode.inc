// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier mode for d2d's test_rules (d2lod-bot card verifyprs2): the npc.txt / quest multipliers (D9) and
// the repair gate (D10) through the REAL rules::item_price and rules::repair_cost of the build under test.
// Pasted at the top of test_rules.cpp's main() in a throwaway worktree, `int main()` turned into
// `int main(int argc, char** argv)`:
//   test_rules --verify-price CASES  -> `id value` per case
// kind 9 `which P mult done q1 q2 q3 bits`: which 0 buy, 1 sell, 2 repair; P the price before the
//   multipliers (an item of cost P, normal quality, no armor; repair: max durability 1, durability 0, so
//   the repair base is P); mult the npc.txt column; quest flags 1, 2, 3 with mults q1..q3, done when bit
//   k of `done` is set (save bit 0 or 1 by bit k of `bits`); max buy at INT_MAX; value = the price.
// kind 10 `ident eth repair throw stack kind2 nodur durcol max dur indes qty maxstack c1l c1m c2l c2m`:
//   one item and its tables row; value = repair_cost (charges of skill 1 cost mult 1024, add 100).
if (argc > 2 && std::string_view(argv[1]) == "--verify-price") {
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        long long id = 0;
        int kind = 0;
        if (!(fields >> id >> kind)) continue;
        Tables vt;
        vt.types["vfyt"] = {};
        auto& info = vt.item_info["vfy"];
        info.type = "vfyt";
        auto& base = vt.item_base["vfy"];
        Item it = item("vfy");
        d2d::d2s::Header header{};
        if (kind == 9) {
            int which = 0, mult = 0, done = 0, bits = 0;
            long long price = 0;
            std::array<int, 3> qm{};
            if (!(fields >> which >> price >> mult >> done >> qm[0] >> qm[1] >> qm[2] >> bits)) continue;
            NpcPrice prices;
            prices.buy = prices.sell = prices.rep = 1024;
            (which == 0 ? prices.buy : which == 1 ? prices.sell : prices.rep) = mult;
            prices.qflag = { 1, 2, 3 };
            for (std::size_t k = 0; k < 3; ++k) {
                (which == 0 ? prices.qbuy : which == 1 ? prices.qsell : prices.qrep)[k] = qm[k];
                if (done >> k & 1) {
                    const int bit_index = int(k + 1) * 16 + (bits >> k & 1);
                    header.quests[0][std::size_t(bit_index >> 3)] |= std::uint8_t(1 << (bit_index & 7));
                }
            }
            prices.max_buy = { std::numeric_limits<int>::max(), std::numeric_limits<int>::max(), std::numeric_limits<int>::max() };
            vt.npc_prices["vnpc"] = prices;
            base.cost = int(price);
            int value = 0;
            if (which < 2) {
                value = item_price(vt, it, "vnpc", which == 1, header);
            } else {
                vt.types["vfyt"].repair = true;
                base.durability = 10;
                it.max_durability = 1; it.durability = 0;
                value = repair_cost(vt, it, "vnpc", header);
            }
            std::printf("%lld %d\n", id, value);
        } else if (kind == 10) {
            int ident = 0, eth = 0, repair = 0, throwable = 0, stack = 0, kind2 = 0, nodur = 0, durcol = 0, max = 0, dur = 0, indes = 0, qty = 0, maxstack = 0;
            std::array<int, 4> charges{};
            if (!(fields >> ident >> eth >> repair >> throwable >> stack >> kind2 >> nodur >> durcol >> max >> dur >> indes >> qty >> maxstack
                  >> charges[0] >> charges[1] >> charges[2] >> charges[3])) continue;
            vt.types["vfyt"].repair = repair != 0;
            vt.types["vfyt"].throwable = throwable != 0;
            info.kind = kind2 ? 2 : 0;
            base.cost = 5000; base.stackable = stack != 0; base.max_stack = maxstack; base.no_durability = nodur != 0; base.durability = durcol;
            vt.skill_costs = { { 0, 0 }, { 1024, 100 } };
            vt.skill_levels = { { 0, 20 }, { 1, 20 } };
            vt.npc_prices["vnpc"] = NpcPrice{};
            it.identified = ident != 0; it.ethereal = eth != 0;
            it.max_durability = max; it.durability = dur; it.quantity = qty;
            if (indes != 0) it.props.push_back({ d2d::d2s::kIndestructible, 0, indes });
            for (std::size_t k = 0; k < 2; ++k)
                if (charges[2 * k + 1] > 0) it.props.push_back({ d2d::d2s::kChargedSkill, 1 << 6 | 5, charges[2 * k + 1] << 8 | charges[2 * k] });
            std::printf("%lld %d\n", id, repair_cost(vt, it, "vnpc", header));
        }
    }
    return 0;
}
