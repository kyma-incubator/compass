package provisioning

import (
	"testing"
	"time"

	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertStep_RunFreshOperation(t *testing.T) {
	// given
	repo := storage.NewMemoryStorage().Operations()
	svc := NewLmsCertificatesStep(nil, repo, false)
	// a fresh operation
	operation := internal.ProvisioningOperation{
		Lms: internal.LMS{},
	}

	// when
	_, _, err := svc.Run(operation, fixLogger())

	//then
	require.Error(t, err)
}

func TestCertStep_Run(t *testing.T) {
	// given
	cli, tID := newFakeClientWithTenant(0)
	repo := storage.NewMemoryStorage().Operations()
	svc := NewLmsCertificatesStep(cli, repo, false)
	operation := internal.ProvisioningOperation{
		Lms: internal.LMS{
			TenantID: tID,
		},
		ProvisioningParameters: `{"name": "awesome"}`,
		InputCreator:           newInputCreator(),
	}
	repo.InsertProvisioningOperation(operation)

	// when
	op, duration, err := svc.Run(operation, fixLogger())

	// then
	require.NoError(t, err)
	assert.Zero(t, duration.Seconds())
	assert.False(t, op.Lms.Failed)

	assert.True(t, cli.IsCertRequestedForTenant(tID))
}

func TestCertStep_TenantNotReady(t *testing.T) {
	// given
	cli, tID := newFakeClientWithTenant(time.Hour)
	repo := storage.NewMemoryStorage().Operations()
	svc := NewLmsCertificatesStep(cli, repo, false)
	operation := internal.ProvisioningOperation{
		Lms: internal.LMS{
			TenantID:    tID,
			RequestedAt: time.Now(),
		},
		ProvisioningParameters: "{}",
	}
	repo.InsertProvisioningOperation(operation)

	// when
	op, duration, err := svc.Run(operation, fixLogger())

	// then
	require.NoError(t, err)
	assert.NotZero(t, duration.Seconds())
	assert.False(t, op.Lms.Failed)

	// do not expect call to LMS
	assert.False(t, cli.IsCertRequestedForTenant(tID))
}

func TestCertStep_TenantNotReadyTimeout(t *testing.T) {
	// given
	cli, tID := newFakeClientWithTenant(time.Hour)
	repo := storage.NewMemoryStorage().Operations()
	svc := NewLmsCertificatesStep(cli, repo, false)
	operation := internal.ProvisioningOperation{
		Lms: internal.LMS{
			TenantID:    tID,
			RequestedAt: time.Now().Add(-10 * time.Hour), // very old
		},
		ProvisioningParameters: `{"name": "awesome"}`,
	}
	repo.InsertProvisioningOperation(operation)

	// when
	op, duration, err := svc.Run(operation, fixLogger())

	// then
	require.NoError(t, err)
	assert.Zero(t, duration.Seconds())
	assert.True(t, op.Lms.Failed)

	// do not expect call to LMS
	assert.False(t, cli.IsCertRequestedForTenant(tID))
}

func TestCertStep_TenantNotReadyTimeout_Mandatory(t *testing.T) {

	for _, tc := range []struct {
		name string
		isMandatory bool
	} {
{name: "step mandatory", isMandatory: true},
{name: "step optional", isMandatory: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			cli, tID := newFakeClientWithTenant(time.Hour)
			repo := storage.NewMemoryStorage().Operations()
			svc := NewLmsCertificatesStep(cli, repo, tc.isMandatory)
			operation := internal.ProvisioningOperation{
				Lms: internal.LMS{
					TenantID:    tID,
					RequestedAt: time.Now().Add(-10 * time.Hour), // very old
				},
				ProvisioningParameters: `{"name": "awesome"}`,
			}
			repo.InsertProvisioningOperation(operation)

			// when
			op, duration, err := svc.Run(operation, fixLogger())

			// then
			if tc.isMandatory {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Zero(t, duration.Seconds())
			assert.True(t, op.Lms.Failed)

			// do not expect call to LMS
			assert.False(t, cli.IsCertRequestedForTenant(tID))
		})

	}


}

func assertLmsMarkedFailed(t *testing.T, operation internal.ProvisioningOperation, duration time.Duration, err error) {
	assert.NoError(t, err)
	assert.Zero(t, duration.Seconds())
	assert.True(t, operation.Lms.Failed)
}

func assertOperationFailed(t *testing.T, operation internal.ProvisioningOperation, duration time.Duration, err error) {
	assert.Error(t, err)
	assert.Zero(t, duration.Seconds())
	assert.True(t, operation.Lms.Failed)
}

func TestLmsStepsHappyPath(t *testing.T) {
	// given
	lmsClient := lms.NewFakeClient(0)
	opRepo := storage.NewMemoryStorage().Operations()
	tRepo := storage.NewMemoryStorage().LMSTenants()
	certStep := NewLmsCertificatesStep(lmsClient, opRepo, false)
	tManager := lms.NewTenantManager(tRepo, lmsClient, fixLogger())
	tenantStep := NewProvideLmsTenantStep(tManager, opRepo, "eu", false)

	inputCreator := newInputCreator()
	operation := internal.ProvisioningOperation{
		Lms:                    internal.LMS{},
		ProvisioningParameters: `{"Parameters": {"name":"Awesome Lms"}}`,
		InputCreator:           inputCreator,
	}
	opRepo.InsertProvisioningOperation(operation)

	// when
	op, when, err := tenantStep.Run(operation, fixLogger())

	// then
	require.NoError(t, err)
	require.Zero(t, when)
	assert.NotEmpty(t, op.Lms.TenantID)

	// when
	op, when, err = certStep.Run(op, fixLogger())

	// then
	require.NoError(t, err)
	require.Zero(t, when)
	lmsClient.IsCertRequestedForTenant(op.Lms.TenantID)

	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.conf.Output.forward.enabled", Value: "true"})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.forward.host", Value: fmt.Sprintf("forward.%s", lms.FakeLmsHost)})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.forward.port", Value: "8443"})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.forward.tls.ca", Value: "Y2VydC1jYS1wYXlsb2Fk"})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.forward.tls.cert", Value: "c2lnbmVkLWNlcnQtcGF5bG9hZA=="})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.forward.tls.key", Value: "cHJpdmF0ZS1rZXk="})

	inputCreator.AssertLabel(t, "operator_lmsUrl", fmt.Sprintf("https://kibana.%s", lms.FakeLmsHost))
}

