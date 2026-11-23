import unittest, strips

def rows(*r): return list(r)

class StripVerdict(unittest.TestCase):
    def test_open_gap_joined_to_the_interior(self):
        g = ["##..##"] + ["#....#"] * 9
        self.assertEqual(strips.verdict(g, "N", band=2, need=10), "open")
    def test_gap_behind_a_door_is_a_pocket(self):
        g = ["##..##", "######"] + ["#....#"] * 8   # two open cells in row 0, wall rows below
        self.assertEqual(strips.verdict(g, "N", band=1, need=10), "pocket")
    def test_no_open_cell_in_the_band_is_closed(self):
        g = ["######", "######", "##..##"] + ["#....#"] * 7
        self.assertEqual(strips.verdict(g, "N", band=2, need=10), "closed")
    def test_every_side(self):
        g = ["#" * 6] + ["#....#"] * 8 + ["#" * 6]
        for s in "NSWE":
            self.assertEqual(strips.verdict(g, s, band=1, need=5), "closed", s)
        self.assertEqual(strips.verdict(g, "N", band=2, need=5), "open", "band 2 reaches the interior")
    def test_side_of(self):
        self.assertEqual(strips.side_of(0, -40, 64, 18), "N")
        self.assertEqual(strips.side_of(0, 18, 64, 18), "S")
        self.assertEqual(strips.side_of(-30, 0, 64, 18), "W")
        self.assertEqual(strips.side_of(70, 0, 64, 18), "E")
        self.assertIsNone(strips.side_of(5, 5, 64, 18))

if __name__ == "__main__":
    unittest.main()
