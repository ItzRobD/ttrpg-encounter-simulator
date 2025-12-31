package core

import (
	"math/rand/v2"
	"testing"
)

// Test your actual GetAbilityScoreModifier function
func TestGetAbilityScoreModifier(t *testing.T) {
	tests := []struct {
		name    string
		score   int
		wantMod int
		wantErr bool
	}{
		// Valid scores from your main.go (Frank and Jack have 18 STR)
		{"strength 18", 18, 4, false},
		{"dexterity 14", 14, 2, false},
		{"constitution 16", 16, 3, false},
		{"intelligence 10", 10, 0, false},
		{"wisdom 12", 12, 1, false},

		// Edge cases
		{"minimum valid score", 1, -5, false},
		{"maximum valid score", 30, 10, false},
		{"average score", 10, 0, false},

		// Invalid scores
		{"score too low", 0, 0, true},
		{"score too high", 31, 0, true},
		{"negative score", -5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMod, gotErr := GetAbilityScoreModifier(tt.score)

			// Check if error expectation matches
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetAbilityScoreModifier(%d) error = %v, wantErr %v",
					tt.score, gotErr, tt.wantErr)
				return
			}

			// If we expected an error, we're done
			if tt.wantErr {
				return
			}

			// Check if modifier matches
			if gotMod != tt.wantMod {
				t.Errorf("GetAbilityScoreModifier(%d) = %d, want %d",
					tt.score, gotMod, tt.wantMod)
			}
		})
	}
}

// Test proficiency bonus calculation
func TestGetCharacterProficiencyBonus(t *testing.T) {
	tests := []struct {
		name    string
		level   uint8
		want    int
		wantErr bool
	}{
		{"level 1", 1, 2, false},
		{"level 4 (Frank and Jack)", 4, 2, false},
		{"level 5 (tier 2)", 5, 3, false},
		{"level 9 (tier 3)", 9, 4, false},
		{"level 20 (max)", 20, 6, false},
		{"level 0 (invalid)", 0, 0, true},
		{"level 21 (invalid)", 21, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCharacterProficiencyBonus(tt.level)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCharacterProficiencyBonus(%d) error = %v, wantErr %v",
					tt.level, err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetCharacterProficiencyBonus(%d) = %d, want %d",
					tt.level, got, tt.want)
			}
		})
	}
}

func TestGetMonsterProficiencyBonus(t *testing.T) {
	tests := []struct {
		name    string
		cr      float64
		want    int
		wantErr bool
	}{
		{"cr 1/8", 1.0 / 8, 2, false},
		{"cr 2", 2, 2, false},
		{"cr 15", 15, 5, false},
		{"cr 30", 30, 9, false},
		{"cr 40", 40, 0, true},
		{"negative cr", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMonsterProficiencyBonus(tt.cr)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMonsterProficiencyBonus(%f) error = %v, wantErr %v",
					tt.cr, got, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetMonsterProficiencyBonus(%f) = %d, want %d",
					tt.cr, got, tt.want)
			}
		})
	}
}

