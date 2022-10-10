package subject

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

const (
	RuntimeType           = "Runtime"
	IntegrationSystemType = "Integration System"
	ApplicationType       = "Application"
	SuperAdminType        = "Super Admin"
	TechnicalClient       = "Technical Client"
)

type CSRSubjectConfig struct {
	Country            string `envconfig:"default=PL"`
	Organization       string `envconfig:"default=Org"`
	OrganizationalUnit string `envconfig:"default=OrgUnit"`
	Locality           string `envconfig:"default=Locality"`
	Province           string `envconfig:"default=State"`
}

type ExternalIssuerSubjectConfig struct {
	Country                         string `envconfig:"default=DE"`
	Organization                    string `envconfig:"default=Org"`
	OrganizationalUnitPattern       string `envconfig:"default=OrgUnit"`
	OrganizationalUnitRegionPattern string `envconfig:"default=Region"`
}

type subjectConsumerTypeMapping struct {
	Subject            string   `json:"subject"`
	ConsumerType       string   `json:"consumer_type"`
	InternalConsumerID string   `json:"internal_consumer_id"`
	TenantAccessLevels []string `json:"tenant_access_levels"`
}

func (s *subjectConsumerTypeMapping) validate() error {
	if len(s.Subject) < 1 {
		return errors.New("subject is not provided")
	}

	supportedConsumerTypes := map[string]bool{
		RuntimeType:           true,
		IntegrationSystemType: true,
		ApplicationType:       true,
		SuperAdminType:        true,
		TechnicalClient:       true,
	}

	supportedTenantTypes := map[string]bool{
		string(tenantEntity.Customer):   true,
		string(tenantEntity.Account):    true,
		string(tenantEntity.Subaccount): true,
	}

	if !supportedConsumerTypes[s.ConsumerType] {
		return fmt.Errorf("consumer type %s is not valid", s.ConsumerType)
	}
	for _, al := range s.TenantAccessLevels {
		if !supportedTenantTypes[al] {
			return fmt.Errorf("tenant access level %s is not valid", al)
		}
	}

	return nil
}

type processor struct {
	mappings        []subjectConsumerTypeMapping
	ouPattern       string
	ouRegionPattern string
}

// NewProcessor returns a new subject processor configured with the given subject-to-consumer mapping, and subject organization unit pattern.
// If the subject-to-consumer mapping is invalid, an error is returned.
func NewProcessor(subjectConsumerTypeMappingConfig string, ouPattern string, ouRegionPattern string) (*processor, error) {
	mappings, err := unmarshalMappings(subjectConsumerTypeMappingConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while configuring subject processor")
	}
	for _, m := range mappings {
		if err := m.validate(); err != nil {
			return nil, err
		}
	}
	return &processor{
		mappings:        mappings,
		ouPattern:       ouPattern,
		ouRegionPattern: ouRegionPattern,
	}, nil
}

// AuthIDFromSubjectFunc returns a function able to extract the authentication ID from a given certificate subject.
func (p *processor) AuthIDFromSubjectFunc() func(subject string) string {
	authIDFromMappingFunc := p.authIDFromMappings()
	authIDFromOUsFunc := cert.GetRemainingOrganizationalUnit(p.ouPattern, p.ouRegionPattern)
	return func(subject string) string {
		if authIDFromMapping := authIDFromMappingFunc(subject); authIDFromMapping != "" {
			return authIDFromMapping
		}
		return authIDFromOUsFunc(subject)
	}
}

// EmptyAuthSessionExtraFunc returns a function which returns an empty auth session body extra.
func (p *processor) EmptyAuthSessionExtraFunc() func(context.Context, string) map[string]interface{} {
	return func(ctx context.Context, subject string) map[string]interface{} { return nil }
}

// AuthSessionExtraFromSubjectFunc returns a function which returns consumer-specific auth session extra body
// in case the subject matches any of the configured consumers from the mapping.
func (p *processor) AuthSessionExtraFromSubjectFunc() func(context.Context, string) map[string]interface{} {
	return func(ctx context.Context, subject string) map[string]interface{} {
		log.C(ctx).Infof("trying to extract auth session extra from subject %s", subject)
		for _, m := range p.mappings {
			log.C(ctx).Infof("trying to match subject pattern %s", m.Subject)
			if subjectsMatch(subject, m.Subject) {
				log.C(ctx).Infof("pattern matched subject!")
				return cert.GetAuthSessionExtra(m.ConsumerType, m.InternalConsumerID, m.TenantAccessLevels)
			}
		}

		return nil
	}
}

// ExternalCertIssuerSubjectMatcher returns a function matching certificate subjects issued by the external trusted issuer configured
// It checks Country, Organization as single values and OrganizationalUnit as regex pattern for easier matching of multiple values.
func ExternalCertIssuerSubjectMatcher(externalSubjectConsts ExternalIssuerSubjectConfig) func(subject string) bool {
	return func(subject string) bool {
		if cert.GetCountry(subject) != externalSubjectConsts.Country || cert.GetOrganization(subject) != externalSubjectConsts.Organization {
			return false
		}
		return len(cert.GetRemainingOrganizationalUnit(externalSubjectConsts.OrganizationalUnitPattern, externalSubjectConsts.OrganizationalUnitRegionPattern)(subject)) > 0
	}
}

// ConnectorCertificateSubjectMatcher returns a function matching certificate subjects issued by compass's connector
func ConnectorCertificateSubjectMatcher(CSRSubjectConsts CSRSubjectConfig) func(subject string) bool {
	return func(subject string) bool {
		return cert.GetOrganization(subject) == CSRSubjectConsts.Organization && cert.GetOrganizationalUnit(subject) == CSRSubjectConsts.OrganizationalUnit &&
			cert.GetCountry(subject) == CSRSubjectConsts.Country && cert.GetLocality(subject) == CSRSubjectConsts.Locality && cert.GetProvince(subject) == CSRSubjectConsts.Province
	}
}

func (p *processor) authIDFromMappings() func(subject string) string {
	return func(subject string) string {
		for _, m := range p.mappings {
			if subjectsMatch(subject, m.Subject) {
				return m.InternalConsumerID
			}
		}
		return ""
	}
}

func unmarshalMappings(mappingsConfig string) ([]subjectConsumerTypeMapping, error) {
	var mappings []subjectConsumerTypeMapping
	if err := json.Unmarshal([]byte(mappingsConfig), &mappings); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling mappings")
	}

	return mappings, nil
}

func subjectsMatch(actualSubject, expectedSubject string) bool {
	expectedSubjectComponents := strings.Split(expectedSubject, ",")

	for _, expectedSubjectComponent := range expectedSubjectComponents {
		if !strings.Contains(actualSubject, strings.TrimSpace(expectedSubjectComponent)) {
			return false
		}
	}

	return true
}
