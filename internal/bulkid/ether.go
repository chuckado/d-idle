package bulkid

import (
	"errors"
	"fmt"

	"d2lod-bot/internal/vault"
)

// Ether errors.
var (
	ErrEtherOff   = errors.New("bulkid: Ether is not a payer of reveals (identify.etherPerReveal is 0)")
	ErrEtherShort = errors.New("bulkid: not enough Ether for the reveal")
)

// EtherPayer charges the account wallet for vault reveals the player asked for by hand. It is never called by the Cain
// trip or any automatic path: only an explicit tap of a lazy reveal reaches it. Price 0 means Ether is off.
type EtherPayer struct {
	W     vault.Wallet
	Price func() int // identify.etherPerReveal
}

func (p *EtherPayer) price() int {
	if p == nil || p.Price == nil {
		return 0
	}
	return max(p.Price(), 0)
}

// Enabled reports whether a reveal can be paid in Ether at all.
func (p *EtherPayer) Enabled() bool { return p != nil && p.W != nil && p.price() > 0 }

// Quote is the Ether n reveals cost (0 when Ether is off).
func (p *EtherPayer) Quote(n int) int {
	if !p.Enabled() {
		return 0
	}
	return n * p.price()
}

// Check refuses (without spending) when Ether is off or the balance is short for n reveals.
func (p *EtherPayer) Check(n int) error {
	if !p.Enabled() {
		return ErrEtherOff
	}
	bal, err := p.W.Balance()
	if err != nil {
		return err
	}
	if bal < p.Quote(n) {
		return fmt.Errorf("%w: need %d, have %d", ErrEtherShort, p.Quote(n), bal)
	}
	return nil
}

// Charge spends the price of one reveal of record id. The ledger key names the record, so a repeated charge for the
// same record is a no-op (the wallet's spends are idempotent per key) and the ledger shows what each Ether bought.
func (p *EtherPayer) Charge(id string) error {
	if err := p.Check(1); err != nil {
		return err
	}
	return p.W.Spend("reveal:"+id, p.price())
}
