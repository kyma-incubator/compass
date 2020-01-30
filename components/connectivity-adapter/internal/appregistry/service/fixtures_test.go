package service_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	svcautomock "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixGQLApplicationRegisterInput(name, description string) graphql.ApplicationRegisterInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}
	return graphql.ApplicationRegisterInput{
		Name:        name,
		Description: &description,
		Labels:      &labels,
		Webhooks: []*graphql.WebhookInput{
			{URL: "webhook1.foo.bar"},
			{URL: "webhook2.foo.bar"},
		},
		APIDefinitions: []*graphql.APIDefinitionInput{
			{Name: "api1", TargetURL: "foo.bar"},
			{Name: "api2", TargetURL: "foo.bar2"},
		},
	}
}

func SuccessfulValidatorFn(input model.ServiceDetails) func() *svcautomock.Validator {
	return func() *svcautomock.Validator {
		validator := &svcautomock.Validator{}
		validator.On("Validate", input).Return(nil).Once()
		return validator
	}
}

func SuccessfulDetailsToGQLInputConverterFn(input model.ServiceDetails, output graphql.ApplicationRegisterInput) func() *svcautomock.Converter {
	return func() *svcautomock.Converter {
		converter := &svcautomock.Converter{}
		converter.On("DetailsToGraphQLInput", input).Return(output, nil).Once()
		return converter
	}
}

func EmptyConverterFn() func() *svcautomock.Converter {
	return func() *svcautomock.Converter {
		return &svcautomock.Converter{}
	}
}

func EmptyValidatorFn() func() *svcautomock.Validator {
	return func() *svcautomock.Validator {
		return &svcautomock.Validator{}
	}
}

func EmptyGraphQLClientFn() func() *automock.GraphQLClient {
	return func() *automock.GraphQLClient {
		return &automock.GraphQLClient{}
	}
}

func EmptyGraphQLRequestBuilderFn() func() *svcautomock.GraphQLRequestBuilder {
	return func() *svcautomock.GraphQLRequestBuilder {
		return &svcautomock.GraphQLRequestBuilder{}
	}
}

func SuccessfulRegisterAppGraphQLRequestBuilderFn(input graphql.ApplicationRegisterInput, output *gcli.Request) func() *svcautomock.GraphQLRequestBuilder {
	return func() *svcautomock.GraphQLRequestBuilder {
		gqlRequestBuilder := &svcautomock.GraphQLRequestBuilder{}
		gqlRequestBuilder.On("RegisterApplicationRequest", input).Return(output, nil).Once()
		return gqlRequestBuilder
	}
}

func SingleErrorLoggerAssertions(errMessage string) func(t *testing.T, hook *test.Hook) {
	return func(t *testing.T, hook *test.Hook) {
		assert.Equal(t, 1, len(hook.AllEntries()))
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, log.ErrorLevel, entry.Level)
		assert.Equal(t, errMessage, entry.Message)
	}
}
