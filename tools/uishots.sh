#!/bin/sh
# Watch UI screenshot suite (docs/ui-regression-plan.md section 4). Runs ON the worker host (through labrun):
# a frozen fixture viewer (go run ./tools/uishots serve), headless google-chrome-stable on a private Xvfb display at
# nice 19 (never the user's desktop), one PNG and one DOM dump per page, width and theme, then the DOM contract check
# and a pixel diff against the golden set kept on the worker host (not in git: pictures change with every UI PR).
#   tools/uishots.sh [--update] [--out DIR] [--only NAME]   --update makes this run the golden set
#   tools/uishots.sh --cockpit [--out DIR] [--only NAME]     pictures of the cockpit pages at the supported sizes (ck-NAME-WxH.png), no check, no golden set
#   tools/uishots.sh --char --url '/watch?layout=cockpit-demo&theme=dark' --click '#btn-more|#btn-help' --sizes '1366x768 390x844' --tag NAME   the same for any /watch page (a cockpit sheet or drawer), after the clicks
#   tools/uishots.sh --char [--out DIR] [--hover SEL | --ctx SEL | --click SEL] [--sizes 'WxH ...'] [--theme dark|light] [--fx VARIANT] [--tag NAME]   pictures of the character canvas (/watch/inventory) at the eight fit sizes, in the fit harness
#   tools/uishots.sh --vt [--out DIR] [--only GLOB] [--sizes 'WxH ...']   pictures of the Ethereal Vault page (/watch/vault, fake vaults), vt-NAME-WxH.png, in the fit harness
#   tools/uishots.sh --skills [--out DIR] [--sizes 'WxH ...'] [--tag NAME]   pictures of the Skills page (/watch/skills, fake trees) at a few sizes
#   tools/uishots.sh --pics [--out DIR] [--only GLOB]   pictures of the cockpit (strip and Needs you variants) at the eight fit sizes, in the fit harness
#   tools/uishots.sh --fit [--out DIR]                       the cockpit fit check (tools/uishots/fit.go), on the builder:
#     every cockpit page at the supported sizes loads in an iframe of exactly that size; the page must not scroll, no slot
#     may overflow, no dead space; the deliberately overflowing fixture must FAIL. No pictures, no golden set.
# Exit: 0 ok, 1 a check failed, 2 a tool is missing. Every process has a hard limit; one run per port at a time (flock).
# UISHOTS_PORT and UISHOTS_DISPLAY pick another port and private display; UISHOTS_GOLDEN another golden directory.
set -u
update=0 out=${TMPDIR:-/tmp}/uishots-$(id -u)/run only= fit=0 cockpit=0 char=0 skills=0 charu= extra= sizes_only= theme=dark fx= tag= pics=0 pq= cs=0 vt=0
while [ $# -gt 0 ]; do
	case $1 in
	--update) update=1; shift ;;
	--out) out=$2; shift 2 ;;
	--only) only=$2; shift 2 ;;
	--fit) fit=1; shift ;;
	--cockpit) cockpit=1; shift ;;
	--char) char=1; shift ;;
	--pics) pics=1; shift ;;
	--cs) cs=1; shift ;;
	--vt) vt=1; shift ;;
	--skills) skills=1; shift ;;
	--url) charu=$(printf %s "$2" | sed 's/&/%26/g; s/?/%3F/g'); shift 2 ;;
	--hover) extra="$extra&hover=$(printf %s "$2" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g')"; shift 2 ;;
	--ctx) extra="$extra&ctx=$(printf %s "$2" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g')"; shift 2 ;;
	--sizes) sizes_only=$2; shift 2 ;;
	--theme) theme=$2; shift 2 ;;
	--fx) fx=$2; shift 2 ;;
	--tag) tag=$2-; shift 2 ;;
	--pq) pq="$pq%26$(printf %s "$2" | sed 's/&/%26/g')"; shift 2 ;;
	--click) extra="$extra&click=$(printf %s "$2" | sed 's/_/ /g; s/#/%23/g; s/\[/%5B/g; s/\]/%5D/g; s/|/%7C/g')"; shift 2 ;;
	*) echo "uishots: unknown argument $1" >&2; exit 2 ;;
	esac
done
golden=${UISHOTS_GOLDEN:-$HOME/.cache/didle-uishots/golden}
chrome=$(command -v google-chrome-stable || command -v chromium || command -v chromium-browser) || { echo "uishots: no Chrome or Chromium on this host" >&2; exit 2; }
command -v Xvfb >/dev/null || { echo "uishots: no Xvfb on this host" >&2; exit 2; }
port=${UISHOTS_PORT:-18941} disp=${UISHOTS_DISPLAY:-:241}
exec 9>/tmp/uishots-$(id -u)-$port.lock
flock -w 600 9 || { echo "uishots: another run holds the lock" >&2; exit 2; }
cd "$(dirname "$0")/.." || exit 2
bin=$(mktemp -d)
trap 'kill $sp $xp 2>/dev/null; rm -rf "${bin:?}"' EXIT INT TERM
go build -o "$bin/uishots" ./tools/uishots || exit 2
timeout 2400 nice -n 19 "$bin/uishots" serve -addr 127.0.0.1:$port >"$bin/serve.log" 2>&1 &
sp=$!
timeout 2400 Xvfb $disp -screen 0 1920x1200x24 -nolisten tcp >/dev/null 2>&1 &
xp=$!
sleep 2
if [ "$cockpit" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/ck-*
	sizes="1920x1080 1440x900 1366x768 1914x930 1904x1000 1536x730 1280x620 390x844"
	pages="demo watch?layout=cockpit-demo&theme=dark
real watch?layout=cockpit&theme=dark
zone watch?layout=cockpit&theme=dark&zone=cp
zone-loop watch?layout=cockpit&theme=dark&zone=cp&loop=1
zone-hard watch?layout=cockpit&theme=dark&zone=ftw
zone-town watch?layout=cockpit&theme=dark&zone=camp
zone-phone watch?layout=cockpit&theme=dark&tab=Map&zone=cp
light watch?layout=cockpit-demo&theme=light
tab-orders watch?layout=cockpit-demo&theme=dark&tab=Orders
tab-map watch?layout=cockpit-demo&theme=dark&tab=Map
tab-progress watch?layout=cockpit-demo&theme=dark&tab=Progress
drawer-log watch?layout=cockpit-demo&theme=dark&drawer=log
drawer-start watch?layout=cockpit-demo&theme=dark&drawer=onboarding"
	echo "$pages" | while read -r pname pq; do
		case $pname in $only) ;; *) [ -n "$only" ] && continue ;; esac
		for sz in $sizes; do
			w=${sz%x*} h=${sz#*x}
			case $pname in tab-*|drawer-*) [ "$sz" = 390x844 ] || [ "${pname#drawer}" != "$pname" ] || continue ;; esac
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 60 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--force-prefers-reduced-motion --window-size="$w,$h" --virtual-time-budget=12000 --timeout=30000 --screenshot="$out/ck-$pname-$sz.png" "http://127.0.0.1:$port/$pq" >/dev/null 2>&1
			rm -rf "${prof:?}"
		done
	done
	ls "$out" | wc -l
	exit 0
