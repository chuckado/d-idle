package d2dctl

import (
	"context"
	"fmt"
	"strings"
)

// CapLabels is the capability of the fork's ground item label verb (`labels on|off|toggle|status`, the fork branch
// didle/labels-verb, docs/video-feed.md section "Item labels"). The fork advertises it with --item-verbs or --lab-verbs.
// V (fork headless test labels_d2d.py): caps, the reply shape and the persistence; U live on the watch page.
const CapLabels = "ui.labels"

// Labels sets (mode "on", "off", "toggle") or reads ("status") the ground item name labels and returns the state the
// game reports back. The flag is held by the game until it is set off; a new d2d process starts with it off. A fork
// without the verb, or without --item-verbs, answers an error (a *CmdError): the caller shows the button as unavailable.
func (c *Client) Labels(ctx context.Context, mode string) (bool, error) {
	switch mode {
	case "on", "off", "toggle", "status":
	default:
		return false, fmt.Errorf("d2dctl: labels mode %q", mode)
	}
	r, err := c.Do(ctx, "labels "+mode)
	if err != nil {
		return false, err
	}
	for _, l := range append(append([]string(nil), r.Lines...), r.Rest) {
		if v, ok := ParseKV(strings.TrimSpace(l)).Get("labels"); ok {
			return v == "on", nil
		}
	}
	return false, fmt.Errorf("d2dctl: labels: no state in reply %q", r.Text())
}
