package osb

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb/osbfakes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBindGet(t *testing.T) {
	instanceID := "instanceID"
	bindingID := "bindingID"

	var (
		fakeCredentialsGetter *osbfakes.FakePackageCredentialsFetcherForInstance
		be                    *GetBindingEndpoint
		packageInstanceAuth   *director.PackageInstanceCredentialsOutput
	)

	setup := func() {
		fakeCredentialsGetter = &osbfakes.FakePackageCredentialsFetcherForInstance{}

		be = &GetBindingEndpoint{
			credentialsGetter: fakeCredentialsGetter,
		}

		packageInstanceAuth = &director.PackageInstanceCredentialsOutput{
			InstanceAuth: &graphql.PackageInstanceAuth{
				Status: &graphql.PackageInstanceAuthStatus{
					Condition: graphql.PackageInstanceAuthStatusConditionSucceeded,
				},
				Auth: &graphql.Auth{
					Credential: &graphql.BasicCredentialData{
						Username: "username",
						Password: "password",
					},
				},
			},
			TargetURLs: map[string]string{},
		}
	}

	t.Run("Success", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			packageInstanceAuth,
			nil,
		)
		binding, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.NoError(t, err)
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
		creds, ok := binding.Credentials.(BindingCredentials)
		assert.True(t, ok)
		basicCreds := creds.AuthDetails.Credentials.ToCredentials().BasicAuth
		assert.Equal(t, "username", basicCreds.Username)
		assert.Equal(t, "password", basicCreds.Password)
	})

	t.Run("When credentials mapper returns an error", func(t *testing.T) {
		setup()
		packageInstanceAuth.InstanceAuth.Auth.Credential = graphql.BasicCredentialData{
			Username: "username",
			Password: "password",
		}
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			packageInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while mapping to binding credentials: got unknown credential type")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter returns an error", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			nil,
			errors.New("some-error"),
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while getting package instance credentials from director")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter returns not found", func(t *testing.T) {
		setup()
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			nil,
			&notFoundErr{},
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter is in pending state", func(t *testing.T) {
		setup()
		packageInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionPending
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			packageInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter is in unused state", func(t *testing.T) {
		setup()
		packageInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionUnused
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			packageInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "binding cannot be fetched")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
	})

	t.Run("When credentials getter is in failed state", func(t *testing.T) {
		setup()
		packageInstanceAuth.InstanceAuth.Status.Condition = graphql.PackageInstanceAuthStatusConditionFailed
		fakeCredentialsGetter.FetchPackageInstanceCredentialsReturns(
			packageInstanceAuth,
			nil,
		)
		_, err := be.GetBinding(context.TODO(), instanceID, bindingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credentials status is not success")
		assert.Equal(t, 1, fakeCredentialsGetter.FetchPackageInstanceCredentialsCallCount())
	})
}
