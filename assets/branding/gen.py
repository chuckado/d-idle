#!/usr/bin/env python3
"""Generates assets/branding/concepts/* (stdlib only). Run: python3 assets/branding/gen.py
Everything is drawn from rects, polygons and strokes. The pixel font below is original to this project
(5x7 cells, x-height 5). No installed fonts are used, so the SVGs render the same everywhere."""
import math, os, sys
import xml.etree.ElementTree as ET

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "concepts")

# ---------------------------------------------------------------- pixel font
GLYPHS = {
    "d": ["....#", "....#", ".####", "#...#", "#...#", "#...#", ".####"],
    "i": ["#", ".", "#", "#", "#", "#", "#"],
    "l": ["#", "#", "#", "#", "#", "#", "#"],
    "e": [".....", ".....", ".###.", "#...#", "#####", "#....", ".####"],
    ".": [".", ".", ".", ".", ".", ".", "#"],
    "n": [".....", ".....", "####.", "#...#", "#...#", "#...#", "#...#"],
    "u": [".....", ".....", "#...#", "#...#", "#...#", "#...#", ".####"],
    "a": [".....", ".....", ".###.", "....#", ".####", "#...#", ".####"],
    "t": ["#..", "#..", "###", "#..", "#..", "#..", ".##"],
    "o": [".....", ".....", ".###.", "#...#", "#...#", "#...#", ".###."],
    "f": [".##", "#..", "###", "#..", "#..", "#..", "#.."],
    " ": ["..", "..", "..", "..", "..", "..", ".."],
}
ROWS = 7


def layout(text, gap=1):
    """-> (cells set {(col,row)}, width in cols, per-glyph column spans)."""
    cells, col, spans = set(), 0, []
    for k, ch in enumerate(text):
        g = GLYPHS[ch]
        w = len(g[0])
        assert all(len(r) == w for r in g) and len(g) == ROWS, ch
        for r, line in enumerate(g):
            for c, v in enumerate(line):
                if v == "#":
                    cells.add((col + c, r))
        spans.append((ch, col, col + w))
        col += w + (gap if k < len(text) - 1 else 0)
    return cells, col, spans


def runs(cells):
    """horizontal runs -> rects (col,row,w,h) merged vertically where identical."""
    byrow = {}
    for c, r in sorted(cells, key=lambda t: (t[1], t[0])):
        byrow.setdefault(r, []).append(c)
    out = []
    for r, cs in byrow.items():
        start = prev = cs[0]
        for c in cs[1:]:
            if c != prev + 1:
                out.append((start, r, prev - start + 1, 1))
                start = c
            prev = c
        out.append((start, r, prev - start + 1, 1))
    merged, used = [], set()
    for a in out:
        if a in used:
            continue
        h = 1
        while (a[0], a[1] + h, a[2], 1) in out:
            used.add((a[0], a[1] + h, a[2], 1))
            h += 1
        merged.append((a[0], a[1], a[2], h))
    return merged


def N(v):
    return ("%.2f" % v).rstrip("0").rstrip(".")


def path_of(cells, ox, oy, s):
    d = []
    for c, r, w, h in runs(cells):
        d.append("M%s %sh%sv%sh-%sz" % (N(ox + c * s), N(oy + r * s), N(w * s), N(h * s), N(w * s)))
    return "".join(d)


def px(cells, ox, oy, s, fill, extra=""):
    return '<path d="%s" fill="%s"%s/>' % (path_of(cells, ox, oy, s), fill, extra)


def occupancy(text):
    cells, w, spans = layout(text)
    lines = ["occupancy %r (%d cols x %d rows):" % (text, w, ROWS)]
    owner = {}
    for ch, a, b in spans:
        for c in range(a, b):
            for r in range(ROWS):
                if (c, r) in cells:
                    assert (c, r) not in owner
                    owner[(c, r)] = ch
    for r in range(ROWS):
        lines.append("".join(owner.get((c, r), ".") if (c, r) in cells else " " for c in range(w)))
    # gap columns must be empty
    for (_, _, b), (_, a2, _) in zip(spans, spans[1:]):
        for c in range(b, a2):
            assert not any((c, r) in cells for r in range(ROWS)), "glyphs touch"
    return "\n".join(lines)


