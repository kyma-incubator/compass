package model

import "time"

//  Compass performs fetch to validate if request is correct and stores a copy
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

type FetchRequestReferenceObjectType string

const (
	APIFetchRequestReference      FetchRequestReferenceObjectType = "API"
	EventAPIFetchRequestReference FetchRequestReferenceObjectType = "EventAPI"
	DocumentFetchRequestReference FetchRequestReferenceObjectType = "Document"
)

type FetchRequestStatus struct {
	Condition FetchRequestStatusCondition
	Timestamp time.Time
}

type FetchMode string

const (
	FetchModeSingle  FetchMode = "SINGLE"
	FetchModePackage FetchMode = "PACKAGE"
	FetchModeIndex   FetchMode = "INDEX"
)

type FetchRequestStatusCondition string

const (
	FetchRequestStatusConditionInitial   FetchRequestStatusCondition = "INITIAL"
	FetchRequestStatusConditionSucceeded FetchRequestStatusCondition = "SUCCEEDED"
	FetchRequestStatusConditionFailed    FetchRequestStatusCondition = "FAILED"
)

type FetchRequestInput struct {
	URL    string
	Auth   *AuthInput
	Mode   *FetchMode
	Filter *string
}

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
