#!/bin/sh
# SORCREST: disassemble game.exe by address (the worker host). usage: sorcrest_dis.sh ADDR [COUNT]
# Uses upstream's disasm.py from the scratch checkout; the MPQ stack for game.exe is the emulator's.
cd ~/d2d-lab/pr/skmiss/tools/emu || exit 1
export D2_MPQ_DIR=${D2_MPQ_DIR:-/tmp/ug-emu} D2_PATCH_INSTALLER=${D2_PATCH_INSTALLER:-$HOME/Downloads/LODPatch_114d.exe}
exec nice ${PY:-$HOME/d2d-lab/pr/walk2-venv/bin/python} disasm.py "$@"
