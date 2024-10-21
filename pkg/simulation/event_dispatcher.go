package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/events"
	"fmt"
)

// EventDispatcher dispatches events to registered listeners.
type EventDispatcher struct {
	listeners []events.CombatListener
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners: []events.CombatListener{},
	}
}

func (d *EventDispatcher) RegisterListener(listener events.CombatListener) {
	d.listeners = append(d.listeners, listener)
}

func (d *EventDispatcher) DispatchEvent(event events.CombatEvent) {
	for _, listener := range d.listeners {
		if listener != nil {
			listener.HandleEvent(event)
		} else {
			fmt.Errorf("Listener is nil")
		}
	}
}
