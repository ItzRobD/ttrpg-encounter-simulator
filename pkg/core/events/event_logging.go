package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"time"
)

type ActionChoiceData struct {
	Choice       core.ActionType
	AllScores    []ActionUtilityScore
	TopReasons   []DecisionFactor
	UtilityScore float64
}

type DragonbornBreathWeaponData struct {
	Target      core.Entity
	DamageTotal int
	DamageType  string
	DC          int
	SaveAbility string
	SaveSuccess bool
	SaveResult  int
}

type DamageData struct {
	Target     core.Entity
	DamageType string
	Damage     int
	Rolls      []int
}

type SpellChoiceData struct {
	Choice *core.SpellChoice
	Status *spells.SpellcastingManagerStatus
}

type SpellDCData struct {
	Target core.Entity
	Spell  *spells.Spell
	DC     int
	Save   int
	IsHit  bool
}

type LayOnHandsHealData struct {
	Subject core.Entity
	Value   int
}

type DamageModifiedData struct {
	Subject      core.Entity
	Res          core.DamageModificationResult
	SourceRollID string
}

type HPModifiedData struct {
	Subject      core.Entity
	Res          core.HPModificationResult
	SourceRollID string
}

type HPRollData struct {
	RollSum int
	Rolls   []int
	ToAdd   int
}

type SavingThrowData struct {
	Result   int
	Roll     int
	Modifier int
	Success  bool
}

type TargetChoiceData struct {
	Target  core.Entity
	Score   float64
	Factors map[DecisionFactor]float64
}

type SpecialAbilityData struct {
	AbilityName string
	Description string
	TargetName  string
	Value       int
}

func setupBaseEvent(ctx *core.EventContext, actor core.Entity, event CombatEvent) {
	event.SetActor(actor)
	event.SetTimestamp(time.Now())
	event.SetContext(ctx)
	if ctx != nil {
		// If it's an action choice, we want to use the parent ID (the Action ID) as the ID of this event
		// so that child events correctly link to it.
		if event.GetEventType() == ETActionChoiceEvent {
			event.SetID(ctx.GetParentID())
		} else {
			ctx.GenerateCurrentID()
			event.SetID(ctx.GetCurrentID())
		}
	} else {
		event.MakeNewEventID()
	}
}

