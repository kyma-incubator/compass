package k8s

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/op-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=K8SClient -output=automock -outpkg=automock -case=underscore
type K8SClient interface {
	Create(ctx context.Context, operation *v1beta1.Operation) (*v1beta1.Operation, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1beta1.Operation, error)
	Update(ctx context.Context, operation *v1beta1.Operation) (*v1beta1.Operation, error)
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
	if errors.IsNotFound(err) {
		k8sOp := toK8SOperation(op)
		createdOperation, err := s.kcli.Create(ctx, k8sOp)
		return string(createdOperation.UID), err
	}
	// TODO: Check if operation is in progress, if true: return an error that op is in progress; otherwise proceed with op update
	getOp = toK8SOperation(op)
	updatedOperation, err := s.kcli.Update(ctx, getOp)
	if errors.IsConflict(err) {
		return "", fmt.Errorf("another operation is in progress")
	}
	return string(updatedOperation.UID), err
}

func toK8SOperation(op *operation.Operation) *v1beta1.Operation {
	// TODO: Implement when controller skeleton is merged
	return &v1beta1.Operation{}
}
