package oathkeeper

import "regexp"

func GetOrganization(subject string) string {
	return getRegexMatch("O=([^,]+)", subject)
}

func GetOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^,]+)", subject)
}

func GetAllOrganizationalUnits(subject string) []string{
	return getAllRegexMatches("OU=([^,]+)", subject)
}

func GetCountry(subject string) string {
	return getRegexMatch("C=([^,]+)", subject)
}

func GetProvince(subject string) string {
	return getRegexMatch("ST=([^,]+)", subject)
}

func GetLocality(subject string) string {
	return getRegexMatch("L=([^,]+)", subject)
}

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
