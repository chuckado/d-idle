#!/usr/bin/env bash
# Image builds on the NAS through the restricted key (build-hosts stage 3, docs/process/build-hosts-stage3-2026-11-14.md).
#   tools/imgbuild.sh build NAME TAG [--arg K=V]... -- CONTEXT-CMD [ARGS...]
#       1. runs CONTEXT-CMD on the builder (ssh $D2_IMGCTX_SSH) in a copy of this working tree (git ls-files), with the target
#          directory appended as the LAST argument (e.g. `-- deploy/d2d-container/context.sh`);
#       2. asks the NAS to build that directory as TAG (forced-command verb `build`, only allowed repos and build args);
#       3. removes the context from the builder. Streams the build log.
#   tools/imgbuild.sh save TAG OUT.tar.gz     docker save on the NAS, gzip -n here
#   tools/imgbuild.sh rmi TAG                 remove a scratch-* tag; tools/imgbuild.sh list    the allowed images
# Env: D2_IMGBUILD_SSH (ssh alias of the restricted key, default nas-img), D2_IMGCTX_SSH (ssh alias of the builder that
# assembles contexts and gets cleaned, default nas-build). Run it from the repo root.
set -uo pipefail
nas=${D2_IMGBUILD_SSH:-nas-img} ctxssh=${D2_IMGCTX_SSH:-nas-build}
ctxroot=/data/ctx # the builder's view of the NAS directory the forced command builds from
die() { echo "imgbuild: $*" >&2; exit 2; }
verb=${1:-}; [ $# -gt 0 ] && shift
case $verb in
build)
	name=${1:?usage: imgbuild.sh build NAME TAG [--arg K=V]... -- CONTEXT-CMD...}; tag=${2:?missing TAG}; shift 2
	args=()
	while [ $# -gt 0 ] && [ "$1" != -- ]; do
		[ "$1" = --arg ] || die "unexpected $1"
		args+=("${2:?--arg needs K=V}"); shift 2
	done
	[ "${1:-}" = -- ] || die "missing -- CONTEXT-CMD"; shift
	[ $# -gt 0 ] || die "missing the context command"
	cleanup() { ssh "$ctxssh" "rm -rf $ctxroot/$name" >/dev/null 2>&1 || true; }
	trap cleanup EXIT
	# the working tree (tracked and untracked files that are not ignored) goes to the builder over ssh, NOT through labrun: labrun
	# jobs queue behind the one-at-a-time gate lock for up to an hour, and assembling a context is light
	src=/data/ctxsrc/$name
	cmd=$(printf '%q ' "$@")
	git ls-files -z -co --exclude-standard | tar --null -T - -cf - | ssh "$ctxssh" \
		"rm -rf $src && mkdir -p $src && tar -x -m -C $src && cd $src && export PATH=/opt/go/bin:\$PATH GOCACHE=/data/gocache GOMODCACHE=/data/gomod GOTOOLCHAIN=local GOFLAGS=-buildvcs=false && nice -n 10 $cmd $ctxroot/$name; rc=\$?; cd /; rm -rf $src; exit \$rc" || die "context step failed"
	ssh "$nas" "build $name $tag ${args[*]:-}" ;;
save)
	tag=${1:?usage: imgbuild.sh save TAG OUT.tar.gz}; out=${2:?missing OUT}
	ssh "$nas" "save $tag" | gzip -n >"$out.part" && [ "${PIPESTATUS[0]}" = 0 ] && mv "$out.part" "$out" || { rm -f "$out.part"; die "save failed"; } ;;
rmi) ssh "$nas" "rmi ${1:?usage: imgbuild.sh rmi TAG}" ;;
list) ssh "$nas" list ;;
*) sed -n 2,13p "$0"; exit 2 ;;
esac
