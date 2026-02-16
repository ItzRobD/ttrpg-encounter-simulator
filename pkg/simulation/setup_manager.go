package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

type SetupManager struct {
	ctx         context.Context
	rollManager *roll_manager.RollManager
}

func NewSetupManager(ctx context.Context, rm *roll_manager.RollManager) *SetupManager {
	return &SetupManager{
		ctx:         ctx,
		rollManager: rm,
	}
}

func (sm *SetupManager) SetupActor(cfg actor.ActorConfig) (*actor.Actor, error) {
	var a *actor.Actor
	var err error

	switch cfg.ActorType {
	case core.ActorTypeCharacter:
		a, err = sm.setupCharacter(cfg)
	case core.ActorTypeMonster:
		a, err = sm.setupMonster(cfg)
	case core.ActorTypeLair:
		a, err = sm.setupLair(cfg)
	default:
		return nil, fmt.Errorf("invalid actor type: %v", cfg.ActorType)
	}

	if err != nil {
		return nil, err
	}

	// Apply initial state if provided (overwrites hydrated defaults)
	if cfg.InitialState != nil {
		a.StateManager.CurrentHP = cfg.InitialState.CurrentHP
		a.StateManager.MaxHP = cfg.InitialState.MaxHP
		a.StateManager.TempHP = cfg.InitialState.TempHP
		a.StateManager.Conditions = cfg.InitialState.Conditions
		a.StateManager.HealthState = cfg.InitialState.HealthState
	}

	return a, nil
}

func (sm *SetupManager) setupCharacter(cfg actor.ActorConfig) (*actor.Actor, error) {
	// Hydrate basic actor properties from configuration
	a, err := actor.NewActorFromConfig(&cfg)
	if err != nil {
		return nil, err
	}

	// Extra attack
	attackCount, err := repo.GetAttackCountByClassAndLevel(sm.ctx, int(a.Metadata.ClassID), a.Metadata.Level)
	if err == nil {
		a.StateManager.AttackCount = attackCount
	}

	// Hydrate Spellcasting
	spellSlots, err := repo.GetSpellSlotsByClassAndLevel(sm.ctx, int(a.Metadata.ClassID), a.Metadata.Level)
	if err == nil && len(spellSlots) > 0 {
		a.StateManager.MaxSlots = spellSlots
		a.StateManager.CurrentSlots = make(spells.SpellSlots)
		for k, v := range spellSlots {
			a.StateManager.CurrentSlots[k] = v
		}

		spellAbility, err := repo.GetSpellcastingAbilityByClassID(sm.ctx, int(a.Metadata.ClassID))
		if err == nil {
			a.SpellManager.SpellcastingAbility = spellAbility
			a.Metadata.SpellcasterMetadata.IsSpellcaster = true
			a.Metadata.SpellcasterMetadata.SpellcastingAbility = spellAbility
			a.Metadata.SpellcasterMetadata.SpellcastingLevel = a.Metadata.Level
		}
	} else if a.Metadata.SpellcasterMetadata.IsSpellcaster && len(a.StateManager.MaxSlots) > 0 {
		// Fallback for custom characters if slots are already provided in config
		if len(a.StateManager.CurrentSlots) == 0 {
			a.StateManager.CurrentSlots = make(spells.SpellSlots)
			for k, v := range a.StateManager.MaxSlots {
				a.StateManager.CurrentSlots[k] = v
			}
		}
	}

	// Hydrate Class Features
	classFeatures, err := repo.HydrateClassFeaturesSRD(sm.ctx, int(a.Metadata.ClassID), a.Metadata.Level)
	if err == nil {
		a.Features = append(a.Features, classFeatures...)
	}

	// Hydrate Race Features
	raceFeatures, err := repo.HydrateRaceFeaturesSRD(sm.ctx, int(a.Metadata.RaceID), a.Metadata.Level, string(a.Metadata.DragonbornColor))
	if err == nil {
		a.Features = append(a.Features, raceFeatures...)
	}

	// TODO: Handle custom equipment
	// Hydrate Equipment
	eq, err := repo.HydrateEquipment(sm.ctx, &cfg)
	if err == nil {
		for _, item := range eq {
			// Select slot based on equipment type
			slot := equipment_manager.EquipmentSlotPrimary
			if item.Type == equipment.EquipmentTypeArmor {
				slot = equipment_manager.EquipmentSlotArmor
			} else if item.Type == equipment.EquipmentTypeShield {
				slot = equipment_manager.EquipmentSlotShield
			}
			err := a.Equipment.AddItem(slot, item)
			if err != nil {
				return nil, err
			}
		}
	}

	// TODO: Handle custom spells
	// Hydrate Known Spells
	if len(cfg.KnownSpellIDs) > 0 {
		knownSpells, err := repo.QuerySpellData(sm.ctx, spells.SpellQueryParams{ID: cfg.KnownSpellIDs})
		if err == nil {
			for _, s := range knownSpells {
				// Create a copy of the spell to get a pointer to it
				spellCopy := s
				err := a.SpellManager.AddKnownSpell(&spellCopy)
				if err != nil {
					fmt.Printf("failed to add known spell %s: %v\n", s.ID, err)
				}
			}
		} else {
			fmt.Printf("failed to hydrate known spells: %v\n", err)
		}
	}

	// Populate Actions
	// Custom spells from config are already added to SpellManager via NewActorFromConfig
	a.UpdateActionsFromEquipment()
	a.UpdateActionsFromSpells()

	// Initialize Recharge Actions as charged
	for _, act := range a.Actions {
		if act.RechargeValue > 0 {
			a.StateManager.Resource[act.Name] = 1
		}
	}

	// Process Features (apply complex logic, AC, etc.)
	a.ProcessFeatures()

	// Calculate Action Averages
	for i := range a.Actions {
		sm.calculateActionAverage(&a.Actions[i])
	}

	// Initialize HP based on HPConfig
	sm.initializeActorHP(a)

	// Initialize Hit Dice for adventuring day
	if len(a.HPConfig.HitDice) > 0 {
		a.StateManager.MaxHitDice = make(map[core.DiceType]int)
		a.StateManager.CurrentHitDice = make(map[core.DiceType]int)
		for die, count := range a.HPConfig.HitDice {
			a.StateManager.MaxHitDice[die] = count
			a.StateManager.CurrentHitDice[die] = count
		}
	}

	a.UpdateOffensiveValues()
	a.Behavior = cfg.Behavior

	return a, nil
}

