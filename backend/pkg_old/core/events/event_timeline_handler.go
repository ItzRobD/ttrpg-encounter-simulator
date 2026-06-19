package events

type TimelineHandler struct {
	Timeline []TimelineEvent
}

func (h *TimelineHandler) HandleEvent(event CombatEvent) {
	timelineEvent := MakeTimelineEvent(event)
	if timelineEvent != nil {
		h.Timeline = append(h.Timeline, *timelineEvent)
	}
}
