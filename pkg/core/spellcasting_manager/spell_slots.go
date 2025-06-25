package spellcasting_manager

func (s *SpellcastingManager) getSpellSlots() SpellSlots {
	return s.currentSlots
}

func (s *SpellcastingManager) getMaxSpellSlots() SpellSlots {
	return s.maxSlots
}

func (s *SpellcastingManager) getSpellSlotsAtLevel(slot int) int {
	return s.currentSlots[slot]
}

func (s *SpellcastingManager) getMaxSpellSlotsAtLevel(slot int) int {
	return s.maxSlots[slot]
}

func (s *SpellcastingManager) HasSpellSlotsAtLevel(slot int) bool {
	return s.currentSlots[slot] > 0
}

func (s *SpellcastingManager) HasAnySpellSlots() bool {
	for _, slot := range s.currentSlots {
		if slot > 0 {
			return true
		}
	}
	return false
}

func (s *SpellcastingManager) ExpendSpellSlot(slot int) error {
	if !s.HasSpellSlotsAtLevel(slot) {
		return NewSpellSlotError(slot, "no spell slots available", ERROR_NO_SLOTS_AVAILABLE)
	}
	s.currentSlots[slot]--
	return nil
}

func (s *SpellcastingManager) RecoverSpellSlotByAmount(slot int, amount int) error {
	if s.currentSlots[slot] < s.maxSlots[slot] {
		s.currentSlots[slot] += amount
		return nil
	}
	return NewSpellSlotError(slot, "unable to recover spell slot", ERROR_GENERIC_SLOT)
}

func (s *SpellcastingManager) RecoverSpellSlotToMax(slot int) error {
	return s.RecoverSpellSlotByAmount(slot, s.maxSlots[slot]-s.currentSlots[slot])
}

func (s *SpellcastingManager) getHighestAvailableSpellSlot() (int, error) {
	for i := len(s.currentSlots) - 1; i > 0; i-- {
		if s.currentSlots[i] > 0 {
			return i, nil
		}
	}
	return 0, NewSpellSlotError(0, "no spell slots available", ERROR_NO_SLOTS_AVAILABLE)
}
