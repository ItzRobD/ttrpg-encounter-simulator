package monster

func (m *Monster) RefreshLegendaryActions() {
	m.EntityStateManager.ReplenishLegendaryActionPoints(m.EntityStateManager.GetLegendaryActionPointsMax())
}
