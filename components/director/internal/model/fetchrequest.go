package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/go-ozzo/ozzo-validation/is"

	validation "github.com/go-ozzo/ozzo-validation"
)

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

func (i *FetchRequestInput) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.URL, validation.Required, is.URL, validation.Length(1, 256), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.Auth),
		validation.Field(&i.Mode, validation.Required, validation.In(FetchModeSingle, FetchModePackage, FetchModeIndex), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.Filter, validation.NilOrNotEmpty, validation.Length(1, 256), validation.By(inputvalidation.ValidatePrintable)),
	)
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
