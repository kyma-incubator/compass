package gqlschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// UnmarshalJSON is used to handle unmarshalling ClusterConfig interface properly
func (a *RuntimeConfig) UnmarshalJSON(data []byte) error {
	err := a.unmarshalGardener(data)
	if err != nil {
		return a.unmarshalGCP(data)
	}

	return nil
}

func (a *RuntimeConfig) unmarshalGardener(data []byte) error {
	type Alias RuntimeConfig

	tempGardener := &struct {
		*Alias
		ClusterConfig *GardenerConfig `json:"clusterConfig"`
	}{
		Alias: (*Alias)(a),
	}

	decoder := newDecoder(data)
	if err := decoder.Decode(&tempGardener); err != nil {
		return err
	}

	a.ClusterConfig = tempGardener.ClusterConfig

	return nil
}

func (a *RuntimeConfig) unmarshalGCP(data []byte) error {
	type Alias RuntimeConfig

	tempGCP := &struct {
		*Alias
		ClusterConfig *GCPConfig `json:"clusterConfig"`
	}{
		Alias: (*Alias)(a),
	}

	decoder := newDecoder(data)
	if err := decoder.Decode(&tempGCP); err != nil {
		return err
	}

	a.ClusterConfig = tempGCP.ClusterConfig

	return nil
}

// UnmarshalJSON is used to handle unmarshaling ProviderSpecificConfig interface properly
func (g *GardenerConfig) UnmarshalJSON(data []byte) error {
	type Alias GardenerConfig

	temp := &struct {
		*Alias
		ProviderSpecificConfig json.RawMessage `json:"providerSpecificConfig"`
	}{
		Alias: (*Alias)(g),
	}
	decoder := newDecoder(data)
	if err := decoder.Decode(&temp); err != nil {
		return err
	}
	if temp.Provider == nil {
		return errors.New("provider field is required")
	}

	switch *temp.Provider {
	case "azure": // TODO to enum which will be validated
		g.ProviderSpecificConfig = &AzureProviderConfig{}
	case "gcp": // TODO to enum which will be validated
		g.ProviderSpecificConfig = &GCPProviderConfig{}
	case "aws": // TODO to enum which will be validated
		g.ProviderSpecificConfig = &AWSProviderConfig{}
	default:
		return fmt.Errorf("got unknown provider type %q", *temp.Provider)
	}

	if err := json.Unmarshal(temp.ProviderSpecificConfig, g.ProviderSpecificConfig); err != nil {
		return err
	}

	return nil
}

func newDecoder(data []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.DisallowUnknownFields()
	return decoder
}
