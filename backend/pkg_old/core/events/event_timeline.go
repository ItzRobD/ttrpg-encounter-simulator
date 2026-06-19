package events

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"fmt"
	"strings"
)

func MakeTimelineEvent(event CombatEvent) *TimelineEvent {
	seqID := ""
	parentID := ""
	if event.Context() != nil {
		seqID = event.Context().GetSequenceID()
		parentID = event.Context().GetParentID()
	}

	te := &TimelineEvent{
		Timestamp:  event.GetTimestamp(),
		ID:         event.GetID(),
		SequenceID: seqID,
		ParentID:   parentID,
		Round:      event.GetRound(),
		Type:       "",
		Data:       nil,
	}

	switch e := event.(type) {
	case *TurnStartEvent:
		te.Type = TimelineTurnStartType
		te.ParentID = te.SequenceID
		te.Data = TimelineTurnStart{
			Actor: mapTimelineEntity(e.GetActor()),
		}
	case *ActionChoiceEvent: // parent id is the sequence id, the target choice parent id is the action id, therefore parent id
		te.Type = TimelineChoiceType
		choiceStr := e.ActionChoice.String()
		te.Data = TimelineChoice{
			Actor:      mapTimelineEntity(e.GetActor()),
			ChoiceType: "action",
			Choice:     &choiceStr,
			Scores:     makeTimelineScores(e.UtilityScore, e.Factors, e.TopReasons),
		}
	case *TargetChoiceEvent:
		te.Type = TimelineChoiceType
		te.Data = TimelineChoice{
			Actor:      mapTimelineEntity(e.GetActor()),
			Target:     mapTimelineEntity(e.GetTargetEntity()),
			ChoiceType: "target",
			Scores:     makeTimelineScores(e.Score, e.Factors, nil),
		}
	case *SpellChoiceEvent:
		te.Type = TimelineChoiceType
		choiceStr := e.SpellChoice.GetSpell().GetName()
		te.Data = TimelineChoice{
			Actor:      mapTimelineEntity(e.GetActor()),
			Target:     mapTimelineEntity(e.GetTargetEntity()),
			ChoiceType: "spell",
			Choice:     &choiceStr,
		}
	case *MartialAttackEvent:
		te.Type = TimelineAttackType
		target := TimelineEntity{Name: e.Target}
		if e.GetTargetEntity() != nil {
			target = mapTimelineEntity(e.GetTargetEntity())
		}
		attackType := "melee"
		if e.IsRanged {
			attackType = "ranged"
		}
		te.Data = TimelineAttack{
			Actor:        mapTimelineEntity(e.GetActor()),
			Target:       target,
			AttackType:   attackType,
			DiceRoll:     e,
			ActionDetail: e.ActionDetail,
		}
	case *SpellAttackEvent:
		// If the spell has a DC, we skip logging it as an "attack" type event in the timeline
		// to avoid redundancy with the saving throw and damage roll events.
		// For attack roll spells, it's still needed.
		if e.HasDC {
			return nil
		}
		te.Type = TimelineAttackType
		target := TimelineEntity{Name: e.Target}
		if e.GetTargetEntity() != nil {
			target = mapTimelineEntity(e.GetTargetEntity())
		}
		te.Data = TimelineAttack{
			Actor:        mapTimelineEntity(e.GetActor()),
			Target:       target,
			AttackType:   "spell",
			DiceRoll:     e,
			ActionDetail: e.ActionDetail,
		}
	case *DiceRollEvent:
		if e.RollType == core.DiceRollSavingThrow || e.RollType == core.DiceRollDeathSavingThrow {
			te.Type = TimelineSaveType
			target := TimelineEntity{Name: e.Name}
			if e.GetTargetEntity() != nil {
				target = mapTimelineEntity(e.GetTargetEntity())
			}
			te.Data = TimelineSavingThrow{
				Actor:    mapTimelineEntity(e.GetActor()),
				Target:   target,
				DC:       e.TargetValue,
				DiceRoll: e,
			}
		} else if e.RollType == core.DiceRollInitiative {
			te.Type = TimelineInitiativeType
			te.Data = TimelineRoll{
				Actor: mapTimelineEntity(e.GetActor()),
				Roll:  e,
			}
		} else {
			te.Type = TimelineDamageType
			target := TimelineEntity{}
			if e.GetTargetEntity() != nil {
				target = mapTimelineEntity(e.GetTargetEntity())
			}
			te.Data = TimelineRoll{
				Actor:      mapTimelineEntity(e.GetActor()),
				Target:     target,
				Roll:       e,
				DamageType: e.DamageType,
			}
		}
	case *EquipmentEvent:
		te.Type = TimelineEquipmentType
		if e.GetActor() != nil && e.GetActor().IsMonster() {
			te.Type = TimelineActionDetailType
		}
		te.Data = TimelineEquipment{
			Actor:        mapTimelineEntity(e.GetActor()),
			Name:         e.Name,
			NumberOfDice: e.NumberOfDice,
			Die:          e.Die,
			DamageType:   e.DamageType,
			AttackBonus:  e.AttackBonus,
			DamageBonus:  e.DamageBonus,
			IsRanged:     e.IsRanged,
			Properties:   e.Properties,
			Modifiers:    e.Modifiers,
		}
	case *HPModifiedEvent:
		te.Type = TimelineHPModifiedType
		target := TimelineEntity{Name: e.SubjectName}
		if e.GetSubjectEntity() != nil {
			target = mapTimelineEntity(e.GetSubjectEntity())
		}
		te.Data = TimelineEffect{
			Actor:          mapTimelineEntity(e.GetActor()),
			Target:         target,
			Type:           core.EffectDamage,
			Value:          e.ModificationValue,
			DamageType:     e.DamageType,
			OriginalHP:     e.OriginalHP,
			FinalHP:        e.NewHP,
			OriginalTempHP: e.OriginalTempHP,
			FinalTempHP:    e.NewTempHP,
			SourceRollID:   e.SourceRollID,
		}
		// Force parent to the originating damage/heal roll when available
		if e.SourceRollID != "" {
			te.ParentID = e.SourceRollID
		}
	case *DamageModifiedEvent:
		te.Type = TimelineDamageModifiedType
		target := TimelineEntity{Name: e.SubjectName}
		if e.GetSubjectEntity() != nil {
			target = mapTimelineEntity(e.GetSubjectEntity())
		}
		te.Data = TimelineEffect{
			Actor:            mapTimelineEntity(e.GetActor()),
			Target:           target,
			Type:             core.EffectDamage,
			Value:            e.FinalValue,
			DamageType:       e.DamageType,
			SourceRollID:     e.SourceRollID,
			OriginalValue:    e.OriginalValue,
			FinalValue:       e.FinalValue,
			WasModified:      e.WasModified,
			ResistanceType:   e.ResistanceType,
			ResistanceBroken: e.ResistanceBroken,
			Note:             fmt.Sprintf("Original: %d, Final: %d", e.OriginalValue, e.FinalValue),
		}
		// Force parent to the originating damage roll when available
		if e.SourceRollID != "" {
			te.ParentID = e.SourceRollID
		}
	case *HealEvent:
		te.Type = TimelineHealType
		target := TimelineEntity{Name: e.Target}
		if e.GetTargetEntity() != nil {
			target = mapTimelineEntity(e.GetTargetEntity())
		}
		te.Data = TimelineEffect{
			Actor:  mapTimelineEntity(e.GetActor()),
			Target: target,
			Type:   core.EffectHealing,
			Value:  e.HealTotal,
			Note:   e.Name,
		}
	case *DeathEvent:
		te.Type = TimelineDeathType
		te.Data = TimelineEffect{
			Actor:  mapTimelineEntity(e.GetActor()),
			Target: mapTimelineEntity(e.GetActor()), // Subject of death is the actor of the event
			Type:   core.EffectCondition,
			Note:   "Died",
		}
	case *UnconsciousEvent:
		te.Type = TimelineUnconsciousType
		te.Data = TimelineEffect{
			Actor:  mapTimelineEntity(e.GetActor()),
			Target: mapTimelineEntity(e.GetActor()),
			Type:   core.EffectCondition,
			Note:   "Unconscious",
		}
	case *ConditionEvent:
		te.Type = TimelineConditionType
		note := "Added"
		if !e.IsAdded {
			note = "Removed"
		}
		te.Data = TimelineEffect{
			Actor:     mapTimelineEntity(e.GetActor()),
			Target:    mapTimelineEntity(e.GetActor()),
			Type:      core.EffectCondition,
			Condition: &e.Condition,
			Note:      note,
		}
	case *VictoryEvent:
		te.Type = TimelineVictoryType
		te.ParentID = "" // Always root
		te.Data = TimelineVictory{
			Winner: e.WinningSide,
			Rounds: e.Rounds,
		}
	case *CombatEventMessage:
		msg := e.Message
		lowerMsg := strings.ToLower(msg)
		if strings.Contains(lowerMsg, "multiattack") {
			te.Type = TimelineMultiattackType
		} else if strings.Contains(lowerMsg, " action ") || strings.Contains(lowerMsg, " uses ") {
			te.Type = TimelineActionType
		} else {
			te.Type = TimelineMessageType
		}
		te.Data = TimelineMessage{
			Actor:   mapTimelineEntity(e.GetActor()),
			Message: msg,
		}
	}

	return te
}

func makeTimelineScores(utilityScore float64, factors map[DecisionFactor]float64, topReasons []DecisionFactor) TimelineScores {
	stringFactors := make(map[string]float64)
	for k, v := range factors {
		stringFactors[string(k)] = v
	}

	stringTopFactors := make([]string, len(topReasons))
	for i, r := range topReasons {
		stringTopFactors[i] = string(r)
	}

	return TimelineScores{
		UtilityScore: utilityScore,
		Factors:      stringFactors,
		TopFactors:   stringTopFactors,
	}
}

func mapTimelineEntity(e core.Entity) TimelineEntity {
	if e == nil {
		return TimelineEntity{}
	}
	return TimelineEntity{
		Name:       e.GetName(),
		InstanceID: e.GetInstanceID(),
		Type:       e.GetEntityType(),
	}
}

func eventTypeToStringPtr(t EventType) *string {
	return (*string)(&t)
}