func newFakeClientWithTenant(timeToReady time.Duration) (*lms.FakeClient, string) {
	lmsClient := lms.NewFakeClient(timeToReady)
	out, _ := lmsClient.CreateTenant(lms.CreateTenantInput{
		Name: "some-tenant",
	})

	return lmsClient, out.ID
}

func newInputCreator() *simpleInputCreator {
	return &simpleInputCreator{
		overrides: make(map[string][]*gqlschema.ConfigEntryInput, 0),
		labels:    make(map[string]string),
	}
}

type simpleInputCreator struct {
	overrides map[string][]*gqlschema.ConfigEntryInput
	labels    map[string]string
}

func (c *simpleInputCreator) SetLabel(key, val string) internal.ProvisionInputCreator {
	c.labels[key] = val
	return c
}

func (c *simpleInputCreator) SetOverrides(component string, overrides []*gqlschema.ConfigEntryInput) internal.ProvisionInputCreator {
	return c
}

func (c *simpleInputCreator) Create() (gqlschema.ProvisionRuntimeInput, error) {
	return gqlschema.ProvisionRuntimeInput{}, nil
}

func (c *simpleInputCreator) SetProvisioningParameters(params internal.ProvisioningParametersDTO) internal.ProvisionInputCreator {
	return c
}

func (c *simpleInputCreator) AppendOverrides(component string, overrides []*gqlschema.ConfigEntryInput) internal.ProvisionInputCreator {
	c.overrides[component] = append(c.overrides[component], overrides...)
	return c
}

func (c *simpleInputCreator) AppendGlobalOverrides(overrides []*gqlschema.ConfigEntryInput) internal.ProvisionInputCreator {
	return c
}

func (c *simpleInputCreator) AssertOverride(t *testing.T, component string, cei gqlschema.ConfigEntryInput) {
	cmpOverrides, found := c.overrides[component]
	require.True(t, found)

	for _, item := range cmpOverrides {
		if item.Key == cei.Key {
			assert.Equal(t, cei, *item)
			return
		}
	}
	assert.Failf(t, "Overrides assert failed", "Expected component override not found: %+v", cei)
}

func (c *simpleInputCreator) AssertLabel(t *testing.T, key, expectedValue string) {
	value, found := c.labels[key]
	require.True(t, found)
	assert.Equal(t, expectedValue, value)
}
