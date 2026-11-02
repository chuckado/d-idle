#!/bin/bash
# playerdef: disassemble game.exe on the worker host (upstream disasm.py from the palskill read checkout).
cd ~/d2d-lab/pr/palskill/tools/emu || exit 2
export D2_MPQ_DIR=/tmp/ug-emu
exec nice $HOME/d2d-lab/pr/walk2-venv/bin/python disasm.py "$@"
