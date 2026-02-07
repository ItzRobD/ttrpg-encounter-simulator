package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
	"strings"
)

// CombatLogger handles human-readable formatting of the hierarchical timeline.
type CombatLogger struct {
	indentSize int
}

func NewCombatLogger() *CombatLogger {
	return &CombatLogger{
		indentSize: 2,
	}
}

// PrintTimeline prints a formatted view of the combat events.
func (cl *CombatLogger) PrintTimeline(timeline []events.TimelineEvent) {
	// Build a map for quick lookup and to track levels
	eventMap := make(map[string]*events.TimelineEvent)
	for i := range timeline {
		eventMap[timeline[i].ID] = &timeline[i]
	}

	// Calculate indentation levels
	levels := make(map[string]int)
	for _, event := range timeline {
		levels[event.ID] = cl.calculateLevel(event, eventMap)
	}

	fmt.Println("--- Combat Simulation Log ---")
	for _, event := range timeline {
		indent := strings.Repeat(" ", levels[event.ID]*cl.indentSize)
		msg := cl.formatEvent(event)
		if msg != "" {
			fmt.Printf("%s%s\n", indent, msg)
		}
	}
	fmt.Println("--- End of Log ---")
}

func (cl *CombatLogger) calculateLevel(event events.TimelineEvent, eventMap map[string]*events.TimelineEvent) int {
	level := 0
	curr := event
	for curr.ParentID != "" {
		parent, ok := eventMap[curr.ParentID]
		if !ok {
			break
		}
		level++
		curr = *parent
	}
	return level
}

func (cl *CombatLogger) formatEvent(event events.TimelineEvent) string {
	actorName := "Unknown"
	if event.Actor != nil {
		actorName = event.Actor.Name
	}

	switch event.Type {
	case events.EventInitiative:
		if event.Actor == nil {
			return "[initiative] Rolling for all actors..."
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			return fmt.Sprintf("[initiative] %s", actorName)
		}

		if data["is_lair"] != nil && data["is_lair"].(bool) {
			return fmt.Sprintf("[initiative] %s: 20 (Lair)", actorName)
		}

		rollRes, ok := data["roll"].(*roll_manager.RollResult)
		if !ok {
			// Try map if coming from JSON
			if rollMap, ok := data["roll"].(map[string]interface{}); ok {
				total := int(rollMap["total"].(float64))
				return fmt.Sprintf("[initiative] %s: %v", actorName, total)
			}
			if total, ok := data["total"].(int); ok {
				return fmt.Sprintf("[initiative] %s: %v", actorName, total)
			}
			return fmt.Sprintf("[initiative] %s", actorName)
		}

		return fmt.Sprintf("[initiative] %s: %v %s", actorName, rollRes.Total, cl.formatRollDetails(rollRes))

	case events.EventTurnStart:
		return fmt.Sprintf(">>> Turn Start: %s (Round %d)", actorName, event.Round)
	case events.EventTurnEnd:
		return "" // Often redundant in console
	case events.EventRecharge:
		data := event.Data.(map[string]interface{})
		status := "Failed"
		if data["success"].(bool) {
			status = "Success"
		}
		return fmt.Sprintf("[Recharge] %s: %v (Roll: %v) - %s", actorName, data["action"], data["roll"], status)
	case events.EventDecisionStart:
		data := event.Data.(map[string]interface{})
		return fmt.Sprintf("[Decision] %s: %v", actorName, data["decision"])
	case events.EventActionStart:
		data := event.Data.(map[string]interface{})
		return fmt.Sprintf("[Action] %s uses %v", actorName, data["action_name"])
	case events.EventResolution:
		data, ok := event.Data.(map[string]interface{})
		if ok && data["is_aoe"] != nil && data["is_aoe"].(bool) {
			return "" // Hide per-target AOE scope to avoid clutter, or maybe show it?
		}
		return "" // Structural container
	case events.EventAttackRoll:
		data := event.Data.(map[string]interface{})
		res := "Miss"
		rollTotal := 0

		// Handle both map[string]interface{} and *roll_manager.RollResult
		if rollMap, ok := data["roll"].(map[string]interface{}); ok {
			rollTotal = cl.toInt(rollMap["total"])
		} else if rollRes, ok := data["roll"].(*roll_manager.RollResult); ok {
			rollTotal = rollRes.Total
		}

		if data["is_hit"].(bool) {
			res = "Hit"
			if data["is_critical"].(bool) {
				res = "CRITICAL HIT"
			}
		}
		return fmt.Sprintf("[Attack] %s vs Target %v: %v (%s) %s", actorName, data["target_id"], rollTotal, res, cl.formatRollDetailsFromData(data["roll"]))
	case events.EventSavingThrow:
		data := event.Data.(map[string]interface{})
		res := "Failed"
		if data["save_success"].(bool) {
			res = "Success"
		}
		rollTotal := 0
		if roll, ok := data["roll"].(map[string]interface{}); ok {
			rollTotal = cl.toInt(roll["total"])
		} else if rollRes, ok := data["roll"].(*roll_manager.RollResult); ok {
			rollTotal = rollRes.Total
		}
		return fmt.Sprintf("[Save] %s vs %v DC %v: %s (Total: %v) %s", actorName, data["ability"], data["dc"], res, rollTotal, cl.formatRollDetailsFromData(data["roll"]))
	case events.EventOutcome:
		return "" // Structural container
	case events.EventDamageRoll:
		data := event.Data.(map[string]interface{})
		rollTotal := 0
		if rollMap, ok := data["roll"].(map[string]interface{}); ok {
			rollTotal = cl.toInt(rollMap["total"])
		} else if rollRes, ok := data["roll"].(*roll_manager.RollResult); ok {
			rollTotal = rollRes.Total
		}
		return fmt.Sprintf("[Damage] %v %v (Total: %v) %s", actorName, data["damage_type"], rollTotal, cl.formatRollDetailsFromData(data["roll"]))
	case events.EventHPModified:
		data := event.Data.(map[string]interface{})
		var oldHP, newHP, tempHP int

		if resMap, ok := data["result"].(map[string]interface{}); ok {
			oldHP = cl.toInt(resMap["old_hp"])
			newHP = cl.toInt(resMap["new_hp"])
			tempHP = cl.toInt(resMap["temp_hp_used"])
		} else if resMap, ok := data["result"].(core.HPModificationResult); ok {
			oldHP = resMap.OriginalHP
			newHP = resMap.NewHP
			tempHP = resMap.TempHPUsed
		}

		hpStr := fmt.Sprintf("%v -> %v", oldHP, newHP)
		if tempHP > 0 {
			hpStr += fmt.Sprintf(" (Temp HP used: %v)", tempHP)
		}
		return fmt.Sprintf("[HP] %s: %s", actorName, hpStr)
	case events.EventFeatureTrigger:
		data := event.Data.(map[string]interface{})
		return fmt.Sprintf("[Feature] %s triggered %v", actorName, data["feature"])
	case events.EventDeath:
		return fmt.Sprintf("!!! %s has died !!!", actorName)
	case events.EventUnconscious:
		return fmt.Sprintf("!!! %s falls unconscious !!!", actorName)
	case events.EventDamageModified:
		data := event.Data.(map[string]interface{})
		resType := fmt.Sprintf("%v", data["resistance_type"])
		dt := fmt.Sprintf("%v", data["damage_type"])

		orig := cl.toInt(data["original_value"])
		final := cl.toInt(data["final_value"])
		return fmt.Sprintf("  [Resist] %s %v: %v -> %v (%s)", actorName, dt, orig, final, resType)
	case events.EventVictory:
		data := event.Data.(map[string]interface{})
		return fmt.Sprintf("=== VICTORY: %v Side Wins! (Rounds: %v) ===", data["winner"], data["rounds"])
	case events.EventMessage:
		data := event.Data.(map[string]interface{})
		return fmt.Sprintf("> %s", data["message"])
	default:
		return fmt.Sprintf("[%s] %s", event.Type, actorName)
	}
}

