package gqlschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

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
