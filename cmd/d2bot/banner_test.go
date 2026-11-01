package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func maxCols(s string) int {
	m := 0
	for _, l := range strings.Split(s, "\n") {
		if n := utf8.RuneCountInString(l); n > m {
			m = n
		}
	}
	return m
}

func TestBannerShape(t *testing.T) {
	if c := maxCols(bannerBlocks); c > 72 || c < 40 {
		t.Fatalf("blocks banner is %d columns", c)
	}
	if c := maxCols(bannerASCII); c > 72 || c < 40 {
		t.Fatalf("ascii banner is %d columns", c)
	}
	for _, r := range bannerASCII {
		if r > 126 || (r < 32 && r != '\n') {
			t.Fatalf("non-ASCII rune %q in the plain banner", r)
		}
	}
	if !strings.Contains(bannerBlocks, "unattended. offline. idle.") || !strings.Contains(bannerASCII, "unattended. offline. idle.") {
		t.Fatal("tagline missing")
	}
	if bannerText(true) != bannerBlocks || bannerText(false) != bannerASCII {
		t.Fatal("bannerText picks the wrong variant")
	}
}

// The Go copies must equal the asset files (skipped where the assets are absent).
func TestBannerMatchesAssets(t *testing.T) {
	for name, want := range map[string]string{"banner.txt": bannerBlocks, "banner-ascii.txt": bannerASCII} {
		b, err := os.ReadFile("../../assets/branding/final/" + name)
		if err != nil {
			t.Skip("assets not present:", err)
		}
		if string(b) != want {
			t.Errorf("%s differs from the Go copy: regenerate assets and update cmd/d2bot/banner.go", name)
		}
	}
}

func TestBannerUnicodeNeedsTerminalAndUTF8(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if bannerUnicode(&bytes.Buffer{}, env(map[string]string{"LANG": "en_US.UTF-8"})) {
		t.Fatal("a buffer is not a terminal")
	}
	f, err := os.Open(os.DevNull) // a character device
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	for _, c := range []struct {
		env  map[string]string
		want bool
	}{
		{map[string]string{"LANG": "en_US.UTF-8"}, true},
		{map[string]string{"LC_ALL": "C.utf8", "LANG": "C"}, true},
		{map[string]string{"LC_ALL": "C", "LANG": "en_US.UTF-8"}, false},
		{map[string]string{"TERM": "dumb", "LANG": "en_US.UTF-8"}, false},
		{map[string]string{}, false},
	} {
		if got := bannerUnicode(f, env(c.env)); got != c.want {
			t.Errorf("%v: got %v", c.env, got)
		}
	}
}

func TestUsagePrintsBannerThenFlags(t *testing.T) {
	var buf bytes.Buffer
	fs := flag.NewFlagSet("d2bot", flag.ContinueOnError)
	fs.SetOutput(&buf)
	fs.Bool("sim", false, "simulate")
	printUsage(fs, false)
	out := buf.String()
	if !strings.HasPrefix(out, bannerASCII) || !strings.Contains(out, "Usage of d2bot:") || !strings.Contains(out, "-sim") {
		t.Fatalf("usage:\n%s", out)
	}
}

// --version and the normal parse never print the banner (script-parsed output).
func TestVersionOutputUnchanged(t *testing.T) {
	if strings.Contains(version, "unattended") {
		t.Fatal("version must stay a bare string")
	}
}
