package d2dctl

import (
	"context"
	"encoding/json"
	"strings"
)

// Points verbs (docs/points.md). `cmd stat <stat> [n]` and `cmd skillpt <index>` are stock d2d commands: "what a remote
// client would send" (fork docs/control_channel.md, the `cmd` row; V: apps/d2d/devctl/verbs.hpp sends cmd::StatPoint and
// cmd::SkillPoint, the same messages the stat button and the skill icon send). They need no lab flag. Arguments, V from
// the fork source: stat is the stat id (0 Strength, 1 Energy, 2 Dexterity, 3 Vitality, panels.hpp kStatButtons) and n a
// count (the world spends min(n, unspent)); skillpt takes the skill's INDEX 0..29 within its class, not the skills.txt id
// (components/game/world/commands.cpp hands the number to rules::can_learn as skill_index). The world applies a command at
// its next tick, so a caller reads the points back afterwards. Whether a stock build without the verb answers "err" is
// U; the caller treats any error as "no verb" and says so.

// StatID values of CmdStat.
const (
	StatStrength  = 0
	StatEnergy    = 1
	StatDexterity = 2
	StatVitality  = 3
)

// BaseAttrs reads the base attributes (str, dex, vit, ene) without item bonuses from `dump stats` (`stats.base`, V
// dump_head_late.json). It needs the fork's dump verb; ok is false on stock d2d.
func (c *Client) BaseAttrs(ctx context.Context) (attrs [4]int, ok bool, err error) {
	r, err := c.Do(ctx, "dump stats")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unknown") {
			return attrs, false, nil
		}
		return attrs, false, err
	}
	line := strings.TrimSpace(r.Rest)
	for _, l := range r.Lines {
		if strings.TrimSpace(l) != "" {
			line = l
			break
		}
	}
	var d struct {
		Stats struct {
			Base map[string]int `json:"base"`
		} `json:"stats"`
		Base map[string]int `json:"base"`
	}
	if err := json.Unmarshal([]byte(line), &d); err != nil {
		return attrs, false, nil
	}
	b := d.Stats.Base
	if len(b) == 0 {
		b = d.Base
	}
	for i, k := range [4]string{"str", "dex", "vit", "ene"} {
		v, has := b[k]
		if !has {
			return attrs, false, nil
		}
		attrs[i] = v
	}
	return attrs, true, nil
}
