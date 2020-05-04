package provisioning

import (
	"errors"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProvideLmsTenantStep_TenantProviderWithRetry(t *testing.T) {
	// given
	now := time.Now()
	opRepo := storage.NewMemoryStorage().Operations()
	tenantStep := NewProvideLmsTenantStep(fakeErrorTenantProvider{}, opRepo, "eu", true)

	inputCreator := newInputCreator()
	operation := internal.ProvisioningOperation{
		Operation: internal.Operation{
			UpdatedAt: now,
		},
		Lms:                    internal.LMS{},
		ProvisioningParameters: `{"Parameters": {"name":"Awesome Lms"}}`,
		InputCreator:           inputCreator,
	}
	opRepo.InsertProvisioningOperation(operation)

	// when
	_, when, err := tenantStep.Run(operation, fixLogger())

	// then
	require.NoError(t, err)
	require.NotZero(t, when)
}

func TestProvideLmsTenantStep_TenantProviderWithError(t *testing.T) {
	runForOptionalAndMandatory(t, func(t *testing.T, isMandatory bool, a asserter) {
		// given
		now := time.Now().Add(-10 * time.Hour)
		opRepo := storage.NewMemoryStorage().Operations()
		tenantStep := NewProvideLmsTenantStep(fakeErrorTenantProvider{}, opRepo, "eu", isMandatory)

		inputCreator := newInputCreator()
		operation := internal.ProvisioningOperation{
			Operation: internal.Operation{
				UpdatedAt: now,
			},
			Lms:                    internal.LMS{},
			ProvisioningParameters: `{"Parameters": {"name":"Awesome Lms"}}`,
			InputCreator:           inputCreator,
		}
		opRepo.InsertProvisioningOperation(operation)

		// when
		op, when, err := tenantStep.Run(operation, fixLogger())

		// then
		a.AssertError(t, err)
		assert.Zero(t, when)
		assert.True(t, op.Lms.Failed)
	})
}

type fakeErrorTenantProvider struct {
}

func (fakeErrorTenantProvider) ProvideLMSTenantID(name, region string) (string, error) {
	return "", errors.New("some error")
}
