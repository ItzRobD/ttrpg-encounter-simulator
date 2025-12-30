package equipment_manager

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
	"testing"
)

func TestGetWeaponAttackData_ModifiersAndVersatile(t *testing.T) {
	parent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14}, nil)
	em, err := NewEquipmentManager(parent)
	if err != nil {
		t.Fatalf("NewEquipmentManager: %v", err)
	}

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged:    false,
			IsVersatile: true,
		},
	}
	rapier := &weapon.Weapon{
		Name:         "Rapier",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamagePiercing,
		Properties: weapon.Properties{
			IsRanged:  false,
			IsFinesse: true,
		},
	}

	tests := []struct {
		name      string
		w         *weapon.Weapon
		prof      bool
		wantAtk   int
		wantDmg   int
		versatile bool
	}{
		{"longsword proficient", longsword, true, 7, 4, false}, // +4 STR +3 prof (lvl 5)
		{"longsword versatile", longsword, true, 7, 4, true},   // same mods, bigger die
		{"rapier proficient (finesse uses best of STR/DEX)", rapier, true, 7, 4, false},
		{"longsword not proficient", longsword, false, 4, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := em.SetWeapon(core.WSPrimary, tt.w, tt.prof); err != nil {
				t.Fatalf("SetWeapon: %v", err)
			}
			ad, err := em.GetWeaponAttackData(core.WSPrimary, tt.versatile)
			if err != nil {
				t.Fatalf("GetWeaponAttackData: %v", err)
			}

			if ad.AttackModifier != tt.wantAtk {
				t.Errorf("AttackModifier=%d want %d", ad.AttackModifier, tt.wantAtk)
			}
			if ad.DamageModifier != tt.wantDmg {
				t.Errorf("DamageModifier=%d want %d", ad.DamageModifier, tt.wantDmg)
			}
			if tt.versatile != ad.IsVersatileAttack {
				t.Errorf("IsVersatileAttack=%v want %v", ad.IsVersatileAttack, tt.versatile)
			}
			if tt.versatile && ad.Die != tt.w.Die+2 {
				t.Errorf("Versatile die should be +2: got %v", ad.Die)
			}
		})
	}
}

func TestGetAC(t *testing.T) {
	// Build separate parents to differentiate monk vs non‑monk
	nonMonk := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 16, Wisdom: 18}, nil)
	monkID := uint8(classes.Monk)
	monk := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 16, Wisdom: 18}, &monkID)

	leather, _ := armor.New("Leather Armor", 11, true, false, 0)
	ringmail, _ := armor.New("Ring Mail", 14, false, false, 0)
	plate, _ := armor.New("Plate Armor", 18, false, false, 15)
	shield, _ := armor.New("Shield", 2, false, false, 0)
	chainshirt, _ := armor.New("Chain Shirt", 13, true, true, 0)

	// Use nonMonk to compute modifiers (same scores as monk parent)
	dm, _ := nonMonk.GetAbilityScoreModifier(core.AbilityDexterity) // Dex 16 => +3
	wm, _ := nonMonk.GetAbilityScoreModifier(core.AbilityWisdom)    // Wis 18 => +4

	tests := []struct {
		name   string
		armor  armor.Armor
		shield armor.Armor
		wantAC int
		isMonk bool
	}{
		{name: "leather", armor: leather, wantAC: 11 + dm},                                                // 14
		{name: "ringmail", armor: ringmail, wantAC: 14},                                                   // no Dex bonus
		{name: "plate", armor: plate, wantAC: 18},                                                         // no Dex bonus
		{name: "leather and shield", armor: leather, shield: shield, wantAC: 11 + dm + shield.ArmorClass}, // shield +2
		{name: "plate and shield", armor: plate, shield: shield, wantAC: 18 + shield.ArmorClass},          // shield +2
		{name: "chain shirt", armor: chainshirt, wantAC: 13 + 2},
		{name: "no armor", wantAC: 10 + dm},
		{name: "shield only", shield: shield, wantAC: 10 + dm + shield.ArmorClass},
		{name: "unarmed monk", isMonk: true, wantAC: 10 + wm + dm},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fresh EquipmentManager per subtest to avoid shield/armor state leakage
			var p core.Entity
			if tt.isMonk {
				p = monk
			} else {
				p = nonMonk
			}
			em, err := NewEquipmentManager(p)
			if err != nil {
				t.Fatalf("NewEquipmentManager: %v", err)
			}
			if tt.armor.ArmorClass > 0 {
				em.SetArmor(tt.armor)
			}
			if tt.shield.ArmorClass > 0 {
				em.SetShield(tt.shield)
			}
			if got := em.GetAC(); got != tt.wantAC {
				t.Errorf("GetAC() = %d, want %d", got, tt.wantAC)
			}
		})
	}
}

