package scenario

import "testing"

func res(asserts ...AssertResult) Result {
	return Result{Run: Run{ID: "x", Class: "Paladin", Seed: 11}, Status: Green, Asserts: asserts}
}

// Gold 700 against 712 with the same verdict is the same run; a different quest word is not (design 2.6).
func TestAssertDigestRangeVersusDiscrete(t *testing.T) {
	g700 := res(AssertResult{Kind: AGoldBetween, OK: true, Got: "gold 700"})
	g712 := res(AssertResult{Kind: AGoldBetween, OK: true, Got: "gold 712"})
	if AssertDigest(g700) != AssertDigest(g712) {
		t.Fatal("a range fact must enter the digest as its bucket only")
	}
	q1 := res(AssertResult{Kind: AQuestBits, OK: true, Got: "0x2011"})
	q2 := res(AssertResult{Kind: AQuestBits, OK: true, Got: "0x2019"})
	if AssertDigest(q1) == AssertDigest(q2) {
		t.Fatal("two GREEN runs with different quest words are not the same run")
	}
	f := res(AssertResult{Kind: AGoldBetween, OK: false, Got: "gold 5"})
	if AssertDigest(f) == AssertDigest(g700) {
		t.Fatal("the pass or fail bucket must differ")
	}
}

func TestCheckDeterminismGatesOnAssertDigest(t *testing.T) {
	a := res(AssertResult{Kind: AQuestBits, OK: true, Got: "0x2011"})
	b := res(AssertResult{Kind: AQuestBits, OK: true, Got: "0x2019"})
	a.Digest, b.Digest = "s1", "s2"
	ok, vd, dd, _ := CheckDeterminism([]Result{a}, []Result{b})
	if ok || len(vd) != 1 || len(dd) != 0 {
		t.Fatalf("a changed discrete fact must fail hard: ok=%v vd=%v dd=%v", ok, vd, dd)
	}
	c := res(AssertResult{Kind: AGoldBetween, OK: true, Got: "gold 700"})
	d := res(AssertResult{Kind: AGoldBetween, OK: true, Got: "gold 712"})
	c.Digest, d.Digest = "s1", "s2"
	ok, vd, dd, _ = CheckDeterminism([]Result{c}, []Result{d})
	if !ok || len(vd) != 0 || len(dd) != 1 {
		t.Fatalf("a state-digest drift under equal D1 is reported, not failed: ok=%v vd=%v dd=%v", ok, vd, dd)
	}
}
