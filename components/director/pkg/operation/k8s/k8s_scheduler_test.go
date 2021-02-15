package k8s_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/k8s"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/k8s/automock"
)

const (
	resourceID  = "c7092c57-7a5c-4ebe-8c58-03c0f85ade6c"
	operationID = "0b4fc816-da70-4505-961e-db346388fdb7"
)

func TestScheduler_Schedule(t *testing.T) {
	t.Run("when the k8s client fails to retrieve the operation it should fail", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		op := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		expErr := errors.New("error")
		operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)

		cli := &automock.K8SClient{}
		cli.On("Get", ctx, operationName, metav1.GetOptions{}).Return(nil, expErr).Once()
		s := k8s.NewScheduler(cli)

		// WHEN
		_, err := s.Schedule(ctx, op)
		// THEN
		require.Equal(t, expErr, err)
	})

	t.Run("when no previous operation exists and the k8s client fails to create a new one it should fail", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		op := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)

		cli := &automock.K8SClient{}
		cli.On("Get", ctx, operationName, metav1.GetOptions{}).Return(nil, k8s_errors.NewNotFound(schema.GroupResource{}, "test")).Once()

		k8sOp := toK8SOperation(op)
		expErr := errors.New("error")
		cli.On("Create", ctx, k8sOp).Return(nil, expErr).Once()

		s := k8s.NewScheduler(cli)

		// WHEN
		_, err := s.Schedule(ctx, op)
		// THEN
		require.Error(t, err)
		require.Equal(t, expErr, err)
	})

	t.Run("when no previous operation exists it should return the ID of a newly created operation", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		op := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)

		cli := &automock.K8SClient{}
		notFoundErr := k8s_errors.NewNotFound(schema.GroupResource{}, operationName)
		cli.On("Get", ctx, operationName, metav1.GetOptions{}).Return(nil, notFoundErr).Once()

		k8sOp := toK8SOperation(op)
		k8sOp.UID = operationID
		k8sOpWithoutID := toK8SOperation(op)

		cli.On("Create", ctx, k8sOpWithoutID).Return(k8sOp, nil).Once()

		s := k8s.NewScheduler(cli)

		// WHEN
		opID, err := s.Schedule(ctx, op)
		// THEN
		require.NoError(t, err)
		require.Equal(t, string(k8sOp.UID), opID)
	})

	t.Run("when a previous operation is in progress it should return an error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		op := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)

		cli := &automock.K8SClient{}

		k8sOp := toK8SOperation(op)
		k8sOp.UID = operationID

		cli.On("Get", ctx, operationName, metav1.GetOptions{}).Return(k8sOp, nil).Once()

		s := k8s.NewScheduler(cli)

		// WHEN
		_, err := s.Schedule(ctx, op)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("another operation is in progress for resource with ID %q", op.ResourceID))
	})

	t.Run("when operation update fails it should return an error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		op := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)

		cli := &automock.K8SClient{}

		k8sOp := toK8SOperation(op)
		k8sOp.UID = operationID
		k8sOp.Status.Conditions = []v1alpha1.Condition{{
			Status: v1.ConditionTrue,
		}}

		cli.On("Get", ctx, operationName, metav1.GetOptions{}).Return(k8sOp, nil).Once()
		expErr := errors.New("error")
		cli.On("Update", ctx, k8sOp).Return(nil, expErr).Once()

		defer cli.AssertExpectations(t)
		s := k8s.NewScheduler(cli)

		// WHEN
		_, err := s.Schedule(ctx, op)
		// THEN
		require.Error(t, err)
		require.Equal(t, expErr, err)
	})

	t.Run("when another operation update is in progress it should return an error", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		op := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)

		cli := &automock.K8SClient{}

		k8sOp := toK8SOperation(op)
		k8sOp.UID = operationID
		k8sOp.Status.Conditions = []v1alpha1.Condition{{
			Status: v1.ConditionTrue,
		}}

		cli.On("Get", ctx, operationName, metav1.GetOptions{}).Return(k8sOp, nil).Once()
		conflictErr := k8s_errors.NewConflict(schema.GroupResource{}, operationName, errors.New("error"))
		cli.On("Update", ctx, k8sOp).Return(nil, conflictErr).Once()

		defer cli.AssertExpectations(t)
		s := k8s.NewScheduler(cli)

		// WHEN
		_, err := s.Schedule(ctx, op)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("another operation is in progress for resource with ID %q", op.ResourceID))
	})

	t.Run("when the operation is updated successfully it should return the ID of the operation", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		cli := &automock.K8SClient{}

		completedOp := &operation.Operation{OperationType: operation.OperationTypeCreate, ResourceID: resourceID}
		completedOpName := fmt.Sprintf("%s-%s", completedOp.ResourceType, completedOp.ResourceID)
		completedk8sOp := toK8SOperation(completedOp)
		completedk8sOp.UID = operationID
		completedk8sOp.Status.Conditions = []v1alpha1.Condition{{
			Status: v1.ConditionTrue,
		}}

		newOp := &operation.Operation{OperationType: operation.OperationTypeUpdate, ResourceID: resourceID}
		newK8sOp := toK8SOperation(newOp)
		newK8sOp.UID = operationID

		cli.On("Get", ctx, completedOpName, metav1.GetOptions{}).Return(completedk8sOp, nil).Once()
		cli.On("Update", ctx, completedk8sOp).Return(newK8sOp, nil).Once()

		s := k8s.NewScheduler(cli)

		// WHEN
		opID, err := s.Schedule(ctx, newOp)
		// THEN
		require.NoError(t, err)
		require.Equal(t, string(newK8sOp.UID), opID)
	})
}

func toK8SOperation(op *operation.Operation) *v1alpha1.Operation {
	operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)
	return &v1alpha1.Operation{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: operationName,
		},
		Spec: v1alpha1.OperationSpec{
			OperationType: "Create",
			ResourceID:    op.ResourceID,
		},
		Status: v1alpha1.OperationStatus{},
	}
}
