package model

import (
	"time"

	"github.com/pkg/errors"
)

type PackageInstanceAuth struct {
	ID          string
	PackageID   string
	Tenant      string
	Context     *string
	InputParams *string
	Auth        *Auth
	Status      *PackageInstanceAuthStatus
}

func (a *PackageInstanceAuth) SetDefaultStatus(condition PackageInstanceAuthStatusCondition, timestamp time.Time) error {
	if a == nil {
		return nil
	}

	var reason, message string

	switch condition {
	case PackageInstanceAuthStatusConditionSucceeded:
		reason = "CredentialsProvided"
		message = "Credentials were provided."
		break
	case PackageInstanceAuthStatusConditionPending:
		reason = "CredentialsNotProvided"
		message = "Credentials were not yet provided."
		break
	case PackageInstanceAuthStatusConditionUnused:
		reason = "PendingDeletion"
		message = "Credentials for given Package Instance Auth are ready for being deleted by Application or Integration System."
		break
	default:
		return errors.New("invalid status condition")
	}

	a.Status = &PackageInstanceAuthStatus{
		Condition: condition,
		Timestamp: timestamp,
		Message:   &message,
		Reason:    &reason,
	}

	return nil
}

type PackageInstanceAuthStatus struct {
	Condition PackageInstanceAuthStatusCondition
	Timestamp time.Time
	Message   *string
	Reason    *string
}

type PackageInstanceAuthStatusCondition string

const (
	PackageInstanceAuthStatusConditionPending   PackageInstanceAuthStatusCondition = "PENDING"
	PackageInstanceAuthStatusConditionSucceeded PackageInstanceAuthStatusCondition = "SUCCEEDED"
	PackageInstanceAuthStatusConditionFailed    PackageInstanceAuthStatusCondition = "FAILED"
	PackageInstanceAuthStatusConditionUnused    PackageInstanceAuthStatusCondition = "UNUSED"
)

// Input type for requestPackageInstanceAuthCreation
type PackageInstanceAuthRequestInput struct {
	Context     *string
	InputParams *string
}

func (ri PackageInstanceAuthRequestInput) ToPackageInstanceAuth(id, packageID, tenant string, auth *Auth, status PackageInstanceAuthStatus) PackageInstanceAuth {
	return PackageInstanceAuth{
		ID:          id,
		PackageID:   packageID,
		Tenant:      tenant,
		Context:     ri.Context,
		InputParams: ri.InputParams,
		Auth:        auth,
		Status:      &status,
	}
}

// Input type for setPackageInstanceAuth
type PackageInstanceAuthSetInput struct {
	Auth   *AuthInput
	Status *PackageInstanceAuthStatusInput
}

type PackageInstanceAuthStatusInput struct {
	Condition PackageInstanceAuthSetStatusConditionInput
	Message   *string
	Reason    *string
}

func (si *PackageInstanceAuthStatusInput) ToPackageInstanceAuthStatus(timestamp time.Time) *PackageInstanceAuthStatus {
	if si == nil {
		return nil
	}

	return &PackageInstanceAuthStatus{
		Condition: PackageInstanceAuthStatusCondition(si.Condition),
		Timestamp: timestamp,
		Message:   si.Message,
		Reason:    si.Reason,
	}
}

type PackageInstanceAuthSetStatusConditionInput string

const (
	PackageInstanceAuthSetStatusConditionInputSucceeded PackageInstanceAuthSetStatusConditionInput = "SUCCEEDED"
	PackageInstanceAuthSetStatusConditionInputFailed    PackageInstanceAuthSetStatusConditionInput = "FAILED"
)
