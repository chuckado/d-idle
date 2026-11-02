# SPDX-License-Identifier: GPL-3.0-or-later
"""treasure card: print table headers and a few rows as game.exe's MPQ stack gives them (worker host)."""
import sys

import drops


def main():
    o = drops.Oracle()
    e = o.e
    for name in sys.argv[1:]:
        rows = drops.txt_rows(e, name)
        print(name, len(rows), "rows")
        print(list(rows[0].keys()))
        for r in rows[:3]:
            print({k: v for k, v in r.items() if v})


if __name__ == "__main__":
    main()
