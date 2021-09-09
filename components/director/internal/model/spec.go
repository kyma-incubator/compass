package model

import (
	"github.com/pkg/errors"
)

// Spec missing godoc
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

// SpecReferenceObjectType missing godoc
type SpecReferenceObjectType string

const (
	// APISpecReference missing godoc
	APISpecReference SpecReferenceObjectType = "API"
	// EventSpecReference missing godoc
	EventSpecReference SpecReferenceObjectType = "Event"
)

// SpecFormat missing godoc
type SpecFormat string

const (
	// SpecFormatYaml missing godoc
	SpecFormatYaml SpecFormat = "YAML"
	// SpecFormatJSON missing godoc
	SpecFormatJSON SpecFormat = "JSON"
	// SpecFormatXML missing godoc
	SpecFormatXML SpecFormat = "XML"

	// ORD Formats

	// SpecFormatApplicationJSON missing godoc
	SpecFormatApplicationJSON SpecFormat = "application/json"
	// SpecFormatTextYAML missing godoc
	SpecFormatTextYAML SpecFormat = "text/yaml"
	// SpecFormatApplicationXML missing godoc
	SpecFormatApplicationXML SpecFormat = "application/xml"
	// SpecFormatPlainText missing godoc
	SpecFormatPlainText SpecFormat = "text/plain"
	// SpecFormatOctetStream missing godoc
	SpecFormatOctetStream SpecFormat = "application/octet-stream"
)

// APISpecType missing godoc
type APISpecType string

const (
	// APISpecTypeOdata missing godoc
	APISpecTypeOdata APISpecType = "ODATA"
	// APISpecTypeOpenAPI missing godoc
	APISpecTypeOpenAPI APISpecType = "OPEN_API"

	// ORD Formats

	// APISpecTypeOpenAPIV2 missing godoc
	APISpecTypeOpenAPIV2 APISpecType = "openapi-v2"
	// APISpecTypeOpenAPIV3 missing godoc
	APISpecTypeOpenAPIV3 APISpecType = "openapi-v3"
	// APISpecTypeRaml missing godoc
	APISpecTypeRaml APISpecType = "raml-v1"
	// APISpecTypeEDMX missing godoc
	APISpecTypeEDMX APISpecType = "edmx"
	// APISpecTypeCsdl missing godoc
	APISpecTypeCsdl APISpecType = "csdl-json"
	// APISpecTypeWsdlV1 missing godoc
	APISpecTypeWsdlV1 APISpecType = "wsdl-v1"
	// APISpecTypeWsdlV2 missing godoc
	APISpecTypeWsdlV2 APISpecType = "wsdl-v2"
	// APISpecTypeRfcMetadata missing godoc
	APISpecTypeRfcMetadata APISpecType = "sap-rfc-metadata-v1"
	// APISpecTypeCustom missing godoc
	APISpecTypeCustom APISpecType = "custom"
)

// EventSpecType missing godoc
type EventSpecType string

const (
	// EventSpecTypeAsyncAPI missing godoc
	EventSpecTypeAsyncAPI EventSpecType = "ASYNC_API"

	// ORD Formats

	// EventSpecTypeAsyncAPIV2 missing godoc
	EventSpecTypeAsyncAPIV2 EventSpecType = "asyncapi-v2"
	// EventSpecTypeCustom missing godoc
	EventSpecTypeCustom EventSpecType = "custom"
)

// SpecInput missing godoc
type SpecInput struct {
	Data       *string
	Format     SpecFormat
	APIType    *APISpecType
	EventType  *EventSpecType
	CustomType *string

	FetchRequest *FetchRequestInput
}

// ToSpec missing godoc
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
