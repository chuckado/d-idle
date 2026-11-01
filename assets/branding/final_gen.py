#!/usr/bin/env python3
"""Final asset set (cursor concept). Called by gen.py; stdlib only.
One source of truth: every asset is a Scene (a list of axis-aligned rects, an optional rounded background and
optional scanlines). The SVG files and scenes.json (read by render.py) are both written from the same Scene,
so the PNG renders match the SVGs. All coordinates are whole pixels at the asset's native size."""
import json, os

import gen
from gen import CUR, ROWS, WORD, TAGLINE, layout, runs, N, lum, cr, crisp

OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "final")

# light-background logo colours (darker than the dual-use P/Q; AA text 4.5:1 on paper)
LP, LQ = "#b34700", "#7d5200"
COAL = CUR["dark"]


class Scene:
    def __init__(self, name, w, h, title, png=None, scale=1):
        self.name, self.w, self.h, self.title, self.png, self.scale = name, w, h, title, png, scale
        self.items = []   # dicts: t=bg|r|scan
        self.probes = []  # [x, y, "#rrggbb"] opaque pixels at png scale, verified after rendering

    def bg(self, c, rx=0):
        self.items.append(dict(t="bg", c=c, rx=rx))

    def rect(self, x, y, w, h, c, o=1):
        assert 0 <= x and 0 <= y and x + w <= self.w and y + h <= self.h, (self.name, x, y, w, h)
        assert (x, y, w, h) == tuple(int(v) for v in (x, y, w, h)), (self.name, "not whole pixel", x, y, w, h)
        self.items.append(dict(t="r", x=int(x), y=int(y), w=int(w), h=int(h), c=c, o=o))

    def cells(self, cs, ox, oy, s, c, o=1):
        for col, row, w, h in runs(cs):
            self.rect(ox + col * s, oy + row * s, w * s, h * s, c, o)

    def scan(self, step, c="#ffffff", o=0.035):
        self.items.append(dict(t="scan", step=step, h=1, c=c, o=o))

    def probe(self, x, y, c):
        self.probes.append([int(x * self.scale), int(y * self.scale), c])

    # ---- SVG
    def svg(self):
        parts, group, last = [], [], None

        def flush():
            nonlocal group, last
            if group:
                d = "".join("M%s %sh%sv%sh-%sz" % (N(x), N(y), N(w), N(h), N(w)) for x, y, w, h in group)
                fill, o = last
                parts.append('<path d="%s" fill="%s"%s/>' % (d, fill, "" if o == 1 else ' fill-opacity="%s"' % o))
            group, last = [], None
        for it in self.items:
            if it["t"] == "bg":
                flush()
                parts.append(gen.rect(0, 0, self.w, self.h, it["c"], ' rx="%d"' % it["rx"] if it["rx"] else ""))
            elif it["t"] == "scan":
                flush()
                parts.append('<defs><pattern id="cs" width="8" height="%d" patternUnits="userSpaceOnUse">'
                             '<rect width="8" height="1" fill="%s" fill-opacity="%s"/></pattern></defs>' % (it["step"], it["c"], it["o"]))
                parts.append('<rect width="%d" height="%d" fill="url(#cs)"/>' % (self.w, self.h))
            else:
                key = (it["c"], it["o"])
                if key != last:
                    flush()
                    last = key
                group.append((it["x"], it["y"], it["w"], it["h"]))
        flush()
        body = "".join(p if p.startswith("<rect") or p.startswith("<defs") else crisp(p) for p in parts)
        return gen.svg(self.w, self.h, body, self.title)

    def json(self):
        return dict(w=self.w, h=self.h, scale=self.scale, png=self.png, items=self.items, probes=self.probes)


def chev():
    return gen.chev_cells(0)


def cursor():
    return gen.block_cells(7, 5, 4, 2)


def word_cells():
    cs, ww, _ = layout(WORD)
    return cs - {(6, 6)}, {(6, 6)}, ww


def mark_scene(name, s, c1, c2, pad=0, title="d.idle mark", png=None, scale=1):
    sc = Scene(name, (11 + 2 * pad) * s, (7 + 2 * pad) * s, title, png, scale)
    sc.cells(chev(), pad * s, pad * s, s, c1)
    sc.cells(cursor(), pad * s, pad * s, s, c2)
    return sc