func (cl *CombatLogger) toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	default:
		return 0
	}
}

func (cl *CombatLogger) formatRollDetailsFromData(rollData interface{}) string {
	if rollRes, ok := rollData.(*roll_manager.RollResult); ok {
		return cl.formatRollDetails(rollRes)
	}
	if rollMap, ok := rollData.(map[string]interface{}); ok {
		// Reconstruct basic info from map if possible, or just skip for now
		// For brevity in console, if we don't have the full object, we might just show total which is already shown.
		// But we can try to extract FinalRolls
		if rolls, ok := rollMap["final_rolls"].([]interface{}); ok {
			strRolls := make([]string, len(rolls))
			for i, r := range rolls {
				strRolls[i] = fmt.Sprintf("%v", r)
			}
			mod := cl.toInt(rollMap["modifier"])
			modStr := ""
			if mod > 0 {
				modStr = fmt.Sprintf("+%d", mod)
			} else if mod < 0 {
				modStr = fmt.Sprintf("%d", mod)
			}
			return fmt.Sprintf("(%s)%s", strings.Join(strRolls, "+"), modStr)
		}
	}
	return ""
}

func (cl *CombatLogger) formatRollDetails(res *roll_manager.RollResult) string {
	if res == nil {
		return ""
	}

	diceStr := ""
	if len(res.FinalRolls) > 0 {
		strRolls := make([]string, len(res.FinalRolls))
		for i, r := range res.FinalRolls {
			strRolls[i] = fmt.Sprintf("%d", r)
		}
		diceStr = strings.Join(strRolls, "+")
	}

	modStr := ""
	if res.Modifier > 0 {
		modStr = fmt.Sprintf("+%d", res.Modifier)
	} else if res.Modifier < 0 {
		modStr = fmt.Sprintf("%d", res.Modifier)
	}

	rerollStr := ""
	if len(res.RerollEvents) > 0 {
		var events []string
		for _, e := range res.RerollEvents {
			events = append(events, fmt.Sprintf("%d->%d (%s)", e.OriginalRoll, e.NewRoll, e.Reason))
		}
		rerollStr = fmt.Sprintf(" [Reroll: %s]", strings.Join(events, ", "))
	}

	advStr := ""
	if res.Advantage == core.RollAdvantage {
		advStr = " [Adv]"
	} else if res.Advantage == core.RollDisadvantage {
		advStr = " [Disadv]"
	}

	return fmt.Sprintf("(%s)%s%s%s", diceStr, modStr, advStr, rerollStr)
}
