package subject

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/hydrator/internal/certsubjectmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
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

type processor struct {
	certSubjectMappingCache certsubjectmapping.Cache
	ouPattern               string
	ouRegionPattern         string
}

// NewProcessor returns a new subject processor configured with the given subject-to-consumer mapping cache, and subject organization unit pattern.
// If the subject-to-consumer mapping is invalid, an error is returned.
func NewProcessor(ctx context.Context, certSubjectMappingCache certsubjectmapping.Cache, ouPattern, ouRegionPattern string) (*processor, error) {
	mappings := certSubjectMappingCache.Get()
	if len(mappings) == 0 {
		log.C(ctx).Warnf("The certificate subject mapping cache is empty.")
	}

	for _, m := range mappings {
		if err := m.Validate(); err != nil {
			return nil, err
		}
	}

	return &processor{
		certSubjectMappingCache: certSubjectMappingCache,
		ouPattern:               ouPattern,
		ouRegionPattern:         ouRegionPattern,
	}, nil
}

// AuthIDFromSubjectFunc returns a function able to extract the authentication ID from a given certificate subject or from the certificate subject mappings.
func (p *processor) AuthIDFromSubjectFunc(ctx context.Context) func(subject string) string {
	authIDFromMappingFunc := p.authIDFromMappings()
	authIDFromOUsFunc := cert.GetRemainingOrganizationalUnit(p.ouPattern, p.ouRegionPattern)

	return func(subject string) string {
		if authIDFromMapping := authIDFromMappingFunc(subject); authIDFromMapping != "" {
			log.C(ctx).Infof("Extracting Auth ID from certificate subject mapping cache")
			return authIDFromMapping
		}

		log.C(ctx).Infof("Extracting Auth ID from the certificate attributes(OU)")
		return authIDFromOUsFunc(subject)
	}
}

// EmptyAuthSessionExtraFunc returns a function which returns an empty auth session body extra.
func (p *processor) EmptyAuthSessionExtraFunc() func(context.Context, string) map[string]interface{} {
	return func(ctx context.Context, subject string) map[string]interface{} { return nil }
}

// AuthSessionExtraFromSubjectFunc returns a function which returns consumer-specific auth session extra body
// in case the subject matches any of the configured certificate subject mappings.
func (p *processor) AuthSessionExtraFromSubjectFunc() func(context.Context, string) map[string]interface{} {
	return func(ctx context.Context, subject string) map[string]interface{} {
		log.C(ctx).Infof("Trying to extract auth session extra from consumer subject: %q", subject)

		mappings := p.certSubjectMappingCache.Get()
		if len(mappings) == 0 {
			log.C(ctx).Warnf("The certificate subject mapping cache is empty.")
		}

		for _, m := range mappings {
			log.C(ctx).Infof("Trying to match the consumer subject DN with certificate subject mappings DN: %q", m.Subject)
			if subjectsMatch(subject, m.Subject) {
				log.C(ctx).Infof("Subject's DNs matched!")
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
		mappings := p.certSubjectMappingCache.Get()
		for _, m := range mappings {
			if subjectsMatch(subject, m.Subject) {
				return m.InternalConsumerID
			}
		}
		return ""
	}
}

func subjectsMatch(actualSubject, expectedSubject string) bool {
	return cert.GetCommonName(expectedSubject) == cert.GetCommonName(actualSubject) &&
		cert.GetCountry(expectedSubject) == cert.GetCountry(actualSubject) &&
		cert.GetLocality(expectedSubject) == cert.GetLocality(actualSubject) &&
		cert.GetOrganization(expectedSubject) == cert.GetOrganization(actualSubject) &&
		matchOrganizationalUnits(cert.GetAllOrganizationalUnits(actualSubject), cert.GetAllOrganizationalUnits(expectedSubject))
}

func matchOrganizationalUnits(actualOrgUnits, expectedOrgUnits []string) bool {
	if len(expectedOrgUnits) != len(actualOrgUnits) {
		return false
	}

	expectedOrgUnitsMap := make(map[string]struct{}, len(expectedOrgUnits))
	for _, expectedOrgUnit := range expectedOrgUnits {
		expectedOrgUnitsMap[strings.TrimSpace(expectedOrgUnit)] = struct{}{}
	}

	for _, actualOrgUnit := range actualOrgUnits {
		if _, exist := expectedOrgUnitsMap[strings.TrimSpace(actualOrgUnit)]; !exist {
			return false
		}
	}

	return true
}
