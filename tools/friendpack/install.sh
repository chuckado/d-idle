#!/usr/bin/env bash
# d.idle friends pack installer. Never uses sudo, touches nothing outside this folder, the Docker image and one
# named volume. Usage: ./install.sh [--game-dir PATH] [--port N] [--host NAME] [--project NAME] [--yes]
#   --port N     watch page port (answers use N+1); default 8891
#   --host NAME  also publish on all interfaces and accept the page at http://NAME:PORT (LAN use); default loopback only
#   --project N  compose project name, default didle (the volume is <project>_d2dbot-data)
#   --yes        skip the question only when this version's licence was already accepted here
set -euo pipefail
cd "$(dirname "$(readlink -f "$0")")"

game= port=8891 host= project=didle yes=0 minfree_kb=3145728
while [ $# -gt 0 ]; do
	case "$1" in
	--game-dir) game=${2:?--game-dir needs a path}; shift 2 ;;
	--port) port=${2:?--port needs a number}; shift 2 ;;
	--host) host=${2:?--host needs a name}; shift 2 ;;
	--project) project=${2:?--project needs a name}; shift 2 ;;
	--yes) yes=1; shift ;;
	*) echo "STOP: unknown option $1 (see the top of install.sh)" >&2; exit 2 ;;
	esac
done
stop() { echo "STOP: $*" >&2; exit 1; }
ok() { echo "ok: $*"; }

version=$(cat VERSION)
[[ $port =~ ^[0-9]+$ ]] && [ "$port" -ge 1024 ] && [ "$port" -lt 65535 ] || stop "--port must be a number from 1024 to 65534"
[[ $project =~ ^[a-z0-9][a-z0-9_-]*$ ]] || stop "--project must be lowercase letters, digits, - and _"
answer=$((port + 1))

# 1 tools and host
command -v docker >/dev/null || stop "Docker is not installed. Install Docker Engine and the compose plugin, then run this again."
docker compose version >/dev/null 2>&1 || stop "the docker compose plugin (v2) is missing. Install docker-compose-plugin."
docker info >/dev/null 2>&1 || stop "Docker is not running or your user cannot reach it (try: sudo usermod -aG docker \$USER, then log in again)."
case "$(uname -m)" in x86_64 | amd64) ;; *) stop "this pack is built for x86-64 only; this machine is $(uname -m)." ;; esac
ok "docker and compose are usable on $(uname -m)"

# 2 the download is intact
sha256sum -c --quiet SHA256SUMS >/dev/null 2>&1 || stop "the download is damaged or changed (sha256sum -c SHA256SUMS fails). Fetch it again."
ok "pack checksums match"

# 3 licence note, once per version
if [ "$(cat .licence-ack 2>/dev/null || true)" != "$version" ]; then
	[ "$yes" = 0 ] || stop "--yes is only allowed after you accepted the licence note once; run ./install.sh without it."
	cat licenses/licence-note.txt
	[ -t 0 ] || stop "the licence note needs an answer; run ./install.sh from a terminal."
	read -r -p "Type y to accept and continue: " a
	[ "$a" = y ] || stop "not accepted; nothing was installed."
	echo "$version" >.licence-ack
fi
ok "licence note accepted for $version"

