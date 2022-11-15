package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// Spec represents a specification of a resource.
type Spec struct {
	ID         string
	ObjectType SpecReferenceObjectType
	ObjectID   string

	Data       *string
	DataHash   *string
	Format     SpecFormat
	APIType    *APISpecType
	EventType  *EventSpecType
	CustomType *string
}

// SpecReferenceObjectType represents the type of the referenced object by the Spec.
type SpecReferenceObjectType string

const (
	// APISpecReference is a reference to an API Specification.
	APISpecReference SpecReferenceObjectType = "API"
	// EventSpecReference is a reference to an Event Specification.
	EventSpecReference SpecReferenceObjectType = "Event"
)

// GetResourceType returns the resource type of the specification based on the referenced entity.
func (obj SpecReferenceObjectType) GetResourceType() resource.Type {
	switch obj {
	case APISpecReference:
		return resource.APISpecification
	case EventSpecReference:
		return resource.EventSpecification
	}
	return ""
}

// SpecFormat is the format of the specification.
type SpecFormat string

const (
	// SpecFormatYaml is the YAML format.
	SpecFormatYaml SpecFormat = "YAML"
	// SpecFormatJSON is the JSON format.
	SpecFormatJSON SpecFormat = "JSON"
	// SpecFormatXML is the XML format.
	SpecFormatXML SpecFormat = "XML"

	// ORD Formats

	// SpecFormatApplicationJSON is the Application JSON format.
	SpecFormatApplicationJSON SpecFormat = "application/json"
	// SpecFormatTextYAML is the Text YAML format.
	SpecFormatTextYAML SpecFormat = "text/yaml"
	// SpecFormatApplicationXML is the Application XML format.
	SpecFormatApplicationXML SpecFormat = "application/xml"
	// SpecFormatPlainText is the Plain Text format.
	SpecFormatPlainText SpecFormat = "text/plain"
	// SpecFormatOctetStream is the Octet Stream format.
	SpecFormatOctetStream SpecFormat = "application/octet-stream"
)

// APISpecType is the type of the API Specification.
type APISpecType string

const (
	// APISpecTypeOdata is the OData Specification.
	APISpecTypeOdata APISpecType = "ODATA"
	// APISpecTypeOpenAPI is the OpenAPI Specification.
	APISpecTypeOpenAPI APISpecType = "OPEN_API"

	// ORD Formats

	// APISpecTypeOpenAPIV2 is the OpenAPI V2 Specification.
	APISpecTypeOpenAPIV2 APISpecType = "openapi-v2"
	// APISpecTypeOpenAPIV3 is the OpenAPI V3 Specification.
	APISpecTypeOpenAPIV3 APISpecType = "openapi-v3"
	// APISpecTypeRaml is the RAML Specification.
	APISpecTypeRaml APISpecType = "raml-v1"
	// APISpecTypeEDMX is the EDMX Specification.
	APISpecTypeEDMX APISpecType = "edmx"
	// APISpecTypeCsdl is the CSDL Specification.
	APISpecTypeCsdl APISpecType = "csdl-json"
	// APISpecTypeWsdlV1 is the WSDL V1 Specification.
	APISpecTypeWsdlV1 APISpecType = "wsdl-v1"
	// APISpecTypeWsdlV2 is the WSDL V2 Specification.
	APISpecTypeWsdlV2 APISpecType = "wsdl-v2"
	// APISpecTypeRfcMetadata is the RFC Metadata Specification.
	APISpecTypeRfcMetadata APISpecType = "sap-rfc-metadata-v1"
	// APISpecTypeCustom is the Custom Specification.
	APISpecTypeCustom APISpecType = "custom"
)

// EventSpecType is the type of the Event Specification.
type EventSpecType string

const (
	// EventSpecTypeAsyncAPI is the AsyncAPI Specification.
	EventSpecTypeAsyncAPI EventSpecType = "ASYNC_API"

	// ORD Formats

	// EventSpecTypeAsyncAPIV2 is the AsyncAPI V2 Specification.
	EventSpecTypeAsyncAPIV2 EventSpecType = "asyncapi-v2"
	// EventSpecTypeCustom is the Custom Specification.
	EventSpecTypeCustom EventSpecType = "custom"
)

// SpecInput is an input for creating/updating specification.
type SpecInput struct {
	Data       *string
	DataHash   *string
	Format     SpecFormat
	APIType    *APISpecType
	EventType  *EventSpecType
	CustomType *string

	FetchRequest *FetchRequestInput
}

// ToSpec converts the input to a Spec.
func (s *SpecInput) ToSpec(id string, objectType SpecReferenceObjectType, objectID string) (*Spec, error) {
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
		ObjectType: objectType,
		ObjectID:   objectID,
		Data:       s.Data,
		DataHash:   s.DataHash,
		Format:     s.Format,
		APIType:    s.APIType,
		EventType:  s.EventType,
		CustomType: s.CustomType,
	}, nil
}
