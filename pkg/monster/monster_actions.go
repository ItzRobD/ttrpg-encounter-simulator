package monster

//func (m *Monster) logHPRollEvent(rollSum int, rolls []int, toAdd int) {
//	if m.EventListener != nil {
//		event := events.CombatEvent{
//			EventType: events.ETHPRollEvent,
//			ActorName:     m.Name,
//			Value:     rollSum + toAdd,
//			Rolls:     rolls,
//			Modifier:     toAdd,
//		}
//		m.EventListener(event)
//	}
//}

//// TODO: Think about how to handle the values better. should value be the combined amount? maybe roll, rolls, added, total
//func (m *Monster) logRollEvent(rollSum int, rolls []int, toAdd int) {
//	if m.EventListener != nil {
//		event := events.CombatEvent{
//			EventType: events.ETRollEvent,
//			ActorName:     m.Name,
//			Value:     rollSum + toAdd,
//			Rolls:     rolls,
//			Modifier:     toAdd,
//		}
//		m.EventListener(event)
//	}
//}
