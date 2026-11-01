package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// The d.idle startup banner (assets/branding/final/banner*.txt, cursor concept). It is printed ONLY as the head of
// the usage text (-h, or a flag error), on stderr, never on the log or on --version: scripts parse both
// (packaging/release_test.sh, tools/soak/soak24.sh, the soak log readers). The text is a copy of the asset files; banner_test.go
// keeps them equal.
const bannerBlocks = `        ██      ▀▀          ██  ██              ██████
▄▄▀▀▀▀▀▀██      ██  ▄▄▀▀▀▀▀▀██  ██  ▄▄▀▀▀▀▀▀▄▄  ██████
██      ██      ██  ██      ██  ██  ██▀▀▀▀▀▀▀▀  ██████
  ▀▀▀▀▀▀▀▀  ▀▀  ▀▀    ▀▀▀▀▀▀▀▀  ▀▀    ▀▀▀▀▀▀▀▀  ▀▀▀▀▀▀

> unattended. offline. idle.
`

const bannerASCII = `        ##      ##          ##  ##              ######
        ##                  ##  ##              ######
  ########      ##    ########  ##    ######    ######
##      ##      ##  ##      ##  ##  ##      ##  ######
##      ##      ##  ##      ##  ##  ##########  ######
##      ##      ##  ##      ##  ##  ##          ######
  ########  ##  ##    ########  ##    ########  ######

> unattended. offline. idle.
`

// bannerText returns the half-block banner when the terminal can show it, else the plain ASCII one.
func bannerText(unicode bool) string {
	if unicode {
		return bannerBlocks
	}
	return bannerASCII
}

// bannerUnicode reports whether w is a terminal with a UTF-8 locale and not TERM=dumb.
func bannerUnicode(w interface{}, getenv func(string) string) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil || st.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	if getenv("TERM") == "dumb" {
		return false
	}
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := strings.ToLower(getenv(k)); v != "" {
			return strings.Contains(v, "utf-8") || strings.Contains(v, "utf8")
		}
	}
	return false
}

// printUsage is the flag set's Usage: banner, blank line, then the stock flag listing.
func printUsage(fs *flag.FlagSet, unicode bool) {
	w := fs.Output()
	fmt.Fprint(w, bannerText(unicode))
	fmt.Fprintf(w, "\nUsage of %s:\n", fs.Name())
	fs.PrintDefaults()
}
