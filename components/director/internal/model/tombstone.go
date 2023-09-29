package model

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// Tombstone missing godoc
type Tombstone struct {
	ID                           string
	OrdID                        string
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	RemovalDate                  string
	Description                  *string
}

// TombstoneInput missing godoc
type TombstoneInput struct {
	OrdID       string  `json:"ordId"`
	RemovalDate string  `json:"removalDate"`
	Description *string `json:"description"`
}

// ToTombstone missing godoc
func (i *TombstoneInput) ToTombstone(id string, resourceType resource.Type, resourceID string) *Tombstone {
	if i == nil {
		return nil
	}

	tombstone := &Tombstone{
		ID:          id,
		OrdID:       i.OrdID,
		RemovalDate: i.RemovalDate,
		Description: i.Description,
	}

	if resourceType.IsTenantIgnorable() {
		tombstone.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		tombstone.ApplicationID = &resourceID
	}

	return tombstone
}

// SetFromUpdateInput missing godoc
func (p *Tombstone) SetFromUpdateInput(update TombstoneInput) {
	p.RemovalDate = update.RemovalDate
}
