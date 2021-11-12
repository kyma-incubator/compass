package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// FetchRequest compass performs fetch to validate if request is correct and stores a copy
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

// FetchRequestReferenceObjectType missing godoc
type FetchRequestReferenceObjectType string

const (
	// APISpecFetchRequestReference missing godoc
	APISpecFetchRequestReference FetchRequestReferenceObjectType = "APISpec"
	// EventSpecFetchRequestReference missing godoc
	EventSpecFetchRequestReference FetchRequestReferenceObjectType = "EventSpec"
	// DocumentFetchRequestReference missing godoc
	DocumentFetchRequestReference FetchRequestReferenceObjectType = "Document"
)

func (obj FetchRequestReferenceObjectType) GetResourceType() resource.Type {
	switch obj {
	case APISpecFetchRequestReference:
		return resource.APISpecFetchRequest
	case EventSpecFetchRequestReference:
		return resource.EventSpecFetchRequest
	case DocumentFetchRequestReference:
		return resource.DocFetchRequest
	}
	return ""
}

// FetchRequestStatus missing godoc
type FetchRequestStatus struct {
	Condition FetchRequestStatusCondition
	Message   *string
	Timestamp time.Time
}

// FetchMode missing godoc
type FetchMode string

const (
	// FetchModeSingle missing godoc
	FetchModeSingle FetchMode = "SINGLE"
	// FetchModeBundle missing godoc
	FetchModeBundle FetchMode = "BUNDLE"
	// FetchModeIndex missing godoc
	FetchModeIndex FetchMode = "INDEX"
)

// FetchRequestStatusCondition missing godoc
type FetchRequestStatusCondition string

const (
	// FetchRequestStatusConditionInitial missing godoc
	FetchRequestStatusConditionInitial FetchRequestStatusCondition = "INITIAL"
	// FetchRequestStatusConditionSucceeded missing godoc
	FetchRequestStatusConditionSucceeded FetchRequestStatusCondition = "SUCCEEDED"
	// FetchRequestStatusConditionFailed missing godoc
	FetchRequestStatusConditionFailed FetchRequestStatusCondition = "FAILED"
)

// FetchRequestInput missing godoc
type FetchRequestInput struct {
	URL    string
	Auth   *AuthInput
	Mode   *FetchMode
	Filter *string
}

// ToFetchRequest missing godoc
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
