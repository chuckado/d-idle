package main

// Defects of the "..." menu policy (card menufit2, the recurring 1280x620 failure of 2026-11-16): the real page with ONE defect put
// back, each over a menu made taller than its window (a group at least 700 px high stands for "someone added ten entries"). The fit
// pages menudefect-NAME must FAIL with the verdict mustFail names (fit.go). The healthy page is default-menu-tall: the same tall
// menu with no defect, which must pass because the body scrolls and says so.
func init() {
	const tall = `.d-menu-gb { min-block-size: 700px !important; }`
	defectCSS["mn-tall"] = tall
	defectCSS["mn-noscroll"] = tall + ` .d-menu-body { overflow-y: hidden !important; }`
	defectCSS["mn-primary"] = tall + ` .d-menu-primary { position: relative !important; inset-block-start: 4000px !important; }`
	defectCSS["mn-hint"] = tall + ` .d-menu-body::after { content: none !important; }`
}