fi
if [ "$char" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/char-${tag}*
	for sz in ${sizes_only:-1914x930 1536x730 1280x620 1440x900 1920x1080 1366x768 1904x1000 390x844}; do
		w=${sz%x*} h=${sz#*x}
		prof=$(mktemp -d)
		DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
			--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/char-$tag$sz.png" "http://127.0.0.1:$port/__fit?u=${charu:-/watch/inventory%3Ftheme=$theme${fx:+%26fx=$fx}$pq}&w=$w&h=$h$extra" >/dev/null 2>&1
		rm -rf "${prof:?}"
	done
	ls "$out"/char-*
	exit 0
fi
if [ "$cs" = 1 ]; then
	# pictures of the Characters page (docs/charselect-ui.md): name, query, optional click selector; cs-NAME-WxH.png, in the fit harness
	mkdir -p "$out" && rm -f "$out"/cs-*
	echo "c1 chars=1&theme=dark
c3 chars=3&theme=dark
c3l chars=3&theme=light
c8 chars=8&theme=dark
c8l chars=8&theme=light
clong chars=long&theme=dark
cdel chars=deleted&theme=dark
cswitch chars=switching&theme=dark
coff chars=off&theme=dark
cnew chars=3&theme=dark [data-key=new-top]
cnewl chars=3&theme=light [data-key=new-top]
cdelete chars=8&theme=dark [data-key=delete-Zapp]
cdeletel chars=8&theme=light [data-key=delete-Zapp]" | while read -r pname pq pclick; do
		case $pname in $only) ;; *) [ -n "$only" ] && continue ;; esac
		for sz in ${sizes_only:-1914x930 1366x768 1280x620 412x780 360x640}; do
			w=${sz%x*} h=${sz#*x}
			ck=; [ -n "$pclick" ] && ck="&click=$(printf %s "$pclick" | sed 's/#/%23/g')"
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/cs-$pname-$sz.png" "http://127.0.0.1:$port/__fit?u=/watch/characters%3F$(printf %s "$pq" | sed 's/&/%26/g')&w=$w&h=$h$ck" >/dev/null 2>&1
			rm -rf "${prof:?}"
		done
	done
	ls "$out"/cs-* | wc -l
	exit 0
fi
if [ "$vt" = 1 ]; then
	# pictures of the Ethereal Vault page (docs/vault.md): name, query, optional click selectors or hover=SEL; vt-NAME-WxH.png, in the fit harness
	mkdir -p "$out" && rm -f "$out"/vt-*
	echo "pop vault=pop&theme=dark
popl vault=pop&theme=light
list vault=pop&theme=dark [data-key=view-toggle]
warn vault=warn&theme=dark
full vault=full&theme=dark
over vault=over&theme=dark
overl vault=over&theme=light
empty vault=empty&theme=dark
none vault=none&theme=dark
off vault=off&theme=dark
error vault=error&theme=dark
big vault=big&theme=dark
filters vault=pop&theme=dark [data-key=filters-toggle]
select vault=pop&theme=dark [data-key=select-toggle]|.vt-cell|.vt-cell:nth-child(3)
sheet vault=pop&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=discard]
sheettyped vault=pop&theme=dark [data-key=select-toggle]|[data-key=select-all]|[data-key=discard]
sheetl vault=pop&theme=light [data-key=select-toggle]|.vt-cell|[data-key=discard]
tip vault=pop&theme=dark hover=.vt-cell:nth-child(3)" | while read -r pname pq pclick; do
		case $pname in $only) ;; *) [ -n "$only" ] && continue ;; esac
		for sz in ${sizes_only:-1914x930 1366x768 1280x620 412x780 360x640}; do
			w=${sz%x*} h=${sz#*x}
			ck=
			case $pclick in
			hover=*) ck="&$(printf %s "$pclick" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g')" ;;
			?*) ck="&click=$(printf %s "$pclick" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g; s/|/%7C/g')" ;;
			esac
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/vt-$pname-$sz.png" "http://127.0.0.1:$port/__fit?u=/watch/vault%3F$(printf %s "$pq" | sed 's/&/%26/g')&w=$w&h=$h$ck" >/dev/null 2>&1
			rm -rf "${prof:?}"
		done
	done
	ls "$out"/vt-* | wc -l
	exit 0
