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

func TestDeleteBinding(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsDeleter      *osbfakes.FakePackageCredentialsDeleteRequester
		fakeCredentialsGetter       *osbfakes.FakePackageCredentialsFetcher
		be                          *UnbindEndpoint
		details                     domain.UnbindDetails
		packageInstanceAuth         *director.PackageInstanceAuthOutput
		packageInstanceAuthDeletion *director.PackageInstanceAuthDeletionOutput
	)

	setup := func() {
		fakeCredentialsDeleter = &osbfakes.FakePackageCredentialsDeleteRequester{}
		fakeCredentialsGetter = &osbfakes.FakePackageCredentialsFetcher{}

		be = &UnbindEndpoint{
			credentialsDeleter: fakeCredentialsDeleter,
			credentialsGetter:  fakeCredentialsGetter,
		}

		details = domain.UnbindDetails{
			ServiceID: "serviceID",
			PlanID:    "planID",
		}

		packageInstanceAuth = &director.PackageInstanceAuthOutput{
			InstanceAuth: &graphql.PackageInstanceAuth{
				ID: "instanceAuthID",
				Status: &graphql.PackageInstanceAuthStatus{
					Condition: graphql.PackageInstanceAuthStatusConditionSucceeded,
				},
			},
		}

		packageInstanceAuthDeletion = &director.PackageInstanceAuthDeletionOutput{
			ID: "instanceAuthID",
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			packageInstanceAuth,
			nil,
		)
		fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionReturns(
			packageInstanceAuthDeletion,
			nil,
		)
		binding, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, 1, fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionCallCount())
		assert.Equal(t, "unbind_operation", binding.OperationData)
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting package instance credentials from director")
	})

	t.Run("When credentials are not found", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			nil,
			&notFoundErr{},
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding does not exist")
	})

	t.Run("When credentials are already being deleted", func(t *testing.T) {
		setup()
		packageInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionUnused
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			packageInstanceAuth,
			nil,
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, 0, fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionCallCount())
	})

	t.Run("When credentials deleter returns not found", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			packageInstanceAuth,
			nil,
		)
		fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionReturns(
			nil,
			&notFoundErr{},
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding does not exist")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, 1, fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionCallCount())
	})

	t.Run("When credentials deleter returns error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceAuthReturns(
			packageInstanceAuth,
			nil,
		)
		fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionReturns(
			nil,
			errors.New("some error"),
		)
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while requesting package instance credentials deletion from director")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceAuthCallCount())
		assert.Equal(t, 1, fakeCredentialsDeleter.RequestPackageInstanceCredentialsDeletionCallCount())
	})

	t.Run("When async is not supported by platform", func(t *testing.T) {
		setup()
		_, err := be.Unbind(context.TODO(), instanceID, bindingID, details, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "This service plan requires client support for asynchronous service operations")
	})
}
