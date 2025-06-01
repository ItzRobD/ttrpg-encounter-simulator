package events

import (
	"fmt"
)

// EventDispatcher dispatches events to registered handlers.
type EventDispatcher struct {
	handlers []EventHandler
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: []EventHandler{},
	}
}

func (d *EventDispatcher) RegisterHandler(handler EventHandler) {
	d.handlers = append(d.handlers, handler)
}

func (d *EventDispatcher) DispatchEvent(event CombatEvent) {
	for _, listener := range d.handlers {
		if listener != nil {
			listener.HandleEvent(event)
		} else {
			fmt.Errorf("Listener is nil")
		}
	}
}
