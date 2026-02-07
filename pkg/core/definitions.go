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
	UseHPAverageMonster            bool                      `json:"useHPAverageMonster"`
	UseHPAverageCharacter          bool                      `json:"useHPAverageCharacter"`
	CanMonstersCrit                bool                      `json:"canMonstersCrit"`
	CanCharactersCrit              bool                      `json:"canCharactersCrit"`
	HasIncreasedCrits              bool                      `json:"hasIncreasedCrits"`
	UseImprovedCritical            bool                      `json:"useImprovedCritical"`
	CharactersAlwaysUpcast         bool                      `json:"charactersAlwaysUpcast"`
	MonstersAlwaysUpcast           bool                      `json:"monstersAlwaysUpcast"`
	AllowCharacterHeals            bool                      `json:"allowCharacterHeals"`
	AllowMonsterHeals              bool                      `json:"allowMonsterHeals"`
	AOEHitsAllEnemies              bool                      `json:"aoeHitsAllEnemies"`
	CharacterHealThresholdPct      int                       `json:"characterHealThresholdPct"`
	MonsterHealThresholdPct        int                       `json:"monsterHealThresholdPct"`
	CharacterEmergencyThresholdPct int                       `json:"characterEmergencyThresholdPct"`
	MonsterEmergencyThresholdPct   int                       `json:"monsterEmergencyThresholdPct"`
	LimitedLegendaryActions        bool                      `json:"limitedLegendaryActions"`
	AllowLairActions               bool                      `json:"allowLairActions"`
	AllowDragonbornBreathAttack    bool                      `json:"allowDragonbornBreathAttack"`
	EnableClassFeatures            bool                      `json:"enableClassFeatures"`
	EnableRacialFeatures           bool                      `json:"enableRacialFeatures"`
	BarbarianAlwaysRecklessAttack  bool                      `json:"barbarianAlwaysRecklessAttack"`
	PaladinAlwaysSmite             bool                      `json:"paladinAlwaysSmite"`
	PaladinUseHighestSmiteSlot     bool                      `json:"paladinUseHighestSmiteSlot"`
	UseMassiveDamage               bool                      `json:"useMassiveDamage"`
	EnableSpecialAbilities         bool                      `json:"enableSpecialAbilities"`
	MonsterDeathEffectsHitAllies   bool                      `json:"monsterDeathEffectsHitAllies"`
	AlwaysUseSneakAttack           bool                      `json:"alwaysUseSneakAttack"`
	UseWeightedAI                  bool                      `json:"useWeightedAI"`
	DebugAI                        bool                      `json:"debugAI"`
	HPVisibilityMode               HPVisibilityMode          `json:"hpVisibilityMode"`
	EnableMonsterNoise             bool                      `json:"enableMonsterNoise"`
	MonsterNoiseWeight             float64                   `json:"monsterNoiseWeight"`
	AOETargetThreshold             int                       `json:"aoeTargetThreshold,omitempty"`
	MultiattackPolicy              MultiattackFollowUpPolicy `json:"multiattackPolicy,omitempty"`
	ActionSelectionPolicy          ActionSelectionPolicy     `json:"actionSelectionPolicy,omitempty"`
	DebugEnableMonsterTurns        bool                      `json:"debugEnableMonsterTurns,omitempty"`
	DebugEnableCharacterTurns      bool                      `json:"debugEnableCharacterTurns,omitempty"`
	DebugEnableLairTurns           bool                      `json:"debugEnableLairTurns,omitempty"`
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
