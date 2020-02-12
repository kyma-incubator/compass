package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
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

func (a *PackageInstanceAuth) SetAuth(setInput PackageInstanceAuthSetInput, timestamp time.Time) {
	if a == nil {
		return
	}

	a.Auth = setInput.Auth.ToAuth()
	a.Status = setInput.Status.ToPackageInstanceAuthStatus(timestamp)
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

func (ri PackageInstanceAuthRequestInput) ToPackageInstanceAuth(id, packageID, tenant string, auth *Auth, timestamp time.Time) PackageInstanceAuth {
	return PackageInstanceAuth{
		ID:          id,
		PackageID:   packageID,
		Tenant:      tenant,
		Context:     ri.Context,
		InputParams: ri.InputParams,
		Auth:        auth,
		Status:      ri.statusFromAuth(auth, timestamp),
	}
}

func (ri PackageInstanceAuthRequestInput) statusFromAuth(auth *Auth, timestamp time.Time) *PackageInstanceAuthStatus {
	if auth == nil {
		return &PackageInstanceAuthStatus{
			Condition: PackageInstanceAuthStatusConditionPending,
			Timestamp: timestamp,
			Message:   str.Ptr("Credentials were not yet provided."),
			Reason:    str.Ptr("CredentialsNotProvided"),
		}
	}

	return &PackageInstanceAuthStatus{
		Condition: PackageInstanceAuthStatusConditionSucceeded,
		Timestamp: timestamp,
		Message:   str.Ptr("Credentials were provided."),
		Reason:    str.Ptr("CredentialsProvided"),
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
