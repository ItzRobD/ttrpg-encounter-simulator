package shared

// SpellSlots Definition
// Spell slots are handled as maps
// Max Spell slots are of type map[int]map[int]int where the outer value is the character level
// and the inner value is a map of spell slots which are of type map[int]int
// where the key is the spell slot level and the value is the number of slots
type SpellSlots map[int]int
