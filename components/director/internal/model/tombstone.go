package model

// Tombstone missing godoc
type Tombstone struct {
	ID            string
	OrdID         string
	ApplicationID string
	RemovalDate   string
}

// TombstoneInput missing godoc
type TombstoneInput struct {
	OrdID       string `json:"ordId"`
	RemovalDate string `json:"removalDate"`
}

// ToTombstone missing godoc
func (i *TombstoneInput) ToTombstone(id, appID string) *Tombstone {
	if i == nil {
		return nil
	}

	return &Tombstone{
		ID:            id,
		OrdID:         i.OrdID,
		ApplicationID: appID,
		RemovalDate:   i.RemovalDate,
	}
}

// SetFromUpdateInput missing godoc
func (p *Tombstone) SetFromUpdateInput(update TombstoneInput) {
	p.RemovalDate = update.RemovalDate
}
