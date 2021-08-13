package oathkeeper

import (
	"regexp"
	"strings"

	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
)

// GetOrganization returns the O part of the subject
func GetOrganization(subject string) string {
	return getRegexMatch("O=([^,]+)", subject)
}

// GetOrganizationalUnit returns the first OU of the subject
func GetOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^,]+)", subject)
}

// GetAllOrganizationalUnits returns all OU parts of the subject
func GetAllOrganizationalUnits(subject string) []string {
	return getAllRegexMatches("OU=([^,]+)", subject)
}

// GetCountry returns the C part of the subject
func GetCountry(subject string) string {
	return getRegexMatch("C=([^,]+)", subject)
}

// GetProvince returns the ST part of the subject
func GetProvince(subject string) string {
	return getRegexMatch("ST=([^,]+)", subject)
}

// GetLocality returns the L part of the subject
func GetLocality(subject string) string {
	return getRegexMatch("L=([^,]+)", subject)
}

// GetCommonName returns the CN part of the subject
func GetCommonName(subject string) string {
	return getRegexMatch("CN=([^,]+)", subject)
}

// ExternalCertIssuerSubjectMatcher returns a function matching certificate subjects issued by the external trusted issuer configured
// It checks Country, Organization as single values and OrganizationalUnit as regex pattern for easier matching of multiple values (joined by ',').
func ExternalCertIssuerSubjectMatcher(externalSubjectConsts certificates.SubjectConsts) func(subject string) bool {
	return func(subject string) bool {
		if GetCountry(subject) != externalSubjectConsts.Country || GetOrganization(subject) != externalSubjectConsts.Organization {
			return false
		}
		orgUnitRegex := regexp.MustCompile(externalSubjectConsts.OrganizationalUnit)
		ou := strings.Join(GetAllOrganizationalUnits(subject), ",")
		return orgUnitRegex.MatchString(ou)
	}
}

// ConnectorCertificateSubjectMatcher returns a function matching certificate subjects issued by compass's connector
func ConnectorCertificateSubjectMatcher(CSRSubjectConsts certificates.SubjectConsts) func(subject string) bool {
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
