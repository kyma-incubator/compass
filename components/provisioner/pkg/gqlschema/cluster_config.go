package gqlschema

import (
	"bytes"
	"encoding/json"
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

// UnmarshalJSON is used to handle unmarshalling ProviderSpecificConfig interface properly
func (g *GardenerConfig) UnmarshalJSON(data []byte) error {
	err := g.unmarshalAzure(data)
	if err != nil {
		err := g.unmarshalGCP(data)
		if err != nil {
			return g.unmarshalAWS(data)
		}
		return nil
	}

	return nil
}

func (g *GardenerConfig) unmarshalAzure(data []byte) error {
	type Alias GardenerConfig

	temp := &struct {
		*Alias
		ProviderSpecificConfig *AzureProviderConfig `json:"providerSpecificConfig"`
	}{
		Alias: (*Alias)(g),
	}

	decoder := newDecoder(data)
	if err := decoder.Decode(&temp); err != nil {
		return err
	}

	g.ProviderSpecificConfig = temp.ProviderSpecificConfig

	return nil
}

func (g *GardenerConfig) unmarshalGCP(data []byte) error {
	type Alias GardenerConfig

	temp := &struct {
		*Alias
		ProviderSpecificConfig *GCPProviderConfig `json:"providerSpecificConfig"`
	}{
		Alias: (*Alias)(g),
	}

	decoder := newDecoder(data)
	if err := decoder.Decode(&temp); err != nil {
		return err
	}

	g.ProviderSpecificConfig = temp.ProviderSpecificConfig

	return nil
}

func (g *GardenerConfig) unmarshalAWS(data []byte) error {
	type Alias GardenerConfig

	temp := &struct {
		*Alias
		ProviderSpecificConfig *AWSProviderConfig `json:"providerSpecificConfig"`
	}{
		Alias: (*Alias)(g),
	}

	decoder := newDecoder(data)
	if err := decoder.Decode(&temp); err != nil {
		return err
	}

	g.ProviderSpecificConfig = temp.ProviderSpecificConfig

	return nil
}

func newDecoder(data []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.DisallowUnknownFields()
	return decoder
}
