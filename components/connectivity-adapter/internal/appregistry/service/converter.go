package service

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) DetailsToGraphQLInput(in model.ServiceDetails) (graphql.ApplicationRegisterInput, error) {
	// This is just a temporary fixture for testing purposes
	// TODO: Replace with production-grade implementation
	return graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
	}, nil
}

func (c *converter) GraphQLToDetailsModel(in graphql.ApplicationExt) (model.ServiceDetails, error) {
	// This is just a temporary fixture for testing purposes
	// TODO: Replace with production-grade implementation

	return model.ServiceDetails{
		Name:     "foo",
		Provider: "test",
		Labels: &map[string]string{
			"foo": "bar",
		},
	}, nil

}

func (c *converter) GraphQLToModel(in graphql.ApplicationExt) (model.Service, error) {
	panic("not implemented")
}
