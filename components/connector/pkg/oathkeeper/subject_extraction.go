package oathkeeper

import (
	"regexp"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
)

// GetOrganization returns the O part of the subject
func GetOrganization(subject string) string {
	return getRegexMatch("O=([^(,|+)]+)", subject)
}

// GetOrganizationalUnit returns the first OU of the subject
func GetOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^(,|+)]+)", subject)
}

// GetUUIDOrganizationalUnit returns the OU that is a valid UUID or empty string if there is no OU that is a valid UUID
func GetUUIDOrganizationalUnit(subject string) string {
	orgUnits := GetAllOrganizationalUnits(subject)
	for _, orgUnit := range orgUnits {
		if _, err := uuid.Parse(orgUnit); err == nil {
			return orgUnit
		}
	}
	return ""
}

// GetAllOrganizationalUnits returns all OU parts of the subject
func GetAllOrganizationalUnits(subject string) []string {
	return getAllRegexMatches("OU=([^(,|+)]+)", subject)
}

// GetCountry returns the C part of the subject
func GetCountry(subject string) string {
	return getRegexMatch("C=([^(,|+)]+)", subject)
}

// GetProvince returns the ST part of the subject
func GetProvince(subject string) string {
	return getRegexMatch("ST=([^(,|+)]+)", subject)
}

// GetLocality returns the L part of the subject
func GetLocality(subject string) string {
	return getRegexMatch("L=([^(,|+)]+)", subject)
}

// GetCommonName returns the CN part of the subject
func GetCommonName(subject string) string {
	return getRegexMatch("CN=([^,]+)", subject)
}

// ExternalCertIssuerSubjectMatcher returns a function matching certificate subjects issued by the external trusted issuer configured
// It checks Country, Organization as single values and OrganizationalUnit as regex pattern for easier matching of multiple values.
func ExternalCertIssuerSubjectMatcher(externalSubjectConsts certificates.ExternalIssuerSubjectConsts) func(subject string) bool {
	return func(subject string) bool {
		if GetCountry(subject) != externalSubjectConsts.Country || GetOrganization(subject) != externalSubjectConsts.Organization {
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
		return GetOrganization(subject) == CSRSubjectConsts.Organization && GetOrganizationalUnit(subject) == CSRSubjectConsts.OrganizationalUnit &&
			GetCountry(subject) == CSRSubjectConsts.Country && GetLocality(subject) == CSRSubjectConsts.Locality && GetProvince(subject) == CSRSubjectConsts.Province
	}
}

func getRegexMatch(regex, text string) string {
	matches := getAllRegexMatches(regex, text)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func getAllRegexMatches(regex, text string) []string {
	cnRegex := regexp.MustCompile(regex)
	matches := cnRegex.FindAllStringSubmatch(text, -1)

	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		result = append(result, match[1])
	}

	return result
}
