package formationassignment

// IsConfigEmpty checks for different "empty" json values that could be in the formation assignment configuration
func IsConfigEmpty(configuration string) bool {
	if configuration == "" || configuration == "{}" || configuration == "\"\"" || configuration == "[]" || configuration == "null" {
		return true
	}

	return false
}
