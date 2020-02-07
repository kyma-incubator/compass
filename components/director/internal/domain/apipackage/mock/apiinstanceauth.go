package mock

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func FixAPIInstanceAuth(id string, condition graphql.APIInstanceAuthStatusCondition) *graphql.APIInstanceAuth {
	var reason string
	var auth *graphql.Auth

	switch condition {

	case graphql.APIInstanceAuthStatusConditionSucceeded:
		reason = "CredentialsProvided"
		auth = &graphql.Auth{
			Credential: graphql.BasicCredentialData{
				Username: "username",
				Password: "password",
			},
		}
	case graphql.APIInstanceAuthStatusConditionFailed:
		reason = "CredentialsNotProvided"
	}

	return &graphql.APIInstanceAuth{
		ID:      id,
		Context: nil,
		Auth:    auth,
		Status: &graphql.APIInstanceAuthStatus{
			Condition: condition,
			Timestamp: graphql.Timestamp(time.Now()),
			Message:   "Message",
			Reason:    reason,
		},
	}

}