func TestSetWeapon(t *testing.T) {
	parent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14}, nil)
	em, err := NewEquipmentManager(parent)
	if err != nil {
		t.Fatalf("NewEquipmentManager: %v", err)
	}

	longsword := getTestWeapon(t, "Longsword")
	//club := getTestWeapon(t, "club")
	//greatsword := getTestWeapon(t, "greatsword")

	tests := []struct {
		name       string
		slot       core.WeaponSlot
		weapon     *weapon.Weapon
		proficient bool
		wantErr    bool
	}{
		{
			name:       "equip ls primary proficient",
			slot:       core.WSPrimary,
			weapon:     longsword,
			proficient: true,
			wantErr:    false,
		},
		{
			name:       "equip secondary weapon not proficient",
			slot:       core.WSSecondary,
			weapon:     longsword,
			proficient: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terr := em.SetWeapon(tt.slot, tt.weapon, tt.proficient)
			if terr != nil && !tt.wantErr {
				t.Errorf("SetWeapon() error = %v, wantErr %v", terr, tt.wantErr)
				return
			}

			if !tt.wantErr {
				w, tErr := em.GetWeaponFromSlot(tt.slot)
				if tErr != nil {
					t.Fatalf("GetWeaponFromSlot() error = %v", tErr)
				}
				if w.Name != tt.weapon.Name {
					t.Errorf("Weapon name = %s, want %s", w.Name, tt.weapon.Name)
				}

				// Verify proficiency was set
				if em.GetIsProficientWithSlot(tt.slot) != tt.proficient {
					t.Errorf("Proficiency = %v, want %v", em.GetIsProficientWithSlot(tt.slot), tt.proficient)
				}

				// Verify attack data was computed
				_, tErr = em.GetWeaponAttackData(tt.slot, false)
				if tErr != nil {
					t.Errorf("Attack data not computed: %v", tErr)
				}
			}
		})
	}
}

// Test has melee weapon check
func TestHasMeleeWeapon(t *testing.T) {
	parent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(parent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties:   weapon.Properties{IsRanged: false},
	}

	longbow := &weapon.Weapon{
		Name:         "Longbow",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamagePiercing,
		Properties:   weapon.Properties{IsRanged: true, IsOnlyRanged: true},
	}

	tests := []struct {
		name      string
		setupFunc func()
		wantMelee bool
	}{
		{
			name:      "no weapons equipped",
			setupFunc: func() {},
			wantMelee: false,
		},
		{
			name: "melee weapon equipped",
			setupFunc: func() {
				em.SetWeapon(core.WSPrimary, longsword, true)
			},
			wantMelee: true,
		},
		{
			name: "only ranged weapon equipped",
			setupFunc: func() {
				em, _ = NewEquipmentManager(parent)
				em.SetWeapon(core.WSPrimary, longbow, true)
			},
			wantMelee: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em, _ = NewEquipmentManager(parent) // Reset for each test
			tt.setupFunc()

			got := em.HasMeleeWeapon()
			if got != tt.wantMelee {
				t.Errorf("HasMeleeWeapon() = %v, want %v", got, tt.wantMelee)
			}
		})
	}
}

