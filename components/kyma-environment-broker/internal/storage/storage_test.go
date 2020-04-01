// +build database_integration

package storage

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"

	"time"

	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaInitializer(t *testing.T) {
	ctx := context.Background()

	cleanupNetwork, err := EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	t.Run("Init tests", func(t *testing.T) {
		t.Run("Should initialize database when schema not applied", func(t *testing.T) {
			// given
			containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			// when
			connection, err := postsql.InitializeDatabase(cfg.ConnectionURL())

			require.NoError(t, err)
			require.NotNil(t, connection)

			defer CloseDatabase(t, connection)

			// then
			assert.NoError(t, err)
		})

		t.Run("Should return error when failed to connect to the database", func(t *testing.T) {
			containerCleanupFunc, _, err := InitTestDBContainer(t, ctx, "test_DB_3")
			require.NoError(t, err)
			defer containerCleanupFunc()

			// given
			connString := "bad connection string"

			// when
			connection, err := postsql.InitializeDatabase(connString)

			// then
			assert.Error(t, err)
			assert.Nil(t, connection)
		})
	})

	// Instances
	t.Run("Instances", func(t *testing.T) {
		t.Run("Should create and update instance", func(t *testing.T) {
			// given
			containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			err = InitTestDBTables(t, cfg.ConnectionURL())
			require.NoError(t, err)

			// when
			brokerStorage, err := NewFromConfig(cfg, logrus.StandardLogger())

			require.NoError(t, err)
			require.NotNil(t, brokerStorage)

			testData := "test"

			fixInstance := fixInstance(testData)
			err = brokerStorage.Instances().Insert(*fixInstance)
			require.NoError(t, err)

			fixInstance.DashboardURL = "diff"
			err = brokerStorage.Instances().Update(*fixInstance)
			require.NoError(t, err)

			// then
			inst, err := brokerStorage.Instances().GetByID(testData)
			assert.NoError(t, err)
			require.NotNil(t, inst)

			assert.Equal(t, fixInstance.InstanceID, inst.InstanceID)
			assert.Equal(t, fixInstance.RuntimeID, inst.RuntimeID)
			assert.Equal(t, fixInstance.GlobalAccountID, inst.GlobalAccountID)
			assert.Equal(t, fixInstance.ServiceID, inst.ServiceID)
			assert.Equal(t, fixInstance.ServicePlanID, inst.ServicePlanID)
			assert.Equal(t, fixInstance.DashboardURL, inst.DashboardURL)
			assert.Equal(t, fixInstance.ProvisioningParameters, inst.ProvisioningParameters)
			assert.NotEmpty(t, inst.CreatedAt)
			assert.NotEmpty(t, inst.UpdatedAt)
			assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", inst.DelatedAt.String())

			// when
			err = brokerStorage.Instances().Delete(fixInstance.InstanceID)

			// then
			assert.NoError(t, err)
			_, err = brokerStorage.Instances().GetByID(fixInstance.InstanceID)
			assert.True(t, dberr.IsNotFound(err))

			// when
			err = brokerStorage.Instances().Delete(fixInstance.InstanceID)
			assert.NoError(t, err, "deletion non existing instance must not cause any error")
		})
	})

	t.Run("Operations", func(t *testing.T) {
		t.Run("Provisioning", func(t *testing.T) {
			containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			givenOperation := internal.ProvisioningOperation{
				Operation: internal.Operation{
					ID:    "operation-id",
					State: domain.InProgress,
					// used Round and set timezone to be able to compare timestamps
					CreatedAt:              time.Now().Truncate(time.Millisecond),
					UpdatedAt:              time.Now().Truncate(time.Millisecond).Add(time.Second),
					InstanceID:             "inst-id",
					ProvisionerOperationID: "target-op-id",
					Description:            "description",
					Version:                1,
				},
				Lms:                    internal.LMS{TenantID: "tenant-id"},
				ProvisioningParameters: `{"k":"v"}`,
			}

			err = InitTestDBTables(t, cfg.ConnectionURL())
			require.NoError(t, err)

			brokerStorage, err := NewFromConfig(cfg, logrus.StandardLogger())
			require.NoError(t, err)

			svc := brokerStorage.Operations()

			// when
			err = svc.InsertProvisioningOperation(givenOperation)
			require.NoError(t, err)

			ops, err := svc.GetOperationsInProgressByType(dbmodel.OperationTypeProvision)
			require.NoError(t, err)
			assert.Len(t, ops, 1)
			assertOperation(t, givenOperation.Operation, ops[0])

			gotOperation, err := svc.GetProvisioningOperationByID("operation-id")
			require.NoError(t, err)

			op, err := svc.GetOperationByID("operation-id")
			require.NoError(t, err)
			assert.Equal(t, givenOperation.Operation.ID, op.ID)

			// then
			assertProvisioningOperation(t, givenOperation, *gotOperation)

			// when
			gotOperation.Description = "new modified description"
			_, err = svc.UpdateProvisioningOperation(*gotOperation)
			require.NoError(t, err)

			// then
			gotOperation2, err := svc.GetProvisioningOperationByID("operation-id")
			require.NoError(t, err)

			assert.Equal(t, "new modified description", gotOperation2.Description)

		})
		t.Run("Deprovisioning", func(t *testing.T) {
			containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			givenOperation := internal.DeprovisioningOperation{
				Operation: internal.Operation{
					ID:    "operation-id",
					State: domain.InProgress,
					// used Round and set timezone to be able to compare timestamps
					CreatedAt:              time.Now().Truncate(time.Millisecond),
					UpdatedAt:              time.Now().Truncate(time.Millisecond).Add(time.Second),
					InstanceID:             "inst-id",
					ProvisionerOperationID: "target-op-id",
					Description:            "description",
					Version:                1,
				},
			}

			err = InitTestDBTables(t, cfg.ConnectionURL())
			require.NoError(t, err)

			brokerStorage, err := NewFromConfig(cfg, logrus.StandardLogger())
			require.NoError(t, err)

			svc := brokerStorage.Operations()

			// when
			err = svc.InsertDeprovisioningOperation(givenOperation)
			require.NoError(t, err)

			ops, err := svc.GetOperationsInProgressByType(dbmodel.OperationTypeDeprovision)
			require.NoError(t, err)
			assert.Len(t, ops, 1)
			assertOperation(t, givenOperation.Operation, ops[0])

			gotOperation, err := svc.GetDeprovisioningOperationByID("operation-id")
			require.NoError(t, err)

			op, err := svc.GetOperationByID("operation-id")
			require.NoError(t, err)
			assert.Equal(t, givenOperation.Operation.ID, op.ID)

			// then
			assertDeprovisioningOperation(t, givenOperation, *gotOperation)

			// when
			gotOperation.Description = "new modified description"
			_, err = svc.UpdateDeprovisioningOperation(*gotOperation)
			require.NoError(t, err)

			// then
			gotOperation2, err := svc.GetDeprovisioningOperationByID("operation-id")
			require.NoError(t, err)

			assert.Equal(t, "new modified description", gotOperation2.Description)

		})
	})

	t.Run("Operations conflicts", func(t *testing.T) {
		t.Run("Provisioning", func(t *testing.T) {
			containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			givenOperation := internal.ProvisioningOperation{
				Operation: internal.Operation{
					ID:    "operation-001",
					State: domain.InProgress,
					// used Round and set timezone to be able to compare timestamps
					CreatedAt:              time.Now(),
					UpdatedAt:              time.Now().Add(time.Second),
					InstanceID:             "inst-id",
					ProvisionerOperationID: "target-op-id",
					Description:            "description",
				},
				Lms:                    internal.LMS{TenantID: "tenant-id"},
				ProvisioningParameters: `{"key":"value"}`,
			}

			err = InitTestDBTables(t, cfg.ConnectionURL())
			require.NoError(t, err)

			brokerStorage, err := NewFromConfig(cfg, logrus.StandardLogger())
			svc := brokerStorage.Provisioning()

			require.NoError(t, err)
			require.NotNil(t, brokerStorage)
			err = svc.InsertProvisioningOperation(givenOperation)
			require.NoError(t, err)

			// when
			gotOperation1, err := svc.GetProvisioningOperationByID("operation-001")
			require.NoError(t, err)

			gotOperation2, err := svc.GetProvisioningOperationByID("operation-001")
			require.NoError(t, err)

			// when
			gotOperation1.Description = "new modified description 1"
			gotOperation2.Description = "new modified description 2"
			_, err = svc.UpdateProvisioningOperation(*gotOperation1)
			require.NoError(t, err)

			_, err = svc.UpdateProvisioningOperation(*gotOperation2)

			// then
			assertError(t, dberr.CodeConflict, err)

			// when
			err = svc.InsertProvisioningOperation(*gotOperation1)

			// then
			assertError(t, dberr.CodeAlreadyExists, err)
		})
		t.Run("Deprovisioning", func(t *testing.T) {
			containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			givenOperation := internal.DeprovisioningOperation{
				Operation: internal.Operation{
					ID:    "operation-001",
					State: domain.InProgress,
					// used Round and set timezone to be able to compare timestamps
					CreatedAt:              time.Now(),
					UpdatedAt:              time.Now().Add(time.Second),
					InstanceID:             "inst-id",
					ProvisionerOperationID: "target-op-id",
					Description:            "description",
				},
			}

			err = InitTestDBTables(t, cfg.ConnectionURL())
			require.NoError(t, err)

			brokerStorage, err := NewFromConfig(cfg, logrus.StandardLogger())
			require.NoError(t, err)

			svc := brokerStorage.Deprovisioning()

			err = svc.InsertDeprovisioningOperation(givenOperation)
			require.NoError(t, err)

			// when
			gotOperation1, err := svc.GetDeprovisioningOperationByID("operation-001")
			require.NoError(t, err)

			gotOperation2, err := svc.GetDeprovisioningOperationByID("operation-001")
			require.NoError(t, err)

			// when
			gotOperation1.Description = "new modified description 1"
			gotOperation2.Description = "new modified description 2"
			_, err = svc.UpdateDeprovisioningOperation(*gotOperation1)
			require.NoError(t, err)

			_, err = svc.UpdateDeprovisioningOperation(*gotOperation2)

			// then
			assertError(t, dberr.CodeConflict, err)

			// when
			err = svc.InsertDeprovisioningOperation(*gotOperation1)

			// then
			assertError(t, dberr.CodeAlreadyExists, err)
		})
	})

	t.Run("LMS Tenants", func(t *testing.T) {
		containerCleanupFunc, cfg, err := InitTestDBContainer(t, ctx, "test_DB_1")
		require.NoError(t, err)
		defer containerCleanupFunc()

		lmsTenant := internal.LMSTenant{
			ID:     "tenant-001",
			Region: "na",
			Name:   "some-company",
		}
		err = InitTestDBTables(t, cfg.ConnectionURL())
		require.NoError(t, err)

		brokerStorage, err := NewFromConfig(cfg, logrus.StandardLogger())
		svc := brokerStorage.LMSTenants()
		require.NoError(t, err)
		require.NotNil(t, brokerStorage)

		// when
		err = svc.InsertTenant(lmsTenant)
		require.NoError(t, err)
		gotTenant, found, err := svc.FindTenantByName("some-company", "na")
		_, differentRegionExists, drErr := svc.FindTenantByName("some-company", "us")
		_, differentNameExists, dnErr := svc.FindTenantByName("some-company1", "na")

		// then
		assert.Equal(t, lmsTenant.Name, gotTenant.Name)
		assert.True(t, found)
		assert.NoError(t, err)
		assert.False(t, differentRegionExists)
		assert.NoError(t, drErr)
		assert.False(t, differentNameExists)
		assert.NoError(t, dnErr)
	})
}

