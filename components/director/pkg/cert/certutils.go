package cert

import (
	"regexp"

	"github.com/google/uuid"
)

const (
	ConsumerTypeExtraField = "consumer_type"
	AccessLevelExtraField  = "access_level"
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

// GetRemainingOrganizationalUnit returns the OU that is remaining after matching previously expected ones based on a given pattern
func GetRemainingOrganizationalUnit(organizationalUnitPattern string) func(string) string {
	return func(subject string) string {
		orgUnitRegex := regexp.MustCompile(organizationalUnitPattern)
		orgUnits := GetAllOrganizationalUnits(subject)

		remainingOrgUnit := ""
		matchedOrgUnits := 0
		for _, orgUnit := range orgUnits {
			if !orgUnitRegex.MatchString(orgUnit) {
				remainingOrgUnit = orgUnit
			} else {
				matchedOrgUnits++
			}
		}

		expectedOrgUnits := GetPossibleRegexTopLevelMatches(organizationalUnitPattern)
		singleRemainingOrgUnitExists := len(orgUnits)-expectedOrgUnits == 1 || expectedOrgUnits-matchedOrgUnits == 0
		if !singleRemainingOrgUnitExists {
			return ""
		}

		return remainingOrgUnit
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

func GetExtra(consumerType, accessLevel string) (map[string]interface{}, error) {
	//ct := model.SystemAuthReferenceObjectType(consumerType)
	//if ct != .ApplicationReference && ct != model.RuntimeReference && ct != model.IntegrationSystemReference {
	//	return nil, fmt.Errorf("invalid consumer type %q", consumerType)
	//}
	//te := tenantEntity.Type(accessLevel)
	//if te != tenantEntity.Account && te != tenantEntity.Subaccount && te != tenantEntity.Customer {
	//	return nil, fmt.Errorf("invalid access level %q", accessLevel)
	//}

	return map[string]interface{}{
		ConsumerTypeExtraField: consumerType,
		AccessLevelExtraField:  accessLevel,
	}, nil
}

// GetPossibleRegexTopLevelMatches returns the number of possible top level matches of a regex pattern.
// This means that the pattern will be inspected and split only based on the most top level '|' delimiter
// and inner group '|' delimiters won't be taken into account.
func GetPossibleRegexTopLevelMatches(pattern string) int {
	if pattern == "" {
		return 0
	}
	count := 1
	openedGroups := 0
	for _, char := range pattern {
		switch char {
		case '|':
			if openedGroups == 0 {
				count++
			}
		case '(':
			openedGroups++
		case ')':
			openedGroups--
		default:
			continue
		}
	}
	return count
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