fi
if [ "$skills" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/skills-*
	for sz in ${sizes_only:-1914x930 1366x768 1024x768 390x844}; do
		w=${sz%x*} h=${sz#*x}
		prof=$(mktemp -d)
		DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
			--window-size="$w,$h" --virtual-time-budget=12000 --timeout=30000 --screenshot="$out/skills-$tag$sz.png" "http://127.0.0.1:$port/watch/skills?theme=$theme" >/dev/null 2>&1
		rm -rf "${prof:?}"
	done
	ls "$out"/skills-*
	exit 0
fi
if [ "$pics" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/pic-*
	echo "default first=0&theme=dark&needs=2
default-needs0 first=0&theme=dark&needs=0
default-needs5 first=0&theme=dark&needs=5
default-measuring first=0&theme=dark&needs=0&glance=measuring
default-blocked first=0&theme=dark&needs=1&glance=blocked
default-unknown first=0&theme=dark&needs=0&glance=unknown
default-light first=0&theme=light&needs=2
default-progresstab first=0&theme=dark&needs=2&tab=Progress
default-worst first=0&theme=dark&needs=5&cmd=eight&glance=full,mercdead,supplieslow
default-worst-light first=0&theme=light&needs=5&cmd=eight&glance=full,mercdead,supplieslow
default-paused first=0&theme=dark&needs=2&cmd=paused&glance=full,paused
default-loop first=0&theme=dark&needs=2&cmd=loop
default-live first=0&theme=dark&needs=2&stream=live
default-zone first=0&theme=dark&needs=2&zone=cp&cmd=loop
default-update first=0&theme=dark&needs=2&update=available
default-update-light first=0&theme=light&needs=2&update=ready
default-frozen first=0&theme=dark&needs=2&pause=frozen&glance=full,paused
default-frozen-light first=0&theme=light&needs=2&pause=frozen&glance=full,paused
default-question3 first=0&theme=dark&needs=1&question=3
default-question5-light first=0&theme=light&needs=2&question=5
quest-offer first=0&theme=dark&needs=0&quests=offer
quest-running first=0&theme=dark&needs=0&quests=running
quest-awaiting first=0&theme=dark&needs=0&quests=awaiting
quest-handing first=0&theme=dark&needs=0&quests=handing
quest-confirming first=0&theme=dark&needs=0&quests=confirming
quest-failed first=0&theme=dark&needs=0&quests=failed
quest-blocked first=0&theme=dark&needs=0&quests=blocked
quest-training first=0&theme=dark&needs=0&quests=training
quest-done first=0&theme=dark&needs=0&quests=done
quest-off first=0&theme=dark&needs=0&quests=off
quest-offer-light first=0&theme=light&needs=0&quests=offer
quest-awaiting-light first=0&theme=light&needs=0&quests=awaiting
quest-blocked-light first=0&theme=light&needs=0&quests=blocked
quest-offer-sorceress first=0&theme=dark&needs=0&quests=offer&class=sorceress
quest-needs3 first=0&theme=dark&needs=2&quests=awaiting
quest-phone-tab first=0&theme=dark&needs=0&quests=offer&tab=Quests
quest-phone-tab-awaiting first=0&theme=dark&needs=0&quests=awaiting&tab=Quests
quest-phone-tab-off first=0&theme=dark&needs=0&quests=off&tab=Quests
quest-phone-tab-light first=0&theme=light&needs=0&quests=running&tab=Quests
quest-sheet-offer first=0&theme=dark&needs=0&quests=offer #btn-more|[data-drawer-open=quest-board]
quest-sheet-running first=0&theme=dark&needs=0&quests=running #btn-more|[data-drawer-open=quest-board]
quest-sheet-awaiting first=0&theme=dark&needs=0&quests=awaiting #btn-more|[data-drawer-open=quest-board]
quest-sheet-blocked first=0&theme=dark&needs=0&quests=blocked #btn-more|[data-drawer-open=quest-board]
quest-sheet-auto first=0&theme=dark&needs=0&quests=auto #btn-more|[data-drawer-open=quest-board]
quest-sheet-mid first=0&theme=dark&needs=0&quests=mid #btn-more|[data-drawer-open=quest-board]
quest-sheet-off first=0&theme=dark&needs=0&quests=off #btn-more|[data-drawer-open=quest-board]
quest-sheet-training first=0&theme=dark&needs=0&quests=training #btn-more|[data-drawer-open=quest-board]
quest-sheet-light first=0&theme=light&needs=0&quests=offer #btn-more|[data-drawer-open=quest-board]
quest-sheet-auto-light first=0&theme=light&needs=0&quests=auto #btn-more|[data-drawer-open=quest-board]
quest-sheet-sorceress first=0&theme=dark&needs=0&quests=offer&class=sorceress #btn-more|[data-drawer-open=quest-board]
overlay-menu-zone first=0&theme=dark&zone=cp&tab=Map #btn-more
overlay-menu-zone-light first=0&theme=light&zone=cp&tab=Map #btn-more
overlay-charter first=0&theme=dark&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go]
overlay-charter-light first=0&theme=light&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go]
overlay-charter-menu first=0&theme=dark&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go]|#btn-more" | while read -r pname pq pclick; do
		case $pname in $only) ;; *) [ -n "$only" ] && continue ;; esac
		ck=
		[ -n "$pclick" ] && ck="&click=$(printf %s "$pclick" | sed 's/_/ /g; s/#/%23/g; s/?/%3F/g; s/|/%7C/g; s/&/%26/g')"
		for sz in ${sizes_only:-1914x930 1536x730 1280x620 1440x900 1920x1080 1366x768 1904x1000 390x844}; do
			w=${sz%x*} h=${sz#*x}
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/pic-$pname-$sz.png" "http://127.0.0.1:$port/__fit?u=/watch%3F$(printf %s "$pq" | sed 's/&/%26/g')&w=$w&h=$h$ck" >/dev/null 2>&1
			rm -rf "${prof:?}"
		done
	done
	ls "$out"/pic-* | head -80
	exit 0
fi
if [ "$fit" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/fit-*
	# the full-screen sizes, the browser-window sizes (a window with its chrome is shorter than the screen) and the phone
	sizes="1920x1080 1440x900 1366x768 1914x930 1904x1000 1536x730 1280x620 390x844"
	# the default page also at the window sizes between the phone and the wide cockpit (the Electron viewer opens /watch in any of them)
	midsizes="1024x768 900x700 1279x720"
	phonesizes="412x780 393x700 360x640 412x890 390x844"
	# name, query, and an optional selector to press before the second measurement (the "..." sheet must be hittable)
	pages="default watch?first=0&theme=dark
default-light watch?first=0&theme=light
default-menu watch?first=0&theme=dark #btn-more
default-needs0 watch?first=0&theme=dark&needs=0
default-needs5-orders8 watch?first=0&theme=dark&needs=5&cmd=eight
default-needs5-orders8-light watch?first=0&theme=light&needs=5&cmd=eight&glance=full,mercdead,supplieslow
default-empty watch?first=0&theme=dark&needs=0&cmd=empty
default-live-stream watch?first=0&theme=dark&needs=2&stream=live
default-live-stream-light watch?first=0&theme=light&needs=2&stream=live
default-live-pal watch?first=0&theme=dark&needs=2&pal=1
default-pfdebug watch?first=0&theme=dark&needs=2&pal=1&pfdebug=1
default-paused watch?first=0&theme=dark&needs=2&cmd=paused&glance=full,paused
default-phone-progress watch?first=0&theme=dark&needs=2&tab=Progress
default-phone-map watch?first=0&theme=dark&needs=2&tab=Map
default-zone watch?first=0&theme=dark&zone=cp
default-drawer-log watch?first=0&theme=dark #btn-more|[data-drawer-open=log] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
default-update watch?first=0&theme=dark&needs=2&update=available
default-update-light watch?first=0&theme=light&needs=2&update=ready
default-frozen watch?first=0&theme=dark&needs=2&pause=frozen&glance=full,paused expect=#pause-banner
default-frozen-light watch?first=0&theme=light&needs=2&pause=frozen&glance=full,paused expect=#pause-banner
default-pause-menu watch?first=0&theme=dark&needs=2&pause=can #btn-more
default-question3 watch?first=0&theme=dark&needs=1&question=3
default-question5-light watch?first=0&theme=light&needs=2&question=5
quest-offer watch?first=0&theme=dark&needs=0&quests=offer
quest-running watch?first=0&theme=dark&needs=0&quests=running
quest-awaiting watch?first=0&theme=dark&needs=0&quests=awaiting
quest-handing watch?first=0&theme=dark&needs=0&quests=handing
quest-confirming watch?first=0&theme=dark&needs=0&quests=confirming
quest-failed watch?first=0&theme=dark&needs=0&quests=failed
quest-blocked watch?first=0&theme=dark&needs=0&quests=blocked
quest-training watch?first=0&theme=dark&needs=0&quests=training
quest-done watch?first=0&theme=dark&needs=0&quests=done
quest-off watch?first=0&theme=dark&needs=0&quests=off
quest-offer-light watch?first=0&theme=light&needs=0&quests=offer
quest-awaiting-light watch?first=0&theme=light&needs=0&quests=awaiting
quest-offer-sorceress watch?first=0&theme=dark&needs=0&quests=offer&class=sorceress
quest-needs3 watch?first=0&theme=dark&needs=2&quests=awaiting
quest-phone-tab watch?first=0&theme=dark&needs=0&quests=offer&tab=Quests
quest-phone-tab-awaiting watch?first=0&theme=dark&needs=0&quests=awaiting&tab=Quests
quest-phone-tab-off watch?first=0&theme=dark&needs=0&quests=off&tab=Quests
quest-sheet-offer watch?first=0&theme=dark&needs=0&quests=offer #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-running watch?first=0&theme=dark&needs=0&quests=running #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-awaiting watch?first=0&theme=dark&needs=0&quests=awaiting #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-blocked watch?first=0&theme=dark&needs=0&quests=blocked #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-auto watch?first=0&theme=dark&needs=0&quests=auto #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-mid watch?first=0&theme=dark&needs=0&quests=mid #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-off watch?first=0&theme=dark&needs=0&quests=off #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-training watch?first=0&theme=dark&needs=0&quests=training #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-light watch?first=0&theme=light&needs=0&quests=offer #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-sorceress watch?first=0&theme=dark&needs=0&quests=offer&class=sorceress #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
questdefect-clip watch?first=0&theme=dark&needs=0&quests=offer&defect=q-clip #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
questdefect-small watch?first=0&theme=dark&needs=0&quests=offer&defect=q-small #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
questdefect-words watch?first=0&theme=dark&needs=0&quests=offer&defect=q-words #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
questdefect-two watch?first=0&theme=dark&needs=0&quests=offer&defect=q-two #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
reset-header watch?first=0&theme=dark&needs=0 #btn-reset #btn-reset
reset-header-light watch?first=0&theme=light&needs=0 #btn-reset #btn-reset
reset-pressed watch?first=0&theme=dark&needs=0&reset=resetting #btn-reset #btn-reset,#reset-progress
reset-pressed-light watch?first=0&theme=light&needs=0&reset=rejoining #btn-reset #btn-reset,#reset-progress
reset-return-menu watch?first=0&theme=dark&needs=0 #btn-more [data-key=menu-return]
reset-return-menu-light watch?first=0&theme=light&needs=0 #btn-more [data-key=menu-return]
reset-menu watch?first=0&theme=dark&needs=0&reset=resetting #btn-more|[data-key=reset-open] #reset-progress
reset-more watch?first=0&theme=dark&needs=0&reset=resetting [data-key=verb-more]|[data-key=more-reset] #reset-progress
reset-hard-confirm watch?first=0&theme=dark&needs=0&reset=failed [data-key=reset-hard] #reset-confirm,#reset-go,[data-key=reset-cancel]
reset-hard-confirm-light watch?first=0&theme=light&needs=0&reset=failed [data-key=reset-hard] #reset-confirm,#reset-go,[data-key=reset-cancel]
support-sheet watch?first=0&theme=dark&needs=0 #btn-more|[data-key=support-open] #support-sheet,#support-build,[data-key=support-cancel],#support-shot,#support-save
support-sheet-light watch?first=0&theme=light&needs=0 #btn-more|[data-key=support-open] #support-sheet,#support-build,[data-key=support-cancel],#support-shot,#support-save
support-preview watch?first=0&theme=dark&needs=0 #btn-more|[data-key=support-open]|#support-build #support-sheet,#support-send,#support-save-file,[data-key=support-back],[data-key=support-cancel]
support-preview-light watch?first=0&theme=light&needs=0 #btn-more|[data-key=support-open]|#support-build #support-sheet,#support-send,#support-save-file,[data-key=support-back],[data-key=support-cancel]
lootrules-sheet watch?first=0&theme=dark&needs=0 #btn-more|[data-key=lootrules-open] #lootrules-sheet,[data-key=lr-mode-starter],[data-key=lr-mode-custom],#lr-apply,[data-key=lr-close]
lootrules-sheet-light watch?first=0&theme=light&needs=0 #btn-more|[data-key=lootrules-open] #lootrules-sheet,[data-key=lr-mode-starter],[data-key=lr-mode-custom],#lr-apply,[data-key=lr-close]
lootrules-custom watch?first=0&theme=dark&needs=0&lootrules=custom #btn-more|[data-key=lootrules-open] #lootrules-sheet,[data-key=lr-mode-starter],[data-key=lr-save],[data-key=lr-close]
lootrules-custom-light watch?first=0&theme=light&needs=0&lootrules=custom #btn-more|[data-key=lootrules-open] #lootrules-sheet,[data-key=lr-mode-starter],[data-key=lr-save],[data-key=lr-close]
lootrules-empty watch?first=0&theme=dark&needs=0&lootrules=empty #btn-more|[data-key=lootrules-open] #lootrules-sheet,[data-key=lr-close]
lootrules-preview watch?first=0&theme=dark&needs=0 #btn-more|[data-key=lootrules-open]|[data-key=lr-preview] #lootrules-sheet,[data-key=lr-preview],[data-key=lr-undo],[data-key=lr-import],[data-key=lr-close]
lootrules-preview-light watch?first=0&theme=light&needs=0 #btn-more|[data-key=lootrules-open]|[data-key=lr-preview] #lootrules-sheet,[data-key=lr-preview],[data-key=lr-undo],[data-key=lr-import],[data-key=lr-close]
lootrules-sellkinds watch?first=0&theme=dark&needs=0 #btn-more|[data-key=lootrules-open]|[data-key=lr-sellkinds] #lootrules-sheet,[data-key=lr-sellkinds],[data-key=lr-close]
lootrules-sellkinds-light watch?first=0&theme=light&needs=0 #btn-more|[data-key=lootrules-open]|[data-key=lr-sellkinds] #lootrules-sheet,[data-key=lr-sellkinds],[data-key=lr-close]
lootrules-error watch?first=0&theme=dark&needs=0 #btn-more|[data-key=lootrules-open]|#lr-apply #lootrules-sheet,[data-key=lr-close]
support-blocked watch?first=0&theme=dark&needs=0&support=blocked #btn-more|[data-key=support-open]|#support-build #support-sheet,[data-key=support-back],[data-key=support-cancel]
support-nosend watch?first=0&theme=dark&needs=0&support=nosend #btn-more|[data-key=support-open]|#support-build #support-sheet,#support-save-file,[data-key=support-back]
reset-resetting watch?first=0&theme=dark&needs=0&reset=resetting expect=#reset-progress
reset-resetting-light watch?first=0&theme=light&needs=0&reset=rejoining expect=#reset-progress
phonedefect-resetclip watch?first=0&theme=dark&needs=0&defect=resetclip #btn-reset #btn-reset
default-firstrun watch?theme=dark&needs=2 ?#none .d-drawer:not([hidden])_[data-key=drawer-close]
cockpit watch?layout=cockpit&theme=dark
cockpit-menu watch?layout=cockpit&theme=dark #btn-more
orders-eight watch?layout=cockpit&theme=dark&cmd=eight
strip-needs5-orders8 watch?layout=cockpit&theme=dark&needs=5&cmd=eight
strip-needs5-orders8-light watch?layout=cockpit&theme=light&needs=5&cmd=eight&glance=full,mercdead,supplieslow
paused watch?layout=cockpit&theme=dark&needs=2&cmd=paused&glance=full,paused
paused-light watch?layout=cockpit&theme=light&needs=2&cmd=paused&glance=full,paused
frozen watch?layout=cockpit&theme=dark&needs=2&pause=frozen&glance=full,paused expect=#pause-banner
frozen-light watch?layout=cockpit&theme=light&needs=2&pause=frozen&glance=full,paused expect=#pause-banner
frozen-menu watch?layout=cockpit&theme=dark&needs=2&pause=can #btn-more
loop-row watch?layout=cockpit&theme=dark&needs=2&cmd=loop
loop-row-light watch?layout=cockpit&theme=light&needs=2&cmd=loop
merc-dead watch?layout=cockpit&theme=dark&needs=4&glance=full,mercdead
low-supplies watch?layout=cockpit&theme=dark&needs=2&glance=full,supplieslow
merc-dead-low-supplies-light watch?layout=cockpit&theme=light&needs=4&glance=full,mercdead,supplieslow
no-merc-no-potions watch?layout=cockpit&theme=dark&needs=2&glance=full,nopotions
live-stream watch?layout=cockpit&theme=dark&needs=2&stream=live
live-stream-light watch?layout=cockpit&theme=light&needs=2&stream=live
live-pal watch?layout=cockpit&theme=dark&needs=2&pal=1
live-pfdebug watch?layout=cockpit&theme=dark&needs=2&pal=1&pfdebug=1
sentence-popover watch?layout=cockpit&theme=dark&needs=2&glance=full,paused [data-key=say] .d-bandpop:not([hidden])
journey-popover watch?layout=cockpit&theme=dark&needs=2&tab=Progress [data-panel=journey-strip]_.jr-dot[data-cur] [data-panel=journey-strip]_[data-key=pact-0],[data-panel=journey-strip]_[data-key=pact-1]
needs-popover watch?layout=cockpit&theme=dark&needs=5 [data-panel=needs-you]_.n-dots [data-panel=needs-you]_.n-pop
orders-rail watch?layout=cockpit-orders&theme=dark&needs=2
zoneclear-rail watch?layout=cockpit-zoneclear&theme=dark&needs=2
zoneclear-rail-light watch?layout=cockpit-zoneclear&theme=light&needs=2
orders-rail-eight watch?layout=cockpit-orders&theme=dark&needs=0&cmd=eight
orders-rail-all watch?layout=cockpit-orders&theme=dark&needs=0&cmd=cancel-all
orders-empty watch?layout=cockpit&theme=dark&cmd=empty
orders-safety watch?layout=cockpit&theme=dark&cmd=safety
orders-blocked watch?layout=cockpit&theme=dark&cmd=blocked
orders-all watch?layout=cockpit&theme=dark&cmd=cancel-all
strip-needs0 watch?layout=cockpit&theme=dark&needs=0
strip-needs2 watch?layout=cockpit&theme=dark&needs=2
strip-needs5 watch?layout=cockpit&theme=dark&needs=5
strip-needs9 watch?layout=cockpit&theme=dark&needs=9
strip-measuring watch?layout=cockpit&theme=dark&needs=0&glance=measuring
strip-blocked watch?layout=cockpit&theme=dark&needs=3&glance=blocked
strip-ch9 watch?layout=cockpit&theme=dark&needs=2&glance=ch9
strip-gearbroken watch?layout=cockpit&theme=dark&needs=2&glance=gearbroken
strip-unknown watch?layout=cockpit&theme=dark&needs=0&glance=unknown
strip-nojourney watch?layout=cockpit&theme=dark&needs=2&glance=nojourney
strip-light watch?layout=cockpit&theme=light&needs=2
strip-phonetab watch?layout=cockpit&theme=dark&needs=2&tab=Progress
demo-needs0 watch?layout=cockpit-demo&theme=dark&needs=0
demo-needs2 watch?layout=cockpit-demo&theme=dark&needs=2
demo-needs5 watch?layout=cockpit-demo&theme=dark&needs=5
demo-orders20 watch?layout=cockpit-demo&theme=dark&needs=0&orders=20
demo-menu watch?layout=cockpit-demo&theme=dark #btn-more
cockpit-tab-map watch?layout=cockpit&theme=dark&tab=Map
zone-menu watch?layout=cockpit&theme=dark&zone=cp
zone-menu-phone watch?layout=cockpit&theme=dark&tab=Map&zone=cp
zone-menu-loop watch?layout=cockpit&theme=dark&zone=cp&loop=1
zone-menu-town watch?layout=cockpit&theme=dark&zone=camp
zone-menu-hard watch?layout=cockpit&theme=dark&zone=ftw
zone-menu-here watch?layout=cockpit&theme=dark&zone=bm
zone-menu-orders watch?layout=cockpit-orders&theme=dark&zone=cp
demo-tab-orders watch?layout=cockpit-demo&theme=dark&tab=Orders
demo-tab-map watch?layout=cockpit-demo&theme=dark&tab=Map
demo-tab-progress watch?layout=cockpit-demo&theme=dark&tab=Progress
demo-areaswitch watch?layout=cockpit-demo&theme=dark&tab=Map ?.d-mapswitch_>_.d-btn:nth-child(2) [data-panel=area-map]
demo-layers watch?layout=cockpit-demo&theme=dark&tab=Map ?.d-mapswitch_>_.d-btn:nth-child(2)|[data-panel=area-map]_[data-key=layers] [data-panel=area-map]_.layers,[data-panel=area-map]_[data-key=coverage]
demo-drawer-log watch?layout=cockpit-demo&theme=dark #btn-more|[data-drawer-open=log] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close],.d-drawer:not([hidden])_[data-key=tab-alerts]
demo-drawer-alerts watch?layout=cockpit-demo&theme=dark #btn-more|[data-drawer-open=log]|[data-key=tab-alerts] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
demo-drawer-start watch?layout=cockpit-demo&theme=dark #btn-more|[data-drawer-open=onboarding] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
demo-drawer-items watch?layout=cockpit-demo&theme=dark #btn-more|[data-drawer-open=log] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close],.d-drawer:not([hidden])_[data-key=tab-items]
demo-help watch?layout=cockpit-demo&theme=dark #btn-more|#btn-help .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
overflow-fixture watch?layout=cockpit-overflow&theme=dark
overflow-nested watch?layout=cockpit-nested&theme=dark
overflow-fakescroll watch?layout=cockpit-fakescroll&theme=dark
overflow-classic watch?layout=classic&theme=dark
blank-rail watch?layout=cockpit-blank&theme=dark
tall-frame watch?layout=cockpit-tall&theme=dark
chars-1 watch/characters?theme=dark&chars=1
chars-1-light watch/characters?theme=light&chars=1
chars-3 watch/characters?theme=dark&chars=3
chars-3-light watch/characters?theme=light&chars=3
chars-8 watch/characters?theme=dark&chars=8
chars-8-light watch/characters?theme=light&chars=8
chars-long watch/characters?theme=dark&chars=long
chars-long-light watch/characters?theme=light&chars=long
chars-deleted watch/characters?theme=dark&chars=deleted
chars-deleted-light watch/characters?theme=light&chars=deleted
chars-switching watch/characters?theme=dark&chars=switching
chars-switching-light watch/characters?theme=light&chars=creating
chars-measured watch/characters?theme=dark&chars=measured
chars-empty watch/characters?theme=dark&chars=empty
chars-off watch/characters?theme=dark&chars=off
chars-off-light watch/characters?theme=light&chars=off
chars-nosave watch/characters?theme=dark&chars=nosave
chars-new watch/characters?theme=dark&chars=3 [data-key=new-top] #cs-sheet,#cs-name,[data-key=sheet-go],[data-key=sheet-cancel]
chars-new-light watch/characters?theme=light&chars=8 [data-key=new-top] #cs-sheet,#cs-name,[data-key=sheet-go],[data-key=sheet-cancel]
chars-switch-sheet watch/characters?theme=dark&chars=3 [data-key=switch-Zapp] #cs-sheet,[data-key=sheet-go],[data-key=sheet-cancel]
chars-delete watch/characters?theme=dark&chars=8 [data-key=delete-Zapp] #cs-sheet,#cs-typed,[data-key=sheet-go],[data-key=sheet-cancel]
chars-delete-light watch/characters?theme=light&chars=long [data-key=delete-Zzzzzzzzzzzzzzz] #cs-sheet,#cs-typed,[data-key=sheet-go],[data-key=sheet-cancel]
chars-delete-active watch/characters?theme=dark&chars=3 [data-key=delete-Holy] #cs-sheet,[data-key=sheet-cancel]
chars-restore watch/characters?theme=dark&chars=deleted .cs-delrow_.d-btn #cs-sheet,[data-key=sheet-go],[data-key=sheet-cancel]
chars-menu-entry watch?first=0&theme=dark&needs=0 #btn-more [data-key=characters-open]
chars-more-entry watch?first=0&theme=dark&needs=0 [data-key=verb-more] [data-key=more-characters]
overlay-menu-over-zone watch?first=0&theme=dark&zone=cp&tab=Map #btn-more #btn-more
overlay-menu-over-zone-light watch?first=0&theme=light&zone=cp&tab=Map #btn-more #btn-more
overlay-menu-over-zone-cockpit watch?layout=cockpit&theme=dark&zone=cp&tab=Map #btn-more #btn-more
overlay-zone-after-menu watch?first=0&theme=dark&tab=Map #btn-more|[data-panel=zone-map]_.zm-n #btn-more
overlay-charter watch?first=0&theme=dark&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go] #btn-more
overlay-charter-light watch?first=0&theme=light&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go] #btn-more
overlay-charter-cockpit watch?layout=cockpit&theme=dark&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go] #btn-more
overlay-charter-menu watch?first=0&theme=dark&zone=cp&charter=1&tab=Map [data-panel=zone-map]_[data-key=go]|#btn-more #btn-more
overlay-zone-sheet-light watch?first=0&theme=light&zone=cp&tab=Map
overlay-zone-sheet-loop watch?first=0&theme=dark&zone=cp&loop=1&tab=Map
overlaydefect-under watch?first=0&theme=dark&zone=cp&defect=ov-under&tab=Map #btn-more #btn-more
overlaydefect-clip watch?first=0&theme=dark&zone=cp&charter=1&defect=ov-clip&tab=Map [data-panel=zone-map]_[data-key=go] #btn-more
overlaydefect-header watch?first=0&theme=dark&zone=cp&defect=ov-header&tab=Map
overlaydefect-two watch?first=0&theme=dark&defect=ov-two&tab=Map #btn-more #btn-more
character watch/inventory?theme=dark
character-light watch/inventory?theme=light
character-tip-stash watch/inventory?theme=dark hover=[data-key=item-31]
character-tip-pack watch/inventory?theme=dark hover=[data-key=item-10]
character-tip-ring watch/inventory?theme=dark hover=[data-key=item-11]
character-tip-worn watch/inventory?theme=dark hover=[data-key=slot-rhand]
character-tip-plate watch/inventory?theme=dark hover=[data-key=item-16]
character-menu-pack watch/inventory?theme=dark ctx=[data-key=item-10]
character-menu-worn watch/inventory?theme=dark ctx=[data-key=slot-torso]
character-menu-stash watch/inventory?theme=dark ctx=[data-key=item-31]
character-nomerc watch/inventory?theme=dark&fx=nomerc
character-absentmerc watch/inventory?theme=dark&fx=absentmerc
character-deadmerc watch/inventory?theme=dark&fx=deadmerc
character-deadmerc-light watch/inventory?theme=light&fx=deadmerc
character-longnames watch/inventory?theme=dark&fx=longnames
character-longnames-light watch/inventory?theme=light&fx=longnames
character-pending watch/inventory?theme=dark&fx=pending&tab=pack expect=[data-key=item-11][data-pending=stash][data-wait=town]|[data-key=item-10][data-pending=equip]
character-phonepack watch/inventory?theme=dark ?#cv-tabbar_[data-tab=pack]|?#pack_[data-key=item-10] #belt,#cv-sel
character-idwait watch/inventory?theme=dark&fx=idwait&tab=pack expect=[data-key=item-16][data-pending=identify][data-waiting]
character-iddoing watch/inventory?theme=dark&fx=iddoing&tab=pack expect=[data-key=item-16][data-pending=identify]
character-tip-light watch/inventory?theme=light hover=[data-key=item-16]
character-tip-plate-nomerc watch/inventory?theme=dark&fx=nomerc hover=[data-key=item-16]
character-phone watch/inventory?theme=dark&fx=phone&tab=pack
character-phone-light watch/inventory?theme=light&fx=phone&tab=pack
character-phone-sheet watch/inventory?theme=dark&fx=phone&tab=pack&sheet=1 expect=.cv-acts>button
character-phone-sheet-light watch/inventory?theme=light&fx=phone&tab=pack&sheet=1 expect=.cv-acts>button
character-phone-errands watch/inventory?theme=dark&fx=phone&tab=pack ?#cv-errtoggle #errands,#stashgold
character-phone-menu watch/inventory?theme=dark&fx=phone&tab=pack ?#cv-more .cv-morelist_a
character-phonehero watch/inventory?theme=dark
phonedefect-label watch/inventory?theme=dark&fx=phone&tab=pack&defect=label
phonedefect-acts watch/inventory?theme=dark&fx=phone&tab=pack&sheet=1&defect=acts
phonedefect-cutchips watch/inventory?theme=dark&fx=phone&tab=pack&defect=cutchips
phonedefect-lock watch/inventory?theme=dark&fx=phone&tab=pack&defect=lock
phonedefect-hover watch/inventory?theme=dark&fx=phone&tab=pack&defect=hover
phonedefect-narrow watch/inventory?theme=dark&fx=phone&tab=pack&defect=narrow
charsdefect-clip watch/characters?theme=dark&chars=long&defect=chars-clip
charsdefect-wide watch/characters?theme=dark&chars=3&defect=chars-wide
charsdefect-under watch/characters?theme=dark&chars=3&defect=chars-under
charsdefect-small watch/characters?theme=dark&chars=3&defect=chars-small
charsdefect-hole watch/characters?theme=dark&chars=1&defect=chars-hole
charsdefect-sheet watch/characters?theme=dark&chars=3&defect=chars-sheet [data-key=delete-Zapp] #cs-sheet,[data-key=sheet-go]
charsdefect-covered watch/characters?theme=dark&chars=3&defect=chars-covered
vault-pop watch/vault?theme=dark&vault=pop
vault-pop-light watch/vault?theme=light&vault=pop
vault-list watch/vault?theme=dark&vault=pop [data-key=view-toggle] .vt-cell,[data-key=view-toggle]
vault-warn watch/vault?theme=dark&vault=warn
vault-full watch/vault?theme=dark&vault=full
vault-over watch/vault?theme=dark&vault=over
vault-over-light watch/vault?theme=light&vault=over
vault-empty watch/vault?theme=dark&vault=empty
vault-none watch/vault?theme=dark&vault=none
vault-off watch/vault?theme=dark&vault=off
vault-error watch/vault?theme=dark&vault=error
vault-big watch/vault?theme=dark&vault=big
vault-filters watch/vault?theme=dark&vault=pop [data-key=filters-toggle] [data-key=q-7],[data-key=search]
vault-select watch/vault?theme=dark&vault=pop [data-key=select-toggle]|.vt-cell|.vt-cell:nth-child(3) [data-key=select-toggle],[data-key=pin],[data-key=discard]
vault-sheet watch/vault?theme=dark&vault=pop [data-key=select-toggle]|.vt-cell|[data-key=discard] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-sheet-typed watch/vault?theme=dark&vault=pop [data-key=select-toggle]|[data-key=select-all]|[data-key=discard] .vt-sheet-card,[data-key=sheet-typed],[data-key=sheet-cancel]
vault-sheet-light watch/vault?theme=light&vault=pop [data-key=select-toggle]|.vt-cell|[data-key=discard] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-tip watch/vault?theme=dark&vault=pop hover=.vt-cell:nth-child(3)
vault-menu-entry watch?first=0&theme=dark&needs=0 #btn-more [data-key=vault-open]
vault-more-entry watch?first=0&theme=dark&needs=0 [data-key=verb-more] [data-key=more-vault]
vaultdefect-clip watch/vault?theme=dark&vault=pop&defect=vault-clip
vaultdefect-wide watch/vault?theme=dark&vault=pop&defect=vault-wide
vaultdefect-under watch/vault?theme=dark&vault=pop&defect=vault-under
vaultdefect-small watch/vault?theme=dark&vault=pop&defect=vault-small
vaultdefect-covered watch/vault?theme=dark&vault=pop&defect=vault-covered
vaultdefect-sheet watch/vault?theme=dark&vault=pop&defect=vault-sheet [data-key=select-toggle]|.vt-cell|[data-key=discard] .vt-sheet-card,[data-key=sheet-cancel]
vaultdefect-noname watch/vault?theme=dark&vault=pop&defect=vault-noname
vaultdefect-enabled watch/vault?theme=dark&vault=pop&defect=vault-enabled"
	# in a click or hit selector "_" stands for a space and a leading "?" makes a step optional (no switch on a wide rail);
	# a third column hover=SEL or ctx=SEL hovers or right-clicks that cell before the measurement (the tooltip, the menu)
	echo "$pages" | while read -r pname pq pclick phit; do
		szs=$sizes
		case $pname in
		default | default-light) szs="$sizes $midsizes" ;;
		reset-return-menu*) szs="$sizes 900x700 412x780 360x640" ;; # the Return screen link lives in the menu on every width
		reset-* | support-*) szs="$sizes 412x780 360x640" ;; # the real phone sizes too
		# the real phone sizes (visible area of Chrome with its URL bar; the first fixtures were 390x844 only and missed the 2026-11-15 bug)
		lootrules-*) szs="$sizes 412x780 360x640" ;; # the real phone sizes too
		phonedefect-*) szs="412x780 360x640" ;;
		charsdefect-*) szs="1280x620 412x780 360x640" ;;
		chars-*) szs="$sizes 412x780 360x640" ;;
		vaultdefect-*) szs="1280x620 412x780 360x640" ;;
		vault-menu-entry | vault-more-entry) ;;
		vault-*) szs="$sizes 412x780 360x640" ;;
		questdefect-clip | questdefect-words | questdefect-two) szs="1366x768 412x780" ;;
		questdefect-small) szs="412x780 360x640" ;;
		quest-phone-tab*) szs="412x780 390x844 360x640" ;;
		quest-sheet-*) szs="$sizes 412x780 360x640" ;;
		quest-*) szs="$sizes 412x780" ;;
		overlaydefect-two) szs="1366x768" ;; # the other tiles are not on screen on a phone: nothing to show
		overlaydefect-*) szs="1366x768 412x780" ;;
		overlay-*) szs="$sizes 412x780" ;;
		character-phone*) szs="$phonesizes" ;;
		esac
		for sz in $szs; do
			w=${sz%x*} h=${sz#*x} ck=
			enc() { printf %s "$1" | sed 's/_/ /g; s/#/%23/g; s/?/%3F/g; s/|/%7C/g; s/&/%26/g'; }
			case $pclick in
			hover=* | ctx=* | expect=*) ck="&$(printf %s "$pclick" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g; s/|/%7C/g')" ;;
			?*) ck="&click=$(enc "$pclick")" ;;
			esac
			[ -n "$phit" ] && ck="$ck&hit=$(enc "$phit" | sed 's/%7C/|/g')"
			if [ -n "$only" ]; then case $pname in $only) ;; *) continue ;; esac; fi
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$h" --virtual-time-budget=25000 --timeout=40000 --dump-dom "http://127.0.0.1:$port/__fit?u=/$(printf %s "$pq" | sed 's/&/%26/g')&w=$w&h=$h$ck" >"$out/fit-$pname-$sz.html" 2>/dev/null
			rm -rf "${prof:?}"
		done
	done
	"$bin/uishots" fit "$out"
	exit $?
