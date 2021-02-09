package model

import (
	"github.com/pkg/errors"
)

type Spec struct {
	ID         string
	Tenant     string
	ObjectType SpecReferenceObjectType
	ObjectID   string

	Data       *string
	Format     SpecFormat
	APIType    *APISpecType
	EventType  *EventSpecType
	CustomType *string
}

type SpecReferenceObjectType string

const (
	APISpecReference   SpecReferenceObjectType = "API"
	EventSpecReference SpecReferenceObjectType = "Event"
)

type SpecFormat string

const (
	SpecFormatYaml SpecFormat = "YAML"
	SpecFormatJSON SpecFormat = "JSON"
	SpecFormatXML  SpecFormat = "XML"

	// ORD Formats
	SpecFormatApplicationJSON SpecFormat = "application/json"
	SpecFormatTextYAML        SpecFormat = "text/yaml"
	SpecFormatApplicationXML  SpecFormat = "application/xml"
	SpecFormatPlainText       SpecFormat = "text/plain"
	SpecFormatOctetStream     SpecFormat = "application/octet-stream"
)

type APISpecType string

const (
	APISpecTypeOdata   APISpecType = "ODATA"
	APISpecTypeOpenAPI APISpecType = "OPEN_API"

	// ORD Formats
	APISpecTypeOpenAPIV2   APISpecType = "openapi-v2"
	APISpecTypeOpenAPIV3   APISpecType = "openapi-v3"
	APISpecTypeRaml        APISpecType = "raml-v1"
	APISpecTypeEDMX        APISpecType = "edmx"
	APISpecTypeCsdl        APISpecType = "csdl-json"
	APISpecTypeWsdlV1      APISpecType = "wsdl-v1"
	APISpecTypeWsdlV2      APISpecType = "wsdl-v2"
	APISpecTypeRfcMetadata APISpecType = "sap-rfc-metadata-v1"
	APISpecTypeCustom      APISpecType = "custom"
)

type EventSpecType string

const (
	EventSpecTypeAsyncAPI EventSpecType = "ASYNC_API"

	// ORD Formats
	EventSpecTypeAsyncAPIV2 EventSpecType = "asyncapi-v2"
	EventSpecTypeCustom     EventSpecType = "custom"
)

type SpecInput struct {
	Data       *string
	Format     SpecFormat
	APIType    *APISpecType
	EventType  *EventSpecType
	CustomType *string

	FetchRequest *FetchRequestInput
}

func (s *SpecInput) ToSpec(id, tenant string, objectType SpecReferenceObjectType, objectID string) (*Spec, error) {
	if s == nil {
		return nil, nil
	}

	if objectType == APISpecReference && s.APIType == nil {
		return nil, errors.New("API Spec type cannot be empty")
	}

	if objectType == EventSpecReference && s.EventType == nil {
		return nil, errors.New("event spec type cannot be empty")
	}

	return &Spec{
		ID:         id,
		Tenant:     tenant,
		ObjectType: objectType,
		ObjectID:   objectID,
		Data:       s.Data,
		Format:     s.Format,
		APIType:    s.APIType,
		EventType:  s.EventType,
		CustomType: s.CustomType,
	}, nil
}