// Test has ranged weapon check
func TestHasRangedWeapon(t *testing.T) {
	parent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(parent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties:   weapon.Properties{IsRanged: false},
	}

	longbow := &weapon.Weapon{
		Name:         "Longbow",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamagePiercing,
		Properties:   weapon.Properties{IsRanged: true, IsOnlyRanged: true},
	}

	tests := []struct {
		name       string
		setupFunc  func()
		wantRanged bool
	}{
		{
			name:       "no weapons equipped",
			setupFunc:  func() {},
			wantRanged: false,
		},
		{
			name: "ranged weapon equipped",
			setupFunc: func() {
				em.SetWeapon(core.WSPrimary, longbow, true)
			},
			wantRanged: true,
		},
		{
			name: "only melee weapon equipped",
			setupFunc: func() {
				em, _ = NewEquipmentManager(parent)
				em.SetWeapon(core.WSPrimary, longsword, true)
			},
			wantRanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em, _ = NewEquipmentManager(parent) // Reset for each test
			tt.setupFunc()

			got := em.HasRangedWeapon()
			if got != tt.wantRanged {
				t.Errorf("HasRangedWeapon() = %v, want %v", got, tt.wantRanged)
			}
		})
	}
}

// Test versatile weapon handling
func TestCanUseVersatile(t *testing.T) {
	emParent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(emParent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged:    false,
			IsVersatile: true,
		},
	}

	dagger := &weapon.Weapon{
		Name:         "Dagger",
		NumberOfDice: 1,
		Die:          core.D4,
		DamageType:   core.DamagePiercing,
		Properties: weapon.Properties{
			IsRanged:    false,
			IsVersatile: false,
		},
	}

	tests := []struct {
		name          string
		weapon        *weapon.Weapon
		slot          core.WeaponSlot
		wantVersatile bool
	}{
		{
			name:          "versatile weapon",
			weapon:        longsword,
			slot:          core.WSPrimary,
			wantVersatile: true,
		},
		{
			name:          "non-versatile weapon",
			weapon:        dagger,
			slot:          core.WSSecondary,
			wantVersatile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em.SetWeapon(tt.slot, tt.weapon, true)

			got := em.CanUseVersatile(tt.slot)
			if got != tt.wantVersatile {
				t.Errorf("CanUseVersatile() = %v, want %v", got, tt.wantVersatile)
			}
		})
	}
}

// Test versatile damage dice upgrade
func TestGetWeaponAttackData_Versatile(t *testing.T) {
	emParent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(emParent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8, // 1d8 normal
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged:    false,
			IsVersatile: true, // 1d10 versatile (d8 + 2)
		},
	}

	em.SetWeapon(core.WSPrimary, longsword, true)

	tests := []struct {
		name         string
		useVersatile bool
		wantDie      core.DiceType
	}{
		{"normal attack", false, core.D8},
		{"versatile attack", true, core.D10}, // D8 + 2 = D10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad, err := em.GetWeaponAttackData(core.WSPrimary, tt.useVersatile)
			if err != nil {
				t.Fatalf("GetWeaponAttackData() error = %v", err)
			}

			if ad.Die != tt.wantDie {
				t.Errorf("Die = %v, want %v", ad.Die, tt.wantDie)
			}

			if ad.IsVersatileAttack != tt.useVersatile {
				t.Errorf("IsVersatileAttack = %v, want %v", ad.IsVersatileAttack, tt.useVersatile)
			}
		})
	}
}

// Test available weapon slots
func TestGetAvailableWeaponSlots(t *testing.T) {
	emParent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(emParent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged: false,
		},
	}

	tests := []struct {
		name      string
		setupFunc func()
		wantCount int
	}{
		{
			name:      "no weapons equipped",
			setupFunc: func() {},
			wantCount: 0,
		},
		{
			name: "one weapon equipped",
			setupFunc: func() {
				em.SetWeapon(core.WSPrimary, longsword, true)
			},
			wantCount: 1,
		},
		{
			name: "two weapons equipped",
			setupFunc: func() {
				em.SetWeapon(core.WSPrimary, longsword, true)
				em.SetWeapon(core.WSSecondary, longsword, true)
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em, _ = NewEquipmentManager(emParent) // Reset for each test
			tt.setupFunc()

			slots := em.GetAvailableWeaponSlots()
			if len(slots) != tt.wantCount {
				t.Errorf("GetAvailableWeaponSlots() count = %d, want %d", len(slots), tt.wantCount)
			}
		})
	}
}

