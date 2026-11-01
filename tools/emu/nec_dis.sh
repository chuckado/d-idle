#!/bin/bash
cd ~/d2d-lab/pr/palskill/tools/emu
export D2_MPQ_DIR=/tmp/ug-emu
exec nice $HOME/d2d-lab/pr/walk2-venv/bin/python disasm.py "$@"
