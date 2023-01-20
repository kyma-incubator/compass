package certsubjectmapping

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct{}

// NewConverter creates a new certificate subject mapping converter
func NewConverter() *converter {
	return &converter{}
}

// FromGraphql converts from graphql.CertificateSubjectMappingInput to model.CertSubjectMapping
func (c *converter) FromGraphql(id string, in graphql.CertificateSubjectMappingInput) *model.CertSubjectMapping {
	return &model.CertSubjectMapping{
		ID:                 id,
		Subject:            in.Subject,
		ConsumerType:       in.ConsumerType,
		InternalConsumerID: in.InternalConsumerID,
		TenantAccessLevels: in.TenantAccessLevels,
	}
}

// ToGraphQL converts from model.CertSubjectMapping to graphql.CertificateSubjectMapping
func (c *converter) ToGraphQL(in *model.CertSubjectMapping) *graphql.CertificateSubjectMapping {
	if in == nil {
		return nil
	}

	return &graphql.CertificateSubjectMapping{
		ID: in.ID,
		Subject: in.Subject,
		ConsumerType: in.ConsumerType,
		InternalConsumerID: in.InternalConsumerID,
		TenantAccessLevels: in.TenantAccessLevels,
	}
}

// MultipleToGraphQL converts multiple model.CertSubjectMapping models to graphql.CertificateSubjectMapping
func (c *converter) MultipleToGraphQL(in []*model.CertSubjectMapping) []*graphql.CertificateSubjectMapping {
	if in == nil {
		return nil
	}

	certSubjectMappings := make([]*graphql.CertificateSubjectMapping, 0, len(in))
	for _, i := range in {
		if i == nil {
			continue
		}
		certSubjectMappings = append(certSubjectMappings, c.ToGraphQL(i))
	}

	return certSubjectMappings
}

// ToEntity converts model.CertSubjectMapping to Entity
func (c *converter) ToEntity(in *model.CertSubjectMapping) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	marshalledTntAccessLevels, err := json.Marshal(in.TenantAccessLevels)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling tenant access levels")
	}

	return &Entity{
		ID:                 in.ID,
		Subject:            in.Subject,
		ConsumerType:       in.ConsumerType,
		InternalConsumerID: in.InternalConsumerID,
		TenantAccessLevels: string(marshalledTntAccessLevels),
	}, nil
}

// FromEntity converts Entity to model.CertSubjectMapping
func (c *converter) FromEntity(e *Entity) (*model.CertSubjectMapping, error) {
	if e == nil {
		return nil, nil
	}

	var unmarshalledTntAccessLevels	[]string
	err := json.Unmarshal([]byte(e.TenantAccessLevels), &unmarshalledTntAccessLevels)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling tenant access levels")
	}

	return &model.CertSubjectMapping{
		ID:                 e.ID,
		Subject:            e.Subject,
		ConsumerType:       e.ConsumerType,
		InternalConsumerID: e.InternalConsumerID,
		TenantAccessLevels: unmarshalledTntAccessLevels,
	}, nil
}
