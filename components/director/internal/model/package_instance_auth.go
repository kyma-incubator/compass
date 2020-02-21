package model

import "time"

type PackageInstanceAuth struct {
	ID          string
	PackageID   string
	Tenant      string
	Context     *string
	InputParams *string
	Auth        *Auth
	Status      PackageInstanceAuthStatus
}

// TODO: Final models and unit tests will be introduced in https://github.com/kyma-incubator/compass/issues/806
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
	Message   string
	Reason    string
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
	out := PackageInstanceAuth{
		ID:          id,
		PackageID:   packageID,
		Tenant:      tenant,
		Context:     ri.Context,
		InputParams: ri.InputParams,
		Auth:        nil,
		Status: PackageInstanceAuthStatus{
			Condition: PackageInstanceAuthStatusConditionPending,
			Timestamp: timestamp,
			Message:   "Credentials were not yet provided.",
			Reason:    "CredentialsNotProvided",
		},
	}
	if auth != nil {
		out.Auth = auth
		out.Status = PackageInstanceAuthStatus{
			Condition: PackageInstanceAuthStatusConditionSucceeded,
			Timestamp: timestamp,
			Message:   "Credentials were provided.",
			Reason:    "CredentialsProvided",
		}
	}
	return out
}

// Input type for setPackageInstanceAuth
type PackageInstanceAuthSetInput struct {
	Auth   *AuthInput
	Status PackageInstanceAuthStatusInput
}
type PackageInstanceAuthStatusInput struct {
	Condition PackageInstanceAuthSetStatusConditionInput
	Message   string
	Reason    string
}

func (si PackageInstanceAuthStatusInput) ToPackageInstanceAuthStatus(timestamp time.Time) PackageInstanceAuthStatus {
	return PackageInstanceAuthStatus{
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
