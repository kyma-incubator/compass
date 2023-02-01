package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// CertSubjectMapping is a structure that represents a certificate subject
// mapped to internal consumer with given tenant access levels
type CertSubjectMapping struct {
	ID                 string   `json:"id"`
	Subject            string   `json:"subject"`
	ConsumerType       string   `json:"consumer_type"`
	InternalConsumerID *string  `json:"internal_consumer_id"`
	TenantAccessLevels []string `json:"tenant_access_levels"`
}

// CertSubjectMappingInput is an input for creating a new CertSubjectMapping
type CertSubjectMappingInput struct {
	Subject            string   `json:"subject"`
	ConsumerType       string   `json:"consumer_type"`
	InternalConsumerID *string  `json:"internal_consumer_id"`
	TenantAccessLevels []string `json:"tenant_access_levels"`
}

// CertSubjectMappingPage contains CertSubjectMapping data with page info
type CertSubjectMappingPage struct {
	Data       []*CertSubjectMapping
	PageInfo   *pagination.Page
	TotalCount int
}

// ToModel converts CertSubjectMappingInput to CertSubjectMapping
func (c *CertSubjectMappingInput) ToModel(id string) *CertSubjectMapping {
	if c == nil {
		return nil
	}

	return &CertSubjectMapping{
		ID:                 id,
		Subject:            c.Subject,
		ConsumerType:       c.ConsumerType,
		InternalConsumerID: c.InternalConsumerID,
		TenantAccessLevels: c.TenantAccessLevels,
	}
}
