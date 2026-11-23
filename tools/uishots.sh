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
#   tools/uishots.sh --builds [--out DIR] [--only GLOB]   pictures of the Build tab (card buildcards, builds-NAME-WIDTH.png: the card list and one opened detail, 390 px and 1366 px wide, full height), fixture Paladin, no real hero data
#   tools/uishots.sh --skills [--out DIR] [--sizes 'WxH ...'] [--tag NAME]   pictures of the Skills page (/watch/skills, fake trees) at a few sizes
#   tools/uishots.sh --pics [--out DIR] [--only GLOB]   pictures of the cockpit (strip and Needs you variants) at the eight fit sizes, in the fit harness
#   tools/uishots.sh --nav [--out DIR] [--only GLOB] [--sizes 'WxH ...']   pictures of every page of the shared navigation (card uinav2, nav-NAME-WxH.png: the bottom bar at 390x844, the band at 1366x768), in the fit harness
#   tools/uishots.sh --fit [--out DIR]                       the cockpit fit check (tools/uishots/fit.go), on the builder:
#     every cockpit page at the supported sizes loads in an iframe of exactly that size; the page must not scroll, no slot
#     may overflow, no dead space; the deliberately overflowing fixture must FAIL. No pictures, no golden set.
#   tools/uishots.sh --map [--out DIR] [--sizes 'WxH ...'] [--only VARIANT]   the Map-tab check (card mapphone, tools/uishots/mapphone.js): the Area map of the cockpit in the senses states legacy, fresh, walked and waiting;
#     a picture map-VARIANT-WxH.png and a MAPPHONE PASS|FAIL line per case; exit 1 when one fails
# Exit: 0 ok, 1 a check failed, 2 a tool is missing. Every process has a hard limit; one run per port at a time (flock).
# UISHOTS_PORT and UISHOTS_DISPLAY pick another port and private display; UISHOTS_GOLDEN another golden directory.
set -u
update=0 out=${TMPDIR:-/tmp}/uishots-$(id -u)/run only= fit=0 cockpit=0 char=0 map=0 skills=0 charu= extra= sizes_only= theme=dark fx= tag= pics=0 pq= cs=0 vt=0 nav=0 builds=0
while [ $# -gt 0 ]; do
	case $1 in
	--update) update=1; shift ;;
	--out) out=$2; shift 2 ;;
	--only) only=$2; shift 2 ;;
	--fit) fit=1; shift ;;
	--cockpit) cockpit=1; shift ;;
	--char) char=1; shift ;;
	--map) map=1; shift ;;
	--pics) pics=1; shift ;;
	--cs) cs=1; shift ;;
	--vt) vt=1; shift ;;
	--nav) nav=1; shift ;;
	--builds) builds=1; shift ;;
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
. "$(dirname "$0")/lib/heavy-guard.sh" && heavy_guard uishots || exit $?   # memory guard: docs/process/worker-capacity.md
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
# The pages of the shared navigation (card uinav2; web/nav.js): name, query, optional click and hit. --fit judges them (nav-NAME: five
# destinations, one aria-current, the bar clear of the content, the More sheet inside the window) and --nav draws them.
# settings-rules and settings-alerts (card uisettings): the Settings sheet open over the Skills page (?settings=SECTION opens it as More > Settings does).
navpages="nav-live watch?first=0&theme=dark&needs=0
nav-gear watch/inventory?theme=dark
nav-skills watch/skills?theme=dark
nav-points watch/points?theme=dark
nav-build watch/points?theme=dark&tab=build
nav-vault watch/vault?theme=dark&vault=pop
nav-characters watch/characters?theme=dark&chars=3
nav-return watch/return?theme=dark
nav-policy watch/policy?theme=dark
nav-hall watch/hall?theme=dark
nav-light-skills watch/skills?theme=light
nav-more-live watch?first=0&theme=dark&needs=0 #btn-more #menu-more,#btn-more
nav-more-gear watch/inventory?theme=dark #btn-more #menu-more,#btn-more
nav-more-skills watch/skills?theme=dark #btn-more #menu-more,#btn-more
nav-more-vault watch/vault?theme=dark&vault=pop #btn-more #menu-more,#btn-more
nav-more-characters watch/characters?theme=dark&chars=3 #btn-more #menu-more,#btn-more
settings-rules watch/skills?theme=dark&settings=rules ?.d-set-h .d-sheet-ok,.d-set-sec
settings-alerts watch/skills?theme=dark&settings=alerts ?.d-set-h .d-sheet-ok,.d-set-sec
navdefect-covers watch/skills?theme=dark&defect=nav-covers
navdefect-small watch/skills?theme=dark&defect=nav-small
navdefect-sheet watch/skills?theme=dark&defect=nav-sheet #btn-more #menu-more,#btn-more
navdefect-wide watch/skills?theme=dark&defect=nav-wide"
# The Build tab (card buildcards; web/buildcards.js): name, query, click, hit. --fit judges them (builds-NAME: one column on a phone, a grid
# above, no scroll box, 44 px buttons, a text equivalent on every star row) and --builds draws them.
buildspages="builds-list watch/points?theme=dark&tab=build
builds-detail watch/points?theme=dark&tab=build .bc-details .bc-choose
builds-choose watch/points?theme=dark&tab=build .bc-details|.bc-choose .d-sheet-ok,.d-sheet-adv-btn
builds-choose-adv watch/points?theme=dark&tab=build .bc-details|.bc-choose|.d-sheet-adv-btn .d-sheet-ok,.d-sheet-opt"
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
if [ "$map" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/map-*
	bad=0
	click=$(printf %s '.d-mapswitch [data-view="1"]' | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g; s/"/%22/g; s/ /%20/g')
	for fx in legacy fresh walked waiting; do
		[ -n "$only" ] && [ "$only" != "$fx" ] && continue
		for sz in ${sizes_only:-390x844 1366x768}; do
			w=${sz%x*} h=${sz#*x}
			url="http://127.0.0.1:$port/__mapphone?u=/watch%3Flayout=cockpit%26theme=dark%26tab=Map&w=$w&h=$h&fx=$fx&click=$click"
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --dump-dom "$url" >"$out/map-$fx-$sz.html" 2>/dev/null
			rm -rf "${prof:?}"; prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/map-$fx-$sz.png" "$url" >/dev/null 2>&1
			rm -rf "${prof:?}"
			line=$(grep -o 'MAPPHONE [A-Z]* [^<]*' "$out/map-$fx-$sz.html" | head -1)
			[ -n "$line" ] || line="MAPPHONE FAIL $fx $sz: no result in the page"
			echo "$line"
			case $line in "MAPPHONE PASS"*) ;; *) bad=1 ;; esac
		done
	done
	exit $bad
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
if [ "$nav" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/nav-*.png
	echo "$navpages" | while read -r pname pq pclick phit; do
		case $pname in navdefect-*) continue ;; esac
		[ -n "$only" ] && case $pname in $only) ;; *) continue ;; esac
		for sz in ${sizes_only:-390x844 1366x768}; do
			w=${sz%x*} h=${sz#*x}
			ck=; [ -n "$pclick" ] && ck="&click=$(printf %s "$pclick" | sed 's/_/ /g; s/#/%23/g')"
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/$pname-$sz.png" "http://127.0.0.1:$port/__fit?u=/$(printf %s "$pq" | sed 's/&/%26/g; s/?/%3F/g')&w=$w&h=$h&flow=1$ck" >/dev/null 2>&1
			rm -rf "${prof:?}"
		done
	done
	ls "$out"/nav-*.png | wc -l
	exit 0
fi
if [ "$builds" = 1 ]; then
	mkdir -p "$out" && rm -f "$out"/builds-*.png
	echo "$buildspages" | while read -r pname pq pclick phit; do
		[ -n "$only" ] && case $pname in $only) ;; *) continue ;; esac
		for sz in ${sizes_only:-390x2400 1366x1500}; do
			w=${sz%x*} h=${sz#*x}
			ck=; [ -n "$pclick" ] && ck="&click=$(printf %s "$pclick" | sed 's/_/ /g; s/#/%23/g')"
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$h" --virtual-time-budget=25000 --timeout=40000 --screenshot="$out/$pname-$w.png" "http://127.0.0.1:$port/__fit?u=/$(printf %s "$pq" | sed 's/&/%26/g; s/?/%3F/g')&w=$w&h=$h&flow=1$ck" >/dev/null 2>&1
			rm -rf "${prof:?}"
		done
	done
	ls "$out"/builds-*.png
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
tip vault=pop&theme=dark hover=.vt-cell:nth-child(3)
tipu vault=pop&theme=dark hover=.vt-cell:has([data-tag=unid])
tipl vault=pop&theme=light hover=.vt-cell:nth-child(2)
bigdeep vault=big&theme=dark scroll=0.55
bigend vault=big&theme=dark scroll=1
eth vault=eth&theme=dark
ethl vault=eth&theme=light
ethshort vault=eth-short&theme=dark
ethowned vault=eth-owned&theme=dark
ethnofit vault=eth-nofit&theme=dark
ethmax vault=eth-max&theme=dark
ethsheet vault=eth&theme=dark [data-key=upgrade]
ethsheetl vault=eth&theme=light [data-key=upgrade]
ethrefund vault=eth-owned&theme=dark [data-key=downgrade]
keeping vault=keeping&theme=dark
keepingl vault=keeping&theme=light
keepfull vault=keepfull&theme=dark
keeplist vault=keeping&theme=dark [data-key=view-toggle]
keeplistl vault=keeping&theme=light [data-key=view-toggle]
kept vault=keeping&theme=dark ?[data-key=filters-toggle]|[data-key=chip-kept]
keeppress vault=keeppress&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=keep]
keepdead vault=keepdead&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=keep]
revealing vault=revealing&theme=dark
revealingl vault=revealing&theme=light
revealpress vault=revealpress&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=reveal]
revealnone vault=revealnone&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=reveal]
revealcain vault=revealcain&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=reveal]
sellsheet vault=revealpress&theme=dark [data-key=select-toggle]|.vt-cell|[data-key=sell]
sellall vault=sellall&theme=dark
sellallarm vault=sellall&theme=dark [data-key=sell-all]
sellallfull vault=sellall-full&theme=dark [data-key=sell-all]" | while read -r pname pq pclick; do
		case $pname in $only) ;; *) [ -n "$only" ] && continue ;; esac
		for sz in ${sizes_only:-1914x930 1366x768 1280x620 412x780 360x640}; do
			w=${sz%x*} h=${sz#*x}
			ck=
			case $pclick in
			hover=* | scroll=*) ck="&$(printf %s "$pclick" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g')" ;;
			?*) ck="&click=$(printf %s "$pclick" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g; s/|/%7C/g')" ;;
			esac
			prof=$(mktemp -d)
			DISPLAY=$disp timeout 90 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$prof" --force-device-scale-factor=1 \
				--window-size="$w,$((h + 40))" --virtual-time-budget=${UISHOTS_VT_BUDGET:-25000} --timeout=40000 --screenshot="$out/vt-$pname-$sz.png" "http://127.0.0.1:$port/__fit?u=/watch/vault%3F$(printf %s "$pq" | sed 's/&/%26/g')&w=$w&h=$h$ck" >/dev/null 2>&1
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
quest-runlong first=0&theme=dark&needs=0&quests=runlong
quest-runlong-light first=0&theme=light&needs=0&quests=runlong
quest-runlong-pal first=0&theme=dark&needs=0&quests=runlong&pal=1&pfdebug=1
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
quest-countdown first=0&theme=dark&needs=0&quests=countdown
quest-countdown-sorceress first=0&theme=dark&needs=0&quests=countdown&class=sorceress
quest-away first=0&theme=dark&needs=0&quests=away
quest-ether-offer first=0&theme=dark&needs=0&quests=ether
quest-ether-paid first=0&theme=dark&needs=0&quests=paid
quest-ether-offer-light first=0&theme=light&needs=0&quests=ether
quest-ether-offer-sorceress first=0&theme=dark&needs=0&quests=ether&class=sorceress
quest-away-light first=0&theme=light&needs=0&quests=away
quest-phone-tab-away first=0&theme=dark&needs=0&quests=away&tab=Quests
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
quest-sheet-countdown first=0&theme=dark&needs=0&quests=countdown #btn-more|[data-drawer-open=quest-board]
quest-sheet-away first=0&theme=dark&needs=0&quests=away #btn-more|[data-drawer-open=quest-board]
quest-sheet-ether first=0&theme=dark&needs=0&quests=ether #btn-more|[data-drawer-open=quest-board]
quest-sheet-ether-paid first=0&theme=dark&needs=0&quests=paid #btn-more|[data-drawer-open=quest-board]
quest-sheet-away-sorceress first=0&theme=dark&needs=0&quests=away&class=sorceress #btn-more|[data-drawer-open=quest-board]
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
default-ether watch?first=0&theme=dark&needs=0&ether=1
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
default-labels watch?first=0&theme=dark&needs=2&labels=on
default-labels-light watch?first=0&theme=light&needs=2&labels=off
default-labels-menu watch?first=0&theme=dark&needs=2&labels=on #btn-more
default-question3 watch?first=0&theme=dark&needs=1&question=3
default-question5-light watch?first=0&theme=light&needs=2&question=5
quest-offer watch?first=0&theme=dark&needs=0&quests=offer
quest-running watch?first=0&theme=dark&needs=0&quests=running
quest-runlong watch?first=0&theme=dark&needs=0&quests=runlong
quest-runlong-light watch?first=0&theme=light&needs=0&quests=runlong
quest-runlong-pal watch?first=0&theme=dark&needs=0&quests=runlong&pal=1
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
quest-countdown watch?first=0&theme=dark&needs=0&quests=countdown
quest-countdown-sorceress watch?first=0&theme=dark&needs=0&quests=countdown&class=sorceress
quest-away watch?first=0&theme=dark&needs=0&quests=away
quest-away-light watch?first=0&theme=light&needs=0&quests=away
quest-phone-tab-away watch?first=0&theme=dark&needs=0&quests=away&tab=Quests
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
quest-sheet-countdown watch?first=0&theme=dark&needs=0&quests=countdown #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-away watch?first=0&theme=dark&needs=0&quests=away #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-away-sorceress watch?first=0&theme=dark&needs=0&quests=away&class=sorceress #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
quest-sheet-sorceress watch?first=0&theme=dark&needs=0&quests=offer&class=sorceress #btn-more|[data-drawer-open=quest-board] .d-drawer:not([hidden]),.d-drawer:not([hidden])_[data-key=drawer-close]
viddefect-ctl watch?first=0&theme=dark&needs=0&quests=runlong&defect=vid-ctl
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
lootrules-link watch?first=0&theme=dark&needs=0&lootrules=1 ?.d-lr-card #lootrules-sheet,[data-key=lr-mode-starter],[data-key=lr-mode-custom],#lr-apply,[data-key=lr-close]
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
chars-unsaved watch/characters?theme=dark&chars=unsaved
chars-unsaved-light watch/characters?theme=light&chars=unsaved
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
short-base watch?first=0&theme=dark&needs=2
short-base-cockpit watch?layout=cockpit&theme=dark&needs=2
more-needs watch?first=0&theme=dark&needs=2 [data-key=verb-more] .o-filter,.o-item
more-needs-light watch?first=0&theme=light&needs=2 [data-key=verb-more] .o-filter,.o-item
more-three watch?first=0&theme=dark&needs=0&cmd=eight [data-key=verb-more] .o-filter,.o-item
more-needs-cockpit watch?layout=cockpit&theme=dark&needs=2 [data-key=verb-more] .o-filter,.o-item
more-three-cockpit watch?layout=cockpit&theme=dark&needs=0&cmd=eight [data-key=verb-more] .o-filter,.o-item
moredefect-clip watch?first=0&theme=dark&needs=2&defect=mf-clip [data-key=verb-more] .o-filter,.o-item
moredefect-phone watch?first=0&theme=dark&needs=2&defect=mf-phone [data-key=verb-more] .o-filter,.o-item
menu-tall watch?first=0&theme=dark&defect=mn-tall #btn-more
menu-tall-light watch?first=0&theme=light&defect=mn-tall #btn-more
menu-tall-cockpit watch?layout=cockpit&theme=dark&defect=mn-tall #btn-more
menudefect-noscroll watch?first=0&theme=dark&defect=mn-noscroll #btn-more
menudefect-primary watch?first=0&theme=dark&defect=mn-primary #btn-more
menudefect-hint watch?first=0&theme=dark&defect=mn-hint #btn-more
menu-groups watch?first=0&theme=dark&needs=2&pause=can #btn-more|[data-group=layout]
menu-groups-light watch?first=0&theme=light&needs=2&pause=can #btn-more|[data-group=layout]
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
character-phone-menu watch/inventory?theme=dark&fx=phone&tab=pack ?#btn-more #menu-more
character-phonehero watch/inventory?theme=dark
phonedefect-label watch/inventory?theme=dark&fx=phone&tab=pack&defect=label
phonedefect-acts watch/inventory?theme=dark&fx=phone&tab=pack&sheet=1&defect=acts
phonedefect-cutchips watch/inventory?theme=dark&fx=phone&tab=pack&defect=cutchips
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
vault-select watch/vault?theme=dark&vault=pop [data-key=select-toggle]|.vt-cell|.vt-cell:nth-child(3) [data-key=select-done],[data-key=pin],[data-key=discard]
vault-sheet watch/vault?theme=dark&vault=pop [data-key=select-toggle]|.vt-cell|[data-key=discard] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-sheet-typed watch/vault?theme=dark&vault=pop [data-key=select-toggle]|[data-key=select-all]|[data-key=discard] .vt-sheet-card,[data-key=sheet-typed],[data-key=sheet-cancel]
vault-sheet-light watch/vault?theme=light&vault=pop [data-key=select-toggle]|.vt-cell|[data-key=discard] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-tip watch/vault?theme=dark&vault=pop hover=.vt-cell:nth-child(3)
vault-tip-unid watch/vault?theme=dark&vault=pop hover=.vt-cell:has([data-tag=unid])
vault-tip-light watch/vault?theme=light&vault=pop hover=.vt-cell:nth-child(2)
vault-tip-last watch/vault?theme=dark&vault=pop hover=.vt-cell:last-child
vault-big-deep watch/vault?theme=dark&vault=big scroll=0.55
vault-big-end watch/vault?theme=dark&vault=big scroll=1
vault-big-deep-light watch/vault?theme=light&vault=big scroll=0.3
vaultdefect-tip watch/vault?theme=dark&vault=pop&defect=vault-tip hover=.vt-cell:nth-child(3)
vaultdefect-unwindowed watch/vault?theme=dark&vault=big&defect=vault-unwindowed
vault-eth watch/vault?theme=dark&vault=eth
vault-eth-light watch/vault?theme=light&vault=eth
vault-eth-short watch/vault?theme=dark&vault=eth-short
vault-eth-owned watch/vault?theme=dark&vault=eth-owned
vault-eth-nofit watch/vault?theme=dark&vault=eth-nofit
vault-eth-max watch/vault?theme=dark&vault=eth-max
vault-eth-upgrade watch/vault?theme=dark&vault=eth [data-key=upgrade] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-eth-upgrade-light watch/vault?theme=light&vault=eth [data-key=upgrade] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-eth-refund watch/vault?theme=dark&vault=eth-owned [data-key=downgrade] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-keeping watch/vault?theme=dark&vault=keeping
vault-keeping-light watch/vault?theme=light&vault=keeping
vault-keeping-list watch/vault?theme=dark&vault=keeping [data-key=view-toggle] .vt-cell,[data-key=view-toggle]
vault-keepfull watch/vault?theme=dark&vault=keepfull
vault-kept watch/vault?theme=dark&vault=keeping ?[data-key=filters-toggle]|[data-key=chip-kept] [data-key=chip-kept]
vault-keep-press watch/vault?theme=dark&vault=keeppress [data-key=select-toggle]|.vt-cell|[data-key=keep] [data-key=select-toggle]
vault-keep-dead watch/vault?theme=dark&vault=keepdead [data-key=select-toggle]|.vt-cell|[data-key=keep] [data-key=select-toggle]
vault-revealing watch/vault?theme=dark&vault=revealing
vault-revealing-light watch/vault?theme=light&vault=revealing
vault-revealing-list watch/vault?theme=dark&vault=revealing [data-key=view-toggle] .vt-cell,[data-key=view-toggle]
vault-reveal-press watch/vault?theme=dark&vault=revealpress [data-key=select-toggle]|.vt-cell|[data-key=reveal] .vt-head
vault-reveal-none watch/vault?theme=dark&vault=revealnone [data-key=select-toggle]|.vt-cell|[data-key=reveal] .vt-head
vault-reveal-cain watch/vault?theme=dark&vault=revealcain [data-key=select-toggle]|.vt-cell|[data-key=reveal] .vt-head
vault-reveal-multi watch/vault?theme=dark&vault=revealcain [data-key=filters-toggle]|[data-key=unid]|[data-key=select-toggle]|.vt-cell:nth-child(1)|.vt-cell:nth-child(2)|.vt-cell:nth-child(3) .vt-head,.vt-costline
vault-reveal-multi-light watch/vault?theme=light&vault=revealcain [data-key=filters-toggle]|[data-key=unid]|[data-key=select-toggle]|.vt-cell:nth-child(1)|.vt-cell:nth-child(2)|.vt-cell:nth-child(3) .vt-head,.vt-costline
vault-reveal-dead watch/vault?theme=dark&vault=revealdead [data-key=select-toggle]|.vt-cell|[data-key=reveal] .vt-head
vault-sell-sheet watch/vault?theme=dark&vault=revealpress [data-key=select-toggle]|.vt-cell|[data-key=sell] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-sell-sheet-light watch/vault?theme=light&vault=revealpress [data-key=select-toggle]|.vt-cell|[data-key=sell] .vt-sheet-card,[data-key=sheet-go],[data-key=sheet-cancel]
vault-more-entry watch?first=0&theme=dark&needs=0 [data-key=verb-more] [data-key=more-vault]
vaultdefect-clip watch/vault?theme=dark&vault=pop&defect=vault-clip
vaultdefect-wide watch/vault?theme=dark&vault=pop&defect=vault-wide
vaultdefect-under watch/vault?theme=dark&vault=pop&defect=vault-under
vaultdefect-small watch/vault?theme=dark&vault=pop&defect=vault-small
vaultdefect-covered watch/vault?theme=dark&vault=pop&defect=vault-covered
vaultdefect-sheet watch/vault?theme=dark&vault=pop&defect=vault-sheet [data-key=select-toggle]|.vt-cell|[data-key=discard] .vt-sheet-card,[data-key=sheet-cancel]
vaultdefect-noname watch/vault?theme=dark&vault=pop&defect=vault-noname
vaultdefect-enabled watch/vault?theme=dark&vault=pop&defect=vault-enabled
vaultdefect-ethenabled watch/vault?theme=dark&vault=eth-short&defect=vault-ethenabled
$navpages
$buildspages"
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
		vault-more-entry) ;;
		navdefect-wide) szs="1366x768" ;;
		# the must-fail fixtures of the desk frame (an overflowing, nested or fake-scrolling box, a tall frame): the phone cockpit is a scrolling page by design
		overflow-fixture | overflow-nested | overflow-fakescroll | tall-frame) szs="1920x1080 1440x900 1366x768 1914x930 1904x1000 1536x730 1280x620" ;;
		navdefect-*) szs="390x844" ;;
		nav-* | settings-*) szs="390x844 1366x768" ;;
		builds-*) szs="390x844 1366x768" ;;
		vault-*) szs="$sizes 412x780 360x640" ;;
		questdefect-words) szs="1366x768" ;;
		questdefect-clip | questdefect-two) szs="1366x768 412x780" ;;
		questdefect-small) szs="412x780 360x640" ;;
		quest-phone-tab*) szs="412x780 390x844 360x640" ;;
		quest-sheet-*) szs="$sizes 412x780 360x640" ;;
		quest-*) szs="$sizes 412x780 360x640" ;; # the 360x640 phone too (the Needs you row of a waiting hand-in no longer pushes Orders over, slice S8)
		short-base*) szs="870x520 1000x560 1100x600 1280x480 1280x620" ;;
		moredefect-phone) szs="390x844 360x640" ;;
		more-* | moredefect-*) szs="870x520 1000x560 1100x600 1280x480 1280x620" ;;
		menudefect-*) szs="1280x620 412x780" ;;
		menu-tall* | menu-groups*) szs="1280x620 1366x768 900x700 412x780 360x640" ;;
		overlaydefect-two) szs="1366x768" ;; # the other tiles are not on screen on a phone: nothing to show
		overlaydefect-*) szs="1366x768 412x780" ;;
		overlay-*) szs="$sizes 412x780" ;;
		character-phone*) szs="$phonesizes" ;;
		esac
		[ -n "$sizes_only" ] && szs=$sizes_only # --sizes narrows a fit run to those sizes
		for sz in $szs; do
			w=${sz%x*} h=${sz#*x} ck=
			enc() { printf %s "$1" | sed 's/_/ /g; s/#/%23/g; s/?/%3F/g; s/|/%7C/g; s/&/%26/g'; }
			case $pclick in
			hover=* | ctx=* | expect=* | scroll=*) ck="&$(printf %s "$pclick" | sed 's/#/%23/g; s/\[/%5B/g; s/\]/%5D/g; s/|/%7C/g')" ;;
			?*) ck="&click=$(enc "$pclick")" ;;
			esac
			case $pname in nav-* | navdefect-* | builds-* | settings-*) ck="$ck&flow=1" ;; esac
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
