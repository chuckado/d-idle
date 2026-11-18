package env

import "os"

// WatchTuning holds the watch screen's picture knobs (docs/video-feed.md). Owner: the videofeed cards.
type WatchTuning struct {
	PaletteFeed int // 1: offer the lossless tile-delta feed (/stream.pfeed) next to MJPEG; 0: MJPEG only
}

func DefaultWatchTuning() WatchTuning { return WatchTuning{} }

func watchKnobs() []Knob {
	return []Knob{
		{Key: "watch.paletteFeed", Label: "Lossless picture feed", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: the viewer (d2dview, and the watch screen through the tap relay) also serves the lossless RGBA tile-delta feed of the game picture (/stream.pfeed; measured 0.4 to 0.7 MB/s against 1.8 MB/s of JPEG q72, docs/probes/video-feed-2026-11-15.md) and the page draws it on a canvas when the glance reply lists it; any decode trouble falls back to MJPEG. Needs the d2d framefeed ring. d2dview and the friends pack read the environment switch D2_PALETTE_FEED=1 (same meaning; the d2dview process has no tuning file). 0 (default): MJPEG only, nothing changes. Off until the live A/B has run (docs/video-feed.md section 9)."},
	}
}

// PaletteFeedOn is the one rule every host uses: the environment switch D2_PALETTE_FEED=1 turns the feed on, D2_PALETTE_FEED=0
// forces it off, else the knob value decides (0 unless a tuning file sets it).
func PaletteFeedOn(knob float64) bool {
	switch os.Getenv("D2_PALETTE_FEED") {
	case "1":
		return true
	case "0":
		return false
	}
	return knob >= 1
}
