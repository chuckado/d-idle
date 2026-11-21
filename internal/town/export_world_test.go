package town

import (
	"context"

	"d2lod-bot/internal/d2"
)

// BuyByVerbForTest and SellByVerbForTest expose the verb halves of the vendor visit to the data-free tests.
func (t *Town) BuyByVerbForTest(ctx context.Context, it d2.Unit, g0 int) (int, bool, error) {
	return t.buyByVerb(ctx, it, g0)
}

func (t *Town) SellByVerbForTest(ctx context.Context, u d2.Unit, sold func(*d2.Snapshot) bool) (bool, bool, error) {
	_, handled, ok, err := t.sellByVerb(ctx, u, sold)
	return handled, ok, err
}
