package env

// LootTuning holds the loot knobs. Owner: the loot agent; add fields, defaults and Knob entries here only.
type LootTuning struct {
	PickRadius             int // subtiles around the player in which ground items are considered
	PickStrictness         int // 0 loose (also sellable items), 1 normal, 2 strict (keepers only), 3 gold/runes/gems/uniques only
	PickTries              int // click attempts per item
	WalkDist               int // walk to within this many subtiles of an item before clicking it
	ConfirmMs              int // after a click, wait this long for the item to leave the ground
	GapMs                  int // pause between two pickups
	GoldMinStack           int // ignore gold piles smaller than this (a ceiling: GoldPileShare lowers it in early areas)
	GoldPileShare          int // percent of an area's piles the min-pile bar must let through (model: data.PileMinWorth); 0 = the flat GoldMinStack only
	FreeCellsKeep          int // inventory cells that must stay free after a pickup (the pickit may override for gold and runes)
	PotionBuffer           int // extra healing potions to carry in the pack beyond the belt (also rejuvenation-free kinds other than mana)
	PotionBufferMana       int // extra mana potions to carry in the pack beyond the belt (SME: fewer than healing)
	RejuvSpare             int // rejuvenation potions carried in total, belt plus pack (the rest is left or sold, Full preferred)
	LowGoldBelow           int // below this much gold, also pick items worth selling
	MinValuePerCell        int // sell price per inventory cell an item needs to be worth picking in low-gold mode
	ValuableMinGold        int // an unequipped item worth at least this (estimated sell value) counts as a valuable
	ValuableMinPrio        int // ... or a pickit rule priority at least this (runes, gems, uniques sit at 70+)
	CarryValuables         int // valuables the bot may carry out of town; the rest are stashed before a dangerous trip
	MaxPerCall             int // pickups per PickUp call
	PoorGoldBelow          int // below this gold the looter takes every gold pile and any junk weapon or armor worth PoorMinSell (a naked, broke character earns its rebuy that way)
	PoorMinSell            int // estimated sell value an item needs to be picked in poor mode
	MemEnabled             int // 1: remember wanted drops beyond the pickup radius and walk back for them (docs/lootmemory.md)
	MemMinGold             int // smallest remembered gold pile worth a walk-back at clvl 1; it grows by 10 percent per clvl
	MemMaxDist             int // remembered drops farther than this (subtiles, straight line) are left
	MemMaxTrips            int // walk-back trips per visit of an area
	MemCap                 int // remembered drops kept at most (lowest value evicted)
	GoldPriority           int // 1: gold piles are fetched right below survival (mid-fight when safe, before roaming on, before leaving an area; docs/goldprio.md)
	GoldGrabDist           int // subtiles: a pile this close is taken in the middle of a fight when the fight allows
	GoldSafeDist           int // subtiles: no hostile closer than this to the hero for a mid-fight grab
	GoldMaxDanger          int // highest pack danger (combat.Danger*) at which a mid-fight grab is still allowed
	GoldMinLifePct         int // life percent under which gold is ignored (survival first)
	GoldRetries            int // failed pickups of one pile (over all calls) before it is given up
	Vault                  int // 1: the Ethereal Vault preset decides the ground stage; vault items are recorded, not picked up (docs/vault.md); forced 0 in measured runs
	VaultAll               int // 1: with Vault, every other non-gold item that is not a gem or a stack is recorded too
	VaultKeep              int // 1 (with Vault): the Keep queue may create kept items in the stash; 0 stops creating (kill switch)
	VaultKeepPerVisit      int // most items the Keep queue creates in one town visit
	VaultEquip             int // 1 (with Vault and VaultKeep): the sheet of a vault item offers one-tap Equip (withdraw and wear in one town visit)
	VaultSellPerVisit      int // most vault items the sell errand creates and sells in one town visit
	VaultSellPrompt        int // 1: a Needs-you row asks "short on gold: sell these vault items?" (never sells by itself)
	VaultAutoTrim          int // 1: a full vault trims its least valuable unpinned item for a worthwhile drop and may sell flagged items by itself (old behaviour, an unlock for later); 0 (default): full means new drops are not recorded and nothing leaves
	VaultPoorPickup        int // 1 (with Vault): in poor mode an item worth selling is picked up instead of only recorded
	VaultRevealGold        int // 1 (with Vault and Keep): the page may order a reveal paid in gold (the bot buys the scrolls at the next town visit); 0 (default) off
	VaultRevealGoldConfirm int // gold from which the page asks for a confirm step before a gold reveal
	RulesEngine            int // loot rules engine stage: 0 off, 1 ground, 2 ground and tidy, 3 reserved (sell, needs a user decision; not implemented)
}

