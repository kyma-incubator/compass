package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
)

func systemAuthQuery(authID string) string {
	return fmt.Sprintf(`query {
	  result: systemAuth(id: "%s") {
		id
		auth {
		  certCommonName
		}
	  }
	}`, authID)
}

func tenantByExternalIDQuery(tenantID string) string {
	return fmt.Sprintf(`query {
	  	result: tenantByExternalID(id: "%s") {
			id
	  	}
	}`, tenantID)
}

func updateSystemAuthQuery(authID string, gqlAuth graphql.Auth) (string, error) {
	authInput, err := auth.ToGraphQLInput(gqlAuth)
	if err != nil {
		return "", err
	}

	g := graphqlizer.Graphqlizer{}
	gqlAuthInput, err := g.AuthInputToGQL(authInput)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`mutation {
	  	result: updateSystemAuth(authID: "%s", in: %s) {
			id
	  	}
	}`, authID, gqlAuthInput), nil
}
