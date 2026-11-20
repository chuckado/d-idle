package data

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// StatLines words an item's property list, d2d's "stat:param:value;..." (value as the game keeps it), with the game's own
// itemstatcost columns (descpriority, descfunc, descval, descstrpos/neg, descstr2) and string table, the way the item
// tooltip does. One function for every page that shows an item's properties (the Vault page uses it).
//
// S: the descfunc meanings are the community table of itemstatcost.txt (1-5, 12 plain and percent forms; 6-10 the same with
// a per-level value; 13 class skills, 14 skill tabs, 15 chance to cast, 16 aura, 19 a printf string, 20 and 21 negated, 24
// charges, 27 and 28 one skill), plus the grouped lines the game builds out of several stats (All Resistances, the damage
// ranges, poison over time). U: the string KEYS of the grouped lines and of 13 and 14 were not read from a real install here.
// A stat or a descfunc this file does not know is left out, never worded by guess: a missing key, a format that does not fit
// its arguments or an unknown skill drops that one line only.
func (db *DB) StatLines(props string) []string {
	type ent struct {
		s          *StatCost
		param, val int
		used       bool
	}
	var ents []*ent
	byName := map[string]*ent{}
	for _, p := range strings.Split(props, ";") {
		f := strings.Split(p, ":")
		if len(f) != 3 {
			continue
		}
		id, e1 := strconv.Atoi(f[0])
		par, e2 := strconv.Atoi(f[1])
		val, e3 := strconv.Atoi(f[2])
		s := db.StatByID(id)
		if e1 != nil || e2 != nil || e3 != nil || s == nil {
			continue
		}
		e := &ent{s: s, param: par, val: s.FromMemory(val)}
		ents = append(ents, e)
		if _, dup := byName[lc(s.Name)]; !dup {
			byName[lc(s.Name)] = e
		}
	}
	type line struct {
		prio int
		text string
	}
	var out []line
	add := func(prio int, t string) {
		if t != "" {
			out = append(out, line{prio, t})
		}
	}
	str := func(key string) string {
		v, ok := db.Str(key)
		if !ok {
			return ""
		}
		return strings.TrimSpace(CleanString(v))
	}

	// the lines the game builds from several stats
	if r := [4]*ent{byName["fireresist"], byName["lightresist"], byName["coldresist"], byName["poisonresist"]}; r[0] != nil && r[1] != nil && r[2] != nil && r[3] != nil &&
		r[0].val == r[1].val && r[1].val == r[2].val && r[2].val == r[3].val {
		if s := str("strModAllResistances"); s != "" {
			add(r[0].s.DescPriority, s+" "+signedInt(r[0].val))
			for _, e := range r {
				e.used = true
			}
		}
	}
	for _, g := range [][4]string{
		{"mindamage", "maxdamage", "strModMinDamageRange", "strModMinDamage"},
		{"firemindam", "firemaxdam", "strModFireDamageRange", "strModFireDamage"},
		{"lightmindam", "lightmaxdam", "strModLightningDamageRange", "strModLightningDamage"},
		{"coldmindam", "coldmaxdam", "strModColdDamageRange", "strModColdDamage"},
		{"magicmindam", "magicmaxdam", "strModMagicDamageRange", "strModMagicDamage"},
	} {
		lo, hi := byName[g[0]], byName[g[1]]
		if lo == nil || hi == nil {
			continue
		}
		var t string
		if lo.val == hi.val {
			t, _ = gfmt(str(g[3]), lo.val)
		}
		if t == "" {
			t, _ = gfmt(str(g[2]), lo.val, hi.val)
		}
		if t != "" {
			add(lo.s.DescPriority, t)
			lo.used, hi.used = true, true
		}
	}
	if lo, hi, ln := byName["poisonmindam"], byName["poisonmaxdam"], byName["poisonlength"]; lo != nil && hi != nil && ln != nil && ln.val > 0 {
		// the stored value is damage per frame times 256; the length is in frames (25 a second)
		if t, ok := gfmt(str("strModPoisonDamage"), (lo.val*ln.val+hi.val*ln.val)/512, ln.val/25); ok {
			add(lo.s.DescPriority, t)
			lo.used, hi.used, ln.used = true, true, true
		}
	}

	for _, e := range ents {
		if e.used || e.s.DescFunc == 0 {
			continue
		}
		if t := db.statLine(e.s, e.param, e.val, str); t != "" {
			add(e.s.DescPriority, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].prio > out[j].prio })
	lines := make([]string, len(out))
	for i, l := range out {
		lines[i] = l.text
	}
	return lines
}

