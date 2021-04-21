package model

import (
	"time"

	"github.com/pkg/errors"
)

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

func (a *BundleInstanceAuth) SetDefaultStatus(condition BundleInstanceAuthStatusCondition, timestamp time.Time) error {
	if a == nil {
		return nil
	}

	var reason, message string

	switch condition {
	case BundleInstanceAuthStatusConditionSucceeded:
		reason = "CredentialsProvided"
		message = "Credentials were provided."
		break
	case BundleInstanceAuthStatusConditionPending:
		reason = "CredentialsNotProvided"
		message = "Credentials were not yet provided."
		break
	case BundleInstanceAuthStatusConditionUnused:
		reason = "PendingDeletion"
		message = "Credentials for given Bundle Instance Auth are ready for being deleted by Application or Integration System."
		break
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

type BundleInstanceAuthStatus struct {
	Condition BundleInstanceAuthStatusCondition
	Timestamp time.Time
	Message   string
	Reason    string
}

type BundleInstanceAuthStatusCondition string

const (
	BundleInstanceAuthStatusConditionPending   BundleInstanceAuthStatusCondition = "PENDING"
	BundleInstanceAuthStatusConditionSucceeded BundleInstanceAuthStatusCondition = "SUCCEEDED"
	BundleInstanceAuthStatusConditionFailed    BundleInstanceAuthStatusCondition = "FAILED"
	BundleInstanceAuthStatusConditionUnused    BundleInstanceAuthStatusCondition = "UNUSED"
)

// Input type for requestBundleInstanceAuthCreation
type BundleInstanceAuthRequestInput struct {
	ID          *string
	Context     *string
	InputParams *string
}

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

// Input type for setBundleInstanceAuth
type BundleInstanceAuthSetInput struct {
	Auth   *AuthInput
	Status *BundleInstanceAuthStatusInput
}

type BundleInstanceAuthStatusInput struct {
	Condition BundleInstanceAuthSetStatusConditionInput
	Message   string
	Reason    string
}

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

type BundleInstanceAuthSetStatusConditionInput string

const (
	BundleInstanceAuthSetStatusConditionInputSucceeded BundleInstanceAuthSetStatusConditionInput = "SUCCEEDED"
	BundleInstanceAuthSetStatusConditionInputFailed    BundleInstanceAuthSetStatusConditionInput = "FAILED"
)
