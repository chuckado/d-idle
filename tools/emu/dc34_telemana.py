"""decomp34: does a refused Teleport cost mana in game.exe?  FUN_0056f7f0 (the do wrapper; pays at 0x56fa70 via FUN_0056bfe0 only when the do
returned non-zero for a skill with no missile, 0x56fa3d) run natively with the REAL do 27 (FUN_005ca360) on sk_tele_spot's synthetic rooms.
    dc34_telemana.py game N SEED  ->  `id do x y paid`   (paid = points taken off stat 8, 256ths; do = FUN_005ca360 put the unit somewhere)
Hooked (not the claim): sk_tele_spot's set, plus sk_mana's: FUN_00620250 (no current skill), FUN_0056c160 (the mana check: passes),
FUN_00643ad0, FUN_00644140, FUN_006442a0, FUN_00625480 (stat 8 = 100 x 256), FUN_0063a400, FUN_00639df0, FUN_006272b0 (recorded).
Tags: V run for the game column; the hook list is V (the run completes)."""
import os, random, sys
sys.path.insert(0, os.environ["EMU"])
import emu, skill_tips
_BOOTED = skill_tips.boot()
emu.Emu = lambda: _BOOTED          # the Skills.txt records are loaded (sk_mana.py's boot); sk_tele_spot builds its rooms on it
import sk_tele_spot as T

RECORD, FAKE, TELE = 0x23c, 0x30000000, 54

class TM(T.Oracle):
    def __init__(self):
        super().__init__()
        import skill_tips
        e = self.e
        tables = e.r32(0x744304)
        self.skills = e.r32(tables + 0xb98)
        self.calls, self.added, self.lvl = 0, None, 1
        def cur(e):
            self.calls += 1
            return FAKE if self.calls == 2 else 0
        e.hook(0x620250, lambda e: 0, 1)
        e.stub(1, {0x56c160: 0})      # the mana check (sk_mana.py's claim): passes; the payment is what is read
        e.stub(0xFFFFFFFF, {0x643ad0: 1})
        e.hook(0x644140, lambda e: self.skills + TELE * RECORD, 0)
        e.hook(0x6442a0, lambda e: self.lvl, 3)
        e.hook(0x625480, lambda e: 100 * 256 if e.arg(1) == 8 else 0, 3)
        e.stub(0, {0x63a400: 1, 0x639df0: 2})
        e.hook(0x6272b0, self.add, 4)
        r = self.skills + TELE * RECORD
        print('# row srvst %d srvdo %d srvmissile %d flags %#x' % (e.s32(r + 0x2c) & 0xffff, e.s32(r + 0x2e) & 0xffff, e.s32(r + 0x46) & 0xffff, e.r32(r + 4)), file=sys.stderr)
        self.pays = 0
        e.hook(0x56bfe0, self.pay, 2)
    def pay(self, e):
        self.pays += 1
        return 1
    def add(self, e):
        if e.arg(1) == 8:
            v = e.arg(2); self.added = v - (1 << 32) if v & 0x80000000 else v
        return 0
    def cast(self, tele, cx, cy, tx, ty, words):
        self.load(words)
        self.e.view("levels", self.levelrow).teleport = tele
        self.here = self.rooms[(cx // T.ROOM, cy // T.ROOM)]
        self.xy, self.moved, self.line = (tx & 0xFFFFFFFF, ty & 0xFFFFFFFF), None, None
        self.calls, self.added, self.pays = 0, None, 0
        ret = self.e.call(0x56f7f0, TELE, self.lvl, 1, 1, 0, ecx=0x2000, edx=self.unit) & 0xFFFFFFFF
        return ret, self.moved, (self.pays)

def main():
    n, seed = int(sys.argv[2]), int(sys.argv[3])
    rng, o = random.Random(seed), TM()
    for cid in range(n):
        tele, cx, cy, tx, ty, words = T.make_case(rng, False)
        ret, pos, paid = o.cast(tele, cx, cy, tx, ty, words)
        print(cid, int(pos is not None), *(pos if pos else ("-", "-")), paid, "tele", tele, "ret", ret)
main()
