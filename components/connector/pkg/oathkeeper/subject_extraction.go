package oathkeeper

import (
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"regexp"

	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
)

// ExternalCertIssuerSubjectMatcher returns a function matching certificate subjects issued by the external trusted issuer configured
// It checks Country, Organization as single values and OrganizationalUnit as regex pattern for easier matching of multiple values.
func ExternalCertIssuerSubjectMatcher(externalSubjectConsts certificates.ExternalIssuerSubjectConsts) func(subject string) bool {
	return func(subject string) bool {
		if cert.GetCountry(subject) != externalSubjectConsts.Country || cert.GetOrganization(subject) != externalSubjectConsts.Organization {
			return false
		}
		orgUnitRegex := regexp.MustCompile(externalSubjectConsts.OrganizationalUnitPattern)
		orgUnits := GetAllOrganizationalUnits(subject)
		for _, orgUnit := range orgUnits {
			if !orgUnitRegex.MatchString(orgUnit) {
				return false
			}
		}
		return true
	}
}

// ConnectorCertificateSubjectMatcher returns a function matching certificate subjects issued by compass's connector
func ConnectorCertificateSubjectMatcher(CSRSubjectConsts certificates.CSRSubjectConsts) func(subject string) bool {
	return func(subject string) bool {
		return cert.GetOrganization(subject) == CSRSubjectConsts.Organization && cert.GetOrganizationalUnit(subject) == CSRSubjectConsts.OrganizationalUnit &&
			cert.GetCountry(subject) == CSRSubjectConsts.Country && cert.GetLocality(subject) == CSRSubjectConsts.Locality && cert.GetProvince(subject) == CSRSubjectConsts.Province
	}
}
