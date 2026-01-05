package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
)

type LairAI struct {
	parent    *Lair
	combatCtx *core.CombatContext
	eventCtx  *core.EventContext
}

func NewLairAI(l *Lair) *LairAI { return &LairAI{parent: l} }

func (lai *LairAI) UpdateCombatContext(ctx *core.CombatContext) { lai.combatCtx = ctx }
func (lai *LairAI) UpdateEventContext(ctx *core.EventContext)   { lai.eventCtx = ctx }

// BuildLairActionRequest chooses the first available action (respecting recharge)
// and selects a target according to that action's TargetSide/TargetPolicy.
func (lai *LairAI) BuildLairActionRequest() (*core.AIRequest, error) {
	if lai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set for lair")
	}
	if len(lai.parent.actionManager.Actions) == 0 {
		return nil, fmt.Errorf("no lair actions configured")
	}

	// Pick first available action
	var actionIndex int = -1
	var action LairAction
	for idx, a := range lai.parent.actionManager.Actions {
		if lai.parent.actionManager.IsActionAvailable(idx) {
			actionIndex = idx
			action = a
			break
		}
	}
	if actionIndex == -1 {
		return nil, fmt.Errorf("no lair actions available")
	}

	// Build target map according to action's TargetSide
	candidates := make(map[int]*core.Combatant)
	for id, ci := range lai.combatCtx.CombatantInfo {
		if ci == nil || ci.Combatant == nil || ci.Combatant.IsLair {
			continue
		}
		e := ci.Combatant.GetEntity()
		if action.TargetSide == TargetCharacters && e.IsCharacter() && !e.IsUnconscious() && !e.IsDead() {
			candidates[id] = ci.Combatant
		} else if action.TargetSide == TargetMonsters && e.IsMonster() && !e.IsUnconscious() && !e.IsDead() {
			candidates[id] = ci.Combatant
		}
	}
	if len(candidates) == 0 {
		lai.parent.LogEvent(events.ECombatEventMessage, "No valid lair targets")
		return nil, nil
	}

	// Select representative target for logging/UI; AOE effects will apply to all candidates later
	priority := action.TargetPolicy
	tStatus, targetID, err := core.SelectTargetFromMap(candidates, priority, lai.parent.GetRNG())
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		lai.parent.LogEvent(events.ECombatEventMessage, "No valid lair targets")
		return nil, nil
	}
	target := candidates[targetID].GetEntity()

	req := &core.AIRequest{
		Actor:       lai.parent,
		ActorType:   core.EntityMonster,
		TargetID:    targetID,
		Target:      target,
		ActionType:  core.ATLairAction,
		ActionIndex: actionIndex,
		Advantage:   core.RollNormal,
		Request:     core.AIReqNormalAction,
	}

	// Structured logging: chosen action and target
	lai.parent.LogEvent(events.ECombatEventMessage, fmt.Sprintf("Lair chooses lair action: %s", action.Name))
	lai.parent.LogEvent(events.ETTargetChoiceEvent, &events.TargetChoiceData{
		Target:  target,
		Score:   1.0,
		Factors: nil,
	})

	return req, nil
}
