package model

type Tombstone struct {
	ID            string
	OrdID         string
	TenantID      string
	ApplicationID string
	RemovalDate   string
}

type TombstoneInput struct {
	OrdID       string `json:"ordId"`
	RemovalDate string `json:"removalDate"`
}

func (i *TombstoneInput) ToTombstone(id, tenantID, appID string) *Tombstone {
	if i == nil {
		return nil
	}

	return &Tombstone{
		ID:            id,
		OrdID:         i.OrdID,
		TenantID:      tenantID,
		ApplicationID: appID,
		RemovalDate:   i.RemovalDate,
	}
}

func (p *Tombstone) SetFromUpdateInput(update TombstoneInput) {
	p.RemovalDate = update.RemovalDate
}