func LogCharacterActionChoiceEvent(ctx *core.EventContext, actor core.Entity, choice core.ActionType, allScores []ActionUtilityScore, topReasons []DecisionFactor, utilityScore float64, listener func(event interface{})) {
	event := &ActionChoiceEvent{
		ActionChoice: choice,
		AllScores:    allScores,
		TopReasons:   topReasons,
		UtilityScore: utilityScore,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogMonsterActionChoiceEvent(ctx *core.EventContext, actor core.Entity, choice core.ActionType, allScores []ActionUtilityScore, topReasons []DecisionFactor, utilityScore float64, listener func(event interface{})) {
	event := &ActionChoiceEvent{
		ActionChoice: choice,
		AllScores:    allScores,
		TopReasons:   topReasons,
		UtilityScore: utilityScore,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogMeleeAttackEvent(ctx *core.EventContext, actor core.Entity, attackResult *core.AttackResult, listener func(event interface{})) {
	var damageTotal int
	var damageType string
	if attackResult.GetDamageResult() != nil {
		damageTotal = attackResult.GetDamageResult().GetTotal()
		damageType = attackResult.GetDamageType().String()
	}

	event := &MeleeAttackEvent{
		Target:         attackResult.GetTargetName(),
		AttackName:     attackResult.GetAttackName(),
		AttackCount:    attackResult.GetAttackCount(),
		AttackRoll:     attackResult.GetAttackRoll(),
		AttackModifier: attackResult.GetAttackTotal() - attackResult.GetAttackRoll(),
		AttackTotal:    attackResult.GetAttackTotal(),
		TargetValue:    attackResult.GetTargetValue(),
		Success:        attackResult.GetIsHit(),
		CriticalHit:    attackResult.GetIsCriticalHit(),
		DamageTotal:    damageTotal,
		DamageType:     damageType,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogDragonbornBreathWeaponEvent(ctx *core.EventContext, actor core.Entity, target core.Entity, damageTotal int, damageType string, dc int, saveAbility string, saveSuccess bool, saveResult int, listener func(event interface{})) {
	event := &DragonbornBreathWeaponEvent{
		Target:             core.FormatEntityName(target),
		DamageTotal:        damageTotal,
		DamageType:         damageType,
		DC:                 dc,
		SaveAbility:        saveAbility,
		SavingThrowSuccess: saveSuccess,
		SavingThrowResult:  saveResult,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

//func LogMonsterAttackEvent(actor core.Entity, attackResult *core.AttackResult, listener func(event interface{})) {
//	event := &
//}

func LogDamageEvent(ctx *core.EventContext, actor core.Entity, target core.Entity, damageType string, damage int, rolls []int, listener func(event interface{})) {
	event := &DamageEvent{
		Target:     core.FormatEntityName(target),
		Amount:     damage,
		DamageType: damageType,
		Rolls:      rolls,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogSpellChoiceEvent(ctx *core.EventContext, actor core.Entity, choice *core.SpellChoice, status *spells.SpellcastingManagerStatus, listener func(event interface{})) {
	event := &SpellChoiceEvent{
		SpellChoice:   choice,
		ManagerStatus: status,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogSpellAttackEvent(ctx *core.EventContext, actor core.Entity, res core.SpellResult, listener func(event interface{})) {
	var damageTotal int
	if res.GetValueResult() != nil {
		damageTotal = res.GetValueResult().GetTotal()
	} else {
		damageTotal = 0
	}

	event := &SpellAttackEvent{
		Target:             res.GetTargetName(),
		SpellName:          res.GetSpellName(),
		SpellLevel:         res.GetSpellLevel(),
		AttackTotal:        res.GetAttackTotal(),
		AttackModifier:     res.GetAttackTotal() - res.GetAttackRoll(),
		AttackRoll:         res.GetAttackRoll(),
		Success:            res.GetIsHit(),
		CriticalHit:        res.GetIsCriticalHit(),
		DamageTotal:        damageTotal,
		DamageType:         res.GetDamageType().String(),
		HasDC:              res.GetHasDC(),
		DCAbility:          res.GetSpellSaveAbility().String(),
		SaveEffect:         res.GetSpellSaveEffect().String(),
		DCValue:            res.GetTargetDCValue(),
		SavingThrowSuccess: res.GetSpellSaveSuccess(),
		SavingThrowResult:  res.GetSpellSaveTotal(),
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogSpellDCEvent(ctx *core.EventContext, actor core.Entity, target core.Entity, spell *spells.Spell, dc int, save int, isHit bool, listener func(event interface{})) {
	event := &SpellDCEvent{
		Target:      core.FormatEntityName(target),
		SpellChoice: spell,
		DC:          dc,
		SavingThrow: save,
		Success:     isHit,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogSpellHealEvent(ctx *core.EventContext, actor core.Entity, res core.SpellResult, listener func(event interface{})) {
	event := &HealEvent{
		Target:     res.GetTargetName(),
		Name:       res.GetSpellName(),
		SpellLevel: res.GetSpellLevel(),
		IsSpell:    true,
		HealTotal:  res.GetValueResult().GetTotal(),
		HealRolls:  res.GetValueResult().GetFinalRolls(),
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogLayOnHandsHealEvent(ctx *core.EventContext, actor core.Entity, subject core.Entity, value int, listener func(event interface{})) {
	event := &HealEvent{
		Target:     core.FormatEntityName(subject),
		Name:       "Lay on Hands",
		SpellLevel: 0,
		IsSpell:    false,
		HealTotal:  value,
		HealRolls:  []int{value},
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogDamageModifiedEvent(ctx *core.EventContext, actor core.Entity, subject core.Entity, res core.DamageModificationResult, sourceRollID string, listener func(event interface{})) {
	event := &DamageModifiedEvent{
		BaseEvent:        BaseEvent{},
		SubjectName:      core.FormatEntityName(subject),
		OriginalValue:    res.OriginalValue,
		FinalValue:       res.FinalValue,
		WasModified:      res.WasModified,
		ResistanceType:   res.ResistanceType,
		ResistanceBroken: res.ResistanceBroken,
		SourceRollID:     sourceRollID,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogHPModifiedEvent(ctx *core.EventContext, actor core.Entity, subject core.Entity, res core.HPModificationResult, sourceRollID string, listener func(event interface{})) {
	event := &HPModifiedEvent{
		BaseEvent:         BaseEvent{},
		SubjectName:       core.FormatEntityName(subject),
		ModificationValue: res.GetModificationValue(),
		OriginalHP:        res.GetOriginalHP(),
		OriginalTempHP:    res.GetOriginalTempHP(),
		NewHP:             res.GetNewHP(),
		NewTempHP:         res.GetNewTempHP(),
		DidHealHP:         res.GetDidHealHP(),
		DidHealTempHP:     res.GetDidHealTempHP(),
		DidTempDamage:     res.GetDidTempDamage(),
		DidHPDamage:       res.GetDidHPDamage(),
		IsUnconscious:     res.GetIsUnconscious(),
		IsMaxHealth:       res.GetIsMaxHealth(),
		SourceRollID:      sourceRollID,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogHPRollEvent(ctx *core.EventContext, actor core.Entity, rollSum int, rolls []int, toAdd int, listener func(event interface{})) {
	event := &HPRollEvent{
		Value:    rollSum,
		Rolls:    rolls,
		Modifier: toAdd,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogDiceRollEvent(ctx *core.EventContext, actor core.Entity, res core.RollResult, listener func(event interface{})) {
	event := &DiceRollEvent{
		RollType:       res.GetDiceRollType(),
		NumberOfDice:   res.GetNumberOfDice(),
		Die:            res.GetDiceType(),
		FinalRollValue: res.GetFinalRollValue(),
		FinalRolls:     res.GetFinalRolls(),
		Modifier:       res.GetModifier(),
		Total:          res.GetTotal(),
		Advantage:      res.GetAdvantage(),
		OriginalRolls:  res.GetOriginalRolls(),
		RerollEvents:   res.GetRerollEvents(),
		WasRerolled:    res.GetWasRerolled(),
		IsCritical:     res.GetIsCritical(),
		IsNaturalOne:   res.GetIsNaturalOne(),
		IsSuccess:      res.GetIsSuccess(),
		TargetValue:    res.GetTargetValue(),
	}
	setupBaseEvent(ctx, actor, event)
	res.SetID(event.GetID()) // sync the generated event id back to the roll result

	if listener != nil {
		listener(event)
	}
}

func LogSavingThrowEvent(ctx *core.EventContext, actor core.Entity, result int, roll int, modifier int, success bool, listener func(event interface{})) {
	event := &SavingThrowEvent{
		Result:   result,
		Roll:     roll,
		Modifier: modifier,
		Success:  success,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogTargetChoiceEvent(ctx *core.EventContext, actor core.Entity, target core.Entity, score float64, factors map[DecisionFactor]float64, listener func(event interface{})) {
	event := &TargetChoiceEvent{
		Target:  core.FormatEntityName(target),
		Score:   score,
		Factors: factors,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogSpecialAbilityEvent(ctx *core.EventContext, actor core.Entity, abilityName string, description string, targetName string, value int, listener func(event interface{})) {
	event := &SpecialAbilityEvent{
		AbilityName: abilityName,
		Description: description,
		Target:      targetName,
		Value:       value,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogCombatEventMessage(ctx *core.EventContext, actor core.Entity, message string, listener func(event interface{})) {
	event := &CombatEventMessage{
		Message: message,
	}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogDeathEvent(ctx *core.EventContext, actor core.Entity, listener func(event interface{})) {
	event := &DeathEvent{}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}

func LogUnconsciousEvent(ctx *core.EventContext, actor core.Entity, listener func(event interface{})) {
	event := &UnconsciousEvent{}
	setupBaseEvent(ctx, actor, event)

	if listener != nil {
		listener(event)
	}
}
