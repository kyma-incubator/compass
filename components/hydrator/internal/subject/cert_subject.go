package subject

import (
	"context"
	"github.com/kyma-incubator/compass/components/hydrator/internal/certsubjectmapping"
	"strings"

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
func NewProcessor(certSubjectMappingCache certsubjectmapping.Cache, ouPattern, ouRegionPattern string) (*processor, error) {
	mappings := certSubjectMappingCache.Get()
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
		mappings := p.certSubjectMappingCache.Get()
		for _, m := range mappings {
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
	expectedSubjectComponents := strings.Split(expectedSubject, ",")

	for _, expectedSubjectComponent := range expectedSubjectComponents {
		if !strings.Contains(actualSubject, strings.TrimSpace(expectedSubjectComponent)) {
			return false
		}
	}

	return true
}