func DefaultLootTuning() LootTuning {
	return LootTuning{
		PickRadius: 30, PickStrictness: 1, PickTries: 3, WalkDist: 4, ConfirmMs: 1500, GapMs: 150,
		GoldMinStack: 10, GoldPileShare: 50, FreeCellsKeep: 2, PotionBuffer: 4, PotionBufferMana: 2, RejuvSpare: 3, LowGoldBelow: 2000, MinValuePerCell: 100, ValuableMinGold: 1500, ValuableMinPrio: 70, CarryValuables: 0, MaxPerCall: 30,
		PoorGoldBelow: 300, PoorMinSell: 20,
		MemEnabled: 1, MemMinGold: 20, MemMaxDist: 120, MemMaxTrips: 6, MemCap: 64,
		GoldPriority: 1, GoldGrabDist: 20, GoldSafeDist: 8, GoldMaxDanger: 0, GoldMinLifePct: 60, GoldRetries: 4, RulesEngine: 0, VaultKeep: 1, VaultKeepPerVisit: 20, VaultSellPerVisit: 10, VaultPoorPickup: 1, VaultRevealGoldConfirm: 2000,
	}
}

func lootK(key, label, unit, desc string, def, min, max int, higher bool) Knob {
	return Knob{Key: "loot." + key, Label: label, Unit: unit, Desc: desc,
		Default: float64(def), Min: float64(min), Max: float64(max), HigherIsBetter: higher}
}

