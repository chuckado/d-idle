#!/usr/bin/env python3
"""PNG renders for the final d.idle assets, stdlib only (zlib + struct). Renders final/scenes.json, the rect
lists that gen.py writes next to the SVGs (axis-aligned rects, opacity, scanlines, a rounded background).
Rects are anti-aliased by exact area coverage; the rounded corners by 4x4 supersampling.
Run: python3 assets/branding/gen.py && python3 assets/branding/render.py [--preview]
--preview prints an ASCII occupancy preview of the icon at 16/32/48/64 px."""
import json, math, os, struct, sys, zlib

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(HERE, "final")
LIMIT = 300_000


def rgb(h):
    h = h.lstrip("#")
    return tuple(int(h[i:i + 2], 16) for i in (0, 2, 4))


class Canvas:
    def __init__(self, w, h):
        self.w, self.h = w, h
        self.rows = [bytearray(4 * w) for _ in range(h)]  # RGBA straight alpha, transparent

    def blend(self, row, x, c, a):
        if a <= 0:
            return
        i = 4 * x
        da = row[i + 3] / 255.0
        oa = a + da * (1 - a)
        for k in range(3):
            row[i + k] = int(round((c[k] * a + row[i + k] * da * (1 - a)) / oa))
        row[i + 3] = int(round(oa * 255))

    def rect(self, x, y, w, h, c, o=1.0):
        x1, y1 = x + w, y + h
        for iy in range(max(int(math.floor(y)), 0), min(int(math.ceil(y1)), self.h)):
            vc = min(iy + 1, y1) - max(iy, y)
            if vc <= 0:
                continue
            row = self.rows[iy]
            ix0, ix1 = max(int(math.floor(x)), 0), min(int(math.ceil(x1)), self.w)
            full0, full1 = int(math.ceil(x)), int(math.floor(x1))
            a_row = vc * o
            if a_row >= 1.0 and full1 > full0:
                row[4 * full0:4 * full1] = (bytes(c) + b"\xff") * (full1 - full0)
                xs = [ix for ix in (list(range(ix0, full0)) + list(range(full1, ix1)))]
            else:
                xs = range(ix0, ix1)
            for ix in xs:
                hc = min(ix + 1, x1) - max(ix, x)
                if hc > 0:
                    self.blend(row, ix, c, a_row * hc)

    def rounded_bg(self, c, rx):
        W, H = self.w, self.h

        def inside(px, py):
            cx = rx if px < rx else W - rx if px > W - rx else None
            cy = rx if py < rx else H - rx if py > H - rx else None
            if cx is None or cy is None:
                return True
            return (px - cx) ** 2 + (py - cy) ** 2 <= rx * rx
        solid = bytes(c) + b"\xff"
        for iy in range(H):
            row = self.rows[iy]
            if rx <= iy < H - rx:
                row[:] = solid * W
                continue
            for ix in range(W):
                if rx <= ix < W - rx:
                    row[4 * ix:4 * ix + 4] = solid
                    continue
                n = sum(inside(ix + (a + .5) / 4, iy + (b + .5) / 4) for a in range(4) for b in range(4))
                if n:
                    self.blend(row, ix, c, n / 16.0)

    def scan(self, step, h, c, o):
        for iy in range(0, self.h, step):
            self.rect(0, iy, self.w, h, c, o)

    def opaque(self):
        return all(r[3::4] == b"\xff" * self.w for r in self.rows)


def render(scene):
    k = scene.get("scale", 1)
    cv = Canvas(scene["w"] * k, scene["h"] * k)
    for it in scene["items"]:
        c = rgb(it["c"])
        if it["t"] == "bg":
            cv.rounded_bg(c, it["rx"] * k)
        elif it["t"] == "scan":
            cv.scan(it["step"] * k, it["h"] * k, c, it["o"])
        else:
            cv.rect(it["x"] * k, it["y"] * k, it["w"] * k, it["h"] * k, c, it["o"])
    return cv


# ---------------------------------------------------------------- PNG / ICO
def chunk(t, d):
    return struct.pack(">I", len(d)) + t + d + struct.pack(">I", zlib.crc32(t + d) & 0xffffffff)