def logo_scene(name, p, q, title, png=None, scale=1):
    s, pad = 12, 1
    word, dot, ww = word_cells()
    cols = 11 + 3 + ww
    sc = Scene(name, (cols + 2 * pad) * s, (ROWS + 2 * pad) * s, title, png, scale)
    sc.cells(chev(), pad * s, pad * s, s, p)
    sc.cells(cursor(), pad * s, pad * s, s, q)
    ox = (pad + 14) * s
    sc.cells(word, ox, pad * s, s, p)
    sc.cells(dot, ox, pad * s, s, q)
    return sc


def wordmark_scene(name, s, title, png=None):
    """tight spaces: wordmark + block cursor only, 4 px cells -> 108 x 28."""
    word, dot, ww = word_cells()
    cols = ww + 1 + 3
    sc = Scene(name, cols * s, ROWS * s, title, png)
    sc.cells(word, 0, 0, s, CUR["P"])
    sc.cells(dot, 0, 0, s, CUR["Q"])
    sc.cells(gen.block_cells(ww + 1, 0, 3, ROWS), 0, 0, s, CUR["Q"])
    return sc


def icon_scene(size, name=None, png=True):
    """hinted: 16 cells across, whole-pixel cells at every size (s = size/16, floored at 1)."""
    s = max(size // 16, 1)
    ox, oy = (size - 11 * s) // 2, (size - 7 * s) // 2
    rx = round(size * 104 / 512)
    sc = Scene(name or "icon-%d" % size, size, size, "d.idle icon", "icon-%d.png" % size if png else None)
    sc.bg(CUR["icon_bg"], rx)
    sc.cells(chev(), ox, oy, s, CUR["ember"])
    sc.cells(cursor(), ox, oy, s, CUR["amber"])
    # probes: middle of the chevron's top-left cell and of the cursor block
    sc.probe(ox + s // 2, oy + s // 2, CUR["ember"])
    sc.probe(ox + 8 * s + s // 2, oy + 5 * s + s // 2, CUR["amber"])
    return sc


def header_scene(name, W, H, s, ts, ms, png):
    """native integer layout (no scale transform): s wordmark cell, ts tagline cell, ms corner-mark cell."""
    c = CUR
    sc = Scene(name, W, H, "d.idle header", png)
    sc.bg(c["dark"])
    k = W / 1280.0
    step = 8 if W == 1280 else 6
    sc.scan(step)
    word, dot, ww = word_cells()
    cols = ww + 1 + 3
    tag, tw, _ = layout(TAGLINE)
    block = ROWS * s + 2 * s + ROWS * ts
    x0, y0 = (W - cols * s) // 2, (H - block) // 2
    bbox = (x0, y0, x0 + cols * s, y0 + block)
    # embers: deterministic specks, checked not to touch the lockup or the corner mark
    z = 8 if W == 1280 else 6
    specks = [(120, 120, .5), (200, 480, .3), (330, 90, .35), (1100, 140, .5), (1180, 420, .3),
              (960, 540, .4), (72, 330, .25), (1010, 60, .25), (240, 560, .5), (1220, 250, .35)]
    mx = my = 64 if W == 1280 else 48
    mark_box = (mx, my, mx + 11 * ms, my + 7 * ms)
    for x, y, o in specks:
        x, y = round(x * k), round(y * k)
        box = (x, y, x + z, y + z)
        for b in (bbox, mark_box):
            assert box[2] <= b[0] or box[0] >= b[2] or box[3] <= b[1] or box[1] >= b[3], (name, "speck overlaps", box, b)
        assert box[2] <= W and box[3] <= H
        sc.rect(x, y, z, z, c["ember"], o)
    sc.cells(word, x0, y0, s, c["ash"])
    sc.cells(dot, x0, y0, s, c["ember"])
    sc.cells(gen.block_cells(ww + 1, 0, 3, ROWS), x0, y0, s, c["ember"])
    sc.cells(tag, (W - tw * ts) // 2, y0 + ROWS * s + 2 * s, ts, c["muted"])
    sc.cells(chev(), mx, my, ms, c["ember"])
    sc.cells(cursor(), mx, my, ms, c["amber"])
    # probes: opaque pixels away from scanlines (y % step != 0) and specks
    for px_, py_, col in ((5, 3, c["dark"]), (x0 + 4 * s + s // 2, y0 + s // 2, c["ash"]),
                          (x0 + (ww + 1) * s + s, y0 + 3 * s, c["ember"]), (mx + ms // 2, my + ms // 2, c["ember"])):
        assert py_ % step != 0, (name, py_)
        sc.probe(px_, py_, col)
    return sc


# ---------------------------------------------------------------- banner
def banner():
    """big: half-block art from the pixel grid (2 chars per pixel column, 2 pixel rows per text line);
    ascii: '#' per pixel, 2 chars wide, 7 lines. Both <= 72 columns."""
    word, dot, ww = word_cells()
    cs = set(word) | set(dot) | {(c + ww + 1, r) for c in range(3) for r in range(ROWS)}
    cols = ww + 1 + 3
    ascii_lines = ["".join("##" if (c, r) in cs else "  " for c in range(cols)).rstrip() for r in range(ROWS)]
    big = []
    for r in range(0, ROWS, 2):
        line = ""
        for c in range(cols):
            top, bot = (c, r) in cs, (c, r + 1) in cs
            ch = "█" if top and bot else "▀" if top else "▄" if bot else " "
            line += ch * 2
        big.append(line.rstrip())
    tag = "> " + TAGLINE
    return "\n".join(big) + "\n\n" + tag + "\n", "\n".join(ascii_lines) + "\n\n" + tag + "\n"


def write(path, text):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(text)


def main():
    scenes = []
    logo = logo_scene("logo", CUR["P"], CUR["Q"], "d.idle logo", "logo.png", 2)
    logo.probe(1 * 12 + 6, 1 * 12 + 6, CUR["P"])  # chevron top-left cell: (12+6, 12+6) at 1x -> scaled in probe
    scenes += [logo,
               logo_scene("logo-light", LP, LQ, "d.idle logo, light background", "logo-light.png", 2),
               mark_scene("mark", 12, CUR["P"], CUR["Q"], title="d.idle mark"),
               mark_scene("mark-mono", 12, COAL, COAL, title="d.idle mark, monochrome"),
               wordmark_scene("wordmark", 4, "d.idle wordmark (small)", "wordmark.png")]
    scenes.append(icon_scene(512, "icon"))
    scenes[-1].png = "icon-512.png"
    for sz in (256, 128, 64, 48, 32, 16):
        scenes.append(icon_scene(sz))
    scenes.append(header_scene("header", 1280, 640, 32, 6, 6, "header-1280x640.png"))
    scenes.append(header_scene("header-discord", 960, 540, 24, 4, 4, "header-discord-960x540.png"))

    for sc in scenes:
        if sc.name.startswith("icon-") and sc.name != "icon-512":
            continue  # PNG-only hinted sizes
        write(os.path.join(OUT, sc.name + ".svg"), sc.svg())
    # logo probes: scale-aware (probe() multiplies by scale; the logo probes above are given at 1x)
    write(os.path.join(OUT, "scenes.json"), json.dumps({s.name: s.json() for s in scenes}, separators=(",", ":")) + "\n")
    b_big, b_ascii = banner()
    write(os.path.join(OUT, "banner.txt"), b_big)
    write(os.path.join(OUT, "banner-ascii.txt"), b_ascii)
    write(os.path.join(OUT, "banner-line.txt"), "> d.idle_\nd.idle: unattended. offline. idle.\n")
    # contrast report
    rows = []
    for fg, bg, what in ((CUR["P"], CUR["dark"], "logo P on dark"), (CUR["Q"], CUR["dark"], "logo Q on dark"),
                         (CUR["P"], CUR["light"], "logo P on light"), (CUR["Q"], CUR["light"], "logo Q on light"),
                         (LP, CUR["light"], "light logo P on paper"), (LQ, CUR["light"], "light logo Q on paper"),
                         (CUR["ash"], CUR["dark"], "header wordmark on coal"), (CUR["muted"], CUR["dark"], "tagline on coal"),
                         (CUR["ember"], CUR["icon_bg"], "icon chevron on tile"), (CUR["amber"], CUR["icon_bg"], "icon cursor on tile"),
                         (COAL, CUR["light"], "mono mark on paper")):
        r = cr(fg, bg)
        rows.append("%-26s %s on %s  %.2f" % (what, fg, bg, r))
        assert r >= 3, what
    print("\n".join(rows))
    return scenes
