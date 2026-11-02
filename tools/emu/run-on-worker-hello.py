# Smoke script for tools/emu/run-on-worker.sh: proves the emulator venv, unicorn and the MPQ dir are reachable on the worker host.
import os, sys
import unicorn
print("hello from the worker host; unicorn", unicorn.__version__, "mpq", os.environ.get("D2_MPQ_DIR"), "scratch", os.environ.get("D2_SCRATCH"), "args", sys.argv[1:])
print("mpq files:", sorted(os.listdir(os.environ["D2_MPQ_DIR"]))[:4])
