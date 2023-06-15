package fixtures

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixAPIDefinitionInputWithTargetURL(targetURL string) graphql.APIDefinitionInput {
	apiInput := FixAPIDefinitionInput()
	apiInput.TargetURL = targetURL

	return apiInput
}
