package combat

// SkillKey is the verified hotkey of the skill with that name, or "" when the character does not know it or it is not
// bound. The key map only fills with skills the character knows (EnsureHotkeys binds layout skills of KnownSkills,
// d2dcombat BindSkills the skills in the player's list), so a class or build without the skill reads "".
func (f *Fighter) SkillKey(name string) string {
	if f.cat == nil {
		return ""
	}
	sk := f.cat.SkillByName(name)
	if sk == nil {
		return ""
	}
	return f.keys[sk.ID]
}

// TeleportKey is the hotkey Teleport is bound to, or "" (navigation then walks).
func (f *Fighter) TeleportKey() string { return f.SkillKey("Teleport") }
