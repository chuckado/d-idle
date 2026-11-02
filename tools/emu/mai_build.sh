#!/bin/bash
# monsterai: build d2d's test_monsters in a scratch worktree (worker host), 2 jobs, nice.
# usage: mai_build.sh [src dir]   (default ~/d2d-lab/pr/monsterai-src)
SRC=${1:-$HOME/d2d-lab/pr/monsterai-src}
D=$HOME/d2d-lab/src/build/_deps
B=$HOME/d2d-lab/pr/monsterai-out/build
mkdir -p "$B"
cd "$SRC" || exit 1
[ -f "$B/build.ninja" ] || cmake -S . -B "$B" -G Ninja -DCMAKE_BUILD_TYPE=Release -DFETCHCONTENT_SOURCE_DIR_CLI11=$D/cli11-src -DFETCHCONTENT_SOURCE_DIR_SENTRY_NATIVE=$D/sentry_native-src -DFETCHCONTENT_SOURCE_DIR_STORMLIB=$D/stormlib-src "-DCMAKE_CXX_FLAGS=-Wno-error=maybe-uninitialized -Wno-error=array-bounds" > "$B/../cmake.log" 2>&1 || { tail -20 "$B/../cmake.log"; exit 1; }
nice ninja -C "$B" -j2 test_monsters > "$B/../build.log" 2>&1; rc=$?; tail -3 "$B/../build.log"; exit $rc
