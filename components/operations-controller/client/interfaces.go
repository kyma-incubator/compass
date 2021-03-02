package client

import (
	"context"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type OperationsInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.OperationList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Operation, error)
	Create(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Update(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}
