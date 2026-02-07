package spellcasting_manager

import "fmt"

type SpellSlotErrorType string

const (
	ERROR_NO_SLOTS_AVAILABLE SpellSlotErrorType = "no slots available"
	ERROR_NO_SLOTS_LEFT      SpellSlotErrorType = "no slots left"
	ERROR_SLOT_ALREADY_USED  SpellSlotErrorType = "slot already used"
	ERROR_SLOT_NOT_AVAILABLE SpellSlotErrorType = "slot not available"
	ERROR_SLOT_NOT_FOUND     SpellSlotErrorType = "slot not found"
	ERROR_GENERIC_SLOT       SpellSlotErrorType = "generic slot error"
)

type SpellSlotError struct {
	Level   int
	Message string
	Type    SpellSlotErrorType
}

func (e *SpellSlotError) Error() string {
	return fmt.Sprintf("Spell slot error: %s. Level: %d. Message: %s", e.Type, e.Level, e.Message)
}

func NewSpellSlotError(level int, message string, errType SpellSlotErrorType) *SpellSlotError {
	return &SpellSlotError{
		Level:   level,
		Message: message,
		Type:    errType,
	}
}

func NewSpellSlotErrorOutOfSlots(level int) *SpellSlotError {
	return NewSpellSlotError(level, "out of slots", ERROR_NO_SLOTS_AVAILABLE)
}

func NewSpellSlotErrorNoSlotsLeft(level int) *SpellSlotError {
	return NewSpellSlotError(level, "no slots left", ERROR_NO_SLOTS_LEFT)
}

func NewSpellSlotErrorAlreadyUsed(level int) *SpellSlotError {
	return NewSpellSlotError(level, "already used", ERROR_SLOT_ALREADY_USED)
}

func NewSpellSlotErrorNotAvailable(level int) *SpellSlotError {
	return NewSpellSlotError(level, "not available", ERROR_SLOT_NOT_AVAILABLE)
}

func NewSpellSlotErrorNotFound(level int) *SpellSlotError {
	return NewSpellSlotError(level, "not found", ERROR_SLOT_NOT_FOUND)
}

type SpellcastingErrorType string

const (
	ERROR_SPELL_NOT_FOUND      SpellcastingErrorType = "spell not found"
	ERROR_SPELL_ALREADY_KNOWN  SpellcastingErrorType = "spell already known"
	ERROR_SPELL_NOT_KNOWN      SpellcastingErrorType = "spell not known"
	ERROR_SPELL_LEVEL_TOO_LOW  SpellcastingErrorType = "spell level too low"
	ERROR_SPELL_LEVEL_TOO_HIGH SpellcastingErrorType = "spell level too high"
	ERROR_GENERIC_SPELL        SpellcastingErrorType = "generic spell error"
)

type SpellcastingError struct {
	Name    string
	Message string
	Type    SpellcastingErrorType
}

func (e *SpellcastingError) Error() string {
	return fmt.Sprintf("SpellManager error: %s. Spell: %s. Message: %s", e.Type, e.Name, e.Message)
}

func NewSpellcastingError(name string, message string, errType SpellcastingErrorType) *SpellcastingError {
	return &SpellcastingError{
		Name:    name,
		Message: message,
		Type:    errType,
	}
}

func NewSpellcastingErrorNotFound(name string) *SpellcastingError {
	return NewSpellcastingError(name, "not found", ERROR_SPELL_NOT_FOUND)
}

func NewSpellcastingErrorAlreadyKnown(name string) *SpellcastingError {
	return NewSpellcastingError(name, "already known", ERROR_SPELL_ALREADY_KNOWN)
}

func NewSpellcastingErrorNotKnown(name string) *SpellcastingError {
	return NewSpellcastingError(name, "not known", ERROR_SPELL_NOT_KNOWN)
}

func NewSpellcastingErrorLevelTooLow(name string) *SpellcastingError {
	return NewSpellcastingError(name, "level too low", ERROR_SPELL_LEVEL_TOO_LOW)
}

func NewSpellcastingErrorLevelTooHigh(name string) *SpellcastingError {
	return NewSpellcastingError(name, "level too high", ERROR_SPELL_LEVEL_TOO_HIGH)
}