func assertProvisioningOperation(t *testing.T, expected, got internal.ProvisioningOperation) {
	// do not check zones and monothonic clock, see: https://golang.org/pkg/time/#Time
	assert.True(t, expected.CreatedAt.Equal(got.CreatedAt), fmt.Sprintf("Expected %s got %s", expected.CreatedAt, got.CreatedAt))
	assert.JSONEq(t, expected.ProvisioningParameters, got.ProvisioningParameters)

	expected.CreatedAt = got.CreatedAt
	expected.UpdatedAt = got.UpdatedAt
	expected.ProvisioningParameters = got.ProvisioningParameters
	assert.Equal(t, expected, got)
}

func assertDeprovisioningOperation(t *testing.T, expected, got internal.DeprovisioningOperation) {
	// do not check zones and monothonic clock, see: https://golang.org/pkg/time/#Time
	assert.True(t, expected.CreatedAt.Equal(got.CreatedAt), fmt.Sprintf("Expected %s got %s", expected.CreatedAt, got.CreatedAt))

	expected.CreatedAt = got.CreatedAt
	expected.UpdatedAt = got.UpdatedAt
	assert.Equal(t, expected, got)
}

func assertOperation(t *testing.T, expected, got internal.Operation) {
	// do not check zones and monothonic clock, see: https://golang.org/pkg/time/#Time
	assert.True(t, expected.CreatedAt.Equal(got.CreatedAt), fmt.Sprintf("Expected %s got %s", expected.CreatedAt, got.CreatedAt))

	expected.CreatedAt = got.CreatedAt
	expected.UpdatedAt = got.UpdatedAt
	assert.Equal(t, expected, got)
}

func assertError(t *testing.T, expectedCode int, err error) {
	require.Error(t, err)

	dbe, ok := err.(dberr.Error)
	if !ok {
		assert.Fail(t, "expected DB Error Conflict")
	}
	assert.Equal(t, expectedCode, dbe.Code())
}

func fixInstance(testData string) *internal.Instance {
	return &internal.Instance{
		InstanceID:             testData,
		RuntimeID:              testData,
		GlobalAccountID:        testData,
		ServiceID:              testData,
		ServicePlanID:          testData,
		DashboardURL:           testData,
		ProvisioningParameters: testData,
	}
}
