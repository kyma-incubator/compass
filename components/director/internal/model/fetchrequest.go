package model

import (
	"time"
)

// FetchRequest missing godoc
// FetchRequest compass performs fetch to validate if request is correct and stores a copy
type FetchRequest struct {
	ID         string
	Tenant     string
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
	// SpecFetchRequestReference missing godoc
	SpecFetchRequestReference FetchRequestReferenceObjectType = "Spec"
	// DocumentFetchRequestReference missing godoc
	DocumentFetchRequestReference FetchRequestReferenceObjectType = "Document"
)

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
func (f *FetchRequestInput) ToFetchRequest(timestamp time.Time, id, tenant string, objectType FetchRequestReferenceObjectType, objectID string) *FetchRequest {
	if f == nil {
		return nil
	}

	fetchMode := FetchModeSingle
	if f.Mode != nil {
		fetchMode = *f.Mode
	}

	return &FetchRequest{
		ID:     id,
		Tenant: tenant,
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
