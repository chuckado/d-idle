package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// The support sheet's fit pages exist at all sizes and the phone, the sheet's own controls must be hittable, and the
// fixture variants carry what each page needs (a blocked preview has findings and no Send, no upload has Save only).
func TestSupportFitPages(t *testing.T) {
	b, err := os.ReadFile("../uishots.sh")
	if err != nil {
		t.Fatal(err)
	}
	sh := string(b)
	for _, p := range []string{"support-sheet ", "support-sheet-light ", "support-preview ", "support-preview-light ", "support-blocked ", "support-nosend "} {
		if !strings.Contains(sh, "\n"+p) {
			t.Errorf("no fit page %q", p)
		}
	}
	if !strings.Contains(sh, "reset-* | support-*) szs=") || !strings.Contains(sh, "412x780 360x640") {
		t.Error("the support pages do not run the phone sizes")
	}
	for _, sel := range []string{"#support-sheet", "#support-send", "#support-save-file", "[data-key=support-back]", "#support-build", "[data-key=support-cancel]"} {
		if !strings.Contains(sh, sel) {
			t.Errorf("no fit page hits %s", sel)
		}
	}
}

func TestSupportFixtureVariants(t *testing.T) {
	get := func(q string) map[string]any {
		r := httptest.NewRequest("POST", "/api/support/bundle", nil)
		r.Header.Set("Referer", "http://x/watch?"+q)
		w := httptest.NewRecorder()
		supportFixture(w, r)
		var m map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	if m := get(""); m["canSend"] != true || m["canSave"] != true || len(m["files"].([]any)) < 5 {
		t.Errorf("default: %v", m)
	}
	if m := get("support=blocked"); m["canSend"] != false || m["canSave"] != false || len(m["findings"].([]any)) == 0 {
		t.Errorf("blocked: %v", m)
	}
	if m := get("support=nosend"); m["canSend"] != false || m["canSave"] != true || m["configured"] != false {
		t.Errorf("nosend: %v", m)
	}
}
