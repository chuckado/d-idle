package d2s

// Tail is a copy of the bytes from the "if" skill block to the end: skills, items, corpse and golem
// data, uninterpreted. Hashing it identifies a character's kit independent of header and stats.
func (f *File) Tail() []byte { return append([]byte(nil), f.tail...) }
