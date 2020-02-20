// +build database-integration

package storage

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"

	"time"

	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/pivotal-cf/brokerapi/v7/domain"
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
			containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			// when
			connection, err := postsql.InitializeDatabase(connString)

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
			containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			err = InitTestDBTables(t, connString)
			require.NoError(t, err)

			// when
			brokerStorage, err := New(connString)

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

		})
	})

	t.Run("Operations", func(t *testing.T) {
		containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
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
			LmsTenantID:            "tenant-id",
			ProvisioningParameters: `{"k":"v"}`,
		}

		err = InitTestDBTables(t, connString)
		require.NoError(t, err)

		brokerStorage, err := New(connString)
		svc := brokerStorage.Operations()

		require.NoError(t, err)
		require.NotNil(t, brokerStorage)

		// when
		err = svc.InsertProvisioningOperation(givenOperation)
		require.NoError(t, err)

		gotOperation, err := svc.GetProvisioningOperationByID("operation-id")
		require.NoError(t, err)

		op, err := svc.GetOperation("operation-id")
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

	t.Run("Operations conflicts", func(t *testing.T) {
		containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
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
			LmsTenantID:            "tenant-id",
			ProvisioningParameters: `{"key":"value"}`,
		}

		err = InitTestDBTables(t, connString)
		require.NoError(t, err)

		brokerStorage, err := New(connString)
		svc := brokerStorage.Operations()

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
