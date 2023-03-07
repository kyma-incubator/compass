package fixtures

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixLabelSelector(key, value string) graphql.Label {
	return graphql.Label{
		Key:   key,
		Value: value,
	}
}
