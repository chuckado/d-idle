#!/bin/sh
# usage: dis.sh ADDR [COUNT]
cd ~/d2d-lab/pr/skmiss/tools/emu
export D2_MPQ_DIR=$HOME/d2d-lab/data D2_PATCH_INSTALLER=$HOME/Downloads/LODPatch_114d.exe
nice ~/d2d-lab/pr/walk2-venv/bin/python disasm.py --exe ~/d2d-lab/pr/tome-scratch/game.exe "$@"
