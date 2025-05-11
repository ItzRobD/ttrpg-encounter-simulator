package shared

type SaveProficiencies struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// MakeSavingThrow attempts a saving throw for an actor based on a specific ability score and target difficulty.
// Returns true if the saving throw is successful, otherwise false, and an error if there is any.
//func MakeSavingThrow(actor core.Entity, ability AbilityScore, target int) (bool, error) {
//	var mod int
//	var err error
//	switch c := actor.(type) {
//	case *character.Character:
//		// TODO: Implement character saving throw logic
//	case *monster.Monster:
//		switch ability {
//		case spells.SpellDCStrength:
//			if c.SaveProficiencies.Strength == 0 {
//				mod, err = GetAbilityScoreModifier(c.AbilityScores.Strength)
//				if err != nil {
//					return false, err
//				}
//			} else {
//				mod = c.SaveProficiencies.Strength
//			}
//		case spells.SpellDCDexterity:
//			if c.SaveProficiencies.Dexterity == 0 {
//				mod, err = GetAbilityScoreModifier(c.AbilityScores.Dexterity)
//				if err != nil {
//					return false, err
//				}
//			} else {
//				mod = c.SaveProficiencies.Dexterity
//			}
//		case spells.SpellDCConstitution:
//			if c.SaveProficiencies.Constitution == 0 {
//				mod, err = GetAbilityScoreModifier(c.AbilityScores.Constitution)
//				if err != nil {
//					return false, err
//				}
//			} else {
//				mod = c.SaveProficiencies.Constitution
//			}
//		case spells.SpellDCIntelligence:
//			if c.SaveProficiencies.Intelligence == 0 {
//				mod, err = GetAbilityScoreModifier(c.AbilityScores.Intelligence)
//				if err != nil {
//					return false, err
//				}
//			} else {
//				mod = c.SaveProficiencies.Intelligence
//			}
//		case spells.SpellDCWisdom:
//			if c.SaveProficiencies.Wisdom == 0 {
//				mod, err = GetAbilityScoreModifier(c.AbilityScores.Wisdom)
//				if err != nil {
//					return false, err
//				}
//			} else {
//				mod = c.SaveProficiencies.Wisdom
//			}
//		case spells.SpellDCCharisma:
//			if c.SaveProficiencies.Charisma == 0 {
//				mod, err = GetAbilityScoreModifier(c.AbilityScores.Charisma)
//				if err != nil {
//					return false, err
//				}
//			} else {
//				mod = c.SaveProficiencies.Charisma
//			}
//		default:
//			return false, fmt.Errorf("invalid ability: %s", ability)
//		}
//	}
//
//	roll, rolls, err := RollDice(1, 20)
//	if err != nil {
//		return false, err
//	}
//	result := roll + mod
//
//	if result >= target {
//		events.LogSavingThrowEvent(actor, result, rolls, mod, true, actor.GetEventListener())
//		return true, nil
//	}
//	events.LogSavingThrowEvent(actor, result, rolls, mod, false, actor.GetEventListener())
//	return false, nil
//}
