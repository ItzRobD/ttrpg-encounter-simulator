package monster

func (m *Monster) RefreshLegendaryActions() {
	m.EntityState.ReplenishLegendaryActionPoints(m.EntityState.LegendaryActionPointsMax)
}
