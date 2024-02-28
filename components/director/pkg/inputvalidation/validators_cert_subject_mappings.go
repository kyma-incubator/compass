package inputvalidation

import (
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

const (
	// GlobalAccessLevel is an access level that is not tied to a specific tenant entity but rather it's used globally
	GlobalAccessLevel = "global"
)

// SupportedConsumerTypes is a map of all supported consumer types
var SupportedConsumerTypes = map[consumer.Type]bool{
	consumer.Runtime:                            true,
	consumer.IntegrationSystem:                  true,
	consumer.Application:                        true,
	consumer.SuperAdmin:                         true,
	consumer.BusinessIntegration:                true,
	consumer.ManagedApplicationProviderOperator: true,
	consumer.ManagedApplicationConsumer:         true,
	consumer.LandscapeResourceOperator:          true,
	consumer.TenantDiscoveryOperator:            true,
	consumer.InstanceCreator:                    true,
	consumer.TechnicalClient:                    true,
}

// SupportedAccessLevels is a map of all supported tenant access levels
var SupportedAccessLevels = map[string]bool{
	string(tenantEntity.Customer):      true,
	string(tenantEntity.Account):       true,
	string(tenantEntity.Subaccount):    true,
	string(tenantEntity.Organization):  true,
	string(tenantEntity.Folder):        true,
	string(tenantEntity.ResourceGroup): true,
	string(tenantEntity.CostObject):    true,
	GlobalAccessLevel:                  true,
}

type certMappingSubjectValidator struct{}
type certMappingConsumerTypeValidator struct{}
type certMappingTenantAccessLevelValidator struct{}

// IsValidCertSubject is a custom validation rule that validates certificate subject mapping's subject input
var IsValidCertSubject = &certMappingSubjectValidator{}

// IsValidConsumerType is a custom validation rule that validates certificate subject mapping's consumer type input
var IsValidConsumerType = &certMappingConsumerTypeValidator{}

// AreTenantAccessLevelsValid is a custom validation rule that validates certificate subject mapping's tenant access levels input
var AreTenantAccessLevelsValid = &certMappingTenantAccessLevelValidator{}

func (v *certMappingSubjectValidator) Validate(value interface{}) error {
	s, isNil, err := ensureIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	expectedSubjectComponents := strings.Split(s, ",")
	if len(expectedSubjectComponents) < 5 { // 5 because that's the number of certificate relative distinguished names that we expect - CountryName(C), Organization(O), OrganizationalUnit(OU), Locality(L) and CommonName(CN)
		return errors.Errorf("the number of certificate attributes are different than the expected ones. We got: %d and we need at least 5 - C, O, OU, L and CN", len(expectedSubjectComponents))
	}

	if country := cert.GetCountry(s); country == "" {
		return errors.New("missing Country property in the subject")
	}

	if org := cert.GetOrganization(s); org == "" {
		return errors.New("missing Organization property in the subject")
	}

	OUs := cert.GetAllOrganizationalUnits(s)
	if len(OUs) < 1 {
		return errors.New("missing Organization Unit property in the subject")
	}

	if locality := cert.GetLocality(s); locality == "" {
		return errors.New("missing Locality property in the subject")
	}

	if cm := cert.GetCommonName(s); cm == "" {
		return errors.New("missing Common Name property in the subject")
	}

	return nil
}

func (v certMappingConsumerTypeValidator) Validate(value interface{}) error {
	consumerType, isNil, err := ensureIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	if !SupportedConsumerTypes[consumer.Type(consumerType)] {
		return errors.Errorf("consumer type %s is not valid", consumerType)
	}

	return nil
}

func (v certMappingTenantAccessLevelValidator) Validate(value interface{}) error {
	tenantAccessLevels, ok := value.([]string)
	if !ok {
		return errors.Errorf("invalid type, expected []string, got: %T", value)
	}

	for _, al := range tenantAccessLevels {
		if !SupportedAccessLevels[al] {
			return errors.Errorf("tenant access level %s is not valid", al)
		}
	}

	return nil
}
