package bulkid

import (
	"context"
	"errors"

	"d2lod-bot/internal/env"
)

// ScrollIO is the game side of the scroll path (policy 2): the town wires the pack tools and the vendor.
type ScrollIO interface {
	Charges(ctx context.Context) int // scrolls plus tome charges held
	Gold(ctx context.Context) int
	Identify(ctx context.Context, id uint32) error // one charge; ErrNoTool when none is left
	Buy(ctx context.Context, n int) (int, error)   // scrolls bought (may be fewer); nil Buyer: never called
}

// ErrNoTool is returned by ScrollIO.Identify when no scroll or tome charge is left.
var ErrNoTool = errors.New("no identify scroll or tome charge")

// ScrollReport says what the scroll path did.
type ScrollReport struct {
	Identified int
	Bought     int
	Spent      int    // charges used
	Reason     string // why it stopped early
}

// RunScrolls identifies the scroll picks with scrolls or tome charges from the pack. It spends at most
// identify.maxScrolls charges, buys only when allowBuy is set (town.identifyTome) and only so that the purse keeps
// identify.goldFloor, and never touches an item the plan did not pick.
func RunScrolls(ctx context.Context, tu env.IdentifyTuning, picks []Pick, needBuy, price int, allowBuy bool, io ScrollIO) ScrollReport {
	var rep ScrollReport
	if tu.Policy != env.IdentifyScrolls || len(picks) == 0 {
		return rep
	}
	if needBuy > 0 {
		switch {
		case !allowBuy:
			rep.Reason = "scrolls are not bought (town.identifyTome is 0)"
		case price <= 0:
			rep.Reason = "the scroll price is unknown"
		default:
			n := min(needBuy, (io.Gold(ctx)-tu.GoldFloor)/price, tu.MaxScrolls-io.Charges(ctx))
			if n <= 0 {
				rep.Reason = WhyGoldFloor
			} else if got, err := io.Buy(ctx, n); err != nil {
				rep.Reason = "buying scrolls failed: " + err.Error()
			} else {
				rep.Bought = got
			}
		}
	}
	for _, p := range picks {
		if rep.Spent >= tu.MaxScrolls {
			rep.Reason = WhyCap
			break
		}
		if err := ctx.Err(); err != nil {
			rep.Reason = err.Error()
			break
		}
		if err := io.Identify(ctx, p.Item.ID); err != nil {
			if errors.Is(err, ErrNoTool) {
				rep.Reason = "out of identify scrolls"
			} else {
				rep.Reason = err.Error()
			}
			break
		}
		rep.Identified++
		rep.Spent++
	}
	return rep
}