fi
mkdir -p "$out" && rm -f "$out"/*.png "$out"/*.html "$out"/*.log
# name width height path?query
shots="watch-390-dark 390 844 watch?theme=dark
watch-390-light 390 844 watch?theme=light
watch-768-dark 768 1024 watch?theme=dark
watch-1280-dark 1280 800 watch?theme=dark
watch-1440-dark 1440 900 watch?theme=dark
watch-1440-light 1440 900 watch?theme=light
watch-1440full-dark 1440 2600 watch?theme=dark
watch-1920-dark 1920 1080 watch?theme=dark
watch-1920full-dark 1920 2400 watch?theme=dark
watch-1280full-dark 1280 2800 watch?theme=dark
watch-768full-dark 768 5000 watch?theme=dark
watch-1440-classic 1440 2600 watch?theme=dark&layout=classic
watch-1440-simple 1440 2600 watch?theme=dark&layout=simple
watch-1440-standard 1440 2600 watch?theme=dark&layout=standard
watch-390-classic 390 844 watch?theme=dark&layout=classic
watch-390-simple 390 844 watch?theme=dark&layout=simple
watch-1440-nolayout 1440 2600 watch?theme=dark&layout=none
watch-1920-mockup 1920 2400 watch?theme=dark&layout=mockup
watch-1920-live 1920 2800 watch?theme=dark&layout=live
watch-1920-command 1920 2600 watch?theme=dark&layout=command
watch-1920-edit 1920 2400 watch?theme=dark&edit=1
watch-1920-editdrag 1920 2400 watch?theme=dark&edit=1&drag=map:1:5
watch-1280-edit 1280 2400 watch?theme=dark&edit=1
cockpit-orders-1366 1366 768 watch?theme=dark&layout=cockpit-orders&needs=2
cockpit-orders-1440 1440 900 watch?theme=dark&layout=cockpit-orders&needs=2
cockpit-orders-1920 1920 1080 watch?theme=dark&layout=cockpit-orders&needs=2
cockpit-orders-1280 1280 620 watch?theme=dark&layout=cockpit-orders&needs=2&cmd=eight
cockpit-orders-390 390 844 watch?theme=dark&layout=cockpit-orders&needs=2
cockpit-orders-1366-eight 1366 768 watch?theme=dark&layout=cockpit-orders&needs=3&cmd=eight
overlay-1920-dark 1920&tab=Map 160 #btn-more
inventory-390-dark 390 844 watch/inventory?theme=dark
inventory-390full-dark 390 2400 watch/inventory?theme=dark
inventory-1440-dark 1440 900 watch/inventory?theme=dark
inventory-1440-light 1440 900 watch/inventory?theme=light
gearsets-390full-dark 390 2800 watch/inventory?theme=dark
gearsets-1440-dark 1440 1500 watch/inventory?theme=dark
gearsets-drag-red-1440 1440 1500 demo/drag-red?theme=dark
gearsets-drag-green-1440 1440 1500 demo/drag-green?theme=dark
gearsets-drag-green-390 390 1400 demo/drag-green?theme=dark
gearsets-preview-1440 1440 1900 demo/preview?theme=dark
gearsets-preview-390 390 3200 demo/preview?theme=dark
policy-390-dark 390 1600 watch/policy?theme=dark
policy-1440-dark 1440 900 watch/policy?theme=dark
return-390-dark 390 1600 watch/return?theme=dark
return-1440-dark 1440 900 watch/return?theme=dark"
echo "$shots" | while read -r name w h q; do
	[ -n "$only" ] && [ "$only" != "$name" ] && continue
	url="http://127.0.0.1:$port/$q"
	prof=$(mktemp -d)
	DISPLAY=$disp timeout 40 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run \
		--user-data-dir="$prof" --force-device-scale-factor=1 --force-prefers-reduced-motion --disable-lcd-text \
		--window-size="$w,$h" --virtual-time-budget=8000 --timeout=20000 --screenshot="$out/$name.png" "$url" >/dev/null 2>&1
	DISPLAY=$disp timeout 40 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --no-first-run \
		--user-data-dir="$prof" --window-size="$w,$h" --virtual-time-budget=8000 --timeout=20000 --enable-logging=stderr --v=0 --dump-dom "$url" >"$out/$name.html" 2>"$out/$name.log"
	rm -rf "${prof:?}"
done
rc=0
"$bin/uishots" check "$out" || rc=1
if [ "$update" = 1 ]; then
	mkdir -p "$golden" && cp "$out"/*.png "$golden"/ && echo "uishots: golden set updated in $golden"
elif [ -d "$golden" ]; then
	"$bin/uishots" diff -golden "$golden" -got "$out" || rc=1
else
	echo "uishots: no golden set yet (run with --update once)"
fi
echo "uishots: pictures in $out"
exit $rc