func (sm *SetupManager) calculateActionAverage(act *core.Action) {
	totalAvg := 0
	for _, db := range act.DiceBlock {
		avg, _ := core.GetAverageRoll(db.NumberOfDice, db.Die, db.Modifier)
		totalAvg += avg
	}
	act.AverageDamage = totalAvg
}

func (sm *SetupManager) setupMonster(cfg actor.ActorConfig) (*actor.Actor, error) {
	var config *actor.ActorConfig
	var err error

	if cfg.IsCustom {
		config = &cfg
	} else {
		config, err = repo.HydrateMonsterConfig(sm.ctx, cfg.ID)
		if err != nil || config == nil {
			return nil, err
		}

		// Merge features from input config if any
		if len(cfg.Features) > 0 {
			config.Features = append(config.Features, cfg.Features...)
		}

		// Preserve side and behavior from input if provided
		if cfg.Side != "" {
			config.Side = cfg.Side
		}
		if cfg.Behavior.ActionPreference != "" {
			config.Behavior.ActionPreference = cfg.Behavior.ActionPreference
		}
		if cfg.Behavior.SecondaryActionPreference != "" {
			config.Behavior.SecondaryActionPreference = cfg.Behavior.SecondaryActionPreference
		}
		if cfg.Behavior.TargetPriority != "" {
			config.Behavior.TargetPriority = cfg.Behavior.TargetPriority
		}
		if cfg.Behavior.SecondaryTargetPriority != "" {
			config.Behavior.SecondaryTargetPriority = cfg.Behavior.SecondaryTargetPriority
		}
	}

	a, err := actor.NewActorFromConfig(config)
	if err != nil {
		return nil, err
	}

	// Ensure the side is set from the initial request if not specified in the monster data
	if a.Side == "" {
		a.Side = cfg.Side
	}
	// Failsafe for monsters
	if a.Side == "" {
		a.Side = core.SideMonsters
	}

	// Fill actions
	a.UpdateActionsFromEquipment()
	a.UpdateActionsFromSpells()

	// Initialize Recharge Actions as charged
	for _, act := range a.Actions {
		if act.RechargeValue > 0 {
			a.StateManager.Resource[act.Name] = 1
		}
	}

	// Calculate Action Averages
	for i := range a.Actions {
		sm.calculateActionAverage(&a.Actions[i])
	}

	// Initialize HP based on HPConfig
	sm.initializeActorHP(a)

	// Initialize Hit Dice for adventuring day
	if len(a.HPConfig.HitDice) > 0 {
		a.StateManager.MaxHitDice = make(map[core.DiceType]int)
		a.StateManager.CurrentHitDice = make(map[core.DiceType]int)
		for die, count := range a.HPConfig.HitDice {
			a.StateManager.MaxHitDice[die] = count
			a.StateManager.CurrentHitDice[die] = count
		}
	}

	a.UpdateOffensiveValues()
	a.Behavior = config.Behavior

	return a, nil
}

func (sm *SetupManager) initializeActorHP(a *actor.Actor) {
	hp := 0
	switch a.HPConfig.HPMethod {
	case core.HPSetValue:
		if a.HPConfig.Value > 0 {
			hp = a.HPConfig.Value
		}
	case core.HPSetRoll:
		if sm.rollManager != nil && len(a.HPConfig.HitDice) > 0 {
			totalHP := a.HPConfig.AmountToAdd
			for die, count := range a.HPConfig.HitDice {
				res := sm.rollManager.RollDice(count, die, roll_manager.RollOptions{})
				totalHP += res.Total
			}
			hp = totalHP
		}
	case core.HPSetAverage:
		hp = a.HPConfig.HPAverage
	default:
		hp = a.HPConfig.HPAverage
	}

	if hp > 0 {
		a.StateManager.MaxHP = hp
		a.StateManager.CurrentHP = hp
	}
}

func (sm *SetupManager) setupLair(cfg actor.ActorConfig) (*actor.Actor, error) {
	// Lair hydration
	a, err := actor.NewActorFromConfig(&cfg)
	if err != nil {
		return nil, err
	}
	return a, nil
}
