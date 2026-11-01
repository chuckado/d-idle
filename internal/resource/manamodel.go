package resource

// The bot's model of the player's mana: natural regeneration and when a cast's cost leaves the pool. Evidence tags:
// V read = game.exe 1.14d disassembly (docs/research/sk-mana-trace.md), V-d2d = measured on the d2d backend
// (docs/probes/drills/README.md DRILLS13 M1, DRILLS15 M4). docs/fixmanamodel.md has the derivation.

const (
	// ManaRegenFrames: game.exe FUN_005806f0 adds max(maxmana(256ths) / ManaRegenFrames, 1) 256ths a frame (CharStats
	// ManaRegen 120 x 25 frames, the same for every class). V read; d2d does the same (V-d2d: 45 max mana, 750 ticks
	// = +2250/256 = 3 a frame, DRILLS15 M4). Both backends share this model, so there is no backend switch.
	ManaRegenFrames = 3000
	// FramesPerSec: game frames per second, a tick of the d2d lockstep is one frame (V-d2d).
	FramesPerSec = 25
	// ManaPotSecs: a mana potion's points arrive over 128 frames, not at once. V read (misc.txt len + FUN_005be3f0).
	ManaPotSecs = 128.0 / FramesPerSec
)

// ManaPerFrame256 is the 256ths of mana regenerated in one frame: the whole-number base, then the percent bonus on
// the truncated base (MulDiv, FUN_00483360), as game.exe does. maxMana is in points.
func ManaPerFrame256(maxMana, bonusPct int) int {
	if maxMana <= 0 {
		return 0
	}
	base := max(maxMana*256/ManaRegenFrames, 1)
	return base * (100 + bonusPct) / 100
}

// ManaPerSec is the natural regeneration rate in points a second: 45 max mana gives 3 x 25 / 256 = 0.29 (153 s to
// fill), 250 gives 2.05 (122 s). The flat stat 26 (manarecovery) is not modelled.
func ManaPerSec(maxMana, bonusPct int) float64 {
	return float64(ManaPerFrame256(maxMana, bonusPct)) * FramesPerSec / 256
}

// ManaBooking says when a cast's cost leaves the pool and whether a cut cast still pays.
type ManaBooking struct {
	DelaySec float64 // command to payment
	Refund   bool    // a cast cut off before the payment never pays
}

var (
	// BookD2D: d2d pays 1 tick (40 ms) after the command, even for a cast cut off afterwards (V-d2d DRILLS13 M1
	// "1 tick after the command", DRILLS15 M2: 7 of 7 cut casts paid).
	BookD2D = ManaBooking{DelaySec: 0.04}
	// BookGameExe: game.exe pays at the action frame and not at all when the cast is cut first (V read, trace D1 and
	// D3). The action frame is 8 ticks for Fire Bolt level 1 without FCR (V-d2d: d2d pays 7 ticks before it); other
	// skills and FCR move it (U), and the four start-paying skills pay at once (not modelled).
	BookGameExe = ManaBooking{DelaySec: 0.32, Refund: true}
)

type booked struct {
	due  float64
	cost int
}

// ManaBook holds casts that were commanded but not yet paid, so the ManaBurnTracker takes a cost off its samples
// when the mana actually drops and not when the command is sent. Pure; time is the caller's clock in seconds.
type ManaBook struct {
	Model   ManaBooking
	pending []booked
}

// Add books a cast sent at now.
func (b *ManaBook) Add(now float64, cost int) {
	if cost > 0 {
		b.pending = append(b.pending, booked{now + b.Model.DelaySec, cost})
	}
}

// Cut tells the book that the player was interrupted at now: casts not yet paid are dropped when the backend
// refunds a cut cast (game.exe); on d2d it changes nothing.
func (b *ManaBook) Cut() {
	if b.Model.Refund {
		b.pending = b.pending[:0]
	}
}

// Take returns the cost of the casts due by now and forgets them.
func (b *ManaBook) Take(now float64) int {
	n, keep := 0, b.pending[:0]
	for _, c := range b.pending {
		if c.due <= now {
			n += c.cost
		} else {
			keep = append(keep, c)
		}
	}
	b.pending = keep
	return n
}

// Reset drops every booking (a new fight).
func (b *ManaBook) Reset() { b.pending = b.pending[:0] }
