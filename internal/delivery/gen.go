package delivery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

// SkillsJSON renders the generated combat skill view, one skill per line.
func SkillsJSON(t *Table) ([]byte, error) {
	v, err := t.CombatView()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	b.WriteString("{\n\"generated\": \"by tools/capgen from data/capabilities/capabilities.json; do not edit\",\n\"skills\": [\n")
	for i, s := range v {
		line, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		b.Write(line)
		if i < len(v)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("]}\n")
	return b.Bytes(), nil
}

var kindsBlock = regexp.MustCompile(`(?s)"skillKinds": \{.*?\n \}`)

// ProfilesJSON returns profiles.json with its skillKinds block replaced by the generated one.
func ProfilesJSON(old []byte, t *Table) ([]byte, error) {
	kinds := t.PowerKinds()
	names := make([]string, 0, len(kinds))
	for n := range kinds {
		names = append(names, n)
	}
	sort.Strings(names)
	var b bytes.Buffer
	b.WriteString("\"skillKinds\": {\n")
	for i, n := range names {
		k, _ := json.Marshal(n)
		v, _ := json.Marshal(kinds[n])
		fmt.Fprintf(&b, "  %s: %s", k, v)
		if i < len(names)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString(" }")
	if !kindsBlock.Match(old) {
		return nil, fmt.Errorf("delivery: profiles.json has no skillKinds block")
	}
	return kindsBlock.ReplaceAllLiteral(old, b.Bytes()), nil
}
