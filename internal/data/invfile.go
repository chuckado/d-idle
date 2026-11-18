package data

import "strings"

// InvQuality selects which icon columns InvFile consults.
type InvQuality int

const (
	InvPlain  InvQuality = iota // magic, rare, crafted, runeword, normal: the base invfile
	InvUnique                   // uniqueitems invfile, then the base uniqueinvfile
	InvSet                      // setitems invfile, then the base setinvfile
)

// InvFile resolves the inventory icon name (lower case, no extension, as in data\global\items\<name>.dc6)
// of an item. class is the base item row (index into Items), fileIndex the unique or set id of the item.
//
// Order: unique -> uniqueitems.invfile, else base uniqueinvfile, else base invfile; set -> setitems.invfile,
// else base setinvfile, else base invfile; everything else -> base invfile. VERIFIED that every named file
// exists in the 1.14d archives (docs/research/paperdoll-art.md section 2); the exact game rule and the row
// semantics of fileIndex are UNVERIFIED, so an override row whose base code is not the item's own is ignored
// (a wrong index can then only fall back to the base icon, never show another item's).
// ok is false when the class is unknown or the base has no icon name at all.
func (db *DB) InvFile(class int, q InvQuality, fileIndex int) (name string, ok bool) {
	return db.InvFilePic(class, q, fileIndex, -1)
}

// InvFilePic is InvFile for an item that carries its own picture index (ItemTypes InvGfx1..6, the item record's picture
// field; pic < 0 = none). The picture replaces the BASE icon only: a unique or set item that has an icon of its own
// keeps it. An index the type does not list (empty slot, past the end, a base type without pictures) falls back to the
// base icon, so an older d2d build or a wrong index can only show the plain art. Tags: table columns V (itemtypes.txt of
// the 1.14d install), the pic semantics V in the fork (Item::picture indexes InvGfx), unique-override rule U.
func (db *DB) InvFilePic(class int, q InvQuality, fileIndex, pic int) (name string, ok bool) {
	if db == nil || class < 0 || class >= len(db.Items) {
		return "", false
	}
	b := db.Items[class]
	var over, special string
	switch q {
	case InvUnique:
		special = b.UniqueInvFile
		for _, u := range db.Uniques {
			if u.Row == fileIndex {
				if strings.EqualFold(u.Code, b.Code) {
					over = u.InvFile
				}
				break
			}
		}
	case InvSet:
		special = b.SetInvFile
		for _, s := range db.SetItems {
			if s.Row == fileIndex {
				if strings.EqualFold(s.Code, b.Code) {
					over = s.InvFile
				}
				break
			}
		}
	}
	if over == "" && special == "" && pic >= 0 && pic < 6 {
		if g := db.invGfx(b.Type, pic); g != "" {
			return g, true
		}
	}
	for _, n := range []string{over, special, b.InvFile} {
		if n != "" {
			return strings.ToLower(n), true
		}
	}
	return "", false
}

// invGfx is InvGfx[pic] of the item type; it scans the rows when the by-code index is not built (a hand-made DB).
func (db *DB) invGfx(typeCode string, pic int) string {
	if t := db.ItemType(typeCode); t != nil {
		return t.InvGfx[pic]
	}
	for _, t := range db.ItemTypes {
		if t.Code == typeCode {
			return t.InvGfx[pic]
		}
	}
	return ""
}
