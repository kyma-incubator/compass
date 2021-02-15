package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=K8SClient -output=automock -outpkg=automock -case=underscore
type K8SClient interface {
	Create(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Operation, error)
	Update(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

type Scheduler struct {
	kcli K8SClient
}

func NewScheduler(kcli K8SClient) *Scheduler {
	return &Scheduler{
		kcli: kcli,
	}
}

func (s *Scheduler) Schedule(ctx context.Context, op *operation.Operation) (string, error) {
	operationName := fmt.Sprintf("%s-%s", op.ResourceType, op.ResourceID)
	getOp, err := s.kcli.Get(ctx, operationName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			k8sOp := toK8SOperation(op)
			createdOperation, err := s.kcli.Create(ctx, k8sOp)
			return string(createdOperation.UID), err
		}
		return "", err
	}
	if isOpInProgress(getOp) {
		return "", fmt.Errorf("another operation is in progress")
	}
	if err := s.kcli.Delete(ctx, operationName, metav1.DeleteOptions{}); err != nil {
		return "", fmt.Errorf("could not delete operation: %s", err)
	}
	k8sOp := toK8SOperation(op)
	createdOperation, err := s.kcli.Create(ctx, k8sOp)
	if err != nil {
		return "", err
	}
	return string(createdOperation.UID), err
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
		OperationType:     v1alpha1.OperationType(title(string(op.OperationType))),
		ResourceType:      string(op.ResourceType),
		ResourceID:        op.ResourceID,
		CorrelationID:     op.CorrelationID,
		WebhookIDs:        op.WebhookIDs,
		RequestData:       op.RequestData,
	}
	return k8sOp
}

func title(s string) string {
	return strings.Title(strings.ToLower(s))
}
