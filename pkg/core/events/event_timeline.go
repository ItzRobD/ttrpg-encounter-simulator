package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
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
	case *ActionChoiceEvent: // parent id is the sequence id, the target choice parent id is the action id, therefore parent id
		te.Type = TimelineChoiceType
		te.ParentID = te.SequenceID // override parent id to be the sequence id, this is the start of action logic
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
	case *MeleeAttackEvent:
		te.Type = TimelineAttackType
		target := TimelineEntity{Name: e.Target}
		if e.GetTargetEntity() != nil {
			target = mapTimelineEntity(e.GetTargetEntity())
		}
		te.Data = TimelineAttack{
			Actor:      mapTimelineEntity(e.GetActor()),
			Target:     target,
			AttackType: "melee",
			DiceRoll:   e,
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
			Actor:      mapTimelineEntity(e.GetActor()),
			Target:     target,
			AttackType: "spell",
			DiceRoll:   e,
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
			te.Type = "initiative"
			te.Data = TimelineRoll{
				Actor: mapTimelineEntity(e.GetActor()),
				Roll:  e,
			}
		} else {
			te.Type = TimelineDamageType
			te.Data = TimelineRoll{
				Actor: mapTimelineEntity(e.GetActor()),
				Roll:  e,
			}
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
			OriginalHP:     e.OriginalHP,
			FinalHP:        e.NewHP,
			OriginalTempHP: e.OriginalTempHP,
			FinalTempHP:    e.NewTempHP,
			SourceRollID:   e.SourceRollID,
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
			SourceRollID:     e.SourceRollID,
			OriginalValue:    e.OriginalValue,
			FinalValue:       e.FinalValue,
			WasModified:      e.WasModified,
			ResistanceType:   e.ResistanceType,
			ResistanceBroken: e.ResistanceBroken,
			Note:             fmt.Sprintf("Original: %d, Final: %d", e.OriginalValue, e.FinalValue),
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
	}

	return te
}

func makeTimelineScores(utilityScore float64, factors map[DecisionFactor]float64, topReasons []DecisionFactor) TimelineScores {
	return TimelineScores{
		UtilityScore: utilityScore,
		Factors:      factors,
		TopReasons:   topReasons,
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
