package equipment

type EquipmentConfig struct {
	ID   string        `json:"id"`
	Type EquipmentType `json:"type"`
	Slot string        `json:"slot"`
}
