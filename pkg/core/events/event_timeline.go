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
			Target:     TimelineEntity{Name: e.Target},
			ChoiceType: "target",
			Scores:     makeTimelineScores(e.Score, e.Factors, nil),
		}
	case *SpellChoiceEvent:
		te.Type = TimelineChoiceType
		choiceStr := e.SpellChoice.GetSpell().GetName()
		te.Data = TimelineChoice{
			Actor:      mapTimelineEntity(e.GetActor()),
			ChoiceType: "spell",
			Choice:     &choiceStr,
		}
	case *MeleeAttackEvent:
		te.Type = TimelineAttackType
		te.Data = TimelineAttack{
			Actor:      mapTimelineEntity(e.GetActor()),
			Target:     TimelineEntity{Name: e.Target},
			AttackType: "melee",
			DiceRoll:   e,
		}
	case *SpellAttackEvent:
		te.Type = TimelineAttackType
		te.Data = TimelineAttack{
			Actor:      mapTimelineEntity(e.GetActor()),
			Target:     TimelineEntity{Name: e.Target},
			AttackType: "spell",
			DiceRoll:   e,
		}
	case *DiceRollEvent:
		if e.RollType == core.DiceRollSavingThrow || e.RollType == core.DiceRollDeathSavingThrow {
			te.Type = TimelineSaveType
			te.Data = TimelineSavingThrow{
				Actor:    mapTimelineEntity(e.GetActor()),
				Target:   TimelineEntity{Name: e.Name}, // e.Name is often target name for saves
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
		te.Data = TimelineEffect{
			Actor:        mapTimelineEntity(e.GetActor()),
			Target:       TimelineEntity{Name: e.SubjectName},
			Type:         core.EffectDamage,
			Value:        e.ModificationValue,
			OriginalHP:   e.OriginalHP,
			FinalHP:      e.NewHP,
			SourceRollID: e.SourceRollID,
		}
	case *DamageModifiedEvent:
		te.Type = TimelineDamageModifiedType
		te.Data = TimelineEffect{
			Actor:        mapTimelineEntity(e.GetActor()),
			Target:       TimelineEntity{Name: e.SubjectName},
			SourceRollID: e.SourceRollID,
			Note:         fmt.Sprintf("Original: %d, Final: %d", e.OriginalValue, e.FinalValue),
		}
	case *HealEvent:
		te.Type = TimelineHealType
		te.Data = TimelineEffect{
			Actor:  mapTimelineEntity(e.GetActor()),
			Target: TimelineEntity{Name: e.Target},
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
	return TimelineEntity{
		Name:       e.GetName(),
		InstanceID: e.GetInstanceID(),
		Type:       e.GetEntityType(),
	}
}

func eventTypeToStringPtr(t EventType) *string {
	return (*string)(&t)
}
