package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

// LairAction is a richer action descriptor supporting attack or DC, AOE, and recharge.
type LairAction struct {
	Name         string
	Mode         LairActionMode
	TargetSide   TargetSide
	TargetPolicy core.TargetPriority
	IsAOE        bool
	Recharge     int // 0 = none; else target value (e.g., 5 -> recharge 5-6)

	// Attack mode
	AttackData core.AttackData

	// DC mode
	DCAbility core.Ability
	DCValue   int
	OnSuccess core.DCOnSuccess
}

// LairActionManager processes lair actions, including AOE/DC and recharge.
type LairActionManager struct {
	parent      *Lair
	rollManager *roll_manager.RollManager
	Actions     map[int]LairAction // index -> action
	rechargeOK  map[int]bool       // index -> available this round
}

func NewLairActionManager(parent *Lair, rm *roll_manager.RollManager) *LairActionManager {
	return &LairActionManager{
		parent:      parent,
		rollManager: rm,
		Actions:     make(map[int]LairAction),
		rechargeOK:  make(map[int]bool),
	}
}

func (lam *LairActionManager) AddLairAction(index int, a LairAction) {
	lam.Actions[index] = a
	// If action has recharge, mark as available at start (first use allowed)
	if a.Recharge > 0 {
		lam.rechargeOK[index] = true
	}
}

// IsActionAvailable reports whether an action is available to use (recharge ok or no recharge).
func (lam *LairActionManager) IsActionAvailable(index int) bool {
	a, ok := lam.Actions[index]
	if !ok {
		return false
	}
	if a.Recharge <= 0 {
		return true
	}
	return lam.rechargeOK[index]
}

// RollRechargeActions rolls for each expended action with a recharge value.
func (lam *LairActionManager) RollRechargeActions() {
	for idx, a := range lam.Actions {
		if a.Recharge <= 0 {
			continue
		}
		if lam.rechargeOK[idx] {
			continue // already available
		}
		// Recharge on d6 >= a.Recharge
		opts := roll_manager.NewRollOptions()
		opts.RollType = core.DiceRollRecharge
		opts.TargetValue = a.Recharge
		res := lam.rollManager.RollRecharge(opts)
		res.Name = a.Name
		if res.IsSuccess {
			lam.rechargeOK[idx] = true
		}
		events.LogDiceRollEvent(lam.parent, res, lam.parent.GetEventListener())
	}
}

// expend marks a recharge action as used, if applicable.
func (lam *LairActionManager) expend(idx int) {
	if a, ok := lam.Actions[idx]; ok && a.Recharge > 0 {
		lam.rechargeOK[idx] = false
	}
}

// ProcessAttackRequest handles only Attack-mode, single-target actions.
// Retained for compatibility with Lair.ExecuteAIRequest building AttackRequest.
func (lam *LairActionManager) ProcessAttackRequest(req *core.AttackRequest) ([]core.AttackResult, error) {
	var results []core.AttackResult

	for idx, ad := range req.AttackData {
		rollOpts := roll_manager.NewRollOptions()
		rollOpts.Advantage = req.AttackOptions.GetAdvantage()
		rollOpts.Modifier = ad.AttackModifier + req.AttackOptions.GetBonusToAttackRoll()
		rollOpts.CriticalThreshold = 20
		if req.AttackOptions.GetIsImprovedCritical() {
			rollOpts.CriticalThreshold = 19
		}
		rollOpts.RollType = core.DiceRollAttack
		rollOpts.TargetValue = req.Target.GetAC()

		attackRollResult, err := lam.rollManager.RollAttack(rollOpts)
		if err != nil {
			return nil, err
		}

		// Damage
		dOpts := roll_manager.NewRollOptions()
		dOpts.Modifier = ad.DamageModifier + req.AttackOptions.GetBonusToDamageRoll()
		dOpts.RollType = core.DiceRollDamage
		dmgRollResult, err := lam.rollManager.RollDamage(req, idx, attackRollResult.IsCritical, dOpts)
		if err != nil {
			return nil, err
		}

		ar := core.AttackResult{
			ActorName:     lam.parent.GetName(),
			TargetName:    req.Target.GetName(),
			AttackName:    ad.Name,
			AttackCount:   idx + 1,
			TargetValue:   attackRollResult.TargetValue,
			IsHit:         attackRollResult.IsSuccess,
			IsCriticalHit: attackRollResult.IsCritical,
			AttackTotal:   attackRollResult.Total,
			AttackRoll:    attackRollResult.FinalRollValue,
			DamageRoll:    dmgRollResult,
			DamageType:    ad.DamageType,
		}
		events.LogMeleeAttackEvent(lam.parent, &ar, lam.parent.GetEventListener())
		if attackRollResult.IsSuccess {
			events.LogDiceRollEvent(lam.parent, dmgRollResult, lam.parent.GetEventListener())
		}
		results = append(results, ar)
	}
	return results, nil
}