// Test proficiency changes
func TestSetWeaponProficiencyBySlot(t *testing.T) {
	emParent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(emParent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged: false,
		},
	}

	// Equip weapon as not proficient
	em.SetWeapon(core.WSPrimary, longsword, false)

	if em.GetIsProficientWithSlot(core.WSPrimary) {
		t.Error("Expected not proficient initially")
	}

	// Change to proficient
	em.SetWeaponProficiencyBySlot(core.WSPrimary, true)

	if !em.GetIsProficientWithSlot(core.WSPrimary) {
		t.Error("Expected proficient after change")
	}
}

// Test attack data computation
func TestComputeAttackDataForSlot(t *testing.T) {
	emParent := testhelpers.NewEmEntity(5, core.AbilityScores{Strength: 18, Dexterity: 14, Constitution: 12, Intelligence: 8, Wisdom: 10, Charisma: 10}, nil)
	em, _ := NewEquipmentManager(emParent)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged:    false,
			IsFinesse:   false,
			IsVersatile: false,
		},
	}

	rapier := &weapon.Weapon{
		Name:         "Rapier",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamagePiercing,
		Properties: weapon.Properties{
			IsRanged:    false,
			IsFinesse:   true, // Can use DEX
			IsVersatile: false,
		},
	}

	tests := []struct {
		name          string
		weapon        *weapon.Weapon
		isProficient  bool
		wantAttackMod int // Expected attack modifier (ability + prof if proficient)
		wantDamageMod int // Expected damage modifier (ability only)
	}{
		{
			name:          "longsword proficient",
			weapon:        longsword,
			isProficient:  true,
			wantAttackMod: 7, // +4 STR + +3 prof (lvl 5)
			wantDamageMod: 4, // +4 STR
		},
		{
			name:          "longsword not proficient",
			weapon:        longsword,
			isProficient:  false,
			wantAttackMod: 4, // +4 STR only
			wantDamageMod: 4, // +4 STR
		},
		{
			name:          "rapier proficient (finesse uses best of STR/DEX)",
			weapon:        rapier,
			isProficient:  true,
			wantAttackMod: 7, // best of STR(+4) or DEX(+2) + prof(+3)
			wantDamageMod: 4, // best of STR(+4) or DEX(+2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em.SetWeapon(core.WSPrimary, tt.weapon, tt.isProficient)

			ad, err := em.GetWeaponAttackData(core.WSPrimary, false)
			if err != nil {
				t.Fatalf("GetWeaponAttackData() error = %v", err)
			}

			if ad.AttackModifier != tt.wantAttackMod {
				t.Errorf("AttackModifier = %d, want %d", ad.AttackModifier, tt.wantAttackMod)
			}

			if ad.DamageModifier != tt.wantDamageMod {
				t.Errorf("DamageModifier = %d, want %d", ad.DamageModifier, tt.wantDamageMod)
			}

			if ad.Name != tt.weapon.Name {
				t.Errorf("Name = %s, want %s", ad.Name, tt.weapon.Name)
			}

			if ad.NumberOfDice != tt.weapon.NumberOfDice {
				t.Errorf("NumberOfDice = %d, want %d", ad.NumberOfDice, tt.weapon.NumberOfDice)
			}

			if ad.Die != tt.weapon.Die {
				t.Errorf("Die = %v, want %v", ad.Die, tt.weapon.Die)
			}
		})
	}
}

func getTestArmor(t testing.TB, id int) *armor.Armor {
	t.Helper()
	dbErr := database.InitDb(&database.InitOpts{EnvPath: "../../../.env"})

	if dbErr != nil {
		fmt.Println(dbErr)
		return nil
	}
	defer database.CloseDb()

	ctx := context.Background()
	armor, err := armor.QueryArmorData(ctx, armor.ArmorQueryParams{ID: id})
	if err != nil {
		t.Fatalf("QueryArmorData: %v", err)
	}
	return &armor
}

func getTestWeapon(t testing.TB, name string) *weapon.Weapon {
	t.Helper()
	dbErr := database.InitDb(&database.InitOpts{EnvPath: "../../../.env"})

	if dbErr != nil {
		fmt.Println(dbErr)
		return nil
	}
	defer database.CloseDb()

	ctx := context.Background()
	armor, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{Name: name})
	if err != nil {
		t.Fatalf("QueryWeaponData: %v", err)
	}
	return &armor
}
