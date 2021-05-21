package model

type Tombstone struct {
	OrdID         string
	TenantID      string
	ApplicationID string
	RemovalDate   string
}

type TombstoneInput struct {
	OrdID       string `json:"ordId"`
	RemovalDate string `json:"removalDate"`
}

func (i *TombstoneInput) ToTombstone(tenantID, appID string) *Tombstone {
	if i == nil {
		return nil
	}

	return &Tombstone{
		OrdID:         i.OrdID,
		TenantID:      tenantID,
		ApplicationID: appID,
		RemovalDate:   i.RemovalDate,
	}
}

func (p *Tombstone) SetFromUpdateInput(update TombstoneInput) {
	p.RemovalDate = update.RemovalDate
}