// ExecuteAdvanced executes a lair action (attack or DC), with optional AOE.
// Returns total effective effects as AttackResults for logging symmetry.
func (lam *LairActionManager) ExecuteAdvanced(actionIndex int, primaryTarget core.Entity) ([]core.AttackResult, []core.Effect, error) {
	a, ok := lam.Actions[actionIndex]
	if !ok {
		return nil, nil, fmt.Errorf("invalid lair action index")
	}
	if a.Recharge > 0 && !lam.rechargeOK[actionIndex] {
		return nil, nil, fmt.Errorf("lair action on cooldown")
	}

	// Build target set
	targets := lam.collectTargets(a, primaryTarget)
	if len(targets) == 0 {
		return nil, nil, fmt.Errorf("no valid targets for lair action")
	}

	var results []core.AttackResult
	var effects []core.Effect

	switch a.Mode {
	case LAMAttack:
		// Single damage roll per target (attack per target)
		for i, t := range targets {
			// Build synthetic AttackRequest with one AttackData
			req := &core.AttackRequest{AttackData: []core.AttackData{a.AttackData}, AttackOptions: core.AttackOptions{Advantage: core.RollNormal, ShouldApplyDamageMod: true}, Target: t}
			rs, err := lam.ProcessAttackRequest(req)
			if err != nil {
				return nil, nil, err
			}
			for _, r := range rs {
				results = append(results, r)
				if r.GetIsHit() {
					effects = append(effects, core.Effect{Type: core.EffectDamage, Value: r.GetDamageResult().GetTotal(), DamageType: r.GetDamageType()})
				}
			}
			_ = i
		}
	case LAMDC:
		// Roll one damage value for the action, apply per target with save
		dOpts := roll_manager.NewRollOptions()
		dOpts.RollType = core.DiceRollDamage
		// Reuse AttackRequest shell to drive RollDamage API
		req := &core.AttackRequest{AttackData: []core.AttackData{a.AttackData}, Target: targets[0]}
		dmgRoll, err := lam.rollManager.RollDamage(req, 0, false, dOpts)
		if err != nil {
			return nil, nil, err
		}
		events.LogDiceRollEvent(lam.parent, dmgRoll, lam.parent.GetEventListener())

		for _, t := range targets {
			saveRes, err := t.MakeSavingThrow(a.DCAbility, a.DCValue, false)
			if err != nil {
				return nil, nil, err
			}
			// Log DC event (basic)
			// Using generic spell DC event structure with lair name
			events.LogSpellDCEvent(lam.parent, t, &spells.Spell{Name: a.Name, Level: 0}, a.DCValue, saveRes.GetTotal(), saveRes.GetIsSuccess(), lam.parent.GetEventListener())

			applied := dmgRoll.Total
			if saveRes.GetIsSuccess() {
				if a.OnSuccess == core.DCOnSuccessHalf {
					applied = applied / 2
				} else {
					applied = 0
				}
			}
			ar := core.AttackResult{ // reuse for summary
				ActorName:  lam.parent.GetName(),
				TargetName: t.GetName(),
				AttackName: a.Name,
				DamageRoll: dmgRoll,
				DamageType: a.AttackData.DamageType,
				IsHit:      applied > 0,
			}
			results = append(results, ar)
			if applied > 0 {
				effects = append(effects, core.Effect{Type: core.EffectDamage, Value: applied, DamageType: a.AttackData.DamageType})
			}
		}
	}

	lam.expend(actionIndex)
	return results, effects, nil
}

func (lam *LairActionManager) collectTargets(a LairAction, primary core.Entity) []core.Entity {
	// If not AOE, return primary only
	if !a.IsAOE {
		if primary != nil {
			return []core.Entity{primary}
		}
		return nil
	}
	if lam.parent.combatCtx == nil {
		return nil
	}
	targets := make([]core.Entity, 0)
	for _, ci := range lam.parent.combatCtx.CombatantInfo {
		if ci == nil || ci.Combatant == nil || ci.Combatant.IsLair {
			continue
		}
		e := ci.Combatant.GetEntity()
		if a.TargetSide == TargetCharacters && e.IsCharacter() && !e.IsUnconscious() && !e.IsDead() {
			targets = append(targets, e)
		} else if a.TargetSide == TargetMonsters && e.IsMonster() && !e.IsUnconscious() && !e.IsDead() {
			targets = append(targets, e)
		}
	}
	return targets
}
