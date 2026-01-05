package events

import "dnd5e-encounter-simulator-backend/pkg/core"

func MakeTimelineEvent(event CombatEvent) *TimelineEvent {
	if event.Context() == nil {
		panic("event context is nil while making timeline event")
	}
	te := &TimelineEvent{
		Timestamp:  event.GetTimestamp(),
		ID:         event.GetID(),
		SequenceID: event.Context().GetSequenceID(),
		ParentID:   event.Context().GetParentID(),
		Type:       "",
		Data:       nil,
	}

	switch e := event.(type) {
	case *ActionChoiceEvent: // parent id is the sequence id, the target choice parent id is the action id, therefore parent id
		te.Type = TimelineChoiceType
		te.ParentID = te.SequenceID // override parent id to be the sequence id, this is the start of action logic
		te.Data = TimelineChoice{
			Actor:      mapTimelineEntity(e.GetActor()),
			ChoiceType: "",
			Choice:     eventTypeToStringPtr(e.GetEventType()),
			Scores:     makeTimelineScores(e.UtilityScore, e.Factors, e.TopReasons),
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
