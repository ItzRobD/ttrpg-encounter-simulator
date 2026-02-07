package monster

import "dnd5e-encounter-simulator-backend/pkg_old/core"

func (m *Monster) GetCasterLevel() int { return m.SpellCastingManager.GetCasterLevel() }

func (m *Monster) GetSpellSaveDC(ability *core.Ability) (int, error) {
	return m.SpellCastingManager.GetSaveDC(), nil
}

func (m *Monster) GetHealingSpellCount() int {
	return m.SpellCastingManager.GetHealingSpellCount()
}

func (m *Monster) ChooseSpellByHealingEfficiency(targetValue int) (*core.SpellChoice, error) {
	choice, err := m.SpellCastingManager.GetMostEfficientHealingSpell(targetValue)
	if err != nil {
		return nil, err
	}
	return choice, nil
}

func (m *Monster) ChooseDamageSpellByPriority(p core.SpellPriority) (*core.SpellChoice, error) {
	return m.SpellCastingManager.ChooseSpellByPriority(core.STDamage, p)
}

func (m *Monster) GetDamageSpellCount() int {
	return m.SpellCastingManager.GetDamageSpellCount()
}
