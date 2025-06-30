package core

func DoesAttackHit(attackTotal int, ac int) bool {
	if attackTotal >= ac {
		return true
	}
	return false
}

func IsCriticalHit(attackRoll int, critThreshold int) bool {
	if attackRoll >= critThreshold {
		return true
	}
	return false
}
