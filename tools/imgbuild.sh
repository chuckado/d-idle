#!/usr/bin/env bash
# Image builds on the NAS through the restricted key (build-hosts stage 3, docs/process/build-hosts-stage3-2026-11-14.md).
#   tools/imgbuild.sh build NAME TAG [--arg K=V]... -- CONTEXT-CMD [ARGS...]
#       1. runs CONTEXT-CMD on the builder (labrun host $D2_IMGCTX_HOST, default tower) from the repo mirror, with the target
#          directory appended as the LAST argument (e.g. `-- deploy/d2d-container/context.sh`);
#       2. asks the NAS to build that directory as TAG (forced-command verb `build`, only allowed repos and build args);
#       3. removes the context from the builder. Streams the build log.
#   tools/imgbuild.sh save TAG OUT.tar.gz     docker save on the NAS, gzip -n here
#   tools/imgbuild.sh rmi TAG                 remove a scratch-* tag; tools/imgbuild.sh list    the allowed images
# Env: D2_IMGBUILD_SSH (ssh alias of the restricted key, default tower-img), D2_IMGCTX_HOST (labrun host that assembles contexts,
# default tower), D2_IMGCTX_SSH (ssh alias of the builder for cleanup, default tower-build).
set -uo pipefail
nas=${D2_IMGBUILD_SSH:-tower-img} ctxhost=${D2_IMGCTX_HOST:-tower} ctxssh=${D2_IMGCTX_SSH:-tower-build}
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
	labrun "$ctxhost" -- "$@" "$ctxroot/$name" || die "context step failed"
	ssh "$nas" "build $name $tag ${args[*]:-}" ;;
save)
	tag=${1:?usage: imgbuild.sh save TAG OUT.tar.gz}; out=${2:?missing OUT}
	ssh "$nas" "save $tag" | gzip -n >"$out.part" && [ "${PIPESTATUS[0]}" = 0 ] && mv "$out.part" "$out" || { rm -f "$out.part"; die "save failed"; } ;;
rmi) ssh "$nas" "rmi ${1:?usage: imgbuild.sh rmi TAG}" ;;
list) ssh "$nas" list ;;
*) sed -n 2,13p "$0"; exit 2 ;;
esac
