package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
)

type LairAI struct {
	parent    *Lair
	combatCtx *core.CombatContext
}

func NewLairAI(l *Lair) *LairAI { return &LairAI{parent: l} }

func (lai *LairAI) UpdateCombatContext(ctx *core.CombatContext) { lai.combatCtx = ctx }

// BuildLairActionRequest picks the first configured action and targets a character.
func (lai *LairAI) BuildLairActionRequest() (*core.AIRequest, error) {
	if lai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set for lair")
	}
	if len(lai.parent.actionManager.Actions) == 0 {
		return nil, fmt.Errorf("no lair actions configured")
	}

	// Select first action index deterministically
	var actionIndex int = -1
	for idx := range lai.parent.actionManager.Actions {
		actionIndex = idx
		break
	}
	if actionIndex == -1 {
		return nil, fmt.Errorf("no lair actions available")
	}

	// Build target map of enemies: choose any character (treat lair as hostile to characters)
	enemies := make(map[int]*core.Combatant)
	for id, ci := range lai.combatCtx.CombatantInfo {
		if ci == nil || ci.Combatant == nil || ci.Combatant.IsLair {
			continue
		}
		e := ci.Combatant.GetEntity()
		if e.IsCharacter() && !e.IsUnconscious() && !e.IsDead() {
			enemies[id] = ci.Combatant
		}
	}
	if len(enemies) == 0 {
		return nil, fmt.Errorf("no valid lair targets")
	}

	targetID, err := core.SelectTargetFromMap(enemies, core.PrioritizeLowestMaxHP, lai.parent.GetRNG())
	if err != nil {
		return nil, err
	}
	target := enemies[targetID].GetEntity()

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
	events.LogCombatEventMessage(lai.parent, "Lair chooses lair action", lai.parent.GetEventListener())
	events.LogTargetChoiceEvent(lai.parent, target, lai.parent.GetEventListener())

	return req, nil
}
