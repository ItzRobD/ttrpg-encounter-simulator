package core

// MultiattackFollowUpPolicy controls how monsters handle remaining swings in a multiattack
// after a target drops to 0 HP.
type MultiattackFollowUpPolicy string

const (
	// MultiattackPolicyNone preserves legacy behavior (keep swinging same target)
	MultiattackPolicyNone MultiattackFollowUpPolicy = "none"
	// MultiattackPolicyRetargetOnDown retargets the next swing to a new valid target when the current one drops
	MultiattackPolicyRetargetOnDown MultiattackFollowUpPolicy = "retarget_on_down"
	// MultiattackPolicyWasteRemaining stops executing remaining swings when the current target drops
	MultiattackPolicyWasteRemaining MultiattackFollowUpPolicy = "waste_remaining"
)

type SimulationOptions struct {
	Seed                          Seed `json:"seed"`
	UseHPAverageMonster           bool `json:"useHPAverageMonster"`
	UseHPAverageCharacter         bool `json:"useHPAverageCharacter"`
	CanMonstersCrit               bool `json:"canMonstersCrit"`
	CanCharactersCrit             bool `json:"canCharactersCrit"`
	HasIncreasedCrits             bool `json:"hasIncreasedCrits"`
	UseImprovedCritical           bool `json:"useImprovedCritical"`
	CharactersAlwaysUpcast        bool `json:"charactersAlwaysUpcast"`
	MonstersAlwaysUpcast          bool `json:"monstersAlwaysUpcast"`
	AllowCharacterHeals           bool `json:"allowCharacterHeals"`
	AllowMonsterHeals             bool `json:"allowMonsterHeals"`
	AOEHitsAllEnemies             bool `json:"aoeHitsAllEnemies"`
	CharacterHealThresholdPct     int  `json:"characterHealThresholdPct"`
	MonsterHealThresholdPct       int  `json:"monsterHealThresholdPct"`
	LimitedLegendaryActions       bool `json:"limitedLegendaryActions"`
	AllowLairActions              bool `json:"allowLairActions"`
	AllowDragonbornBreathAttack   bool `json:"allowDragonbornBreathAttack"`
	EnableClassFeatures           bool `json:"enableClassFeatures"`
	EnableRacialFeatures          bool `json:"enableRacialFeatures"`
	BarbarianAlwaysRecklessAttack bool `json:"barbarianAlwaysRecklessAttack"`
	PaladinAlwaysSmite            bool `json:"paladinAlwaysSmite"`
	PaladinUseHighestSmiteSlot    bool `json:"paladinUseHighestSmiteSlot"`
	UseMassiveDamage              bool `json:"useMassiveDamage"`
	EnableSpecialAbilities        bool `json:"enableSpecialAbilities"`
	MonsterDeathEffectsHitAllies  bool `json:"monsterDeathEffectsHitAllies"`
	AlwaysUseSneakAttack          bool `json:"alwaysUseSneakAttack"`

	// Premium AI Options
	UseWeightedAI      bool             `json:"useWeightedAI"`
	DebugAI            bool             `json:"debugAI"`
	HPVisibilityMode   HPVisibilityMode `json:"hpVisibilityMode"`
	EnableMonsterNoise bool             `json:"enableMonsterNoise"`
	MonsterNoiseWeight float64          `json:"monsterNoiseWeight"`

	// Multiattack follow-up policy (single enum)
	MultiattackPolicy MultiattackFollowUpPolicy `json:"multiattackPolicy,omitempty"`
}

// GetMultiattackPolicy returns a safe, non-empty policy value.
func (so *SimulationOptions) GetMultiattackPolicy() MultiattackFollowUpPolicy {
	if so == nil {
		return MultiattackPolicyNone
	}
	if so.MultiattackPolicy == "" {
		return MultiattackPolicyNone
	}
	return so.MultiattackPolicy
}
