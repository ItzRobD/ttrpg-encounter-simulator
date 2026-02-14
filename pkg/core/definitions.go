package core

// #DEFINE

const SIM_BIG_NUMBER = 1000000

// #END DEFINE

type Side string

const (
	SideCharacters Side = "character"
	SideMonsters   Side = "monster"
)

type VictoryStatus string

const (
	VictoryStatusNone       VictoryStatus = "none"
	VictoryStatusCharacters VictoryStatus = "characters"
	VictoryStatusMonsters   VictoryStatus = "monsters"
	VictoryStatusDraw       VictoryStatus = "draw"
)

type ActorType string

const (
	ActorTypeCharacter ActorType = "character"
	ActorTypeMonster   ActorType = "monster"
	ActorTypeLair      ActorType = "lair"
	ActorTypeNone      ActorType = ""
)

type Seed struct {
	Seed1 uint64 `json:"seed1"`
	Seed2 uint64 `json:"seed2"`
}

type HPVisibilityMode string

const (
	HPVisible      HPVisibilityMode = "visible"
	HPPercentage   HPVisibilityMode = "percentage"
	HPStatusHidden HPVisibilityMode = "hidden"
)

type MultiattackFollowUpPolicy string

const (
	MultiattackPolicyNone           MultiattackFollowUpPolicy = "none"
	MultiattackPolicyRetargetOnDown MultiattackFollowUpPolicy = "retarget_on_down"
	MultiattackPolicyWasteRemaining MultiattackFollowUpPolicy = "waste_remaining"
)

type ActionSelectionPolicy string

const (
	ActionPolicyHighestDamage ActionSelectionPolicy = "highest_damage"
	ActionPolicyRandom        ActionSelectionPolicy = "random"
	ActionPolicyPriority      ActionSelectionPolicy = "priority"
)

type SimulationOptions struct {
	Seed                           Seed                      `json:"seed"`
	UseHPAverageMonster            bool                      `json:"use_hp_average_monster"`
	UseHPAverageCharacter          bool                      `json:"use_hp_average_character"`
	CanMonstersCrit                bool                      `json:"can_monsters_crit"`
	CanCharactersCrit              bool                      `json:"can_characters_crit"`
	HasIncreasedCrits              bool                      `json:"has_increased_crits"`
	UseImprovedCritical            bool                      `json:"use_improved_critical"`
	CharactersAlwaysUpcast         bool                      `json:"characters_always_upcast"`
	MonstersAlwaysUpcast           bool                      `json:"monsters_always_upcast"`
	AllowCharacterHeals            bool                      `json:"allow_character_heals"`
	AllowMonsterHeals              bool                      `json:"allow_monster_heals"`
	AOEHitsAllEnemies              bool                      `json:"aoe_hits_all_enemies"`
	CharacterHealThresholdPct      int                       `json:"character_heal_threshold_pct"`
	MonsterHealThresholdPct        int                       `json:"monster_heal_threshold_pct"`
	CharacterEmergencyThresholdPct int                       `json:"character_emergency_threshold_pct"`
	MonsterEmergencyThresholdPct   int                       `json:"monster_emergency_threshold_pct"`
	LimitedLegendaryActions        bool                      `json:"limited_legendary_actions"`
	AllowLairActions               bool                      `json:"allow_lair_actions"`
	AllowDragonbornBreathAttack    bool                      `json:"allow_dragonborn_breath_attack"`
	EnableClassFeatures            bool                      `json:"enable_class_features"`
	EnableRacialFeatures           bool                      `json:"enable_racial_features"`
	BarbarianAlwaysRecklessAttack  bool                      `json:"barbarian_always_reckless_attack"`
	PaladinAlwaysSmite             bool                      `json:"paladin_always_smite"`
	PaladinUseHighestSmiteSlot     bool                      `json:"paladin_use_highest_smite_slot"`
	UseMassiveDamage               bool                      `json:"use_massive_damage"`
	EnableSpecialAbilities         bool                      `json:"enable_special_abilities"`
	MonsterDeathEffectsHitAllies   bool                      `json:"monster_death_effects_hit_allies"`
	AlwaysUseSneakAttack           bool                      `json:"always_use_sneak_attack"`
	UseWeightedAI                  bool                      `json:"use_weighted_ai"`
	DebugAI                        bool                      `json:"debug_ai"`
	HPVisibilityMode               HPVisibilityMode          `json:"hp_visibility_mode"`
	EnableMonsterNoise             bool                      `json:"enable_monster_noise"`
	MonsterNoiseWeight             float64                   `json:"monster_noise_weight"`
	AOETargetThreshold             int                       `json:"aoe_target_threshold,omitempty"`
	MultiattackPolicy              MultiattackFollowUpPolicy `json:"multiattack_policy,omitempty"`
	ActionSelectionPolicy          ActionSelectionPolicy     `json:"action_selection_policy,omitempty"`
	DisableMonsterTurns            bool                      `json:"disable_monster_turns,omitempty"`
	DisableCharacterTurns          bool                      `json:"disable_character_turns,omitempty"`
	DisableLairTurns               bool                      `json:"disable_lair_turns,omitempty"`
}

