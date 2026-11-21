package invdesk

// The item, gold and identify verbs became ordinary input in the fork (Q13, ADR 0021): a build that advertises the
// new name needs no lab flag. The bot adopts the new name only behind the knob input.planner (InputCaps on the mover
// and the equipper); with the knob off only the lab names count, byte for byte as before.
const (
	InputCapItemEquip   = "input.item-equip"
	InputCapItemUnequip = "input.item-unequip"
	InputCapItemMove    = "input.item-move"
	InputCapItemIdent   = "input.item-identify"
	InputCapItemDrop    = "input.item-drop"
	InputCapGoldStash   = "input.gold-stash"
)

// inputAlias maps a lab capability name to the input name of the same verb.
var inputAlias = map[string]string{
	CapItemEquip:    InputCapItemEquip,
	"item.unequip":  InputCapItemUnequip,
	CapItemMove:     InputCapItemMove,
	CapItemIdentify: InputCapItemIdent,
	CapItemDrop:     InputCapItemDrop,
	CapGoldStash:    InputCapGoldStash,
}

// capHas reports whether set lists name, or (when alias is true) the input name of the same verb.
func capHas(set map[string]bool, name string, alias bool) bool {
	if set[name] {
		return true
	}
	return alias && set[inputAlias[name]]
}
