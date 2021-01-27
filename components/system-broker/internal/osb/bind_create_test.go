package osb

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb/osbfakes"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBindCreate(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsCreator *osbfakes.FakePackageCredentialsCreateRequester
		fakeCredentialsGetter  *osbfakes.FakePackageCredentialsFetcher
		be                     *BindEndpoint
		details                domain.BindDetails
		bundleInstanceAuth     *director.BundleInstanceAuthOutput
	)

	setup := func() {
		fakeCredentialsCreator = &osbfakes.FakePackageCredentialsCreateRequester{}
		fakeCredentialsGetter = &osbfakes.FakePackageCredentialsFetcher{}

		be = &BindEndpoint{
			credentialsCreator: fakeCredentialsCreator,
			credentialsGetter:  fakeCredentialsGetter,
		}

		details = domain.BindDetails{
			ServiceID: "serviceID",
			PlanID:    "planID",
		}

		bundleInstanceAuth = &director.BundleInstanceAuthOutput{
			InstanceAuth: &graphql.PackageInstanceAuth{
				ID: "instanceAuthID",
				Status: &graphql.PackageInstanceAuthStatus{
					Condition: graphql.PackageInstanceAuthStatusConditionSucceeded,
				},
			},
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		binding, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.NoError(t, err)
		assert.Equal(t, 0, fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationCallCount())
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, "bind_operation", binding.OperationData)
	})

	t.Run("When no package instance auth exists", func(t *testing.T) {
		setup()
		details.RawContext = []byte(`{"org_guid": "orgID"}`)
		details.RawParameters = []byte(`{}`)
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionPending
		fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationReturns(
			bundleInstanceAuth,
			nil,
		)
		binding, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationCallCount())
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, "bind_operation", binding.OperationData)
	})

	t.Run("When async is not supported by platform", func(t *testing.T) {
		setup()
		_, err := be.Bind(context.TODO(), instanceID, bindingID, details, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "This service plan requires client support for asynchronous service operations")
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting package instance credentials from director")
	})

	t.Run("When raw parameters are not JSON", func(t *testing.T) {
		setup()
		details.RawParameters = []byte(`not a json`)
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionPending
		fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while unmarshaling raw parameters")
	})

	t.Run("When OSB context is not JSON", func(t *testing.T) {
		setup()
		details.RawParameters = []byte(`{}`)
		details.RawContext = []byte(`not a json`)
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionPending
		fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while unmarshaling raw context")
	})

	t.Run("When package instance credential requester returns an error", func(t *testing.T) {
		setup()
		details.RawContext = []byte(`{"org_guid": "orgID"}`)
		details.RawParameters = []byte(`{}`)
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionPending
		fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while requesting package instance credentials creation from director")
		assert.Equal(t, 1, fakeCredentialsCreator.RequestPackageInstanceCredentialsCreationCallCount())
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
	})

	t.Run("When package instance auth status is failed", func(t *testing.T) {
		setup()
		bundleInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionFailed
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			bundleInstanceAuth,
			nil,
		)
		_, err := be.Bind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requesting package instance credentials from director failed, got status")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
	})
}

type notFoundErr struct{}

func (e *notFoundErr) Error() string {
	return "fake not found error"
}

func (e *notFoundErr) NotFound() bool {
	return true
}