func lootKnobs() []Knob {
	d := DefaultLootTuning()
	return []Knob{
		lootK("vault", "Ethereal Vault recording", "0/1", "1: items the Vault preset (docs/vault.md) selects are peeked and recorded in the vault instead of being picked up; the bot walks to them as for a pickup, sends no pickup click and uses no pack space. Gold, potions, scrolls, keys, gems and stacks keep their own rules. Needs d2d started with --vault-verbs; without it the knob does nothing and a Needs-you info line says why. Forced off in measured runs (gates, soaks, baselines: D2BOT_RUN). 0 (default): today's pickit, byte-identical.", d.Vault, 0, 1, false),
		lootK("vaultKeep", "Vault Keep queue", "0/1", "1 (default, with loot.vault 1): the Keep button of the Ethereal Vault queues the item and the bot creates it in the hero's stash at the next safe moment in town, exactly as dropped (docs/vault.md, Keep). 0: the queue stops creating at once (kill switch); queued items stay queued. Needs d2d with --vault-verbs; forced off in measured runs (D2BOT_RUN).", d.VaultKeep, 0, 1, false),
		lootK("vaultEquip", "Vault one-tap Equip", "0/1", "1 (with loot.vault 1 and loot.vaultKeep 1): the item sheet of the Ethereal Vault offers Equip: the bot takes the item out and wears it in the same town visit, vaults the item it replaced and pins the slot (docs/vault.md, Equip). 0 (default): the sheet offers Take out only. Forced off in measured runs (D2BOT_RUN).", d.VaultEquip, 0, 1, false),
		lootK("vaultKeepPerVisit", "Keep items per visit", "items", "Most items the Keep queue creates in one town visit, so a long queue does not stall the trip; the rest follows at the next safe moment.", d.VaultKeepPerVisit, 1, 100, false),
		lootK("vaultSellPerVisit", "Vault sell items per visit", "items", "Most vault items the Sell errand creates in the pack and sells at the vendor in one town visit (docs/vault.md, Sell); the rest follows at the next safe moment.", d.VaultSellPerVisit, 1, 50, false),
		lootK("vaultSellPrompt", "Short on gold: sell vault items?", "0/1", "1 (with loot.vault 1): when the town budget (repair, potion floor, mercenary reserve) is short of gold, a Needs-you row offers to sell the lowest-value unpinned vault items (never rare or better); Yes queues them for the Sell errand. 0 (default): no prompt. Nothing is ever sold without the answer.", d.VaultSellPrompt, 0, 1, false),
		lootK("vaultAutoTrim", "Vault removes items by itself", "0/1", "0 (default, user decision 2026-11-21): a full vault never removes, sells or replaces an item; new drops are not recorded (they are counted as missed drops on the Vault page) and only the player takes items out. 1: an unlock for later, the old behaviour: a worthwhile drop trims the least valuable unpinned item, and items marked for sale are sold at the next visit without a confirm (docs/vault.md, Decision 2026-11-21).", d.VaultAutoTrim, 0, 1, false),
		lootK("vaultPoorPickup", "Vault poor-mode pickup", "0/1", "1 (default, with loot.vault 1): below loot.poorGoldBelow gold the looter picks up the items worth loot.poorMinSell gold or more instead of only recording them, so a broke hero earns gold (docs/vault.md). 0: it records them.", d.VaultPoorPickup, 0, 1, false),
		lootK("vaultRevealGold", "Vault reveal paid in gold", "0/1", "1 (with loot.vault 1 and loot.vaultKeep 1): the Reveal bar of the Vault page offers Gold as a payment. The order makes the bot buy that many Scrolls of Identify at the vendor at the next town visit (real gold, never below the repair and potion-kit reserve) and read them on the vault items the same visit (docs/vault.md, Reveal paid in gold). 0 (default): off.", d.VaultRevealGold, 0, 1, false),
		lootK("vaultRevealGoldConfirm", "Gold reveal confirm from", "gold", "A gold reveal that costs this much or more asks for a second confirming press on the page.", d.VaultRevealGoldConfirm, 0, 99999, false),
		lootK("vaultAll", "Record everything non-gold", "0/1", "1 (with loot.vault 1): every ground item that is not gold, a gem or a stack and that no other rule takes is recorded too, plain and low items included. 0 (default): the Vault preset only.", d.VaultAll, 0, 1, false),
		lootK("pickRadius", "Pickup radius", "subtiles", "Ground items within this distance are considered.", d.PickRadius, 5, 60, true),
		lootK("pickStrictness", "Filter strictness", "level", "0 loose (also items worth selling), 1 normal, 2 keepers only, 3 gold/runes/gems/uniques only.", d.PickStrictness, 0, 3, false),
		lootK("pickTries", "Pickup attempts", "count", "Click attempts per item.", d.PickTries, 1, 6, true),
		lootK("walkDist", "Walk-to distance", "subtiles", "Walk to within this distance before clicking an item.", d.WalkDist, 2, 8, false),
		lootK("confirmMs", "Pickup confirm wait", "ms", "Wait after a click for the item to leave the ground.", d.ConfirmMs, 400, 4000, true),
		lootK("gapMs", "Pickup gap", "ms", "Pause between two pickups.", d.GapMs, 0, 1000, false),
		lootK("goldMinStack", "Min gold pile", "gold", "Ignore gold piles smaller than this (a ceiling: see goldPileShare).", d.GoldMinStack, 0, 2000, false),
		lootK("goldPileShare", "Gold pile share", "percent", "The min-pile bar never skips more than this share of the area's piles: pile size is ilvl to 6 ilvl - 1 coins (treasure-trace section 4), so Blood Moor piles (1..5) would all fall under 10. 0 keeps the flat bar.", d.GoldPileShare, 0, 100, false),
		lootK("freeCellsKeep", "Free cells kept", "cells", "Inventory cells that must stay free after a pickup.", d.FreeCellsKeep, 0, 12, true),
		lootK("potionBuffer", "Potion buffer", "count", "Extra healing potions carried in the pack beyond the belt (SME EXP, docs/levelplan.md: 4).", d.PotionBuffer, 0, 10, true),
		lootK("potionBufferMana", "Mana potion buffer", "count", "Extra mana potions carried in the pack beyond the belt.", d.PotionBufferMana, 0, 10, true),
		lootK("rejuvSpare", "Rejuvenation cap", "count", "Rejuvenation potions carried in total (belt plus pack); more are left on the ground or sold, Full Rejuvenation preferred over regular. Pack ones are moved to a free belt slot in town so the belt hotkey can drink them.", d.RejuvSpare, 0, 6, true),
		lootK("lowGoldBelow", "Low gold mode below", "gold", "Below this gold, also pick items worth selling.", d.LowGoldBelow, 0, 100000, false),
		lootK("minValuePerCell", "Min value per cell", "gold", "Sell value per cell an item needs to be picked in low-gold mode.", d.MinValuePerCell, 0, 5000, false),
		lootK("valuableMinGold", "Valuable: min value", "gold", "Unequipped items worth at least this are valuables (stash before dangerous trips).", d.ValuableMinGold, 100, 100000, false),
		lootK("valuableMinPrio", "Valuable: min rule prio", "prio", "Items whose pickit rule prio is at least this are valuables.", d.ValuableMinPrio, 1, 100, false),
		lootK("carryValuables", "Valuables carried out", "count", "Valuables allowed in the pack outside town (insurance: the rest stay in the stash).", d.CarryValuables, 0, 20, true),
		lootK("poorGoldBelow", "Poor mode below", "gold", "Below this gold every gold pile and junk weapon or armor worth selling is picked.", d.PoorGoldBelow, 0, 10000, false),
		lootK("poorMinSell", "Poor mode min sell value", "gold", "Estimated sell value a junk item needs to be picked in poor mode.", d.PoorMinSell, 0, 500, false),
		lootK("maxPerCall", "Pickups per call", "count", "Upper bound on pickups in one PickUp call.", d.MaxPerCall, 1, 100, true),
		lootK("memEnabled", "Loot memory", "0/1", "1: remember wanted drops beyond the pickup radius and walk back for them at quiet moments.", d.MemEnabled, 0, 1, true),
		lootK("memMinGold", "Loot memory: min gold", "gold", "Smallest remembered gold pile worth a walk-back at clvl 1; the bar grows by 10 percent per character level.", d.MemMinGold, 0, 5000, false),
		lootK("memMaxDist", "Loot memory: max distance", "subtiles", "Remembered drops farther away than this are not fetched.", d.MemMaxDist, 20, 400, true),
		lootK("memMaxTrips", "Loot memory: trips per area", "count", "Walk-back trips per visit of an area.", d.MemMaxTrips, 0, 30, true),
		lootK("goldPriority", "Gold priority", "0/1", "1: gold piles are fetched right below survival: mid-fight when safe, before the next roam leg, and before an area is left (walk-back also for small piles, no trip cap). 0: old behaviour.", d.GoldPriority, 0, 1, true),
		lootK("goldGrabDist", "Gold: mid-fight grab distance", "subtiles", "A pile this close is picked up during a fight when no hostile is near.", d.GoldGrabDist, 5, 40, true),
		lootK("goldSafeDist", "Gold: hostile-free distance", "subtiles", "A mid-fight grab needs every hostile at least this far from the hero.", d.GoldSafeDist, 2, 30, true),
		lootK("goldMaxDanger", "Gold: max pack danger", "level", "Highest pack danger (0 none, 1 champions or shamans, 2 elite, 3 deadly) at which a mid-fight grab is still made.", d.GoldMaxDanger, 0, 3, true),
		lootK("goldMinLifePct", "Gold: min life", "percent", "Below this life percent gold is ignored (survival comes first).", d.GoldMinLifePct, 0, 100, false),
		lootK("rulesEngine", "Loot rules engine stage", "stage", "0: today's pickit and tidy table decide (default). 1: the loot rules engine (docs/lootrules.md) decides what to pick up from the ground, with today's strictness, minimum priority, per-cell and small-gold filters still applied on top. 2: ground and the tidy chore's stash and drop decisions. 3: plus the sell chore (the five sell decisions of docs/lootrules.md section 13: marks first, then the sell rules, then the ground rules as the pickit judges today). If the engine errors or panics, that one decision falls back to today's code and is logged once.", d.RulesEngine, 0, 3, false),
		lootK("goldRetries", "Gold: pickup retries", "count", "Failed pickups of one pile, over all calls, before it is given up.", d.GoldRetries, 1, 10, true),
		lootK("memCap", "Loot memory: size", "count", "Remembered drops kept at most; the lowest value is evicted.", d.MemCap, 4, 512, true),
	}
}
