package model

import (
	"time"

	"github.com/pkg/errors"
)

// BundleInstanceAuth missing godoc
type BundleInstanceAuth struct {
	ID               string
	BundleID         string
	RuntimeID        *string
	RuntimeContextID *string
	Tenant           string
	Context          *string
	InputParams      *string
	Auth             *Auth
	Status           *BundleInstanceAuthStatus
}

// SetDefaultStatus missing godoc
func (a *BundleInstanceAuth) SetDefaultStatus(condition BundleInstanceAuthStatusCondition, timestamp time.Time) error {
	if a == nil {
		return nil
	}

	var reason, message string

	switch condition {
	case BundleInstanceAuthStatusConditionSucceeded:
		reason = "CredentialsProvided"
		message = "Credentials were provided."
	case BundleInstanceAuthStatusConditionPending:
		reason = "CredentialsNotProvided"
		message = "Credentials were not yet provided."
	case BundleInstanceAuthStatusConditionUnused:
		reason = "PendingDeletion"
		message = "Credentials for given Bundle Instance Auth are ready for being deleted by Application or Integration System."
	default:
		return errors.Errorf("invalid status condition: %s", condition)
	}

	a.Status = &BundleInstanceAuthStatus{
		Condition: condition,
		Timestamp: timestamp,
		Message:   message,
		Reason:    reason,
	}

	return nil
}

// BundleInstanceAuthStatus missing godoc
type BundleInstanceAuthStatus struct {
	Condition BundleInstanceAuthStatusCondition
	Timestamp time.Time
	Message   string
	Reason    string
}

// BundleInstanceAuthStatusCondition missing godoc
type BundleInstanceAuthStatusCondition string

const (
	// BundleInstanceAuthStatusConditionPending missing godoc
	BundleInstanceAuthStatusConditionPending BundleInstanceAuthStatusCondition = "PENDING"
	// BundleInstanceAuthStatusConditionSucceeded missing godoc
	BundleInstanceAuthStatusConditionSucceeded BundleInstanceAuthStatusCondition = "SUCCEEDED"
	// BundleInstanceAuthStatusConditionFailed missing godoc
	BundleInstanceAuthStatusConditionFailed BundleInstanceAuthStatusCondition = "FAILED"
	// BundleInstanceAuthStatusConditionUnused missing godoc
	BundleInstanceAuthStatusConditionUnused BundleInstanceAuthStatusCondition = "UNUSED"
)

// BundleInstanceAuthRequestInput missing godoc
// BundleInstanceAuthRequestInput type for requestBundleInstanceAuthCreation
type BundleInstanceAuthRequestInput struct {
	ID          *string
	Context     *string
	InputParams *string
}

// ToBundleInstanceAuth missing godoc
func (ri BundleInstanceAuthRequestInput) ToBundleInstanceAuth(id, bundleID, tenant string, auth *Auth, status *BundleInstanceAuthStatus, runtimeID *string, runtimeContextID *string) BundleInstanceAuth {
	return BundleInstanceAuth{
		ID:               id,
		BundleID:         bundleID,
		RuntimeID:        runtimeID,
		RuntimeContextID: runtimeContextID,
		Tenant:           tenant,
		Context:          ri.Context,
		InputParams:      ri.InputParams,
		Auth:             auth,
		Status:           status,
	}
}

// BundleInstanceAuthSetInput missing godoc
// BundleInstanceAuthSetInput type for setBundleInstanceAuth
type BundleInstanceAuthSetInput struct {
	Auth   *AuthInput
	Status *BundleInstanceAuthStatusInput
}

// BundleInstanceAuthStatusInput missing godoc
type BundleInstanceAuthStatusInput struct {
	Condition BundleInstanceAuthSetStatusConditionInput
	Message   string
	Reason    string
}

// ToBundleInstanceAuthStatus missing godoc
func (si *BundleInstanceAuthStatusInput) ToBundleInstanceAuthStatus(timestamp time.Time) *BundleInstanceAuthStatus {
	if si == nil {
		return nil
	}

	return &BundleInstanceAuthStatus{
		Condition: BundleInstanceAuthStatusCondition(si.Condition),
		Timestamp: timestamp,
		Message:   si.Message,
		Reason:    si.Reason,
	}
}

// BundleInstanceAuthSetStatusConditionInput missing godoc
type BundleInstanceAuthSetStatusConditionInput string

const (
	// BundleInstanceAuthSetStatusConditionInputSucceeded missing godoc
	BundleInstanceAuthSetStatusConditionInputSucceeded BundleInstanceAuthSetStatusConditionInput = "SUCCEEDED"
	// BundleInstanceAuthSetStatusConditionInputFailed missing godoc
	BundleInstanceAuthSetStatusConditionInputFailed BundleInstanceAuthSetStatusConditionInput = "FAILED"
)
