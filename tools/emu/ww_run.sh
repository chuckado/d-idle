#!/bin/bash
# ww_run.sh script.py args : run a python script with the emulator venv (decomp23). On the worker host.
source ~/d2d-lab/emu/env.sh
export GAMEEXE EMU D2_MPQ_DIR
cd "$(dirname "$0")"
PYTHONPATH="$EMU${WW_PYPATH:+:$WW_PYPATH}" exec nice -n 10 "$PY" "$@"
