package watch

import (
	"regexp"
	"strings"
	"testing"
)

// Landmarks, headings and accessible names of every page's static HTML (card uia11y): one h1, a main region, a language and a title,
// and no button, link or form control without a name. Names the script adds later are covered by the node tests of each page; this is the
// part that exists before any script runs.
func TestPagesLandmarksAndNames(t *testing.T) {
	live, _ := webFS.ReadFile("web/watch.html")
	vault, _ := webFS.ReadFile("web/vault.html")
	chars, _ := webFS.ReadFile("web/characters.html")
	pages := map[string]string{
		"live": string(live), "inventory": charPage, "skills": skillsPage, "points": pointsPage, "vault": string(vault), "characters": string(chars),
		"return": string(RenderReturn(ReturnData{}, false)), "policy": string(RenderPolicy(PolicyView{})), "hall": string(RenderHall(&HallView{})),
	}
	h1 := regexp.MustCompile(`(?i)<h1[\s>]`)
	main := regexp.MustCompile(`(?i)<main[\s>]|role="main"`)
	btn := regexp.MustCompile(`(?is)<(button|a)\b([^>]*)>(.*?)</(?:button|a)>`)
	tag := regexp.MustCompile(`(?s)<[^>]*>`)
	ctl := regexp.MustCompile(`(?is)<(input|select|textarea)\b([^>]*)>`)
	for name, src := range pages {
		js := map[string]string{"vault": "web/vault.js", "characters": "web/characters.js"}[name]
		if js != "" { // these two pages draw their heading and main region from script (the title and h1 come from the page's own texts)
			b, _ := webFS.ReadFile(js)
			if !strings.Contains(string(b), "el('h1'") || !strings.Contains(string(b), "el('main'") {
				t.Errorf("%s: its script no longer draws the h1 and the main region", name)
			}
			src = strings.Replace(src, "<body", "<body><h1>x</h1><main>x</main><b", 1)
		}
		if n := len(h1.FindAllString(src, -1)); n != 1 {
			t.Errorf("%s: %d h1 elements, want exactly 1", name, n)
		}
		if !main.MatchString(src) {
			t.Errorf("%s: no main region", name)
		}
		if !regexp.MustCompile(`(?i)<html[^>]*\blang=`).MatchString(src) {
			t.Errorf("%s: <html> has no lang", name)
		}
		if !regexp.MustCompile(`(?is)<title>\s*\S`).MatchString(src) {
			t.Errorf("%s: no title", name)
		}
		for _, m := range btn.FindAllStringSubmatch(src, -1) {
			attrs, inner := m[2], strings.TrimSpace(tag.ReplaceAllString(m[3], ""))
			if inner == "" && !strings.Contains(attrs, ` id="`) && !strings.Contains(attrs, "aria-label") && !strings.Contains(attrs, "aria-labelledby") && !strings.Contains(attrs, "title=") && !strings.Contains(m[3], "aria-label") {
				t.Errorf("%s: a <%s %s> has no name", name, m[1], strings.TrimSpace(attrs))
			}
		}
		for _, m := range ctl.FindAllStringSubmatch(src, -1) {
			a := m[2]
			if strings.Contains(a, `type="hidden"`) {
				continue
			}
			id := regexp.MustCompile(`\bid="([^"]+)"`).FindStringSubmatch(a)
			wrapped := regexp.MustCompile(`(?is)<label[^>]*>[^<]*<input[^>]*` + regexp.QuoteMeta(a) + `[^>]*>\s*[^<]*\S`).MatchString(src)
			labelled := wrapped || strings.Contains(a, "aria-label") || (id != nil && strings.Contains(src, `for="`+id[1]+`"`)) || strings.Contains(a, "title=")
			if !labelled {
				t.Errorf("%s: a <%s %s> has no label", name, m[1], strings.TrimSpace(a))
			}
		}
	}
}
