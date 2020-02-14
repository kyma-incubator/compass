package provisioning

import (
	"testing"
	"time"

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
	svc := NewLmsCertificatesStep(nil, repo)
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
	svc := NewLmsCertificatesStep(cli, repo)
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
	svc := NewLmsCertificatesStep(cli, repo)
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
	svc := NewLmsCertificatesStep(cli, repo)
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

func TestLmsStepsHappyPath(t *testing.T) {
	// given
	lmsClient := lms.NewFakeClient(0)
	opRepo := storage.NewMemoryStorage().Operations()
	tRepo := storage.NewMemoryStorage().LMSTenants()
	certStep := NewLmsCertificatesStep(lmsClient, opRepo)
	tManager := lms.NewTenantManager(tRepo, lmsClient, fixLogger())
	tenantStep := NewProvideLmsTenantStep(tManager, opRepo)

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
		Key: "fluent-bit.conf.Service.Flush", Value: "30"})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.conf.Output.Elasticsearch.enabled", Value: "true"})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.es.host", Value: lms.FakeLmsHost})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.es.port", Value: "443"})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.es.tls_ca", Value: lms.FakeCaCertificate})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.es.tls_crt", Value: lms.FakeSignedCertificate})
	inputCreator.AssertOverride(t, "logging", gqlschema.ConfigEntryInput{
		Key: "fluent-bit.backend.es.tls_key", Value: lms.FakePrivateKey})
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
	}
}

type simpleInputCreator struct {
	overrides map[string][]*gqlschema.ConfigEntryInput
}

func (c *simpleInputCreator) SetRuntimeLabels(instanceID, SubAccountID string) internal.ProvisionInputCreator {
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
