package invdesk

import (
	"context"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
)

// A queued identify protects the item from every town chore that reads the marks (sell, tidy, stash) like a keep mark,
// even over a sell mark, until the request is done or cancelled (user report 2026-11-15: the sell chore sold them).
func TestIdentifyRequestProtectsTheItemFromTheSellChore(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		t.Run(class.String(), func(t *testing.T) {
			d, _, snap := idRig(t, class, town1, false)
			u := snap.Inventory.Items[0]
			if got := d.MarkOf(u); got != "" {
				t.Fatalf("unmarked item reads %q", got)
			}
			r, err := d.RequestIdentify(1)
			if err != nil {
				t.Fatal(err)
			}
			if got := d.MarkOf(u); got != MarkKeep {
				t.Fatalf("identify pending: mark %q, want keep (the sell chore must skip it)", got)
			}
			if err := d.Cancel(r.ID, KindIdentify); err != nil {
				t.Fatal(err)
			}
			if got := d.MarkOf(u); got != "" {
				t.Fatalf("after cancel: mark %q", got)
			}
		})
	}
}

func TestLostIdentifyRequestIsReportedNotDropped(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		t.Run(class.String(), func(t *testing.T) {
			d, sv, snap := idRig(t, class, field, false)
			if _, err := d.RequestIdentify(1); err != nil {
				t.Fatal(err)
			}
			snap.Inventory.Items = snap.Inventory.Items[1:] // sold behind our back
			if lines := d.RunPending(context.Background()); len(lines) != 0 || len(sv.identified) != 0 {
				t.Fatalf("ran for a missing item: %v", lines)
			}
			v := d.View()
			if len(v.Queue) != 0 || len(v.Recent) == 0 {
				t.Fatalf("view %+v", v)
			}
			rc := v.Recent[0]
			if rc.State != StateFailed || rc.Code != CodeLost || !strings.Contains(rc.Reason, "lost") {
				t.Fatalf("recent %+v", rc)
			}
		})
	}
}
