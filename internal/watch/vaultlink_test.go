package watch

import (
	"strings"
	"testing"
)

// The Vault is one tap from the band (card mobilequests): a link next to Inventory, Skills and Points whose href shell.js
// sets from VAULT_URL (the theme choice passes through), and the Vault page has its way back.
func TestBandVaultLink(t *testing.T) {
	page, _ := webFS.ReadFile("web/watch.html")
	if !strings.Contains(string(page), `id="vaultlink" href="/watch/vault"`) {
		t.Error("watch.html: the band has no Vault link")
	}
	shell, _ := webFS.ReadFile("web/shell.js")
	if !strings.Contains(string(shell), "getElementById('vaultlink'); if (vl) vl.href = VAULT_URL") {
		t.Error("shell.js: the Vault link does not take VAULT_URL (theme passthrough)")
	}
	vjs, _ := webFS.ReadFile("web/vault.js")
	if !strings.Contains(string(vjs), "back.href = d.backHref || '/watch'") {
		t.Error("vault.js: the Vault page has no Back link")
	}
}

// The phone is a page: no panel list on the phone tabs may stay an inner scroll box (the 2026-11-21 report, Quests tab).
func TestPhoneTabsAreNotScrollTraps(t *testing.T) {
	css, _ := webFS.ReadFile("web/utilities.css")
	for _, want := range []string{`.d-main[data-cmode="phone"] .d-slottabs :is(.qb-box, .qb-list`, `overscroll-behavior: auto`, `data-cmode="phone"] { block-size: auto`} {
		if !strings.Contains(string(css), want) {
			t.Errorf("utilities.css: missing %q", want)
		}
	}
}
