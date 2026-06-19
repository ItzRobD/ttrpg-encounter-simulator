package spellcasting_manager

import "dnd5e-encounter-simulator-backend/pkg_old/spells"

func (scm *SpellcastingManager) getSpellSlots() spells.SpellSlots {
	return scm.currentSlots
}

func (scm *SpellcastingManager) getMaxSpellSlots() spells.SpellSlots {
	return scm.maxSlots
}

func (scm *SpellcastingManager) getSpellSlotsAtLevel(slot int) int {
	return scm.currentSlots[slot]
}

func (scm *SpellcastingManager) getMaxSpellSlotsAtLevel(slot int) int {
	return scm.maxSlots[slot]
}

func (scm *SpellcastingManager) HasSpellSlotsAtLevel(slot int) bool {
	return scm.currentSlots[slot] > 0
}

func (scm *SpellcastingManager) HasAnySpellSlots() bool {
	for _, slot := range scm.currentSlots {
		if slot > 0 {
			return true
		}
	}
	return false
}

func (scm *SpellcastingManager) ExpendSpellSlot(slot int) error {
	if !scm.HasSpellSlotsAtLevel(slot) {
		return NewSpellSlotError(slot, "no spell slots available", ERROR_NO_SLOTS_AVAILABLE)
	}
	scm.currentSlots[slot]--
	return nil
}

func (scm *SpellcastingManager) RecoverSpellSlotByAmount(slot int, amount int) error {
	if scm.currentSlots[slot] < scm.maxSlots[slot] {
		scm.currentSlots[slot] += amount
		return nil
	}
	return NewSpellSlotError(slot, "unable to recover spell slot", ERROR_GENERIC_SLOT)
}

func (scm *SpellcastingManager) RecoverSpellSlotToMax(slot int) error {
	return scm.RecoverSpellSlotByAmount(slot, scm.maxSlots[slot]-scm.currentSlots[slot])
}

func (scm *SpellcastingManager) getHighestAvailableSpellSlot() (int, error) {
	for i := 9; i >= 1; i-- {
		if scm.currentSlots[i] > 0 {
			return i, nil
		}
	}
	return 0, NewSpellSlotError(0, "no spell slots available", ERROR_NO_SLOTS_AVAILABLE)
}
