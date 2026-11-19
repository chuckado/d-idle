package bot

import (
	"strings"
	"sync/atomic"

	"d2lod-bot/internal/data"
)

// PotionKind classifies belt and inventory potions.
type PotionKind int

const (
	PotionNone PotionKind = iota
	PotionHealing
	PotionMana
	PotionRejuv
)

// ItemMeta is what the bot needs to know about an item base. Class is d2.Unit.Class.
type ItemMeta struct {
	Name     string
	Code     string // base item code (bks, bkd, ...); empty when the catalog has none
	GoalType string // the goals.ItemFilter type word: rune, gem, charm, ring, amulet, ...
	W, H     int
	Potion   PotionKind
	Antidote bool // antidote potion (code yps); its own flag because PotionKind has no antidote
	Cost     int  // base vendor price, for the corpse value estimate
	Tier     int  // potion tier 1 (minor) .. 5 (super): the digit of the base code hp1..hp5 (same rule as combat.ItemClass); 0 = not a potion
}

// MonsterMeta describes a monster class.
type MonsterMeta struct {
	Name     string
	Killable bool
	NPC      bool
	Boss     bool
	AI       string // monstats AI column (friendly summons are recognised by it)
	Ranged   bool   // monstats rangedtype
	Level    [3]int // monstats level per difficulty (Normal, Nightmare, Hell); area overrides are not applied
}

// Catalog is the bot's read-only view of the game tables. *data.DB is wrapped by NewDataCatalog.
type Catalog interface {
	Item(class int) (ItemMeta, bool)
	Monster(class int) (MonsterMeta, bool)
	AreaName(id int) string
	AreaByName(name string) (int, bool)
	// ExpToReach is the total experience needed to reach level; 0 when unknown.
	ExpToReach(level int) int64
}

type dataCatalog struct {
	db   *data.DB
	memo []atomic.Pointer[ItemMeta] // ItemMeta per class, filled on first use: the item table is immutable and Item runs per item per poll
}

// NewDataCatalog adapts the extracted game tables. A nil db yields an empty catalog.
func NewDataCatalog(db *data.DB) Catalog {
	c := dataCatalog{db: db}
	if db != nil {
		c.memo = make([]atomic.Pointer[ItemMeta], len(db.Items))
	}
	return c
}

// goalTypes maps item type codes (and their ancestors) to the words used by goals.ItemFilter.
// UNVERIFIED against the real itemtypes.txt: codes are from memory of the standard table.
var goalTypes = []struct{ code, word string }{
	{"rune", "rune"}, {"gem", "gem"}, {"char", "charm"}, {"ring", "ring"}, {"amul", "amulet"}, {"jewl", "jewel"},
	{"helm", "helm"}, {"tors", "armor"}, {"shld", "shield"}, {"boot", "boots"}, {"glov", "gloves"}, {"belt", "belt"},
	{"weap", "weapon"}, {"key", "key"}, {"gold", "gold"}, {"pott", "potion"},
}

func (c dataCatalog) Item(class int) (ItemMeta, bool) {
	if c.db == nil || class < 0 || class >= len(c.db.Items) {
		return ItemMeta{}, false
	}
	if class >= len(c.memo) {
		return c.item(class), true
	}
	if p := c.memo[class].Load(); p != nil {
		return *p, true
	}
	m := c.item(class)
	c.memo[class].Store(&m)
	return m, true
}

func (c dataCatalog) item(class int) ItemMeta {
	b := c.db.Items[class]
	m := ItemMeta{Name: b.Display, Code: b.Code, W: b.Width, H: b.Height, Cost: b.Cost, Antidote: b.Code == "yps"}
	if m.Name == "" {
		m.Name = b.Name
	}
	for _, g := range goalTypes {
		if c.db.IsA(b.Type, g.code) {
			m.GoalType = g.word
			break
		}
	}
	switch {
	case c.db.IsA(b.Type, "hpot"):
		m.Potion = PotionHealing
		m.Tier = potionTier(b.Code)
	case c.db.IsA(b.Type, "mpot"):
		m.Potion = PotionMana
	case c.db.IsA(b.Type, "rpot"):
		m.Potion = PotionRejuv
	}
	return m
}

// InvFile is the inventory icon name of an item (data.DB.InvFile); ok is false when the tables do not name one.
func (c dataCatalog) InvFile(class int, q data.InvQuality, fileIndex int) (string, bool) {
	return c.db.InvFile(class, q, fileIndex)
}

// InvFilePic is InvFile for an item with its own picture index (data.DB.InvFilePic).
func (c dataCatalog) InvFilePic(class int, q data.InvQuality, fileIndex, pic int) (string, bool) {
	return c.db.InvFilePic(class, q, fileIndex, pic)
}

func (c dataCatalog) Monster(class int) (MonsterMeta, bool) {
	if c.db == nil {
		return MonsterMeta{}, false
	}
	m := c.db.MonsterByClassID(class)
	if m == nil {
		return MonsterMeta{}, false
	}
	n := m.Name
	if n == "" {
		n = m.Key
	}
	return MonsterMeta{Name: n, Killable: m.Killable, NPC: m.NPC, Boss: m.Boss, AI: m.AI, Ranged: m.Ranged, Level: m.Level}, true
}

func (c dataCatalog) AreaName(id int) string {
	if c.db == nil {
		return ""
	}
	if l := c.db.AreaByID(id); l != nil {
		if l.Name != "" {
			return l.Name
		}
		return l.Key
	}
	return ""
}

func (c dataCatalog) AreaByName(name string) (int, bool) {
	if c.db == nil {
		return 0, false
	}
	if l := c.db.AreaByName(strings.TrimSpace(name)); l != nil {
		return l.ID, true
	}
	return 0, false
}

func (c dataCatalog) ExpToReach(level int) int64 {
	if c.db == nil || c.db.Exp == nil {
		return 0
	}
	return c.db.ExpToLevel(level)
}

func potionTier(code string) int {
	if len(code) == 3 && code[2] >= '1' && code[2] <= '5' {
		return int(code[2] - '0')
	}
	return 0
}
