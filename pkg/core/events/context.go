package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

type EventContext struct {
	sequenceID  string
	parentStack []string
	parentID    string
	currentID   string
}

func NewEventContext() *EventContext {
	return &EventContext{
		sequenceID:  "",
		parentStack: make([]string, 0),
		parentID:    "",
		currentID:   "",
	}
}

func (ctx *EventContext) GetSequenceID() string { return ctx.sequenceID }
func (ctx *EventContext) GenerateSequenceID()   { ctx.sequenceID = core.NewUUIDv7() }
func (ctx *EventContext) GetParentID() string   { return ctx.parentID }
func (ctx *EventContext) SetParentID(id string) { ctx.parentID = id }
func (ctx *EventContext) GenerateParentID()     { ctx.parentID = core.NewUUIDv7() }
func (ctx *EventContext) GetCurrentID() string  { return ctx.currentID }
func (ctx *EventContext) GenerateCurrentID()    { ctx.currentID = core.NewUUIDv7() }

func (ctx *EventContext) PushParent(id string) {
	if ctx.parentID != "" {
		ctx.parentStack = append(ctx.parentStack, ctx.parentID)
	}
	ctx.parentID = id
}

func (ctx *EventContext) PopParent() {
	if len(ctx.parentStack) > 0 {
		ctx.parentID = ctx.parentStack[len(ctx.parentStack)-1]
		ctx.parentStack = ctx.parentStack[:len(ctx.parentStack)-1]
	} else {
		ctx.parentID = ""
	}
}

func (ctx *EventContext) AdvanceScope() {
	ctx.parentID = ctx.currentID
	ctx.GenerateCurrentID()
}
