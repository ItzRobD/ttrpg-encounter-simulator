package martial_attack_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"math/rand/v2"
	"testing"
)

// emEntityRNG wraps the testhelpers.EmEntity to provide a deterministic RNG
type emEntityRNG struct {
	testhelpers.EmEntity
	rng *rand.Rand
}

func (e emEntityRNG) Regenerate() {
	//TODO implement me
	panic("implement me")
}

func (e emEntityRNG) GetRNG() *rand.Rand { return e.rng }
func (e emEntityRNG) GetName() string    { return "Actor" }

// targetStub embeds EmEntity and overrides AC/Name
type targetStub struct {
	testhelpers.EmEntity
	ac int
}

func (t targetStub) Regenerate() {
	//TODO implement me
	panic("implement me")
}

func (t targetStub) GetAC() int      { return t.ac }
func (t targetStub) GetName() string { return "Target" }

func newDeterministicParents() (core.Entity, core.Entity) {
	rng1 := rand.New(rand.NewPCG(1, 2))
	rng2 := rand.New(rand.NewPCG(1, 2)) // identical sequence to compare option effects
	as := core.AbilityScores{Strength: 16, Dexterity: 14}
	p1 := emEntityRNG{EmEntity: testhelpers.NewEmEntity(5, as, nil), rng: rng1}
	p2 := emEntityRNG{EmEntity: testhelpers.NewEmEntity(5, as, nil), rng: rng2}
	return p1, p2
}

func TestProcessAttackRequest_CountAndHitMiss(t *testing.T) {
	parent, _ := newDeterministicParents()
	rm := roll_manager.NewRollManager(parent, roll_manager.RerollAbilities{})
	mam := NewMartialAttackManager(parent, rm)

	ad := []core.AttackData{
		{Name: "Slash", NumberOfDice: 1, Die: core.D8, AttackModifier: 5, DamageModifier: 3, DamageType: core.DamageSlashing},
		{Name: "Stab", NumberOfDice: 2, Die: core.D6, AttackModifier: 4, DamageModifier: 2, DamageType: core.DamagePiercing},
	}

	// Advantage and 2 attacks each
	opts := core.AttackOptions{Advantage: core.RollAdvantage, NumberOfAttacks: 2}

	// Miss path: AC very high
	missReq := &core.AttackRequest{AttackData: ad, AttackOptions: opts, Target: targetStub{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), ac: 30}}
	results, err := mam.ProcessAttackRequest(missReq)
	if err != nil {
		t.Fatalf("miss path error: %v", err)
	}
	want := len(ad) * opts.GetNumberOfAttacks()
	if len(results) != want {
		t.Fatalf("results len=%d want %d", len(results), want)
	}
	if results[0].AttackName == "" || results[0].AttackRoll == 0 {
		t.Errorf("attack fields not populated: %+v", results[0])
	}
	if results[0].DamageRoll == nil {
		t.Errorf("expected non-nil DamageRoll result")
	}

	// Hit path: AC = 0
	hitReq := &core.AttackRequest{AttackData: ad, AttackOptions: opts, Target: targetStub{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), ac: 0}}
	results, err = mam.ProcessAttackRequest(hitReq)
	if err != nil {
		t.Fatalf("hit path error: %v", err)
	}
	for _, r := range results {
		if !r.IsHit {
			t.Errorf("expected hit with AC=0, got IsHit=false: %+v", r)
		}
		if r.DamageRoll == nil {
			t.Errorf("expected non-nil DamageRoll on hit")
		}
	}
}

func TestProcessAttackRequest_OptionBonusAffectsTotal(t *testing.T) {
	// Build two managers with identical RNG sequences
	pA, pB := newDeterministicParents()
	rmA := roll_manager.NewRollManager(pA, roll_manager.RerollAbilities{})
	rmB := roll_manager.NewRollManager(pB, roll_manager.RerollAbilities{})
	mamA := NewMartialAttackManager(pA, rmA)
	mamB := NewMartialAttackManager(pB, rmB)

	ad := []core.AttackData{{Name: "Slash", NumberOfDice: 1, Die: core.D8, AttackModifier: 3, DamageModifier: 0, DamageType: core.DamageSlashing}}
	base := core.AttackOptions{NumberOfAttacks: 1}
	withBonus := core.AttackOptions{NumberOfAttacks: 1, BonusToAttackRoll: 5}
	tgt := targetStub{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), ac: 10}

	resA, err := mamA.ProcessAttackRequest(&core.AttackRequest{AttackData: ad, AttackOptions: base, Target: tgt})
	if err != nil || len(resA) != 1 {
		t.Fatalf("base run err=%v len=%d", err, len(resA))
	}
	resB, err := mamB.ProcessAttackRequest(&core.AttackRequest{AttackData: ad, AttackOptions: withBonus, Target: tgt})
	if err != nil || len(resB) != 1 {
		t.Fatalf("bonus run err=%v len=%d", err, len(resB))
	}
	if resB[0].AttackTotal <= resA[0].AttackTotal {
		t.Errorf("expected AttackTotal to increase with bonus: base=%d bonus=%d", resA[0].AttackTotal, resB[0].AttackTotal)
	}
}
