package k8s

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8SClient missing godoc
//go:generate mockery --name=K8SClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type K8SClient interface {
	Create(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Operation, error)
	Update(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error)
}

// Scheduler missing godoc
type Scheduler struct {
	kcli K8SClient
}

// NewScheduler missing godoc
func NewScheduler(kcli K8SClient) *Scheduler {
	return &Scheduler{
		kcli: kcli,
	}
}

// Schedule missing godoc
func (s *Scheduler) Schedule(ctx context.Context, op *operation.Operation) (string, error) {
	operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)
	getOp, err := s.kcli.Get(ctx, operationName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			k8sOp := toK8SOperation(op)
			createdOperation, err := s.kcli.Create(ctx, k8sOp)
			if err != nil {
				return "", err
			}
			return string(createdOperation.UID), nil
		}
		return "", err
	}
	if isOpInProgress(getOp) {
		return "", fmt.Errorf("another operation is in progress for resource with ID %q", op.ResourceID)
	}
	getOp = updateOperationSpec(op, getOp)
	updatedOperation, err := s.kcli.Update(ctx, getOp)
	if err != nil {
		if errors.IsConflict(err) {
			return "", fmt.Errorf("another operation is in progress for resource with ID %q", op.ResourceID)
		}
		return "", err
	}
	return string(updatedOperation.UID), err
}

func isOpInProgress(op *v1alpha1.Operation) bool {
	for _, cond := range op.Status.Conditions {
		if cond.Status == v1.ConditionTrue {
			return false
		}
	}
	return true
}

func toK8SOperation(op *operation.Operation) *v1alpha1.Operation {
	operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)
	result := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name: operationName,
		},
	}
	return updateOperationSpec(op, result)
}

func updateOperationSpec(op *operation.Operation, k8sOp *v1alpha1.Operation) *v1alpha1.Operation {
	k8sOp.Spec = v1alpha1.OperationSpec{
		OperationCategory: op.OperationCategory,
		OperationType:     v1alpha1.OperationType(str.Title(string(op.OperationType))),
		ResourceType:      string(op.ResourceType),
		ResourceID:        op.ResourceID,
		CorrelationID:     op.CorrelationID,
		WebhookIDs:        op.WebhookIDs,
		RequestObject:     op.RequestObject,
	}
	return k8sOp
}
