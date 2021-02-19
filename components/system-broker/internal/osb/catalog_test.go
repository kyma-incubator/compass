package osb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types/typesfakes"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb/osbfakes"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
)

func TestServicesReturnsErrorForInvalidApplications(t *testing.T) {
	var (
		fakeApplicationsLister *typesfakes.FakeApplicationsLister
		fakeConverter          *osbfakes.FakeConverter
		endpoint               *osb.CatalogEndpoint
		app                    *schema.ApplicationExt
		output                 *director.ApplicationsOutput
		svc                    *domain.Service
	)

	setup := func() {
		fakeApplicationsLister = &typesfakes.FakeApplicationsLister{}
		fakeConverter = &osbfakes.FakeConverter{}
		endpoint = osb.NewCatalogEndpoint(fakeApplicationsLister, fakeConverter)

		app = &schema.ApplicationExt{
			Application: schema.Application{
				Name: "test-app",
			},
		}

		output = &director.ApplicationsOutput{
			Result: &schema.ApplicationPageExt{
				ApplicationPage: schema.ApplicationPage{},
				Data:            []*schema.ApplicationExt{app},
			},
		}

		svc = &domain.Service{
			ID:                   "id",
			Name:                 "app1",
			Description:          "service generated from system with name app1",
			Bindable:             true,
			InstancesRetrievable: false,
			BindingsRetrievable:  true,
			PlanUpdatable:        false,
			Plans:                generateExpectations(1, 1, 1),
			Metadata: &domain.ServiceMetadata{
				DisplayName:         "app1",
				ProviderDisplayName: "provider",
				AdditionalMetadata:  schema.Labels{"key": "value"},
			},
		}
	}

	t.Run("Invalid application", func(t *testing.T) {
		setup()

		testError := errors.New("test-error")

		fakeApplicationsLister.FetchApplicationsReturns(output, nil)
		fakeConverter.ConvertReturns(nil, testError)

		_, err := endpoint.Services(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.Equal(t, 1, fakeApplicationsLister.FetchApplicationsCallCount())
		assert.Equal(t, 1, fakeConverter.ConvertCallCount())

	})

	t.Run("Fetch application returns error", func(t *testing.T) {
		setup()

		testError := errors.New("test-error")

		fakeApplicationsLister.FetchApplicationsReturns(nil, testError)

		_, err := endpoint.Services(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not build catalog")
		assert.Equal(t, 1, fakeApplicationsLister.FetchApplicationsCallCount())
		assert.Equal(t, 0, fakeConverter.ConvertCallCount())

	})

	t.Run("Current application data is nil", func(t *testing.T) {
		setup()

		output.Result.Data = []*schema.ApplicationExt{nil}
		fakeApplicationsLister.FetchApplicationsReturns(output, nil)

		response, err := endpoint.Services(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 0, len(response))
		assert.Equal(t, 1, fakeApplicationsLister.FetchApplicationsCallCount())
		assert.Equal(t, 0, fakeConverter.ConvertCallCount())
	})

	t.Run("Service has one plan", func(t *testing.T) {
		setup()

		fakeApplicationsLister.FetchApplicationsReturns(output, nil)
		fakeConverter.ConvertReturns(svc, nil)

		response, err := endpoint.Services(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, []domain.Service{*svc}, response)
		assert.Equal(t, 1, fakeApplicationsLister.FetchApplicationsCallCount())
		assert.Equal(t, 1, fakeConverter.ConvertCallCount())
	})

	t.Run("Service has multiple plans", func(t *testing.T) {
		setup()

		svc.Plans = generateExpectations(2, 2, 3)

		fakeApplicationsLister.FetchApplicationsReturns(output, nil)
		fakeConverter.ConvertReturns(svc, nil)

		response, err := endpoint.Services(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, []domain.Service{*svc}, response)
		assert.Equal(t, 1, fakeApplicationsLister.FetchApplicationsCallCount())
		assert.Equal(t, 1, fakeConverter.ConvertCallCount())
	})

	t.Run("Two services with one plan", func(t *testing.T) {
		setup()

		output.Result.Data = []*schema.ApplicationExt{app, app}

		fakeApplicationsLister.FetchApplicationsReturns(output, nil)
		fakeConverter.ConvertReturns(svc, nil)

		response, err := endpoint.Services(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, []domain.Service{*svc, *svc}, response)
		assert.Equal(t, 1, fakeApplicationsLister.FetchApplicationsCallCount())
		assert.Equal(t, 2, fakeConverter.ConvertCallCount())
	})
}
