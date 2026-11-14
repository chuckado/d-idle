package lootrules

// Stat names a rule may use (condition field "stat:<name>"). Values are display values (life already divided by 256).
// Ids are S: the stat numbers of the existing pickit (internal/loot/pickit_starter.json); the others are U until the
// adapter card reads them.
var statNames = map[string]bool{
	"life": true, "mana": true, "str": true, "dex": true, "vit": true, "enr": true,
	"fireres": true, "coldres": true, "lightres": true, "poisres": true,
	"fcr": true, "fhr": true, "frw": true, "ias": true, "mf": true, "gf": true,
	"allskills": true, "classskills": true, "skilltab": true, "sockets": true, "ed": true,
}

// derived stats computed by the evaluator
var resNames = []string{"fireres", "coldres", "lightres", "poisres"}

func statOf(it *Item, name string) int {
	if name == "ressum" {
		n := 0
		for _, r := range resNames {
			n += it.Stats[r]
		}
		return n
	}
	return it.Stats[name]
}