# ---------------------------------------------------------------- colour
def lum(h):
    h = h.lstrip("#")
    ch = [int(h[i:i + 2], 16) / 255 for i in (0, 2, 4)]
    ch = [c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4 for c in ch]
    return 0.2126 * ch[0] + 0.7152 * ch[1] + 0.0722 * ch[2]


def cr(a, b):
    la, lb = sorted((lum(a), lum(b)), reverse=True)
    return (la + 0.05) / (lb + 0.05)


# ---------------------------------------------------------------- svg helpers
def svg(w, h, body, title):
    return ('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" role="img" '
            'aria-label="%s">\n<title>%s</title>\n%s\n</svg>\n') % (w, h, w, h, title, title, body)


def crisp(inner):
    return '<g shape-rendering="crispEdges">%s</g>' % inner


def rect(x, y, w, h, fill, extra=""):
    return '<rect x="%s" y="%s" width="%s" height="%s" fill="%s"%s/>' % (N(x), N(y), N(w), N(h), fill, extra)


TAGLINE = "unattended. offline. idle."
WORD = "d.idle"

CONCEPTS = {}


# ================================================================ 1 cursor
CUR = dict(
    name="cursor", dark="#0e0c0a", light="#f6f1e7",
    P="#e0630f", Q="#a8700a",           # logo colours: >=3:1 on both dark and light
    ember="#ff8a1f", amber="#ffc15a", ash="#ece2d4", muted="#b4a797", icon_bg="#15100c",
)

CHEV = ["##...", ".##..", "..##.", "...##", "..##.", ".##..", "##..."]


def chev_cells(ox=0):
    return {(ox + c, r) for r, line in enumerate(CHEV) for c, v in enumerate(line) if v == "#"}


def block_cells(ox, row0, w, h):
    return {(ox + c, row0 + r) for c in range(w) for r in range(h)}


def cur_mark(s, ox, oy, p, q):
    # prompt chevron + block cursor: 11 cols x 7 rows
    cells = chev_cells(0)
    cur = block_cells(7, 5, 4, 2)
    return crisp(px(cells, ox, oy, s, p) + px(cur, ox, oy, s, q))


def cur_logo():
    c = CUR
    s = 12
    pad = 1
    word, ww, _ = layout(WORD)
    cols = 11 + 3 + ww
    W, H = (cols + 2 * pad) * s, (ROWS + 2 * pad) * s
    body = cur_mark(s, pad * s, pad * s, c["P"], c["Q"])
    body += crisp(px(word - {(5 + 1 + 0, 6)}, (pad + 14) * s, pad * s, s, c["P"]) )
    # the dot is a block in the second colour
    body += crisp(px({(6, 6)}, (pad + 14) * s, pad * s, s, c["Q"]))
    return svg(W, H, body, "d.idle logo, cursor concept")


def cur_icon():
    c = CUR
    s = 32
    body = rect(0, 0, 512, 512, c["icon_bg"], ' rx="104"')
    body += crisp(px(chev_cells(0), 80, 144, s, c["ember"]) + px(block_cells(7, 5, 4, 2), 80, 144, s, c["amber"]))
    return svg(512, 512, body, "d.idle icon, cursor concept")


