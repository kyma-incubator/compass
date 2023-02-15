package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// Validate validates the CertificateSubjectMappingInput structure's properties
func (i CertificateSubjectMappingInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Subject, validation.Required, validation.Length(1, 255), inputvalidation.IsValidCertSubject),
		validation.Field(&i.ConsumerType, validation.Required, validation.Length(1, 255), inputvalidation.IsValidConsumerType),
		validation.Field(&i.InternalConsumerID, validation.Length(1, 255)),
		validation.Field(&i.TenantAccessLevels, validation.Required, inputvalidation.AreTenantAccessLevelsValid),
	)
}