def png_bytes(cv):
    opaque = cv.opaque()
    ct, bpp = (2, 3) if opaque else (6, 4)
    raw, prev = bytearray(), bytes(cv.w * bpp)
    for r in cv.rows:
        if opaque:
            cur = bytes(b for i in range(cv.w) for b in r[4 * i:4 * i + 3])
        else:
            cur = bytes(r)
        raw.append(2)  # filter: Up
        raw += bytes((a - b) & 255 for a, b in zip(cur, prev))
        prev = cur
    return (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", cv.w, cv.h, 8, ct, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(bytes(raw), 9)) + chunk(b"IEND", b""))


def ico_bytes(pngs):
    """pngs: list of (size, png bytes); PNG-compressed entries (Vista+ ICO, also read by every browser)."""
    head = struct.pack("<HHH", 0, 1, len(pngs))
    off = 6 + 16 * len(pngs)
    ent, data = b"", b""
    for sz, p in pngs:
        ent += struct.pack("<BBBBHHII", sz % 256, sz % 256, 0, 0, 1, 32, len(p), off + len(data))
        data += p
    return head + ent + data


# ---------------------------------------------------------------- decoder (independent check)
def decode_png(b):
    assert b[:8] == b"\x89PNG\r\n\x1a\n", "signature"
    pos, idat, ihdr = 8, b"", None
    while pos < len(b):
        n, t = struct.unpack(">I4s", b[pos:pos + 8])
        d = b[pos + 8:pos + 8 + n]
        assert struct.unpack(">I", b[pos + 8 + n:pos + 12 + n])[0] == zlib.crc32(t + d) & 0xffffffff, "crc " + t.decode()
        if t == b"IHDR":
            ihdr = struct.unpack(">IIBBBBB", d)
        elif t == b"IDAT":
            idat += d
        pos += 12 + n
    w, h, depth, ct, _, _, _ = ihdr
    assert depth == 8 and ct in (2, 6)
    bpp = 3 if ct == 2 else 4
    raw = zlib.decompress(idat)
    assert len(raw) == h * (1 + w * bpp), "idat length"
    rows, prev, p = [], bytearray(w * bpp), 0
    for _ in range(h):
        f = raw[p]
        line = bytearray(raw[p + 1:p + 1 + w * bpp])
        p += 1 + w * bpp
        for i in range(len(line)):
            a = line[i - bpp] if i >= bpp else 0
            up = prev[i]
            c = prev[i - bpp] if i >= bpp else 0
            if f == 1:
                line[i] = (line[i] + a) & 255
            elif f == 2:
                line[i] = (line[i] + up) & 255
            elif f == 3:
                line[i] = (line[i] + (a + up) // 2) & 255
            elif f == 4:
                pa, pb, pc = abs(up - c), abs(a - c), abs(a + up - 2 * c)
                pr = a if pa <= pb and pa <= pc else up if pb <= pc else c
                line[i] = (line[i] + pr) & 255
        rows.append(line)
        prev = line
    return w, h, ct, rows


def pixel(dec, x, y):
    w, h, ct, rows = dec
    bpp = 3 if ct == 2 else 4
    r = rows[y][bpp * x:bpp * x + bpp]
    return tuple(r) + ((255,) if ct == 2 else ())


def hexs(t):
    return "#%02x%02x%02x" % t[:3]


# ---------------------------------------------------------------- main
def preview(scenes):
    """ASCII occupancy of the icon: C chevron, # cursor block, . tile, ' ' outside the rounded tile."""
    ember, amber = rgb("#ff8a1f"), rgb("#ffc15a")
    for sz in (16, 32, 48, 64):
        cv = render(scenes["icon-%d" % sz])
        print("icon %d px (C chevron, # cursor, . tile, space transparent):" % sz)
        for r in cv.rows:
            line = ""
            for i in range(sz):
                px = r[4 * i:4 * i + 4]
                if px[3] < 128:
                    line += " "
                else:
                    d_e = sum((px[k] - ember[k]) ** 2 for k in range(3))
                    d_a = sum((px[k] - amber[k]) ** 2 for k in range(3))
                    d_t = sum((px[k] - rgb("#15100c")[k]) ** 2 for k in range(3))
                    m = min(d_e, d_a, d_t)
                    line += "C" if m == d_e else "#" if m == d_a else "."
            print(line)


def main():
    with open(os.path.join(OUT, "scenes.json")) as f:
        scenes = json.load(f)
    if "--preview" in sys.argv:
        return preview(scenes)
    pngs, bad = {}, 0
    for name, sc in scenes.items():
        if not sc.get("png"):
            continue
        cv = render(sc)
        data = png_bytes(cv)
        pngs[sc["png"]] = data
        with open(os.path.join(OUT, sc["png"]), "wb") as f:
            f.write(data)
    ico = ico_bytes([(sz, pngs["icon-%d.png" % sz]) for sz in (16, 32, 48)])
    with open(os.path.join(OUT, "favicon.ico"), "wb") as f:
        f.write(ico)
    # verify: decode every PNG, compare the probes
    for name, sc in scenes.items():
        if not sc.get("png"):
            continue
        dec = decode_png(pngs[sc["png"]])
        assert (dec[0], dec[1]) == (sc["w"] * sc["scale"], sc["h"] * sc["scale"]), name
        res = []
        for x, y, want in sc["probes"]:
            got = hexs(pixel(dec, x, y))
            ok = got == want
            bad += not ok
            res.append("(%d,%d)=%s%s" % (x, y, got, "" if ok else " WANT " + want))
        print("%-26s %4dx%-4d type%d %7d B  %s" % (sc["png"], dec[0], dec[1], dec[2], len(pngs[sc["png"]]), " ".join(res)))
    # corner / alpha checks
    d512 = decode_png(pngs["icon-512.png"])
    print("icon-512 corner (0,0) alpha=%d, edge-aa (12,12) alpha=%d, centre-top (256,0) alpha=%d"
          % (pixel(d512, 0, 0)[3], pixel(d512, 12, 12)[3], pixel(d512, 256, 0)[3]))
    assert pixel(d512, 0, 0)[3] == 0 and pixel(d512, 256, 0)[3] == 255
    # ico
    n = struct.unpack("<HHH", ico[:6])
    assert n == (0, 1, 3), n
    for i in range(3):
        wd, ht, _, _, pl, bp, ln, off = struct.unpack("<BBBBHHII", ico[6 + 16 * i:22 + 16 * i])
        dec = decode_png(ico[off:off + ln])
        assert (dec[0], dec[1]) == (wd, ht)
        print("favicon.ico entry %d: %dx%d %d bpp, %d B PNG ok" % (i, wd, ht, bp, ln))
    for fn in sorted(os.listdir(OUT)):
        p = os.path.join(OUT, fn)
        if os.path.getsize(p) > LIMIT:
            print("TOO BIG", fn)
            bad += 1
    sys.exit(1 if bad else 0)


if __name__ == "__main__":
    main()
