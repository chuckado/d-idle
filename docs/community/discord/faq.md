# FAQ (paste into #welcome or pin in #help)

**What is it?** A Go bot that plays a Diablo II LoD 1.14d character offline and reads game state to decide what to do; d.idle is the
idle-game layer on top. See the README.

**Is it safe or allowed?** It is designed for offline single player on a copy you own. It never touches Battle.net or chat. Using
automation online is against Blizzard's rules and is not discussed here. You run it at your own risk; back up your saves.

**Which backend?** The supported one is d2d, an open-source native reimplementation of the 1.14d engine, driven over its local control
socket (container: `docs/container.md`, bot: `docs/d2d-bot.md`). The original Game.exe under Wine is a frozen reference path.

**What is proven?** The README Status section and `docs/plan/STATUS.md` are the source. We tag facts verified or unverified; if you
see a number elsewhere, trust the README.

**How do I get help?** Read `docs/quickstart.md` and `docs/troubleshooting.md`, then ask in `#help` using the pinned template.

**How do I report a bug?** `#bug-reports`, template in `templates.md`: what you ran, backend, expected, got, evidence, verified or not.

**How do I share a run?** `#show-your-runs`, template in `templates.md`: class, build, hours, deaths, how measured.

**Can I contribute?** Yes: open an issue or a pull request, read `AGENTS.md`, tag facts, no game data, Go standard library only.

**Who writes the code?** Mostly AI agents working from written briefs, reviewed by humans.