# 4 game folder: read only, the four required files by name
# .env holds the game path in single quotes (a folder name such as "Diablo II" has a space); unquote it again here
shq() { local s=${1//\'/\'\\\'\'}; printf "'%s'" "$s"; }
if [ -z "$game" ] && [ -f .env ]; then
	game=$(sed -n 's/^D2_GAME_DIR=//p' .env)
	game=${game#\'}; game=${game%\'}; game=${game//\'\\\'\'/\'}
fi
if [ -z "$game" ] && [ -t 0 ]; then read -r -p "Path of your own Diablo II 1.14d folder: " game; fi
[ -n "$game" ] || stop "give your Diablo II folder: ./install.sh --game-dir /path/to/Diablo-II"
[ -d "$game" ] || stop "$game is not a folder"
game=$(cd "$game" && pwd)
missing=
for n in d2data.mpq d2exp.mpq d2char.mpq Patch_D2.mpq; do
	f=$(find "$game" -maxdepth 1 -iname "$n" -size +0 -print -quit)
	[ -n "$f" ] || missing="$missing $n"
done
if [ -n "$missing" ]; then
	# the unzipped download has the game one folder down ("Diablo II"): find it instead of only saying "missing"
	hits=$(find "$game" -mindepth 2 -maxdepth 2 -iname d2data.mpq -size +0 -printf '%h\n' 2>/dev/null | sort || true)
	if [ -n "$hits" ] && [ "$(printf '%s\n' "$hits" | wc -l)" = 1 ]; then
		echo "note: the four files are not in $game but in $hits; using that folder."
		game=$hits
		missing=
		for n in d2data.mpq d2exp.mpq d2char.mpq Patch_D2.mpq; do
			f=$(find "$game" -maxdepth 1 -iname "$n" -size +0 -print -quit)
			[ -n "$f" ] || missing="$missing $n"
		done
	fi
fi
[ -z "$missing" ] || stop "missing in $game:$missing. Point --game-dir at your own Diablo II folder (not a copy of this pack)."
ok "game folder $game (only its data archives are mounted, read-only)"

# 5 disk and ports
free=$(df -Pk . | awk 'NR==2 {print $4}')
[ "$free" -ge "$minfree_kb" ] || stop "less than 3 GB free here; free some space."
for p in $port $answer; do
	if command -v ss >/dev/null && ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "[:.]$p\$"; then
		running=$(docker compose -p "$project" ps -q 2>/dev/null || true)
		[ -n "$running" ] || stop "port $p is in use. Pick another: ./install.sh --port 9100 (uses 9100 and 9101)."
	fi
done
ok "free space and ports $port, $answer"

# 6 load the image, check it is the pack's
image=$(cat image/IMAGE_TAG)
docker load -i image/didle-image.tar.gz >/dev/null || stop "docker load failed"
abi=$(docker image inspect --format '{{ index .Config.Labels "io.didle.image.abi" }}' "$image" 2>/dev/null || true)
[ -n "$abi" ] || stop "the loaded image $image has no ABI label; it does not match this pack."
ok "image $image loaded (abi $abi)"

# 7 write .env and start
publish=127.0.0.1
origin="http://localhost:$port,http://127.0.0.1:$port"
shown="http://localhost:$port"
if [ -n "$host" ]; then
	publish=0.0.0.0
	origin="http://$host:$port,$origin"
	shown="http://$host:$port"
fi
# only the data archives d2d reads are mounted, one read-only file each (d2d reads no other file of the game folder: V in the
# fork source, gamedata_load.cpp, apps/d2d/main.cpp). The key file, the game program, its libraries, saves, videos and every other file of the
# folder stay out of the container; no directory is mounted, so no listing shows them either.
yq() { local v=${1//\\/\\\\}; v=${v//\"/\\\"}; printf '"%s"' "$v"; }
{
	echo "# written by install.sh: one read-only mount per data archive of your game folder"
	echo "services:"
	echo "  d2dbot:"
	echo "    volumes:"
	for n in d2data.mpq d2exp.mpq d2char.mpq patch_d2.mpq d2sfx.mpq d2speech.mpq d2xtalk.mpq d2music.mpq d2xmusic.mpq; do
		f=$(find "$game" -maxdepth 1 -iname "$n" -size +0 -print -quit)
		[ -n "$f" ] || continue
		echo "      - type: bind"
		echo "        source: $(yq "$f")"
		echo "        target: /game/$n"
		echo "        read_only: true"
	done
} >compose.game.yml
umask 077
cat >.env <<ENV
COMPOSE_PROJECT_NAME=$project
D2_IMAGE=$image
COMPOSE_FILE=compose.yml:compose.game.yml
D2_GAME_DIR=$(shq "$game")
D2_WATCH_PORT=$port
D2_ANSWER_PORT=$answer
D2_PUBLISH=$publish
D2_WATCH_ORIGIN=$origin
D2_SHOWN_URL=$shown
D2BOT_UPDATE_FEED=$(cat UPDATE_FEED 2>/dev/null || true)
D2BOT_UPDATE_CHECK=on
ENV
docker compose up -d || stop "docker compose could not start the container (output above)"

# 8 wait for the page, print the URL once
for _ in $(seq 1 60); do
	curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$port/" 2>/dev/null && break
	sleep 2
done
curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$port/" 2>/dev/null || stop "the page did not come up in 120 s. Run ./didle report and send the file."
tok=
for _ in $(seq 1 15); do
	tok=$(docker compose exec -T d2dbot cat /data/watch.token 2>/dev/null | tr -d '\r\n' || true)
	[ -n "$tok" ] && break
	sleep 2
done
echo
echo "d.idle $version is running."
echo "Open this exact address in your browser (the token is your password, do not share it):"
echo
echo "    $shown/setup.html?t=${tok:-<run ./didle status>}"
echo
echo "The first start walks you through a short setup, then a 1 to 3 minute warm-up. ./didle status shows this address again."
