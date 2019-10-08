package gqlschema

import "encoding/json"

type clusterConfig struct {
	*GardenerConfig
	*GCPConfig
}

func (cfg *RuntimeConfig) UnmarshalJSON(data []byte) error {
	type Alias RuntimeConfig

	aux := &struct {
		*Alias
		ClusterConfig clusterConfig `json:"clusterConfig"`
	}{
		Alias: (*Alias)(cfg),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	cfg.ClusterConfig = retrieveConfig(aux.ClusterConfig)

	return nil
}

func retrieveConfig(umarshaledConfig clusterConfig) ClusterConfig {
	if umarshaledConfig.GardenerConfig != nil {
		return umarshaledConfig.GardenerConfig
	}
	return umarshaledConfig.GCPConfig
}
