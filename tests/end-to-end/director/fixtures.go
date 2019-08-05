package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixBasicAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "admin",
				Password: "secret",
			},
		},
		AdditionalHeaders: &graphql.HttpHeaders{
			"headerA": []string{"ha1", "ha2"},
			"headerB": []string{"hb1", "hb2"},
		},
		AdditionalQueryParams: &graphql.QueryParams{
			"qA": []string{"qa1", "qa2"},
			"qB": []string{"qb1", "qb2"},
		},
	}
}

func fixOauthAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Oauth: &graphql.OAuthCredentialDataInput{
				URL:          "http://oauth/token",
				ClientID:     "clientID",
				ClientSecret: "clientSecret",
			},
		},
	}
}

func fixDepracatedVersion1() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           "v1",
		Deprecated:      ptrBool(true),
		ForRemoval:      ptrBool(true),
		DeprecatedSince: ptrString("v5"),
	}
}

func fixRuntimeInput(placeholder string) graphql.RuntimeInput {
	return graphql.RuntimeInput{
		Name:        placeholder,
		Description: ptrString(fmt.Sprintf("%s-description", placeholder)),
		Labels:      &graphql.Labels{"placeholder": []interface{}{"placeholder"}},
	}
}
