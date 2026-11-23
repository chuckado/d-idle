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
	// the Orders More list squeezed to nothing again (card morefit): the shape of the 2026-11-17 screenshot
	defectCSS["mf-clip"] = `[data-panel="orders"] .o-items { flex: 1 1 0 !important; min-block-size: 0 !important; max-block-size: 30px !important; overflow-y: hidden !important; }`
	// the phone sheet cannot scroll again (data-scroll once beat .d-tilesheet in the phone rule of utilities.css)
	defectCSS["mf-phone"] = `.d-tilesheet { overflow: visible !important; max-block-size: none !important; } .d-tilesheet > * { flex: none !important; } .d-tilesheet { min-block-size: 0; padding-block-end: 80px !important; }`
	defectCSS["mn-hint"] = tall + ` .d-menu-body::after { content: none !important; }`
}
