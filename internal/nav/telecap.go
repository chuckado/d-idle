package nav

// teleKey is the hotkey Teleport is cast with: the fixed TeleportKey (rigs, tests) or what TeleportKeyFn reports now.
// Empty means the character has no usable Teleport and every caller walks.
func (n *Navigator) teleKey() string {
	if n.TeleportKey != "" {
		return n.TeleportKey
	}
	if n.TeleportKeyFn != nil {
		return n.TeleportKeyFn()
	}
	return ""
}

// HasTeleport reports whether a Teleport key is bound right now (task, boss and farm read this as CanTeleport).
func (n *Navigator) HasTeleport() bool { return n.hasTeleport() }

// teleAllowed applies the level rule (towns, Duriel's Lair: data.CanTeleportIn) when the bot installed one.
func (n *Navigator) teleAllowed(level int) bool { return n.TeleportOK == nil || n.TeleportOK(level) }
