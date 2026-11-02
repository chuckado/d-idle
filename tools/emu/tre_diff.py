# SPDX-License-Identifier: GPL-3.0-or-later
"""The upstream drop differential (tools/emu/drops.py of the read checkout) with the d2d side pointed at a given
drop-dump binary ($TRE_DROPDUMP): game.exe's FUN_0055a6d0 and friends natively against d2d's rules, line by line.

    bash tools/emu/tre_run.sh tre_diff.py 100001-200000        # rolls
    bash tools/emu/tre_run.sh tre_diff.py tables                # every class's entries
    bash tools/emu/tre_run.sh tre_diff.py items 1-40000
    bash tools/emu/tre_run.sh tre_diff.py objects 1-20000
    bash tools/emu/tre_run.sh tre_diff.py stands 1-20000

Nothing is changed in the oracle; only run_ours is replaced (the upstream one looks in <checkout>/build).
"""
import os
import subprocess

import drops
import emu


def run_ours(args, stdin):
    env = dict(os.environ)
    out = subprocess.run([os.environ["TRE_DROPDUMP"], emu.data_path()] + args, input=stdin, env=env,
                         capture_output=True, text=True, check=True).stdout.splitlines()
    return out[-stdin.count("\n"):]


drops.run_ours = run_ours

if __name__ == "__main__":
    drops.main()
