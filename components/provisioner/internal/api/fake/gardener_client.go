package fake

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"testing"
)

func NewFakeShootsInterface(t *testing.T, config *rest.Config) gardener_apis.ShootInterface {
	dynamicClient, err := dynamic.NewForConfig(config)
	require.NoError(t, err)

	resourceInterface := dynamicClient.Resource(gardener_types.SchemeGroupVersion.WithResource("shoots"))
	return &fakeShootsInterface{
		client: resourceInterface,
	}
}

type fakeShootsInterface struct {
	client dynamic.ResourceInterface
}

func (f fakeShootsInterface) Create(shoot *gardener_types.Shoot) (*gardener_types.Shoot, error) {
	addTypeMeta(shoot)

	shoot.SetFinalizers([]string{"finalizer"})

	seedName := "az-us2"
	shoot.Spec.SeedName = &seedName

	unstructuredShoot, err := toUnstructured(shoot)

	if err != nil {
		return nil, err
	}

	create, err := f.client.Create(unstructuredShoot, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return fromUnstructured(create)
}

func (f *fakeShootsInterface) Update(shoot *gardener_types.Shoot) (*gardener_types.Shoot, error) {
	obj, err := toUnstructured(shoot)

	if err != nil {
		return nil, err
	}

	updated, err := f.client.Update(obj, metav1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	return fromUnstructured(updated)
}

func (f *fakeShootsInterface) UpdateStatus(*gardener_types.Shoot) (*gardener_types.Shoot, error) {
	return nil, nil
}

func (f *fakeShootsInterface) Delete(name string, options *metav1.DeleteOptions) error {
	return f.client.Delete(name, options)
}

func (f *fakeShootsInterface) DeleteCollection(_ *metav1.DeleteOptions, _ metav1.ListOptions) error {
	return nil
}

func (f *fakeShootsInterface) Get(name string, options metav1.GetOptions) (*gardener_types.Shoot, error) {
	obj, err := f.client.Get(name, options)

	if err != nil {
		return nil, err
	}

	return fromUnstructured(obj)
}
func (f *fakeShootsInterface) List(opts metav1.ListOptions) (*gardener_types.ShootList, error) {
	list, err := f.client.List(opts)

	if err != nil {
		return nil, err
	}

	return listFromUnstructured(list)
}
func (f *fakeShootsInterface) Watch(_ metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (f *fakeShootsInterface) Patch(_ string, pt types.PatchType, _ []byte, _ ...string) (result *gardener_types.Shoot, err error) {
	return nil, nil
}

func addTypeMeta(shoot *gardener_types.Shoot) {
	shoot.TypeMeta = metav1.TypeMeta{
		Kind:       "Shoot",
		APIVersion: "core.gardener.cloud/v1beta1",
	}
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)

	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func fromUnstructured(object *unstructured.Unstructured) (*gardener_types.Shoot, error) {
	var newShoot gardener_types.Shoot

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, &newShoot)

	if err != nil {
		return nil, err
	}

	return &newShoot, err
}

func listFromUnstructured(list *unstructured.UnstructuredList) (*gardener_types.ShootList, error) {
	shootList := &gardener_types.ShootList{
		Items: []gardener_types.Shoot{},
	}

	for _, obj := range list.Items {
		shoot, err := fromUnstructured(&obj)
		if err != nil {
			return &gardener_types.ShootList{}, err
		}
		shootList.Items = append(shootList.Items, *shoot)
	}
	return shootList, nil
}
