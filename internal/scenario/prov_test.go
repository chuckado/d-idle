package scenario

import "testing"

func TestRunLengthReached(t *testing.T) {
	if !RunLengthReached([]byte(`{"msg":"virtual run length reached","virtual":3600000000000}`)) {
		t.Fatal("line present: reached")
	}
	if RunLengthReached([]byte(`{"msg":"zone entry"}`)) {
		t.Fatal("a log without the line is a cut run")
	}
}

func TestParseBuildInfo(t *testing.T) {
	out := "d2bot: go1.27.3\n\tbuild\tvcs.revision=5b12ce5e7\n\tbuild\tvcs.modified=true\n"
	c, d := ParseBuildInfo(out)
	if c != "5b12ce5e7" || !d {
		t.Fatalf("got %q %v", c, d)
	}
	c, d = ParseBuildInfo("build\t-ldflags=x\n")
	if c != "" || d {
		t.Fatalf("no stamp must give empty, got %q %v", c, d)
	}
}

func TestPaceOf(t *testing.T) {
	if PaceOf(10, 60) != 10 || PaceOf(10, 0) != 0 {
		t.Fatal("pace")
	}
}
