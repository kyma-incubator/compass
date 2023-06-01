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
	Owner            string
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

// SetFromUpdateInput sets fields to BundleInstanceAuth from BundleInstanceAuthUpdateInput
func (a *BundleInstanceAuth) SetFromUpdateInput(in BundleInstanceAuthUpdateInput) {
	if in.Context != nil {
		a.Context = in.Context
	}

	if in.InputParams != nil {
		a.InputParams = in.InputParams
	}

	if in.Auth != nil {
		a.Auth = in.Auth.ToAuth()
	}
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
		Owner:            tenant,
		Context:          ri.Context,
		InputParams:      ri.InputParams,
		Auth:             auth,
		Status:           status,
	}
}

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

// BundleInstanceAuthCreateInput type is input for createBundleInstanceAuth mutation
type BundleInstanceAuthCreateInput struct {
	Context          *string
	InputParams      *string
	Auth             *AuthInput
	RuntimeID        *string
	RuntimeContextID *string
}

// ToBundleInstanceAuth creates BundleInstanceAuth from BundleInstanceAuthCreateInput
func (ri BundleInstanceAuthCreateInput) ToBundleInstanceAuth(id, bundleID, tenant string, status *BundleInstanceAuthStatus) BundleInstanceAuth {
	return BundleInstanceAuth{
		ID:               id,
		BundleID:         bundleID,
		RuntimeID:        ri.RuntimeID,
		RuntimeContextID: ri.RuntimeContextID,
		Owner:            tenant,
		Context:          ri.Context,
		InputParams:      ri.InputParams,
		Auth:             ri.Auth.ToAuth(),
		Status:           status,
	}
}

// BundleInstanceAuthUpdateInput type is input for updateBundleInstanceAuth mutation
type BundleInstanceAuthUpdateInput struct {
	Context     *string
	InputParams *string
	Auth        *AuthInput
}
