package cluster_config

func updateString(toUpdate *string, value *string) {
	if value != nil {
		*toUpdate = *value
	}
}
