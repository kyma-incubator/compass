package cert

import (
	"github.com/google/uuid"
	"regexp"
	"strings"
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

// GetUnmatchedOrganizationalUnit returns the OU that is remaining after matching previously expected ones based on a given pattern
func GetUnmatchedOrganizationalUnit(organizationalUnitPattern string) func(string) string {
	return func(subject string) string {
		orgUnitRegex := regexp.MustCompile(organizationalUnitPattern)
		orgUnits := GetAllOrganizationalUnits(subject)

		unmatchedOrgUnit := ""
		matchedOrgUnits := 0
		for _, orgUnit := range orgUnits {
			if !orgUnitRegex.MatchString(orgUnit) {
				unmatchedOrgUnit = orgUnit
			} else {
				matchedOrgUnits++
			}
		}

		expectedOrgUnits := len(strings.Split(organizationalUnitPattern, "|"))
		if expectedOrgUnits-matchedOrgUnits != 1 { // there should be only 1 unmatched OU
			return ""
		}

		return unmatchedOrgUnit
	}
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
