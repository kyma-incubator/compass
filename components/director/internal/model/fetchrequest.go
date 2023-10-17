package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// FetchRequest represents a request to fetch a specification resource from a remote system.
type FetchRequest struct {
	ID         string
	URL        string
	Auth       *Auth
	Mode       FetchMode
	Filter     *string
	Status     *FetchRequestStatus
	ObjectType FetchRequestReferenceObjectType
	ObjectID   string
}

// FetchRequestReferenceObjectType represents the type of the object that the fetch request is referencing.
type FetchRequestReferenceObjectType string

const (
	// APISpecFetchRequestReference represents a fetch request for an API Specification.
	APISpecFetchRequestReference FetchRequestReferenceObjectType = "APISpec"
	// EventSpecFetchRequestReference represents a fetch request for an Event Specification.
	EventSpecFetchRequestReference FetchRequestReferenceObjectType = "EventSpec"
	// CapabilitySpecFetchRequestReference represents a fetch request for a Capability Specification.
	CapabilitySpecFetchRequestReference FetchRequestReferenceObjectType = "CapabilitySpec"
	// DocumentFetchRequestReference represents a fetch request for an Document Specification.
	DocumentFetchRequestReference FetchRequestReferenceObjectType = "Document"
)

// GetResourceType returns the resource type of the fetch request based on the referenced entity.
func (obj FetchRequestReferenceObjectType) GetResourceType() resource.Type {
	switch obj {
	case APISpecFetchRequestReference:
		return resource.APISpecFetchRequest
	case EventSpecFetchRequestReference:
		return resource.EventSpecFetchRequest
	case CapabilitySpecFetchRequestReference:
		return resource.CapabilitySpecFetchRequest
	case DocumentFetchRequestReference:
		return resource.DocFetchRequest
	}
	return ""
}

// FetchRequestStatus is the status of an executed fetch request.
type FetchRequestStatus struct {
	Condition FetchRequestStatusCondition
	Message   *string
	Timestamp time.Time
}

// FetchMode is a legacy never delivered feature.
type FetchMode string

const (
	// FetchModeSingle is a legacy never delivered feature.
	FetchModeSingle FetchMode = "SINGLE"
	// FetchModeBundle is a legacy never delivered feature.
	FetchModeBundle FetchMode = "BUNDLE"
	// FetchModeIndex is a legacy never delivered feature.
	FetchModeIndex FetchMode = "INDEX"
)

// FetchRequestStatusCondition represents the condition of a fetch request.
type FetchRequestStatusCondition string

const (
	// FetchRequestStatusConditionInitial represents the initial state of a fetch request.
	FetchRequestStatusConditionInitial FetchRequestStatusCondition = "INITIAL"
	// FetchRequestStatusConditionSucceeded represents the state of a fetch request after it has been successfully executed.
	FetchRequestStatusConditionSucceeded FetchRequestStatusCondition = "SUCCEEDED"
	// FetchRequestStatusConditionFailed represents the state of a fetch request after it has failed.
	FetchRequestStatusConditionFailed FetchRequestStatusCondition = "FAILED"
)

// FetchRequestInput represents the input for creating a fetch request.
type FetchRequestInput struct {
	URL    string
	Auth   *AuthInput
	Mode   *FetchMode
	Filter *string
}

// ToFetchRequest converts a FetchRequestInput to a FetchRequest.
func (f *FetchRequestInput) ToFetchRequest(timestamp time.Time, id string, objectType FetchRequestReferenceObjectType, objectID string) *FetchRequest {
	if f == nil {
		return nil
	}

	fetchMode := FetchModeSingle
	if f.Mode != nil {
		fetchMode = *f.Mode
	}

	return &FetchRequest{
		ID:     id,
		URL:    f.URL,
		Auth:   f.Auth.ToAuth(),
		Mode:   fetchMode,
		Filter: f.Filter,
		Status: &FetchRequestStatus{
			Condition: FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: objectType,
		ObjectID:   objectID,
	}
}