type HPModificationResult struct {
	ModificationValue int  `json:"modification_value"`
	OriginalHP        int  `json:"original_hp"`
	OriginalTempHP    int  `json:"original_temp_hp"`
	NewHP             int  `json:"new_hp"`
	NewTempHP         int  `json:"new_temp_hp"`
	TempHPUsed        int  `json:"temp_hp_used"`
	DidHealHP         bool `json:"did_heal_hp"`
	DidHealTempHP     bool `json:"did_heal_temp_hp"`
	DidTempDamage     bool `json:"did_temp_damage"`
	DidHPDamage       bool `json:"did_hp_damage"`
	IsUnconscious     bool `json:"is_unconscious"`
	IsMaxHealth       bool `json:"is_max_health"`
}

type SpecialAbility string

const (
	SpecAbilityAssassinate          SpecialAbility = "Assassinate"
	SpecAbilityBerserk              SpecialAbility = "Berserk"
	SpecAbilityBloodFrenzy          SpecialAbility = "Blood Frenzy"
	SpecAbilityConsumeLife          SpecialAbility = "Consume Life"
	SpecAbilityCorrosiveForm        SpecialAbility = "Corrosive Form"
	SpecAbilityDeathBurst           SpecialAbility = "Death Burst"
	SpecAbilityDeathThroes          SpecialAbility = "Death Throes"
	SpecAbilityDivineEminence       SpecialAbility = "Divine Eminence"
	SpecAbilityDivineSmite          SpecialAbility = "Divine Smite"
	SpecAbilityEvasion              SpecialAbility = "Evasion"
	SpecAbilityFireAura             SpecialAbility = "Aura"
	SpecAbilityFireForm             SpecialAbility = "Fire Form"
	SpecAbilityGibbering            SpecialAbility = "Gibbering"
	SpecAbilityCunning              SpecialAbility = "Cunning"
	SpecAbilityHeatedBody           SpecialAbility = "Heated Body"
	SpecAbilityLegendaryResistance  SpecialAbility = "Legendary Resistance"
	SpecAbilityAbsorption           SpecialAbility = "Absorption"
	SpecAbilityLimitedMagicImmunity SpecialAbility = "Limited Magic Immunity"
	SpecAbilityMagicResistance      SpecialAbility = "Magic Resistance"
	SpecAbilityMagicWeapons         SpecialAbility = "Magic Weapons"
	SpecAbilityMartialAdvantage     SpecialAbility = "Martial Advantage"
	SpecAbilityPackTactics          SpecialAbility = "Pack Tactics"
	SpecAbilityReckless             SpecialAbility = "Reckless"
	SpecAbilityReflectiveCarapace   SpecialAbility = "Reflective Carapace"
	SpecAbilityRegeneration         SpecialAbility = "Regeneration"
	SpecAbilityRelentless           SpecialAbility = "Relentless"
	SpecAbilityRelentlessEndurance  SpecialAbility = "Relentless Endurance"
	SpecAbilitySneakAttack          SpecialAbility = "Sneak Attack (1/Turn)"
	SpecAbilityUndeadFortitude      SpecialAbility = "Undead Fortitude"
	SpecAbilitySecondWind           SpecialAbility = "Second Wind"

	// New constants for existing features
	SpecAbilityDwarvenResilience    SpecialAbility = "Dwarven Resilience"
	SpecAbilityHellishResistance    SpecialAbility = "Hellish Resistance"
	SpecAbilityDraconicResistance   SpecialAbility = "Draconic Resistance"
	SpecAbilityBreathWeapon         SpecialAbility = "Breath Weapon"
	SpecAbilityIndomitable          SpecialAbility = "Indomitable"
	SpecAbilityLayOnHands           SpecialAbility = "Lay on Hands"
	SpecAbilityFightingStyleDef     SpecialAbility = "Fighting Style: Defense"
	SpecAbilityFeralInstinct        SpecialAbility = "Feral Instinct"
	SpecAbilityDangerSense          SpecialAbility = "Danger Sense"
	SpecAbilitySlipperyMind         SpecialAbility = "Slippery Mind"
	SpecAbilityRageStrengthSave     SpecialAbility = "Rage: Strength Save Advantage"
	SpecAbilityImprovedDivineSmite  SpecialAbility = "Improved Divine Smite"
	SpecAbilityDeflectMissiles      SpecialAbility = "Deflect Missiles"
	SpecAbilityFightingStyleArchery SpecialAbility = "Fighting Style: Archery"
	SpecAbilityFightingStyleDuel    SpecialAbility = "Fighting Style: Dueling"
	SpecAbilityFightingStyleGWF     SpecialAbility = "Fighting Style: Great Weapon Fighting"
	SpecAbilityFightingStyleTWF     SpecialAbility = "Fighting Style: Two-Weapon Fighting"
	SpecAbilityHalflingLucky        SpecialAbility = "Halfling Lucky"
	SpecAbilitySavageAttacks        SpecialAbility = "Savage Attacks"
	SpecAbilityBrutalCritical       SpecialAbility = "Brutal Critical"
	SpecAbilityRelentlessRage       SpecialAbility = "Relentless Rage"
	SpecAbilityRageExtraDamage      SpecialAbility = "Rage: Extra Damage"
	SpecAbilityRageResistance       SpecialAbility = "Rage: Resistance"
	SpecAbilityUncannyDodge         SpecialAbility = "Uncanny Dodge"
	SpecAbilityElusive              SpecialAbility = "Elusive"
)
