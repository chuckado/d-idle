#!/bin/bash
# palskill (the worker host): disassemble game.exe addresses from the scratch upstream checkout
cd ~/d2d-lab/pr/palskill/tools/emu
export D2_MPQ_DIR=/tmp/ug-emu
exec $HOME/d2d-lab/pr/walk2-venv/bin/python disasm.py "$@"
