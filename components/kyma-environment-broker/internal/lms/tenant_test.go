package lms_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerProvideLMSTenantIDIfNotExists(t *testing.T) {
	// given
	lmsStorage := storage.NewMemoryStorage().LMSTenants()
	tCreator := &automock.TenantCreator{}
	tCreator.On("CreateTenant", lms.CreateTenantInput{
		Name:   "newtenant",
		Region: "eu",
	}).Return(lms.CreateTenantOutput{ID: "tenant-id-001"}, nil)
	defer tCreator.AssertExpectations(t)

	svc := lms.NewTenantManager(lmsStorage, tCreator, logrus.StandardLogger())

	// when
	id, err := svc.ProvideLMSTenantID("newtenant", "eu")
	require.NoError(t, err)

	// then
	assert.Equal(t, "tenant-id-001", id)
	tenantInStorage, _, _ := lmsStorage.FindTenantByName("newtenant", "eu")
	assertTenant(t, internal.LMSTenant{Region: "eu", Name: "newtenant", ID: "tenant-id-001"}, tenantInStorage)
}

func TestManagerProvideLMSTenantIDIfExists(t *testing.T) {
	// given
	lmsStorage := storage.NewMemoryStorage().LMSTenants()
	lmsStorage.InsertTenant(internal.LMSTenant{Region: "eu", Name: "newtenant", ID: "tenant-id-001"})
	tCreator := &automock.TenantCreator{}
	defer tCreator.AssertExpectations(t)

	svc := lms.NewTenantManager(lmsStorage, tCreator, logrus.StandardLogger())

	// when
	id, err := svc.ProvideLMSTenantID("newtenant", "eu")
	require.NoError(t, err)

	// then
	assert.Equal(t, "tenant-id-001", id)
	tenantInStorage, _, _ := lmsStorage.FindTenantByName("newtenant", "eu")
	assertTenant(t, internal.LMSTenant{Region: "eu", Name: "newtenant", ID: "tenant-id-001"}, tenantInStorage)
}

func assertTenant(t *testing.T, expected internal.LMSTenant, given internal.LMSTenant) {
	expected.CreatedAt = given.CreatedAt //do not compare created at
	assert.Equal(t, expected, given)
}
