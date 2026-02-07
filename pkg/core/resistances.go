package core

import (
	"fmt"
	"strings"
)

type DamageResistances map[DamageType]DamageResistance

type DamageResistance struct {
	Resistance ResistanceType  `json:"resistance"`
	Breakers   []ResistBreaker `json:"breakers"`
}

func NewEmptyDamageResistance() DamageResistance {
	return DamageResistance{
		Resistance: ResistanceNone,
		Breakers:   make([]ResistBreaker, 0),
	}
}

func NewDamageResistance(rt ResistanceType, rb []ResistBreaker) DamageResistance {
	return DamageResistance{
		Resistance: rt,
		Breakers:   rb,
	}
}

type ResistanceType string

const (
	ResistanceNone       ResistanceType = "none"
	ResistanceResistant  ResistanceType = "resist"
	ResistanceVulnerable ResistanceType = "vulnerable"
	ResistanceImmune     ResistanceType = "immune"
)

func (rt ResistanceType) String() string {
	return string(rt)
}

func MakeResistanceType(s string) (ResistanceType, error) {
	switch strings.ToLower(s) {
	case "none":
		return ResistanceNone, nil
	case "resistant":
		return ResistanceResistant, nil
	case "vulnerable":
		return ResistanceVulnerable, nil
	case "immune":
		return ResistanceImmune, nil
	default:
		return ResistanceNone, fmt.Errorf("invalid resistance type")
	}
}

func NewDamageResistances() DamageResistances {
	return map[DamageType]DamageResistance{
		DamageAcid:        NewEmptyDamageResistance(),
		DamageCold:        NewEmptyDamageResistance(),
		DamageFire:        NewEmptyDamageResistance(),
		DamageForce:       NewEmptyDamageResistance(),
		DamageLightning:   NewEmptyDamageResistance(),
		DamageNecrotic:    NewEmptyDamageResistance(),
		DamagePoison:      NewEmptyDamageResistance(),
		DamagePsychic:     NewEmptyDamageResistance(),
		DamageRadiant:     NewEmptyDamageResistance(),
		DamageThunder:     NewEmptyDamageResistance(),
		DamageSlashing:    NewEmptyDamageResistance(),
		DamageBludgeoning: NewEmptyDamageResistance(),
		DamagePiercing:    NewEmptyDamageResistance(),
	}
}

func NewDamageResistancesAll(rt ResistanceType) DamageResistances {
	return map[DamageType]DamageResistance{
		DamageAcid:        NewDamageResistance(rt, nil),
		DamageCold:        NewDamageResistance(rt, nil),
		DamageFire:        NewDamageResistance(rt, nil),
		DamageForce:       NewDamageResistance(rt, nil),
		DamageLightning:   NewDamageResistance(rt, nil),
		DamageNecrotic:    NewDamageResistance(rt, nil),
		DamagePoison:      NewDamageResistance(rt, nil),
		DamagePsychic:     NewDamageResistance(rt, nil),
		DamageRadiant:     NewDamageResistance(rt, nil),
		DamageThunder:     NewDamageResistance(rt, nil),
		DamageSlashing:    NewDamageResistance(rt, nil),
		DamageBludgeoning: NewDamageResistance(rt, nil),
		DamagePiercing:    NewDamageResistance(rt, nil),
	}
}

func (dr DamageResistances) GetResistanceType(dt DamageType) ResistanceType {
	if rt, ok := dr[dt]; ok {
		return rt.Resistance
	}
	return ResistanceNone
}

func (dr DamageResistances) GetResistance(dt DamageType) DamageResistance {
	if res, ok := dr[dt]; ok {
		return res
	}
	// Return a safe default instead of zero-value so callers never see empty ResistanceType
	return NewEmptyDamageResistance()
}

func (dr DamageResistances) SetResistance(d DamageType, rt ResistanceType, rb []ResistBreaker) {
	dr[d] = NewDamageResistance(rt, rb)
}

func (dr DamageResistances) SetResistanceType(dt DamageType, rt ResistanceType) {
	res := dr[dt]
	res.Resistance = rt
	dr[dt] = res
}

func (dr DamageResistances) ResetResistance(dt DamageType) {
	dr[dt] = NewEmptyDamageResistance()
}

func (dr DamageResistances) SetPhysicalResistance(rt ResistanceType) {
	res := dr[DamageSlashing]
	res.Resistance = rt
	dr[DamageSlashing] = res
	res = dr[DamageBludgeoning]
	res.Resistance = rt
	dr[DamageBludgeoning] = res
	res = dr[DamagePiercing]
	res.Resistance = rt
	dr[DamagePiercing] = res
}

func (dr DamageResistances) AddBreaker(dt DamageType, rb ResistBreaker) {
	res := dr[dt]
	res.Breakers = append(res.Breakers, rb)
	dr[dt] = res
}

func (dr DamageResistances) RemoveBreaker(dt DamageType, rb ResistBreaker) {
	res := dr[dt]
	for i, br := range res.Breakers {
		if br == rb {
			res.Breakers = append(res.Breakers[:i], res.Breakers[i+1:]...)
			break
		}
	}
}

func (dr DamageResistances) GetBreakers(dt DamageType) []ResistBreaker {
	if res, ok := dr[dt]; ok {
		return res.Breakers
	}
	return nil
}

func (dr DamageResistances) DamageTypeContainsBreaker(dt DamageType, rb ResistBreaker) bool {
	if res, ok := dr[dt]; ok {
		for _, br := range res.Breakers {
			if br == rb {
				return true
			}
		}
	}
	return false
}

func (dr DamageResistances) DamageTypeContainsAllBreakers(dt DamageType, rb []ResistBreaker) bool {
	res, ok := dr[dt]
	if !ok {
		return false
	}

	for _, br := range rb {
		found := false
		for _, existingBreaker := range res.Breakers {
			if existingBreaker == br {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type ResistBreaker string

const (
	ResistBreakerNone           ResistBreaker = "none"
	ResistBreakerMagic          ResistBreaker = "magic"
	ResistBreakerSilvered       ResistBreaker = "silvered"
	ResistBreakerAdamantine     ResistBreaker = "adamantine"
	ResistBreakerColdForgedIron ResistBreaker = "cold forged iron"
)

func (rb ResistBreaker) String() string {
	return string(rb)
}

func MakeResistBreaker(s string) (ResistBreaker, error) {
	switch strings.ToLower(s) {
	case "none", "":
		return ResistBreakerNone, nil
	case "magic":
		return ResistBreakerMagic, nil
	case "silvered":
		return ResistBreakerSilvered, nil
	case "adamantine":
		return ResistBreakerAdamantine, nil
	case "cold forged iron":
		return ResistBreakerColdForgedIron, nil
	default:
		return ResistBreakerNone, fmt.Errorf("invalid resistance breaker")
	}
}