// statLine words one stat; "" when the function or its data is not known.
func (db *DB) statLine(s *StatCost, param, val int, str func(string) string) string {
	key := s.DescStrPos
	if val < 0 && s.DescStrNeg != "" {
		key = s.DescStrNeg
	}
	text := str(key)
	if text == "" {
		return ""
	}
	with := func(v string) string { // descval: 0 no value, 1 value first, 2 value after the text
		switch s.DescVal {
		case 0:
			return text
		case 2:
			return text + " " + v
		}
		return v + " " + text
	}
	skill := func(id int) string {
		if sk := db.SkillByID(id); sk != nil && sk.Name != "" {
			return sk.Name
		}
		return ""
	}
	switch s.DescFunc {
	case 1, 12:
		return with(signedInt(val))
	case 2:
		return with(strconv.Itoa(val) + "%")
	case 3:
		return with(strconv.Itoa(val))
	case 4:
		return with(signedInt(val) + "%")
	case 5:
		return with(strconv.Itoa(val*100/128) + "%")
	case 6, 7, 8, 9, 10: // per character level: the value is in eighths per level; shown per level (U: the game multiplies by the character level)
		per := strconv.FormatFloat(float64(val)/8, 'f', -1, 64)
		if s.DescFunc == 10 {
			per = strconv.FormatFloat(float64(val*100/128)/8, 'f', -1, 64)
		}
		if s.DescFunc == 6 || s.DescFunc == 8 {
			per = "+" + per
		}
		if s.DescFunc >= 7 && s.DescFunc != 9 {
			per += "%"
		}
		return with(per) + " per level"
	case 13: // class skill levels: the param is the class, one string per class (U: key)
		if c := str("ModStr3" + string(rune('a'+param))); c != "" && param >= 0 && param < 7 {
			return signedInt(val) + " " + c
		}
	case 14: // skill tab: param is tab + 8 * class, one string per tab holding the %d (U: key)
		if p, ok := gfmt(str("StrSklTabItem"+strconv.Itoa((param>>3)*3+param&7+1)), val); ok && param>>3 < 7 && param&7 < 3 {
			return p
		}
	case 15: // chance to cast: param is skill*64 + level
		if n := skill(param >> 6); n != "" {
			if t, ok := gfmt(text, val, param&63, n); ok {
				return t
			}
		}
	case 16: // aura on the item
		if n := skill(param); n != "" {
			if t, ok := gfmt(text, val, n); ok {
				return t
			}
		}
	case 19: // the string is itself a format for the value
		if t, ok := gfmt(text, val); ok {
			return t
		}
	case 20:
		return with(strconv.Itoa(-val) + "%")
	case 21:
		return with(strconv.Itoa(-val))
	case 24: // charges: param is skill*64 + level, value is current + 256 * maximum
		if n := skill(param >> 6); n != "" {
			if t, ok := gfmt(text, param&63, n, val&255, val>>8); ok {
				return t
			}
		}
	case 27, 28: // +N to one skill; 27 adds the class that owns it
		sk := db.SkillByID(param)
		if sk == nil || sk.Name == "" {
			return ""
		}
		args := []any{sk.Name}
		if s.DescFunc == 27 {
			c := db.ClassByCode(sk.Class)
			if c == nil {
				return ""
			}
			args = append(args, c.Name)
		}
		if t, ok := gfmt(text, args...); ok {
			return signedInt(val) + " " + t
		}
	}
	return ""
}

// ClassByCode finds a class by its three-letter skills.txt code, nil when unknown.
func (db *DB) ClassByCode(code string) *CharClass {
	for _, c := range db.Classes {
		if code != "" && c.Code == code {
			return c
		}
	}
	return nil
}

func signedInt(v int) string {
	if v >= 0 {
		return fmt.Sprintf("+%d", v)
	}
	return strconv.Itoa(v)
}

// gfmt fills a string-table format (%d, %+d, %s, %%) in order; ok is false when the verbs do not fit the arguments, so a
// string the table does not word the way we assume is dropped instead of printed half filled.
func gfmt(f string, args ...any) (string, bool) {
	if f == "" {
		return "", false
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(f); i++ {
		if f[i] != '%' {
			b.WriteByte(f[i])
			continue
		}
		i++
		plus := false
		if i < len(f) && f[i] == '+' {
			plus = true
			i++
		}
		if i >= len(f) {
			return "", false
		}
		switch f[i] {
		case '%':
			b.WriteByte('%')
		case 'd', 'i':
			v, ok := argAt(args, n).(int)
			if !ok {
				return "", false
			}
			if plus {
				b.WriteString(signedInt(v))
			} else {
				b.WriteString(strconv.Itoa(v))
			}
			n++
		case 's':
			v, ok := argAt(args, n).(string)
			if !ok {
				return "", false
			}
			b.WriteString(v)
			n++
		default:
			return "", false
		}
	}
	return b.String(), n == len(args)
}

func argAt(a []any, i int) any {
	if i < len(a) {
		return a[i]
	}
	return nil
}
