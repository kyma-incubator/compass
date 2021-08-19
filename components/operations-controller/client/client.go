package client

import (
	"context"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type OperationsClient struct {
	restClient rest.Interface
	namespace  string
}

type OperationsClientInterface interface {
	Operations(string) OperationsInterface
}

func NewForConfig(cfg *rest.Config) (OperationsClientInterface, error) {
	if err := v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	cfg.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupVersion.Group, Version: v1alpha1.GroupVersion.Version}
	cfg.APIPath = "/apis"
	cfg.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{
		CodecFactory: scheme.Codecs,
	}
	cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	c, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}
	return &OperationsClient{
		restClient: c,
	}, nil
}

func (c *OperationsClient) Operations(namespace string) OperationsInterface {
	c.namespace = namespace
	return c
}

func (c *OperationsClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.OperationList, error) {
	result := v1alpha1.OperationList{}
	err := c.restClient.Get().
		Namespace(c.namespace).
		Resource("operations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *OperationsClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Operation, error) {
	result := v1alpha1.Operation{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource("operations").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *OperationsClient) Create(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error) {
	result := v1alpha1.Operation{}
	err := c.restClient.
		Post().
		Namespace(c.namespace).
		Resource("operations").
		Body(operation).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *OperationsClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(c.namespace).
		Resource("operations").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Error()
}

func (c *OperationsClient) Update(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error) {
	err := c.restClient.
		Put().
		Namespace(c.namespace).
		Resource("operations").
		Name(operation.Name).
		Body(operation).
		Do(ctx).
		Into(operation)

	return operation, err
}

func (c *OperationsClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.namespace).
		Resource("operations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