func TestDetermineAttackAdvantageFromConditions(t *testing.T) {
	tests := []struct {
		name             string
		actorConditions  EntityConditions
		targetConditions EntityConditions
		want             AdvantageType
	}{
		{
			"no conditions",
			nil,
			nil,
			RollNormal,
		},
		{
			"actor blinded (outgoing disadvantage)",
			EntityConditions{ConditionBlinded: true},
			nil,
			RollDisadvantage,
		},
		{
			"target prone (generic incoming is normal in our system, handled in unified helper)",
			nil,
			EntityConditions{ConditionProne: true},
			RollNormal,
		},
		{
			"both blinded and prone (blinded gives disadvantage, prone generic is normal -> disadvantage)",
			EntityConditions{ConditionBlinded: true},
			EntityConditions{ConditionProne: true},
			RollDisadvantage,
		},
		{
			"invisible actor (advantage)",
			EntityConditions{ConditionInvisible: true},
			nil,
			RollAdvantage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineAttackAdvantageFromConditions(tt.actorConditions, tt.targetConditions)
			if got != tt.want {
				t.Errorf("DetermineAttackAdvantageFromConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineSaveAdvantageFromConditions(t *testing.T) {
	tests := []struct {
		name            string
		actorConditions EntityConditions
		ability         Ability
		want            AdvantageType
	}{
		{
			"no conditions",
			nil,
			AbilityDexterity,
			RollNormal,
		},
		{
			"restrained (disadvantage on dex saves)",
			EntityConditions{ConditionRestrained: true},
			AbilityDexterity,
			RollDisadvantage,
		},
		{
			"restrained (normal on str saves)",
			EntityConditions{ConditionRestrained: true},
			AbilityStrength,
			RollNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineSaveAdvantageFromConditions(tt.actorConditions, tt.ability)
			if got != tt.want {
				t.Errorf("DetermineSaveAdvantageFromConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFinalAdvantageType(t *testing.T) {
	tests := []struct {
		name string
		advs []AdvantageType
		want AdvantageType
	}{
		{"empty", []AdvantageType{}, RollNormal},
		{"single advantage", []AdvantageType{RollAdvantage}, RollAdvantage},
		{"single disadvantage", []AdvantageType{RollDisadvantage}, RollDisadvantage},
		{"mixed cancels", []AdvantageType{RollAdvantage, RollDisadvantage}, RollNormal},
		{"multiple advantage", []AdvantageType{RollAdvantage, RollAdvantage}, RollAdvantage},
		{"more advantage than disadvantage (5e cancels to normal)", []AdvantageType{RollAdvantage, RollAdvantage, RollDisadvantage}, RollNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFinalAdvantageType(tt.advs)
			if got != tt.want {
				t.Errorf("GetFinalAdvantageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineAttackAdvantage(t *testing.T) {
	tests := []struct {
		name             string
		actorConditions  EntityConditions
		targetConditions EntityConditions
		isRanged         bool
		base             AdvantageType
		want             AdvantageType
	}{
		{
			"normal attack",
			nil,
			nil,
			false,
			RollNormal,
			RollNormal,
		},
		{
			"reckless attack (base advantage)",
			nil,
			nil,
			false,
			RollAdvantage,
			RollAdvantage,
		},
		{
			"prone target melee (advantage)",
			nil,
			EntityConditions{ConditionProne: true},
			false,
			RollNormal,
			RollAdvantage,
		},
		{
			"prone target ranged (disadvantage)",
			nil,
			EntityConditions{ConditionProne: true},
			true,
			RollNormal,
			RollDisadvantage,
		},
		{
			"blinded actor (disadvantage)",
			EntityConditions{ConditionBlinded: true},
			nil,
			false,
			RollNormal,
			RollDisadvantage,
		},
		{
			"blinded actor hitting prone target (cancels)",
			EntityConditions{ConditionBlinded: true},
			EntityConditions{ConditionProne: true},
			false,
			RollNormal,
			RollNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineAttackAdvantage(tt.actorConditions, tt.targetConditions, tt.isRanged, tt.base)
			if got != tt.want {
				t.Errorf("DetermineAttackAdvantage() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockEntity struct {
	entityStub
	conditions EntityConditions
	elusive    bool
}

func (m *mockEntity) GetConditions() EntityConditions {
	return m.conditions
}

func (m *mockEntity) HasElusive() bool {
	return m.elusive
}

func (m *mockEntity) GetType() string {
	return "Humanoid"
}

func (m *mockEntity) BreakConcentration()           {}
func (m *mockEntity) IsConcentrating() bool         { return false }
func (m *mockEntity) SetConcentrating(bool, string) {}

// Minimal implementation of Entity interface via embedding entityStub
type entityStub struct{}

func (entityStub) GetClassID() uint8                                 { return 0 }
func (entityStub) IsDead() bool                                      { return false }
func (entityStub) IsUnconscious() bool                               { return false }
func (entityStub) GetHPStatus() HPStatus                             { return nil }
func (entityStub) GetName() string                                   { return "" }
func (entityStub) GetAC() int                                        { return 0 }
func (entityStub) GetEventListener() func(event interface{})         { return nil }
func (entityStub) SetEventListener(listener func(event interface{})) {}
func (entityStub) GetLevel() float64                                 { return 0 }
func (entityStub) GetHitDie() DiceType                               { return 0 }
func (entityStub) GetCasterLevel() int                               { return 0 }
func (entityStub) MakeSavingThrow(ability Ability, targetValue int, isSpell bool, damageType DamageType) (RollResult, error) {
	return nil, nil
}
func (entityStub) GetSpellSaveDC(ability *Ability) (int, error)         { return 0, nil }
func (entityStub) GetAbilityScores() AbilityScores                      { return AbilityScores{} }
func (entityStub) GetAbilityScore(ability Ability) int                  { return 0 }
func (entityStub) GetAbilityScoreModifier(ability Ability) (int, error) { return 0, nil }
func (entityStub) GetSavingThrowBonus(ability Ability) (int, error)     { return 0, nil }
func (entityStub) IsCharacter() bool                                    { return false }
func (entityStub) IsMonster() bool                                      { return false }
func (entityStub) GetIsLegendary() bool                                 { return false }
func (entityStub) GetHPConfig() HPConfig                                { return HPConfig{} }
func (entityStub) GetState() interface{}                                { return nil }
func (entityStub) RollInitiative() (int, error)                         { return 0, nil }
func (entityStub) InitializeHP() error                                  { return nil }
func (entityStub) IsSpellcaster() bool                                  { return false }
func (entityStub) IsHealer() bool                                       { return false }
func (entityStub) GetTargetPriority() TargetPriority                    { return 0 }
func (entityStub) SetTargetPriority(priority TargetPriority)            {}
func (entityStub) ChooseSpellByHealingEfficiency(targetValue int) (*SpellChoice, error) {
	return nil, nil
}
func (entityStub) ChooseDamageSpellByPriority(p SpellPriority) (*SpellChoice, error) {
	return nil, nil
}
func (entityStub) GetHealingSpellCount() int                                     { return 0 }
func (entityStub) GetDamageSpellCount() int                                      { return 0 }
func (entityStub) GetRNG() *rand.Rand                                            { return nil }
func (entityStub) GetAIRequest(actorID int, t AIRequestType) (*AIRequest, error) { return nil, nil }
func (entityStub) ExecuteAIRequest(req *AIRequest) (*ActionOutcome, error)       { return nil, nil }
func (entityStub) UpdateAICombatContext(ctx *CombatContext) error                { return nil }
func (entityStub) ModifyHP(value int, isTemp bool, tempStacking bool, allowMassiveDamage bool) (HPModificationResult, error) {
	return nil, nil
}
func (entityStub) RefreshLegendaryActions() {}
func (entityStub) CanTakeActions() bool     { return true }
func (entityStub) ProcessTurn(actorID int, turnType TurnType) (*TurnResult, *AIRequest, error) {
	return nil, nil, nil
}
func (entityStub) GetConditions() EntityConditions { return nil }
func (entityStub) GetType() string                 { return "Humanoid" }
func (entityStub) IsConcentrating() bool           { return false }
func (entityStub) BreakConcentration()             {}
func (entityStub) SetConcentrating(bool, string)   {}

func TestDetermineAttackAdvantageForEntities(t *testing.T) {
	tests := []struct {
		name           string
		attacker       *mockEntity
		target         *mockEntity
		isRangedAttack bool
		base           AdvantageType
		want           AdvantageType
	}{
		{
			"simple normal",
			&mockEntity{conditions: nil},
			&mockEntity{conditions: nil},
			false,
			RollNormal,
			RollNormal,
		},
		{
			"elusive target cancels advantage",
			&mockEntity{conditions: EntityConditions{ConditionInvisible: true}},
			&mockEntity{conditions: nil, elusive: true},
			false,
			RollNormal,
			RollNormal,
		},
		{
			"elusive target doesn't cancel when incapacitated",
			&mockEntity{conditions: EntityConditions{ConditionInvisible: true}},
			&mockEntity{conditions: EntityConditions{ConditionIncapacitated: true}, elusive: true},
			false,
			RollNormal,
			RollAdvantage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineAttackAdvantageForEntities(tt.attacker, tt.target, tt.isRangedAttack, tt.base)
			if got != tt.want {
				t.Errorf("DetermineAttackAdvantageForEntities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateDie(t *testing.T) {
	tests := []struct {
		die  int
		want bool
	}{
		{4, true},
		{6, true},
		{8, true},
		{10, true},
		{12, true},
		{20, true},
		{100, true},
		{0, false},
		{7, false},
		{50, false},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.die)), func(t *testing.T) {
			if got := ValidateDie(tt.die); got != tt.want {
				t.Errorf("ValidateDie(%d) = %v, want %v", tt.die, got, tt.want)
			}
		})
	}
}

func TestGetDieAverage(t *testing.T) {
	tests := []struct {
		die  DiceType
		want float64
	}{
		{D4, 2.5},
		{D6, 3.5},
		{D8, 4.5},
		{D10, 5.5},
		{D12, 6.5},
		{D20, 10.5},
	}

	for _, tt := range tests {
		t.Run(tt.die.String(), func(t *testing.T) {
			got, _ := GetDieAverage(tt.die)
			if got != tt.want {
				t.Errorf("GetDieAverage(%v) = %v, want %v", tt.die, got, tt.want)
			}
		})
	}
}

func TestGetAverageRoll(t *testing.T) {
	tests := []struct {
		name     string
		numDice  int
		die      DiceType
		amtToAdd int
		want     int
	}{
		{"1d8+0", 1, D8, 0, 4},  // int(4.5) = 4
		{"1d8+2", 1, D8, 2, 6},  // int(4.5 + 2) = 6
		{"2d8+0", 2, D8, 0, 9},  // int(9.0) = 9
		{"2d6+3", 2, D6, 3, 10}, // int(7.0 + 3) = 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := GetAverageRoll(tt.numDice, tt.die, tt.amtToAdd)
			if got != tt.want {
				t.Errorf("GetAverageRoll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateDamageType(t *testing.T) {
	tests := []struct {
		dt   string
		want bool
	}{
		{"acid", true},
		{"fire", true},
		{"slashing", true},
		{"invalid", false},
		{"FIRE", false}, // Currently case-sensitive in implementation
	}

	for _, tt := range tests {
		t.Run(tt.dt, func(t *testing.T) {
			if got := ValidateDamageType(tt.dt); got != tt.want {
				t.Errorf("ValidateDamageType(%s) = %v, want %v", tt.dt, got, tt.want)
			}
		})
	}
}

func TestGetNormalizedAbility(t *testing.T) {
	tests := []struct {
		input   string
		want    Ability
		wantErr bool
	}{
		{"str", AbilityStrength, false},
		{"Strength", AbilityStrength, false},
		{"DEX", AbilityDexterity, false},
		{"con", AbilityConstitution, false},
		{"INT", AbilityIntelligence, false},
		{"Wisdom", AbilityWisdom, false},
		{"cha", AbilityCharisma, false},
		{"invalid", AbilityNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := GetNormalizedAbility(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNormalizedAbility(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNormalizedAbility(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
