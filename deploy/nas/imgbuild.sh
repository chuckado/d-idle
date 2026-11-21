#!/bin/bash
# Forced command for the restricted image-build ssh key (build-hosts stage 3, docs/process/build-hosts-stage3-2026-11-14.md).
# Installed on the NAS as a root-owned file the builder container cannot write; the key line in root's authorized_keys is
#   command="/path/imgbuild",restrict,from="<head box>" ssh-ed25519 AAAA... d2-imgbuild
# `restrict` turns off pty, forwarding, agent and X11; only the verbs below run, nothing else reaches the docker socket.
#   build NAME TAG [ARG=VALUE ...]   docker build of the directory $CTX_ROOT/NAME (made by the builder container, see labrun host)
#   save TAG                         docker save TAG to stdout (the caller pipes it through gzip)
#   rmi TAG                          remove a scratch tag (tag part must start with scratch-)
#   list                             images of the allowed repositories
# TAG must be REPO:VERSION with REPO in $REPOS. Build args are limited to $ARGS and a safe value alphabet.
# One build at a time (flock), nice 19, --cpu-shares 2, 4 GB memory cap, refused when the NAS has under 3 GiB available.
# Verified (V) by the drills in the stage 3 doc; every refusal exits 2 with one line on stderr.
set -u
CTX_ROOT=${IMGBUILD_CTX_ROOT:-/mnt/cache/appdata/d2-builder/ctx}
REPOS='didle d2dbot imgscratch'
ARGS='D2D_SRC DEV APP_VERSION BUILD_JOBS UBUNTU'
LOCK=/tmp/imgbuild.lock
die() { echo "imgbuild: $*" >&2; exit 2; }

read -r -a a <<<"${SSH_ORIGINAL_COMMAND:-${IMGBUILD_CMD:-}}"
verb=${a[0]:-}
valid_tag() { # REPO:VERSION, REPO in the allow list
	[[ $1 =~ ^([a-z0-9]+):([A-Za-z0-9][A-Za-z0-9._-]{0,40})$ ]] || return 1
	[[ " $REPOS " == *" ${BASH_REMATCH[1]} "* ]]
}
case $verb in
build)
	name=${a[1]:-} tag=${a[2]:-}
	[[ $name =~ ^[a-z0-9][a-z0-9._-]{0,40}$ ]] || die "bad context name"
	valid_tag "$tag" || die "tag not allowed (repos: $REPOS)"
	dir=$CTX_ROOT/$name
	[ -d "$dir" ] && [ ! -L "$dir" ] || die "no such context $name under the fixed directory"
	[ -f "$dir/Dockerfile" ] && [ ! -L "$dir/Dockerfile" ] || die "context has no Dockerfile (or it is a link)"
	# a "# syntax=" line would make BuildKit run an arbitrary frontend image
	if grep -Eiq '^[[:space:]]*#[[:space:]]*(syntax|escape)[[:space:]]*=' "$dir/Dockerfile"; then die "Dockerfile syntax directive refused"; fi
	if grep -Eq -- '--(security|network)=(insecure|host)' "$dir/Dockerfile"; then die "insecure RUN flag refused"; fi
	args=()
	for kv in "${a[@]:3}"; do
		k=${kv%%=*} v=${kv#*=}
		[[ " $ARGS " == *" $k "* ]] || die "build arg $k not allowed"
		[[ $v =~ ^[A-Za-z0-9._/:@+-]{0,80}$ ]] || die "build arg value for $k not allowed"
		args+=(--build-arg "$k=$v")
	done
	avail=$(awk '/MemAvailable/ {print int($2/1024)}' /proc/meminfo)
	[ "$avail" -ge 3072 ] || die "NAS has only $avail MiB available, refusing"
	exec 9>"$LOCK"
	flock -w 3600 9 || die "another image build holds the lock"
	nice -n 19 ionice -c3 docker build --cpu-shares 2 --memory 4g --memory-swap 4g "${args[@]}" -t "$tag" "$dir"
	;;
save)
	valid_tag "${a[1]:-}" || die "tag not allowed (repos: $REPOS)"
	exec docker save "${a[1]}"
	;;
rmi)
	t=${a[1]:-}
	valid_tag "$t" || die "tag not allowed"
	[[ ${t#*:} == scratch-* ]] || die "rmi only removes tags that start with scratch-"
	exec docker rmi "$t"
	;;
list)
	for r in $REPOS; do docker images --format '{{.Repository}}:{{.Tag}} {{.ID}} {{.Size}} {{.CreatedSince}}' "$r"; done
	;;
*)
	die "unknown verb '${verb}' (build, save, rmi, list)"
	;;
esac