def cur_header(W, H):
    c = CUR
    sc = W / 1280.0
    pat = ('<defs><pattern id="cs" width="8" height="8" patternUnits="userSpaceOnUse">'
           '<rect width="8" height="1" fill="#ffffff" fill-opacity="0.035"/></pattern></defs>')
    body = pat + rect(0, 0, W, H, c["dark"]) + rect(0, 0, W, H, "url(#cs)")
    ty = (H - 640 * sc) / 2
    g = ""
    # embers: deterministic specks
    specks = [(120, 120, 8, .5), (200, 480, 8, .3), (330, 90, 8, .35), (1100, 140, 8, .5), (1180, 420, 8, .3),
              (960, 540, 8, .4), (72, 330, 8, .25), (1010, 60, 8, .25), (240, 560, 8, .5), (1220, 250, 8, .35)]
    for x, y, z, o in specks:
        g += rect(x, y, z, z, c["ember"], ' fill-opacity="%s"' % o)
    word, ww, _ = layout(WORD)
    s = 32
    cols = ww + 1 + 3
    x0 = (1280 - cols * s) // 2
    y0 = 170
    g += px(word - {(6, 6)}, x0, y0, s, c["ash"])
    g += px({(6, 6)}, x0, y0, s, c["ember"])
    g += px(block_cells(ww + 1, 0, 3, ROWS), x0, y0, s, c["ember"])
    tag, tw, _ = layout(TAGLINE)
    ts = 6
    g += px(tag, (1280 - tw * ts) // 2, y0 + ROWS * s + 64, ts, c["muted"])
    g += px(chev_cells(0), 64, 64, 6, c["ember"]) + px(block_cells(7, 5, 4, 2), 64, 64, 6, c["amber"])
    body += '<g transform="translate(0 %s) scale(%s)">%s</g>' % (N(ty), N(sc), crisp(g))
    return svg(W, H, body, "d.idle header, cursor concept")


# ================================================================ 2 rune
RUN = dict(
    name="rune", dark="#100d1c", light="#f5f2fa",
    P="#8a63e8", Q="#1f8f82",
    violet="#a98bff", teal="#5fd4c4", text="#e6e0f5", muted="#a79fc4", icon_bg="#161128",
)


def f(v):
    return N(v)


def rune_mark(cx, cy, k, p, q, ticks=True, sw_ring=6, sw_rune=8):
    """Original sigil, designed on a 128 box centred (64,64), scaled by k.
    ring + stem + triangular bowl (a 'd' read sideways) + a diamond for the dot."""
    def X(v):
        return cx + (v - 64) * k

    def Y(v):
        return cy + (v - 64) * k
    out = '<circle cx="%s" cy="%s" r="%s" fill="none" stroke="%s" stroke-width="%s"/>' % (
        f(X(64)), f(Y(64)), f(56 * k), p, f(sw_ring * k))
    if ticks:
        for i in range(12):
            a = math.radians(i * 30)
            r0, r1 = (42, 49) if i % 3 == 0 else (45, 49)
            x0, y0 = 64 + r0 * math.sin(a), 64 - r0 * math.cos(a)
            x1, y1 = 64 + r1 * math.sin(a), 64 - r1 * math.cos(a)
            out += '<path d="M%s %sL%s %s" stroke="%s" stroke-width="%s" fill="none"/>' % (
                f(X(x0)), f(Y(y0)), f(X(x1)), f(Y(y1)), q, f(3 * k))
    stem = "M%s %sL%s %s" % (f(X(74)), f(Y(30)), f(X(74)), f(Y(98)))
    bowl = "M%s %sL%s %sL%s %sZ" % (f(X(74)), f(Y(60)), f(X(46)), f(Y(79)), f(X(74)), f(Y(98)))
    out += '<path d="%s" stroke="%s" stroke-width="%s" fill="none" stroke-linecap="square"/>' % (stem, p, f(sw_rune * k))
    out += '<path d="%s" stroke="%s" stroke-width="%s" fill="none" stroke-linejoin="miter"/>' % (bowl, p, f(sw_rune * k))
    d = "M%s %sL%s %sL%s %sL%s %sZ" % (f(X(50)), f(Y(34)), f(X(56)), f(Y(40)), f(X(50)), f(Y(46)), f(X(44)), f(Y(40)))
    out += '<path d="%s" fill="%s"/>' % (d, q)
    return out


def rune_word(ox, oy, s, p, q, dotshape=True):
    word, ww, _ = layout(WORD)
    out = crisp(px(word - {(6, 6)}, ox, oy, s, p))
    # the dot as a diamond in the second colour
    dx, dy = ox + 6 * s, oy + 6 * s
    out += '<path d="M%s %sL%s %sL%s %sL%s %sZ" fill="%s"/>' % (
        f(dx + s / 2), f(dy - s * .25), f(dx + s * 1.25), f(dy + s / 2), f(dx + s / 2), f(dy + s * 1.25), f(dx - s * .25), f(dy + s / 2), q)
    return out, ww


def rune_logo():
    c = RUN
    s = 10
    ring = 128
    word_svg, ww = rune_word(ring + 24 + 4, (ring - ROWS * s) // 2, s, c["P"], c["Q"])
    W, H = ring + 24 + 8 + ww * s, ring
    body = rune_mark(64, 64, 1.0, c["P"], c["Q"]) + word_svg
    return svg(W, H, body, "d.idle logo, rune concept")


def rune_icon():
    c = RUN
    body = rect(0, 0, 512, 512, c["icon_bg"], ' rx="104"')
    body += rune_mark(256, 256, 3.25, c["violet"], c["teal"], ticks=False, sw_ring=6, sw_rune=8)
    return svg(512, 512, body, "d.idle icon, rune concept")


def rune_header(W, H):
    c = RUN
    sc = W / 1280.0
    pat = ('<defs><pattern id="rs" width="64" height="64" patternUnits="userSpaceOnUse">'
           '<path d="M32 0V64M0 32H64" stroke="#ffffff" stroke-opacity="0.04" stroke-width="1" fill="none"/></pattern></defs>')
    body = pat + rect(0, 0, W, H, c["dark"]) + rect(0, 0, W, H, "url(#rs)")
    ty = (H - 640 * sc) / 2
    g = '<circle cx="640" cy="320" r="300" fill="none" stroke="%s" stroke-opacity="0.12" stroke-width="2"/>' % c["violet"]
    g += '<circle cx="640" cy="320" r="244" fill="none" stroke="%s" stroke-opacity="0.08" stroke-width="2"/>' % c["teal"]
    s = 14
    ring = 3.0 * 128
    ww = layout(WORD)[1]
    total = ring + 40 + ww * s
    x0 = (1280 - total) / 2
    g += rune_mark(x0 + ring / 2, 300, 3.0, c["violet"], c["teal"])
    ws, _ = rune_word(x0 + ring + 40, 300 - ROWS * s / 2, s, c["text"], c["teal"])
    g += ws
    tag, tw, _ = layout(TAGLINE)
    ts = 4
    g += crisp(px(tag, x0 + ring + 40 + 2, 300 + ROWS * s / 2 + 34, ts, c["muted"]))
    body += '<g transform="translate(0 %s) scale(%s)">%s</g>' % (N(ty), N(sc), g)
    return svg(W, H, body, "d.idle header, rune concept")


# ================================================================ 3 tick
TIK = dict(
    name="tick", dark="#0a111d", light="#f1f6fa",
    P="#1c8fb4", Q="#3d63c4",
    cyan="#4fd0f0", glow="#b8f4ff", robe="#4d86e8", face="#f2d4b0", text="#e2ecf8", muted="#9fb0c8", icon_bg="#0f1a2c",
)
HERO = ["....G..", "..H..S.", ".HHH.S.", "..F..S.", ".HHHHS.", ".HHH.S.", ".HHH.S.", "HHHHHS.", ".H.H.S."]


def hero(ox, oy, s, robe, face, staff, glow, halo=True):
    cells = {k: set() for k in "HFSG"}
    for r, line in enumerate(HERO):
        for c, v in enumerate(line):
            if v in cells:
                cells[v].add((c, r))
    out = ""
    if halo:
        out += rect(ox + 3 * s, oy - s, 3 * s, 3 * s, glow, ' fill-opacity="0.35"')
    out += px(cells["H"], ox, oy, s, robe) + px(cells["F"], ox, oy, s, face) + px(cells["S"], ox, oy, s, staff) + px(cells["G"], ox, oy, s, glow)
    return out


def tile(k, ox, oy, top, side_l, side_r, line):
    """iso tile in a 128 box: diamond centre (64,92); coordinates integer when k is integer."""
    def pt(x, y):
        return "%s %s" % (f(ox + x * k), f(oy + y * k))
    d_top = "M%sL%sL%sL%sZ" % (pt(16, 92), pt(64, 68), pt(112, 92), pt(64, 116))
    d_l = "M%sL%sL%sL%sZ" % (pt(16, 92), pt(64, 116), pt(64, 126), pt(16, 102))
    d_r = "M%sL%sL%sL%sZ" % (pt(112, 92), pt(64, 116), pt(64, 126), pt(112, 102))
    return ('<path d="%s" fill="%s"/><path d="%s" fill="%s"/><path d="%s" fill="%s" stroke="%s" stroke-width="%s" stroke-linejoin="miter"/>'
            % (d_l, side_l, d_r, side_r, d_top, top, line, f(3 * k)))


def tick_mark(ox, oy, k, p, q, hero_cols, tile_top, halo=True):
    s = 6 * k
    out = tile(k, ox, oy, tile_top, q, q, p)
    out += crisp(hero(ox + 43 * k, oy + 42 * k, s, hero_cols[0], hero_cols[1], hero_cols[2], hero_cols[3], halo))
    return out


def bar(x, y, segs, w, gap, h, on, off, filled):
    out = ""
    for i in range(segs):
        out += rect(x + i * (w + gap), y, w, h, on if i < filled else off)
    return crisp(out)


def tick_logo():
    c = TIK
    s = 10
    word, ww, _ = layout(WORD)
    mark = 128
    x0 = mark + 24
    W, H = x0 + ww * s + 8, 128
    body = tick_mark(0, 0, 1, c["P"], c["Q"], (c["P"], c["Q"], c["P"], c["Q"]), "none", halo=False)
    body += crisp(px(word, x0 + 4, 34, s, c["P"]))
    body += bar(x0 + 4, 34 + ROWS * s + 14, 6, 35, 4, 6, c["P"], c["Q"], 4)
    return svg(W, H, body, "d.idle logo, tick concept")


def tick_icon():
    c = TIK
    body = rect(0, 0, 512, 512, c["icon_bg"], ' rx="104"')
    body += tick_mark(48, 48, 3, c["cyan"], "#1d3a73", (c["robe"], c["face"], c["cyan"], c["glow"]), "#142848")
    return svg(512, 512, body, "d.idle icon, tick concept")


def tick_header(W, H):
    c = TIK
    sc = W / 1280.0
    pat = ('<defs><pattern id="ts" width="64" height="32" patternUnits="userSpaceOnUse">'
           '<path d="M0 16L32 0L64 16L32 32Z" stroke="#7fb4ff" stroke-opacity="0.07" stroke-width="1" fill="none"/></pattern></defs>')
    body = pat + rect(0, 0, W, H, c["dark"]) + rect(0, 0, W, H, "url(#ts)")
    ty = (H - 640 * sc) / 2
    s = 14
    word, ww, _ = layout(WORD)
    m = 3 * 128
    total = m + 36 + ww * s
    x0 = (1280 - total) // 2
    g = tick_mark(x0, 76, 3, c["cyan"], "#1d3a73", (c["robe"], c["face"], c["cyan"], c["glow"]), "#142848")
    wx = x0 + m + 36
    wy = 270
    g += crisp(px(word - {(6, 6)}, wx, wy, s, c["text"]) + px({(6, 6)}, wx, wy, s, c["cyan"]))
    tag, tw, _ = layout(TAGLINE)
    g += crisp(px(tag, wx + 2, wy + ROWS * s + 30, 4, c["muted"]))
    bw = (ww * s - 5 * 6) // 6
    g += bar(wx + 2, wy + ROWS * s + 30 + ROWS * 4 + 28, 6, bw, 6, 10, c["cyan"], "#1d3a73", 4)
    body += '<g transform="translate(0 %s) scale(%s)">%s</g>' % (N(ty), N(sc), g)
    return svg(W, H, body, "d.idle header, tick concept")


# ================================================================ ASCII
def halfblock(cells, w, dbl=2):
    lines = []
    for r in range(0, ROWS + 1, 2):
        line = ""
        for c in range(w):
            t = (c, r) in cells
            b = (c, r + 1) in cells
            ch = "█" if t and b else "▀" if t else "▄" if b else " "
            line += ch * dbl
        lines.append(line.rstrip())
    return lines


def word_ascii(extra=None):
    cells, w, _ = layout(WORD)
    if extra:
        cells = cells | extra(w)
        w += 4
    return halfblock(cells, w)


def pad_block(lines, n):
    wd = max(len(l) for l in lines)
    return [l.ljust(wd) for l in lines] + [" " * wd] * (n - len(lines))


ASC = {}
w1 = word_ascii(lambda w: block_cells(w + 1, 0, 3, ROWS))
ASC["cursor"] = ("\n".join(w1) + "\n\n> " + TAGLINE + "\n", "> d.idle_\n")

ring = [
    "    .-------.    ",
    "  .'    |  * '.  ",
    " /      |      \\ ",
    "|       |\\      |",
    "|       | >     |",
    " \\      |/     / ",
    "  '.    |    .'  ",
    "    '-------'    ",
]
w2 = halfblock(layout(WORD)[0], layout(WORD)[1])
blk = []
top = 2
for i in range(len(ring)):
    j = i - top
    blk.append(ring[i] + "   " + (w2[j] if 0 <= j < len(w2) else ""))
ASC["rune"] = ("\n".join(l.rstrip() for l in blk) + "\n" + " " * 22 + TAGLINE + "\n", "(|>) d.idle\n")

hero_tile = [
    "     *     ",
    "     |     ",
    "    /#\\    ",
    "  ,-'-+-'-,  ",
    " <    '    > ",
    "  '-,   ,-'  ",
    "     '-'     ",
]
hero_tile = [l.rstrip() for l in hero_tile]
wt = []
for i in range(len(hero_tile)):
    j = i - 1
    wt.append(hero_tile[i].ljust(16) + (w2[j] if 0 <= j < len(w2) else ""))
ASC["tick"] = ("\n".join(l.rstrip() for l in wt) + "\n" + " " * 16 + "[####..] " + TAGLINE + "\n", "[####..] d.idle\n")


# ================================================================ palette
def palette_md(c, items, wordP, wordQ, note):
    d, l = c["dark"], c["light"]
    out = "# %s: palette\n\n" % c["name"]
    out += note + "\n\n## Colours\n\n| Name | Hex | Use |\n|---|---|---|\n"
    for n, hx, u in items:
        out += "| %s | `%s` | %s |\n" % (n, hx, u)
    out += "\n## Contrast (WCAG 2.1 relative luminance; AA = 4.5:1 text, 3:1 large text and graphics)\n\n"
    out += "| Foreground | Background | Ratio | Result |\n|---|---|---|---|\n"
    rows = [(n, fg, bgn, bg) for n, fg, bgn, bg in [
        ("logo colour P", wordP, "dark `%s`" % d, d), ("logo colour P", wordP, "light `%s`" % l, l),
        ("logo colour Q", wordQ, "dark `%s`" % d, d), ("logo colour Q", wordQ, "light `%s`" % l, l)]]
    for n, fg, bgn, bg in rows:
        r = cr(fg, bg)
        out += "| %s `%s` | %s | %.2f | %s |\n" % (n, fg, bgn, r, "AA text" if r >= 4.5 else "AA large/graphics" if r >= 3 else "FAIL")
    for n, fg, bgn, bg in note_rows[c["name"]]:
        r = cr(fg, bg)
        out += "| %s `%s` | %s `%s` | %.2f | %s |\n" % (n, fg, bgn, bg, r, "AA text" if r >= 4.5 else "AA large/graphics" if r >= 3 else "FAIL")
    out += ("\n## Usage\n\n- logo.svg and mark.svg use only P and Q, chosen so both pass AA for large text and graphics (3:1) on the dark "
            "and the light background above; the wordmark is large display type (letters are 70 px tall at 1x).\n"
            "- header.svg and icon.svg are opaque and use the brighter dark-UI colours; they are meant for dark surfaces.\n"
            "- On dark UI (stream page `#0d1016`, status page dark theme `#14110f`) use logo.svg as is, or the brighter header colours.\n"
            "- On light surfaces (status page light theme `#f3efe6`) use logo.svg as is; never put the header colours on light.\n")
    return out


note_rows = {
    "cursor": [("ash text", CUR["ash"], "header", CUR["dark"]), ("muted tagline", CUR["muted"], "header", CUR["dark"]),
               ("ember", CUR["ember"], "header", CUR["dark"]), ("icon ember", CUR["ember"], "icon", CUR["icon_bg"])],
    "rune": [("text", RUN["text"], "header", RUN["dark"]), ("muted tagline", RUN["muted"], "header", RUN["dark"]),
             ("violet", RUN["violet"], "header", RUN["dark"]), ("icon violet", RUN["violet"], "icon", RUN["icon_bg"])],
    "tick": [("text", TIK["text"], "header", TIK["dark"]), ("muted tagline", TIK["muted"], "header", TIK["dark"]),
             ("cyan", TIK["cyan"], "header", TIK["dark"]), ("icon cyan", TIK["cyan"], "icon", TIK["icon_bg"])],
}

# ================================================================ write
def write(path, text):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as fh:
        fh.write(text)


def build():
    out = {}
    out["cursor"] = dict(logo=cur_logo(), icon=cur_icon(), h=cur_header(1280, 640), hd=cur_header(960, 540),
                         mark=svg(132, 84, cur_mark(12, 0, 0, CUR["P"], CUR["Q"]), "d.idle mark, cursor concept"))
    out["rune"] = dict(logo=rune_logo(), icon=rune_icon(), h=rune_header(1280, 640), hd=rune_header(960, 540),
                       mark=svg(128, 128, rune_mark(64, 64, 1.0, RUN["P"], RUN["Q"]), "d.idle mark, rune concept"))
    out["tick"] = dict(logo=tick_logo(), icon=tick_icon(), h=tick_header(1280, 640), hd=tick_header(960, 540),
                       mark=svg(128, 128, tick_mark(0, 0, 1, TIK["P"], TIK["Q"], (TIK["P"], TIK["Q"], TIK["P"], TIK["Q"]), "none", halo=False), "d.idle mark, tick concept"))
    pals = {
        "cursor": [("coal", CUR["dark"], "dark background"), ("paper", CUR["light"], "light background"),
                   ("ember P", CUR["P"], "logo/mark primary (both backgrounds)"), ("amber-dark Q", CUR["Q"], "logo/mark cursor (both backgrounds)"),
                   ("ember bright", CUR["ember"], "header/icon accent (dark only)"), ("amber", CUR["amber"], "icon cursor (dark only)"),
                   ("ash", CUR["ash"], "header wordmark"), ("muted", CUR["muted"], "tagline"), ("icon coal", CUR["icon_bg"], "icon tile")],
        "rune": [("night", RUN["dark"], "dark background"), ("mist", RUN["light"], "light background"),
                 ("violet P", RUN["P"], "logo/mark primary (both)"), ("teal-dark Q", RUN["Q"], "logo/mark secondary (both)"),
                 ("violet bright", RUN["violet"], "header/icon ring and rune (dark only)"), ("teal", RUN["teal"], "header/icon accents (dark only)"),
                 ("text", RUN["text"], "header wordmark"), ("muted", RUN["muted"], "tagline"), ("icon night", RUN["icon_bg"], "icon tile")],
        "tick": [("deep", TIK["dark"], "dark background"), ("frost", TIK["light"], "light background"),
                 ("tile blue P", TIK["P"], "logo/mark primary (both)"), ("royal Q", TIK["Q"], "logo/mark secondary (both)"),
                 ("cyan", TIK["cyan"], "header/icon accent (dark only)"), ("glow", TIK["glow"], "staff glow"),
                 ("robe", TIK["robe"], "hero robe (header/icon)"), ("face", TIK["face"], "hero face"),
                 ("text", TIK["text"], "header wordmark"), ("muted", TIK["muted"], "tagline"), ("icon deep", TIK["icon_bg"], "icon tile")],
    }
    P = {"cursor": CUR, "rune": RUN, "tick": TIK}
    notes = {
        "cursor": "Terminal ember: a command prompt with a block cursor waiting for input. The bot sits at the prompt for days.",
        "rune": "Arcane dark: an original sigil in a ring (a stem, a triangular bowl, a diamond dot) with twelve tick marks like a clock face.",
        "tick": "Cool and lockstep: a pixel mage with a glowing staff on an isometric tile, and a segmented progress bar that never finishes.",
    }
    for k, d in out.items():
        base = os.path.join(ROOT, k)
        write(os.path.join(base, "logo.svg"), d["logo"])
        write(os.path.join(base, "mark.svg"), d["mark"])
        write(os.path.join(base, "icon.svg"), d["icon"])
        write(os.path.join(base, "header.svg"), d["h"])
        write(os.path.join(base, "header-discord.svg"), d["hd"])
        write(os.path.join(base, "ascii.txt"), ASC[k][0] + "\nsmall:\n" + ASC[k][1])
        write(os.path.join(base, "palette.md"), palette_md(P[k], pals[k], P[k]["P"], P[k]["Q"], notes[k]))
    return out


def index_html(out):
    def inline(s, pre):
        t = s.replace(' width="', ' data-w="', 1).replace(' height="', ' data-h="', 1)
        t = t.replace('id="', 'id="%s-' % pre).replace("url(#", "url(#%s-" % pre)
        return t
    h = ['<!DOCTYPE html>\n<html lang="en"><head><meta charset="utf-8"><title>d.idle branding concepts</title>'
         '<meta name="viewport" content="width=device-width,initial-scale=1">\n<style>\n'
         'body{margin:0;font:15px/1.5 system-ui,sans-serif;background:#0d0d10;color:#e8e8ec}body.light{background:#f4f1ea;color:#1c1a17}\n'
         'header{display:flex;gap:16px;align-items:center;padding:16px 24px}h1{font-size:20px;margin:0;flex:1}button{font:inherit;padding:6px 14px;cursor:pointer}\n'
         '.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(380px,1fr));gap:24px;padding:0 24px 32px}\n'
         'section{border:1px solid #8884;border-radius:8px;padding:16px}h2{margin:0 0 4px;font-size:17px}p.n{margin:0 0 12px;opacity:.75;font-size:13px}\n'
         '.logo svg{width:100%;height:auto;max-height:120px}.icons{display:flex;gap:16px;align-items:flex-end;margin:12px 0}\n'
         '.hdr svg{width:100%;height:auto;display:block;margin:8px 0}pre{font:12px/1.15 ui-monospace,Menlo,Consolas,monospace;overflow:auto;background:#000;color:#ddd;padding:10px;border-radius:6px}\n'
         'small{opacity:.7}\n</style></head><body>\n'
         '<header><h1>d.idle branding concepts</h1><button id="t">Toggle dark / light</button></header>\n'
         '<script>document.getElementById("t").onclick=function(){document.body.classList.toggle("light")}</script>\n<div class="grid">\n']
    desc = {"cursor": "prompt chevron + block cursor; terminal ember", "rune": "original ring sigil; arcane dark",
            "tick": "pixel mage on an iso tile + progress bar; cool"}
    for k in ("cursor", "rune", "tick"):
        d = out[k]
        h.append('<section><h2>%s</h2><p class="n">%s</p>' % (k, desc[k]))
        h.append('<div class="logo">%s</div><div class="icons">' % inline(d["logo"], k + "l"))
        for sz in (128, 64, 32):
            h.append('<div><div style="width:%dpx;height:%dpx">%s</div><small>%d px</small></div>' % (sz, sz, inline(d["icon"], "%si%d" % (k, sz)), sz))
        h.append('</div><div class="hdr">%s<small>header 1280x640</small></div>' % inline(d["h"], k + "h"))
        h.append('<div class="hdr">%s<small>discord banner 960x540</small></div>' % inline(d["hd"], k + "d"))
        h.append("<pre>%s</pre><pre>%s</pre></section>\n" % (ASC[k][0].rstrip("\n"), ASC[k][1].rstrip("\n")))
    h.append("</div></body></html>\n")
    return "".join(h)


if __name__ == "__main__":
    print(occupancy(WORD))
    print(occupancy(TAGLINE))
    out = build()
    write(os.path.join(ROOT, "index.html"), index_html(out))
    # checks
    bad = 0
    for dp, _, fs in os.walk(ROOT):
        for fn in fs:
            p = os.path.join(dp, fn)
            if fn.endswith(".svg"):
                ET.parse(p)
            if os.path.getsize(p) > 100_000:
                print("TOO BIG", p)
                bad += 1
    for k, c in (("cursor", CUR), ("rune", RUN), ("tick", TIK)):
        for q in ("P", "Q"):
            for bg in ("dark", "light"):
                r = cr(c[q], c[bg])
                print(k, q, bg, "%.2f" % r, "" if r >= 3 else "<3 FAIL")
                bad += r < 3
    import final_gen
    final_gen.main()
    sys.exit(1 if bad else 0)
