package events

import (
	"fmt"
)

// EventDispatcher dispatches events to registered listeners.
type EventDispatcher struct {
	listeners []CombatListener
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners: []CombatListener{},
	}
}

func (d *EventDispatcher) RegisterListener(listener CombatListener) {
	d.listeners = append(d.listeners, listener)
}

func (d *EventDispatcher) DispatchEvent(event CombatEvent) {
	for _, listener := range d.listeners {
		if listener != nil {
			listener.HandleEvent(event)
		} else {
			fmt.Errorf("Listener is nil")
		}
	}
}
