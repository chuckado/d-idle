package boss

import "testing"

// The Summoner is class 250 alone: the lab lists him without a name and the run never found him (V drill.act2.summoner).
// Must fail without ClassOnly. A nameless monster of another class with a name in the script stays a non-match.
func TestSummonerMatchesByClassWithoutAName(t *testing.T) {
	sc := summoner()
	if !sc.Match.Is(Mon{Class: ClassSummoner, UniqueNo: -1}) {
		t.Fatal("a nameless class 250 is not the Summoner")
	}
	if sc.Match.Is(Mon{Class: 56, UniqueNo: -1}) {
		t.Fatal("another class matched")
	}
	if (Match{Class: 156, UniqueNo: -1, Name: "Andariel"}).Is(Mon{Class: 156, UniqueNo: -1}) {
		t.Fatal("without ClassOnly a nameless monster of the class must not match")
	}
}
