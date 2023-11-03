package formationassignment

func IsConfigEmpty(configuration string) bool {
	if configuration == "" || configuration == "{}" || configuration == "\"\"" || configuration == "[]" || configuration == "null" {
		return true
	}

	return false
}
