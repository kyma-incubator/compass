package mock

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func FixPackageInstanceAuth(id string, condition graphql.PackageInstanceAuthStatusCondition) *graphql.PackageInstanceAuth {
	var reason string
	var auth *graphql.Auth

	switch condition {

	case graphql.PackageInstanceAuthStatusConditionSucceeded:
		reason = "CredentialsProvided"
		auth = &graphql.Auth{
			Credential: graphql.BasicCredentialData{
				Username: "username",
				Password: "password",
			},
		}
	case graphql.PackageInstanceAuthStatusConditionFailed:
		reason = "CredentialsNotProvided"
	}

	return &graphql.PackageInstanceAuth{
		ID:      id,
		Context: nil,
		Auth:    auth,
		Status: &graphql.PackageInstanceAuthStatus{
			Condition: condition,
			Timestamp: graphql.Timestamp(time.Now()),
			Message:   "Message",
			Reason:    reason,
		},
	}

}
