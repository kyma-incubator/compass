package graphql

func (f SpecFormat) isOneOf(formats []SpecFormat) bool {
	for _, value := range formats {
		if value == f {
			return true
		}
	}
	return false
}
